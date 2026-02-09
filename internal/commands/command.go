package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Command is the interface all slash commands implement.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Execute(args []string, ctx *Context) tea.Cmd
}

// Completable is an optional interface for commands that support argument completion.
type Completable interface {
	// Complete returns completion suggestions for the given arguments.
	// args contains the arguments typed so far (may be empty).
	// Returns a list of suggestions to replace the last argument (or add if args is empty).
	Complete(args []string, ctx *Context) []string
}

// Context provides commands access to the app's resources.
type Context struct {
	Client     client.DaemonClient
	SocketPath string // Unix socket path (if connected via socket)
	HTTPUrl    string // HTTP URL (if connected via TCP)
	Theme      *theme.Theme
	Styles     *theme.Styles
	Width      int
	Height     int

	// Callbacks for commands that need to affect app state
	SetMode    func(mode int) // triggers mode change (use app.Mode* constants via int)
	InjectChat func(msg ChatMessage)

	// Chat access
	GetMessages     func() []ChatExportMsg
	GetSystemPrompt func() string
	SetSystemPrompt func(prompt string)

	// Tool system access
	GetToolExecutor func() *llmtools.Executor
	ToolsEnabled    func() bool

	// Config access for personality/roles
	GetActiveRole    func() string
	SetActiveRole    func(role string) error
	GetRoleNames     func() []string
	RebuildPrompt    func() string // rebuilds system prompt from config

	// ALC context access
	GetALCContext func() *alc.State
}

// Ctx returns a background context. Used for tool execution.
func (c *Context) Ctx() context.Context {
	return context.Background()
}

// ChatExportMsg represents a message for export purposes.
type ChatExportMsg struct {
	Role    string
	Content string
	Time    string
}

// ChatMessage represents a message to inject into the chat stream.
// Commands produce these as output — they appear as system messages.
type ChatMessage struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// SystemMessage is a convenience for creating a system chat message.
func SystemMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "system",
		Content: content,
	}
}

// InjectSystemMsg is a tea.Msg that tells the app to add a system message to chat.
type InjectSystemMsg struct {
	Content string
}

// SetModeMsg is a tea.Msg that tells the app to switch modes.
type SetModeMsg struct {
	Mode int
}

// NewConversationMsg tells the app to start a new conversation.
type NewConversationMsg struct{}

// LoadConversationMsg tells the app to load a specific conversation.
type LoadConversationMsg struct {
	ID string
}

// SwitchRoleMsg tells the app to switch to a different ALC role.
type SwitchRoleMsg struct {
	Role string // dna, anp, tni, dno
}

// SetALCContextMsg tells the app to switch ALC context (Chat/Torch/Cartwheel).
type SetALCContextMsg struct {
	Context   alc.Context
	Torch     *alc.TorchInfo
	Cartwheel *alc.CartwheelInfo
	Source    string // "manual", "git", "config"
}

// ShowFormMsg tells the app to display a form overlay.
type ShowFormMsg struct {
	FormType string // "torch_init", "cartwheel_init", etc.
}
