package commands

import tea "github.com/charmbracelet/bubbletea"

// ClearCmd clears the chat history.
type ClearCmd struct{}

func (c *ClearCmd) Name() string        { return "clear" }
func (c *ClearCmd) Aliases() []string   { return []string{"cls"} }
func (c *ClearCmd) Description() string { return "Clear chat history" }

// ClearChatMsg tells the app to clear all chat messages.
type ClearChatMsg struct{}

func (c *ClearCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return ClearChatMsg{}
	}
}
