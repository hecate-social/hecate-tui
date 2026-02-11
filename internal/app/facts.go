package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/applyventuresideeffects"
	"github.com/hecate-social/hecate-tui/internal/factbus"
)

// handleFact routes a received fact to the appropriate handler.
func (a *App) handleFact(msg factbus.FactMsg) tea.Cmd {
	switch msg.FactType {
	case "venture_setup_v1":
		ti, err := applyventuresideeffects.HandleVentureInitiated(msg.Data)
		if err != nil {
			if llm := a.llmStudio(); llm != nil {
				llm.InjectSystemMessage("Failed to parse venture fact: " + err.Error())
			}
			return a.factConn.PollCmd()
		}
		a.statusBar.VentureName = ti.Name
		if llm := a.llmStudio(); llm != nil {
			llm.InjectSystemMessage("Venture initiated: " + ti.Name + " (via fact stream)")
		}
	}

	return a.factConn.PollCmd()
}

// scheduleFactPoll returns a command that waits briefly then polls again.
func (a *App) scheduleFactPoll() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return factbus.FactContinueMsg{}
	})
}
