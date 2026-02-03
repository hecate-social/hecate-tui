package projects

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Purple      = lipgloss.Color("#7C3AED")
	PurpleLight = lipgloss.Color("#A78BFA")
	PurpleDark  = lipgloss.Color("#5B21B6")
	Emerald     = lipgloss.Color("#10B981")
	Cyan        = lipgloss.Color("#06B6D4")
	Orange      = lipgloss.Color("#F97316")
	Amber       = lipgloss.Color("#F59E0B")
	Gold        = lipgloss.Color("#FCD34D")
	Gray100     = lipgloss.Color("#F3F4F6")
	Gray400     = lipgloss.Color("#9CA3AF")
	Gray500     = lipgloss.Color("#6B7280")
	Gray600     = lipgloss.Color("#4B5563")
	Gray700     = lipgloss.Color("#374151")
	Gray800     = lipgloss.Color("#1F2937")

	// Title
	TitleStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true).
			MarginBottom(1)

	// Project list
	ProjectListStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Gray700).
				Padding(1, 2)

	ProjectItemStyle = lipgloss.NewStyle().
				Foreground(Gray100).
				PaddingLeft(2)

	ProjectItemSelectedStyle = lipgloss.NewStyle().
					Foreground(Gray100).
					Background(PurpleDark).
					Bold(true).
					PaddingLeft(2)

	ProjectNameStyle = lipgloss.NewStyle().
				Foreground(Gray100).
				Bold(true)

	ProjectPathStyle = lipgloss.NewStyle().
				Foreground(Gray500).
				Italic(true)

	ProjectTypeStyle = lipgloss.NewStyle().
				Foreground(Cyan)

	// Phase tabs
	PhaseTabStyle = lipgloss.NewStyle().
			Foreground(Gray400).
			Padding(0, 2).
			MarginRight(1)

	PhaseTabActiveStyle = lipgloss.NewStyle().
				Foreground(Gray100).
				Background(Purple).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)

	PhaseTabBarStyle = lipgloss.NewStyle().
				MarginTop(1).
				MarginBottom(1)

	// Phase content
	PhaseContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(1, 2)

	PhaseHeaderStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Bold(true)

	PhaseDescStyle = lipgloss.NewStyle().
			Foreground(Gray400).
			Italic(true)

	// Empty state
	EmptyStyle = lipgloss.NewStyle().
			Foreground(Gray500).
			Italic(true).
			Align(lipgloss.Center)

	ComingSoonStyle = lipgloss.NewStyle().
			Foreground(Amber).
			Bold(true).
			Align(lipgloss.Center)

	// Help
	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray500).
			MarginTop(1)

	// Workspace indicator
	WorkspaceActiveStyle = lipgloss.NewStyle().
				Foreground(Emerald).
				Bold(true)

	WorkspaceMissingStyle = lipgloss.NewStyle().
				Foreground(Orange)
)
