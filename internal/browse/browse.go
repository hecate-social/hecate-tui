package browse

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// ViewMode represents the internal browse state.
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeSearch
	ModeDetail
)

// Tab represents a filter tab.
type Tab int

const (
	TabAll Tab = iota
	TabLocal
	TabRemote
	TabLLM
	TabTools
)

func (t Tab) String() string {
	switch t {
	case TabLocal:
		return "Local"
	case TabRemote:
		return "Remote"
	case TabLLM:
		return "LLM"
	case TabTools:
		return "Tools"
	default:
		return "All"
	}
}

// SelectModelMsg is emitted when user selects an LLM model.
type SelectModelMsg struct {
	ModelName string
}

// Model is the Browse mode overlay.
type Model struct {
	client  *client.Client
	theme   *theme.Theme
	styles  *theme.Styles
	width   int
	height  int
	loading bool
	spinner spinner.Model

	// Data
	capabilities []client.Capability
	filtered     []client.Capability
	selected     int
	err          error

	// Tabs
	activeTab Tab

	// Search
	mode        ViewMode
	searchInput textinput.Model
	searchQuery string

	// Detail
	detailCap *client.Capability
}

// capabilitiesMsg carries fetched capabilities.
type capabilitiesMsg struct {
	capabilities []client.Capability
	err          error
}

// New creates a Browse mode model.
func New(c *client.Client, t *theme.Theme, s *theme.Styles) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(t.Primary)

	ti := textinput.New()
	ti.Placeholder = "Filter capabilities..."
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	ti.Prompt = "/"

	return Model{
		client:      c,
		theme:       t,
		styles:      s,
		loading:     true,
		spinner:     sp,
		mode:        ModeList,
		searchInput: ti,
	}
}

// Init starts loading capabilities.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchCapabilities,
	)
}

// Update handles messages routed from the app.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case capabilitiesMsg:
		m.loading = false
		m.capabilities = msg.capabilities
		m.err = msg.err
		m.applyFilter()
	}

	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// HandleKey processes a keypress in Browse mode. Returns true if the key was consumed.
// Returns a tea.Cmd if an action was triggered.
func (m *Model) HandleKey(key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch m.mode {
	case ModeSearch:
		return m.handleSearchKey(key, msg)
	case ModeDetail:
		return m.handleDetailKey(key)
	default:
		return m.handleListKey(key, msg)
	}
}

func (m *Model) handleListKey(key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "j", "down":
		if m.selected < len(m.filtered)-1 {
			m.selected++
		}
		return true, nil
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
		return true, nil
	case "h", "left":
		m.prevTab()
		return true, nil
	case "l", "right":
		m.nextTab()
		return true, nil
	case "tab":
		m.nextTab()
		return true, nil
	case "shift+tab":
		m.prevTab()
		return true, nil
	case "g":
		m.selected = 0
		return true, nil
	case "G":
		if len(m.filtered) > 0 {
			m.selected = len(m.filtered) - 1
		}
		return true, nil
	case "enter":
		if len(m.filtered) > 0 && m.selected < len(m.filtered) {
			cap := m.filtered[m.selected]
			// If it's an LLM model, select it and close browse
			if isLLM(cap) {
				modelName := extractModelName(cap.MRI)
				return false, func() tea.Msg {
					return SelectModelMsg{ModelName: modelName}
				}
			}
			// Otherwise show detail
			m.detailCap = &cap
			m.mode = ModeDetail
		}
		return true, nil
	case "/":
		m.mode = ModeSearch
		m.searchInput.Focus()
		return true, textinput.Blink
	case "r":
		m.loading = true
		m.searchQuery = ""
		m.searchInput.SetValue("")
		return true, m.fetchCapabilities
	case "esc":
		// Esc exits Browse mode — handled by app
		return false, nil
	}
	return false, nil
}

func (m *Model) nextTab() {
	m.activeTab = (m.activeTab + 1) % 5
	m.applyFilter()
	m.selected = 0
}

func (m *Model) prevTab() {
	if m.activeTab == 0 {
		m.activeTab = 4
	} else {
		m.activeTab--
	}
	m.applyFilter()
	m.selected = 0
}

func (m *Model) handleSearchKey(key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "esc":
		m.mode = ModeList
		m.searchInput.Blur()
		return true, nil
	case "enter":
		m.searchQuery = m.searchInput.Value()
		m.applyFilter()
		m.mode = ModeList
		m.searchInput.Blur()
		m.selected = 0
		return true, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Live filter
		m.searchQuery = m.searchInput.Value()
		m.applyFilter()
		m.selected = 0
		return true, cmd
	}
}

func (m *Model) handleDetailKey(key string) (bool, tea.Cmd) {
	switch key {
	case "esc", "q", "enter":
		m.mode = ModeList
		m.detailCap = nil
		return true, nil
	case "r":
		m.loading = true
		m.mode = ModeList
		m.detailCap = nil
		return true, m.fetchCapabilities
	}
	return true, nil
}

