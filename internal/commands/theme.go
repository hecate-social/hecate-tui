package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// ThemeCmd switches or lists themes.
type ThemeCmd struct{}

func (c *ThemeCmd) Name() string        { return "theme" }
func (c *ThemeCmd) Aliases() []string   { return nil }
func (c *ThemeCmd) Description() string { return "Switch theme (/theme <name> or /theme list)" }

// SwitchThemeMsg tells the app to switch to a different theme.
type SwitchThemeMsg struct {
	Theme *theme.Theme
}

func (c *ThemeCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 || args[0] == "list" {
		return c.listThemes(ctx)
	}

	name := strings.ToLower(strings.Join(args, " "))
	themes := theme.BuiltinThemes()

	t, ok := themes[name]
	if !ok {
		return func() tea.Msg {
			var names []string
			for n := range themes {
				names = append(names, n)
			}
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Unknown theme: "+name) +
					"\n" + ctx.Styles.Subtle.Render("Available: "+strings.Join(names, ", ")),
			}
		}
	}

	return func() tea.Msg {
		return SwitchThemeMsg{Theme: t}
	}
}

func (c *ThemeCmd) listThemes(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		themes := theme.BuiltinThemes()

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Available Themes"))
		b.WriteString("\n\n")

		for name, t := range themes {
			marker := "  "
			if t.Name == ctx.Theme.Name {
				marker = "● "
			}
			b.WriteString(s.Bold.Render(marker + name))
			b.WriteString(s.Subtle.Render("  " + t.Name))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Use /theme <name> to switch"))

		return InjectSystemMsg{Content: b.String()}
	}
}
