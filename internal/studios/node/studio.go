// Package ops implements the DevOps Studio — node operations dashboard.
//
// Shows a single-glance dashboard with node identity, health, LLM providers,
// models, capabilities, and agents. Sub-views for models list, providers,
// capabilities, and detailed health.
package ops

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
)

// opsView identifies which sub-view is active.
type opsView int

const (
	viewDashboard opsView = iota
	viewModels
	viewProviders
	viewCapabilities
	viewHealth
)

// Studio is the DevOps workspace — node operations dashboard.
type Studio struct {
	ctx     *studio.Context
	width   int
	height  int
	focused bool

	// Sub-view state
	activeView opsView

	// Dashboard data
	health       *client.Health
	identity     *client.Identity
	providers    map[string]llm.Provider
	models       []llm.Model
	capabilities []client.Capability
	agents       []client.Agent

	// Loading/error state
	loading bool
	loadErr error

	// List navigation (for sub-views with scrollable lists)
	cursor       int
	scrollOffset int
}

// New creates a new DevOps Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx:        ctx,
		activeView: viewDashboard,
	}
}

func (s *Studio) Name() string      { return "DevOps" }
func (s *Studio) ShortName() string { return "Ops" }
func (s *Studio) Icon() string      { return "\u2699\ufe0f" }
func (s *Studio) Mode() modes.Mode  { return modes.Normal }
func (s *Studio) Focused() bool     { return s.focused }

func (s *Studio) SetFocused(focused bool) {
	s.focused = focused
	if focused && s.health == nil {
		// First focus or stale data — trigger refresh
		s.loading = true
		s.loadErr = nil
	}
}

func (s *Studio) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *Studio) Hints() string {
	if s.loading {
		return "Loading..."
	}
	if s.loadErr != nil {
		return "r:refresh"
	}
	switch s.activeView {
	case viewDashboard:
		return "r:refresh  /models  /providers  /caps  /health"
	case viewModels:
		return "j/k:navigate  r:refresh  /back"
	case viewProviders:
		return "j/k:navigate  r:refresh  /back"
	case viewCapabilities:
		return "j/k:navigate  r:refresh  /back"
	case viewHealth:
		return "r:refresh  /back"
	}
	return ""
}

func (s *Studio) StatusInfo() studio.StatusInfo {
	info := studio.StatusInfo{}
	if s.health != nil {
		info.ModelStatus = s.health.Status
	}
	if s.providers != nil {
		var enabledCount int
		for _, p := range s.providers {
			if p.Enabled {
				enabledCount++
			}
		}
		if enabledCount > 0 {
			info.ModelName = pluralize(enabledCount, "provider", "providers")
		}
	}
	return info
}

func (s *Studio) Commands() []commands.Command {
	return []commands.Command{
		&opsCommand{name: "models", desc: "List available LLM models", studio: s},
		&opsCommand{name: "providers", desc: "Manage LLM providers", studio: s},
		&opsCommand{name: "caps", desc: "List announced capabilities", studio: s},
		&opsCommand{name: "health", desc: "Detailed health check", studio: s},
		&opsCommand{name: "back", desc: "Return to dashboard", studio: s},
		&opsCommand{name: "refresh", desc: "Reload all data", studio: s},
	}
}

func (s *Studio) Init() tea.Cmd {
	s.loading = true
	return s.fetchDashboard
}

func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardFetchedMsg:
		s.loading = false
		s.loadErr = nil
		s.health = msg.health
		s.identity = msg.identity
		s.providers = msg.providers
		s.models = msg.models
		s.capabilities = msg.capabilities
		s.agents = msg.agents
		return s, nil

	case dashboardFetchErrMsg:
		s.loading = false
		s.loadErr = msg.err
		return s, nil

	case switchViewMsg:
		s.activeView = msg.view
		s.cursor = 0
		s.scrollOffset = 0
		return s, nil

	case tea.KeyMsg:
		if s.loading {
			return s, nil
		}
		return s, s.handleKey(msg)
	}

	return s, nil
}

// listLen returns the number of items in the current sub-view list.
func (s *Studio) listLen() int {
	switch s.activeView {
	case viewModels:
		return len(s.models)
	case viewProviders:
		return len(s.providers)
	case viewCapabilities:
		return len(s.capabilities)
	}
	return 0
}

// maxVisibleRows returns the number of rows available for list items.
func (s *Studio) maxVisibleRows() int {
	// Subtract header lines (breadcrumb + separator + table header + spacer = ~5)
	rows := s.height - 5
	if rows < 1 {
		return 1
	}
	return rows
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return itoa(n) + " " + plural
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
