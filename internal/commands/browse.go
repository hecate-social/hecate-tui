package commands

import tea "github.com/charmbracelet/bubbletea"

// BrowseCmd enters Browse mode.
type BrowseCmd struct{}

func (c *BrowseCmd) Name() string        { return "browse" }
func (c *BrowseCmd) Aliases() []string   { return []string{"b"} }
func (c *BrowseCmd) Description() string { return "Browse capabilities on the mesh" }

func (c *BrowseCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetModeMsg{Mode: 3} // modes.Browse = 3
	}
}
