package me

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Section styles
var (
	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.Secondary).
				MarginBottom(1)

	SectionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Border).
			Padding(1, 2).
			MarginBottom(1)
)

// Profile styles
var (
	AvatarStyle = lipgloss.NewStyle().
			Foreground(styles.Primary).
			Bold(true)

	MRIStyle = lipgloss.NewStyle().
			Foreground(styles.Primary).
			Bold(true)

	RealmStyle = lipgloss.NewStyle().
			Foreground(styles.Secondary)

	PairedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	UnpairedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))
)

// Stats styles
var (
	StatLabelStyle = lipgloss.NewStyle().
			Width(16).
			Foreground(styles.Muted)

	StatValueStyle = lipgloss.NewStyle().
			Foreground(styles.Text)

	StatHighlightStyle = lipgloss.NewStyle().
				Foreground(styles.Primary).
				Bold(true)
)

// Settings styles
var (
	SettingsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.Primary).
				MarginBottom(1)

	SettingsGroupStyle = lipgloss.NewStyle().
				MarginBottom(1)

	SettingsLabelStyle = lipgloss.NewStyle().
				Width(20).
				Foreground(styles.Muted)

	SettingsValueStyle = lipgloss.NewStyle().
				Foreground(styles.Text)

	SettingsEditableStyle = lipgloss.NewStyle().
				Foreground(styles.Primary).
				Underline(true)

	SettingsDisabledStyle = lipgloss.NewStyle().
				Foreground(styles.Muted).
				Italic(true)

	SettingsSelectedStyle = lipgloss.NewStyle().
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	SettingsHintStyle = lipgloss.NewStyle().
				Foreground(styles.Muted).
				Italic(true)
)

// Menu item styles
var (
	MenuItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Background(styles.BgMedium).
				Padding(0, 1)

	MenuItemActiveStyle = lipgloss.NewStyle().
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)
)

// Helper function to render a setting row
func RenderSettingRow(label, value string, selected bool) string {
	labelPart := SettingsLabelStyle.Render(label)
	valuePart := SettingsValueStyle.Render(value)

	if selected {
		return SettingsSelectedStyle.Render(labelPart + valuePart)
	}
	return labelPart + valuePart
}

// Avatar art for profile
func AvatarArt() string {
	return `    ___
   /   \
  | o o |
  |  >  |
   \___/`
}
