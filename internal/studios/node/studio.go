// Package node implements the Node Studio — node operations dashboard.
//
// Shows a single-glance dashboard with node identity, health, LLM providers,
// models, capabilities, and agents. Sub-views for models list, providers,
// capabilities, and detailed health.
// Also provides command forms for node lifecycle operations (identity,
// capabilities, mesh, subscriptions, security).
package node

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/ui"
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

	// Action overlay state
	actionMode   actionView
	categories   []Category
	catCursor    int
	actionCursor int
	formView     *ui.FormModel
	formReady    bool
	activeAction *Action

	// Flash message after action completes
	flashMsg     string
	flashSuccess bool
}

// New creates a new DevOps Studio.
func New(ctx *studio.Context) *Studio {
	return &Studio{
		ctx:        ctx,
		activeView: viewDashboard,
		categories: nodeCategories(),
	}
}

func (s *Studio) Name() string      { return "Node" }
func (s *Studio) ShortName() string { return "Node" }
func (s *Studio) Icon() string      { return "\U0001F310" }
func (s *Studio) Mode() modes.Mode {
	if s.actionMode == actionViewForm && s.formReady {
		return modes.Form
	}
	return modes.Normal
}
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
	if s.flashMsg != "" {
		return s.flashMsg
	}

	// Action overlay hints
	if s.actionMode == actionViewForm && s.formReady {
		return "Tab:next  Shift+Tab:prev  Enter:submit  Esc:cancel"
	}
	if s.actionMode == actionViewCategories {
		return "j/k:navigate  Enter:select  Esc:back"
	}
	if s.actionMode == actionViewActions {
		return "j/k:navigate  Enter:select  Esc:back"
	}

	if s.loading {
		return "Loading..."
	}
	if s.loadErr != nil {
		return "r:refresh"
	}
	switch s.activeView {
	case viewDashboard:
		return "a:actions  r:refresh  /models  /providers  /caps  /health"
	case viewModels:
		return "a:actions  j/k:navigate  r:refresh  /back"
	case viewProviders:
		return "a:actions  j/k:navigate  r:refresh  /back"
	case viewCapabilities:
		return "a:actions  j/k:navigate  r:refresh  /back"
	case viewHealth:
		return "a:actions  r:refresh  /back"
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

	case ui.FormResult:
		return s, s.handleFormResult(msg)

	case actionResultMsg:
		s.actionMode = actionViewNone
		s.formView = nil
		s.formReady = false
		s.activeAction = nil
		s.flashSuccess = msg.success
		s.flashMsg = msg.message
		return s, s.clearFlashAfterDelay()

	case clearFlashMsg:
		s.flashMsg = ""
		return s, nil

	case tea.KeyMsg:
		if s.loading && s.actionMode == actionViewNone {
			return s, nil
		}
		// Any key press clears flash message
		if s.flashMsg != "" {
			s.flashMsg = ""
		}
		return s, s.handleKey(msg)
	}

	// Forward messages to form when active
	if s.actionMode == actionViewForm && s.formReady && s.formView != nil {
		updated, cmd := s.formView.Update(msg)
		s.formView = updated
		return s, cmd
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

// handleFormResult processes a completed form (submit or cancel).
func (s *Studio) handleFormResult(result ui.FormResult) tea.Cmd {
	if !result.Submitted || s.activeAction == nil {
		// Cancelled — go back to action list
		s.actionMode = actionViewActions
		s.formView = nil
		s.formReady = false
		return nil
	}

	// Submit — execute the action
	action := *s.activeAction
	values := result.Values
	s.actionMode = actionViewNone
	s.formView = nil
	s.formReady = false
	s.activeAction = nil
	return s.executeAction(action, values)
}

// clearFlashMsg is a message type to clear flash text after a delay.
type clearFlashMsg struct{}

// clearFlashAfterDelay returns a Cmd that fires clearFlashMsg after 3 seconds.
func (s *Studio) clearFlashAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}
