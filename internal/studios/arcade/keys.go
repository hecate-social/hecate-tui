package arcade

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleHomeKey processes key events on the Arcade Studio Home screen.
func (s *Studio) handleHomeKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	cols := 2
	totalApps := len(s.apps)

	switch key {
	case "enter":
		if s.appIndex < totalApps && s.apps[s.appIndex].active {
			switch s.apps[s.appIndex].id {
			case "snake_duel":
				return s.openSnakeDuel()
			}
		}
		return nil

	case "j", "down":
		if s.appIndex+cols < totalApps {
			s.appIndex += cols
		}
	case "k", "up":
		if s.appIndex-cols >= 0 {
			s.appIndex -= cols
		}
	case "l", "right":
		if s.appIndex+1 < totalApps && (s.appIndex+1)%cols != 0 {
			s.appIndex++
		}
	case "h", "left":
		if s.appIndex > 0 && s.appIndex%cols != 0 {
			s.appIndex--
		}
	}

	return nil
}
