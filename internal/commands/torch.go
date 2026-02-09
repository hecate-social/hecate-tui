package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/scaffold"
)

// TorchCmd handles all /torch subcommands for business endeavor management.
type TorchCmd struct{}

func (c *TorchCmd) Name() string        { return "torch" }
func (c *TorchCmd) Aliases() []string   { return []string{"t"} }
func (c *TorchCmd) Description() string { return "Manage business endeavors (Torches)" }

// Complete implements Completable for torch argument completion.
func (c *TorchCmd) Complete(args []string, ctx *Context) []string {
	// Subcommands
	subcommands := []string{"init", "new", "list", "ls", "select", "clear", "exit", "archive", "help", "status"}

	if len(args) == 0 {
		return subcommands
	}

	firstArg := strings.ToLower(args[0])

	// If we have 2+ args, we're completing the second argument
	// For "archive" and "select", complete torch IDs
	if len(args) >= 2 {
		switch firstArg {
		case "archive":
			// Archive needs torch IDs (include archived for visibility)
			return c.completeTorchIDs(args[1], ctx, true)
		case "select", "switch", "use":
			// Select needs active torch IDs/names
			return c.completeTorchIDs(args[1], ctx, false)
		case "list", "ls":
			// List can have "all" or "archived" as second arg
			prefix := strings.ToLower(args[1])
			var matches []string
			for _, opt := range []string{"all", "archived"} {
				if strings.HasPrefix(opt, prefix) {
					matches = append(matches, opt)
				}
			}
			return matches
		}
		return nil
	}

	// Single arg - complete subcommands and torch names
	var subMatches []string
	for _, sub := range subcommands {
		if strings.HasPrefix(sub, firstArg) {
			subMatches = append(subMatches, sub)
		}
	}

	// Also try to complete torch IDs/names for direct selection
	torchMatches := c.completeTorchIDs(firstArg, ctx, false)

	// Combine
	return append(subMatches, torchMatches...)
}

// completeTorchIDs returns torch IDs matching the prefix.
func (c *TorchCmd) completeTorchIDs(prefix string, ctx *Context, includeArchived bool) []string {
	var torches []client.Torch
	var err error
	if includeArchived {
		torches, err = ctx.Client.ListAllTorches()
	} else {
		torches, err = ctx.Client.ListTorches()
	}
	if err != nil {
		return nil
	}

	prefix = strings.ToLower(prefix)
	var matches []string
	for _, torch := range torches {
		// Match against ID
		if strings.HasPrefix(strings.ToLower(torch.TorchID), prefix) {
			matches = append(matches, torch.TorchID)
		}
		// Match against name (if different from ID)
		if strings.HasPrefix(strings.ToLower(torch.Name), prefix) && torch.Name != torch.TorchID {
			matches = append(matches, torch.Name)
		}
	}
	return matches
}

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
		includeArchived := len(args) > 1 && (args[1] == "all" || args[1] == "archived")
		return c.listTorches(ctx, includeArchived)
	case "select", "switch", "use":
		if len(args) < 2 {
			return c.showError(ctx, "Usage: /torch select <id|name|number>")
		}
		return c.selectTorch(args[1], ctx)
	case "clear", "exit":
		return c.clearTorch(ctx)
	case "archive":
		if len(args) < 2 {
			return c.showError(ctx, "Usage: /torch archive <torch-id> [reason]")
		}
		reason := ""
		if len(args) > 2 {
			reason = strings.Join(args[2:], " ")
		}
		return c.archiveTorch(args[1], reason, ctx)
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
		b.WriteString(row("/torch archive <torch-id> [reason]", "Archive a torch (soft delete)"))
		b.WriteString(row("/torch list", "List active torches"))
		b.WriteString(row("/torch list all", "List all torches (including archived)"))
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

func (c *TorchCmd) listTorches(ctx *Context, includeArchived bool) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		var torches []client.Torch
		var err error
		if includeArchived {
			torches, err = ctx.Client.ListAllTorches()
		} else {
			torches, err = ctx.Client.ListTorches()
		}
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list torches: " + err.Error())}
		}

		if len(torches) == 0 {
			var b strings.Builder
			title := "Torches"
			if includeArchived {
				title = "Torches (including archived)"
			}
			b.WriteString(s.CardTitle.Render(title))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("No torches found. Use /torch init <name> [brief] to create one."))
			return InjectSystemMsg{Content: b.String()}
		}

		var b strings.Builder
		title := "Torches"
		if includeArchived {
			title = "Torches (including archived)"
		}
		b.WriteString(s.CardTitle.Render(title))
		b.WriteString("\n\n")

		for i, torch := range torches {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(s.CardLabel.Render("  ID: "))
			b.WriteString(s.CardValue.Render(torch.TorchID))
			// Show archived badge if archived (status bit 32)
			if torch.Status&32 != 0 {
				b.WriteString(" ")
				b.WriteString(s.Error.Render("[archived]"))
			}
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
	// No args → show form
	if len(args) == 0 {
		return func() tea.Msg {
			return ShowFormMsg{FormType: "torch_init"}
		}
	}

	// With args → create directly (power user mode)
	// First arg is path (e.g., "my-venture" or "~/projects/my-venture")
	// Second arg onwards is brief
	cwd, _ := os.Getwd()
	path := expandPath(args[0], cwd)
	name := inferName(path)
	brief := ""
	if len(args) > 1 {
		brief = strings.Join(args[1:], " ")
	}

	return c.doInitiateTorch(path, name, brief, ctx)
}

