package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SystemCmd sets or views the LLM system prompt.
type SystemCmd struct{}

func (c *SystemCmd) Name() string        { return "system" }
func (c *SystemCmd) Aliases() []string   { return []string{"sys"} }
func (c *SystemCmd) Description() string { return "Set/view LLM system prompt (/system [prompt])" }

func (c *SystemCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) == 0 {
			// Show current system prompt
			current := ctx.GetSystemPrompt()
			if current == "" {
				return InjectSystemMsg{
					Content: s.Subtle.Render("No system prompt set.") + "\n" +
						s.Subtle.Render("Use /system <prompt> to set one."),
				}
			}

			var b strings.Builder
			b.WriteString(s.CardTitle.Render("System Prompt"))
			b.WriteString("\n\n")
			b.WriteString(s.CardValue.Render(current))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("Use /system clear to remove."))

			return InjectSystemMsg{Content: b.String()}
		}

		// Special: clear system prompt
		if len(args) == 1 && args[0] == "clear" {
			ctx.SetSystemPrompt("")
			return InjectSystemMsg{
				Content: s.StatusOK.Render("System prompt cleared."),
			}
		}

		// Set system prompt
		prompt := strings.Join(args, " ")
		ctx.SetSystemPrompt(prompt)

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("System prompt set:"))
		b.WriteString("\n")

		display := prompt
		if len(display) > 120 {
			display = display[:117] + "..."
		}
		b.WriteString(s.Subtle.Render("  " + display))

		return InjectSystemMsg{Content: b.String()}
	}
}
