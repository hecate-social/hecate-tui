// Package studio defines the interface for TUI studios — self-contained workspaces
// within the Hecate TUI. Each studio owns its content area, mode management,
// key handling, and commands.
package studio

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/config"
	"github.com/hecate-social/hecate-tui/internal/factbus"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Studio is the interface every studio workspace implements.
// The shell (App) manages the frame (top bar, status bar, command input);
// studios own the content area between them.
type Studio interface {
	// Identity
	Name() string      // "LLM", "Development", etc.
	ShortName() string // Short tab label: "LLM", "Dev", "Ops", "Social", "Arcade"
	Icon() string      // Emoji for tab bar

	// Lifecycle — standard Bubble Tea model methods
	Init() tea.Cmd
	Update(msg tea.Msg) (Studio, tea.Cmd)
	View() string

	// Size — called when the terminal resizes
	SetSize(width, height int)

	// Mode — the studio's current input mode (Normal, Insert, Browse, etc.)
	// The shell reads this to decide key routing.
	Mode() modes.Mode

	// Hints — contextual keybinding hints for the status bar
	Hints() string

	// StatusInfo — data the shell reads to populate the shared status bar
	StatusInfo() StatusInfo

	// Commands — studio-specific slash commands merged with global commands
	Commands() []commands.Command

	// Focus management — for state preservation when switching studios
	Focused() bool
	SetFocused(focused bool)
}

// StatusInfo is a data struct the shell reads from the active studio
// to populate the shared status bar.
type StatusInfo struct {
	// Left side
	ModelName     string // LLM model name (LLM/Dev studios)
	ModelProvider string // "ollama", "openai", "anthropic", etc.
	ModelStatus   string // "ready", "loading", "error"
	ModelError    string // error message when ModelStatus is "error"
	ChannelName   string // IRC channel (Social studio)
	GameName      string // current game (Arcade studio)

	// Right side
	InputLen      int // character count for Insert mode
	SessionTokens int // cumulative tokens for session
	OnlineCount   int // channel members / players online
}

// Context holds shared resources passed to studios at construction time.
// Studios hold a reference — this is passed once, not on every Update.
type Context struct {
	Client  *client.Client
	Theme   *theme.Theme
	Styles  *theme.Styles
	Config  config.Config
	FactBus *factbus.Connection
}

// SwitchStudioMsg tells the shell to switch to a different studio by index.
type SwitchStudioMsg struct {
	Index int
}
