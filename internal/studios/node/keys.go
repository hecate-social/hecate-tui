package ops

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes key events in Normal mode.
func (s *Studio) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Global keys available in all sub-views
	switch key {
	case "r":
		s.loading = true
		s.loadErr = nil
		return s.fetchDashboard
	}

	// Sub-view specific keys
	switch s.activeView {
	case viewDashboard:
		return s.handleDashboardKey(key)
	case viewModels, viewProviders, viewCapabilities:
		return s.handleListKey(key)
	case viewHealth:
		// Health view has no special keys beyond 'r' (handled above)
		return nil
	}

	return nil
}

// handleDashboardKey handles keys on the dashboard view.
func (s *Studio) handleDashboardKey(_ string) tea.Cmd {
	// Dashboard is a static overview — no navigation keys needed.
	// Navigation happens via slash commands (/models, /providers, etc.)
	return nil
}

// handleListKey handles j/k navigation for list sub-views.
func (s *Studio) handleListKey(key string) tea.Cmd {
	total := s.listLen()
	if total == 0 {
		return nil
	}

	switch key {
	case "j", "down":
		if s.cursor < total-1 {
			s.cursor++
			s.ensureVisible()
		}
	case "k", "up":
		if s.cursor > 0 {
			s.cursor--
			s.ensureVisible()
		}
	case "g", "home":
		s.cursor = 0
		s.scrollOffset = 0
	case "G", "end":
		s.cursor = total - 1
		s.ensureVisible()
	}

	return nil
}

// ensureVisible adjusts scrollOffset so cursor is within the visible window.
func (s *Studio) ensureVisible() {
	maxRows := s.maxVisibleRows()

	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	if s.cursor >= s.scrollOffset+maxRows {
		s.scrollOffset = s.cursor - maxRows + 1
	}
}
