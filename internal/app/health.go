package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/alc"
)

// healthMsg carries daemon health check results.
type healthMsg struct {
	status string
}

// healthTickMsg triggers periodic health polling.
type healthTickMsg struct{}

// torchDetectedMsg carries auto-detected torch info.
type torchDetectedMsg struct {
	torch  *alc.TorchInfo
	source string
}

// detectTorch attempts to auto-detect a torch from git remote or .hecate/torch.json.
func (a *App) detectTorch() tea.Msg {
	result := alc.DetectTorch()
	if !result.Found {
		return nil
	}

	// If detected from config, we have the torch ID directly
	if result.Source == "config" && result.Config != nil && result.Config.TorchID != "" {
		// Try to fetch full torch info from daemon
		torch, err := a.client.GetTorchByID(result.Config.TorchID)
		if err == nil && torch != nil {
			return torchDetectedMsg{
				torch: &alc.TorchInfo{
					ID:    torch.TorchID,
					Name:  torch.Name,
					Brief: torch.Brief,
				},
				source: "config",
			}
		}
	}

	// If detected from git, try to match against known torches
	if result.Source == "git" && result.Config != nil {
		torches, err := a.client.ListTorches()
		if err == nil {
			// Match by name (the normalized git URL is stored in Name)
			for _, t := range torches {
				// TODO: Add git_remote field to torch for proper matching
				// For now, just check if torch name is part of the git URL
				if strings.Contains(result.Config.Name, t.Name) {
					return torchDetectedMsg{
						torch: &alc.TorchInfo{
							ID:    t.TorchID,
							Name:  t.Name,
							Brief: t.Brief,
						},
						source: "git",
					}
				}
			}
		}
	}

	return nil
}

func (a *App) scheduleHealthTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return healthTickMsg{}
	})
}

func (a *App) checkHealth() tea.Msg {
	health, err := a.client.GetHealth()
	if err != nil {
		return healthMsg{status: "error"}
	}
	return healthMsg{status: health.Status}
}