func (m *Model) applyFilter() {
	filtered := make([]client.Capability, 0)

	// First filter by tab
	for _, cap := range m.capabilities {
		switch m.activeTab {
		case TabLocal:
			if !isLocal(cap) {
				continue
			}
		case TabRemote:
			if isLocal(cap) {
				continue
			}
		case TabLLM:
			if !isLLM(cap) {
				continue
			}
		case TabTools:
			if !isTool(cap) {
				continue
			}
		}
		filtered = append(filtered, cap)
	}

	// Then filter by search query
	if m.searchQuery != "" {
		query := strings.ToLower(m.searchQuery)
		searchFiltered := make([]client.Capability, 0)
		for _, cap := range filtered {
			if strings.Contains(strings.ToLower(cap.MRI), query) ||
				strings.Contains(strings.ToLower(cap.Description), query) ||
				containsTag(cap.Tags, query) {
				searchFiltered = append(searchFiltered, cap)
			}
		}
		filtered = searchFiltered
	}

	m.filtered = filtered
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// SetSize updates the browse panel dimensions.
// For modal mode, these are the terminal dimensions (we calculate modal size internally).
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Modal width is 70% of terminal, capped at 80 chars
	modalW := width * 70 / 100
	if modalW > 80 {
		modalW = 80
	}
	if modalW < 50 {
		modalW = width - 4
	}
	m.searchInput.Width = modalW - 10
}

// View renders the browse overlay panel.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case ModeDetail:
		return m.renderDetail()
	default:
		return m.renderList()
	}
}

func (m Model) renderList() string {
	s := m.styles
	var b strings.Builder

	// Title
	b.WriteString(s.CardTitle.Render("Browse Capabilities"))
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Search bar
	if m.mode == ModeSearch || m.searchQuery != "" {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Discovering...")
		return m.wrapModal(b.String())
	}

	if m.err != nil {
		b.WriteString(s.Error.Render("Error: " + m.err.Error()))
		return m.wrapModal(b.String())
	}

	if len(m.capabilities) == 0 {
		b.WriteString(s.Subtle.Render("No capabilities on the mesh."))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Press r to refresh."))
		return m.wrapModal(b.String())
	}

	if len(m.filtered) == 0 {
		if m.searchQuery != "" {
			b.WriteString(s.Subtle.Render("No matches for \"" + m.searchQuery + "\""))
		} else if m.activeTab != TabAll {
			b.WriteString(s.Subtle.Render("No " + strings.ToLower(m.activeTab.String()) + " capabilities found."))
		} else {
			b.WriteString(s.Subtle.Render("No capabilities found."))
		}
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Try another tab or press r to refresh."))
		return m.wrapModal(b.String())
	}

	contentW := m.contentWidth()

	// Column headers
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		s.Subtle.Width(32).Render("CAPABILITY"),
		s.Subtle.Width(10).Render("SOURCE"),
		s.Subtle.Render("TAGS"),
	)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(s.Divider.Render(strings.Repeat("─", contentW)))
	b.WriteString("\n")

	// Visible rows based on modal content height
	visibleRows := m.contentHeight() - 8 // Title, search, header, divider, footer
	if visibleRows < 3 {
		visibleRows = 3
	}

	startIdx := 0
	if m.selected >= visibleRows {
		startIdx = m.selected - visibleRows + 1
	}
	endIdx := startIdx + visibleRows
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		cap := m.filtered[i]

		indicator := "  "
		if i == m.selected {
			indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("▸ ")
		}

		name := formatCapName(cap.MRI)
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		source := "remote"
		sourceStyle := s.Subtle
		if isLocal(cap) {
			source = "local"
			sourceStyle = s.StatusOK
		}

		tags := strings.Join(cap.Tags, ", ")
		if len(tags) > 25 {
			tags = tags[:22] + "..."
		}

		row := indicator +
			lipgloss.NewStyle().Width(28).Foreground(m.theme.Text).Render(name) +
			sourceStyle.Width(10).Render(source) +
			s.Subtle.Render(tags)

		if i == m.selected {
			row = s.Selected.Width(contentW).Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > visibleRows {
		b.WriteString(s.Subtle.Render(fmt.Sprintf("\n  %d / %d", m.selected+1, len(m.filtered))))
	}

	// Help hint
	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("  ←/→ tabs  ↑/↓ navigate  / search  ⏎ select  r refresh  esc close"))

	return m.wrapModal(b.String())
}

func (m Model) renderTabs() string {
	tabs := []Tab{TabAll, TabLocal, TabRemote, TabLLM, TabTools}
	var parts []string

	for _, tab := range tabs {
		label := tab.String()
		if tab == m.activeTab {
			// Active tab
			parts = append(parts, lipgloss.NewStyle().
				Foreground(m.theme.Primary).
				Bold(true).
				Padding(0, 1).
				Render("["+label+"]"))
		} else {
			// Inactive tab
			parts = append(parts, lipgloss.NewStyle().
				Foreground(m.theme.TextMuted).
				Padding(0, 1).
				Render(label))
		}
	}

	return "  " + strings.Join(parts, " ")
}

