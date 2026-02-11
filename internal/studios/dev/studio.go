// Package dev implements the Development Studio — venture lifecycle workspace.
//
// Shows a task list derived from the daemon's venture state, with vim-style
// navigation, collapsible division groups, and task state indicators.
package dev

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
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
}

// New creates a new Development Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx:      ctx,
		taskList: &TaskList{},
	}
}

func (s *Studio) Name() string      { return "Development" }
func (s *Studio) ShortName() string { return "Dev" }
func (s *Studio) Icon() string      { return "\U0001F527" }
func (s *Studio) Mode() modes.Mode  { return modes.Normal }
func (s *Studio) Focused() bool     { return s.focused }

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
	if s.noVenture {
		return "No active venture — use /venture init in LLM Studio"
	}
	if s.loading {
		return "Loading..."
	}
	if s.loadErr != nil {
		return "r:retry"
	}
	return "j/k:navigate  Tab:collapse  r:refresh"
}

func (s *Studio) Init() tea.Cmd {
	s.loading = true
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

	case tea.KeyMsg:
		if s.loading || s.noVenture || s.loadErr != nil {
			// Only handle 'r' for refresh when in error/loaded state
			if msg.String() == "r" && !s.loading {
				s.loading = true
				return s, s.fetchTasks
			}
			return s, nil
		}
		return s, s.handleKey(msg)
	}

	return s, nil
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
