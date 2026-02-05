package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/modes"
)

// PairCmd triggers Pair mode.
type PairCmd struct{}

func (c *PairCmd) Name() string        { return "pair" }
func (c *PairCmd) Aliases() []string   { return []string{"p"} }
func (c *PairCmd) Description() string { return "Start realm pairing flow" }

func (c *PairCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetModeMsg{Mode: int(modes.Pair)}
	}
}
