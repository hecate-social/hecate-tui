package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// StatusCmd shows daemon status as an inline card.
type StatusCmd struct{}

func (c *StatusCmd) Name() string        { return "status" }
func (c *StatusCmd) Aliases() []string   { return nil }
func (c *StatusCmd) Description() string { return "Show daemon status" }

func (c *StatusCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		health, err := ctx.Client.GetHealth()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get status: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Status"))
		b.WriteString("\n\n")

		// Status indicator
		statusIcon := "●"
		statusStyle := s.StatusOK
		switch health.Status {
		case "degraded":
			statusStyle = s.StatusWarning
		case "error", "unhealthy":
			statusStyle = s.StatusError
		}
		b.WriteString(s.CardLabel.Render("  Daemon:"))
		b.WriteString(statusStyle.Render(statusIcon + " " + health.Status))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("  Version:"))
		b.WriteString(s.CardValue.Render(health.Version))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("  Uptime:"))
		b.WriteString(s.CardValue.Render(formatUptime(health.UptimeSeconds)))
		b.WriteString("\n")

		// Mesh status
		identity, identErr := ctx.Client.GetIdentity()
		if identErr == nil && identity != nil {
			b.WriteString(s.CardLabel.Render("  Identity:"))
			b.WriteString(s.CardValue.Render(identity.Identity))
			b.WriteString("\n")
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func formatUptime(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