// expandPath expands ~ and makes path absolute relative to cwd.
func expandPath(path, cwd string) string {
	if path == "" {
		return cwd
	}

	// Expand ~
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			path = home + path[1:]
		}
	}

	// Make absolute if relative
	if !strings.HasPrefix(path, "/") {
		path = cwd + "/" + path
	}

	// Clean the path
	return cleanPath(path)
}

// cleanPath normalizes a path (removes . and ..)
func cleanPath(path string) string {
	parts := strings.Split(path, "/")
	var result []string
	for _, p := range parts {
		if p == ".." && len(result) > 0 {
			result = result[:len(result)-1]
		} else if p != "." && p != "" {
			result = append(result, p)
		}
	}
	return "/" + strings.Join(result, "/")
}

// inferName extracts the project name from a path.
func inferName(path string) string {
	// Get last non-empty component
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return "unnamed"
}

// TorchCreatedMsg is sent after a torch is successfully created and scaffolded.
// It triggers a cd to the new torch directory.
type TorchCreatedMsg struct {
	Path    string
	Message string
}

// doInitiateTorch performs the actual torch creation.
func (c *TorchCmd) doInitiateTorch(path, name, brief string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if strings.TrimSpace(path) == "" {
			return InjectSystemMsg{Content: s.Error.Render("Path is required")}
		}

		if strings.TrimSpace(name) == "" {
			name = inferName(path)
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(path, 0755); err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to create directory: " + err.Error())}
		}

		torch, err := ctx.Client.InitiateTorch(name, brief)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate torch: " + err.Error())}
		}

		// Scaffold the repository structure in the target path
		manifest := scaffold.TorchManifest{
			TorchID:     torch.TorchID,
			Name:        torch.Name,
			Brief:       torch.Brief,
			Root:        path,
			InitiatedAt: torch.InitiatedAt,
			InitiatedBy: torch.InitiatedBy,
		}

		result := scaffold.Scaffold(path, manifest)

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Torch Initiated"))
		b.WriteString("\n\n")
		b.WriteString(c.renderTorchCard(torch, ctx))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Root: " + path))

		// Show scaffolding results
		b.WriteString("\n\n")
		b.WriteString(s.CardTitle.Render("Scaffolded:"))
		b.WriteString("\n")

		if result.Success {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render(".hecate/torch.json"))
			b.WriteString("\n")
		}

		if result.AgentsCloned {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render(".hecate/agents/"))
			b.WriteString("\n")
		}

		if result.ReadmeCreated {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("README.md"))
			b.WriteString("\n")
		}

		if result.ChangelogCreated {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("CHANGELOG.md"))
			b.WriteString("\n")
		}

		if result.GitInitialized {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("git init"))
			b.WriteString("\n")
		}

		if result.GitCommitted {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("git commit"))
			b.WriteString("\n")
		}

		// Show warnings
		for _, warn := range result.Warnings {
			b.WriteString("\n")
			b.WriteString(s.StatusWarning.Render("⚠ " + warn))
		}

		// Hint about next steps
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("Next: gh repo create --public --source=. --push"))

		// Return TorchCreatedMsg to trigger cd
		return TorchCreatedMsg{Path: path, Message: b.String()}
	}
}

func (c *TorchCmd) archiveTorch(torchID, reason string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Only accept torch IDs (not names) to avoid ambiguity
		if !strings.HasPrefix(torchID, "torch-") {
			return InjectSystemMsg{Content: s.Error.Render("Please use torch ID (starts with 'torch-'). Use /torch list to see IDs.")}
		}

		err := ctx.Client.ArchiveTorch(torchID, reason)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to archive torch: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Torch Archived"))
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("ID: " + torchID))
		if reason != "" {
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("Reason: " + reason))
		}

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
			b.WriteString(s.Subtle.Render(time.UnixMilli(torch.InitiatedAt).Format("2006-01-02")))
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
			// Find by ID (case-insensitive) or name (case-insensitive)
			for i := range torches {
				if strings.EqualFold(torches[i].TorchID, idOrName) || strings.EqualFold(torches[i].Name, idOrName) {
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
			InitiatedAt: time.UnixMilli(selected.InitiatedAt),
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
			InitiatedAt: time.UnixMilli(selected.InitiatedAt),
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
	includeArchived := len(args) > 0 && (args[0] == "all" || args[0] == "archived")
	return torchCmd.listTorches(ctx, includeArchived)
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
