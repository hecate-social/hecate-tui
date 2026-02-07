package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// GeoBlockedModel renders the blocked region message.
type GeoBlockedModel struct {
	CountryCode string
	CountryName string
	Width       int
	Height      int
}

// NewGeoBlocked creates a new blocked region display model.
func NewGeoBlocked(countryCode, countryName string) GeoBlockedModel {
	return GeoBlockedModel{
		CountryCode: countryCode,
		CountryName: countryName,
	}
}

// SetSize updates the terminal dimensions.
func (m *GeoBlockedModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// View renders the blocked region message.
func (m GeoBlockedModel) View() string {
	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196"))

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	contactStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39"))

	// Build the message
	title := titleStyle.Render("ACCESS RESTRICTED")

	location := m.CountryName
	if location == "" {
		location = m.CountryCode
	}
	if m.CountryCode != "" && m.CountryName != "" {
		location = fmt.Sprintf("%s (%s)", m.CountryName, m.CountryCode)
	}

	subtitle := subtitleStyle.Render(fmt.Sprintf("Hecate is not available in %s.", location))

	contact := contactStyle.Render("If you believe this is an error, please contact support@hecate.social")

	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		title,
		"",
		subtitle,
		"",
		contact,
		"",
	)

	box := borderStyle.Render(content)

	// Center on screen
	return lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

// RenderGeoBlockedMessage renders a simple terminal message for blocked regions.
// This is used before the TUI starts.
func RenderGeoBlockedMessage(countryCode, countryName string) string {
	location := countryName
	if location == "" {
		location = countryCode
	}
	if countryCode != "" && countryName != "" {
		location = fmt.Sprintf("%s (%s)", countryName, countryCode)
	}

	return fmt.Sprintf(`
╭──────────────────────────────────────────────────╮
│              ACCESS RESTRICTED                   │
├──────────────────────────────────────────────────┤
│                                                  │
│  Hecate is not available in %s.
│                                                  │
│  If you believe this is an error, please         │
│  contact support@hecate.social                   │
│                                                  │
╰──────────────────────────────────────────────────╯
`, location)
}
