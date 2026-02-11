package llm

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/scaffold"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// View renders the LLM studio content area.
func (s *Studio) View() string {
	if s.width == 0 {
		return "Loading..."
	}

	// Browse mode uses modal overlay
	if s.mode == modes.Browse && s.browseReady {
		return s.renderBrowseLayout()
	}

	// Pair mode uses split or full layout
	if s.mode == modes.Pair && s.pairReady {
		return s.renderPairLayout()
	}

	// Edit mode takes full content area
	if s.mode == modes.Edit && s.editorReady {
		return s.editorView.View()
	}

	// Form mode overlays the chat
	if s.mode == modes.Form && s.formReady {
		return s.renderFormLayout()
	}

	var sections []string

	// Chat area
	sections = append(sections, s.chat.ViewChat())

	// Stats/streaming indicator
	if stats := s.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}

	// Error
	if errView := s.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}

	// Input area (mode-dependent)
	if s.mode == modes.Insert {
		sections = append(sections, s.chat.ViewInput())
	}

	content := strings.Join(sections, "\n")

	// Overlay tool approval prompt if there's a pending approval
	if s.chat.HasPendingApproval() && s.approvalPrompt != nil {
		content = s.renderWithApprovalOverlay(content)
	}

	return content
}

func (s *Studio) renderWithApprovalOverlay(content string) string {
	call := s.chat.PendingToolCall()
	if call == nil {
		return content
	}

	registry := s.toolExecutor.Registry()
	tool, _, ok := registry.Get(call.Name)
	if !ok {
		tool = llmtools.Tool{
			Name:        call.Name,
			Description: "Unknown tool",
			Category:    llmtools.CategorySystem,
		}
	}

	dialogWidth := 60
	if s.width > 80 {
		dialogWidth = 70
	}
	if s.width < 70 {
		dialogWidth = s.width - 4
	}
	s.approvalPrompt.SetWidth(dialogWidth)

	prompt := s.approvalPrompt.Render(tool, *call)

	promptLines := strings.Split(prompt, "\n")
	promptHeight := len(promptLines)

	contentLines := strings.Split(content, "\n")
	startLine := (len(contentLines) - promptHeight) / 2
	if startLine < 0 {
		startLine = 0
	}

	maxPromptWidth := 0
	for _, line := range promptLines {
		if w := lipgloss.Width(line); w > maxPromptWidth {
			maxPromptWidth = w
		}
	}
	leftPad := (s.width - maxPromptWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

	for i, line := range promptLines {
		lineIdx := startLine + i
		if lineIdx >= 0 && lineIdx < len(contentLines) {
			contentLines[lineIdx] = padding + line
		}
	}

	return strings.Join(contentLines, "\n")
}

func (s *Studio) renderBrowseLayout() string {
	var sections []string
	sections = append(sections, s.chat.ViewChat())
	if stats := s.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}
	if errView := s.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}
	background := strings.Join(sections, "\n")

	backgroundLines := strings.Split(background, "\n")
	for i, line := range backgroundLines {
		backgroundLines[i] = lipgloss.NewStyle().Foreground(s.ctx.Theme.TextMuted).Render(line)
	}

	modal := s.browseView.View()
	modalLines := strings.Split(modal, "\n")

	result := make([]string, len(backgroundLines))
	copy(result, backgroundLines)

	for i, line := range modalLines {
		if i < len(result) && strings.TrimSpace(line) != "" {
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

func (s *Studio) renderPairLayout() string {
	if s.width >= 100 {
		chatWidth := s.width - s.pairWidth() - 1
		chatHeight := s.pairHeight()

		chatContent := s.chat.ViewChat()
		dimmedChat := lipgloss.NewStyle().
			Width(chatWidth).
			Height(chatHeight).
			Foreground(s.ctx.Theme.TextMuted).
			Render(chatContent)

		pairPanel := s.pairView.View()

		sep := lipgloss.NewStyle().
			Foreground(s.ctx.Theme.Border).
			Render("│")

		return lipgloss.JoinHorizontal(lipgloss.Top, dimmedChat, sep, pairPanel)
	}

	return s.pairView.View()
}

func (s *Studio) renderFormLayout() string {
	var sections []string
	sections = append(sections, s.chat.ViewChat())
	if stats := s.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}
	background := strings.Join(sections, "\n")

	backgroundLines := strings.Split(background, "\n")
	for i, line := range backgroundLines {
		backgroundLines[i] = lipgloss.NewStyle().Foreground(s.ctx.Theme.TextMuted).Render(line)
	}

	if s.formView == nil {
		return strings.Join(backgroundLines, "\n")
	}

	formContent := s.formView.View()
	formLines := strings.Split(formContent, "\n")
	formHeight := len(formLines)

	maxFormWidth := 0
	for _, line := range formLines {
		if w := lipgloss.Width(line); w > maxFormWidth {
			maxFormWidth = w
		}
	}

	startLine := (len(backgroundLines) - formHeight) / 2
	if startLine < 2 {
		startLine = 2
	}

	leftPad := (s.width - maxFormWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

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

// Dimension helpers

func (s *Studio) chatAreaHeight() int {
	inputHeight := 0
	if s.mode == modes.Insert {
		inputHeight = 3 // 1 row + border
	}

	statsHeight := 1
	h := s.height - inputHeight - statsHeight
	if h < 5 {
		h = 5
	}
	return h
}

func (s *Studio) pairWidth() int {
	if s.width >= 100 {
		return s.width / 2
	}
	return s.width - 4
}

func (s *Studio) pairHeight() int {
	return s.height - 1
}

func (s *Studio) editorHeight() int {
	return s.height
}

// buildVentureScaffoldMsg creates a venture scaffold and returns a VentureCreatedMsg.
func buildVentureScaffoldMsg(st *theme.Styles, ventureID, name, brief string, initiatedAt int64, initiatedBy, path string) tea.Msg {
	manifest := scaffold.VentureManifest{
		VentureID:   ventureID,
		Name:        name,
		Brief:       brief,
		Root:        path,
		InitiatedAt: initiatedAt,
		InitiatedBy: initiatedBy,
	}

	result := scaffold.Scaffold(path, manifest)

	var b strings.Builder
	b.WriteString(st.StatusOK.Render("Venture Initiated"))
	b.WriteString("\n\n")
	b.WriteString(st.CardTitle.Render("Venture: " + name))
	b.WriteString("\n")
	b.WriteString(st.CardLabel.Render("    ID: "))
	b.WriteString(st.CardValue.Render(ventureID))
	b.WriteString("\n")
	b.WriteString(st.CardLabel.Render("  Root: "))
	b.WriteString(st.Subtle.Render(path))
	b.WriteString("\n\n")

	b.WriteString(st.CardTitle.Render("Scaffolded:"))
	b.WriteString("\n")

	if result.Success {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render(".hecate/torch.json"))
		b.WriteString("\n")
	}
	if result.AgentsCloned {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render(".hecate/agents/"))
		b.WriteString("\n")
	}
	if result.ReadmeCreated {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render("README.md"))
		b.WriteString("\n")
	}
	if result.ChangelogCreated {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render("CHANGELOG.md"))
		b.WriteString("\n")
	}

	if result.GitInitialized {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render("git init"))
		b.WriteString("\n")
	}

	if result.GitCommitted {
		b.WriteString(st.StatusOK.Render("  ✓ "))
		b.WriteString(st.Subtle.Render("git commit"))
		b.WriteString("\n")
	}

	for _, warn := range result.Warnings {
		b.WriteString(st.StatusWarning.Render("  ⚠ " + warn))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(st.Subtle.Render("Next: gh repo create --public --source=. --push"))

	return commands.VentureCreatedMsg{Path: path, Message: b.String()}
}
