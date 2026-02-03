package anp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true)

	comingSoonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	featureStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
)

// Model is the Architecture & Planning phase view
type Model struct {
	width   int
	height  int
	focused bool
}

// New creates an AnP phase view
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
	var b strings.Builder

	b.WriteString(headerStyle.Render("🏗️ Architecture & Planning"))
	b.WriteString("\n")
	b.WriteString(descStyle.Render("Design the system and plan implementation"))
	b.WriteString("\n\n")

	b.WriteString(comingSoonStyle.Render("Coming Soon"))
	b.WriteString("\n\n")

	features := []string{
		"• System architecture diagrams",
		"• Component design documents",
		"• API specifications",
		"• Data model planning",
		"• Task breakdown and estimation",
	}

	b.WriteString(featureStyle.Render("Planned features:"))
	b.WriteString("\n")
	for _, f := range features {
		b.WriteString(featureStyle.Render(f))
		b.WriteString("\n")
	}

	return b.String()
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
