package pair

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Status colors
var (
	PairedColor   = lipgloss.Color("#10B981") // Emerald
	WaitingColor  = lipgloss.Color("#F59E0B") // Amber
	ErrorColor    = lipgloss.Color("#EF4444") // Red
	IdleColor     = lipgloss.Color("#6B7280") // Gray
)

// Status styles
var (
	PairedStatusStyle = lipgloss.NewStyle().
				Foreground(PairedColor).
				Bold(true)

	WaitingStatusStyle = lipgloss.NewStyle().
				Foreground(WaitingColor).
				Bold(true)

	ErrorStatusStyle = lipgloss.NewStyle().
				Foreground(ErrorColor).
				Bold(true)

	IdleStatusStyle = lipgloss.NewStyle().
			Foreground(IdleColor)
)

// Code display styles
var (
	CodeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(styles.Primary).
			Padding(1, 4).
			Align(lipgloss.Center)

	CodeStyle = lipgloss.NewStyle().
			Foreground(styles.Primary).
			Bold(true).
			Align(lipgloss.Center)

	CodeLabelStyle = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Align(lipgloss.Center)
)

// Instruction styles
var (
	StepNumberStyle = lipgloss.NewStyle().
			Foreground(styles.Primary).
			Bold(true)

	StepTextStyle = lipgloss.NewStyle().
			Foreground(styles.Text)

	URLStyle = lipgloss.NewStyle().
			Foreground(styles.Secondary).
			Underline(true)

	HintStyle = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true)
)

// Progress styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(styles.Primary)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(styles.Muted)
)

// Section styles
var (
	SectionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Border).
			Padding(1, 2)

	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.Secondary).
				MarginBottom(1)
)

// Helper functions
func StatusIndicator(status string) string {
	switch status {
	case "paired":
		return PairedStatusStyle.Render("Paired")
	case "waiting":
		return WaitingStatusStyle.Render("Waiting for confirmation...")
	case "error":
		return ErrorStatusStyle.Render("Error")
	default:
		return IdleStatusStyle.Render("Not paired")
	}
}

func RenderStep(number int, text string) string {
	num := StepNumberStyle.Render(string(rune('0'+number)) + ".")
	txt := StepTextStyle.Render(" " + text)
	return num + txt
}

// QR placeholder art
func QRPlaceholder() string {
	return `
  ██████████████
  ██          ██
  ██  ██████  ██
  ██  ██████  ██
  ██  ██████  ██
  ██          ██
  ██████████████
       ████
  ████████████
  ██    ██  ██
  ██████████████`
}
