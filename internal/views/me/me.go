package me

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ModeProfile ViewMode = iota
	ModeSettings
)

// SettingItem represents a configurable setting
type SettingItem struct {
	Label       string
	Value       string
	Editable    bool
	Description string
}

// Model is the me view model
type Model struct {
	client   *client.Client
	width    int
	height   int
	focused  bool
	loading  bool
	spinner  spinner.Model
	identity *client.Identity
	health   *client.Health
	err      error

	// View state
	mode ViewMode

	// Settings
	settings       []SettingItem
	selectedSetting int

	// Stats
	capabilityCount   int
	subscriptionCount int
}

// New creates a new me view
func New(c *client.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return Model{
		client:   c,
		spinner:  s,
		mode:     ModeProfile,
		settings: defaultSettings(),
	}
}

func defaultSettings() []SettingItem {
	return []SettingItem{
		{Label: "Daemon URL", Value: "http://localhost:4444", Editable: false, Description: "Hecate daemon API endpoint"},
		{Label: "Theme", Value: "Dark", Editable: true, Description: "UI color theme"},
		{Label: "Auto-refresh", Value: "30s", Editable: true, Description: "Automatic data refresh interval"},
		{Label: "Notifications", Value: "Enabled", Editable: true, Description: "Show system notifications"},
		{Label: "Debug Mode", Value: "Off", Editable: true, Description: "Show debug information"},
	}
}

// Messages
type profileDataMsg struct {
	identity          *client.Identity
	health            *client.Health
	capabilityCount   int
	subscriptionCount int
	err               error
}

// Init initializes the view
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchProfileData,
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

		switch m.mode {
		case ModeSettings:
			return m.updateSettings(msg)
		default:
			return m.updateProfile(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case profileDataMsg:
		m.loading = false
		m.identity = msg.identity
		m.health = msg.health
		m.capabilityCount = msg.capabilityCount
		m.subscriptionCount = msg.subscriptionCount
		m.err = msg.err
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateProfile(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		m.mode = ModeSettings
		m.selectedSetting = 0
	case "r":
		m.loading = true
		return m, m.fetchProfileData
	case "p":
		// Navigate to pair tab - handled by parent
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = ModeProfile
	case "up", "k":
		if m.selectedSetting > 0 {
			m.selectedSetting--
		}
	case "down", "j":
		if m.selectedSetting < len(m.settings)-1 {
			m.selectedSetting++
		}
	case "enter", " ":
		// Toggle or edit setting
		setting := &m.settings[m.selectedSetting]
		if setting.Editable {
			m.toggleSetting(setting)
		}
	}
	return m, nil
}

func (m *Model) toggleSetting(setting *SettingItem) {
	switch setting.Label {
	case "Theme":
		if setting.Value == "Dark" {
			setting.Value = "Light"
		} else {
			setting.Value = "Dark"
		}
	case "Notifications":
		if setting.Value == "Enabled" {
			setting.Value = "Disabled"
		} else {
			setting.Value = "Enabled"
		}
	case "Debug Mode":
		if setting.Value == "Off" {
			setting.Value = "On"
		} else {
			setting.Value = "Off"
		}
	case "Auto-refresh":
		values := []string{"10s", "30s", "60s", "Off"}
		for i, v := range values {
			if v == setting.Value {
				setting.Value = values[(i+1)%len(values)]
				break
			}
		}
	}
}

// View renders the view
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case ModeSettings:
		return m.renderSettings()
	default:
		return m.renderProfile()
	}
}

func (m Model) renderProfile() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.TitleStyle.Render("Me"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Loading profile...")
		return styles.BoxStyle.Width(m.width - 4).Render(b.String())
	}

	if m.err != nil {
		b.WriteString(styles.StatusError.Render("Error: " + m.err.Error()))
		return styles.BoxStyle.Width(m.width - 4).Render(b.String())
	}

	// Profile card
	b.WriteString(m.renderProfileCard())
	b.WriteString("\n\n")

	// Stats
	b.WriteString(m.renderStats())
	b.WriteString("\n\n")

	// Quick actions
	b.WriteString(m.renderActions())

	return styles.BoxStyle.Width(m.width - 4).Render(b.String())
}

