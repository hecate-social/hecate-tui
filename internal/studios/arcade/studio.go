// Package arcade implements the Arcade Studio — games workspace.
package arcade

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
)

// Studio is the Arcade (games) workspace placeholder.
type Studio struct {
	ctx     *studio.Context
	width   int
	height  int
	focused bool
}

// New creates a new Arcade Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{ctx: ctx}
}

func (s *Studio) Name() string      { return "Arcade" }
func (s *Studio) ShortName() string { return "Arcade" }
func (s *Studio) Icon() string      { return "🎮" }

func (s *Studio) Init() tea.Cmd                      { return nil }
func (s *Studio) Mode() modes.Mode                   { return modes.Normal }
func (s *Studio) Hints() string                      { return "Coming soon — terminal games" }
func (s *Studio) Focused() bool                      { return s.focused }
func (s *Studio) SetFocused(focused bool)            { s.focused = focused }
func (s *Studio) SetSize(width, height int)          { s.width = width; s.height = height }
func (s *Studio) StatusInfo() studio.StatusInfo       { return studio.StatusInfo{} }
func (s *Studio) Commands() []commands.Command        { return nil }

func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	return s, nil
}

func (s *Studio) View() string {
	if s.width == 0 {
		return ""
	}

	t := s.ctx.Theme
	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Arcade Studio")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Terminal games and entertainment")

	items := []string{
		"Classic terminal games reimagined",
		"Multiplayer over the mesh network",
		"Leaderboards and achievements",
		"Plugin your own games",
	}

	var body strings.Builder
	body.WriteString(title + "\n\n")
	body.WriteString(subtitle + "\n\n")
	for _, item := range items {
		body.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Render("  • " + item) + "\n")
	}
	body.WriteString("\n")
	body.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render("  Coming soon..."))

	content := body.String()
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}
