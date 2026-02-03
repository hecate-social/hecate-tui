package browse

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/ui/styles"
)

// Colors specific to browse view
var (
	LocalColor  = lipgloss.Color("#10B981") // Emerald for local
	RemoteColor = lipgloss.Color("#6B7280") // Gray for remote
	MatchColor  = lipgloss.Color("#FBBF24") // Amber for search matches
)

// Container styles
var (
	ListContainerStyle = lipgloss.NewStyle().
				Padding(0, 1)

	DetailContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.Border).
				Padding(1, 2)
)

// List item styles
var (
	ItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	SelectedUnfocusedStyle = lipgloss.NewStyle().
				Background(styles.BgMedium).
				Padding(0, 1)
)

// Header styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Muted).
			MarginBottom(1)

	ColumnNameStyle = lipgloss.NewStyle().
			Width(32).
			Foreground(styles.Muted)

	ColumnSourceStyle = lipgloss.NewStyle().
				Width(10).
				Foreground(styles.Muted)

	ColumnTagsStyle = lipgloss.NewStyle().
			Foreground(styles.Muted)
)

// Source indicator styles
var (
	LocalSourceStyle = lipgloss.NewStyle().
				Foreground(LocalColor).
				Bold(true)

	RemoteSourceStyle = lipgloss.NewStyle().
				Foreground(RemoteColor)
)

// Search styles
var (
	SearchContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.Primary).
				Padding(0, 1).
				MarginBottom(1)

	SearchLabelStyle = lipgloss.NewStyle().
				Foreground(styles.Primary).
				Bold(true)

	SearchInputStyle = lipgloss.NewStyle().
				Foreground(styles.Text)

	NoResultsStyle = lipgloss.NewStyle().
			Foreground(styles.Warning).
			Italic(true)
)

// Detail panel styles
var (
	DetailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.Primary).
				MarginBottom(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Width(14).
				Foreground(styles.Muted)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(styles.Text)

	DetailMRIStyle = lipgloss.NewStyle().
			Foreground(styles.Secondary).
			Bold(true)

	DetailTagStyle = lipgloss.NewStyle().
			Background(styles.BgMedium).
			Foreground(styles.Text).
			Padding(0, 1).
			MarginRight(1)

	DetailSectionStyle = lipgloss.NewStyle().
				MarginTop(1).
				MarginBottom(1)
)

// Empty state styles
var (
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true).
			Align(lipgloss.Center)

	EmptyIconStyle = lipgloss.NewStyle().
			Foreground(styles.Border).
			Align(lipgloss.Center)
)

// Status indicator
func SourceIndicator(isLocal bool) string {
	if isLocal {
		return LocalSourceStyle.Render("local")
	}
	return RemoteSourceStyle.Render("remote")
}

// Tag rendering
func RenderTags(tags []string) string {
	var rendered []string
	for _, tag := range tags {
		rendered = append(rendered, DetailTagStyle.Render(tag))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, rendered...)
}
