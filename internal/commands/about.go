package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/version"
)

// AboutCmd shows version and project info.
type AboutCmd struct{}

func (c *AboutCmd) Name() string        { return "about" }
func (c *AboutCmd) Aliases() []string   { return []string{"version"} }
func (c *AboutCmd) Description() string { return "Show version and project info" }

func (c *AboutCmd) Execute(args []string, ctx *Context) tea.Cmd {
	info := fmt.Sprintf(`Hecate TUI v%s

Terminal interface for Macula Hecate Daemon

Repository: github.com/hecate-social/hecate-tui
Support:    %s

Press 'i' to chat, '/' for commands, '?' for help`, version.Version, version.DonateURL)

	return func() tea.Msg {
		return InjectSystemMsg{Content: info}
	}
}
