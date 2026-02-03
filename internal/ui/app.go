package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
	"github.com/hecate-social/hecate-tui/internal/views/chat"
)

// Tab represents a navigation tab
type Tab int

const (
	TabStatus Tab = iota
	TabChat
	TabMesh
	TabCapabilities
	TabRPC
	TabLogs
)

func (t Tab) String() string {
	return [...]string{"Status", "Chat", "Mesh", "Capabilities", "RPC", "Logs"}[t]
}

// App is the main TUI application model
type App struct {
	client    *client.Client
	width     int
	height    int
	activeTab Tab
	tabs      []Tab

	// Data
	health       *client.Health
	identity     *client.Identity
	capabilities []client.Capability
	procedures   []client.Procedure

	// Views
	chatView chat.Model

	// UI state
	loading bool
	spinner spinner.Model
	err     error
}

// NewApp creates a new TUI application
func NewApp(hecateURL string) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	c := client.New(hecateURL)

	return &App{
		client:    c,
		activeTab: TabStatus,
		tabs:      []Tab{TabStatus, TabChat, TabMesh, TabCapabilities, TabRPC, TabLogs},
		chatView:  chat.New(c),
		spinner:   s,
		loading:   true,
	}
}

// Messages
type healthMsg struct {
	health   *client.Health
	identity *client.Identity
	err      error
}

type capabilitiesMsg struct {
	capabilities []client.Capability
	err          error
}

type proceduresMsg struct {
	procedures []client.Procedure
	err        error
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
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
		// Global quit
		if msg.String() == "q" && a.activeTab != TabChat {
			return a, tea.Quit
		}
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		// Tab navigation (only when not in chat or chat allows it)
		if a.activeTab != TabChat {
			switch msg.String() {
			case "tab":
				a.activeTab = Tab((int(a.activeTab) + 1) % len(a.tabs))
				a.onTabChange()
				return a, a.fetchDataForTab()
			case "shift+tab":
				a.activeTab = Tab((int(a.activeTab) - 1 + len(a.tabs)) % len(a.tabs))
				a.onTabChange()
				return a, a.fetchDataForTab()
			case "r":
				a.loading = true
				return a, a.fetchDataForTab()
			case "1":
				a.activeTab = TabStatus
				a.onTabChange()
				return a, a.fetchHealth
			case "2":
				a.activeTab = TabChat
				a.onTabChange()
				return a, nil
			case "3":
				a.activeTab = TabMesh
				a.onTabChange()
				return a, nil
			case "4":
				a.activeTab = TabCapabilities
				a.onTabChange()
				return a, a.fetchCapabilities
			case "5":
				a.activeTab = TabRPC
				a.onTabChange()
				return a, a.fetchProcedures
			case "6":
				a.activeTab = TabLogs
				a.onTabChange()
				return a, nil
			case "c":
				// Quick key to jump to chat
				a.activeTab = TabChat
				a.onTabChange()
				return a, nil
			}
		} else {
			// In chat view, allow escape to leave chat
			if msg.String() == "esc" {
				a.activeTab = TabStatus
				a.chatView.Blur()
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Update chat view size
		a.chatView.SetSize(msg.Width, msg.Height-6) // Account for header/tabs/footer
	}

	// Always update spinner
	if tickMsg, ok := msg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(tickMsg)
		cmds = append(cmds, cmd)
	}

	// Handle data messages
	switch msg := msg.(type) {
	case healthMsg:
		a.loading = false
		a.health = msg.health
		a.identity = msg.identity
		a.err = msg.err

	case capabilitiesMsg:
		a.loading = false
		a.capabilities = msg.capabilities
		a.err = msg.err

	case proceduresMsg:
		a.loading = false
		a.procedures = msg.procedures
		a.err = msg.err
	}

	// Update chat view if active
	if a.activeTab == TabChat {
		var cmd tea.Cmd
		a.chatView, cmd = a.chatView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a *App) onTabChange() {
	if a.activeTab == TabChat {
		a.chatView.Focus()
	} else {
		a.chatView.Blur()
	}
}

// View renders the UI
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	b.WriteString(a.renderHeader())
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(a.renderTabs())
	b.WriteString("\n\n")

	// Content
	if a.loading {
		b.WriteString(a.spinner.View() + " Loading...")
	} else if a.err != nil {
		b.WriteString(styles.StatusError.Render("Error: " + a.err.Error()))
	} else {
		b.WriteString(a.renderContent())
	}

	// Footer
	b.WriteString("\n\n")
	b.WriteString(a.renderFooter())

	return b.String()
}

func (a *App) renderHeader() string {
	logo := styles.Logo()
	var status string
	if a.health != nil {
		status = styles.StatusIndicator(a.health.Status) + " " + a.health.Status
	} else {
		status = styles.StatusIndicator("unknown") + " unknown"
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		logo,
		strings.Repeat(" ", a.width-lipgloss.Width(logo)-lipgloss.Width(status)-4),
		status,
	)

	return header
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
	case TabStatus:
		return a.renderStatusView()
	case TabChat:
		return a.chatView.View()
	case TabMesh:
		return a.renderMeshView()
	case TabCapabilities:
		return a.renderCapabilitiesView()
	case TabRPC:
		return a.renderRPCView()
	case TabLogs:
		return a.renderLogsView()
	default:
		return "Unknown view"
	}
}

