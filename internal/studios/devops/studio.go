// Package devops implements the DevOps Studio — venture lifecycle workspace.
//
// Shows a task list derived from the daemon's venture state, with vim-style
// navigation, collapsible division groups, and task state indicators.
// Also provides command forms for venture + division lifecycle operations.
package devops

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// Async message types for Bubble Tea commands.

type tasksFetchedMsg struct {
	taskList *client.VentureTaskList
}

type tasksFetchErrMsg struct {
	err error
}

type noVentureMsg struct{}

// Studio is the Development workspace — task list view for venture lifecycle.
type Studio struct {
	ctx     *studio.Context
	width   int
	height  int
	focused bool

	// Data
	ventureID   string
	ventureName string
	taskList    *TaskList
	loading     bool
	loadErr     error
	noVenture   bool

	// Action overlay state
	actionMode   actionView
	categories   []Category
	catCursor    int
	actionCursor int
	formView     *ui.FormModel
	formReady    bool
	activeAction *Action // the action currently being submitted
	flashMsg     string  // brief result message shown after action
	flashSuccess bool    // whether the flash is a success or error
}

// New creates a new Development Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx:        ctx,
		taskList:   &TaskList{},
		categories: ventureCategories(),
	}
}

func (s *Studio) Name() string      { return "DevOps" }
func (s *Studio) ShortName() string { return "DevOps" }
func (s *Studio) Icon() string      { return "\U0001F527" }
func (s *Studio) Focused() bool     { return s.focused }

func (s *Studio) Mode() modes.Mode {
	if s.actionMode == actionViewForm && s.formReady {
		return modes.Form
	}
	return modes.Normal
}

func (s *Studio) SetFocused(focused bool) {
	s.focused = focused
	if focused {
		// Refresh data when studio gains focus
		s.loading = true
		s.loadErr = nil
	}
}

func (s *Studio) SetSize(width, height int) {
	s.width = width
	s.height = height
	// Reserve lines for header (venture name + separator = 2 lines)
	s.taskList.SetViewHeight(height - 2)
}

func (s *Studio) StatusInfo() studio.StatusInfo {
	return studio.StatusInfo{}
}

func (s *Studio) Commands() []commands.Command { return nil }

func (s *Studio) Hints() string {
	if s.actionMode == actionViewForm && s.formReady {
		return "Tab:next  Shift+Tab:prev  Enter:submit  Esc:cancel"
	}
	if s.actionMode == actionViewCategories {
		return "j/k:navigate  Enter:select  Esc:back"
	}
	if s.actionMode == actionViewActions {
		return "j/k:navigate  Enter:select  Esc:back"
	}
	if s.noVenture {
		return "a:actions  r:refresh"
	}
	if s.loading {
		return "Loading..."
	}
	if s.loadErr != nil {
		return "r:retry"
	}
	return "j/k:navigate  Tab:collapse  a:actions  r:refresh"
}

func (s *Studio) Init() tea.Cmd {
	s.loading = true
	s.categories = ventureCategories()
	return s.fetchTasks
}

func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksFetchedMsg:
		s.loading = false
		s.loadErr = nil
		s.noVenture = false
		s.ventureID = msg.taskList.Venture.ID
		s.ventureName = msg.taskList.Venture.Name
		s.taskList.BuildFromResponse(msg.taskList)
		return s, nil

	case tasksFetchErrMsg:
		s.loading = false
		s.loadErr = msg.err
		return s, nil

	case noVentureMsg:
		s.loading = false
		s.noVenture = true
		return s, nil

	case ui.FormResult:
		return s, s.handleFormResult(msg)

	case actionResultMsg:
		s.flashMsg = msg.message
		s.flashSuccess = msg.success
		s.actionMode = actionViewNone
		s.activeAction = nil
		s.formReady = false
		s.formView = nil
		// Refresh task list after action completes
		s.loading = true
		return s, s.fetchTasks

	case tea.KeyMsg:
		// When in form mode, forward all keys to the form
		if s.actionMode == actionViewForm && s.formReady && s.formView != nil {
			var cmd tea.Cmd
			s.formView, cmd = s.formView.Update(msg)
			return s, cmd
		}

		if s.loading || s.loadErr != nil {
			// Only handle 'r' for refresh when in error/loaded state
			if msg.String() == "r" && !s.loading {
				s.loading = true
				return s, s.fetchTasks
			}
			return s, nil
		}
		return s, s.handleKey(msg)
	}

	// Forward non-key messages to form when active
	if s.actionMode == actionViewForm && s.formReady && s.formView != nil {
		var cmd tea.Cmd
		s.formView, cmd = s.formView.Update(msg)
		return s, cmd
	}

	return s, nil
}

// handleFormResult processes the form result and dispatches the action.
func (s *Studio) handleFormResult(result ui.FormResult) tea.Cmd {
	s.formReady = false
	s.actionMode = actionViewNone

	if !result.Submitted {
		// User cancelled
		s.formView = nil
		s.activeAction = nil
		return nil
	}

	if s.activeAction == nil {
		s.formView = nil
		return nil
	}

	action := *s.activeAction
	values := result.Values
	s.formView = nil
	s.activeAction = nil

	return s.executeAction(action, values)
}

// fetchTasks is a tea.Cmd that fetches the active venture + task list from the daemon.
func (s *Studio) fetchTasks() tea.Msg {
	// Get active venture first
	venture, err := s.ctx.Client.GetVenture()
	if err != nil {
		return noVentureMsg{}
	}

	// Fetch task list
	taskList, err := s.ctx.Client.GetVentureTasks(venture.VentureID)
	if err != nil {
		return tasksFetchErrMsg{err: err}
	}

	return tasksFetchedMsg{taskList: taskList}
}
