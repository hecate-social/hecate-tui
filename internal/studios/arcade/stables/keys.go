package stables

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKey processes key events for the stables sub-app.
func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch m.phase {
	case phaseList:
		return m.handleListKey(key)
	case phaseNewStable:
		return m.handleFormKey(key)
	case phaseDetail:
		return m.handleDetailKey(key)
	case phaseDuel:
		return m.handleDuelKey(key)
	}
	return nil
}

// handleListKey processes keys on the stables list view.
func (m *Model) handleListKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.wantsBack = true
		return nil

	case "j", "down":
		if m.listIndex < len(m.stables)-1 {
			m.listIndex++
		}

	case "k", "up":
		if m.listIndex > 0 {
			m.listIndex--
		}

	case "enter":
		return m.openDetail()

	case "n":
		m.phase = phaseNewStable
		m.formFocused = 0
		m.formSeedID = ""
		m.err = nil

	case "r":
		return FetchStables(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())
	}

	return nil
}

// handleFormKey processes keys on the new stable form.
func (m *Model) handleFormKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.phase = phaseList
		m.err = nil
		return nil

	case "tab":
		m.formFocused = (m.formFocused + 1) % len(m.formFields)

	case "shift+tab":
		m.formFocused = (m.formFocused - 1 + len(m.formFields)) % len(m.formFields)

	case "+", "=":
		m.adjustFormField(1)

	case "-":
		m.adjustFormField(-1)

	case "enter":
		return m.createStable()
	}

	return nil
}

// adjustFormField adjusts the focused form field by a step.
func (m *Model) adjustFormField(dir int) {
	switch m.formFocused {
	case 0: // Population: step by 10, min 10, max 500
		m.formFields[0] = clamp(m.formFields[0]+dir*10, 10, 500)
	case 1: // Max generations: step by 10, min 10, max 1000
		m.formFields[1] = clamp(m.formFields[1]+dir*10, 10, 1000)
	case 2: // Opponent AF: step by 5, min 0, max 100
		m.formFields[2] = clamp(m.formFields[2]+dir*5, 0, 100)
	case 3: // Episodes per eval: step by 1, min 1, max 20
		m.formFields[3] = clamp(m.formFields[3]+dir*1, 1, 20)
	}
}

// handleDetailKey processes keys on the stable detail view.
func (m *Model) handleDetailKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.closeTrainingStream()
		m.phase = phaseList
		m.err = nil
		return FetchStables(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())

	case "d":
		if m.selectedStable.Status == "completed" && m.champion != nil {
			return StartChampionDuel(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedStable.StableID,
				m.selectedStable.OpponentAF,
				100,
			)
		}

	case "h":
		if m.selectedStable.Status == "training" {
			return HaltTraining(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedStable.StableID,
			)
		}

	case "s":
		// Seed a new stable from this one
		m.phase = phaseNewStable
		m.formFocused = 0
		m.formSeedID = m.selectedStable.StableID
		m.err = nil

	case "r":
		return m.refreshDetail()
	}

	return nil
}

// handleDuelKey processes keys during a champion duel.
func (m *Model) handleDuelKey(key string) tea.Cmd {
	switch key {
	case "esc":
		if m.duelStream != nil {
			m.duelStream.Close()
			m.duelStream = nil
		}
		m.phase = phaseDetail
		m.err = nil
		return m.refreshDetail()

	case "n":
		// Only allow new duel if current one is finished
		if m.duelState.Status == "finished" {
			return StartChampionDuel(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedStable.StableID,
				m.selectedStable.OpponentAF,
				100,
			)
		}
	}

	return nil
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
