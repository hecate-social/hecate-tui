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

// VentureCmd handles all /venture subcommands for business endeavor management.
type VentureCmd struct{}

func (c *VentureCmd) Name() string        { return "venture" }
func (c *VentureCmd) Aliases() []string   { return []string{"v"} }
func (c *VentureCmd) Description() string { return "Manage business endeavors (Ventures)" }

// Complete implements Completable for venture argument completion.
func (c *VentureCmd) Complete(args []string, ctx *Context) []string {
	// Subcommands
	subcommands := []string{"init", "new", "list", "ls", "select", "clear", "exit", "archive", "refine-vision", "refine", "rv", "submit-vision", "submit", "sv", "help", "status"}

	if len(args) == 0 {
		return subcommands
	}

	firstArg := strings.ToLower(args[0])

	// If we have 2+ args, we're completing the second argument
	// For "archive" and "select", complete venture IDs
	if len(args) >= 2 {
		switch firstArg {
		case "archive":
			// Archive needs venture IDs (include archived for visibility)
			return c.completeVentureIDs(args[1], ctx, true)
		case "select", "switch", "use":
			// Select needs active venture IDs/names
			return c.completeVentureIDs(args[1], ctx, false)
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

	// Single arg - complete subcommands and venture names
	var subMatches []string
	for _, sub := range subcommands {
		if strings.HasPrefix(sub, firstArg) {
			subMatches = append(subMatches, sub)
		}
	}

	// Also try to complete venture IDs/names for direct selection
	ventureMatches := c.completeVentureIDs(firstArg, ctx, false)

	// Combine
	return append(subMatches, ventureMatches...)
}

// completeVentureIDs returns venture IDs matching the prefix.
func (c *VentureCmd) completeVentureIDs(prefix string, ctx *Context, includeArchived bool) []string {
	var ventures []client.Venture
	var err error
	if includeArchived {
		ventures, err = ctx.Client.ListAllVentures()
	} else {
		ventures, err = ctx.Client.ListVentures()
	}
	if err != nil {
		return nil
	}

	prefix = strings.ToLower(prefix)
	var matches []string
	for _, venture := range ventures {
		// Match against ID
		if strings.HasPrefix(strings.ToLower(venture.VentureID), prefix) {
			matches = append(matches, venture.VentureID)
		}
		// Match against name (if different from ID)
		if strings.HasPrefix(strings.ToLower(venture.Name), prefix) && venture.Name != venture.VentureID {
			matches = append(matches, venture.Name)
		}
	}
	return matches
}

func (c *VentureCmd) Execute(args []string, ctx *Context) tea.Cmd {
	// No args → show current venture or list if none selected
	if len(args) == 0 {
		return c.showOrPick(ctx)
	}

	sub := strings.ToLower(args[0])

	switch sub {
	case "status":
		return c.showCurrentVenture(ctx)
	case "init", "new", "create":
		return c.initiateVenture(args[1:], ctx)
	case "list", "ls":
		includeArchived := len(args) > 1 && (args[1] == "all" || args[1] == "archived")
		return c.listVentures(ctx, includeArchived)
	case "select", "switch", "use":
		if len(args) < 2 {
			return c.showError(ctx, "Usage: /venture select <id|name|number>")
		}
		return c.selectVenture(args[1], ctx)
	case "clear", "exit":
		return c.clearVenture(ctx)
	case "archive":
		if len(args) < 2 {
			return c.showError(ctx, "Usage: /venture archive <venture-id> [reason]")
		}
		reason := ""
		if len(args) > 2 {
			reason = strings.Join(args[2:], " ")
		}
		return c.archiveVenture(args[1], reason, ctx)
	case "refine-vision", "refine", "rv":
		return c.refineVision(args[1:], ctx)
	case "submit-vision", "submit", "sv":
		return c.submitVision(ctx)
	case "help":
		return c.showUsage(ctx)
	default:
		// Check if it's a number (list index)
		if idx := c.parseIndex(sub); idx > 0 {
			return c.selectVentureByIndex(idx, ctx)
		}
		// Check if it looks like a venture ID
		if strings.HasPrefix(sub, "ven-") || strings.HasPrefix(sub, "venture-") {
			return c.selectVenture(sub, ctx)
		}
		// Otherwise treat as venture name to select
		return c.selectVenture(sub, ctx)
	}
}

func (c *VentureCmd) showUsage(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		t := ctx.Theme
		var b strings.Builder

		// Title
		b.WriteString(s.CardTitle.Render("Venture - Business Endeavor Commands"))
		b.WriteString("\n\n")

		// Intro
		b.WriteString(s.Subtle.Render("Manage business endeavors (Ventures) - the highest-level organizational unit in Hecate."))
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
		b.WriteString(row("/venture", "Show current venture status"))
		b.WriteString(row("/venture status", "Show current venture status"))
		b.WriteString(row("/venture init <name> [brief]", "Initiate a new venture"))
		b.WriteString(row("/venture archive <venture-id> [reason]", "Archive a venture (soft delete)"))
		b.WriteString(row("/venture refine-vision", "Open VISION.md for editing"))
		b.WriteString(row("/venture submit-vision", "Submit vision, complete DnA phase"))
		b.WriteString(row("/venture list", "List active ventures"))
		b.WriteString(row("/venture list all", "List all ventures (including archived)"))
		b.WriteString(row("/venture <id>", "Show specific venture by ID"))
		b.WriteString("\n")

		// Aliases for vision commands
		b.WriteString(s.Subtle.Render("Vision aliases: refine/rv, submit/sv"))
		b.WriteString("\n")

		// Aliases
		b.WriteString(s.Subtle.Render("Aliases: /v"))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *VentureCmd) showCurrentVenture(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		venture, err := ctx.Client.GetVenture()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get current venture: " + err.Error())}
		}

		return InjectSystemMsg{Content: c.renderVentureCard(venture, ctx)}
	}
}

