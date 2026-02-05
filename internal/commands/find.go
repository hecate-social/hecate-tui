package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FindCmd searches through current chat messages.
type FindCmd struct{}

func (c *FindCmd) Name() string        { return "find" }
func (c *FindCmd) Aliases() []string   { return []string{"search", "f"} }
func (c *FindCmd) Description() string { return "Search chat messages (/find <term>)" }

func (c *FindCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: "Usage: /find <search term>"}
		}
	}

	term := strings.Join(args, " ")

	return func() tea.Msg {
		s := ctx.Styles
		msgs := ctx.GetMessages()
		termLower := strings.ToLower(term)

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Search: " + term))
		b.WriteString("\n\n")

		matches := 0
		for _, msg := range msgs {
			contentLower := strings.ToLower(msg.Content)
			if !strings.Contains(contentLower, termLower) {
				continue
			}
			matches++

			// Role label
			role := msg.Role
			switch role {
			case "user":
				b.WriteString(s.Bold.Render("You"))
			case "assistant":
				b.WriteString(s.Bold.Render("Hecate"))
			default:
				continue // skip system messages in search
			}

			if msg.Time != "" {
				b.WriteString(s.Subtle.Render(" " + msg.Time))
			}
			b.WriteString("\n")

			// Show context around match
			lines := strings.Split(msg.Content, "\n")
			shown := 0
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), termLower) {
					preview := line
					if len(preview) > 100 {
						preview = preview[:97] + "..."
					}
					b.WriteString(s.CardValue.Render("  " + preview))
					b.WriteString("\n")
					shown++
					if shown >= 3 {
						break
					}
				}
			}
			b.WriteString("\n")
		}

		if matches == 0 {
			b.WriteString(s.Subtle.Render("No matches found."))
		} else {
			b.WriteString(s.Subtle.Render(itoa(matches) + " message(s) matched"))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}
