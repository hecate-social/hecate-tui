package monitor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Model is the monitor view model
type Model struct {
	client   *client.Client
	width    int
	height   int
	focused  bool
	loading  bool
	spinner  spinner.Model
	health   *client.Health
	identity *client.Identity
	err      error
}

// New creates a new monitor view
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
type healthMsg struct {
	health   *client.Health
	identity *client.Identity
	err      error
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchHealth,
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
		case "r":
			m.loading = true
			return m, m.fetchHealth
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case healthMsg:
		m.loading = false
		m.health = msg.health
		m.identity = msg.identity
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
	b.WriteString(styles.TitleStyle.Render("📊 Monitor"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Fetching daemon status...")
		return m.wrapInBox(b.String())
	}

	if m.err != nil {
		b.WriteString(styles.StatusError.Render("⚠ " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("Make sure the Hecate daemon is running:\n  hecate start"))
		return m.wrapInBox(b.String())
	}

	// Daemon status section
	b.WriteString(m.renderDaemonStatus())
	b.WriteString("\n\n")

	// Identity section
	b.WriteString(m.renderIdentity())
	b.WriteString("\n\n")

	// Mesh section
	b.WriteString(m.renderMesh())

	return m.wrapInBox(b.String())
}

func (m Model) renderDaemonStatus() string {
	var rows []string

	rows = append(rows, lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Secondary).
		Render("Daemon Status"))

	rows = append(rows, "")

	if m.health == nil {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("No health data available"))
		return strings.Join(rows, "\n")
	}

	// Status with indicator
	statusStyle := styles.StatusOK
	if m.health.Status != "healthy" {
		statusStyle = styles.StatusError
	}
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Status:"),
		statusStyle.Render("● "+m.health.Status)))

	// Version
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Version:"),
		styles.ValueStyle.Render(m.health.Version)))

	// Uptime
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Uptime:"),
		styles.ValueStyle.Render(formatUptime(m.health.UptimeSeconds))))

	// Port
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Port:"),
		styles.ValueStyle.Render("4444")))

	return strings.Join(rows, "\n")
}

func (m Model) renderIdentity() string {
	var rows []string

	rows = append(rows, lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Secondary).
		Render("Identity"))

	rows = append(rows, "")

	if m.identity == nil {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("  No identity configured"))
		return strings.Join(rows, "\n")
	}

	// MRI
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("MRI:"),
		lipgloss.NewStyle().Foreground(styles.Primary).Render(m.identity.Identity)))

	// Created
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Created:"),
		styles.ValueStyle.Render(m.identity.CreatedAt)))

	return strings.Join(rows, "\n")
}

func (m Model) renderMesh() string {
	var rows []string

	rows = append(rows, lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Secondary).
		Render("Mesh Connection"))

	rows = append(rows, "")

	// TODO: Get actual mesh status from daemon
	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Bootstrap:"),
		styles.ValueStyle.Render("boot.macula.io:443")))

	rows = append(rows, fmt.Sprintf("  %s %s",
		styles.LabelStyle.Render("Status:"),
		styles.StatusOK.Render("● Connected")))

	return strings.Join(rows, "\n")
}

func (m Model) wrapInBox(content string) string {
	return styles.BoxStyle.Width(m.width - 4).Render(content)
}

// Commands
func (m Model) fetchHealth() tea.Msg {
	health, err := m.client.GetHealth()
	if err != nil {
		return healthMsg{err: err}
	}

	identity, _ := m.client.GetIdentity()
	return healthMsg{health: health, identity: identity}
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Monitor"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	return "r: refresh"
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

func formatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
