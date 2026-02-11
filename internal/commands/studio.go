package commands

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SwitchStudioMsg tells the app to switch to a different studio by index.
type SwitchStudioMsg struct {
	Index int
}

// StudioCmd switches between studios.
type StudioCmd struct{}

func (c *StudioCmd) Name() string        { return "studio" }
func (c *StudioCmd) Aliases() []string   { return []string{"s"} }
func (c *StudioCmd) Description() string { return "Switch studio (/studio <name|number>)" }

var studioNames = []struct {
	Index int
	Name  string
	Short string
}{
	{0, "llm", "LLM"},
	{1, "dev", "Dev"},
	{2, "ops", "Ops"},
	{3, "social", "Social"},
	{4, "arcade", "Arcade"},
}

func (c *StudioCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.listStudios(ctx)
	}

	target := strings.ToLower(args[0])

	// Try numeric index first (1-based)
	if n, err := strconv.Atoi(target); err == nil && n >= 1 && n <= len(studioNames) {
		return func() tea.Msg {
			return SwitchStudioMsg{Index: n - 1}
		}
	}

	// Match by name
	for _, s := range studioNames {
		if strings.EqualFold(s.Name, target) || strings.EqualFold(s.Short, target) {
			idx := s.Index
			return func() tea.Msg {
				return SwitchStudioMsg{Index: idx}
			}
		}
	}

	return func() tea.Msg {
		return InjectSystemMsg{
			Content: ctx.Styles.Error.Render("Unknown studio: " + target) +
				"\n" + ctx.Styles.Subtle.Render("Use /studio to list available studios."),
		}
	}
}

func (c *StudioCmd) listStudios(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Studios"))
		b.WriteString("\n\n")
		for _, st := range studioNames {
			b.WriteString(s.Bold.Render(fmt.Sprintf("  %d. %s", st.Index+1, st.Short)))
			b.WriteString(s.Subtle.Render("  " + st.Name))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Use /studio <name|number> to switch"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Or Ctrl+1-5 in Normal mode"))
		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *StudioCmd) Complete(args []string, ctx *Context) []string {
	if len(args) > 1 {
		return nil
	}
	prefix := ""
	if len(args) == 1 {
		prefix = strings.ToLower(args[0])
	}
	var matches []string
	for _, s := range studioNames {
		if strings.HasPrefix(strings.ToLower(s.Name), prefix) {
			matches = append(matches, s.Name)
		}
	}
	return matches
}