func (c *VentureCmd) showVentureByID(ventureID string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		venture, err := ctx.Client.GetVentureByID(ventureID)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to get venture: " + err.Error())}
		}

		return InjectSystemMsg{Content: c.renderVentureCard(venture, ctx)}
	}
}

func (c *VentureCmd) listVentures(ctx *Context, includeArchived bool) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		var ventures []client.Venture
		var err error
		if includeArchived {
			ventures, err = ctx.Client.ListAllVentures()
		} else {
			ventures, err = ctx.Client.ListVentures()
		}
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list ventures: " + err.Error())}
		}

		if len(ventures) == 0 {
			var b strings.Builder
			title := "Ventures"
			if includeArchived {
				title = "Ventures (including archived)"
			}
			b.WriteString(s.CardTitle.Render(title))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("No ventures found. Use /venture init <name> [brief] to create one."))
			return InjectSystemMsg{Content: b.String()}
		}

		var b strings.Builder
		title := "Ventures"
		if includeArchived {
			title = "Ventures (including archived)"
		}
		b.WriteString(s.CardTitle.Render(title))
		b.WriteString("\n\n")

		for i, venture := range ventures {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(s.CardLabel.Render("    ID: "))
			b.WriteString(s.CardValue.Render(venture.VentureID))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("  Name: "))
			b.WriteString(s.CardValue.Render(venture.Name))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Status: "))
			b.WriteString(s.CardValue.Render(formatVentureStatusLabel(venture)))
			b.WriteString("\n")
			if venture.Brief != "" {
				b.WriteString(s.Subtle.Render("      " + venture.Brief))
				b.WriteString("\n")
			}
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *VentureCmd) initiateVenture(args []string, ctx *Context) tea.Cmd {
	// No args → show form
	if len(args) == 0 {
		return func() tea.Msg {
			return ShowFormMsg{FormType: "venture_init"}
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

	return c.doInitiateVenture(path, name, brief, ctx)
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

// VentureCreatedMsg is sent after a venture is successfully created and scaffolded.
// It triggers a cd to the new venture directory.
type VentureCreatedMsg struct {
	Path    string
	Message string
}

// doInitiateVenture performs the actual venture creation.
func (c *VentureCmd) doInitiateVenture(path, name, brief string, ctx *Context) tea.Cmd {
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

		venture, err := ctx.Client.InitiateVenture(name, brief)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to initiate venture: " + err.Error())}
		}

		// Scaffold the repository structure in the target path
		manifest := scaffold.VentureManifest{
			VentureID:   venture.VentureID,
			Name:        venture.Name,
			Brief:       venture.Brief,
			Root:        path,
			InitiatedAt: venture.InitiatedAt,
			InitiatedBy: venture.InitiatedBy,
		}

		result := scaffold.Scaffold(path, manifest)

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Venture Initiated"))
		b.WriteString("\n\n")
		b.WriteString(c.renderVentureCard(venture, ctx))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Root: " + path))

		// Show scaffolding results
		b.WriteString("\n\n")
		b.WriteString(s.CardTitle.Render("Scaffolded:"))
		b.WriteString("\n")

		if result.Success {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render(".hecate/venture.json"))
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

		if result.VisionCreated {
			b.WriteString(s.StatusOK.Render("  ✓ "))
			b.WriteString(s.Subtle.Render("VISION.md"))
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

		// Return VentureCreatedMsg to trigger cd
		return VentureCreatedMsg{Path: path, Message: b.String()}
	}
}

func (c *VentureCmd) archiveVenture(ventureID, reason string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Only accept venture IDs (not names) to avoid ambiguity
		if !strings.HasPrefix(ventureID, "venture-") {
			return InjectSystemMsg{Content: s.Error.Render("Please use venture ID (starts with 'venture-'). Use /venture list to see IDs.")}
		}

		err := ctx.Client.ArchiveVenture(ventureID, reason)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to archive venture: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Venture Archived"))
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("ID: " + ventureID))
		if reason != "" {
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("Reason: " + reason))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

// refineVision opens VISION.md for editing, scaffolding it first if needed.
// The file IS the vision — no key=value API params.
func (c *VentureCmd) refineVision(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Need a venture in context
		if ctx.GetALCContext == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}
		state := ctx.GetALCContext()
		if state == nil || state.Venture == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}

		// Venture root = cwd (the TUI cds into the venture dir on init/select)
		cwd, err := os.Getwd()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Cannot determine working directory: " + err.Error())}
		}

		// Scaffold VISION.md if it doesn't exist
		manifest := scaffold.VentureManifest{
			Name:        state.Venture.Name,
			Brief:       state.Venture.Brief,
			InitiatedBy: userAtHost(),
		}
		created, err := scaffold.ScaffoldVision(cwd, manifest)
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to scaffold VISION.md: " + err.Error())}
		}

		// Tell daemon vision refinement started
		params := map[string]interface{}{
			"refined_by": userAtHost(),
		}
		_ = ctx.Client.RefineVision(state.Venture.ID, params)

		visionPath := scaffold.VisionPath(cwd)
		if created {
			return EditFileMsg{Path: visionPath}
		}

		// Already exists — just open it
		return EditFileMsg{Path: visionPath}
	}
}

