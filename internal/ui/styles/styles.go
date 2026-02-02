package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors - Macula theme
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#10B981") // Emerald
	Success   = lipgloss.Color("#22C55E") // Green
	Warning   = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	Text      = lipgloss.Color("#F3F4F6") // Light gray
	TextDark  = lipgloss.Color("#1F2937") // Dark gray
	BgDark    = lipgloss.Color("#111827") // Very dark
	BgMedium  = lipgloss.Color("#1F2937") // Dark
	Border    = lipgloss.Color("#374151") // Border gray

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(Muted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(Text).
			Background(Primary).
			Bold(true)

	// Content box
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	// Status indicators
	StatusOK = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	StatusError = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	StatusWarning = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	// Labels
	LabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Width(15)

	ValueStyle = lipgloss.NewStyle().
			Foreground(Text)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Border)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(Text)

	TableRowAltStyle = lipgloss.NewStyle().
				Foreground(Text).
				Background(BgMedium)

	// Selected item
	SelectedStyle = lipgloss.NewStyle().
			Background(Primary).
			Foreground(Text)

	// Logo
	LogoStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// Logo returns the hecate logo
func Logo() string {
	return LogoStyle.Render("🗝️ hecate-tui")
}

// StatusIndicator returns a colored status indicator
func StatusIndicator(status string) string {
	switch status {
	case "healthy", "connected", "ok":
		return StatusOK.Render("●")
	case "error", "failed", "disconnected":
		return StatusError.Render("●")
	case "warning", "degraded":
		return StatusWarning.Render("●")
	default:
		return lipgloss.NewStyle().Foreground(Muted).Render("●")
	}
}

// FormatUptime formats uptime in seconds to human readable
func FormatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return lipgloss.NewStyle().Foreground(Text).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render(string(rune('0'+days%10))),
				"d ",
				string(rune('0'+hours%10)),
				"h ",
				string(rune('0'+minutes%10)),
				"m",
			),
		)
	}
	if hours > 0 {
		return lipgloss.NewStyle().Foreground(Text).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render(string(rune('0'+hours%10))),
				"h ",
				string(rune('0'+minutes%10)),
				"m",
			),
		)
	}
	return lipgloss.NewStyle().Foreground(Text).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render(string(rune('0'+minutes%10))),
			"m",
		),
	)
}
