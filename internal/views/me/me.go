package me

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Model is the me view model
type Model struct {
	client   *client.Client
	width    int
	height   int
	focused  bool
	identity *client.Identity
	err      error
}

// New creates a new me view
func New(c *client.Client) Model {
	return Model{
		client: c,
	}
}

// Messages
type identityMsg struct {
	identity *client.Identity
	err      error
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return m.fetchIdentity
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "r":
			return m, m.fetchIdentity
		case "s":
			// TODO: Open settings
		}

	case identityMsg:
		m.identity = msg.identity
		m.err = msg.err
	}

	return m, nil
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title with avatar
	b.WriteString(styles.TitleStyle.Render("👤 Me"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(styles.StatusError.Render("⚠ " + m.err.Error()))
		return styles.BoxStyle.Width(m.width - 4).Render(b.String())
	}

	// Identity section
	b.WriteString(m.renderIdentity())
	b.WriteString("\n\n")

	// Stats section
	b.WriteString(m.renderStats())
	b.WriteString("\n\n")

	// Settings hint
	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true).
		Render("Press [s] for settings"))

	return styles.BoxStyle.Width(m.width - 4).Render(b.String())
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
			Render("  No identity configured. Run: hecate init"))
		return strings.Join(rows, "\n")
	}

	// MRI with nice formatting
	mri := m.identity.Identity
	rows = append(rows, "  "+styles.LabelStyle.Render("MRI:"))
	rows = append(rows, "  "+lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Render(mri))

	rows = append(rows, "")

	// Parse realm from MRI
	realm := parseRealm(mri)
	if realm != "" {
		rows = append(rows, "  "+lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("Realm:"),
			styles.ValueStyle.Render(realm)))
	}

	// Created date
	if m.identity.CreatedAt != "" {
		rows = append(rows, "  "+lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("Created:"),
			styles.ValueStyle.Render(m.identity.CreatedAt)))
	}

	// Pairing status
	paired := realm != ""
	pairStatus := lipgloss.NewStyle().Foreground(styles.Warning).Render("○ Not paired")
	if paired {
		pairStatus = lipgloss.NewStyle().Foreground(styles.Success).Render("✓ Paired")
	}
	rows = append(rows, "  "+lipgloss.JoinHorizontal(lipgloss.Left,
		styles.LabelStyle.Render("Status:"),
		pairStatus))

	return strings.Join(rows, "\n")
}

func (m Model) renderStats() string {
	var rows []string

	rows = append(rows, lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Secondary).
		Render("Statistics"))

	rows = append(rows, "")

	// Placeholder stats
	stats := []struct {
		label string
		value string
	}{
		{"Capabilities:", "0 announced"},
		{"RPC Calls:", "0 served"},
		{"Uptime:", "—"},
	}

	for _, stat := range stats {
		rows = append(rows, "  "+lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render(stat.label),
			styles.ValueStyle.Render(stat.value)))
	}

	return strings.Join(rows, "\n")
}

// Commands
func (m Model) fetchIdentity() tea.Msg {
	identity, err := m.client.GetIdentity()
	return identityMsg{identity: identity, err: err}
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Me"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	return "s: settings • r: refresh"
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

func parseRealm(mri string) string {
	// mri:agent:io.macula/name → io.macula
	if !strings.HasPrefix(mri, "mri:") {
		return ""
	}
	parts := strings.Split(mri, ":")
	if len(parts) < 3 {
		return ""
	}
	pathParts := strings.Split(parts[2], "/")
	if len(pathParts) > 0 {
		return pathParts[0]
	}
	return ""
}
