package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Model is the chat renderer — the always-visible canvas.
// It is NOT a "view" — it's the core of the TUI.
type Model struct {
	client   *client.Client
	theme    *theme.Theme
	styles   *theme.Styles
	viewport viewport.Model
	input    textarea.Model

	// Dimensions
	width  int
	height int

	// State
	messages      []Message
	models        []llm.Model
	activeModel   int
	streaming     bool
	streamBuf     *strings.Builder
	thinkingFrame int

	// Stats
	lastTokenCount    int
	lastDuration      time.Duration
	lastSpeed         float64
	streamStart       time.Time
	sessionTokenCount int // Cumulative tokens for session

	// Error
	err error

	// Input visibility (controlled by mode)
	inputVisible   bool
	commandVisible bool
	commandInput   string

	// System prompt
	systemPrompt string

	// Preferred model (loaded from config, applied when models arrive)
	preferredModel string

	// Tool execution
	toolExecutor    *llmtools.Executor
	toolsEnabled    bool
	pendingToolCall *llm.ToolCall       // Tool waiting for approval
	toolInputBuf    *strings.Builder    // Accumulates streaming tool input JSON
	currentToolUse  *llm.ToolCall       // Tool use being streamed
	executingTool   bool                // Whether we're executing a tool
	toolResults     []llm.ToolResult    // Results to send back to LLM
}

// Message represents a chat message (user, assistant, or system).
type Message struct {
	Role    string    // "user", "assistant", "system"
	Content string
	Time    time.Time // when the message was created
}

// Messages for Bubble Tea
type modelsMsg struct {
	models []llm.Model
	err    error
}

type streamChunkMsg struct {
	chunk llm.ChatResponse
}

type streamDoneMsg struct {
	totalTokens int
	duration    time.Duration
	reason      string // debug: why stream ended
}

type streamErrorMsg struct {
	err error
}

type thinkingTickMsg struct{}

type continueStreamMsg struct{}

// Tool-related messages
type toolUseStartMsg struct {
	id   string
	name string
}

type toolInputDeltaMsg struct {
	delta string
}

type toolUseCompleteMsg struct {
	call llm.ToolCall
}

type toolApprovalRequestMsg struct {
	tool llmtools.Tool
	call llm.ToolCall
}

type toolApprovalResponseMsg struct {
	approved        bool
	grantForSession bool
	call            llm.ToolCall
}

type toolExecutionResultMsg struct {
	result llm.ToolResult
}

type toolContinueMsg struct{} // Signal to continue after tool execution

// New creates a new chat model.
func New(c *client.Client, t *theme.Theme, s *theme.Styles) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	// Custom border with only top/bottom lines (no corners, no sides)
	topBottomBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "",
		Right:       "",
		TopLeft:     "",
		TopRight:    "",
		BottomLeft:  "",
		BottomRight: "",
	}
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(topBottomBorder).
		BorderTop(true).
		BorderBottom(true).
		BorderForeground(t.BorderFocus)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(topBottomBorder).
		BorderTop(true).
		BorderBottom(true).
		BorderForeground(t.Border)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	return Model{
		client:       c,
		theme:        t,
		styles:       s,
		viewport:     vp,
		input:        ta,
		messages:     []Message{},
		streamBuf:    &strings.Builder{},
		toolInputBuf: &strings.Builder{},
	}
}

// SetToolExecutor sets the tool executor for function calling.
func (m *Model) SetToolExecutor(executor *llmtools.Executor) {
	m.toolExecutor = executor
}

// EnableTools enables or disables tool/function calling.
func (m *Model) EnableTools(enabled bool) {
	m.toolsEnabled = enabled
}

// ToolsEnabled returns whether tools are enabled.
func (m Model) ToolsEnabled() bool {
	return m.toolsEnabled && m.toolExecutor != nil
}

