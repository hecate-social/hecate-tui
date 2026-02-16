package stables

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/studios/arcade/snake_duel"
)

// phase describes which view the sub-app is showing.
const (
	phaseList       = "list"
	phaseNewStable  = "new_stable"
	phaseDetail     = "detail"
	phaseDuel       = "duel"
	phaseHeroes     = "heroes"
	phasePromote    = "promote"
	phaseHeroDetail = "hero_detail"
	phaseHeroDuel   = "hero_duel"
)

// Model is the Bubble Tea model for the Stables sub-app.
type Model struct {
	ctx *studio.Context

	// Layout
	width  int
	height int

	// Phase management
	phase string

	// List view state
	stables    []Stable
	listIndex  int
	listLoaded bool

	// New stable form state
	formFields    [4]int    // population, maxGens, opponentAF, episodesPerEval
	formLabels    [4]string // labels for each field
	formFocused   int       // which field has focus
	formSeedID    string    // optional seed stable ID

	// Detail view state
	selectedStable Stable
	champion       *Champion
	generations    []GenerationStats
	trainingStream *TrainingStream
	lastProgress   *TrainingProgress

	// Duel view state (reuses snake_duel)
	duelMatchID string
	duelStream  *snake_duel.MatchStream
	duelState   snake_duel.GameState

	// Fitness weight form state
	formPreset      int        // index into preset names (0=balanced)
	formShowWeights bool       // toggle advanced section
	formWeights     [7]float64 // survival, food, win, draw, kill, proximity, circle
	formWeightNames [7]string
	formWeightFocus int // which weight field has focus in advanced mode

	// Hero state
	heroes       []Hero
	heroIndex    int
	selectedHero *Hero
	promoteName  string // text input for hero name

	// Navigation
	wantsBack bool

	// Error from last operation
	err error
}

// New creates a new Stables model.
func New(ctx *studio.Context) *Model {
	return &Model{
		ctx:   ctx,
		phase: phaseList,
		formFields: [4]int{50, 100, 50, 3},
		formLabels:      [4]string{"Population", "Max Generations", "Opponent AF", "Episodes/Eval"},
		formWeights:     [7]float64{0.1, 50.0, 200.0, 50.0, 100.0, 0.5, -0.2},
		formWeightNames: [7]string{"Survival", "Food", "Win Bonus", "Draw Bonus", "Kill Bonus", "Proximity", "Circle Penalty"},
	}
}

// Init returns the initial command — fetch stables list.
func (m *Model) Init() tea.Cmd {
	return FetchStables(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())
}

// SetSize updates the layout dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// WantsBack returns true if the model wants to return to the arcade home.
func (m *Model) WantsBack() bool {
	return m.wantsBack
}

// ClearWantsBack resets the back signal.
func (m *Model) ClearWantsBack() {
	m.wantsBack = false
}

// View renders the current phase.
func (m *Model) View() string {
	return m.view()
}

// Hints returns contextual keybinding hints.
func (m *Model) Hints() string {
	switch m.phase {
	case phaseList:
		return "j/k:navigate  Enter:open  n:new stable  r:refresh  esc:back"
	case phaseNewStable:
		return "Tab/S-Tab:fields  +/-:adjust  Enter:create  esc:cancel"
	case phaseDetail:
		if m.selectedStable.Status == "training" {
			return "h:halt  r:refresh  esc:back"
		}
		if m.selectedStable.Status == "completed" {
			return "d:duel  P:promote  s:seed new  r:refresh  esc:back"
		}
		return "s:seed new  r:refresh  esc:back"
	case phaseDuel:
		return "esc:stop duel"
	case phaseHeroes:
		return "j/k:navigate  Enter:view  esc:back to stables"
	case phaseHeroDetail:
		return "d:duel  esc:back to heroes"
	case phasePromote:
		return "type name  Enter:confirm  esc:cancel"
	case phaseHeroDuel:
		return "esc:back to hero"
	default:
		return ""
	}
}

// StatusInfo returns data for the shared status bar.
func (m *Model) StatusInfo() studio.StatusInfo {
	return studio.StatusInfo{
		GameName: "Stables",
	}
}

