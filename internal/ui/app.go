package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
	"github.com/hecate-social/hecate-tui/internal/views"
	"github.com/hecate-social/hecate-tui/internal/views/browse"
	"github.com/hecate-social/hecate-tui/internal/views/chat"
	"github.com/hecate-social/hecate-tui/internal/views/me"
	"github.com/hecate-social/hecate-tui/internal/views/monitor"
	"github.com/hecate-social/hecate-tui/internal/views/pair"
	"github.com/hecate-social/hecate-tui/internal/views/projects"
)

// App is the main TUI application model
type App struct {
	client    *client.Client
	width     int
	height    int
	activeTab views.Tab
	tabs      []views.Tab

	// Views
	chatView     chat.Model
	browseView   browse.Model
	projectsView projects.Model
	monitorView  monitor.Model
	pairView     pair.Model
	meView       me.Model

	// Health for header status
	health *client.Health

	// UI state
	spinner spinner.Model
}

// NewApp creates a new TUI application
func NewApp(hecateURL string) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	c := client.New(hecateURL)

	return &App{
		client:       c,
		activeTab:    views.TabChat,
		tabs:         views.AllTabs(),
		chatView:     chat.New(c),
		browseView:   browse.New(c),
		projectsView: projects.New(),
		monitorView:  monitor.New(c),
		pairView:     pair.New(c),
		meView:       me.New(c),
		spinner:      s,
	}
}

// Messages
type healthMsg struct {
	health *client.Health
	err    error
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	// Focus the initial view
	a.chatView.Focus()

	return tea.Batch(
		a.spinner.Tick,
		a.fetchHealth,
		a.chatView.Init(),
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit (except in chat which handles its own keys)
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		// Only handle global keys when not in an input-focused view
		if a.activeTab != views.TabChat {
			if msg.String() == "q" {
				return a, tea.Quit
			}

			// Number key navigation
			switch msg.String() {
			case "1":
				return a, a.switchTab(views.TabChat)
			case "2":
				return a, a.switchTab(views.TabBrowse)
			case "3":
				return a, a.switchTab(views.TabProjects)
			case "4":
				return a, a.switchTab(views.TabMonitor)
			case "5":
				return a, a.switchTab(views.TabPair)
			case "6":
				return a, a.switchTab(views.TabMe)
			case "tab":
				nextTab := views.Tab((int(a.activeTab) + 1) % len(a.tabs))
				return a, a.switchTab(nextTab)
			case "shift+tab":
				prevTab := views.Tab((int(a.activeTab) - 1 + len(a.tabs)) % len(a.tabs))
				return a, a.switchTab(prevTab)
			}
		} else {
			// In chat view, Esc goes back to monitor
			if msg.String() == "esc" && !a.chatView.IsStreaming() {
				return a, a.switchTab(views.TabMonitor)
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Content area: total - header(1) - spacing(2) - tabs(1) - spacing(1) - spacing(1) - footer(1) = 7
		viewHeight := msg.Height - 7
		if viewHeight < 10 {
			viewHeight = 10
		}
		a.chatView.SetSize(msg.Width, viewHeight)
		a.browseView.SetSize(msg.Width, viewHeight)
		a.projectsView.SetSize(msg.Width, viewHeight)
		a.monitorView.SetSize(msg.Width, viewHeight)
		a.pairView.SetSize(msg.Width, viewHeight)
		a.meView.SetSize(msg.Width, viewHeight)

	case healthMsg:
		a.health = msg.health
	}

	// Always update spinner
	if tickMsg, ok := msg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(tickMsg)
		cmds = append(cmds, cmd)
	}

	// Update active view
	var cmd tea.Cmd
	switch a.activeTab {
	case views.TabChat:
		a.chatView, cmd = a.chatView.Update(msg)
	case views.TabBrowse:
		var m tea.Model
		m, cmd = a.browseView.Update(msg)
		a.browseView = m.(browse.Model)
	case views.TabProjects:
		var m tea.Model
		m, cmd = a.projectsView.Update(msg)
		a.projectsView = m.(projects.Model)
	case views.TabMonitor:
		var m tea.Model
		m, cmd = a.monitorView.Update(msg)
		a.monitorView = m.(monitor.Model)
	case views.TabPair:
		var m tea.Model
		m, cmd = a.pairView.Update(msg)
		a.pairView = m.(pair.Model)
	case views.TabMe:
		var m tea.Model
		m, cmd = a.meView.Update(msg)
		a.meView = m.(me.Model)
	}
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *App) switchTab(tab views.Tab) tea.Cmd {
	// Blur current view
	switch a.activeTab {
	case views.TabChat:
		a.chatView.Blur()
	case views.TabBrowse:
		a.browseView.Blur()
	case views.TabProjects:
		a.projectsView.Blur()
	case views.TabMonitor:
		a.monitorView.Blur()
	case views.TabPair:
		a.pairView.Blur()
	case views.TabMe:
		a.meView.Blur()
	}

	a.activeTab = tab

	// Focus new view and return init command
	switch tab {
	case views.TabChat:
		a.chatView.Focus()
		return a.chatView.Init()
	case views.TabBrowse:
		a.browseView.Focus()
		return a.browseView.Init()
	case views.TabProjects:
		a.projectsView.Focus()
		return a.projectsView.Init()
	case views.TabMonitor:
		a.monitorView.Focus()
		return a.monitorView.Init()
	case views.TabPair:
		a.pairView.Focus()
		return a.pairView.Init()
	case views.TabMe:
		a.meView.Focus()
		return a.meView.Init()
	}

	return nil
}

// View renders the UI
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Fixed header section
	header := a.renderHeader() + "\n\n" + a.renderTabs() + "\n"

	// Fixed footer section
	footer := "\n" + a.renderFooter()

	// Content height = total - header lines (4) - footer lines (2) - 1 buffer
	contentHeight := a.height - 7
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Constrain content to exact height
	content := lipgloss.NewStyle().
		Height(contentHeight).
		MaxHeight(contentHeight).
		Render(a.renderContent())

	return header + content + footer
}

