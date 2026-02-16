// Package social implements the Social Studio — IRC chat over Macula Mesh.
//
// Follows the hecate-web UX model: lands on a Studio Home page (app explorer)
// with sub-app cards. When IRC is selected, switches to the full IRC sub-app view.
// Esc returns to home.
package social

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
)

// socialApp describes a sub-app card on the home screen.
type socialApp struct {
	id          string
	name        string
	icon        string
	description string
	active      bool
}

// Studio is the Social workspace — IRC chat over the mesh.
type Studio struct {
	ctx     *studio.Context
	width   int
	height  int
	focused bool

	// Home view state
	activeApp string // "" = home, "irc" = IRC sub-app
	appIndex  int    // selected card on home screen
	apps      []socialApp

	// IRC sub-app (nil until opened)
	irc *ircModel
}

// New creates a new Social Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx: ctx,
		apps: []socialApp{
			{id: "irc", name: "IRC", icon: "#", description: "Chat channels over the mesh", active: true},
			{id: "forum", name: "Forum", icon: "\U0001F4AC", description: "Threaded discussions", active: false},
			{id: "feed", name: "Feed", icon: "\U0001F4E1", description: "Activity feed / timeline", active: false},
			{id: "news", name: "News", icon: "\U0001F4F0", description: "News aggregator", active: false},
			{id: "conferencing", name: "Conferencing", icon: "\U0001F4F9", description: "Voice / video calls", active: false},
		},
	}
}

func (s *Studio) Name() string      { return "Social" }
func (s *Studio) ShortName() string { return "Social" }
func (s *Studio) Icon() string      { return "\U0001F4AC" }
func (s *Studio) Focused() bool     { return s.focused }

func (s *Studio) SetFocused(focused bool) {
	s.focused = focused
}

func (s *Studio) SetSize(width, height int) {
	s.width = width
	s.height = height
	if s.irc != nil {
		s.irc.setSize(width, height)
	}
}

func (s *Studio) Mode() modes.Mode {
	if s.activeApp == "irc" && s.irc != nil {
		return s.irc.mode
	}
	return modes.Normal
}

func (s *Studio) Hints() string {
	if s.activeApp == "irc" && s.irc != nil {
		return s.irc.hints()
	}
	return "\u2191\u2193\u2190\u2192:navigate  Enter:open"
}

func (s *Studio) StatusInfo() studio.StatusInfo {
	if s.activeApp == "irc" && s.irc != nil {
		return s.irc.statusInfo()
	}
	return studio.StatusInfo{}
}

func (s *Studio) Commands() []commands.Command { return nil }

func (s *Studio) Init() tea.Cmd {
	return nil
}

func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	// IRC sub-app active: delegate everything
	if s.activeApp == "irc" && s.irc != nil {
		cmd := s.irc.update(msg)

		// Check if IRC requested going back to home
		if s.irc.wantsBack {
			s.irc.wantsBack = false
			s.closeIrc()
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

// openIrc launches the IRC sub-app.
func (s *Studio) openIrc() tea.Cmd {
	s.activeApp = "irc"
	s.irc = newIrcModel(s.ctx)
	s.irc.setSize(s.width, s.height)
	return s.irc.init()
}

// closeIrc returns to the home screen.
func (s *Studio) closeIrc() {
	if s.irc != nil {
		s.irc.close()
		s.irc = nil
	}
	s.activeApp = ""
}