// submitVision submits the venture vision, completing the DnA phase.
// Validates that VISION.md exists before allowing submission.
func (c *VentureCmd) submitVision(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Need a venture in context
		if ctx.GetALCContext == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}
		state := ctx.GetALCContext()
		if state == nil || state.Venture == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}

		// Check VISION.md exists
		cwd, err := os.Getwd()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Cannot determine working directory: " + err.Error())}
		}
		if !scaffold.VisionExists(cwd) {
			return InjectSystemMsg{Content: s.Error.Render("No VISION.md found. Use /venture refine-vision to create and edit it first.")}
		}

		err = ctx.Client.SubmitVision(state.Venture.ID, userAtHost())
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to submit vision: " + err.Error())}
		}

		var b strings.Builder
		b.WriteString(s.StatusOK.Render("Vision Submitted"))
		b.WriteString("\n\n")
		b.WriteString(s.Subtle.Render("DnA phase complete. Venture is ready for Architecture & Planning."))
		return InjectSystemMsg{Content: b.String()}
	}
}

// userAtHost returns "user@hostname" for attribution.
func userAtHost() string {
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return user + "@" + hostname
}

func (c *VentureCmd) renderVentureCard(venture *client.Venture, ctx *Context) string {
	s := ctx.Styles
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Venture: " + venture.Name))
	b.WriteString("\n\n")

	// Right-align labels for clean formatting
	b.WriteString(s.CardLabel.Render("          ID: "))
	b.WriteString(s.CardValue.Render(venture.VentureID))
	b.WriteString("\n")

	if venture.Brief != "" {
		b.WriteString(s.CardLabel.Render("       Brief: "))
		b.WriteString(s.CardValue.Render(venture.Brief))
		b.WriteString("\n")
	}

	b.WriteString(s.CardLabel.Render("      Status: "))
	b.WriteString(s.CardValue.Render(formatVentureStatusLabel(*venture)))
	b.WriteString("\n")

	if venture.ActiveDepartmentID != "" {
		b.WriteString(s.CardLabel.Render("  Department: "))
		b.WriteString(s.CardValue.Render(venture.ActiveDepartmentID))
		b.WriteString("\n")
	}

	b.WriteString(s.CardLabel.Render("   Initiated: "))
	b.WriteString(s.Subtle.Render(formatTimestamp(venture.InitiatedAt)))
	b.WriteString("\n")

	if venture.InitiatedBy != "" {
		b.WriteString(s.CardLabel.Render("          By: "))
		b.WriteString(s.Subtle.Render(venture.InitiatedBy))
		b.WriteString("\n")
	}

	return b.String()
}

