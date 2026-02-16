package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/version"
)

// renderHome renders the home screen shown on first launch.
func (a *App) renderHome() string {
	if a.width == 0 {
		return "Loading..."
	}

	t := a.theme

	// Title
	title := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render("🔥🗝️🔥  H E C A T E  🔥🗝️🔥")

	versionLine := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("v" + version.Version)

	// Studio cards
	type card struct {
		key   string
		icon  string
		name  string
		desc  string
		color lipgloss.Color
	}

	cards := []card{
		{"1", "🤖", "LLM", "Chat with AI", t.Primary},
		{"2", "🔧", "DevOps", "Ventures", t.Secondary},
		{"3", "🌐", "Node", "Node Mgmt", t.Warning},
		{"4", "💬", "Social", "Chat IRC", t.Success},
		{"5", "🎮", "Arcade", "Games", t.Accent},
	}

	cardWidth := 15
	cardStyle := func(c card) string {
		border := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.color).
			Width(cardWidth).
			Padding(0, 1).
			Align(lipgloss.Center)

		keyLabel := lipgloss.NewStyle().Foreground(t.TextDim).Render(c.key + ".")
		icon := c.icon
		name := lipgloss.NewStyle().Foreground(c.color).Bold(true).Render(c.name)
		desc := lipgloss.NewStyle().Foreground(t.TextMuted).Render(c.desc)

		return border.Render(keyLabel + " " + icon + " " + name + "\n" + desc)
	}

	// Row 1: LLM, Dev, Ops
	var row1Cards []string
	for _, c := range cards[:3] {
		row1Cards = append(row1Cards, cardStyle(c))
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, row1Cards...)

	// Row 2: Social, Arcade
	var row2Cards []string
	for _, c := range cards[3:] {
		row2Cards = append(row2Cards, cardStyle(c))
	}
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, row2Cards...)

	// Daemon status
	daemonLine := ""
	switch a.daemonStatus {
	case "healthy", "ok":
		daemonLine = a.styles.StatusOK.Render("●") + a.styles.Subtle.Render(" daemon healthy")
	case "starting":
		daemonLine = a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon starting...")
	case "degraded":
		daemonLine = a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon degraded")
	default:
		daemonLine = a.styles.Subtle.Render("○ connecting...")
	}

	// Hint
	hint := lipgloss.NewStyle().Foreground(t.TextMuted).Render("Press 1-5 to enter a studio  •  q to quit")

	// Assemble
	var content strings.Builder
	content.WriteString(title + "\n")
	content.WriteString(versionLine + "\n\n")
	content.WriteString(row1 + "\n")
	content.WriteString(row2 + "\n\n")
	content.WriteString(daemonLine + "\n\n")
	content.WriteString(hint)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content.String())
}