// ToolExecutor returns the tool executor (may be nil).
func (m Model) ToolExecutor() *llmtools.Executor {
	return m.toolExecutor
}

// HasPendingApproval returns true if there's a tool waiting for user approval.
func (m Model) HasPendingApproval() bool {
	return m.pendingToolCall != nil
}

// PendingToolCall returns the tool call waiting for approval.
func (m Model) PendingToolCall() *llm.ToolCall {
	return m.pendingToolCall
}

// Init fetches available models.
func (m Model) Init() tea.Cmd {
	return m.fetchModels
}

// Update handles messages routed from the app.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		return m, nil

	case modelsMsg:
		m.models = msg.models
		m.err = msg.err
		// Apply preferred model if set
		if m.preferredModel != "" && len(m.models) > 0 {
			for i, model := range m.models {
				if model.Name == m.preferredModel {
					m.activeModel = i
					break
				}
			}
		}
		return m, nil

	case streamChunkMsg:
		// Handle both formats: nested (msg.chunk.Message.Content) and flat (msg.chunk.Content)
		content := ""
		if msg.chunk.Message != nil && msg.chunk.Message.Content != "" {
			content = msg.chunk.Message.Content
		} else if msg.chunk.Content != "" {
			content = msg.chunk.Content
		}
		if content != "" {
			m.streamBuf.WriteString(content)
			m.updateStreamingMessage()
		}
		// Debug: count chunks received
		m.lastTokenCount++ // Repurpose as chunk counter for debug
		return m, func() tea.Msg { return pollStreamCmd() }

	case continueStreamMsg:
		return m, tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
			return pollStreamCmd()
		})

	case streamDoneMsg:
		m.streaming = false
		m.lastTokenCount = msg.totalTokens
		m.sessionTokenCount += msg.totalTokens // Accumulate session tokens
		m.lastDuration = msg.duration
		if msg.duration > 0 {
			m.lastSpeed = float64(msg.totalTokens) / msg.duration.Seconds()
		}
		bufContent := m.streamBuf.String()
		if len(bufContent) > 0 {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: bufContent,
				Time:    time.Now(),
			})
		} else {
			// Debug: no content received
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("[Debug: Stream ended with no content. Reason: %s, Tokens: %d, Duration: %v]", msg.reason, msg.totalTokens, msg.duration),
				Time:    time.Now(),
			})
		}
		m.streamBuf.Reset()
		m.updateViewport()

		// If we have tool results, continue the conversation
		if len(m.toolResults) > 0 {
			return m, m.ContinueAfterToolResult()
		}
		return m, nil

	case streamErrorMsg:
		m.streaming = false
		// If we have partial content, save it before showing error
		if m.streamBuf.Len() > 0 {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: m.streamBuf.String(),
				Time:    time.Now(),
			})
			m.streamBuf.Reset()
			m.updateViewport()
		}
		// Only show error if it's not a normal EOF
		errStr := msg.err.Error()
		if errStr != "EOF" && errStr != "unexpected EOF" {
			m.err = msg.err
		}
		return m, nil

	case thinkingTickMsg:
		if m.streaming || m.executingTool {
			m.thinkingFrame = (m.thinkingFrame + 1) % len(ThinkingFrames)
			return m, m.thinkingTick()
		}
		return m, nil

	// Tool-related message handling
	case toolUseStartMsg:
		m.currentToolUse = &llm.ToolCall{
			ID:   msg.id,
			Name: msg.name,
		}
		m.toolInputBuf.Reset()
		return m, nil

	case toolInputDeltaMsg:
		m.toolInputBuf.WriteString(msg.delta)
		return m, nil

	case toolUseCompleteMsg:
		// Tool use is complete, check if it needs approval
		return m, m.handleToolUseComplete(msg.call)

	case toolApprovalResponseMsg:
		return m, m.handleApprovalResponse(msg)

	case toolExecutionResultMsg:
		m.toolResults = append(m.toolResults, msg.result)
		m.executingTool = false
		// Show the tool result in chat
		m.showToolResult(msg.result)
		// Automatically continue the conversation with tool results
		return m, m.ContinueAfterToolResult()

	case toolContinueMsg:
		// Continue the conversation with tool results
		return m, m.continueWithToolResults()
	}

	// Update textarea when input is visible and not streaming
	if m.inputVisible && !m.streaming {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// -- Public API for the app to drive the chat --

// SetInputVisible shows/hides the textarea (Insert mode).
func (m *Model) SetInputVisible(visible bool) {
	m.inputVisible = visible
	if visible {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
	m.resize()
}

// ScrollUp scrolls the viewport up by n lines.
func (m *Model) ScrollUp(n int) {
	m.viewport.LineUp(n)
}

// ScrollDown scrolls the viewport down by n lines.
func (m *Model) ScrollDown(n int) {
	m.viewport.LineDown(n)
}

// HalfPageUp scrolls up half a page.
func (m *Model) HalfPageUp() {
	m.viewport.HalfViewUp()
}

// HalfPageDown scrolls down half a page.
func (m *Model) HalfPageDown() {
	m.viewport.HalfViewDown()
}

// GotoTop jumps to the beginning of chat.
func (m *Model) GotoTop() {
	m.viewport.GotoTop()
}

// GotoBottom jumps to the end of chat.
func (m *Model) GotoBottom() {
	m.viewport.GotoBottom()
}

// SendCurrentInput sends the current textarea content as a user message.
func (m *Model) SendCurrentInput() tea.Cmd {
	content := strings.TrimSpace(m.input.Value())
	if content == "" || m.streaming {
		return nil
	}

	m.messages = append(m.messages, Message{
		Role:    "user",
		Content: content,
		Time:    time.Now(),
	})
	m.input.Reset()
	m.streaming = true
	m.streamBuf.Reset()
	m.streamStart = time.Now()
	m.lastTokenCount = 0
	m.lastDuration = 0
	m.lastSpeed = 0
	m.err = nil
	m.updateViewport()

	return tea.Batch(
		m.sendMessage(),
		m.thinkingTick(),
	)
}

// InsertNewline adds a newline at the cursor position in the input.
func (m *Model) InsertNewline() {
	m.input.InsertString("\n")
}

// InputLen returns the current input length in characters.
func (m Model) InputLen() int {
	return len(m.input.Value())
}

// InputValue returns the current input text.
func (m Model) InputValue() string {
	return m.input.Value()
}

// SetInputValue sets the input text (for history navigation).
func (m *Model) SetInputValue(v string) {
	m.input.SetValue(v)
}

// RetryLast re-sends the last user message. Removes the last assistant response
// if it immediately follows the last user message, then re-triggers streaming.
func (m *Model) RetryLast() tea.Cmd {
	if m.streaming || len(m.messages) == 0 {
		return nil
	}

	// Find last user message
	lastUserIdx := -1
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "user" {
			lastUserIdx = i
			break
		}
	}
	if lastUserIdx == -1 {
		return nil
	}

	// Remove any assistant/system messages after the last user message
	m.messages = m.messages[:lastUserIdx+1]

	// Re-trigger streaming
	m.streaming = true
	m.streamBuf.Reset()
	m.streamStart = time.Now()
	m.lastTokenCount = 0
	m.lastDuration = 0
	m.lastSpeed = 0
	m.err = nil
	m.updateViewport()

	return tea.Batch(
		m.sendMessage(),
		m.thinkingTick(),
	)
}

// LastAssistantMessage returns the content of the most recent assistant message.
func (m Model) LastAssistantMessage() string {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" {
			return m.messages[i].Content
		}
	}
	return ""
}