// formatVentureStatusLabel returns the status label, preferring the daemon-provided
// label (with emojis) and falling back to a local mapping.
func formatVentureStatusLabel(venture client.Venture) string {
	if venture.StatusLabel != "" {
		return venture.StatusLabel
	}
	return formatVentureStatus(venture.Status)
}

// formatVentureStatus converts a status bit field to a human-readable string.
// Matches venture_aggregate.erl bit flags.
func formatVentureStatus(status int) string {
	const (
		statusInitiated    = 1  // 2^0
		statusDNAActive    = 2  // 2^1
		statusDNAComplete  = 4  // 2^2
		statusImplementing = 8  // 2^3
		statusCompleted    = 16 // 2^4
		statusArchived     = 32 // 2^5
	)

	switch {
	case status&statusArchived != 0:
		return "Archived"
	case status&statusCompleted != 0:
		return "Completed"
	case status&statusImplementing != 0:
		return "Implementing"
	case status&statusDNAComplete != 0:
		return "Discovery Done"
	case status&statusDNAActive != 0:
		return "Discovering"
	case status&statusInitiated != 0:
		return "Initiated"
	default:
		return fmt.Sprintf("Unknown (%d)", status)
	}
}

// showOrPick shows current venture or lists available ventures to pick.
func (c *VentureCmd) showOrPick(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Check if we have a venture in context
		if ctx.GetALCContext != nil {
			if state := ctx.GetALCContext(); state != nil && state.Venture != nil {
				return c.showCurrentVenture(ctx)()
			}
		}

		// No venture selected - list available ventures
		ventures, err := ctx.Client.ListVentures()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list ventures: " + err.Error())}
		}

		if len(ventures) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No ventures found. Use /venture init <name> to create one.")}
		}

		return c.renderVenturePicker(ventures, ctx)
	}
}

// renderVenturePicker renders a numbered list for selection.
func (c *VentureCmd) renderVenturePicker(ventures []client.Venture, ctx *Context) tea.Msg {
	s := ctx.Styles
	t := ctx.Theme

	var b strings.Builder
	b.WriteString(s.CardTitle.Render("Select a Venture"))
	b.WriteString("\n\n")

	for i, venture := range ventures {
		// Numbered entry
		numStyle := lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
		b.WriteString(numStyle.Render(fmt.Sprintf("  %d. ", i+1)))

		// Name
		b.WriteString(s.CardValue.Render(venture.Name))

		// Brief if present
		if venture.Brief != "" {
			brief := venture.Brief
			if len(brief) > 40 {
				brief = brief[:37] + "..."
			}
			b.WriteString(" - ")
			b.WriteString(s.Subtle.Render(brief))
		}
		b.WriteString("\n")

		// ID on second line, indented
		b.WriteString("     ")
		b.WriteString(s.Subtle.Render(venture.VentureID))
		if venture.InitiatedAt > 0 {
			b.WriteString(" · ")
			b.WriteString(s.Subtle.Render(time.UnixMilli(venture.InitiatedAt).Format("2006-01-02")))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("Select: /venture <number> or /venture <id>"))
	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("Create: /venture init <name> [brief]"))

	return InjectSystemMsg{Content: b.String()}
}

// selectVenture switches to the specified venture.
func (c *VentureCmd) selectVenture(idOrName string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Try to find the venture by ID or name
		ventures, err := ctx.Client.ListVentures()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list ventures: " + err.Error())}
		}

		var selected *client.Venture

		// Check if it's a number (index)
		if idx := c.parseIndex(idOrName); idx > 0 && idx <= len(ventures) {
			selected = &ventures[idx-1]
		} else {
			// Find by ID (case-insensitive) or name (case-insensitive)
			for i := range ventures {
				if strings.EqualFold(ventures[i].VentureID, idOrName) || strings.EqualFold(ventures[i].Name, idOrName) {
					selected = &ventures[i]
					break
				}
			}
		}

		if selected == nil {
			return InjectSystemMsg{Content: s.Error.Render("Venture not found: " + idOrName)}
		}

		// Convert to VentureInfo and send message to switch context
		ventureInfo := &alc.VentureInfo{
			ID:          selected.VentureID,
			Name:        selected.Name,
			Brief:       selected.Brief,
			InitiatedAt: time.UnixMilli(selected.InitiatedAt),
		}

		return SetALCContextMsg{
			Context: alc.Venture,
			Venture: ventureInfo,
			Source:  "manual",
		}
	}
}