func (a *App) renderHeader() string {
	logo := styles.Logo()
	var status string
	if a.health != nil {
		status = styles.StatusIndicator(a.health.Status) + " " + a.health.Status
	} else {
		status = styles.StatusIndicator("unknown") + " connecting..."
	}

	spacer := a.width - lipgloss.Width(logo) - lipgloss.Width(status) - 4
	if spacer < 1 {
		spacer = 1
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		logo,
		strings.Repeat(" ", spacer),
		status,
	)
}

func (a *App) renderTabs() string {
	var tabs []string
	for i, tab := range a.tabs {
		style := styles.TabStyle
		if tab == a.activeTab {
			style = styles.ActiveTabStyle
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d %s", i+1, tab.String())))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (a *App) renderContent() string {
	switch a.activeTab {
	case views.TabChat:
		return a.chatView.View()
	case views.TabBrowse:
		return a.browseView.View()
	case views.TabProjects:
		return a.projectsView.View()
	case views.TabMonitor:
		return a.monitorView.View()
	case views.TabPair:
		return a.pairView.View()
	case views.TabMe:
		return a.meView.View()
	default:
		return "Unknown view"
	}
}

func (a *App) renderFooter() string {
	var help string
	switch a.activeTab {
	case views.TabChat:
		help = a.chatView.ShortHelp()
	case views.TabBrowse:
		help = a.browseView.ShortHelp()
	case views.TabProjects:
		help = a.projectsView.ShortHelp()
	case views.TabMonitor:
		help = a.monitorView.ShortHelp()
	case views.TabPair:
		help = a.pairView.ShortHelp()
	case views.TabMe:
		help = a.meView.ShortHelp()
	}

	globalHelp := "1-6: tabs • q: quit"
	if a.activeTab == views.TabChat {
		globalHelp = "Esc: back • ctrl+c: quit"
	}

	return styles.HelpStyle.Render(help + " │ " + globalHelp)
}

// Commands
func (a *App) fetchHealth() tea.Msg {
	health, _ := a.client.GetHealth()
	return healthMsg{health: health}
}