// CancelStreaming stops the current stream.
func (m *Model) CancelStreaming() {
	if activeStream != nil && activeStream.cancel != nil {
		activeStream.cancel()
		activeStream = nil
	}
	m.streaming = false
	if m.streamBuf.Len() > 0 {
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: m.streamBuf.String() + " [cancelled]",
			Time:    time.Now(),
		})
		m.streamBuf.Reset()
	}
	m.updateViewport()
}

// InjectSystemMessage adds a system message to the chat history.
func (m *Model) InjectSystemMessage(content string) {
	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: content,
		Time:    time.Now(),
	})
	m.updateViewport()
}

// ClearMessages removes all chat messages.
func (m *Model) ClearMessages() {
	m.messages = []Message{}
	m.lastTokenCount = 0
	m.lastSpeed = 0
	m.updateViewport()
}

// SwitchModel switches the active model by name.
func (m *Model) SwitchModel(name string) {
	for i, model := range m.models {
		if strings.EqualFold(model.Name, name) || strings.HasPrefix(strings.ToLower(model.Name), strings.ToLower(name)) {
			m.activeModel = i
			m.InjectSystemMessage("Switched to model: " + model.Name)
			return
		}
	}
	m.InjectSystemMessage("Model not found: " + name)
}

// CycleModel cycles to the next available model.
func (m *Model) CycleModel() {
	if len(m.models) > 0 {
		m.activeModel = (m.activeModel + 1) % len(m.models)
	}
}

