package snake_duel

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKey processes key events for the snake duel sub-app.
func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "esc":
		if m.stream != nil {
			m.stream.Close()
			m.stream = nil
		}
		m.wantsBack = true
		return nil
	}

	switch m.phase {
	case "idle", "finished":
		return m.handleIdleKey(key)
	}

	// During connecting/playing, only esc works (handled above)
	return nil
}

// handleIdleKey processes keys when not in a match.
func (m *Model) handleIdleKey(key string) tea.Cmd {
	switch key {
	case "n":
		return m.startNewMatch()

	// Speed controls
	case "+", "=":
		if m.tickMs > 50 {
			m.tickMs -= 10
		}
	case "-":
		if m.tickMs < 500 {
			m.tickMs += 10
		}

	// Player 1 AF controls
	case "[":
		if m.af1 > 0 {
			m.af1 -= 5
		}
	case "]":
		if m.af1 < 100 {
			m.af1 += 5
		}

	// Player 2 AF controls
	case "{":
		if m.af2 > 0 {
			m.af2 -= 5
		}
	case "}":
		if m.af2 < 100 {
			m.af2 += 5
		}
	}

	return nil
}
