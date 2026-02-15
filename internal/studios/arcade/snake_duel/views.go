package snake_duel

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// view renders the full snake duel interface.
func (m *Model) view() string {
	if m.width == 0 {
		return ""
	}

	switch m.phase {
	case "idle":
		return m.viewIdle()
	case "connecting":
		return m.viewConnecting()
	case "playing":
		return m.viewPlaying()
	case "finished":
		return m.viewFinished()
	default:
		return m.viewIdle()
	}
}

// viewIdle shows instructions and settings before a match.
func (m *Model) viewIdle() string {
	t := m.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Snake Duel")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Two AI snakes battle it out")

	settings := m.renderSettings(t)

	hint := lipgloss.NewStyle().
		Foreground(colorSnake1Head).Bold(true).
		Render("Press n to start a match")

	controls := m.renderControls(t)

	content := title + "\n" + subtitle + "\n\n" +
		settings + "\n\n" +
		hint + "\n\n" +
		controls

	if m.err != nil {
		errMsg := lipgloss.NewStyle().
			Foreground(colorSnake2Head).
			Render("Error: " + m.err.Error())
		content += "\n\n" + errMsg
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewConnecting shows a loading state.
func (m *Model) viewConnecting() string {
	t := m.ctx.Theme
	msg := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Starting match...")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

// viewPlaying renders the active game: header + grid + event log + controls.
func (m *Model) viewPlaying() string {
	t := m.ctx.Theme

	header := m.renderHeader(t)
	grid := RenderGrid(m.state)

	// Check for countdown overlay
	if m.state.Status == "countdown" && m.state.Countdown > 0 {
		grid = m.overlayCountdown(grid, t)
	}

	events := m.renderEventLog(t)
	controls := m.renderPlayingControls(t)

	// Compose vertically
	return header + "\n" + grid + "\n" + events + "\n" + controls
}

// viewFinished shows the game over screen with final scores.
func (m *Model) viewFinished() string {
	t := m.ctx.Theme

	header := m.renderHeader(t)
	grid := RenderGrid(m.state)

	// Winner overlay on top of grid
	overlay := m.renderWinnerOverlay(t)
	grid = overlayCenter(grid, overlay, m.width)

	events := m.renderEventLog(t)

	hint := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("n:new match  esc:back  +/-:speed  [/]:P1 AF  {/}:P2 AF")

	return header + "\n" + grid + "\n" + events + "\n" + hint
}

// renderHeader shows player info and scores.
func (m *Model) renderHeader(t *theme.Theme) string {
	back := lipgloss.NewStyle().Foreground(t.TextMuted).Render("<- Esc")
	sep := lipgloss.NewStyle().Foreground(t.Border).Render(" | ")
	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Snake Duel")

	p1Score := lipgloss.NewStyle().Foreground(colorSnake1Head).Bold(true).
		Render("P1:" + itoa(m.state.Snake1.Score))
	p1AF := lipgloss.NewStyle().Foreground(colorSnake1Body).
		Render("AF" + itoa(m.state.Snake1.AssFactor))

	p2Score := lipgloss.NewStyle().Foreground(colorSnake2Head).Bold(true).
		Render("P2:" + itoa(m.state.Snake2.Score))
	p2AF := lipgloss.NewStyle().Foreground(colorSnake2Body).
		Render("AF" + itoa(m.state.Snake2.AssFactor))

	tick := lipgloss.NewStyle().Foreground(t.TextMuted).
		Render("T" + itoa(m.state.Tick))

	left := back + sep + title + sep + p1Score + " " + p1AF + sep + p2Score + " " + p2AF
	right := tick

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}

	return left + strings.Repeat(" ", gap) + right
}

// renderEventLog shows recent events from both snakes.
func (m *Model) renderEventLog(t *theme.Theme) string {
	maxEvents := 3

	var lines []string

	// Collect events from both snakes, most recent first
	s1Events := m.state.Snake1.Events
	s2Events := m.state.Snake2.Events

	// Show last N events from each snake
	for i := len(s1Events) - 1; i >= 0 && len(lines) < maxEvents; i-- {
		evt := s1Events[i]
		icon := eventIcon(evt.Type)
		line := lipgloss.NewStyle().Foreground(colorSnake1Head).
			Render(icon + " Blue: " + evt.Value)
		lines = append(lines, line)
	}

	for i := len(s2Events) - 1; i >= 0 && len(lines) < maxEvents*2; i-- {
		evt := s2Events[i]
		icon := eventIcon(evt.Type)
		line := lipgloss.NewStyle().Foreground(colorSnake2Head).
			Render(icon + " Red: " + evt.Value)
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render("  Waiting for events...")
	}

	// Limit total displayed
	if len(lines) > maxEvents*2 {
		lines = lines[:maxEvents*2]
	}

	return strings.Join(lines, "\n")
}

// renderSettings shows the current match configuration.
func (m *Model) renderSettings(t *theme.Theme) string {
	labelStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text).Bold(true)

	speed := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
		Render(itoa(m.tickMs) + "ms")
	p1af := lipgloss.NewStyle().Foreground(colorSnake1Head).Bold(true).
		Render(itoa(m.af1))
	p2af := lipgloss.NewStyle().Foreground(colorSnake2Head).Bold(true).
		Render(itoa(m.af2))

	_ = valueStyle // suppress unused

	return labelStyle.Render("Speed: ") + speed +
		labelStyle.Render("  P1 AF: ") + p1af +
		labelStyle.Render("  P2 AF: ") + p2af
}

