package monitor

import (
	"fmt"
	"strings"
	"time"

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

	// Stats
	subscriptionCount int
	capabilityCount   int

	// Refresh
	lastRefresh time.Time
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
	health            *client.Health
	identity          *client.Identity
	subscriptionCount int
	capabilityCount   int
	err               error
}

type refreshTickMsg struct{}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchAll,
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
			return m, m.fetchAll
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case healthMsg:
		m.loading = false
		m.health = msg.health
		m.identity = msg.identity
		m.subscriptionCount = msg.subscriptionCount
		m.capabilityCount = msg.capabilityCount
		m.err = msg.err
		m.lastRefresh = time.Now()

	case refreshTickMsg:
		if m.focused && time.Since(m.lastRefresh) > 30*time.Second {
			return m, m.fetchAll
		}
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
	b.WriteString(styles.TitleStyle.Render("Monitor"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Fetching daemon status...")
		return m.wrapInBox(b.String())
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		return m.wrapInBox(b.String())
	}

	// Stats cards row
	b.WriteString(m.renderStatsRow())
	b.WriteString("\n\n")

	// Two-column layout
	leftCol := m.renderDaemonStatus()
	rightCol := m.renderMesh()

	colWidth := (m.width - 12) / 2
	leftBox := SectionBoxStyle.Width(colWidth).Render(leftCol)
	rightBox := SectionBoxStyle.Width(colWidth).Render(rightCol)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox))
	b.WriteString("\n\n")

	// Identity section (full width)
	b.WriteString(SectionBoxStyle.Width(m.width - 8).Render(m.renderIdentity()))

	// Last refresh time
	b.WriteString("\n\n")
	if !m.lastRefresh.IsZero() {
		refreshTime := m.lastRefresh.Format("15:04:05")
		b.WriteString(lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true).
			Render("Last updated: " + refreshTime))
	}

	return m.wrapInBox(b.String())
}

func (m Model) renderError() string {
	var b strings.Builder

	icon := lipgloss.NewStyle().
		Foreground(UnhealthyColor).
		Render("⚠")

	b.WriteString(icon + " ")
	b.WriteString(StatusUnhealthyStyle.Render("Daemon Offline"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.Text).
		Render(m.err.Error()))
	b.WriteString("\n\n")

	hint := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Render("Make sure the Hecate daemon is running:\n\n  hecate start\n\nOr start the daemon shell:\n\n  cd hecate-daemon && rebar3 shell")

	b.WriteString(hint)

	return b.String()
}

func (m Model) renderStatsRow() string {
	// Uptime stat
	uptimeVal := "--"
	if m.health != nil {
		uptimeVal = formatUptimeShort(m.health.UptimeSeconds)
	}
	uptimeCard := RenderStatCard(uptimeVal, "Uptime")

	// Subscriptions stat
	subVal := fmt.Sprintf("%d", m.subscriptionCount)
	subCard := RenderStatCard(subVal, "Subscriptions")

	// Capabilities stat
	capVal := fmt.Sprintf("%d", m.capabilityCount)
	capCard := RenderStatCard(capVal, "Capabilities")

	// Status stat
	statusVal := "Offline"
	if m.health != nil && m.health.Status == "healthy" {
		statusVal = "Online"
	}
	statusCard := RenderStatCard(statusVal, "Status")

	return lipgloss.JoinHorizontal(lipgloss.Center,
		uptimeCard, "  ",
		subCard, "  ",
		capCard, "  ",
		statusCard,
	)
}

func (m Model) renderDaemonStatus() string {
	var rows []string

	rows = append(rows, SectionTitleStyle.Render("Daemon"))

	if m.health == nil {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("No health data"))
		return strings.Join(rows, "\n")
	}

	// Status
	rows = append(rows, renderRow("Status:", StatusIndicator(m.health.Status)+" "+StatusText(m.health.Status)))

	// Version
	rows = append(rows, renderRow("Version:", m.health.Version))

	// Uptime
	rows = append(rows, renderRow("Uptime:", formatUptime(m.health.UptimeSeconds)))

	// Port
	rows = append(rows, renderRow("Port:", "4444"))

	return strings.Join(rows, "\n")
}

func (m Model) renderIdentity() string {
	var rows []string

	rows = append(rows, SectionTitleStyle.Render("Identity"))

	if m.identity == nil {
		rows = append(rows, lipgloss.NewStyle().
			Foreground(styles.Muted).
			Render("No identity configured. Run: hecate init"))
		return strings.Join(rows, "\n")
	}

	// MRI
	mri := m.identity.Identity
	if len(mri) > 60 {
		mri = mri[:57] + "..."
	}
	rows = append(rows, renderRow("MRI:", RowHighlightStyle.Render(mri)))

	// Public key (truncated)
	pubKey := m.identity.PublicKey
	if len(pubKey) > 40 {
		pubKey = pubKey[:20] + "..." + pubKey[len(pubKey)-16:]
	}
	rows = append(rows, renderRow("Public Key:", pubKey))

	// Created
	if m.identity.CreatedAt != "" {
		rows = append(rows, renderRow("Created:", m.identity.CreatedAt))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderMesh() string {
	var rows []string

	rows = append(rows, SectionTitleStyle.Render("Mesh Connection"))

	// Bootstrap server
	rows = append(rows, renderRow("Bootstrap:", "boot.macula.io:443"))

	// Connection status (derived from health)
	meshStatus := "disconnected"
	if m.health != nil && m.health.Status == "healthy" {
		meshStatus = "connected"
	}
	rows = append(rows, renderRow("Status:", StatusIndicator(meshStatus)+" "+StatusText(meshStatus)))

	// Note about mesh status
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true).
		Render("Mesh status endpoint TBD"))

	return strings.Join(rows, "\n")
}

func renderRow(label, value string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		RowLabelStyle.Render(label),
		RowValueStyle.Render(value),
	)
}

func (m Model) wrapInBox(content string) string {
	return styles.BoxStyle.Width(m.width - 4).Render(content)
}

// Commands
func (m Model) fetchAll() tea.Msg {
	health, err := m.client.GetHealth()
	if err != nil {
		return healthMsg{err: err}
	}

	identity, _ := m.client.GetIdentity()

	// Fetch subscription count
	subs, _ := m.client.ListSubscriptions()
	subCount := 0
	if subs != nil {
		subCount = len(subs)
	}

	// Fetch capability count
	caps, _ := m.client.DiscoverCapabilities("", "", 1000)
	capCount := 0
	if caps != nil {
		capCount = len(caps)
	}

	return healthMsg{
		health:            health,
		identity:          identity,
		subscriptionCount: subCount,
		capabilityCount:   capCount,
	}
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

func formatUptimeShort(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
