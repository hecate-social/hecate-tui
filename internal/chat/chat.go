package chat

import (
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
	client   client.DaemonClient
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

// ExportMsg is a message suitable for export (no internal state).
type ExportMsg struct {
	Role    string
	Content string
	Time    string
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
func New(c client.DaemonClient, t *theme.Theme, s *theme.Styles) Model {
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
	// NOTE: WindowSizeMsg is handled by App via SetSize() — do NOT handle here,
	// as it would override the calculated chat area height with the full terminal height.

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
			// Update the chat area to show the new animation frame
			if m.streaming {
				m.updateStreamingMessage()
			}
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

// -- State queries --

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

// SessionTokenCount returns the cumulative token count for this session.
func (m Model) SessionTokenCount() int {
	return m.sessionTokenCount
}

// -- Scroll API --

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

// -- Messages API --

// Messages returns the current message list.
func (m Model) Messages() []Message {
	return m.messages
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

// LoadMessages replaces all messages (for loading saved conversations).
func (m *Model) LoadMessages(msgs []Message) {
	m.messages = msgs
	m.updateViewport()
}

// ClearMessages removes all chat messages.
func (m *Model) ClearMessages() {
	m.messages = []Message{}
	m.lastTokenCount = 0
	m.lastSpeed = 0
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

// -- System prompt --

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

// -- Streaming control --

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
