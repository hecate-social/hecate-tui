package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// HelpCmd shows available commands.
type HelpCmd struct {
	registry *Registry
}

func (c *HelpCmd) Name() string        { return "help" }
func (c *HelpCmd) Aliases() []string   { return []string{"h", "?"} }
func (c *HelpCmd) Description() string { return "Show available commands" }

func (c *HelpCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		var b strings.Builder
		s := ctx.Styles

		b.WriteString(s.CardTitle.Render("Available Commands"))
		b.WriteString("\n\n")

		for _, cmd := range c.registry.List() {
			name := fmt.Sprintf("  /%s", cmd.Name())
			aliases := ""
			if len(cmd.Aliases()) > 0 {
				aliases = " (" + strings.Join(cmd.Aliases(), ", ") + ")"
			}
			b.WriteString(s.Bold.Render(name))
			b.WriteString(s.Subtle.Render(aliases))
			b.WriteString(s.Subtle.Render("  " + cmd.Description()))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Type / or : to enter command mode"))

		return InjectSystemMsg{Content: b.String()}
	}
}
