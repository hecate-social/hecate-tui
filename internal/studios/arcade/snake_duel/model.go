package snake_duel

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/studio"
)

// Model is the Bubble Tea model for the Snake Duel sub-app.
type Model struct {
	ctx *studio.Context

	// Layout
	width  int
	height int

	// Game state from SSE
	state   GameState
	matchID string

	// Match configuration (adjustable before starting)
	af1    int
	af2    int
	tickMs int

	// Lifecycle phase: idle, connecting, playing, finished
	phase string

	// SSE connection (nil when not playing)
	stream *MatchStream

	// Signal to parent to go back to arcade home
	wantsBack bool

	// Error from last operation
	err error
}

// New creates a new Snake Duel model.
func New(ctx *studio.Context) *Model {
	return &Model{
		ctx:    ctx,
		af1:    50,
		af2:    50,
		tickMs: 100,
		phase:  "idle",
	}
}

// Init returns the initial command.
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize updates the layout dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// WantsBack returns true if the model wants to return to the arcade home.
func (m *Model) WantsBack() bool {
	return m.wantsBack
}

// ClearWantsBack resets the back signal.
func (m *Model) ClearWantsBack() {
	m.wantsBack = false
}

// View renders the current state.
func (m *Model) View() string {
	return m.view()
}

// Hints returns contextual keybinding hints.
func (m *Model) Hints() string {
	switch m.phase {
	case "idle":
		return "n:new  esc:back  +/-:speed  [/]:P1 AF  {/}:P2 AF"
	case "connecting":
		return "esc:cancel"
	case "playing":
		return "esc:stop"
	case "finished":
		return "n:new  esc:back  +/-:speed"
	default:
		return ""
	}
}

// StatusInfo returns data for the shared status bar.
func (m *Model) StatusInfo() studio.StatusInfo {
	return studio.StatusInfo{
		GameName: "Snake Duel",
	}
}

// Close tears down the SSE stream.
func (m *Model) Close() {
	if m.stream != nil {
		m.stream.Close()
		m.stream = nil
	}
}

// Update handles all Bubble Tea messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case MatchStartedMsg:
		m.matchID = msg.MatchID
		m.phase = "playing"
		m.stream = NewMatchStream(
			m.ctx.Client.SocketPath(),
			m.ctx.Client.BaseURL(),
		)
		return m.stream.Connect(m.matchID)

	case MatchStartFailedMsg:
		m.phase = "idle"
		m.err = msg.Err
		return nil

	case MatchStateMsg:
		m.state = msg.State
		if msg.State.Status == "finished" {
			m.phase = "finished"
			if m.stream != nil {
				m.stream.Close()
				m.stream = nil
			}
			return nil
		}
		return m.pollStream()

	case MatchContinueMsg:
		return m.pollStream()

	case MatchDoneMsg:
		if m.phase == "playing" {
			m.phase = "finished"
		}
		m.stream = nil
		return nil

	case MatchErrorMsg:
		m.phase = "idle"
		m.err = msg.Err
		if m.stream != nil {
			m.stream.Close()
			m.stream = nil
		}
		return nil

	case pollTickMsg:
		return m.pollStream()
	}

	return nil
}

// startNewMatch initiates a match via the daemon API.
func (m *Model) startNewMatch() tea.Cmd {
	m.phase = "connecting"
	m.err = nil
	return StartMatch(
		m.ctx.Client.SocketPath(),
		m.ctx.Client.BaseURL(),
		m.af1, m.af2, m.tickMs,
	)
}

// pollStream returns a command to poll the SSE stream after a small delay.
func (m *Model) pollStream() tea.Cmd {
	if m.stream == nil {
		return nil
	}
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return pollTickMsg{}
	})
}

// pollTickMsg triggers a stream poll after a small delay.
type pollTickMsg struct{}
