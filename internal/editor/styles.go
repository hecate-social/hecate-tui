package editor

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Purple      = lipgloss.Color("#7C3AED")
	PurpleLight = lipgloss.Color("#A78BFA")
	Emerald     = lipgloss.Color("#10B981")
	Cyan        = lipgloss.Color("#06B6D4")
	Orange      = lipgloss.Color("#F97316")
	Amber       = lipgloss.Color("#F59E0B")
	Red         = lipgloss.Color("#EF4444")
	Gray100     = lipgloss.Color("#F3F4F6")
	Gray400     = lipgloss.Color("#9CA3AF")
	Gray500     = lipgloss.Color("#6B7280")
	Gray600     = lipgloss.Color("#4B5563")
	Gray700     = lipgloss.Color("#374151")
	Gray800     = lipgloss.Color("#1F2937")
	Gray900     = lipgloss.Color("#111827")

	// Title bar
	TitleBarStyle = lipgloss.NewStyle().
			Background(Purple).
			Foreground(Gray100).
			Padding(0, 1).
			Bold(true)

	TitleFileStyle = lipgloss.NewStyle().
			Foreground(Gray100)

	TitleModifiedStyle = lipgloss.NewStyle().
				Foreground(Amber).
				Bold(true)

	// Line numbers
	LineNumberStyle = lipgloss.NewStyle().
			Foreground(Gray600).
			Width(4).
			Align(lipgloss.Right).
			PaddingRight(1)

	LineNumberActiveStyle = lipgloss.NewStyle().
				Foreground(Amber).
				Width(4).
				Align(lipgloss.Right).
				PaddingRight(1).
				Bold(true)

	// Editor content
	EditorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Gray700).
			Padding(0, 1)

	EditorActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(0, 1)

	// Cursor line highlight
	CursorLineStyle = lipgloss.NewStyle().
			Background(Gray800)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Background(Gray800).
			Foreground(Gray400).
			Padding(0, 1)

	StatusModeStyle = lipgloss.NewStyle().
			Background(Purple).
			Foreground(Gray100).
			Padding(0, 1).
			Bold(true)

	StatusInsertStyle = lipgloss.NewStyle().
				Background(Emerald).
				Foreground(Gray900).
				Padding(0, 1).
				Bold(true)

	StatusPosStyle = lipgloss.NewStyle().
			Foreground(Gray400)

	// Messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Emerald).
			Bold(true)

	// Help
	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray500)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)
)
