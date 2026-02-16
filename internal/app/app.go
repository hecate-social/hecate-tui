package app

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/config"
	"github.com/hecate-social/hecate-tui/internal/factbus"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/statusbar"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/studios/arcade"
	"github.com/hecate-social/hecate-tui/internal/studios/devops"
	llmstudio "github.com/hecate-social/hecate-tui/internal/studios/llm"
	"github.com/hecate-social/hecate-tui/internal/studios/node"
	"github.com/hecate-social/hecate-tui/internal/studios/social"
	"github.com/hecate-social/hecate-tui/internal/theme"

	"github.com/hecate-social/hecate-tui/internal/client"
)

// App is the root Bubble Tea model — the shell that manages studios.
type App struct {
	client *client.Client
	theme  *theme.Theme
	styles *theme.Styles

	// Dimensions
	width  int
	height int

	// Studios
	studios      []studio.Studio
	activeStudio int
	showHome     bool // first launch = home screen

	// Command mode (shell-level)
	inCommandMode bool
	cmdInput      textinput.Model
	cmdHistory    []string
	cmdHistIdx    int
	cmdDraft      string

	// Command registry
	registry *commands.Registry

	// Status bar
	statusBar statusbar.Model

	// Persistent config
	cfg config.Config

	// Health polling
	daemonStatus string

	// Fact stream (SSE from daemon)
	factConn            *factbus.Connection
	factStreamConnected bool
	rxActive            bool
	txActive            bool

	// Flash notification (shown in hints area, auto-clears)
	flashMsg string
}

// New creates a new App with the modal chat interface.
func New(hecateURL string) *App {
	cfg := config.Load()
	if cfg.DaemonURL() != "" && hecateURL == "http://localhost:4444" {
		hecateURL = cfg.DaemonURL()
	}
	return newApp(client.New(hecateURL), cfg)
}

// NewWithSocket creates a new App connected via Unix domain socket.
func NewWithSocket(socketPath string) *App {
	return newApp(client.NewWithSocket(socketPath), config.Load())
}

// newApp builds the App with all shared initialization.
func newApp(c *client.Client, cfg config.Config) *App {
	t := theme.HecateDark()
	if cfg.Theme != "" {
		if saved, ok := theme.BuiltinThemes()[cfg.Theme]; ok {
			t = saved
		}
	}
	s := t.ComputeStyles()

	ci := textinput.New()
	ci.Placeholder = "command..."
	ci.Prompt = "/"
	ci.PromptStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	ci.TextStyle = lipgloss.NewStyle().Foreground(t.Text)
	ci.CharLimit = 256

	sb := statusbar.New(t, s)
	if cwd, err := os.Getwd(); err == nil {
		sb.Cwd = cwd
	}

	// Create factbus connection
	var fc *factbus.Connection
	if c.SocketPath() != "" {
		fc = factbus.NewConnection(c.SocketPath(), "http://localhost")
	} else {
		fc = factbus.NewConnection("", c.BaseURL())
	}

	// Create studio context (shared resources)
	ctx := &studio.Context{
		Client:  c,
		Theme:   t,
		Styles:  s,
		Config:  cfg,
		FactBus: fc,
	}

	// Create all studios
	studios := []studio.Studio{
		llmstudio.New(ctx),
		devops.New(ctx),
		node.New(ctx),
		social.New(ctx),
		arcade.New(ctx),
	}

	// Determine initial studio
	activeStudio := 0
	showHome := true
	if cfg.LastStudio >= 0 && cfg.LastStudio < len(studios) {
		activeStudio = cfg.LastStudio
		showHome = false
	}

	return &App{
		client:       c,
		theme:        t,
		styles:       s,
		cfg:          cfg,
		studios:      studios,
		activeStudio: activeStudio,
		showHome:     showHome,
		statusBar:    sb,
		cmdInput:     ci,
		registry:     commands.NewRegistry(),
		factConn:     fc,
	}
}

// Init starts the app — health polling, fact stream, active studio init.
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{
		a.checkHealth,
		a.scheduleHealthTick(),
		a.factConn.Subscribe(),
	}

	if !a.showHome {
		a.studios[a.activeStudio].SetFocused(true)
		cmds = append(cmds, a.studios[a.activeStudio].Init())
	}

	return tea.Batch(cmds...)
}

// rxFlashDoneMsg resets the RX LED after the flash duration.
type rxFlashDoneMsg struct{}

// txFlashDoneMsg resets the TX LED after the flash duration.
type txFlashDoneMsg struct{}