// Close tears down any active streams.
func (m *Model) Close() {
	if m.trainingStream != nil {
		m.trainingStream.Close()
		m.trainingStream = nil
	}
	if m.duelStream != nil {
		m.duelStream.Close()
		m.duelStream = nil
	}
}

// Update handles all Bubble Tea messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	// List responses
	case StablesListMsg:
		m.stables = msg.Stables
		m.listLoaded = true
		m.err = nil
		if m.listIndex >= len(m.stables) && len(m.stables) > 0 {
			m.listIndex = len(m.stables) - 1
		}
		return nil

	case StablesListErrMsg:
		m.err = msg.Err
		m.listLoaded = true
		return nil

	// Detail responses
	case StableDetailMsg:
		m.selectedStable = msg.Stable
		m.err = nil
		return nil

	case StableDetailErrMsg:
		m.err = msg.Err
		return nil

	case ChampionMsg:
		m.champion = &msg.Champion
		return nil

	case ChampionErrMsg:
		// No champion is not an error — just means training hasn't completed
		m.champion = nil
		return nil

	case GenerationsMsg:
		m.generations = msg.Generations
		return nil

	case GenerationsErrMsg:
		// Non-critical
		return nil

	// Stable creation
	case StableCreatedMsg:
		m.err = nil
		m.phase = phaseList
		return FetchStables(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())

	case StableCreateErrMsg:
		m.err = msg.Err
		return nil

	// Halt
	case TrainingHaltedMsg:
		m.selectedStable.Status = "halted"
		m.closeTrainingStream()
		return m.refreshDetail()

	case TrainingHaltErrMsg:
		m.err = msg.Err
		return nil

	// Training SSE stream
	case TrainingUpdateMsg:
		m.lastProgress = &msg.Progress
		if !msg.Progress.Running {
			m.selectedStable.Status = msg.Progress.Status
			m.closeTrainingStream()
			return m.refreshDetail()
		}
		return m.pollTrainingStream()

	case TrainingStreamContinueMsg:
		return m.pollTrainingStream()

	case TrainingStreamDoneMsg:
		m.closeTrainingStream()
		return m.refreshDetail()

	case trainingPollTickMsg:
		if m.trainingStream != nil {
			return m.trainingStream.PollCmd()
		}
		return nil

	// Duel lifecycle (reuse snake_duel messages)
	case DuelStartedMsg:
		m.duelMatchID = msg.MatchID
		m.phase = phaseDuel
		m.duelStream = snake_duel.NewMatchStream(
			m.ctx.Client.SocketPath(),
			m.ctx.Client.BaseURL(),
		)
		return m.duelStream.Connect(m.duelMatchID)

	case DuelStartErrMsg:
		m.err = msg.Err
		return nil

	// Hero lifecycle
	case HeroesListMsg:
		m.heroes = msg.Heroes
		m.err = nil
		if m.heroIndex >= len(m.heroes) && len(m.heroes) > 0 {
			m.heroIndex = len(m.heroes) - 1
		}
		return nil

	case HeroesListErrMsg:
		m.err = msg.Err
		return nil

	case HeroDetailMsg:
		m.selectedHero = &msg.Hero
		m.err = nil
		return nil

	case HeroDetailErrMsg:
		m.err = msg.Err
		return nil

	case HeroPromotedMsg:
		m.phase = phaseHeroes
		m.promoteName = ""
		m.err = nil
		return FetchHeroes(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL())

	case HeroPromoteErrMsg:
		m.err = msg.Err
		return nil

	case HeroDuelStartedMsg:
		m.duelMatchID = msg.MatchID
		m.phase = phaseHeroDuel
		m.duelStream = snake_duel.NewMatchStream(
			m.ctx.Client.SocketPath(),
			m.ctx.Client.BaseURL(),
		)
		return m.duelStream.Connect(m.duelMatchID)

	case HeroDuelStartErrMsg:
		m.err = msg.Err
		return nil

	// Duel stream messages (forwarded from snake_duel's MatchStream)
	case snake_duel.MatchStateMsg:
		m.duelState = msg.State
		if msg.State.Status == "finished" {
			if m.duelStream != nil {
				m.duelStream.Close()
				m.duelStream = nil
			}
			return nil
		}
		return m.pollDuelStream()

	case snake_duel.MatchContinueMsg:
		return m.pollDuelStream()

	case snake_duel.MatchDoneMsg:
		if m.duelStream != nil {
			m.duelStream.Close()
			m.duelStream = nil
		}
		return nil

	case duelPollTickMsg:
		if m.duelStream != nil {
			return m.duelStream.PollCmd()
		}
		return nil
	}

	return nil
}