func (m Model) renderProfileCard() string {
	var rows []string

	// Avatar and identity side by side
	avatar := AvatarStyle.Render(AvatarArt())

	var identityInfo []string
	if m.identity == nil {
		identityInfo = append(identityInfo, UnpairedStyle.Render("No identity configured"))
		identityInfo = append(identityInfo, "")
		identityInfo = append(identityInfo, SettingsHintStyle.Render("Run: hecate init"))
	} else {
		// MRI
		mri := m.identity.Identity
		if len(mri) > 50 {
			mri = mri[:47] + "..."
		}
		identityInfo = append(identityInfo, MRIStyle.Render(mri))

		// Realm
		realm := parseRealm(m.identity.Identity)
		if realm != "" {
			identityInfo = append(identityInfo, RealmStyle.Render("Realm: "+realm))
			identityInfo = append(identityInfo, PairedStyle.Render("Paired"))
		} else {
			identityInfo = append(identityInfo, UnpairedStyle.Render("Not paired"))
		}

		// Created
		if m.identity.CreatedAt != "" {
			identityInfo = append(identityInfo, SettingsHintStyle.Render("Since "+m.identity.CreatedAt))
		}
	}

	identityBlock := strings.Join(identityInfo, "\n")

	// Join avatar and identity
	profile := lipgloss.JoinHorizontal(lipgloss.Top,
		avatar,
		"    ",
		identityBlock,
	)

	rows = append(rows, SectionTitleStyle.Render("Identity"))
	rows = append(rows, profile)

	return SectionBoxStyle.Width(m.width - 8).Render(strings.Join(rows, "\n"))
}

func (m Model) renderStats() string {
	var rows []string

	rows = append(rows, SectionTitleStyle.Render("Statistics"))
	rows = append(rows, "")

	// Build stat rows
	stats := []struct {
		label string
		value string
	}{
		{"Capabilities:", fmt.Sprintf("%d announced", m.capabilityCount)},
		{"Subscriptions:", fmt.Sprintf("%d active", m.subscriptionCount)},
	}

	if m.health != nil {
		stats = append(stats, struct {
			label string
			value string
		}{"Daemon:", StatHighlightStyle.Render("Online") + " (v" + m.health.Version + ")"})
	} else {
		stats = append(stats, struct {
			label string
			value string
		}{"Daemon:", UnpairedStyle.Render("Offline")})
	}

	for _, stat := range stats {
		rows = append(rows, "  "+lipgloss.JoinHorizontal(lipgloss.Left,
			StatLabelStyle.Render(stat.label),
			StatValueStyle.Render(stat.value)))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderActions() string {
	actions := []string{
		SettingsHintStyle.Render("[s] Settings"),
		SettingsHintStyle.Render("[p] Pairing"),
		SettingsHintStyle.Render("[r] Refresh"),
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, actions[0], "  ", actions[1], "  ", actions[2])
}

func (m Model) renderSettings() string {
	var b strings.Builder

	// Title
	b.WriteString(SettingsTitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	// Settings list
	for i, setting := range m.settings {
		row := m.renderSettingRow(setting, i == m.selectedSetting)
		b.WriteString(row)
		b.WriteString("\n")

		// Show description for selected item
		if i == m.selectedSetting {
			desc := SettingsHintStyle.Render("  " + setting.Description)
			b.WriteString(desc)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Help
	help := SettingsHintStyle.Render("↑↓: navigate • Enter: toggle • Esc: back")
	b.WriteString(help)

	return SectionBoxStyle.Width(m.width - 4).Height(m.height - 4).Render(b.String())
}

func (m Model) renderSettingRow(setting SettingItem, selected bool) string {
	label := SettingsLabelStyle.Render(setting.Label + ":")

	valueStyle := SettingsValueStyle
	if !setting.Editable {
		valueStyle = SettingsDisabledStyle
	}
	value := valueStyle.Render(setting.Value)

	row := label + value

	if selected {
		indicator := lipgloss.NewStyle().Foreground(styles.Primary).Render("> ")
		row = indicator + row
		if setting.Editable {
			row = MenuItemSelectedStyle.Render(row)
		}
	} else {
		row = "  " + row
	}

	return row
}

// Commands
func (m Model) fetchProfileData() tea.Msg {
	identity, _ := m.client.GetIdentity()
	health, _ := m.client.GetHealth()

	caps, _ := m.client.DiscoverCapabilities("", "", 1000)
	capCount := 0
	if caps != nil {
		capCount = len(caps)
	}

	subs, _ := m.client.ListSubscriptions()
	subCount := 0
	if subs != nil {
		subCount = len(subs)
	}

	return profileDataMsg{
		identity:          identity,
		health:            health,
		capabilityCount:   capCount,
		subscriptionCount: subCount,
	}
}

// View interface implementation

// Name returns the tab label
func (m Model) Name() string {
	return "Me"
}

// ShortHelp returns help text
func (m Model) ShortHelp() string {
	if m.mode == ModeSettings {
		return "↑↓: navigate • Enter: toggle • Esc: back"
	}
	return "s: settings • p: pairing • r: refresh"
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
	// Return to profile mode when blurred
	m.mode = ModeProfile
}

// Helper functions

func parseRealm(mri string) string {
	// mri:agent:io.macula/name -> io.macula
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
