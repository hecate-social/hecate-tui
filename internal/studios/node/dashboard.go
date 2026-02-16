package node

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/llm"
)

// Async message types for Bubble Tea commands.

type dashboardFetchedMsg struct {
	health       *client.Health
	identity     *client.Identity
	providers    map[string]llm.Provider
	models       []llm.Model
	capabilities []client.Capability
	agents       []client.Agent
}

type dashboardFetchErrMsg struct {
	err error
}

type switchViewMsg struct {
	view opsView
}

// opsCommand is a simple slash command for ops-specific navigation.
type opsCommand struct {
	name   string
	desc   string
	studio *Studio
}

func (c *opsCommand) Name() string        { return c.name }
func (c *opsCommand) Aliases() []string   { return nil }
func (c *opsCommand) Description() string { return c.desc }

func (c *opsCommand) Execute(_ []string, _ *commands.Context) tea.Cmd {
	switch c.name {
	case "models":
		return switchView(viewModels)
	case "providers":
		return switchView(viewProviders)
	case "caps":
		return switchView(viewCapabilities)
	case "health":
		return switchView(viewHealth)
	case "back":
		return switchView(viewDashboard)
	case "refresh":
		c.studio.loading = true
		c.studio.loadErr = nil
		return c.studio.fetchDashboard
	}
	return nil
}

func switchView(v opsView) tea.Cmd {
	return func() tea.Msg {
		return switchViewMsg{view: v}
	}
}

// fetchDashboard fetches all dashboard data in parallel.
func (s *Studio) fetchDashboard() tea.Msg {
	var (
		health       *client.Health
		identity     *client.Identity
		providers    map[string]llm.Provider
		models       []llm.Model
		capabilities []client.Capability
		agents       []client.Agent
		firstErr     error
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	setErr := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}

	wg.Add(6)

	go func() {
		defer wg.Done()
		h, err := s.ctx.Client.GetHealth()
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		health = h
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		id, err := s.ctx.Client.GetIdentity()
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		identity = id
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p, err := s.ctx.Client.ListProviders()
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		providers = p
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		m, err := s.ctx.Client.ListModels()
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		models = m
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		c, err := s.ctx.Client.DiscoverCapabilities("", "", 0)
		if err != nil {
			// Non-fatal — caps may not be available
			return
		}
		mu.Lock()
		capabilities = c
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		a, err := s.ctx.Client.ListAgents()
		if err != nil {
			// Non-fatal — agents may not be available
			return
		}
		mu.Lock()
		agents = a
		mu.Unlock()
	}()

	wg.Wait()

	// If health failed, the daemon is likely down — report the error
	if health == nil && firstErr != nil {
		return dashboardFetchErrMsg{err: firstErr}
	}

	return dashboardFetchedMsg{
		health:       health,
		identity:     identity,
		providers:    providers,
		models:       models,
		capabilities: capabilities,
		agents:       agents,
	}
}

// formatUptime converts seconds to a human-readable duration.
func formatUptime(seconds int) string {
	if seconds < 60 {
		return itoa(seconds) + "s"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	result := ""
	if days > 0 {
		result += itoa(days) + "d "
	}
	if hours > 0 || days > 0 {
		result += itoa(hours) + "h "
	}
	result += itoa(minutes) + "m"
	return result
}