// openDetail transitions to detail view for the selected stable.
func (m *Model) openDetail() tea.Cmd {
	if len(m.stables) == 0 {
		return nil
	}
	m.selectedStable = m.stables[m.listIndex]
	m.champion = nil
	m.generations = nil
	m.lastProgress = nil
	m.phase = phaseDetail
	m.err = nil

	sp := m.ctx.Client.SocketPath()
	bu := m.ctx.Client.BaseURL()
	sid := m.selectedStable.StableID

	cmds := []tea.Cmd{
		FetchStable(sp, bu, sid),
		FetchChampion(sp, bu, sid),
		FetchGenerations(sp, bu, sid),
	}

	// Auto-connect to training stream if still training
	if m.selectedStable.Status == "training" {
		m.trainingStream = NewTrainingStream(sp, bu)
		cmds = append(cmds, m.trainingStream.Connect(sid))
	}

	return tea.Batch(cmds...)
}

// refreshDetail re-fetches detail data.
func (m *Model) refreshDetail() tea.Cmd {
	sp := m.ctx.Client.SocketPath()
	bu := m.ctx.Client.BaseURL()
	sid := m.selectedStable.StableID
	return tea.Batch(
		FetchStable(sp, bu, sid),
		FetchChampion(sp, bu, sid),
		FetchGenerations(sp, bu, sid),
	)
}

// createStable initiates a new stable from form values.
func (m *Model) createStable() tea.Cmd {
	req := InitiateStableRequest{
		PopulationSize:  m.formFields[0],
		MaxGenerations:  m.formFields[1],
		OpponentAF:      m.formFields[2],
		EpisodesPerEval: m.formFields[3],
		SeedStableID:    m.formSeedID,
	}

	// Add training config with fitness weights if not balanced (default)
	presetNames := []string{"balanced", "aggressive", "forager", "survivor", "assassin"}
	if m.formPreset > 0 || m.formShowWeights {
		tc := &TrainingConfig{}
		if m.formPreset > 0 && m.formPreset < len(presetNames) {
			tc.FitnessPreset = presetNames[m.formPreset]
		} else if m.formShowWeights {
			tc.FitnessWeights = &FitnessWeights{
				SurvivalWeight:  m.formWeights[0],
				FoodWeight:      m.formWeights[1],
				WinBonus:        m.formWeights[2],
				DrawBonus:       m.formWeights[3],
				KillBonus:       m.formWeights[4],
				ProximityWeight: m.formWeights[5],
				CirclePenalty:   m.formWeights[6],
			}
		}
		req.TrainingConfig = tc
	}

	return InitiateStable(m.ctx.Client.SocketPath(), m.ctx.Client.BaseURL(), req)
}

// closeTrainingStream cleans up the training SSE stream.
func (m *Model) closeTrainingStream() {
	if m.trainingStream != nil {
		m.trainingStream.Close()
		m.trainingStream = nil
	}
}

// pollTrainingStream returns a command to poll after a short delay.
func (m *Model) pollTrainingStream() tea.Cmd {
	if m.trainingStream == nil {
		return nil
	}
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return trainingPollTickMsg{}
	})
}

// pollDuelStream returns a command to poll the duel SSE stream.
func (m *Model) pollDuelStream() tea.Cmd {
	if m.duelStream == nil {
		return nil
	}
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return duelPollTickMsg{}
	})
}

type trainingPollTickMsg struct{}
type duelPollTickMsg struct{}
