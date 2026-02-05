package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/llm"
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
	streamBuf     strings.Builder
	thinkingFrame int

	// Stats
	lastTokenCount int
	lastDuration   time.Duration
	lastSpeed      float64
	streamStart    time.Time

	// Error
	err error

	// Input visibility (controlled by mode)
	inputVisible   bool
	commandVisible bool
	commandInput   string

	// System prompt
	systemPrompt string
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
}

type streamErrorMsg struct {
	err error
}

type thinkingTickMsg struct{}

type continueStreamMsg struct{}

// New creates a new chat model.
func New(c *client.Client, t *theme.Theme, s *theme.Styles) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(1, 2)

	return Model{
		client:   c,
		theme:    t,
		styles:   s,
		viewport: vp,
		input:    ta,
		messages: []Message{},
	}
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
		return m, nil

	case streamChunkMsg:
		if msg.chunk.Message != nil && msg.chunk.Message.Content != "" {
			m.streamBuf.WriteString(msg.chunk.Message.Content)
			m.updateStreamingMessage()
		}
		return m, func() tea.Msg { return pollStreamCmd() }

	case continueStreamMsg:
		return m, tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
			return pollStreamCmd()
		})

	case streamDoneMsg:
		m.streaming = false
		m.lastTokenCount = msg.totalTokens
		m.lastDuration = msg.duration
		if msg.duration > 0 {
			m.lastSpeed = float64(msg.totalTokens) / msg.duration.Seconds()
		}
		if m.streamBuf.Len() > 0 {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: m.streamBuf.String(),
				Time:    time.Now(),
			})
		}
		m.streamBuf.Reset()
		m.updateViewport()
		return m, nil

	case streamErrorMsg:
		m.streaming = false
		m.err = msg.err
		m.streamBuf.Reset()
		return m, nil

	case thinkingTickMsg:
		if m.streaming {
			m.thinkingFrame = (m.thinkingFrame + 1) % len(ThinkingFrames)
			return m, m.thinkingTick()
		}
		return m, nil
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

// IsStreaming returns whether a response is being streamed.
func (m Model) IsStreaming() bool {
	return m.streaming
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
	inputHeight := 0
	if m.inputVisible {
		inputHeight = 5
	}
	statsHeight := 1
	padding := 4

	vpHeight := m.height - inputHeight - statsHeight - padding
	if vpHeight < 5 {
		vpHeight = 5
	}

	vpWidth := m.width - 4
	if vpWidth < 40 {
		vpWidth = 40
	}

	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.input.SetWidth(vpWidth - 2)

	m.updateViewport()
}

func (m *Model) updateViewport() {
	content := m.renderMessages()
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
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
			bubble := m.styles.UserBubble.Width(bubbleWidth).Render(msg.Content)
			aligned := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Right, bubble)
			parts = append(parts, label+"\n"+aligned)

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

		req := llm.ChatRequest{
			Model:    modelName,
			Messages: llmMsgs,
			Stream:   true,
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

func pollStreamCmd() tea.Msg {
	if activeStream == nil {
		return streamDoneMsg{totalTokens: 0, duration: 0}
	}

	select {
	case resp, ok := <-activeStream.respChan:
		if !ok {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration}
		}
		if resp.EvalCount > 0 {
			activeStream.totalTokens = resp.EvalCount
		}
		if resp.Done {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration}
		}
		return streamChunkMsg{chunk: resp}

	case err := <-activeStream.errChan:
		duration := time.Since(activeStream.start)
		tokens := activeStream.totalTokens
		activeStream = nil
		if err != nil && err != context.Canceled {
			return streamErrorMsg{err: err}
		}
		return streamDoneMsg{totalTokens: tokens, duration: duration}

	default:
		return continueStreamMsg{}
	}
}

func (m Model) thinkingTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

// SetSize updates the model dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resize()
}
