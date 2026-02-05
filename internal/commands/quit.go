package commands

import tea "github.com/charmbracelet/bubbletea"

// QuitCmd exits the TUI.
type QuitCmd struct{}

func (c *QuitCmd) Name() string        { return "quit" }
func (c *QuitCmd) Aliases() []string   { return []string{"q", "exit"} }
func (c *QuitCmd) Description() string { return "Quit Hecate" }

func (c *QuitCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return tea.Quit
}