// CycleModelReverse cycles to the previous available model.
func (m *Model) CycleModelReverse() {
	if len(m.models) > 0 {
		m.activeModel = (m.activeModel - 1 + len(m.models)) % len(m.models)
	}
}

// ActiveModelName returns the name of the currently active model.
func (m Model) ActiveModelName() string {
	if len(m.models) == 0 {
		return ""
	}
	if m.activeModel < len(m.models) {
		return m.models[m.activeModel].Name
	}
	return ""
}

// ActiveModelProvider returns the provider of the currently active model.
func (m Model) ActiveModelProvider() string {
	if len(m.models) == 0 {
		return ""
	}
	if m.activeModel < len(m.models) {
		return m.models[m.activeModel].Provider
	}
	return ""
}

// SessionTokenCount returns the cumulative token count for this session.
func (m Model) SessionTokenCount() int {
	return m.sessionTokenCount
}

// IsPaidProvider returns true if the active model uses a commercial provider.
func (m Model) IsPaidProvider() bool {
	provider := m.ActiveModelProvider()
	switch provider {
	case "anthropic", "openai", "google", "groq", "together":
		return true
	default:
		return false
	}
}

// IsStreaming returns whether a response is being streamed.
func (m Model) IsStreaming() bool {
	return m.streaming
}

// HasError returns whether there was an error in the last operation.
func (m Model) HasError() bool {
	return m.err != nil
}

// LastError returns the last error message, if any.
func (m Model) LastError() string {
	if m.err == nil {
		return ""
	}
	return m.err.Error()
}

// ClearError clears the last error.
func (m *Model) ClearError() {
	m.err = nil
}

// ExportMessages returns all messages for export.
func (m Model) ExportMessages() []ExportMsg {
	var msgs []ExportMsg
	for _, msg := range m.messages {
		ts := ""
		if !msg.Time.IsZero() {
			ts = msg.Time.Format("2006-01-02 15:04:05")
		}
		msgs = append(msgs, ExportMsg{
			Role:    msg.Role,
			Content: msg.Content,
			Time:    ts,
		})
	}
	return msgs
}

// ExportMsg is a message suitable for export (no internal state).
type ExportMsg struct {
	Role    string
	Content string
	Time    string
}

// LoadMessages replaces all messages (for loading saved conversations).
func (m *Model) LoadMessages(msgs []Message) {
	m.messages = msgs
	m.updateViewport()
}

// Messages returns the current message list.
func (m Model) Messages() []Message {
	return m.messages
}

// SetSystemPrompt sets the system prompt prepended to LLM requests.
func (m *Model) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

