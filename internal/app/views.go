package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/modes"
)

// View renders the entire TUI.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Browse mode uses split or full layout
	if a.mode == modes.Browse && a.browseReady {
		return a.renderBrowseLayout()
	}

	// Pair mode uses split or full layout
	if a.mode == modes.Pair && a.pairReady {
		return a.renderPairLayout()
	}

	// Edit mode takes full screen
	if a.mode == modes.Edit && a.editorReady {
		return a.renderEditLayout()
	}

	// Form mode overlays the chat
	if a.mode == modes.Form && a.formReady {
		return a.renderFormLayout()
	}

	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	// Chat area (always visible)
	sections = append(sections, a.chat.ViewChat())

	// Stats/streaming indicator
	if stats := a.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}

	// Error
	if errView := a.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}

	// Input area (mode-dependent)
	switch a.mode {
	case modes.Insert:
		sections = append(sections, a.chat.ViewInput())
	case modes.Command:
		sections = append(sections, a.renderCommandLine())
	}

	// Status bar (always at bottom)
	sections = append(sections, a.statusBar.View())

	content := strings.Join(sections, "\n")

	// Overlay tool approval prompt if there's a pending approval
	if a.chat.HasPendingApproval() && a.approvalPrompt != nil {
		content = a.renderWithApprovalOverlay(content)
	}

	return content
}

