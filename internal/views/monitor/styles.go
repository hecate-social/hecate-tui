package monitor

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Status colors
var (
	HealthyColor    = lipgloss.Color("#10B981") // Emerald
	DegradedColor   = lipgloss.Color("#F59E0B") // Amber
	UnhealthyColor  = lipgloss.Color("#EF4444") // Red
	DisconnectColor = lipgloss.Color("#6B7280") // Gray
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

// Status indicator styles
var (
	StatusHealthyStyle = lipgloss.NewStyle().
				Foreground(HealthyColor).
				Bold(true)

	StatusDegradedStyle = lipgloss.NewStyle().
				Foreground(DegradedColor).
				Bold(true)

	StatusUnhealthyStyle = lipgloss.NewStyle().
				Foreground(UnhealthyColor).
				Bold(true)

	StatusDisconnectStyle = lipgloss.NewStyle().
				Foreground(DisconnectColor)
)

// Label and value styles
var (
	RowLabelStyle = lipgloss.NewStyle().
			Width(14).
			Foreground(styles.Muted)

	RowValueStyle = lipgloss.NewStyle().
			Foreground(styles.Text)

	RowHighlightStyle = lipgloss.NewStyle().
				Foreground(styles.Primary).
				Bold(true)
)

// Stat card styles
var (
	StatCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Border).
			Padding(0, 2).
			Width(20).
			Align(lipgloss.Center)

	StatValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Primary).
			Align(lipgloss.Center)

	StatLabelStyle = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Align(lipgloss.Center)
)

// Helper functions
func StatusIndicator(status string) string {
	switch status {
	case "healthy", "running", "connected":
		return StatusHealthyStyle.Render("●")
	case "degraded", "connecting":
		return StatusDegradedStyle.Render("●")
	case "unhealthy", "error", "failed":
		return StatusUnhealthyStyle.Render("●")
	default:
		return StatusDisconnectStyle.Render("○")
	}
}

func StatusText(status string) string {
	switch status {
	case "healthy", "running", "connected":
		return StatusHealthyStyle.Render(status)
	case "degraded", "connecting":
		return StatusDegradedStyle.Render(status)
	case "unhealthy", "error", "failed":
		return StatusUnhealthyStyle.Render(status)
	default:
		return StatusDisconnectStyle.Render(status)
	}
}

func RenderStatCard(value, label string) string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		StatValueStyle.Render(value),
		StatLabelStyle.Render(label),
	)
	return StatCardStyle.Render(content)
}
