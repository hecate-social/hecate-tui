package browse

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeSearch
	ModeDetail
)

// Model is the browse view model
type Model struct {
	client       *client.Client
	width        int
	height       int
	focused      bool
	loading      bool
	spinner      spinner.Model
	capabilities []client.Capability
	filtered     []client.Capability
	selected     int
	err          error

	// Search
	mode        ViewMode
	searchInput textinput.Model
	searchQuery string

	// Detail view
	detailCap *client.Capability
}

// New creates a new browse view
func New(c *client.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	ti := textinput.New()
	ti.Placeholder = "Search capabilities..."
	ti.CharLimit = 100
	ti.Width = 40

	return Model{
		client:      c,
		spinner:     s,
		loading:     true,
		mode:        ModeList,
		searchInput: ti,
	}
}

// Messages
type capabilitiesMsg struct {
	capabilities []client.Capability
	err          error
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchCapabilities,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Handle based on current mode
		switch m.mode {
		case ModeSearch:
			return m.updateSearch(msg)
		case ModeDetail:
			return m.updateDetail(msg)
		default:
			return m.updateList(msg)
		}

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

	// Update search input when in search mode
	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.filtered)-1 {
			m.selected++
		}
	case "home", "g":
		m.selected = 0
	case "end", "G":
		if len(m.filtered) > 0 {
			m.selected = len(m.filtered) - 1
		}
	case "r":
		m.loading = true
		m.searchQuery = ""
		m.searchInput.SetValue("")
		return m, m.fetchCapabilities
	case "enter":
		if len(m.filtered) > 0 && m.selected < len(m.filtered) {
			cap := m.filtered[m.selected]
			m.detailCap = &cap
			m.mode = ModeDetail
		}
	case "/":
		m.mode = ModeSearch
		m.searchInput.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeList
		m.searchInput.Blur()
		return m, nil
	case "enter":
		m.searchQuery = m.searchInput.Value()
		m.applyFilter()
		m.mode = ModeList
		m.searchInput.Blur()
		m.selected = 0
		return m, nil
	}

	// Update the text input
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Live filter as user types
	m.searchQuery = m.searchInput.Value()
	m.applyFilter()
	m.selected = 0

	return m, cmd
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = ModeList
		m.detailCap = nil
	case "r":
		// Refresh while in detail view
		m.loading = true
		m.mode = ModeList
		m.detailCap = nil
		return m, m.fetchCapabilities
	}
	return m, nil
}

func (m *Model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.capabilities
		return
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []client.Capability
	for _, cap := range m.capabilities {
		// Search in MRI, tags, and description
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

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case ModeDetail:
		return m.renderDetail()
	default:
		return m.renderListView()
	}
}

func (m Model) renderListView() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("Browse Capabilities"))
	b.WriteString("\n\n")

	// Search bar (always visible when search mode or has query)
	if m.mode == ModeSearch || m.searchQuery != "" {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Discovering capabilities...")
		return m.wrapInBox(b.String())
	}

	if m.err != nil {
		b.WriteString(styles.StatusError.Render("Error: " + m.err.Error()))
		return m.wrapInBox(b.String())
	}

	if len(m.capabilities) == 0 {
		b.WriteString(m.renderEmpty())
		return m.wrapInBox(b.String())
	}

	if len(m.filtered) == 0 && m.searchQuery != "" {
		b.WriteString(NoResultsStyle.Render("No capabilities match \"" + m.searchQuery + "\""))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(styles.Muted).Render("Press / to search again or r to refresh"))
		return m.wrapInBox(b.String())
	}

	// Capability list
	b.WriteString(m.renderCapabilities())

	return m.wrapInBox(b.String())
}

func (m Model) renderSearchBar() string {
	var parts []string

	label := SearchLabelStyle.Render("/")
	parts = append(parts, label)

	if m.mode == ModeSearch {
		parts = append(parts, m.searchInput.View())
	} else {
		// Show current query
		query := SearchInputStyle.Render(m.searchQuery)
		count := lipgloss.NewStyle().Foreground(styles.Muted).Render(
			" (" + formatCount(len(m.filtered), len(m.capabilities)) + ")",
		)
		parts = append(parts, query, count)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left, parts...)
	return SearchContainerStyle.Width(m.width - 8).Render(content)
}

func formatCount(filtered, total int) string {
	if filtered == total {
		return fmt.Sprintf("%d total", total)
	}
	return fmt.Sprintf("%d of %d", filtered, total)
}

func (m Model) renderEmpty() string {
	icon := EmptyIconStyle.Render(`
    .--.
   |o_o |
   |:_/ |
  //   \ \
 (|     | )
/'\_   _/'\
\___)=(___/
`)
	message := EmptyStateStyle.Render("\nNo capabilities discovered on the mesh.\n\nCapabilities appear here when agents announce them.\nPress r to refresh.")

	return lipgloss.JoinVertical(lipgloss.Center, icon, message)
}