// Update is the main message loop.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)
		contentHeight := a.contentAreaHeight()
		for _, s := range a.studios {
			s.SetSize(msg.Width, contentHeight)
		}

	case tea.KeyMsg:
		cmd := a.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// If shell handled the key (command mode, studio switch, quit), don't forward
		if a.shellConsumedKey(msg) {
			return a, tea.Batch(cmds...)
		}

	case healthMsg:
		a.daemonStatus = msg.status
		a.statusBar.DaemonStatus = msg.status

	case healthTickMsg:
		cmds = append(cmds, a.checkHealth, a.scheduleHealthTick())

	case commands.SwitchThemeMsg:
		a.switchTheme(msg.Theme)

	case commands.SwitchStudioMsg:
		cmd := a.switchStudio(msg.Index)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case commands.ChangeDirMsg:
		if err := os.Chdir(msg.Path); err == nil {
			a.statusBar.Cwd = msg.Path
		}
		// Still forward to active studio

	case commands.VentureCreatedMsg:
		if err := os.Chdir(msg.Path); err == nil {
			a.statusBar.Cwd = msg.Path
		}
		// Show flash notification visible in any studio
		cmds = append(cmds, a.setFlash("Venture created: "+msg.Path))

	case commands.InjectSystemMsg:
		// Show flash notification visible in any studio
		cmds = append(cmds, a.setFlash(stripAnsi(msg.Content)))

	// Fact stream messages
	case factbus.FactMsg:
		a.factStreamConnected = true
		a.rxActive = true
		cmds = append(cmds, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return rxFlashDoneMsg{}
		}))
		cmd := a.handleFact(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case factbus.FactContinueMsg:
		a.factStreamConnected = true
		cmds = append(cmds, a.scheduleFactPoll())

	case factbus.FactDisconnectedMsg:
		a.factStreamConnected = false

	case rxFlashDoneMsg:
		a.rxActive = false

	case txFlashDoneMsg:
		a.txActive = false

	case flashClearMsg:
		a.flashMsg = ""
	}

	// Forward message to active studio
	if !a.showHome && a.activeStudio < len(a.studios) {
		// Track streaming state for TX LED
		llm := a.llmStudio()
		wasStreaming := false
		if llm != nil {
			wasStreaming = llm.IsStreaming()
		}

		updated, cmd := a.studios[a.activeStudio].Update(msg)
		a.studios[a.activeStudio] = updated
		cmds = append(cmds, cmd)

		// Check if streaming just started (for TX LED)
		if llm != nil {
			nowStreaming := llm.IsStreaming()
			if !wasStreaming && nowStreaming {
				a.txActive = true
				cmds = append(cmds, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
					return txFlashDoneMsg{}
				}))
				a.rxActive = true
			}
			if nowStreaming {
				a.rxActive = true
			}
			if wasStreaming && !nowStreaming {
				a.rxActive = false
			}
		}
	}

	// Sync status bar from active studio
	a.syncStatusBar()

	return a, tea.Batch(cmds...)
}

// shellConsumedKey returns true if the shell handled a key and it shouldn't
// be forwarded to the studio.
func (a *App) shellConsumedKey(msg tea.KeyMsg) bool {
	key := msg.String()

	// Always consumed by shell
	if key == "ctrl+c" {
		return true
	}

	// Home screen keys
	if a.showHome {
		return true
	}

	// Command mode consumes all keys
	if a.inCommandMode {
		return true
	}

	// Studio switch keys in Normal mode
	activeMode := a.studios[a.activeStudio].Mode()
	if activeMode == modes.Normal {
		switch key {
		case "ctrl+1", "ctrl+2", "ctrl+3", "ctrl+4", "ctrl+5":
			return true
		case "q":
			return true
		case "/", ":":
			return true
		}
	}

	return false
}

func (a *App) switchStudio(index int) tea.Cmd {
	if index < 0 || index >= len(a.studios) {
		return nil
	}
	if index == a.activeStudio && !a.showHome {
		return nil
	}

	// Unfocus current studio
	if !a.showHome {
		a.studios[a.activeStudio].SetFocused(false)
	}

	a.activeStudio = index
	a.showHome = false
	a.studios[index].SetFocused(true)

	// Persist last studio
	a.cfg.LastStudio = index
	_ = a.cfg.Save()

	return a.studios[index].Init()
}

// flashClearMsg clears the flash notification after a delay.
type flashClearMsg struct{}