// selectVentureByIndex selects a venture by its list index (1-based).
func (c *VentureCmd) selectVentureByIndex(index int, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		ventures, err := ctx.Client.ListVentures()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list ventures: " + err.Error())}
		}

		if index < 1 || index > len(ventures) {
			return InjectSystemMsg{Content: s.Error.Render(fmt.Sprintf("Invalid index: %d (have %d ventures)", index, len(ventures)))}
		}

		selected := &ventures[index-1]
		ventureInfo := &alc.VentureInfo{
			ID:          selected.VentureID,
			Name:        selected.Name,
			Brief:       selected.Brief,
			InitiatedAt: time.UnixMilli(selected.InitiatedAt),
		}

		return SetALCContextMsg{
			Context: alc.Venture,
			Venture: ventureInfo,
			Source:  "manual",
		}
	}
}

// clearVenture exits venture mode and returns to chat.
func (c *VentureCmd) clearVenture(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetALCContextMsg{Context: alc.Chat}
	}
}

// parseIndex converts a string to an integer index, returns 0 if not a number.
func (c *VentureCmd) parseIndex(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}

// showError returns an error message.
func (c *VentureCmd) showError(ctx *Context, msg string) tea.Cmd {
	return func() tea.Msg {
		return InjectSystemMsg{Content: ctx.Styles.Error.Render(msg)}
	}
}

// VenturesCmd handles /ventures (alias for /venture list).
type VenturesCmd struct{}

func (c *VenturesCmd) Name() string        { return "ventures" }
func (c *VenturesCmd) Aliases() []string   { return []string{"vs"} }
func (c *VenturesCmd) Description() string { return "List all ventures" }

func (c *VenturesCmd) Execute(args []string, ctx *Context) tea.Cmd {
	ventureCmd := &VentureCmd{}
	includeArchived := len(args) > 0 && (args[0] == "all" || args[0] == "archived")
	return ventureCmd.listVentures(ctx, includeArchived)
}

// ChatCmd handles /chat - returns to Chat mode (clears venture context).
type ChatCmd struct{}

func (c *ChatCmd) Name() string        { return "chat" }
func (c *ChatCmd) Aliases() []string   { return nil }
func (c *ChatCmd) Description() string { return "Return to chat mode (clear venture context)" }

func (c *ChatCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		return SetALCContextMsg{Context: alc.Chat}
	}
}

// BackCmd handles /back - navigate up the context hierarchy.
type BackCmd struct{}

func (c *BackCmd) Name() string        { return "back" }
func (c *BackCmd) Aliases() []string   { return []string{"b", ".."} }
func (c *BackCmd) Description() string { return "Navigate back (Department -> Venture -> Chat)" }

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
		case alc.Department:
			// Back to Venture mode, keep the venture
			return SetALCContextMsg{
				Context: alc.Venture,
				Venture: state.Venture,
			}
		case alc.Venture:
			// Back to Chat mode
			return SetALCContextMsg{Context: alc.Chat}
		default:
			// Already in Chat mode
			ctx.Styles.Subtle.Render("Already in chat mode.")
			return nil
		}
	}
}

// DepartmentsCmd handles /departments - list departments in current venture.
type DepartmentsCmd struct{}

func (c *DepartmentsCmd) Name() string        { return "departments" }
func (c *DepartmentsCmd) Aliases() []string   { return []string{"dpts"} }
func (c *DepartmentsCmd) Description() string { return "List departments in current venture" }

func (c *DepartmentsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Check if we have a venture in context
		if ctx.GetALCContext == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}

		state := ctx.GetALCContext()
		if state == nil || state.Venture == nil {
			return InjectSystemMsg{Content: s.Error.Render("No venture selected. Use /venture to select one first.")}
		}

		// For now, delegate to /department command
		// TODO: Filter departments by current venture when API supports it
		departmentCmd := &DepartmentCmd{}
		return departmentCmd.Execute(nil, ctx)()
	}
}
