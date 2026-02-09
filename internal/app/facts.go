package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/applytorchsideeffects"
	"github.com/hecate-social/hecate-tui/internal/factbus"
)

// handleFact routes a received fact to the appropriate cartwheel handler.
// This is the TUI equivalent of a gen_server's handle_info — the App
// dispatches based on fact_type.
func (a *App) handleFact(msg factbus.FactMsg) tea.Cmd {
	switch msg.FactType {
	case "torch_initiated_v1":
		ti, err := applytorchsideeffects.HandleTorchInitiated(msg.Data)
		if err != nil {
			a.chat.InjectSystemMessage("Failed to parse torch fact: " + err.Error())
			return a.factConn.PollCmd()
		}
		a.statusBar.TorchName = ti.Name
		a.chat.InjectSystemMessage("Torch initiated: " + ti.Name + " (via fact stream)")
	}

	return a.factConn.PollCmd()
}

// scheduleFactPoll returns a command that waits briefly then polls again.
// This prevents tight-looping when no facts are available.
func (a *App) scheduleFactPoll() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return factbus.FactContinueMsg{}
	})
}