// GetSystemPrompt returns the current system prompt.
func (m Model) GetSystemPrompt() string {
	return m.systemPrompt
}

// SetPreferredModel sets the preferred model to select when models are loaded.
func (m *Model) SetPreferredModel(name string) {
	m.preferredModel = name
}

// -- View rendering --

// ViewChat renders just the chat area (messages + streaming).
func (m Model) ViewChat() string {
	return m.viewport.View()
}

// ViewInput renders the textarea (for Insert mode).
func (m Model) ViewInput() string {
	if !m.inputVisible {
		return ""
	}
	return m.input.View()
}

// ViewStats renders the stats line (after streaming completes).
func (m Model) ViewStats() string {
	if m.streaming {
		return m.renderStreaming()
	}
	if m.lastTokenCount > 0 {
		return m.renderStats()
	}
	return ""
}

// ViewError renders any error.
func (m Model) ViewError() string {
	if m.err != nil {
		return m.styles.Error.Render("  " + m.err.Error())
	}
	return ""
}

// -- Internal rendering --

func (m Model) renderStreaming() string {
	frame := ThinkingFrames[m.thinkingFrame]
	sparkle := Sparkles[m.thinkingFrame%len(Sparkles)]

	streamStyle := lipgloss.NewStyle().Foreground(m.theme.StreamingColor).Bold(true)
	subtleStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)

	// Model name
	modelPart := ""
	if name := m.ActiveModelName(); name != "" {
		modelPart = subtleStyle.Render(" via " + name)
	}

	// Elapsed time
	elapsed := time.Since(m.streamStart)
	elapsedPart := subtleStyle.Render(fmt.Sprintf("  %0.1fs", elapsed.Seconds()))

	// Cancel hint
	cancelHint := subtleStyle.Render("  (Esc to cancel)")

	return streamStyle.Render("  "+sparkle+" "+frame+" "+sparkle) + modelPart + elapsedPart + cancelHint
}

func (m Model) renderStats() string {
	durationPart := lipgloss.NewStyle().Foreground(m.theme.TextMuted).
		Render(fmt.Sprintf("  %0.1fs", m.lastDuration.Seconds()))
	return "  " + FormatTokens(m.lastTokenCount, m.theme) + "  " + FormatSpeed(m.lastSpeed, m.theme) + durationPart
}

func (m *Model) resize() {
	// m.height is already the chat area height (app subtracts header, statusbar, input, stats)
	vpHeight := m.height
	if vpHeight < 5 {
		vpHeight = 5
	}

	// Dynamic horizontal padding based on terminal width
	hPadding := 4
	if m.width < 80 {
		hPadding = 2
	} else if m.width < 60 {
		hPadding = 1
	}

	vpWidth := m.width - hPadding
	if vpWidth < 40 {
		vpWidth = 40
	}

	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.input.SetWidth(vpWidth - 2)

	m.updateViewportPreserveScroll()
}

func (m *Model) updateViewport() {
	content := m.renderMessages()
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) updateViewportPreserveScroll() {
	// Preserve scroll position as percentage
	oldTotal := m.viewport.TotalLineCount()
	oldOffset := m.viewport.YOffset
	scrollPercent := 0.0
	if oldTotal > 0 {
		scrollPercent = float64(oldOffset) / float64(oldTotal)
	}
	atBottom := m.viewport.AtBottom()

	content := m.renderMessages()
	m.viewport.SetContent(content)

	// Restore scroll position
	if atBottom {
		m.viewport.GotoBottom()
	} else {
		newTotal := m.viewport.TotalLineCount()
		newOffset := int(scrollPercent * float64(newTotal))
		m.viewport.SetYOffset(newOffset)
	}
}

