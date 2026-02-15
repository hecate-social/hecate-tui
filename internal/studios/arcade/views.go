package arcade

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// viewHome renders the app explorer grid for the Arcade Studio.
func (s *Studio) viewHome() string {
	t := s.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("\U0001F3AE Arcade Studio")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Terminal games and entertainment")

	cols := 2
	gap := 2
	cardWidth := 28

	var rows []string
	for i := 0; i < len(s.apps); i += cols {
		var rowCards []string
		for j := 0; j < cols && i+j < len(s.apps); j++ {
			idx := i + j
			selected := idx == s.appIndex
			rowCards = append(rowCards, renderAppCard(t, s.apps[idx], selected, cardWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			strings.Join(rowCards, strings.Repeat(" ", gap))))
	}

	grid := strings.Join(rows, "\n")

	hints := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("\u2191\u2193\u2190\u2192:navigate  Enter:open")

	content := title + "\n" + subtitle + "\n\n" + grid + "\n\n" + hints
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

// renderAppCard renders a single app card for the home screen.
func renderAppCard(t *theme.Theme, app arcadeApp, selected bool, width int) string {
	borderColor := t.Border
	if selected && app.active {
		borderColor = t.Primary
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	iconStyle := lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(t.TextDim)

	if !app.active {
		iconStyle = iconStyle.Foreground(t.TextMuted)
		nameStyle = nameStyle.Foreground(t.TextMuted)
		descStyle = descStyle.Foreground(t.TextMuted)
	}

	var content strings.Builder
	content.WriteString(iconStyle.Render(app.icon) + " " + nameStyle.Render(app.name) + "\n")
	content.WriteString(descStyle.Render(app.description))

	if !app.active {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render("Coming Soon"))
	}

	return cardStyle.Render(content.String())
}
