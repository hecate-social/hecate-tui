package alc

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

// ViewMode represents the internal overlay state.
type ViewMode int

const (
	ModeList   ViewMode = iota // Project list
	ModeSearch                 // Search input focused
	ModeDetail                 // Single project detail
	ModePhase                  // Phase artifact detail
)

// PhaseTab identifies which phase section is shown in detail view.
type PhaseTab int

const (
	TabDiscovery    PhaseTab = iota
	TabArchitecture
	TabTesting
	TabDeployment
)

// Model is the Projects mode overlay.
type Model struct {
	client  *client.Client
	theme   *theme.Theme
	styles  *theme.Styles
	width   int
	height  int
	loading bool
	spinner spinner.Model

	// Project list
	projects []client.ALCProject
	filtered []client.ALCProject
	selected int
	err      error

	// Search
	mode        ViewMode
	searchInput textinput.Model
	searchQuery string

	// Detail
	detailProject *client.ALCProject
	phaseTab      PhaseTab

	// Phase artifacts (loaded on demand)
	findings        []client.ALCFinding
	terms           []client.ALCTerm
	dossiers        []client.ALCDossier
	spokes          []client.ALCSpoke
	implementations []client.ALCImplementation
	builds          []client.ALCBuild
	deployments     []client.ALCDeployment
	incidents       []client.ALCIncident
	artifactsLoaded bool
	phaseSelected   int
}

// projectsMsg carries fetched projects.
type projectsMsg struct {
	projects []client.ALCProject
	err      error
}

// artifactsMsg carries fetched phase artifacts.
type artifactsMsg struct {
	findings        []client.ALCFinding
	terms           []client.ALCTerm
	dossiers        []client.ALCDossier
	spokes          []client.ALCSpoke
	implementations []client.ALCImplementation
	builds          []client.ALCBuild
	deployments     []client.ALCDeployment
	incidents       []client.ALCIncident
	err             error
}

// New creates a Projects mode model.
func New(c *client.Client, t *theme.Theme, s *theme.Styles) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(t.Primary)

	ti := textinput.New()
	ti.Placeholder = "Filter projects..."
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

// Init starts loading projects.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchProjects,
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

	case projectsMsg:
		m.loading = false
		m.projects = msg.projects
		m.filtered = msg.projects
		m.err = msg.err

	case artifactsMsg:
		m.findings = msg.findings
		m.terms = msg.terms
		m.dossiers = msg.dossiers
		m.spokes = msg.spokes
		m.implementations = msg.implementations
		m.builds = msg.builds
		m.deployments = msg.deployments
		m.incidents = msg.incidents
		m.artifactsLoaded = true
		if msg.err != nil {
			m.err = msg.err
		}
	}

	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// HandleKey processes a keypress in Projects mode. Returns true if the key was consumed.
func (m *Model) HandleKey(key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch m.mode {
	case ModeSearch:
		return m.handleSearchKey(key, msg)
	case ModeDetail:
		return m.handleDetailKey(key)
	case ModePhase:
		return m.handlePhaseKey(key)
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
			p := m.filtered[m.selected]
			m.detailProject = &p
			m.mode = ModeDetail
			m.phaseTab = phaseTabFromPhase(p.CurrentPhase)
			m.artifactsLoaded = false
			m.phaseSelected = 0
			return true, m.fetchArtifacts(p.ProjectID)
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
		return true, m.fetchProjects
	case "esc":
		// Esc exits Projects mode — handled by app
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
	case "esc":
		m.mode = ModeList
		m.detailProject = nil
		m.artifactsLoaded = false
		return true, nil
	case "tab":
		m.phaseTab = (m.phaseTab + 1) % 4
		m.phaseSelected = 0
		return true, nil
	case "shift+tab":
		m.phaseTab = (m.phaseTab + 3) % 4
		m.phaseSelected = 0
		return true, nil
	case "enter":
		m.mode = ModePhase
		m.phaseSelected = 0
		return true, nil
	case "r":
		if m.detailProject != nil {
			m.artifactsLoaded = false
			return true, m.fetchArtifacts(m.detailProject.ProjectID)
		}
		return true, nil
	}
	return true, nil
}