// renderControls shows keybinding help for idle/finished phase.
func (m *Model) renderControls(t *theme.Theme) string {
	return lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("n:new  esc:back  +/-:speed  [/]:P1 AF  {/}:P2 AF")
}

// renderPlayingControls shows minimal controls during gameplay.
func (m *Model) renderPlayingControls(t *theme.Theme) string {
	return lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
		Render("esc:stop & back")
}

// overlayCountdown renders a large countdown number.
func (m *Model) overlayCountdown(grid string, t *theme.Theme) string {
	num := itoa(m.state.Countdown)
	if m.state.Countdown == 0 {
		num = "GO!"
	}

	overlay := lipgloss.NewStyle().
		Foreground(colorFood).
		Bold(true).
		Render(num)

	return overlayCenter(grid, overlay, m.width)
}

// renderWinnerOverlay shows the game result.
func (m *Model) renderWinnerOverlay(t *theme.Theme) string {
	var winnerText string
	var winnerColor lipgloss.Color

	switch m.state.Winner {
	case "player1":
		winnerText = "Blue Wins!"
		winnerColor = colorSnake1Head
	case "player2":
		winnerText = "Red Wins!"
		winnerColor = colorSnake2Head
	case "draw":
		winnerText = "Draw!"
		winnerColor = colorFood
	default:
		winnerText = "Game Over"
		winnerColor = t.TextDim
	}

	title := lipgloss.NewStyle().
		Foreground(winnerColor).Bold(true).
		Render(winnerText)

	scores := lipgloss.NewStyle().Foreground(colorSnake1Head).Render("Blue: "+itoa(m.state.Snake1.Score)) +
		"  " +
		lipgloss.NewStyle().Foreground(colorSnake2Head).Render("Red: "+itoa(m.state.Snake2.Score))

	return title + "\n" + scores
}

// overlayCenter places text in the center of an existing rendered block.
func overlayCenter(background, overlay string, width int) string {
	bgLines := strings.Split(background, "\n")
	ovLines := strings.Split(overlay, "\n")

	midRow := len(bgLines) / 2
	startRow := midRow - len(ovLines)/2

	for i, ovLine := range ovLines {
		row := startRow + i
		if row >= 0 && row < len(bgLines) {
			ovWidth := lipgloss.Width(ovLine)
			bgWidth := lipgloss.Width(bgLines[row])
			padLeft := (bgWidth - ovWidth) / 2
			if padLeft < 0 {
				padLeft = 0
			}
			bgLines[row] = strings.Repeat(" ", padLeft) + ovLine
		}
	}

	return strings.Join(bgLines, "\n")
}

// eventIcon returns a small icon for each event type.
func eventIcon(eventType string) string {
	switch eventType {
	case "food":
		return "*"
	case "poison_drop":
		return "!"
	case "poison_eat":
		return "x"
	case "collision":
		return "#"
	case "win":
		return ">"
	case "turn":
		return "~"
	default:
		return "-"
	}
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
