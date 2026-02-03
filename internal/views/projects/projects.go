package projects

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Model is the projects view model
type Model struct {
	width   int
	height  int
	focused bool
}

// New creates a new projects view
func New() Model {
	return Model{}
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("📁 Projects"))
	b.WriteString("\n\n")

	// Coming soon content
	content := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Render(m.renderPlaceholder())

	b.WriteString(content)

	return styles.BoxStyle.Width(m.width - 4).Render(b.String())
}

func (m Model) renderPlaceholder() string {
	phases := []string{
		"📊 Analysis & Discovery (AnD)",
		"🏗️  Architecture & Planning (AnP)",
		"⚡ Implementation & Testing (InT)",
		"🚀 Deployment & Operations (DoO)",
	}

	var b strings.Builder
	b.WriteString("Project workspace coming soon.\n\n")
	b.WriteString("Development phases:\n\n")

	for _, phase := range phases {
		b.WriteString("  " + phase + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Italic(true).
		Foreground(styles.Primary).
		Render("The TUI will detect HECATE.md files and help you\nnavigate through the development lifecycle."))

	return b.String()
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Projects"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	return "Coming soon..."
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
