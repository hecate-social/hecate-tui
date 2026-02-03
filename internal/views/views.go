package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

// View is the interface that all views must implement
type View interface {
	tea.Model

	// Name returns the tab label for this view
	Name() string

	// ShortHelp returns the help text for the status bar
	ShortHelp() string

	// SetSize updates the view dimensions
	SetSize(width, height int)

	// Focus is called when the view becomes active
	Focus()

	// Blur is called when the view becomes inactive
	Blur()
}

// Tab represents a navigation tab
type Tab int

const (
	TabChat Tab = iota
	TabBrowse
	TabProjects
	TabMonitor
	TabPair
	TabMe
)

// AllTabs returns all tabs in order
func AllTabs() []Tab {
	return []Tab{TabChat, TabBrowse, TabProjects, TabMonitor, TabPair, TabMe}
}

func (t Tab) String() string {
	return [...]string{"Chat", "Browse", "Projects", "Monitor", "Pair", "Me"}[t]
}

// Shortcut returns the keyboard shortcut for this tab
func (t Tab) Shortcut() string {
	return [...]string{"1", "2", "3", "4", "5", "6"}[t]
}
