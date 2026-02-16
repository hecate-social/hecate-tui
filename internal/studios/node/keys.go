package node

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// handleKey processes key events in Normal mode.
func (s *Studio) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Form mode: forward all keys to the form
	if s.actionMode == actionViewForm && s.formReady && s.formView != nil {
		return s.handleFormKey(msg)
	}

	// Action category navigation
	if s.actionMode == actionViewCategories {
		return s.handleCategoryKey(key)
	}

	// Action selection navigation
	if s.actionMode == actionViewActions {
		return s.handleActionKey(key)
	}

	// Global keys available in all sub-views
	switch key {
	case "r":
		s.loading = true
		s.loadErr = nil
		return s.fetchDashboard
	case "a":
		s.actionMode = actionViewCategories
		s.catCursor = 0
		return nil
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

// handleCategoryKey handles navigation within the category list.
func (s *Studio) handleCategoryKey(key string) tea.Cmd {
	total := len(s.categories)
	if total == 0 {
		return nil
	}

	switch key {
	case "j", "down":
		if s.catCursor < total-1 {
			s.catCursor++
		}
	case "k", "up":
		if s.catCursor > 0 {
			s.catCursor--
		}
	case "enter":
		// Select category, move to action list
		s.actionMode = actionViewActions
		s.actionCursor = 0
		return nil
	case "esc":
		s.actionMode = actionViewNone
		return nil
	}

	return nil
}

// handleActionKey handles navigation within the action list for the selected category.
func (s *Studio) handleActionKey(key string) tea.Cmd {
	if s.catCursor >= len(s.categories) {
		s.actionMode = actionViewCategories
		return nil
	}

	cat := s.categories[s.catCursor]
	total := len(cat.Actions)
	if total == 0 {
		return nil
	}

	switch key {
	case "j", "down":
		if s.actionCursor < total-1 {
			s.actionCursor++
		}
	case "k", "up":
		if s.actionCursor > 0 {
			s.actionCursor--
		}
	case "enter":
		if s.actionCursor >= total {
			return nil
		}
		action := cat.Actions[s.actionCursor]
		s.activeAction = &action

		// No form needed — execute immediately (confirm-only action)
		if action.FormSpec == nil {
			s.actionMode = actionViewNone
			return s.executeAction(action, nil)
		}

		// Build the form and enter form mode
		spec := *action.FormSpec
		s.formView = ui.BuildForm(spec, s.ctx.Theme, s.ctx.Styles)
		s.formReady = true
		s.actionMode = actionViewForm
		return s.formView.Init()

	case "esc":
		// Go back to categories
		s.actionMode = actionViewCategories
		return nil
	}

	return nil
}

// handleFormKey forwards key messages to the active form.
func (s *Studio) handleFormKey(msg tea.KeyMsg) tea.Cmd {
	if s.formView == nil {
		return nil
	}

	updated, cmd := s.formView.Update(msg)
	s.formView = updated
	return cmd
}
