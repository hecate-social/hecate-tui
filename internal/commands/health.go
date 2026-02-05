package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// HealthCmd shows a quick daemon health check.
type HealthCmd struct{}

func (c *HealthCmd) Name() string        { return "health" }
func (c *HealthCmd) Aliases() []string   { return nil }
func (c *HealthCmd) Description() string { return "Quick daemon health check" }

func (c *HealthCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		health, err := ctx.Client.GetHealth()
		if err != nil {
			return InjectSystemMsg{Content: s.StatusError.Render("● Daemon unreachable: ") + s.Error.Render(err.Error())}
		}

		var b strings.Builder
		switch health.Status {
		case "healthy", "ok":
			b.WriteString(s.StatusOK.Render("● Healthy"))
		case "degraded":
			b.WriteString(s.StatusWarning.Render("● Degraded"))
		default:
			b.WriteString(s.StatusError.Render("● " + health.Status))
		}

		b.WriteString(s.Subtle.Render("  v" + health.Version))
		b.WriteString(s.Subtle.Render("  up " + formatUptime(health.UptimeSeconds)))

		// LLM health
		llmHealth, llmErr := ctx.Client.GetLLMHealth()
		if llmErr == nil && llmHealth != nil {
			b.WriteString(s.Subtle.Render("  LLM: "))
			switch llmHealth.Status {
			case "ok", "healthy":
				b.WriteString(s.StatusOK.Render("●"))
			default:
				b.WriteString(s.StatusError.Render("●"))
			}
		}

		return InjectSystemMsg{Content: b.String()}
	}
}
