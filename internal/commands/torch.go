package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/client"
)

// TorchCmd handles all /torch subcommands for business endeavor management.
type TorchCmd struct{}

func (c *TorchCmd) Name() string        { return "torch" }
func (c *TorchCmd) Aliases() []string   { return []string{"t"} }
func (c *TorchCmd) Description() string { return "Manage business endeavors (Torches)" }

func (c *TorchCmd) Execute(args []string, ctx *Context) tea.Cmd {
	// No args → show current torch or list if none selected
	if len(args) == 0 {
		return c.showOrPick(ctx)
	}

	sub := strings.ToLower(args[0])

	switch sub {
	case "status":
		return c.showCurrentTorch(ctx)
	case "init", "new", "create":
		return c.initiateTorch(args[1:], ctx)
	case "list", "ls":
		return c.listTorches(ctx)
	case "select", "switch", "use":
		if len(args) < 2 {
			return c.showError(ctx, "Usage: /torch select <id|name|number>")
		}
		return c.selectTorch(args[1], ctx)
	case "clear", "exit":
		return c.clearTorch(ctx)
	case "help":
		return c.showUsage(ctx)
	default:
		// Check if it's a number (list index)
		if idx := c.parseIndex(sub); idx > 0 {
			return c.selectTorchByIndex(idx, ctx)
		}
		// Check if it looks like a torch ID
		if strings.HasPrefix(sub, "tch-") || strings.HasPrefix(sub, "torch-") {
			return c.selectTorch(sub, ctx)
		}
		// Otherwise treat as torch name to select
		return c.selectTorch(sub, ctx)
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

// showOrPick shows current torch or lists available torches to pick.
func (c *TorchCmd) showOrPick(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Check if we have a torch in context
		if ctx.GetALCContext != nil {
			if state := ctx.GetALCContext(); state != nil && state.Torch != nil {
				return c.showCurrentTorch(ctx)()
			}
		}

		// No torch selected - list available torches
		torches, err := ctx.Client.ListTorches()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list torches: " + err.Error())}
		}

		if len(torches) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No torches found. Use /torch init <name> to create one.")}
		}

		return c.renderTorchPicker(torches, ctx)
	}
}

// renderTorchPicker renders a numbered list for selection.
func (c *TorchCmd) renderTorchPicker(torches []client.Torch, ctx *Context) tea.Msg {
	s := ctx.Styles
	t := ctx.Theme

	var b strings.Builder
	b.WriteString(s.CardTitle.Render("Select a Torch"))
	b.WriteString("\n\n")

	for i, torch := range torches {
		// Numbered entry
		numStyle := lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
		b.WriteString(numStyle.Render(fmt.Sprintf("  %d. ", i+1)))

		// Name
		b.WriteString(s.CardValue.Render(torch.Name))

		// Brief if present
		if torch.Brief != "" {
			brief := torch.Brief
			if len(brief) > 40 {
				brief = brief[:37] + "..."
			}
			b.WriteString(" - ")
			b.WriteString(s.Subtle.Render(brief))
		}
		b.WriteString("\n")

		// ID on second line, indented
		b.WriteString("     ")
		b.WriteString(s.Subtle.Render(torch.TorchID))
		if torch.InitiatedAt > 0 {
			b.WriteString(" · ")
			b.WriteString(s.Subtle.Render(time.Unix(torch.InitiatedAt, 0).Format("2006-01-02")))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("Select: /torch <number> or /torch <id>"))
	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("Create: /torch init <name> [brief]"))

	return InjectSystemMsg{Content: b.String()}
}

// selectTorch switches to the specified torch.
func (c *TorchCmd) selectTorch(idOrName string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Try to find the torch by ID or name
		torches, err := ctx.Client.ListTorches()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list torches: " + err.Error())}
		}

		var selected *client.Torch

		// Check if it's a number (index)
		if idx := c.parseIndex(idOrName); idx > 0 && idx <= len(torches) {
			selected = &torches[idx-1]
		} else {
			// Find by ID or name (case-insensitive for name)
			for i := range torches {
				if torches[i].TorchID == idOrName || strings.EqualFold(torches[i].Name, idOrName) {
					selected = &torches[i]
					break
				}
			}
		}

		if selected == nil {
			return InjectSystemMsg{Content: s.Error.Render("Torch not found: " + idOrName)}
		}

		// Convert to TorchInfo and send message to switch context
		torchInfo := &alc.TorchInfo{
			ID:          selected.TorchID,
			Name:        selected.Name,
			Brief:       selected.Brief,
			InitiatedAt: time.Unix(selected.InitiatedAt, 0),
		}

		return SetALCContextMsg{
			Context: alc.Torch,
			Torch:   torchInfo,
			Source:  "manual",
		}
	}
}

