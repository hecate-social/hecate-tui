package dev

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes key events in Normal mode.
func (s *Studio) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		s.taskList.Down()
	case "k", "up":
		s.taskList.Up()
	case "g", "home":
		s.taskList.Top()
	case "G", "end":
		s.taskList.Bottom()
	case "tab":
		s.taskList.ToggleCollapse()
	case "r":
		s.loading = true
		return s.fetchTasks
	case "enter":
		// No-op for now — will open task-specific UI later
	}
	return nil
}