// renderWithApprovalOverlay overlays the approval prompt on top of the content.
func (a *App) renderWithApprovalOverlay(content string) string {
	call := a.chat.PendingToolCall()
	if call == nil {
		return content
	}

	// Get tool info from registry
	registry := a.toolExecutor.Registry()
	tool, _, ok := registry.Get(call.Name)
	if !ok {
		// Unknown tool, just show basic info
		tool = llmtools.Tool{
			Name:        call.Name,
			Description: "Unknown tool",
			Category:    llmtools.CategorySystem,
		}
	}

	// Set width based on terminal
	dialogWidth := 60
	if a.width > 80 {
		dialogWidth = 70
	}
	if a.width < 70 {
		dialogWidth = a.width - 4
	}
	a.approvalPrompt.SetWidth(dialogWidth)

	// Render the approval prompt
	prompt := a.approvalPrompt.Render(tool, *call)

	// Center the prompt on the screen
	promptLines := strings.Split(prompt, "\n")
	promptHeight := len(promptLines)

	// Calculate vertical position (center)
	contentLines := strings.Split(content, "\n")
	startLine := (len(contentLines) - promptHeight) / 2
	if startLine < 0 {
		startLine = 0
	}

	// Calculate horizontal padding to center
	maxPromptWidth := 0
	for _, line := range promptLines {
		if w := lipgloss.Width(line); w > maxPromptWidth {
			maxPromptWidth = w
		}
	}
	leftPad := (a.width - maxPromptWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

	// Overlay the prompt
	for i, line := range promptLines {
		lineIdx := startLine + i
		if lineIdx >= 0 && lineIdx < len(contentLines) {
			contentLines[lineIdx] = padding + line
		}
	}

	return strings.Join(contentLines, "\n")
}

func (a *App) renderBrowseLayout() string {
	// Render the normal chat view as the background
	var sections []string
	sections = append(sections, a.renderHeader())
	sections = append(sections, a.chat.ViewChat())
	if stats := a.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}
	if errView := a.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}
	sections = append(sections, a.statusBar.View())
	background := strings.Join(sections, "\n")

	// Dim the background
	backgroundLines := strings.Split(background, "\n")
	for i, line := range backgroundLines {
		backgroundLines[i] = lipgloss.NewStyle().Foreground(a.theme.TextMuted).Render(line)
	}

	// The browse modal handles its own centering
	modal := a.browseView.View()
	modalLines := strings.Split(modal, "\n")

	// Overlay the modal on the dimmed background
	result := make([]string, len(backgroundLines))
	copy(result, backgroundLines)

	// Overlay modal lines onto background
	for i, line := range modalLines {
		if i < len(result) && strings.TrimSpace(line) != "" {
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

func (a *App) renderPairLayout() string {
	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	if a.width >= 100 {
		// Split pane: dimmed chat left, pair right
		chatWidth := a.width - a.pairWidth() - 1
		chatHeight := a.pairHeight()

		chatContent := a.chat.ViewChat()
		dimmedChat := lipgloss.NewStyle().
			Width(chatWidth).
			Height(chatHeight).
			Foreground(a.theme.TextMuted).
			Render(chatContent)

		pairPanel := a.pairView.View()

		sep := lipgloss.NewStyle().
			Foreground(a.theme.Border).
			Render("│")

		split := lipgloss.JoinHorizontal(lipgloss.Top, dimmedChat, sep, pairPanel)
		sections = append(sections, split)
	} else {
		// Full width pair
		sections = append(sections, a.pairView.View())
	}

	// Status bar
	sections = append(sections, a.statusBar.View())

	return strings.Join(sections, "\n")
}

func (a *App) renderEditLayout() string {
	var sections []string
	sections = append(sections, a.editorView.View())
	sections = append(sections, a.statusBar.View())
	return strings.Join(sections, "\n")
}

func (a *App) renderFormLayout() string {
	// Render the normal chat view as the background
	var sections []string
	sections = append(sections, a.renderHeader())
	sections = append(sections, a.chat.ViewChat())
	if stats := a.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}
	sections = append(sections, a.statusBar.View())
	background := strings.Join(sections, "\n")

	// Dim the background
	backgroundLines := strings.Split(background, "\n")
	for i, line := range backgroundLines {
		backgroundLines[i] = lipgloss.NewStyle().Foreground(a.theme.TextMuted).Render(line)
	}

	// Render the form
	if a.formView == nil {
		return strings.Join(backgroundLines, "\n")
	}

	formContent := a.formView.View()
	formLines := strings.Split(formContent, "\n")
	formHeight := len(formLines)

	// Calculate form width for centering
	maxFormWidth := 0
	for _, line := range formLines {
		if w := lipgloss.Width(line); w > maxFormWidth {
			maxFormWidth = w
		}
	}

	// Center the form vertically and horizontally
	startLine := (len(backgroundLines) - formHeight) / 2
	if startLine < 2 {
		startLine = 2
	}

	leftPad := (a.width - maxFormWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

	// Overlay the form on the dimmed background
	result := make([]string, len(backgroundLines))
	copy(result, backgroundLines)

	for i, line := range formLines {
		lineIdx := startLine + i
		if lineIdx >= 0 && lineIdx < len(result) {
			result[lineIdx] = padding + line
		}
	}

	return strings.Join(result, "\n")
}

func (a *App) renderHeader() string {
	// Context-aware header based on ALC state
	if a.alcState != nil && a.alcState.Context != alc.Chat {
		return a.renderContextHeader()
	}

	// Default Chat mode header
	logo := lipgloss.NewStyle().Foreground(a.theme.Primary).Bold(true).Render("🔥🗝️🔥 Hecate")

	modelSection := ""
	if modelName := a.chat.ActiveModelName(); modelName != "" {
		modelSection = a.styles.Subtle.Render("  ·  " + modelName)
	}

	daemonSection := "  ·  "
	switch a.daemonStatus {
	case "healthy", "ok":
		daemonSection += a.styles.StatusOK.Render("●") + a.styles.Subtle.Render(" daemon")
	case "degraded":
		daemonSection += a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon")
	default:
		daemonSection += a.styles.Subtle.Render("○ daemon")
	}

	titleSection := ""
	if a.conversationTitle != "" {
		titleSection = a.styles.Subtle.Render("  ·  ") + a.styles.CardValue.Render(a.conversationTitle)
	}

	left := logo + modelSection + daemonSection + titleSection

	return lipgloss.NewStyle().Width(a.width).Padding(0, 1).Render(left)
}

// renderContextHeader renders the context-aware header for Torch/Cartwheel modes.
func (a *App) renderContextHeader() string {
	var parts []string

	// Torch name (always present in non-Chat mode)
	if a.alcState.Torch != nil {
		torchStyle := lipgloss.NewStyle().Foreground(a.theme.Warning).Bold(true)
		parts = append(parts, torchStyle.Render("🔥 "+a.alcState.Torch.Name))
	}

	// Cartwheel name and phase (if in Cartwheel mode)
	if a.alcState.Context == alc.Cartwheel && a.alcState.Cartwheel != nil {
		cartwheelStyle := lipgloss.NewStyle().Foreground(a.theme.Secondary)
		parts = append(parts, cartwheelStyle.Render("🎡 "+a.alcState.Cartwheel.Name))

		// Phase badge
		if phase := a.alcState.Cartwheel.CurrentPhase; phase != "" {
			phaseStyle := a.phaseStyle(string(phase))
			parts = append(parts, phaseStyle.Render("📍 "+strings.ToUpper(string(phase))))
		}

		// Model indicator moves to header in Cartwheel mode
		if modelName := a.chat.ActiveModelName(); modelName != "" {
			name := modelName
			if len(name) > 15 {
				name = name[:12] + "..."
			}
			parts = append(parts, a.styles.Subtle.Render("🤖 "+name))
		}
	}

	left := strings.Join(parts, a.styles.Subtle.Render(" › "))

	return lipgloss.NewStyle().Width(a.width).Padding(0, 1).Render(left)
}

// phaseStyle returns a style for ALC phase badges.
func (a *App) phaseStyle(phase string) lipgloss.Style {
	switch strings.ToLower(phase) {
	case "dna":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true) // Purple
	case "anp":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Bold(true) // Blue
	case "tni":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#059669")).Bold(true) // Green
	case "dno":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")).Bold(true) // Red
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Bold(true) // Gray
	}
}

func (a *App) renderCommandLine() string {
	return lipgloss.NewStyle().
		Width(a.width).
		Padding(0, 1).
		Background(a.theme.BgInput).
		Render(a.cmdInput.View())
}

func (a *App) chatAreaHeight() int {
	headerHeight := 2
	statusBarHeight := 1
	inputHeight := 0

	switch a.mode {
	case modes.Insert:
		inputHeight = 3 // 1 row + border
	case modes.Command:
		inputHeight = 1
	}

	statsHeight := 1
	h := a.height - headerHeight - statusBarHeight - inputHeight - statsHeight
	if h < 5 {
		h = 5
	}
	return h
}

// browseWidth returns the width for the browse overlay.
// Split pane on wide terminals (>= 100 cols), full width on narrow.
func (a *App) browseWidth() int {
	if a.width >= 100 {
		return a.width / 2
	}
	return a.width - 4
}

// browseHeight returns the height for the browse overlay.
func (a *App) browseHeight() int {
	return a.height - 4 // header + status bar + padding
}

// pairWidth returns the width for the pair overlay.
func (a *App) pairWidth() int {
	if a.width >= 100 {
		return a.width / 2
	}
	return a.width - 4
}

// pairHeight returns the height for the pair overlay.
func (a *App) pairHeight() int {
	return a.height - 4
}