func (m Model) renderSearchBar() string {
	s := m.styles
	if m.mode == ModeSearch {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.Primary).
			Padding(0, 1).
			Width(m.contentWidth()).
			Render(m.searchInput.View())
	}

	// Show current filter
	count := fmt.Sprintf(" (%d of %d)", len(m.filtered), len(m.capabilities))
	return s.Subtle.Render("  /" + m.searchQuery + count)
}

func (m Model) renderDetail() string {
	if m.detailCap == nil {
		return ""
	}

	s := m.styles
	cap := m.detailCap
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Capability Details"))
	b.WriteString("\n\n")

	b.WriteString(s.CardLabel.Render("MRI: "))
	b.WriteString(lipgloss.NewStyle().Foreground(m.theme.Secondary).Bold(true).Render(cap.MRI))
	b.WriteString("\n\n")

	name := formatCapName(cap.MRI)
	b.WriteString(s.CardLabel.Render("Name: "))
	b.WriteString(s.CardValue.Render(name))
	b.WriteString("\n")

	b.WriteString(s.CardLabel.Render("Source: "))
	if isLocal(*cap) {
		b.WriteString(s.StatusOK.Render("local"))
	} else {
		b.WriteString(s.Subtle.Render("remote"))
	}
	b.WriteString("\n")

	if cap.AgentIdentity != "" {
		b.WriteString(s.CardLabel.Render("Agent: "))
		b.WriteString(s.CardValue.Render(cap.AgentIdentity))
		b.WriteString("\n")
	}

	if cap.Description != "" {
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Description: "))
		b.WriteString("\n  ")
		b.WriteString(s.CardValue.Width(m.contentWidth() - 4).Render(cap.Description))
		b.WriteString("\n")
	}

	if len(cap.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Tags: "))
		b.WriteString("\n  ")
		for _, tag := range cap.Tags {
			b.WriteString(lipgloss.NewStyle().
				Background(m.theme.BgCard).
				Foreground(m.theme.Text).
				Padding(0, 1).
				MarginRight(1).
				Render(tag))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.Subtle.Render("  Esc: back to list"))

	return m.wrapModal(b.String())
}

// modalWidth returns the width for the modal dialog.
func (m Model) modalWidth() int {
	w := m.width * 70 / 100
	if w > 80 {
		w = 80
	}
	if w < 50 {
		w = m.width - 4
	}
	return w
}

// modalHeight returns the height for the modal dialog.
func (m Model) modalHeight() int {
	h := m.height * 70 / 100
	if h > 30 {
		h = 30
	}
	if h < 15 {
		h = m.height - 6
	}
	return h
}

// contentWidth returns the usable width inside the modal (minus border and padding).
func (m Model) contentWidth() int {
	return m.modalWidth() - 6 // 2 border + 4 padding
}

// contentHeight returns the usable height inside the modal.
func (m Model) contentHeight() int {
	return m.modalHeight() - 4 // 2 border + 2 padding
}

func (m Model) wrapModal(content string) string {
	modalW := m.modalWidth()
	modalH := m.modalHeight()

	// Create modal with border and shadow effect
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Background(m.theme.BgCard).
		Padding(1, 2).
		Width(modalW).
		Height(modalH).
		Render(content)

	// Center the modal horizontally and vertically
	leftPad := (m.width - modalW) / 2
	topPad := (m.height - modalH) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	if topPad < 0 {
		topPad = 0
	}

	// Build the centered modal with top padding
	var result strings.Builder
	for i := 0; i < topPad; i++ {
		result.WriteString("\n")
	}
	// Add horizontal padding to each line
	padding := strings.Repeat(" ", leftPad)
	for _, line := range strings.Split(modal, "\n") {
		result.WriteString(padding)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

func (m Model) fetchCapabilities() tea.Msg {
	caps, err := m.client.DiscoverCapabilities("", "", 100)
	return capabilitiesMsg{capabilities: caps, err: err}
}

func isLocal(cap client.Capability) bool {
	return strings.Contains(cap.MRI, "local") || cap.AgentIdentity == ""
}

func isLLM(cap client.Capability) bool {
	// Check tags for "llm" or "model"
	for _, tag := range cap.Tags {
		t := strings.ToLower(tag)
		if t == "llm" || t == "model" || t == "chat" {
			return true
		}
	}
	// Check MRI for llm patterns
	mri := strings.ToLower(cap.MRI)
	return strings.Contains(mri, "llm") || strings.Contains(mri, "model")
}

func isTool(cap client.Capability) bool {
	for _, tag := range cap.Tags {
		t := strings.ToLower(tag)
		if t == "tool" || t == "function" {
			return true
		}
	}
	mri := strings.ToLower(cap.MRI)
	return strings.Contains(mri, "tool")
}

func extractModelName(mri string) string {
	// MRI format: mri:llm:local/modelname or similar
	parts := strings.Split(mri, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	// Try splitting by colon
	parts = strings.Split(mri, ":")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return mri
}

func formatCapName(mri string) string {
	parts := strings.Split(mri, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[1:], "/")
	}
	return mri
}