func (m Model) renderCapabilities() string {
	var rows []string

	// Header
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		ColumnNameStyle.Render("CAPABILITY"),
		ColumnSourceStyle.Render("SOURCE"),
		ColumnTagsStyle.Render("TAGS"),
	)
	rows = append(rows, HeaderStyle.Render(header))
	rows = append(rows, lipgloss.NewStyle().Foreground(styles.Border).Render(strings.Repeat("─", m.width-8)))

	// Calculate visible rows based on height
	visibleRows := m.height - 12
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Determine scroll offset
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

		// Indicator
		indicator := "  "
		if i == m.selected {
			indicator = lipgloss.NewStyle().Foreground(styles.Primary).Render("> ")
		}

		// Format name from MRI
		name := formatCapabilityName(cap.MRI)
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Source
		isLocalCap := isLocal(cap)
		source := SourceIndicator(isLocalCap)

		// Format tags
		tags := strings.Join(cap.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}
		tagsRender := lipgloss.NewStyle().Foreground(styles.Text).Render(tags)

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			indicator,
			lipgloss.NewStyle().Width(28).Render(name),
			lipgloss.NewStyle().Width(10).Render(source),
			tagsRender,
		)

		// Highlight selected
		if i == m.selected && m.focused {
			row = SelectedItemStyle.Width(m.width - 8).Render(row)
		} else if i == m.selected {
			row = SelectedUnfocusedStyle.Width(m.width - 8).Render(row)
		} else {
			row = ItemStyle.Width(m.width - 8).Render(row)
		}

		rows = append(rows, row)
	}

	// Scroll indicator
	if len(m.filtered) > visibleRows {
		scrollInfo := lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true).
			Render("\n  " + formatScrollPosition(m.selected+1, len(m.filtered)))
		rows = append(rows, scrollInfo)
	}

	return strings.Join(rows, "\n")
}

func formatScrollPosition(current, total int) string {
	return fmt.Sprintf("%d / %d", current, total)
}

func (m Model) renderDetail() string {
	if m.detailCap == nil {
		return m.wrapInBox("No capability selected")
	}

	cap := m.detailCap
	var b strings.Builder

	// Header
	title := DetailTitleStyle.Render("Capability Details")
	b.WriteString(title)
	b.WriteString("\n\n")

	// MRI (full)
	b.WriteString(DetailLabelStyle.Render("MRI:"))
	b.WriteString(DetailMRIStyle.Render(cap.MRI))
	b.WriteString("\n\n")

	// Name (extracted)
	name := formatCapabilityName(cap.MRI)
	b.WriteString(DetailLabelStyle.Render("Name:"))
	b.WriteString(DetailValueStyle.Render(name))
	b.WriteString("\n")

	// Source
	isLocalCap := isLocal(*cap)
	b.WriteString(DetailLabelStyle.Render("Source:"))
	if isLocalCap {
		b.WriteString(LocalSourceStyle.Render("local (this agent)"))
	} else {
		b.WriteString(RemoteSourceStyle.Render("remote (mesh)"))
	}
	b.WriteString("\n")

	// Agent Identity
	if cap.AgentIdentity != "" {
		b.WriteString(DetailLabelStyle.Render("Agent:"))
		b.WriteString(DetailValueStyle.Render(cap.AgentIdentity))
		b.WriteString("\n")
	}

	// Description
	if cap.Description != "" {
		b.WriteString("\n")
		b.WriteString(DetailLabelStyle.Render("Description:"))
		b.WriteString("\n")
		desc := lipgloss.NewStyle().
			Foreground(styles.Text).
			Width(m.width - 20).
			Render(cap.Description)
		b.WriteString("  " + desc)
		b.WriteString("\n")
	}

	// Tags
	if len(cap.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString(DetailLabelStyle.Render("Tags:"))
		b.WriteString("\n  ")
		b.WriteString(RenderTags(cap.Tags))
		b.WriteString("\n")
	}

	// Input/Output schemas (if available)
	if cap.InputSchema != "" || cap.OutputSchema != "" {
		b.WriteString("\n")
		b.WriteString(DetailSectionStyle.Render("Schemas"))

		if cap.InputSchema != "" {
			b.WriteString("\n")
			b.WriteString(DetailLabelStyle.Render("Input:"))
			b.WriteString("\n")
			schema := lipgloss.NewStyle().
				Foreground(styles.Muted).
				Background(styles.BgDark).
				Padding(0, 1).
				Render(cap.InputSchema)
			b.WriteString("  " + schema)
		}

		if cap.OutputSchema != "" {
			b.WriteString("\n")
			b.WriteString(DetailLabelStyle.Render("Output:"))
			b.WriteString("\n")
			schema := lipgloss.NewStyle().
				Foreground(styles.Muted).
				Background(styles.BgDark).
				Padding(0, 1).
				Render(cap.OutputSchema)
			b.WriteString("  " + schema)
		}
	}

	// Help
	b.WriteString("\n\n")
	help := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true).
		Render("Press Esc to return to list")
	b.WriteString(help)

	return DetailContainerStyle.Width(m.width - 4).Height(m.height - 4).Render(b.String())
}

func (m Model) wrapInBox(content string) string {
	return styles.BoxStyle.Width(m.width - 4).Render(content)
}

// Commands
func (m Model) fetchCapabilities() tea.Msg {
	caps, err := m.client.DiscoverCapabilities("", "", 100)
	return capabilitiesMsg{capabilities: caps, err: err}
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Browse"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	switch m.mode {
	case ModeSearch:
		return "Enter: apply search | Esc: cancel"
	case ModeDetail:
		return "Esc: back to list | r: refresh"
	default:
		if m.searchQuery != "" {
			return "Enter: details | /: search | Esc: clear search | r: refresh"
		}
		return "Enter: details | /: search | r: refresh"
	}
}

// SetSize updates dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.searchInput.Width = width - 20
}

// Focus activates the view
func (m *Model) Focus() {
	m.focused = true
}

// Blur deactivates the view
func (m *Model) Blur() {
	m.focused = false
	// Exit search mode when blurred
	if m.mode == ModeSearch {
		m.mode = ModeList
		m.searchInput.Blur()
	}
}

// Helper functions

func isLocal(cap client.Capability) bool {
	// TODO: Compare with local agent identity
	return strings.Contains(cap.MRI, "local") || cap.AgentIdentity == ""
}

func formatCapabilityName(mri string) string {
	// Extract capability name from MRI
	// mri:capability:io.macula/agent/name -> agent/name
	parts := strings.Split(mri, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[1:], "/")
	}
	return mri
}
