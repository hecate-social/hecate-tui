package devops

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// handleKey processes key events in Normal mode.
func (s *Studio) handleKey(msg tea.KeyMsg) tea.Cmd {
	// Clear flash message on any keypress
	s.flashMsg = ""

	// Route to overlay-specific handlers when active
	switch s.actionMode {
	case actionViewCategories:
		return s.handleCategoryKey(msg)
	case actionViewActions:
		return s.handleActionKey(msg)
	}

	// Default: task list keys
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
	case "a":
		s.actionMode = actionViewCategories
		s.catCursor = 0
		s.actionCursor = 0
	case "enter":
		// No-op for now — will open task-specific UI later
	}
	return nil
}

// handleCategoryKey handles navigation within the category list.
func (s *Studio) handleCategoryKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		if s.catCursor < len(s.categories)-1 {
			s.catCursor++
		}
	case "k", "up":
		if s.catCursor > 0 {
			s.catCursor--
		}
	case "enter":
		if s.catCursor < len(s.categories) {
			s.actionMode = actionViewActions
			s.actionCursor = 0
		}
	case "esc":
		s.actionMode = actionViewNone
	}
	return nil
}

// handleActionKey handles navigation within an action list inside a category.
func (s *Studio) handleActionKey(msg tea.KeyMsg) tea.Cmd {
	cat := s.categories[s.catCursor]
	switch msg.String() {
	case "j", "down":
		if s.actionCursor < len(cat.Actions)-1 {
			s.actionCursor++
		}
	case "k", "up":
		if s.actionCursor > 0 {
			s.actionCursor--
		}
	case "enter":
		if s.actionCursor < len(cat.Actions) {
			return s.selectAction(cat.Actions[s.actionCursor])
		}
	case "esc":
		s.actionMode = actionViewCategories
		s.actionCursor = 0
	}
	return nil
}

// selectAction activates the chosen action — either opens its form or
// immediately executes it (for confirm-only actions with nil FormSpec).
func (s *Studio) selectAction(action Action) tea.Cmd {
	if action.FormSpec == nil {
		// Confirm-only: execute immediately
		s.activeAction = &action
		s.actionMode = actionViewNone
		return s.executeAction(action, map[string]string{})
	}

	// Build and show the form
	s.activeAction = &action
	s.formView = ui.BuildForm(*action.FormSpec, s.ctx.Theme, s.ctx.Styles)
	s.formReady = true
	s.actionMode = actionViewForm
	return s.formView.Init()
}
