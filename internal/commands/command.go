package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Command is the interface all slash commands implement.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Execute(args []string, ctx *Context) tea.Cmd
}

// Context provides commands access to the app's resources.
type Context struct {
	Client *client.Client
	Theme  *theme.Theme
	Styles *theme.Styles
	Width  int
	Height int

	// Callbacks for commands that need to affect app state
	SetMode    func(mode int) // triggers mode change (use app.Mode* constants via int)
	InjectChat func(msg ChatMessage)

	// Chat access
	GetMessages    func() []ChatExportMsg
	GetSystemPrompt func() string
	SetSystemPrompt func(prompt string)
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
