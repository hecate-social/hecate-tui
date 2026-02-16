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
	case phaseHeroes:
		return m.handleHeroesKey(key)
	case phaseHeroDetail:
		return m.handleHeroDetailKey(key)
	case phasePromote:
		return m.handlePromoteKey(key)
	case phaseHeroDuel:
		return m.handleHeroDuelKey(key)
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

	case "H":
		m.phase = phaseHeroes
		m.heroIndex = 0
		m.err = nil
		return FetchHeroes(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())
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

	case "p":
		// Cycle through presets
		presetWeights := [][7]float64{
			{0.1, 50.0, 200.0, 50.0, 100.0, 0.5, -0.2},  // balanced
			{0.1, 20.0, 400.0, 50.0, 250.0, 0.5, -0.2},   // aggressive
			{0.1, 150.0, 50.0, 50.0, 100.0, 3.0, -0.2},   // forager
			{0.8, 50.0, 200.0, 50.0, 20.0, 0.5, -1.5},    // survivor
			{0.1, 0.0, 500.0, 0.0, 300.0, 0.0, 0.0},      // assassin
		}
		m.formPreset = (m.formPreset + 1) % len(presetWeights)
		m.formWeights = presetWeights[m.formPreset]

	case "w":
		m.formShowWeights = !m.formShowWeights
		if m.formShowWeights {
			m.formWeightFocus = 0
		}

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

	case "P":
		// Promote champion to hero (completed stables with champion only)
		if m.selectedStable.Status == "completed" && m.champion != nil {
			m.phase = phasePromote
			m.promoteName = ""
			m.err = nil
		}

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

// handleHeroesKey processes keys on the heroes list view.
func (m *Model) handleHeroesKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.phase = phaseList
		m.err = nil
		return FetchStables(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())

	case "j", "down":
		if m.heroIndex < len(m.heroes)-1 {
			m.heroIndex++
		}

	case "k", "up":
		if m.heroIndex > 0 {
			m.heroIndex--
		}

	case "enter":
		if len(m.heroes) > 0 {
			hero := m.heroes[m.heroIndex]
			m.selectedHero = &hero
			m.phase = phaseHeroDetail
			m.err = nil
			return FetchHero(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL(), hero.HeroID)
		}

	case "r":
		return FetchHeroes(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())
	}

	return nil
}

// handleHeroDetailKey processes keys on the hero detail view.
func (m *Model) handleHeroDetailKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.phase = phaseHeroes
		m.err = nil
		return FetchHeroes(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())

	case "d":
		if m.selectedHero != nil {
			return StartHeroDuel(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedHero.HeroID,
				50,
				100,
			)
		}
	}

	return nil
}

// handlePromoteKey processes keys on the promote form.
func (m *Model) handlePromoteKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.phase = phaseDetail
		m.promoteName = ""
		m.err = nil
		return nil

	case "enter":
		if m.promoteName != "" {
			return PromoteChampion(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedStable.StableID,
				m.promoteName,
			)
		}

	case "backspace":
		if len(m.promoteName) > 0 {
			m.promoteName = m.promoteName[:len(m.promoteName)-1]
		}

	default:
		// Allow typing alphanumeric and basic punctuation
		if len(key) == 1 && len(m.promoteName) < 30 {
			m.promoteName += key
		}
	}

	return nil
}

// handleHeroDuelKey processes keys during a hero duel.
func (m *Model) handleHeroDuelKey(key string) tea.Cmd {
	switch key {
	case "esc":
		if m.duelStream != nil {
			m.duelStream.Close()
			m.duelStream = nil
		}
		m.phase = phaseHeroDetail
		m.err = nil
		return nil

	case "n":
		if m.duelState.Status == "finished" && m.selectedHero != nil {
			return StartHeroDuel(
				m.ctx.Client.SocketPath(),
				m.ctx.Client.BaseURL(),
				m.selectedHero.HeroID,
				50,
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
