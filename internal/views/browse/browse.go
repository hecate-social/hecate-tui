package browse

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
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
	selected     int
	err          error
}

// New creates a new browse view
func New(c *client.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return Model{
		client:  c,
		spinner: s,
		loading: true,
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
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.capabilities)-1 {
				m.selected++
			}
		case "r":
			m.loading = true
			return m, m.fetchCapabilities
		case "enter":
			// TODO: View capability details
		case "/":
			// TODO: Search
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case capabilitiesMsg:
		m.loading = false
		m.capabilities = msg.capabilities
		m.err = msg.err
	}

	return m, tea.Batch(cmds...)
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("🔍 Browse Capabilities"))
	b.WriteString("\n\n")

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

	// Capability list
	b.WriteString(m.renderCapabilities())

	return m.wrapInBox(b.String())
}

func (m Model) renderEmpty() string {
	return lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true).
		Render("No capabilities discovered on the mesh.\n\nCapabilities will appear here when agents announce them.")
}

func (m Model) renderCapabilities() string {
	var rows []string

	// Header
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(30).Foreground(styles.Muted).Render("CAPABILITY"),
		lipgloss.NewStyle().Width(12).Foreground(styles.Muted).Render("SOURCE"),
		lipgloss.NewStyle().Foreground(styles.Muted).Render("TAGS"),
	)
	rows = append(rows, header)
	rows = append(rows, lipgloss.NewStyle().Foreground(styles.Border).Render(strings.Repeat("─", m.width-8)))

	for i, cap := range m.capabilities {
		// Determine source
		source := "remote"
		sourceStyle := lipgloss.NewStyle().Foreground(styles.Muted)
		if isLocal(cap) {
			source = "local"
			sourceStyle = lipgloss.NewStyle().Foreground(styles.Success)
		}

		// Format name from MRI
		name := formatCapabilityName(cap.MRI)
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		// Format tags
		tags := strings.Join(cap.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Width(30).Render(name),
			sourceStyle.Width(12).Render(source),
			lipgloss.NewStyle().Foreground(styles.Text).Render(tags),
		)

		// Highlight selected
		if i == m.selected && m.focused {
			row = lipgloss.NewStyle().
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Width(m.width - 8).
				Render(row)
		} else if i == m.selected {
			row = lipgloss.NewStyle().
				Background(styles.BgMedium).
				Width(m.width - 8).
				Render(row)
		}

		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
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
	return "↑↓: select • Enter: view • /: search • r: refresh"
}

// SetSize updates dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Focus activates the view
func (m *Model) Focus() {
	m.focused = true
}

// Blur deactivates the view
func (m *Model) Blur() {
	m.focused = false
}

// Helper functions

func isLocal(cap client.Capability) bool {
	// TODO: Compare with local agent identity
	return strings.Contains(cap.MRI, "local") || cap.AgentIdentity == ""
}

func formatCapabilityName(mri string) string {
	// Extract capability name from MRI
	// mri:capability:io.macula/agent/name → agent/name
	parts := strings.Split(mri, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[1:], "/")
	}
	return mri
}
