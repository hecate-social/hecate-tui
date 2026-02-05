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
		m.filtered = msg.capabilities
		m.err = msg.err
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
	if m.searchQuery == "" {
		m.filtered = m.capabilities
		return
	}
	query := strings.ToLower(m.searchQuery)
	var filtered []client.Capability
	for _, cap := range m.capabilities {
		if strings.Contains(strings.ToLower(cap.MRI), query) ||
			strings.Contains(strings.ToLower(cap.Description), query) ||
			containsTag(cap.Tags, query) {
			filtered = append(filtered, cap)
		}
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
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.searchInput.Width = width - 10
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

	// Search bar
	if m.mode == ModeSearch || m.searchQuery != "" {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Discovering...")
		return m.wrapPanel(b.String())
	}

	if m.err != nil {
		b.WriteString(s.Error.Render("Error: " + m.err.Error()))
		return m.wrapPanel(b.String())
	}

	if len(m.capabilities) == 0 {
		b.WriteString(s.Subtle.Render("No capabilities on the mesh."))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Press r to refresh."))
		return m.wrapPanel(b.String())
	}

	if len(m.filtered) == 0 && m.searchQuery != "" {
		b.WriteString(s.Error.Render("No matches for \"" + m.searchQuery + "\""))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Press / to search again"))
		return m.wrapPanel(b.String())
	}

	// Column headers
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		s.Subtle.Width(32).Render("CAPABILITY"),
		s.Subtle.Width(10).Render("SOURCE"),
		s.Subtle.Render("TAGS"),
	)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(s.Divider.Render(strings.Repeat("─", m.width-6)))
	b.WriteString("\n")

	// Visible rows
	visibleRows := m.height - 10
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
			row = s.Selected.Width(m.width - 6).Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > visibleRows {
		b.WriteString(s.Subtle.Render(fmt.Sprintf("\n  %d / %d", m.selected+1, len(m.filtered))))
	}

	return m.wrapPanel(b.String())
}

func (m Model) renderSearchBar() string {
	s := m.styles
	if m.mode == ModeSearch {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.Primary).
			Padding(0, 1).
			Width(m.width - 6).
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

	b.WriteString(s.CardLabel.Render("  MRI:"))
	b.WriteString(lipgloss.NewStyle().Foreground(m.theme.Secondary).Bold(true).Render(cap.MRI))
	b.WriteString("\n\n")

	name := formatCapName(cap.MRI)
	b.WriteString(s.CardLabel.Render("  Name:"))
	b.WriteString(s.CardValue.Render(name))
	b.WriteString("\n")

	b.WriteString(s.CardLabel.Render("  Source:"))
	if isLocal(*cap) {
		b.WriteString(s.StatusOK.Render("local"))
	} else {
		b.WriteString(s.Subtle.Render("remote"))
	}
	b.WriteString("\n")

	if cap.AgentIdentity != "" {
		b.WriteString(s.CardLabel.Render("  Agent:"))
		b.WriteString(s.CardValue.Render(cap.AgentIdentity))
		b.WriteString("\n")
	}

	if cap.Description != "" {
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Description:"))
		b.WriteString("\n  ")
		b.WriteString(s.CardValue.Width(m.width - 10).Render(cap.Description))
		b.WriteString("\n")
	}

	if len(cap.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Tags:"))
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

	return m.wrapPanel(b.String())
}

func (m Model) wrapPanel(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(content)
}

func (m Model) fetchCapabilities() tea.Msg {
	caps, err := m.client.DiscoverCapabilities("", "", 100)
	return capabilitiesMsg{capabilities: caps, err: err}
}

func isLocal(cap client.Capability) bool {
	return strings.Contains(cap.MRI, "local") || cap.AgentIdentity == ""
}

func formatCapName(mri string) string {
	parts := strings.Split(mri, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[1:], "/")
	}
	return mri
}
