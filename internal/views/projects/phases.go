package projects

import (
	"strings"

	"github.com/hecate-social/hecate-tui/internal/projects"
)

// PhaseBar renders the AnD/AnP/InT/DoO tab bar
type PhaseBar struct {
	active projects.Phase
	width  int
}

// NewPhaseBar creates a phase bar
func NewPhaseBar() *PhaseBar {
	return &PhaseBar{
		active: projects.PhaseAnD,
	}
}

// SetActive sets the active phase
func (p *PhaseBar) SetActive(phase projects.Phase) {
	p.active = phase
}

// Active returns the current phase
func (p *PhaseBar) Active() projects.Phase {
	return p.active
}

// SetWidth sets the bar width
func (p *PhaseBar) SetWidth(w int) {
	p.width = w
}

// Next moves to next phase
func (p *PhaseBar) Next() {
	phases := projects.Phases()
	for i, info := range phases {
		if info.Phase == p.active {
			if i < len(phases)-1 {
				p.active = phases[i+1].Phase
			}
			return
		}
	}
}

// Prev moves to previous phase
func (p *PhaseBar) Prev() {
	phases := projects.Phases()
	for i, info := range phases {
		if info.Phase == p.active {
			if i > 0 {
				p.active = phases[i-1].Phase
			}
			return
		}
	}
}

// View renders the phase bar
func (p *PhaseBar) View() string {
	var tabs []string

	for _, info := range projects.Phases() {
		label := info.Icon + " " + info.ShortName
		if info.Phase == p.active {
			tabs = append(tabs, PhaseTabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, PhaseTabStyle.Render(label))
		}
	}

	return PhaseTabBarStyle.Render(strings.Join(tabs, ""))
}

// ActiveInfo returns info about the active phase
func (p *PhaseBar) ActiveInfo() projects.PhaseInfo {
	for _, info := range projects.Phases() {
		if info.Phase == p.active {
			return info
		}
	}
	return projects.Phases()[0]
}
