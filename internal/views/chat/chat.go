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
)

// Model is the chat view model
type Model struct {
	client   *client.Client
	viewport viewport.Model
	input    textarea.Model

	// Dimensions
	width  int
	height int

	// State
	messages      []llm.Message
	models        []llm.Model
	activeModel   int
	streaming     bool
	streamBuf     strings.Builder
	thinkingFrame int

	// Stats
	lastTokenCount int
	lastDuration   time.Duration
	lastSpeed      float64

	// Error
	err error

	// Focus
	focused bool
}

// New creates a new chat view
func New(c *client.Client) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Enter to send)"
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Purple)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Gray600)
	ta.Focus()

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Padding(1, 2)

	return Model{
		client:   c,
		viewport: vp,
		input:    ta,
		messages: []llm.Message{},
		focused:  true,
	}
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

// Init initializes the chat view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchModels,
		textarea.Blink,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.streaming {
			// Allow Escape to cancel streaming
			if msg.String() == "esc" || msg.String() == "ctrl+c" {
				if activeStream != nil && activeStream.cancel != nil {
					activeStream.cancel()
					activeStream = nil
				}
				m.streaming = false
				// Keep what we have so far
				if m.streamBuf.Len() > 0 {
					m.messages = append(m.messages, llm.Message{
						Role:    llm.RoleAssistant,
						Content: m.streamBuf.String() + " [cancelled]",
					})
					m.streamBuf.Reset()
				}
				m.updateViewport()
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Cycle through models
			if len(m.models) > 0 {
				m.activeModel = (m.activeModel + 1) % len(m.models)
			}
			return m, nil

		case "shift+tab":
			// Cycle models backwards
			if len(m.models) > 0 {
				m.activeModel = (m.activeModel - 1 + len(m.models)) % len(m.models)
			}
			return m, nil

		case "enter":
			// Send message (unless shift is held for newline)
			content := strings.TrimSpace(m.input.Value())
			if content != "" && !m.streaming {
				m.messages = append(m.messages, llm.Message{
					Role:    llm.RoleUser,
					Content: content,
				})
				m.input.Reset()
				m.streaming = true
				m.streamBuf.Reset()
				m.err = nil
				m.updateViewport()
				return m, tea.Batch(
					m.sendMessage(),
					m.thinkingTick(),
				)
			}
			return m, nil

		case "ctrl+l":
			// Clear chat
			m.messages = []llm.Message{}
			m.updateViewport()
			return m, nil

		case "pgup", "pgdown", "up", "down":
			// Pass to viewport for scrolling
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

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
		// Continue polling for more chunks
		return m, func() tea.Msg { return pollStreamCmd() }

	case continueStreamMsg:
		// Short delay then poll again
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
		// Finalize the assistant message
		if m.streamBuf.Len() > 0 {
			m.messages = append(m.messages, llm.Message{
				Role:    llm.RoleAssistant,
				Content: m.streamBuf.String(),
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

	// Update textarea
	if !m.streaming {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the chat view
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Model selector
	b.WriteString(m.renderModelSelector())
	b.WriteString("\n")

	// Chat area
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Streaming indicator or stats
	if m.streaming {
		b.WriteString(m.renderStreaming())
	} else if m.lastTokenCount > 0 {
		b.WriteString(m.renderStats())
	}
	b.WriteString("\n")

	// Input area
	b.WriteString(m.input.View())
	b.WriteString("\n")

	// Help
	b.WriteString(m.renderHelp())

	// Error
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("⚠ " + m.err.Error()))
	}

	return b.String()
}

func (m Model) renderModelSelector() string {
	var models []string

	title := HeaderStyle.Render("🗝️ Hecate Chat")

	if len(m.models) == 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left,
			title,
			ModelSelectorStyle.Render("  No models available"),
		)
	}

	for i, model := range m.models {
		name := model.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		if i == m.activeModel {
			models = append(models, ModelActiveStyle.Render("● "+name))
		} else {
			models = append(models, ModelInactiveStyle.Render("○ "+name))
		}
	}

	selector := lipgloss.JoinHorizontal(lipgloss.Left, models...)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		title,
		"  ",
		selector,
	)
}

