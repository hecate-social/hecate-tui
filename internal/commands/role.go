package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/config"
)

// RoleCmd switches between ALC roles (DnA, AnP, TnI, DnO).
type RoleCmd struct{}

func (c *RoleCmd) Name() string        { return "roles" }
func (c *RoleCmd) Aliases() []string   { return []string{"role", "r"} }
func (c *RoleCmd) Description() string { return "Switch ALC role (/roles <dna|anp|tni|dno>)" }

func (c *RoleCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 || args[0] == "list" {
		return c.listRoles(ctx)
	}

	role := strings.ToLower(args[0])

	// Validate role
	if _, ok := config.RoleInfo[role]; !ok {
		return func() tea.Msg {
			return InjectSystemMsg{
				Content: ctx.Styles.Error.Render("Unknown role: "+role) +
					"\n" + ctx.Styles.Subtle.Render("Available: dna, anp, tni, dno"),
			}
		}
	}

	return func() tea.Msg {
		return SwitchRoleMsg{Role: role}
	}
}

func (c *RoleCmd) listRoles(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		activeRole := ""
		if ctx.GetActiveRole != nil {
			activeRole = ctx.GetActiveRole()
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("ALC Roles"))
		b.WriteString("\n\n")

		// Order: dna, anp, tni, dno
		roles := []string{"dna", "anp", "tni", "dno"}
		for _, role := range roles {
			info := config.RoleInfo[role]
			marker := "  "
			if role == activeRole {
				marker = "● "
			}
			b.WriteString(s.Bold.Render(marker + role))
			b.WriteString(s.Subtle.Render("  " + info.DisplayName))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Use /roles <code> to switch"))
		b.WriteString("\n\n")

		// Show descriptions
		b.WriteString(s.CardTitle.Render("Phase Descriptions"))
		b.WriteString("\n\n")
		b.WriteString(s.Bold.Render("DnA"))
		b.WriteString(" — Discovery & Analysis\n")
		b.WriteString(s.Subtle.Render("    Understand the problem before solving it."))
		b.WriteString("\n\n")
		b.WriteString(s.Bold.Render("AnP"))
		b.WriteString(" — Architecture & Planning\n")
		b.WriteString(s.Subtle.Render("    Design the solution before building it."))
		b.WriteString("\n\n")
		b.WriteString(s.Bold.Render("TnI"))
		b.WriteString(" — Testing & Implementation\n")
		b.WriteString(s.Subtle.Render("    Build it right."))
		b.WriteString("\n\n")
		b.WriteString(s.Bold.Render("DnO"))
		b.WriteString(" — Deployment & Operations\n")
		b.WriteString(s.Subtle.Render("    Ship it and keep it running."))

		return InjectSystemMsg{Content: b.String()}
	}
}