func (m *Model) handlePhaseKey(key string) (bool, tea.Cmd) {
	switch key {
	case "esc":
		m.mode = ModeDetail
		return true, nil
	case "j", "down":
		m.phaseSelected++
		maxItems := m.phaseItemCount()
		if m.phaseSelected >= maxItems {
			m.phaseSelected = maxItems - 1
		}
		if m.phaseSelected < 0 {
			m.phaseSelected = 0
		}
		return true, nil
	case "k", "up":
		if m.phaseSelected > 0 {
			m.phaseSelected--
		}
		return true, nil
	case "tab":
		m.phaseTab = (m.phaseTab + 1) % 4
		m.phaseSelected = 0
		return true, nil
	}
	return true, nil
}

func (m *Model) phaseItemCount() int {
	switch m.phaseTab {
	case TabDiscovery:
		return len(m.findings) + len(m.terms)
	case TabArchitecture:
		return len(m.dossiers) + len(m.spokes)
	case TabTesting:
		return len(m.implementations) + len(m.builds)
	case TabDeployment:
		return len(m.deployments) + len(m.incidents)
	}
	return 0
}

func (m *Model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.projects
		return
	}
	query := strings.ToLower(m.searchQuery)
	var filtered []client.ALCProject
	for _, p := range m.projects {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) ||
			strings.Contains(strings.ToLower(p.ProjectID), query) ||
			strings.Contains(strings.ToLower(p.CurrentPhase), query) {
			filtered = append(filtered, p)
		}
	}
	m.filtered = filtered
}

// SetSize updates the overlay dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.searchInput.Width = width - 10
}

// View renders the projects overlay panel.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case ModeDetail:
		return m.renderDetail()
	case ModePhase:
		return m.renderPhase()
	default:
		return m.renderList()
	}
}

func (m Model) renderList() string {
	s := m.styles
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Projects"))
	b.WriteString("\n\n")

	// Search bar
	if m.mode == ModeSearch || m.searchQuery != "" {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Loading projects...")
		return m.wrapPanel(b.String())
	}

	if m.err != nil {
		b.WriteString(s.Error.Render("Error: " + m.err.Error()))
		return m.wrapPanel(b.String())
	}

	if len(m.projects) == 0 {
		b.WriteString(s.Subtle.Render("No projects found."))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Use /alc init <name> to create one."))
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
		s.Subtle.Width(28).Render("PROJECT"),
		s.Subtle.Width(16).Render("PHASE"),
		s.Subtle.Render("PROGRESS"),
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
		p := m.filtered[i]

		indicator := "  "
		if i == m.selected {
			indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("▸ ")
		}

		name := p.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}

		phase := m.renderPhaseBadge(p.CurrentPhase)
		progress := m.renderProgress(p)

		row := indicator +
			lipgloss.NewStyle().Width(24).Foreground(m.theme.Text).Render(name) +
			lipgloss.NewStyle().Width(16).Render(phase) +
			s.Subtle.Render(progress)

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

	count := fmt.Sprintf(" (%d of %d)", len(m.filtered), len(m.projects))
	return s.Subtle.Render("  /" + m.searchQuery + count)
}

