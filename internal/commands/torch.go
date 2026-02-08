package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
)

// TorchCmd handles all /torch subcommands for business endeavor management.
type TorchCmd struct{}

func (c *TorchCmd) Name() string        { return "torch" }
func (c *TorchCmd) Aliases() []string   { return []string{"t"} }
func (c *TorchCmd) Description() string { return "Manage business endeavors (Torches)" }

func (c *TorchCmd) Execute(args []string, ctx *Context) tea.Cmd {
	// No args or "status" → show current torch
	if len(args) == 0 {
		return c.showCurrentTorch(ctx)
	}

	sub := strings.ToLower(args[0])

	switch sub {
	case "status":
		return c.showCurrentTorch(ctx)
	case "init":
		return c.initiateTorch(args[1:], ctx)
	case "list":
		return c.listTorches(ctx)
	case "help":
		return c.showUsage(ctx)
	default:
		// Check if it looks like a torch ID
		if strings.HasPrefix(sub, "tch-") || strings.HasPrefix(sub, "torch-") {
			return c.showTorchByID(sub, ctx)
		}
		// Otherwise treat as torch ID anyway (could be a short ID or name)
		return c.showTorchByID(sub, ctx)
	}
}

func (c *TorchCmd) showUsage(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		t := ctx.Theme
		var b strings.Builder

		// Title
		b.WriteString(s.CardTitle.Render("Torch - Business Endeavor Commands"))
		b.WriteString("\n\n")

		// Intro
		b.WriteString(s.Subtle.Render("Manage business endeavors (Torches) - the highest-level organizational unit in Hecate."))
		b.WriteString("\n\n")

		// Helper for table rows
		cmdStyle := lipgloss.NewStyle().Foreground(t.Secondary)
		descStyle := lipgloss.NewStyle().Foreground(t.Text)

		row := func(cmd, desc string) string {
			padded := cmd
			for len(padded) < 30 {
				padded += " "
			}
			return cmdStyle.Render(padded) + descStyle.Render(desc) + "\n"
		}

		// Commands
		b.WriteString(s.Bold.Render("Commands"))
		b.WriteString("\n")
		b.WriteString(row("/torch", "Show current torch status"))
		b.WriteString(row("/torch status", "Show current torch status"))
		b.WriteString(row("/torch init <name> [brief]", "Initiate a new torch"))
		b.WriteString(row("/torch list", "List all torches"))
		b.WriteString(row("/torch <id>", "Show specific torch by ID"))
		b.WriteString("\n")

		// Aliases
		b.WriteString(s.Subtle.Render("Aliases: /t"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *TorchCmd) showCurrentTorch(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		torch, err := ctx.Client.GetTorch()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get current torch: " + err.Error())}
		}

		return InjectSystemMsg{Content: c.renderTorchCard(torch, ctx)}
	}
}

func (c *TorchCmd) showTorchByID(torchID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		torch, err := ctx.Client.GetTorchByID(torchID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get torch: " + err.Error())}
		}

		return InjectSystemMsg{Content: c.renderTorchCard(torch, ctx)}
	}
}

func (c *TorchCmd) listTorches(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		torches, err := ctx.Client.ListTorches()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list torches: " + err.Error())}
		}

		if len(torches) == 0 {
			var b strings.Builder
			b.WriteString(s.CardTitle.Render("Torches"))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("No torches found. Use /torch init <name> [brief] to create one."))
			return InjectSystemMsg{Content: b.String()}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Torches"))
		b.WriteString("\n\n")

		for i, torch := range torches {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(s.CardLabel.Render("  ID: "))
			b.WriteString(s.CardValue.Render(torch.TorchID))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Name: "))
			b.WriteString(s.CardValue.Render(torch.Name))
			b.WriteString("\n")
			if torch.Brief != "" {
				b.WriteString(s.Subtle.Render("      " + torch.Brief))
				b.WriteString("\n")
			}
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *TorchCmd) initiateTorch(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Error.Render("Usage: /torch init <name> [brief]")}
		}
	}

	name := args[0]
	brief := ""
	if len(args) > 1 {
		brief = strings.Join(args[1:], " ")
	}

	return func() tea.Msg {
		s := ctx.Styles

		torch, err := ctx.Client.InitiateTorch(name, brief)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate torch: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Torch Initiated"))
		b.WriteString("\n\n")
		b.WriteString(c.renderTorchCard(torch, ctx))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *TorchCmd) renderTorchCard(torch *client.Torch, ctx *Context) string {
	s := ctx.Styles
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Torch: " + torch.Name))
	b.WriteString("\n\n")

	// Right-align labels for clean formatting
	b.WriteString(s.CardLabel.Render("          ID: "))
	b.WriteString(s.CardValue.Render(torch.TorchID))
	b.WriteString("\n")

	if torch.Brief != "" {
		b.WriteString(s.CardLabel.Render("       Brief: "))
		b.WriteString(s.CardValue.Render(torch.Brief))
		b.WriteString("\n")
	}

	b.WriteString(s.CardLabel.Render("      Status: "))
	b.WriteString(s.CardValue.Render(formatTorchStatus(torch.Status)))
	b.WriteString("\n")

	if torch.ActiveCartwheelID != "" {
		b.WriteString(s.CardLabel.Render("  Cartwheel: "))
		b.WriteString(s.CardValue.Render(torch.ActiveCartwheelID))
		b.WriteString("\n")
	}

	b.WriteString(s.CardLabel.Render("   Initiated: "))
	b.WriteString(s.Subtle.Render(formatTimestamp(torch.InitiatedAt)))
	b.WriteString("\n")

	if torch.InitiatedBy != "" {
		b.WriteString(s.CardLabel.Render("          By: "))
		b.WriteString(s.Subtle.Render(torch.InitiatedBy))
		b.WriteString("\n")
	}

	return b.String()
}

// formatTorchStatus converts a status bit field to a human-readable string.
func formatTorchStatus(status int) string {
	// Status bit flags (assumed based on typical patterns)
	const (
		statusInitiated = 1 << iota
		statusActive
		statusPaused
		statusCompleted
		statusArchived
	)

	switch {
	case status&statusCompleted != 0:
		return "Completed"
	case status&statusArchived != 0:
		return "Archived"
	case status&statusPaused != 0:
		return "Paused"
	case status&statusActive != 0:
		return "Active"
	case status&statusInitiated != 0:
		return "Initiated"
	default:
		return fmt.Sprintf("Unknown (%d)", status)
	}
}