func (m *Model) updateStreamingMessage() {
	content := m.renderMessages()
	if m.streamBuf.Len() > 0 {
		content += "\n" + m.styles.AssistantLabel.Render("✦ Hecate") + "\n"
		bubble := m.styles.AssistantBubble.Width(m.viewport.Width - 8).Render(m.streamBuf.String() + "▊")
		content += bubble
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		welcome := WelcomeArt(m.theme)
		return lipgloss.Place(
			m.viewport.Width,
			m.viewport.Height,
			lipgloss.Center,
			lipgloss.Center,
			welcome,
		)
	}

	var parts []string
	bubbleWidth := m.viewport.Width - 8
	if bubbleWidth < 30 {
		bubbleWidth = 30
	}

	timeStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)

	for _, msg := range m.messages {
		timestamp := ""
		if !msg.Time.IsZero() {
			timestamp = timeStyle.Render(" " + msg.Time.Format("15:04"))
		}

		switch msg.Role {
		case "user":
			label := m.styles.UserLabel.Render("▸ You") + timestamp
			bubble := m.styles.UserBubble.Render(msg.Content)
			parts = append(parts, label+"\n"+bubble)

		case "assistant":
			label := m.styles.AssistantLabel.Render("◆ Hecate") + timestamp
			rendered := RenderMarkdown(msg.Content, m.theme, bubbleWidth-4)
			bubble := m.styles.AssistantBubble.Width(bubbleWidth).Render(rendered)
			parts = append(parts, label+"\n"+bubble)

		case "system":
			bubble := m.styles.SystemBubble.Width(bubbleWidth).Render(msg.Content)
			parts = append(parts, bubble)
		}
	}

	return strings.Join(parts, "\n\n")
}

// -- Commands (streaming) --

func (m Model) fetchModels() tea.Msg {
	models, err := m.client.ListModels()
	return modelsMsg{models: models, err: err}
}

type streamState struct {
	ctx         context.Context
	cancel      context.CancelFunc
	respChan    <-chan llm.ChatResponse
	errChan     <-chan error
	start       time.Time
	totalTokens int
}

var activeStream *streamState

func (m *Model) sendMessage() tea.Cmd {
	return m.sendMessageWithToolResults(nil)
}

func (m *Model) sendMessageWithToolResults(toolResults []llm.ToolResult) tea.Cmd {
	return func() tea.Msg {
		if len(m.models) == 0 {
			return streamErrorMsg{err: fmt.Errorf("no models available")}
		}

		modelName := m.models[m.activeModel].Name
		ctx, cancel := context.WithCancel(context.Background())

		// Convert our messages to llm.Message
		var llmMsgs []llm.Message

		// Prepend system prompt if set
		if m.systemPrompt != "" {
			llmMsgs = append(llmMsgs, llm.Message{
				Role:    llm.RoleSystem,
				Content: m.systemPrompt,
			})
		}

		for _, msg := range m.messages {
			if msg.Role == "system" {
				continue // Don't send system messages to LLM
			}
			llmMsgs = append(llmMsgs, llm.Message{
				Role:    llm.Role(msg.Role),
				Content: msg.Content,
			})
		}

		// Add tool results if any
		for _, result := range toolResults {
			llmMsgs = append(llmMsgs, llm.Message{
				Role:       llm.RoleTool,
				Content:    result.Content,
				ToolCallID: result.ToolCallID,
			})
		}

		req := llm.ChatRequest{
			Model:    modelName,
			Messages: llmMsgs,
			Stream:   true,
		}

		// Add tool schemas if tools are enabled
		if m.toolsEnabled && m.toolExecutor != nil {
			req.Tools = m.buildToolSchemas()
		}

		start := time.Now()
		respChan, errChan := m.client.ChatStream(ctx, req)

		activeStream = &streamState{
			ctx:      ctx,
			cancel:   cancel,
			respChan: respChan,
			errChan:  errChan,
			start:    start,
		}

		return pollStreamCmd()
	}
}

