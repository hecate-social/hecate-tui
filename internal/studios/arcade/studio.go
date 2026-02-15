// Package arcade implements the Arcade Studio — games workspace.
//
// Follows the hecate-web UX model: lands on a Studio Home page (app explorer)
// with sub-app cards. When a game is selected, switches to the game sub-app view.
// Esc returns to home.
package arcade

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/studios/arcade/snake_duel"
)

// arcadeApp describes a sub-app card on the home screen.
type arcadeApp struct {
	id          string
	name        string
	icon        string
	description string
	active      bool
}

// Studio is the Arcade workspace — terminal games.
type Studio struct {
	ctx     *studio.Context
	width   int
	height  int
	focused bool

	// Home view state
	activeApp string // "" = home, "snake_duel" = Snake Duel sub-app
	appIndex  int    // selected card on home screen
	apps      []arcadeApp

	// Snake Duel sub-app (nil until opened)
	snakeDuel *snake_duel.Model
}

// New creates a new Arcade Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx: ctx,
		apps: []arcadeApp{
			{id: "snake_duel", name: "Snake Duel", icon: "\U0001F40D", description: "Two AI snakes battle it out", active: true},
			{id: "tetris", name: "Tetris", icon: "\U0001F9E9", description: "Classic block stacking", active: false},
			{id: "pong", name: "Pong", icon: "\U0001F3D3", description: "Retro table tennis", active: false},
			{id: "life", name: "Conway's Life", icon: "\U0001F9EC", description: "Cellular automaton", active: false},
		},
	}
}

func (s *Studio) Name() string      { return "Arcade" }
func (s *Studio) ShortName() string { return "Arcade" }
func (s *Studio) Icon() string      { return "\U0001F3AE" }
func (s *Studio) Focused() bool     { return s.focused }

func (s *Studio) SetFocused(focused bool) {
	s.focused = focused
}

func (s *Studio) SetSize(width, height int) {
	s.width = width
	s.height = height
	if s.snakeDuel != nil {
		s.snakeDuel.SetSize(width, height)
	}
}

func (s *Studio) Mode() modes.Mode {
	return modes.Normal
}

func (s *Studio) Hints() string {
	if s.activeApp == "snake_duel" && s.snakeDuel != nil {
		return s.snakeDuel.Hints()
	}
	return "\u2191\u2193\u2190\u2192:navigate  Enter:open"
}

func (s *Studio) StatusInfo() studio.StatusInfo {
	if s.activeApp == "snake_duel" && s.snakeDuel != nil {
		return s.snakeDuel.StatusInfo()
	}
	return studio.StatusInfo{}
}

func (s *Studio) Commands() []commands.Command { return nil }

func (s *Studio) Init() tea.Cmd {
	return nil
}

func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	// Snake Duel sub-app active: delegate everything
	if s.activeApp == "snake_duel" && s.snakeDuel != nil {
		cmd := s.snakeDuel.Update(msg)

		// Check if game requested going back to home
		if s.snakeDuel.WantsBack() {
			s.snakeDuel.ClearWantsBack()
			s.closeSnakeDuel()
			return s, nil
		}
		return s, cmd
	}

	// Home view
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return s, s.handleHomeKey(msg)
	}

	return s, nil
}

func (s *Studio) View() string {
	if s.width == 0 {
		return ""
	}

	if s.activeApp == "snake_duel" && s.snakeDuel != nil {
		return s.snakeDuel.View()
	}

	return s.viewHome()
}

// openSnakeDuel launches the Snake Duel sub-app.
func (s *Studio) openSnakeDuel() tea.Cmd {
	s.activeApp = "snake_duel"
	s.snakeDuel = snake_duel.New(s.ctx)
	s.snakeDuel.SetSize(s.width, s.height)
	return s.snakeDuel.Init()
}

// closeSnakeDuel returns to the home screen.
func (s *Studio) closeSnakeDuel() {
	if s.snakeDuel != nil {
		s.snakeDuel.Close()
		s.snakeDuel = nil
	}
	s.activeApp = ""
}
