package projects

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/projects"
	"github.com/hecate-social/hecate-tui/internal/views/projects/and"
	"github.com/hecate-social/hecate-tui/internal/views/projects/anp"
	pint "github.com/hecate-social/hecate-tui/internal/views/projects/int"
	"github.com/hecate-social/hecate-tui/internal/views/projects/doo"
)

// ViewMode determines what's shown
type ViewMode int

const (
	ModeList   ViewMode = iota // Project list
	ModePhases                 // Phase view for selected project
)

// Model is the projects view model
type Model struct {
	width    int
	height   int
	focused  bool
	mode     ViewMode

	// Project list
	projects    []*projects.Project
	selected    int
	current     *projects.Project
	detector    *projects.Detector

	// Phase navigation
	phaseBar    *PhaseBar

	// Phase views
	andView     and.Model
	anpView     anp.Model
	intView     pint.Model
	dooView     doo.Model
}

// New creates a new projects view
func New() Model {
	detector := projects.NewDetector()

	// Try to detect current project
	current, _ := detector.DetectCurrent()

	m := Model{
		detector: detector,
		current:  current,
		phaseBar: NewPhaseBar(),
		andView:  and.New(),
		anpView:  anp.New(),
		intView:  pint.New(),
		dooView:  doo.New(),
	}

	// If we have a current project, start in phases mode
	if current != nil {
		m.mode = ModePhases
		m.phaseBar.SetActive(current.CurrentPhase)
	}

	return m
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return m.scanProjects
}

// Project scanning
type projectsScannedMsg struct {
	projects []*projects.Project
}

func (m Model) scanProjects() tea.Msg {
	// Scan current directory and parent for projects
	var found []*projects.Project

	// Check current directory
	if p, _ := m.detector.DetectCurrent(); p != nil {
		found = append(found, p)
	}

	return projectsScannedMsg{projects: found}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectsScannedMsg:
		m.projects = msg.projects
		if len(m.projects) > 0 && m.current == nil {
			m.current = m.projects[0]
			m.mode = ModePhases
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case ModeList:
			return m.updateList(msg)
		case ModePhases:
			return m.updatePhases(msg)
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.projects)-1 {
			m.selected++
		}
	case "enter":
		if len(m.projects) > 0 {
			m.current = m.projects[m.selected]
			m.mode = ModePhases
			m.phaseBar.SetActive(m.current.CurrentPhase)
		}
	case "r":
		return m, m.scanProjects
	}
	return m, nil
}

func (m Model) updatePhases(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.phaseBar.Prev()
	case "right", "l":
		m.phaseBar.Next()
	case "esc":
		if len(m.projects) > 1 {
			m.mode = ModeList
		}
	case "i":
		// Initialize workspace
		if m.current != nil && !m.current.HasWorkspace {
			ws, err := projects.OpenWorkspace(m.current)
			if err == nil {
				ws.Init()
			}
		}
	}
	return m, nil
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case ModeList:
		return m.renderList()
	case ModePhases:
		return m.renderPhases()
	}
	return ""
}

func (m Model) renderList() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("📁 Projects"))
	b.WriteString("\n\n")

	if len(m.projects) == 0 {
		b.WriteString(EmptyStyle.Render("No projects detected.\n\n"))
		b.WriteString(EmptyStyle.Render("Navigate to a git repository or create a HECATE.md file."))
		return b.String()
	}

	for i, p := range m.projects {
		var line strings.Builder

		if i == m.selected {
			line.WriteString("› ")
		} else {
			line.WriteString("  ")
		}

		line.WriteString(p.TypeIcon())
		line.WriteString(" ")
		line.WriteString(ProjectNameStyle.Render(p.Name))
		line.WriteString(" ")

		if p.HasWorkspace {
			line.WriteString(WorkspaceActiveStyle.Render("●"))
		} else {
			line.WriteString(WorkspaceMissingStyle.Render("○"))
		}

		if i == m.selected {
			b.WriteString(ProjectItemSelectedStyle.Width(m.width - 6).Render(line.String()))
		} else {
			b.WriteString(ProjectItemStyle.Render(line.String()))
		}
		b.WriteString("\n")

		// Show path for selected
		if i == m.selected {
			b.WriteString("    ")
			b.WriteString(ProjectPathStyle.Render(p.Path))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑↓ navigate • Enter select • r refresh"))

	return b.String()
}

func (m Model) renderPhases() string {
	var b strings.Builder

	// Project header
	if m.current != nil {
		b.WriteString(TitleStyle.Render(m.current.TypeIcon() + " " + m.current.Name))

		if m.current.HasWorkspace {
			b.WriteString("  ")
			b.WriteString(WorkspaceActiveStyle.Render("● workspace active"))
		} else {
			b.WriteString("  ")
			b.WriteString(WorkspaceMissingStyle.Render("○ no workspace (press i to init)"))
		}

		if m.current.GitBranch != "" {
			b.WriteString("  ")
			b.WriteString(ProjectTypeStyle.Render("⎇ " + m.current.GitBranch))
		}
	} else {
		b.WriteString(TitleStyle.Render("📁 Projects"))
	}
	b.WriteString("\n")

	// Phase tab bar
	m.phaseBar.SetWidth(m.width - 4)
	b.WriteString(m.phaseBar.View())
	b.WriteString("\n")

	// Phase content
	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	var phaseContent string
	switch m.phaseBar.Active() {
	case projects.PhaseAnD:
		phaseContent = m.andView.View()
	case projects.PhaseAnP:
		phaseContent = m.anpView.View()
	case projects.PhaseInT:
		phaseContent = m.intView.View()
	case projects.PhaseDoO:
		phaseContent = m.dooView.View()
	}

	b.WriteString(PhaseContentStyle.Width(contentWidth).Render(phaseContent))
	b.WriteString("\n\n")

	// Help
	help := "←→ switch phase"
	if m.current != nil && !m.current.HasWorkspace {
		help += " • i init workspace"
	}
	if len(m.projects) > 1 {
		help += " • Esc project list"
	}
	b.WriteString(HelpStyle.Render(help))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Projects"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	if m.mode == ModePhases {
		info := m.phaseBar.ActiveInfo()
		return info.Icon + " " + info.Name
	}
	return "Select a project"
}

// SetSize updates dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.andView.SetSize(width-12, height-10)
	m.anpView.SetSize(width-12, height-10)
	m.intView.SetSize(width-12, height-10)
	m.dooView.SetSize(width-12, height-10)
}

// Focus activates the view
func (m *Model) Focus() {
	m.focused = true
}

// Blur deactivates the view
func (m *Model) Blur() {
	m.focused = false
}