// buildToolSchemas converts the tool registry to LLM tool schemas.
func (m *Model) buildToolSchemas() []llm.ToolSchema {
	if m.toolExecutor == nil {
		return nil
	}

	registry := m.toolExecutor.Registry()
	tools := registry.All()
	schemas := make([]llm.ToolSchema, len(tools))

	for i, tool := range tools {
		// Convert ToolParameters to map[string]any for JSON schema
		inputSchema := map[string]any{
			"type":       tool.Parameters.Type,
			"properties": tool.Parameters.Properties,
		}
		if len(tool.Parameters.Required) > 0 {
			inputSchema["required"] = tool.Parameters.Required
		}

		schemas[i] = llm.ToolSchema{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: inputSchema,
		}
	}

	return schemas
}

func pollStreamCmd() tea.Msg {
	if activeStream == nil {
		return streamDoneMsg{totalTokens: 0, duration: 0, reason: "activeStream was nil"}
	}

	select {
	case resp, ok := <-activeStream.respChan:
		if !ok {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "channel closed"}
		}
		if resp.EvalCount > 0 {
			activeStream.totalTokens = resp.EvalCount
		}

		// Check for tool use in the response
		if resp.ToolUse != nil {
			// Complete tool call received
			return toolUseCompleteMsg{call: *resp.ToolUse}
		}

		// Check for tool calls - Anthropic uses stop_reason="tool_use",
		// but Ollama uses "stop" with tool_calls present
		if resp.Done && resp.Message != nil && len(resp.Message.ToolCalls) > 0 {
			// Return the first tool call (we'll handle multiple later)
			return toolUseCompleteMsg{call: resp.Message.ToolCalls[0]}
		}

		if resp.Done {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "resp.Done=true"}
		}
		return streamChunkMsg{chunk: resp}

	case err, ok := <-activeStream.errChan:
		if !ok {
			// errChan closed without error - stream ended normally, keep polling respChan
			return continueStreamMsg{}
		}
		duration := time.Since(activeStream.start)
		tokens := activeStream.totalTokens
		activeStream = nil
		if err != nil && err != context.Canceled {
			return streamErrorMsg{err: err}
		}
		return streamDoneMsg{totalTokens: tokens, duration: duration, reason: fmt.Sprintf("errChan: %v", err)}

	default:
		return continueStreamMsg{}
	}
}

func (m Model) thinkingTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

// handleToolUseComplete processes a completed tool use request from the LLM.
func (m *Model) handleToolUseComplete(call llm.ToolCall) tea.Cmd {
	if m.toolExecutor == nil {
		// No executor, return error
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    "Tool execution not available",
					IsError:    true,
				},
			}
		}
	}

	registry := m.toolExecutor.Registry()
	tool, _, ok := registry.Get(call.Name)
	if !ok {
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
					IsError:    true,
				},
			}
		}
	}

	// Check permissions
	permissions := m.toolExecutor.Permissions()
	perm := permissions.Check(call.Name, call.Arguments)

	// For tools that require approval, always ask unless session-granted
	if tool.RequiresApproval && perm == llmtools.PermissionAllow {
		if !permissions.SessionGranted(call.Name) {
			perm = llmtools.PermissionAsk
		}
	}

	switch perm {
	case llmtools.PermissionDeny:
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Tool '%s' execution denied by policy", call.Name),
					IsError:    true,
				},
			}
		}

	case llmtools.PermissionAsk:
		// Store pending call and request approval
		m.pendingToolCall = &call
		return func() tea.Msg {
			return toolApprovalRequestMsg{tool: tool, call: call}
		}

	default: // PermissionAllow
		return m.executeToolCall(call)
	}
}