func (m Model) renderStreaming() string {
	frame := ThinkingFrames[m.thinkingFrame]
	sparkle := Sparkles[m.thinkingFrame%len(Sparkles)]
	return StreamingStyle.Render(sparkle + " " + frame + " " + sparkle)
}

func (m Model) renderStats() string {
	return StatsStyle.Render(
		FormatTokens(m.lastTokenCount) + "  •  " + FormatSpeed(m.lastSpeed),
	)
}

func (m Model) renderHelp() string {
	help := "Enter: send • Tab: model • Ctrl+L: clear • ↑↓: scroll"
	if m.streaming {
		help = "Esc: cancel"
	}
	return HelpStyle.Render(help)
}

func (m *Model) resize() {
	headerHeight := 2
	inputHeight := 5
	statsHeight := 2
	helpHeight := 1
	padding := 2

	vpHeight := m.height - headerHeight - inputHeight - statsHeight - helpHeight - padding
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
	// Add streaming content
	if m.streamBuf.Len() > 0 {
		content += "\n" + AssistantLabelStyle.Render("✦ Assistant") + "\n"
		bubble := AssistantBubbleStyle.Width(m.viewport.Width - 8).Render(m.streamBuf.String() + "▊")
		content += bubble
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		welcome := WelcomeStyle.Render(WelcomeArt())
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

	for _, msg := range m.messages {
		switch msg.Role {
		case llm.RoleUser:
			label := UserLabelStyle.Render("You")
			bubble := UserBubbleStyle.Width(bubbleWidth).Render(msg.Content)
			// Right-align user messages
			aligned := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Right, bubble)
			parts = append(parts, label+"\n"+aligned)

		case llm.RoleAssistant:
			label := AssistantLabelStyle.Render("✦ Assistant")
			bubble := AssistantBubbleStyle.Width(bubbleWidth).Render(msg.Content)
			parts = append(parts, label+"\n"+bubble)

		case llm.RoleSystem:
			bubble := SystemBubbleStyle.Width(bubbleWidth).Render(msg.Content)
			parts = append(parts, bubble)
		}
	}

	return strings.Join(parts, "\n\n")
}

// Commands
func (m Model) fetchModels() tea.Msg {
	models, err := m.client.ListModels()
	return modelsMsg{models: models, err: err}
}

// streamState holds state for an active stream
type streamState struct {
	ctx        context.Context
	cancel     context.CancelFunc
	respChan   <-chan llm.ChatResponse
	errChan    <-chan error
	start      time.Time
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

		req := llm.ChatRequest{
			Model:    modelName,
			Messages: m.messages,
			Stream:   true,
		}

		start := time.Now()
		respChan, errChan := m.client.ChatStream(ctx, req)

		// Store stream state globally for polling
		activeStream = &streamState{
			ctx:      ctx,
			cancel:   cancel,
			respChan: respChan,
			errChan:  errChan,
			start:    start,
		}

		// Start polling for stream chunks
		return pollStreamCmd()
	}
}

// pollStreamCmd creates a command that polls for the next stream chunk
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
		// No data yet, continue polling
		return continueStreamMsg{}
	}
}

// continueStreamMsg signals to continue polling
type continueStreamMsg struct{}

func (m Model) thinkingTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

// SetSize updates the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resize()
}

// Focus sets focus on the chat input
func (m *Model) Focus() {
	m.focused = true
	m.input.Focus()
}

// Blur removes focus from the chat input
func (m *Model) Blur() {
	m.focused = false
	m.input.Blur()
}

// Name returns the tab label
func (m Model) Name() string {
	return "Chat"
}

// ShortHelp returns help text for the status bar
func (m Model) ShortHelp() string {
	if m.streaming {
		return "Esc: cancel streaming"
	}
	return "Enter: send • Tab: model • Ctrl+L: clear • ↑↓: scroll"
}

// IsStreaming returns whether the chat is currently streaming a response
func (m Model) IsStreaming() bool {
	return m.streaming
}
