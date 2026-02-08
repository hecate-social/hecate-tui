package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// AgentsCmd shows active agents in the swarm.
type AgentsCmd struct{}

func (c *AgentsCmd) Name() string        { return "agents" }
func (c *AgentsCmd) Aliases() []string   { return []string{"agent", "ag"} }
func (c *AgentsCmd) Description() string { return "View active agents in the swarm" }

func (c *AgentsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) > 0 {
		return c.showAgent(args[0], ctx)
	}
	return c.listAgents(ctx)
}

func (c *AgentsCmd) listAgents(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		agents, err := ctx.Client.ListAgents()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list agents: " + err.Error())}
		}

		if len(agents) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No active agents.")}
		}

		// Calculate column widths
		maxID := 2 // "ID"
		maxType := 4 // "Type"
		maxRole := 4 // "Role"
		for _, a := range agents {
			idLen := len(truncateID(a.AgentID))
			if idLen > maxID {
				maxID = idLen
			}
			if len(a.AgentType) > maxType {
				maxType = len(a.AgentType)
			}
			if len(a.Role) > maxRole {
				maxRole = len(a.Role)
			}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Active Agents"))
		b.WriteString("\n\n")

		// Header
		header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			maxID, "ID",
			maxType, "Type",
			maxRole, "Role",
			"Status")
		b.WriteString(s.Subtle.Render(header))
		b.WriteString("\n")

		// Separator
		b.WriteString(s.Subtle.Render("  " + strings.Repeat("-", maxID+maxType+maxRole+15)))
		b.WriteString("\n")

		// Rows
		for _, a := range agents {
			shortID := truncateID(a.AgentID)
			agentType := a.AgentType
			if agentType == "" {
				agentType = "-"
			}
			role := strings.ToUpper(a.Role)
			if role == "" {
				role = "-"
			}

			statusText, statusStyle := formatAgentStatus(a.Status, s)

			b.WriteString("  ")
			b.WriteString(s.Bold.Render(fmt.Sprintf("%-*s", maxID, shortID)))
			b.WriteString("  ")
			b.WriteString(s.CardValue.Render(fmt.Sprintf("%-*s", maxType, agentType)))
			b.WriteString("  ")
			b.WriteString(s.Subtle.Render(fmt.Sprintf("%-*s", maxRole, role)))
			b.WriteString("  ")
			b.WriteString(statusStyle.Render(statusText))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Use /agents <id> for details"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *AgentsCmd) showAgent(agentID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		agent, err := ctx.Client.GetAgent(agentID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get agent: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Agent Details"))
		b.WriteString("\n\n")

		b.WriteString(s.CardLabel.Render("       ID: "))
		b.WriteString(s.CardValue.Render(agent.AgentID))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("    Torch: "))
		if agent.TorchID != "" {
			b.WriteString(s.CardValue.Render(agent.TorchID))
		} else {
			b.WriteString(s.Subtle.Render("-"))
		}
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("     Type: "))
		if agent.AgentType != "" {
			b.WriteString(s.CardValue.Render(agent.AgentType))
		} else {
			b.WriteString(s.Subtle.Render("-"))
		}
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("     Role: "))
		if agent.Role != "" {
			b.WriteString(s.Bold.Render(strings.ToUpper(agent.Role)))
		} else {
			b.WriteString(s.Subtle.Render("-"))
		}
		b.WriteString("\n")

		statusText, statusStyle := formatAgentStatus(agent.Status, s)
		b.WriteString(s.CardLabel.Render("   Status: "))
		b.WriteString(statusStyle.Render(statusText))
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("     Task: "))
		if agent.CurrentTaskID != "" {
			b.WriteString(s.CardValue.Render(agent.CurrentTaskID))
		} else {
			b.WriteString(s.Subtle.Render("idle"))
		}
		b.WriteString("\n")

		b.WriteString(s.CardLabel.Render("Activated: "))
		if agent.ActivatedAt > 0 {
			t := time.Unix(agent.ActivatedAt, 0)
			b.WriteString(s.CardValue.Render(t.Format("2006-01-02 15:04:05")))
		} else {
			b.WriteString(s.Subtle.Render("-"))
		}
		b.WriteString("\n")

		return InjectSystemMsg{Content: b.String()}
	}
}

// truncateID shortens agent IDs for display (first 8 chars).
func truncateID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// formatAgentStatus returns human-readable status and appropriate style.
func formatAgentStatus(status int, s *theme.Styles) (string, lipgloss.Style) {
	// Status bit flags (from hecate-daemon):
	// 1 = created, 2 = activated, 4 = working, 8 = idle, 16 = suspended
	switch {
	case status&4 != 0: // working
		return "working", s.StatusOK
	case status&8 != 0: // idle
		return "idle", s.Subtle
	case status&16 != 0: // suspended
		return "suspended", s.StatusWarning
	case status&2 != 0: // activated
		return "active", s.StatusOK
	case status&1 != 0: // created
		return "created", s.Subtle
	default:
		return "unknown", s.StatusError
	}
}