// executeToolCall runs a tool and returns the result message.
func (m *Model) executeToolCall(call llm.ToolCall) tea.Cmd {
	m.executingTool = true

	// Show that we're executing the tool
	m.showToolExecution(call)

	return func() tea.Msg {
		if m.toolExecutor == nil {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    "Tool executor not configured",
					IsError:    true,
				},
			}
		}

		// Convert llm.ToolCall to llmtools.ToolCall
		toolCall := llmtools.ToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		}

		// Execute the tool
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result := m.toolExecutor.Registry().Execute(ctx, toolCall)

		return toolExecutionResultMsg{
			result: llm.ToolResult{
				ToolCallID: result.ToolCallID,
				Content:    result.Content,
				IsError:    result.IsError,
			},
		}
	}
}

// handleApprovalResponse processes the user's approval decision.
func (m *Model) handleApprovalResponse(msg toolApprovalResponseMsg) tea.Cmd {
	m.pendingToolCall = nil

	if !msg.approved {
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: msg.call.ID,
					Content:    "Tool execution denied by user",
					IsError:    true,
				},
			}
		}
	}

	// Grant session permission if requested
	if msg.grantForSession && m.toolExecutor != nil {
		m.toolExecutor.Permissions().GrantForSession(msg.call.Name)
	}

	return m.executeToolCall(msg.call)
}

// continueWithToolResults sends tool results back to the LLM to continue.
func (m *Model) continueWithToolResults() tea.Cmd {
	if len(m.toolResults) == 0 {
		return nil
	}

	results := m.toolResults
	m.toolResults = nil

	m.streaming = true
	m.streamBuf.Reset()
	m.streamStart = time.Now()

	return tea.Batch(
		m.sendMessageWithToolResults(results),
		m.thinkingTick(),
	)
}

// showToolExecution displays that a tool is being executed.
func (m *Model) showToolExecution(call llm.ToolCall) {
	var argsPreview string
	if len(call.Arguments) > 0 {
		var args map[string]any
		if err := json.Unmarshal(call.Arguments, &args); err == nil {
			// Show a brief preview of arguments
			parts := make([]string, 0, len(args))
			for k, v := range args {
				vs := fmt.Sprintf("%v", v)
				if len(vs) > 30 {
					vs = vs[:27] + "..."
				}
				parts = append(parts, fmt.Sprintf("%s=%s", k, vs))
			}
			argsPreview = strings.Join(parts, ", ")
		}
	}

	content := fmt.Sprintf("⚙️ Executing: %s", call.Name)
	if argsPreview != "" {
		content += fmt.Sprintf("\n   Args: %s", argsPreview)
	}

	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: content,
		Time:    time.Now(),
	})
	m.updateViewport()
}

// showToolResult displays the result of a tool execution.
func (m *Model) showToolResult(result llm.ToolResult) {
	status := "✓"
	if result.IsError {
		status = "✗"
	}

	// Truncate long results for display
	content := result.Content
	if len(content) > 500 {
		content = content[:500] + "\n... (truncated)"
	}

	msg := fmt.Sprintf("%s Tool result:\n%s", status, content)

	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: msg,
		Time:    time.Now(),
	})
	m.updateViewport()
}

// ApproveToolCall approves the pending tool call.
func (m *Model) ApproveToolCall(grantForSession bool) tea.Cmd {
	if m.pendingToolCall == nil {
		return nil
	}

	call := *m.pendingToolCall
	return func() tea.Msg {
		return toolApprovalResponseMsg{
			approved:        true,
			grantForSession: grantForSession,
			call:            call,
		}
	}
}

// DenyToolCall denies the pending tool call.
func (m *Model) DenyToolCall() tea.Cmd {
	if m.pendingToolCall == nil {
		return nil
	}

	call := *m.pendingToolCall
	return func() tea.Msg {
		return toolApprovalResponseMsg{
			approved: false,
			call:     call,
		}
	}
}

// ContinueAfterToolResult signals to continue the conversation after a tool result.
func (m *Model) ContinueAfterToolResult() tea.Cmd {
	return func() tea.Msg {
		return toolContinueMsg{}
	}
}

// SetSize updates the model dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resize()
}