// setFlash shows a brief notification in the hints area, auto-clears after 4 seconds.
func (a *App) setFlash(msg string) tea.Cmd {
	// Truncate long messages to a single line
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = msg[:idx]
	}
	if len(msg) > 80 {
		msg = msg[:77] + "..."
	}
	a.flashMsg = msg
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return flashClearMsg{}
	})
}

// stripAnsi removes ANSI escape sequences from a string for flash display.
func stripAnsi(s string) string {
	var result strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func (a *App) switchTheme(t *theme.Theme) {
	a.theme = t
	a.styles = t.ComputeStyles()

	// Rebuild status bar
	a.statusBar = statusbar.New(t, a.styles)
	a.statusBar.SetWidth(a.width)
	a.statusBar.DaemonStatus = a.daemonStatus

	// Update command input styling
	a.cmdInput.PromptStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	a.cmdInput.TextStyle = lipgloss.NewStyle().Foreground(t.Text)

	// Persist theme choice
	a.saveThemeToConfig(t)

	// Update LLM studio theme
	if llm := a.llmStudio(); llm != nil {
		llm.SwitchTheme(t, a.styles)
	}
}

func (a *App) saveThemeToConfig(t *theme.Theme) {
	for key, builtin := range theme.BuiltinThemes() {
		if builtin.Name == t.Name {
			a.cfg.Theme = key
			_ = a.cfg.Save()
			return
		}
	}
}

func (a *App) syncStatusBar() {
	if a.showHome || a.activeStudio >= len(a.studios) {
		return
	}

	active := a.studios[a.activeStudio]
	info := active.StatusInfo()

	a.statusBar.Mode = active.Mode()
	a.statusBar.FlashMsg = a.flashMsg
	a.statusBar.ModelName = info.ModelName
	a.statusBar.ModelProvider = info.ModelProvider
	a.statusBar.ModelStatus = info.ModelStatus
	a.statusBar.ModelError = info.ModelError
	a.statusBar.InputLen = info.InputLen
	a.statusBar.SessionTokens = info.SessionTokens

	// ALC context from LLM studio
	if llm := a.llmStudio(); llm != nil {
		alcState := llm.ALCState()
		if alcState != nil && alcState.Venture != nil {
			a.statusBar.VentureName = alcState.Venture.Name
		} else {
			a.statusBar.VentureName = ""
		}
		if alcState != nil && alcState.Department != nil {
			a.statusBar.ActivePhase = string(alcState.Department.CurrentPhase)
		} else {
			a.statusBar.ActivePhase = ""
		}
	}
}

// llmStudio returns the LLM studio (always index 0), cast to the concrete type.
func (a *App) llmStudio() *llmstudio.Studio {
	if len(a.studios) == 0 {
		return nil
	}
	if s, ok := a.studios[0].(*llmstudio.Studio); ok {
		return s
	}
	return nil
}

// commandContext builds a commands.Context for command dispatch.
// Routes to the LLM studio's context if available.
func (a *App) commandContext() *commands.Context {
	if llm := a.llmStudio(); llm != nil {
		ctx := llm.CommandContext()
		ctx.Width = a.width
		ctx.Height = a.height
		return ctx
	}

	// Fallback for non-LLM studios
	return &commands.Context{
		Client: a.client,
		Theme:  a.theme,
		Styles: a.styles,
		Width:  a.width,
		Height: a.height,
	}
}

// contentAreaHeight returns the height available for studio content.
func (a *App) contentAreaHeight() int {
	headerHeight := 4 // brand row + context row + tab bar + separator
	statusBarHeight := 2
	commandHeight := 0
	if a.inCommandMode {
		commandHeight = 1
	}

	h := a.height - headerHeight - statusBarHeight - commandHeight
	if h < 5 {
		h = 5
	}
	return h
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

// healthMsg carries daemon health check results.
type healthMsg struct {
	status string
}

// healthTickMsg triggers periodic health polling.
type healthTickMsg struct{}

// renderHeader builds the header: brand row + context row + tab bar + separator.
func (a *App) renderHeader() string {
	var rows []string

	// Brand row
	rows = append(rows, a.renderBrandRow())

	// Context row (ALC)
	contextRow := a.renderContextRow()
	if contextRow != "" {
		rows = append(rows, contextRow)
	}

	// Tab bar
	rows = append(rows, a.renderTabBar())

	// Separator
	sep := lipgloss.NewStyle().Foreground(a.theme.Border).Render(strings.Repeat("─", a.width))
	rows = append(rows, sep)

	return strings.Join(rows, "\n")
}
