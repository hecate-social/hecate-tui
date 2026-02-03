package pair

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Model is the pair view model
type Model struct {
	client   *client.Client
	width    int
	height   int
	focused  bool
	identity *client.Identity
	paired   bool
}

// New creates a new pair view
func New(c *client.Client) Model {
	return Model{
		client: c,
	}
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return m.fetchIdentity
}

// Messages
type identityMsg struct {
	identity *client.Identity
	err      error
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "p":
			// TODO: Start pairing flow
		case "r":
			return m, m.fetchIdentity
		}

	case identityMsg:
		m.identity = msg.identity
		// Check if paired (identity exists and has realm)
		if m.identity != nil && m.identity.Identity != "" {
			m.paired = strings.Contains(m.identity.Identity, "io.macula")
		}
	}

	return m, nil
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("🔗 Pair"))
	b.WriteString("\n\n")

	if m.paired {
		b.WriteString(m.renderPaired())
	} else {
		b.WriteString(m.renderUnpaired())
	}

	return styles.BoxStyle.Width(m.width - 4).Render(b.String())
}

func (m Model) renderPaired() string {
	var b strings.Builder

	// Paired status
	status := lipgloss.NewStyle().
		Foreground(styles.Success).
		Bold(true).
		Render("✓ Paired")

	b.WriteString(status)
	b.WriteString("\n\n")

	// Identity info
	if m.identity != nil {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("Identity:"),
			lipgloss.NewStyle().Foreground(styles.Primary).Render(m.identity.Identity),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Actions
	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.Muted).
		Render("Press [p] to re-pair with a different realm"))

	return b.String()
}

func (m Model) renderUnpaired() string {
	var b strings.Builder

	// Unpaired status
	status := lipgloss.NewStyle().
		Foreground(styles.Warning).
		Bold(true).
		Render("○ Not Paired")

	b.WriteString(status)
	b.WriteString("\n\n")

	// Instructions
	instructions := `To pair this agent with a realm:

  1. Go to https://macula.io/pair
  2. Sign in to your account
  3. Click "Pair Device"
  4. Press [p] here to start pairing
  5. Enter the confirmation code shown

Pairing connects this agent to the Macula mesh
and enables capability discovery.`

	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.Text).
		Render(instructions))

	b.WriteString("\n\n")

	// CTA
	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Render("Press [p] to start pairing"))

	return b.String()
}

// Commands
func (m Model) fetchIdentity() tea.Msg {
	identity, err := m.client.GetIdentity()
	return identityMsg{identity: identity, err: err}
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Pair"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	if m.paired {
		return "p: re-pair • r: refresh"
	}
	return "p: start pairing"
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
