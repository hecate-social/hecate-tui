package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// EditCmd opens the built-in editor.
type EditCmd struct{}

func (c *EditCmd) Name() string        { return "edit" }
func (c *EditCmd) Aliases() []string   { return []string{"e"} }
func (c *EditCmd) Description() string { return "Open built-in editor (/edit [file])" }

// EditFileMsg tells the app to open a file in the editor.
type EditFileMsg struct {
	Path string // empty = scratch buffer
}

func (c *EditCmd) Execute(args []string, ctx *Context) tea.Cmd {
	path := ""
	if len(args) > 0 {
		path = strings.Join(args, " ")
	}

	return func() tea.Msg {
		return EditFileMsg{Path: path}
	}
}