// selectTorchByIndex selects a torch by its list index (1-based).
func (c *TorchCmd) selectTorchByIndex(index int, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		torches, err := ctx.Client.ListTorches()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list torches: " + err.Error())}
		}

		if index < 1 || index > len(torches) {
			return InjectSystemMsg{Content: s.Error.Render(fmt.Sprintf("Invalid index: %d (have %d torches)", index, len(torches)))}
		}

		selected := &torches[index-1]
		torchInfo := &alc.TorchInfo{
			ID:          selected.TorchID,
			Name:        selected.Name,
			Brief:       selected.Brief,
			InitiatedAt: time.Unix(selected.InitiatedAt, 0),
		}

		return SetALCContextMsg{
			Context: alc.Torch,
			Torch:   torchInfo,
			Source:  "manual",
		}
	}
}

// clearTorch exits torch mode and returns to chat.
func (c *TorchCmd) clearTorch(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetALCContextMsg{Context: alc.Chat}
	}
}

// parseIndex converts a string to an integer index, returns 0 if not a number.
func (c *TorchCmd) parseIndex(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}

// showError returns an error message.
func (c *TorchCmd) showError(ctx *Context, msg string) tea.Cmd {
	return func() tea.Msg {
		return InjectSystemMsg{Content: ctx.Styles.Error.Render(msg)}
	}
}

// TorchesCmd handles /torches (alias for /torch list).
type TorchesCmd struct{}

func (c *TorchesCmd) Name() string        { return "torches" }
func (c *TorchesCmd) Aliases() []string   { return []string{"ts"} }
func (c *TorchesCmd) Description() string { return "List all torches" }

func (c *TorchesCmd) Execute(args []string, ctx *Context) tea.Cmd {
	torchCmd := &TorchCmd{}
	return torchCmd.listTorches(ctx)
}

// ChatCmd handles /chat - returns to Chat mode (clears torch context).
type ChatCmd struct{}

func (c *ChatCmd) Name() string        { return "chat" }
func (c *ChatCmd) Aliases() []string   { return nil }
func (c *ChatCmd) Description() string { return "Return to chat mode (clear torch context)" }

func (c *ChatCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetALCContextMsg{Context: alc.Chat}
	}
}

// BackCmd handles /back - navigate up the context hierarchy.
type BackCmd struct{}

func (c *BackCmd) Name() string        { return "back" }
func (c *BackCmd) Aliases() []string   { return []string{"b", ".."} }
func (c *BackCmd) Description() string { return "Navigate back (Cartwheel -> Torch -> Chat)" }

func (c *BackCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		if ctx.GetALCContext == nil {
			return SetALCContextMsg{Context: alc.Chat}
		}

		state := ctx.GetALCContext()
		if state == nil {
			return SetALCContextMsg{Context: alc.Chat}
		}

		switch state.Context {
		case alc.Cartwheel:
			// Back to Torch mode, keep the torch
			return SetALCContextMsg{
				Context: alc.Torch,
				Torch:   state.Torch,
			}
		case alc.Torch:
			// Back to Chat mode
			return SetALCContextMsg{Context: alc.Chat}
		default:
			// Already in Chat mode
			ctx.Styles.Subtle.Render("Already in chat mode.")
			return nil
		}
	}
}

// CartwheelsCmd handles /cartwheels - list cartwheels in current torch.
type CartwheelsCmd struct{}

func (c *CartwheelsCmd) Name() string        { return "cartwheels" }
func (c *CartwheelsCmd) Aliases() []string   { return []string{"cws"} }
func (c *CartwheelsCmd) Description() string { return "List cartwheels in current torch" }

func (c *CartwheelsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Check if we have a torch in context
		if ctx.GetALCContext == nil {
			return InjectSystemMsg{Content: s.Error.Render("No torch selected. Use /torch to select one first.")}
		}

		state := ctx.GetALCContext()
		if state == nil || state.Torch == nil {
			return InjectSystemMsg{Content: s.Error.Render("No torch selected. Use /torch to select one first.")}
		}

		// For now, delegate to /cartwheel command
		// TODO: Filter cartwheels by current torch when API supports it
		cartwheelCmd := &CartwheelCmd{}
		return cartwheelCmd.Execute(nil, ctx)()
	}
}