func (m Model) renderPhaseBadge(phase string) string {
	var color lipgloss.Color
	var label string

	switch strings.ToLower(phase) {
	case "discovery", "dna":
		color = m.theme.Secondary
		label = "DnA"
	case "architecture", "anp":
		color = m.theme.Warning
		label = "AnP"
	case "testing", "tni":
		color = m.theme.Success
		label = "TnI"
	case "deployment", "dno":
		color = m.theme.Accent
		label = "DnO"
	case "initiated":
		color = m.theme.TextMuted
		label = "NEW"
	case "completed":
		color = m.theme.Success
		label = "DONE"
	default:
		color = m.theme.TextMuted
		label = phase
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(label)
}

func (m Model) renderProgress(p client.ALCProject) string {
	switch strings.ToLower(p.CurrentPhase) {
	case "discovery", "dna":
		return fmt.Sprintf("%dF %dT", p.FindingCount, p.TermCount)
	case "architecture", "anp":
		return fmt.Sprintf("%dD %dS", p.DossierCount, p.SpokeCount)
	case "testing", "tni":
		return fmt.Sprintf("%d/%dS", p.ImplementedSpokeCount, p.SpokeCount)
	case "deployment", "dno":
		return fmt.Sprintf("%dDep %dInc", p.DeploymentCount, p.ActiveIncidents)
	case "initiated":
		return "—"
	case "completed":
		return "done"
	default:
		return ""
	}
}

func (m Model) renderDetail() string {
	if m.detailProject == nil {
		return ""
	}

	s := m.styles
	p := m.detailProject
	var b strings.Builder

	b.WriteString(s.CardTitle.Render("Project: " + p.Name))
	b.WriteString("\n\n")

	b.WriteString(s.CardLabel.Render("  ID:"))
	b.WriteString(s.CardValue.Render(p.ProjectID))
	b.WriteString("\n")

	if p.Description != "" {
		b.WriteString(s.CardLabel.Render("  Description:"))
		b.WriteString(s.CardValue.Render(p.Description))
		b.WriteString("\n")
	}

	b.WriteString(s.CardLabel.Render("  Phase:"))
	b.WriteString(m.renderPhaseBadge(p.CurrentPhase))
	b.WriteString(" ")
	b.WriteString(s.Subtle.Render(formatPhase(p.CurrentPhase)))
	b.WriteString("\n\n")

	// Phase tabs
	b.WriteString(m.renderPhaseTabs())
	b.WriteString("\n\n")

	// Phase summary (counters for the active tab)
	b.WriteString(m.renderPhaseTabContent())

	b.WriteString("\n\n")
	b.WriteString(s.Subtle.Render("  Tab:cycle phases  Enter:drill in  Esc:back"))

	return m.wrapPanel(b.String())
}

func (m Model) renderPhaseTabs() string {
	tabs := []struct {
		tab   PhaseTab
		label string
	}{
		{TabDiscovery, "Discovery"},
		{TabArchitecture, "Architecture"},
		{TabTesting, "Testing"},
		{TabDeployment, "Deployment"},
	}

	var parts []string
	for _, t := range tabs {
		style := lipgloss.NewStyle().Padding(0, 1)
		if t.tab == m.phaseTab {
			style = style.
				Background(m.theme.Primary).
				Foreground(m.theme.StatusBarFg).
				Bold(true)
		} else {
			style = style.Foreground(m.theme.TextMuted)
		}
		parts = append(parts, style.Render(t.label))
	}

	return "  " + strings.Join(parts, " ")
}

func (m Model) renderPhaseTabContent() string {
	s := m.styles
	var b strings.Builder

	if !m.artifactsLoaded {
		b.WriteString("  " + m.spinner.View() + " Loading...")
		return b.String()
	}

	switch m.phaseTab {
	case TabDiscovery:
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Findings (%d)", len(m.findings))))
		b.WriteString("\n")
		for _, f := range m.findings {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(f.Title))
			if f.Priority != "" {
				b.WriteString(" ")
				b.WriteString(s.Subtle.Render("[" + f.Priority + "]"))
			}
			b.WriteString("\n")
		}
		if len(m.findings) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Terms (%d)", len(m.terms))))
		b.WriteString("\n")
		for _, t := range m.terms {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(t.Term))
			b.WriteString(s.Subtle.Render(" — " + truncate(t.Definition, 40)))
			b.WriteString("\n")
		}
		if len(m.terms) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}

	case TabArchitecture:
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Dossiers (%d)", len(m.dossiers))))
		b.WriteString("\n")
		for _, d := range m.dossiers {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(d.DossierName))
			if d.Description != "" {
				b.WriteString(s.Subtle.Render(" — " + truncate(d.Description, 40)))
			}
			b.WriteString("\n")
		}
		if len(m.dossiers) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Spokes (%d)", len(m.spokes))))
		b.WriteString("\n")
		for _, sp := range m.spokes {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(sp.SpokeName))
			b.WriteString(" ")
			b.WriteString(s.Subtle.Render("[" + sp.SpokeType + "]"))
			if sp.Priority != "" {
				b.WriteString(" ")
				b.WriteString(s.Subtle.Render(sp.Priority))
			}
			b.WriteString("\n")
		}
		if len(m.spokes) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}

	case TabTesting:
		p := m.detailProject
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Implementations (%d/%d spokes)", len(m.implementations), p.SpokeCount)))
		b.WriteString("\n")
		for _, impl := range m.implementations {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(impl.SpokeID))
			if impl.ImplementationNotes != "" {
				b.WriteString(s.Subtle.Render(" — " + truncate(impl.ImplementationNotes, 40)))
			}
			b.WriteString("\n")
		}
		if len(m.implementations) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Builds (%d)", len(m.builds))))
		b.WriteString("\n")
		for _, build := range m.builds {
			b.WriteString("    ")
			if build.Result == "pass" {
				b.WriteString(s.StatusOK.Render("PASS"))
			} else {
				b.WriteString(s.StatusError.Render("FAIL"))
			}
			if build.Notes != "" {
				b.WriteString(s.Subtle.Render(" — " + truncate(build.Notes, 40)))
			}
			b.WriteString("\n")
		}
		if len(m.builds) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Skeleton: "))
		if p.SkeletonCreated {
			b.WriteString(s.StatusOK.Render("created"))
		} else {
			b.WriteString(s.Subtle.Render("pending"))
		}
		b.WriteString("  ")
		b.WriteString(s.CardLabel.Render("Build: "))
		if p.BuildVerified {
			b.WriteString(s.StatusOK.Render("verified"))
		} else {
			b.WriteString(s.Subtle.Render("pending"))
		}

	case TabDeployment:
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Deployments (%d)", len(m.deployments))))
		b.WriteString("\n")
		for _, dep := range m.deployments {
			b.WriteString("    ")
			b.WriteString(s.CardValue.Render(dep.Environment))
			b.WriteString(" ")
			b.WriteString(s.Subtle.Render("v" + dep.Version))
			if dep.Notes != "" {
				b.WriteString(s.Subtle.Render(" — " + truncate(dep.Notes, 30)))
			}
			b.WriteString("\n")
		}
		if len(m.deployments) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(s.Bold.Render(fmt.Sprintf("  Incidents (%d)", len(m.incidents))))
		b.WriteString("\n")
		for _, inc := range m.incidents {
			b.WriteString("    ")
			sevStyle := s.Subtle
			if inc.Severity == "critical" || inc.Severity == "high" {
				sevStyle = s.StatusError
			} else if inc.Severity == "medium" {
				sevStyle = s.StatusWarning
			}
			b.WriteString(sevStyle.Render("[" + inc.Severity + "]"))
			b.WriteString(" ")
			b.WriteString(s.CardValue.Render(truncate(inc.Description, 40)))
			if inc.Resolution != "" {
				b.WriteString(s.StatusOK.Render(" (resolved)"))
			}
			b.WriteString("\n")
		}
		if len(m.incidents) == 0 {
			b.WriteString(s.Subtle.Render("    (none)"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderPhase() string {
	// Phase drilldown — shows the same phase tab content but with j/k item navigation
	s := m.styles
	var b strings.Builder

	tabLabel := ""
	switch m.phaseTab {
	case TabDiscovery:
		tabLabel = "Discovery"
	case TabArchitecture:
		tabLabel = "Architecture"
	case TabTesting:
		tabLabel = "Testing"
	case TabDeployment:
		tabLabel = "Deployment"
	}

	if m.detailProject != nil {
		b.WriteString(s.CardTitle.Render(m.detailProject.Name + " / " + tabLabel))
	} else {
		b.WriteString(s.CardTitle.Render(tabLabel))
	}
	b.WriteString("\n\n")

	if !m.artifactsLoaded {
		b.WriteString("  " + m.spinner.View() + " Loading...")
		return m.wrapPanel(b.String())
	}

	items := m.phaseItems()
	visibleRows := m.height - 8
	if visibleRows < 3 {
		visibleRows = 3
	}

	startIdx := 0
	if m.phaseSelected >= visibleRows {
		startIdx = m.phaseSelected - visibleRows + 1
	}
	endIdx := startIdx + visibleRows
	if endIdx > len(items) {
		endIdx = len(items)
	}

	for i := startIdx; i < endIdx; i++ {
		indicator := "  "
		if i == m.phaseSelected {
			indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("▸ ")
		}

		row := indicator + items[i]
		if i == m.phaseSelected {
			row = s.Selected.Width(m.width - 6).Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	if len(items) == 0 {
		b.WriteString(s.Subtle.Render("  No artifacts in this phase."))
	}

	if len(items) > visibleRows {
		b.WriteString(s.Subtle.Render(fmt.Sprintf("\n  %d / %d", m.phaseSelected+1, len(items))))
	}

	b.WriteString("\n\n")
	b.WriteString(s.Subtle.Render("  j/k:nav  Tab:phase  Esc:back"))

	return m.wrapPanel(b.String())
}

func (m Model) phaseItems() []string {
	s := m.styles
	var items []string

	switch m.phaseTab {
	case TabDiscovery:
		for _, f := range m.findings {
			item := s.CardValue.Render(f.Title)
			if f.Category != "" {
				item += " " + s.Subtle.Render("["+f.Category+"]")
			}
			if f.Priority != "" {
				item += " " + s.Subtle.Render(f.Priority)
			}
			items = append(items, item)
		}
		for _, t := range m.terms {
			item := s.Bold.Render(t.Term) + s.Subtle.Render(" — "+t.Definition)
			items = append(items, item)
		}

	case TabArchitecture:
		for _, d := range m.dossiers {
			item := s.CardValue.Render(d.DossierName)
			if d.StreamPattern != "" {
				item += " " + s.Subtle.Render("("+d.StreamPattern+")")
			}
			items = append(items, item)
		}
		for _, sp := range m.spokes {
			item := s.CardValue.Render(sp.SpokeName) + " " + s.Subtle.Render("["+sp.SpokeType+"]")
			if sp.Priority != "" {
				item += " " + s.Subtle.Render(sp.Priority)
			}
			items = append(items, item)
		}

	case TabTesting:
		for _, impl := range m.implementations {
			item := s.CardValue.Render(impl.SpokeID)
			if impl.ImplementationNotes != "" {
				item += " " + s.Subtle.Render("— "+impl.ImplementationNotes)
			}
			items = append(items, item)
		}
		for _, build := range m.builds {
			label := s.StatusOK.Render("PASS")
			if build.Result != "pass" {
				label = s.StatusError.Render("FAIL")
			}
			item := label
			if build.Notes != "" {
				item += " " + s.Subtle.Render(build.Notes)
			}
			items = append(items, item)
		}

	case TabDeployment:
		for _, dep := range m.deployments {
			item := s.CardValue.Render(dep.Environment) + " " + s.Subtle.Render("v"+dep.Version)
			if dep.Notes != "" {
				item += " " + s.Subtle.Render("— "+dep.Notes)
			}
			items = append(items, item)
		}
		for _, inc := range m.incidents {
			sevStyle := s.Subtle
			if inc.Severity == "critical" || inc.Severity == "high" {
				sevStyle = s.StatusError
			} else if inc.Severity == "medium" {
				sevStyle = s.StatusWarning
			}
			item := sevStyle.Render("["+inc.Severity+"]") + " " + s.CardValue.Render(inc.Description)
			if inc.Resolution != "" {
				item += " " + s.StatusOK.Render("(resolved)")
			}
			items = append(items, item)
		}
	}

	return items
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

func (m Model) fetchProjects() tea.Msg {
	projects, err := m.client.ListProjects()
	return projectsMsg{projects: projects, err: err}
}

func (m Model) fetchArtifacts(projectID string) tea.Cmd {
	return func() tea.Msg {
		var msg artifactsMsg

		// Fetch all artifact types in sequence (could be parallel with goroutines
		// but keeping it simple — the daemon is local so latency is low)
		msg.findings, _ = m.client.ListFindings(projectID)
		msg.terms, _ = m.client.ListTerms(projectID)
		msg.dossiers, _ = m.client.ListDossiers(projectID)
		msg.spokes, _ = m.client.ListSpokes(projectID)
		msg.implementations, _ = m.client.ListImplementations(projectID)
		msg.builds, _ = m.client.ListBuilds(projectID)
		msg.deployments, _ = m.client.ListDeployments(projectID)
		msg.incidents, _ = m.client.ListIncidents(projectID)

		return msg
	}
}

func phaseTabFromPhase(phase string) PhaseTab {
	switch strings.ToLower(phase) {
	case "discovery", "dna":
		return TabDiscovery
	case "architecture", "anp":
		return TabArchitecture
	case "testing", "tni":
		return TabTesting
	case "deployment", "dno":
		return TabDeployment
	default:
		return TabDiscovery
	}
}

func formatPhase(phase string) string {
	switch strings.ToLower(phase) {
	case "discovery", "dna":
		return "Discovery & Analysis"
	case "architecture", "anp":
		return "Architecture & Planning"
	case "testing", "tni":
		return "Testing & Integration"
	case "deployment", "dno":
		return "Deployment & Operations"
	case "initiated":
		return "Initiated"
	case "completed":
		return "Completed"
	default:
		return phase
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