func (a *App) renderStatusView() string {
	if a.health == nil {
		return "No health data available"
	}

	var rows []string

	// Health info
	rows = append(rows, styles.TitleStyle.Render("Daemon Status"))

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		styles.LabelStyle.Render("Status:"),
		styles.StatusIndicator(a.health.Status)+" "+a.health.Status,
	))

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		styles.LabelStyle.Render("Version:"),
		styles.ValueStyle.Render(a.health.Version),
	))

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		styles.LabelStyle.Render("Uptime:"),
		styles.ValueStyle.Render(formatUptime(a.health.UptimeSeconds)),
	))

	// Identity info
	if a.identity != nil {
		rows = append(rows, "")
		rows = append(rows, styles.TitleStyle.Render("Identity"))

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("MRI:"),
			styles.ValueStyle.Render(a.identity.Identity),
		))

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			styles.LabelStyle.Render("Created:"),
			styles.ValueStyle.Render(a.identity.CreatedAt),
		))
	}

	return styles.BoxStyle.Width(a.width - 4).Render(strings.Join(rows, "\n"))
}

func (a *App) renderMeshView() string {
	return styles.BoxStyle.Width(a.width - 4).Render(
		styles.TitleStyle.Render("Mesh Topology") + "\n\n" +
			styles.SubtitleStyle.Render("Coming soon..."),
	)
}

func (a *App) renderCapabilitiesView() string {
	var rows []string
	rows = append(rows, styles.TitleStyle.Render("Discovered Capabilities"))

	if len(a.capabilities) == 0 {
		rows = append(rows, styles.SubtitleStyle.Render("No capabilities discovered"))
	} else {
		for _, cap := range a.capabilities {
			rows = append(rows, "")
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("MRI:"),
				styles.ValueStyle.Render(cap.MRI),
			))
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("Agent:"),
				styles.ValueStyle.Render(cap.AgentIdentity),
			))
			if cap.Description != "" {
				rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
					styles.LabelStyle.Render("Description:"),
					styles.ValueStyle.Render(cap.Description),
				))
			}
			if len(cap.Tags) > 0 {
				rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
					styles.LabelStyle.Render("Tags:"),
					styles.ValueStyle.Render(strings.Join(cap.Tags, ", ")),
				))
			}
		}
	}

	return styles.BoxStyle.Width(a.width - 4).Render(strings.Join(rows, "\n"))
}

func (a *App) renderRPCView() string {
	var rows []string
	rows = append(rows, styles.TitleStyle.Render("Registered Procedures"))

	if len(a.procedures) == 0 {
		rows = append(rows, styles.SubtitleStyle.Render("No procedures registered"))
	} else {
		for _, proc := range a.procedures {
			rows = append(rows, "")
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("Name:"),
				styles.ValueStyle.Render(proc.Name),
			))
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("MRI:"),
				styles.ValueStyle.Render(proc.MRI),
			))
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				styles.LabelStyle.Render("Endpoint:"),
				styles.ValueStyle.Render(proc.Endpoint),
			))
		}
	}

	return styles.BoxStyle.Width(a.width - 4).Render(strings.Join(rows, "\n"))
}

func (a *App) renderLogsView() string {
	return styles.BoxStyle.Width(a.width - 4).Render(
		styles.TitleStyle.Render("Daemon Logs") + "\n\n" +
			styles.SubtitleStyle.Render("Coming soon..."),
	)
}

func (a *App) renderFooter() string {
	if a.activeTab == TabChat {
		return styles.HelpStyle.Render("Enter: send • Tab: model • Ctrl+L: clear • Esc: back • ↑↓: scroll")
	}
	return styles.HelpStyle.Render("Tab/1-6: switch view • c: chat • r: refresh • q: quit")
}

// Commands
func (a *App) fetchHealth() tea.Msg {
	health, err := a.client.GetHealth()
	if err != nil {
		return healthMsg{err: err}
	}

	identity, _ := a.client.GetIdentity()

	return healthMsg{health: health, identity: identity}
}

func (a *App) fetchCapabilities() tea.Msg {
	caps, err := a.client.DiscoverCapabilities("", "", 100)
	return capabilitiesMsg{capabilities: caps, err: err}
}

func (a *App) fetchProcedures() tea.Msg {
	procs, err := a.client.ListProcedures()
	return proceduresMsg{procedures: procs, err: err}
}

func (a *App) fetchDataForTab() tea.Cmd {
	switch a.activeTab {
	case TabStatus:
		return a.fetchHealth
	case TabChat:
		return a.chatView.Init()
	case TabCapabilities:
		return a.fetchCapabilities
	case TabRPC:
		return a.fetchProcedures
	default:
		return nil
	}
}

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
