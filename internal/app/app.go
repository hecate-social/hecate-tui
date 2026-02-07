package app

import (
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/browse"
	"github.com/hecate-social/hecate-tui/internal/chat"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/config"
	"github.com/hecate-social/hecate-tui/internal/editor"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/pair"
	"github.com/hecate-social/hecate-tui/internal/statusbar"
	"github.com/hecate-social/hecate-tui/internal/theme"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// App is the root Bubble Tea model — the modal chat interface.
type App struct {
	client *client.Client
	theme  *theme.Theme
	styles *theme.Styles

	// Dimensions
	width  int
	height int

	// Modal state
	mode     modes.Mode
	prevMode modes.Mode

	// Components
	chat       chat.Model
	browseView browse.Model
	pairView   pair.Model
	editorView editor.Model
	statusBar  statusbar.Model
	cmdInput   textinput.Model
	registry   *commands.Registry

	// Tool system
	toolExecutor   *llmtools.Executor
	approvalPrompt *ui.ApprovalPrompt

	// Overlay modes initialized
	browseReady bool
	pairReady   bool
	editorReady bool

	// Command history
	cmdHistory []string
	cmdHistIdx int    // -1 = new input, 0..N = browsing history
	cmdDraft   string // save draft when browsing history

	// System prompt for LLM
	systemPrompt string

	// Persistent config
	cfg               config.Config
	conversationID    string
	conversationTitle string

	// Health polling
	daemonStatus string
}

// New creates a new App with the modal chat interface.
func New(hecateURL string) *App {
	// Load persistent config
	cfg := config.Load()

	// Config can override daemon URL
	if cfg.DaemonURL() != "" && hecateURL == "http://localhost:4444" {
		hecateURL = cfg.DaemonURL()
	}

	c := client.New(hecateURL)

	// Apply saved theme or default
	t := theme.HecateDark()
	if cfg.Theme != "" {
		if saved, ok := theme.BuiltinThemes()[cfg.Theme]; ok {
			t = saved
		}
	}
	s := t.ComputeStyles()

	// Command line input
	ci := textinput.New()
	ci.Placeholder = "command..."
	ci.Prompt = "/"
	ci.PromptStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	ci.TextStyle = lipgloss.NewStyle().Foreground(t.Text)
	ci.CharLimit = 256

	sb := statusbar.New(t, s)

	chatModel := chat.New(c, t, s)

	// Build and apply system prompt (combines personality, role, and custom prompt)
	systemPrompt := cfg.BuildSystemPrompt()
	if systemPrompt != "" {
		chatModel.SetSystemPrompt(systemPrompt)
	}

	// Apply saved model preference
	if cfg.Model != "" {
		chatModel.SetPreferredModel(cfg.Model)
	}

	// Initialize tool system
	toolRegistry := llmtools.NewDefaultRegistry()
	toolPermissions := llmtools.NewPermissions()
	toolExecutor := llmtools.NewExecutor(toolRegistry, toolPermissions)

	// Wire up tool executor to chat
	chatModel.SetToolExecutor(toolExecutor)
	// Tools disabled by default - most Ollama models don't support function calling.
	// Use /tools enable to turn on for models that support it (Claude, GPT-4, etc.)
	chatModel.EnableTools(false)

	// Set mesh client for mesh tools
	llmtools.SetMeshClient(c)

	// Create approval prompt for tool authorization
	approvalPrompt := ui.NewApprovalPrompt(t, s)

	// Auto-load most recent conversation
	convID := config.NewConversationID()
	convTitle := ""
	if convs := config.ListConversations(); len(convs) > 0 {
		latest := convs[0]
		convID = latest.ID
		convTitle = latest.Title
		var msgs []chat.Message
		for _, m := range latest.Messages {
			msgs = append(msgs, chat.Message{
				Role:    m.Role,
				Content: m.Content,
				Time:    m.Time,
			})
		}
		chatModel.LoadMessages(msgs)
	}

	return &App{
		client:            c,
		theme:             t,
		styles:            s,
		cfg:               cfg,
		conversationID:    convID,
		conversationTitle: convTitle,
		mode:              modes.Normal,
		chat:              chatModel,
		systemPrompt:      systemPrompt,
		statusBar:         sb,
		cmdInput:          ci,
		registry:          commands.NewRegistry(),
		toolExecutor:      toolExecutor,
		approvalPrompt:    approvalPrompt,
	}
}

// NewWithSocket creates a new App connected via Unix domain socket.
func NewWithSocket(socketPath string) *App {
	cfg := config.Load()

	c := client.NewWithSocket(socketPath)

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

	chatModel := chat.New(c, t, s)

	// Build and apply system prompt (combines personality, role, and custom prompt)
	systemPrompt := cfg.BuildSystemPrompt()
	if systemPrompt != "" {
		chatModel.SetSystemPrompt(systemPrompt)
	}

	if cfg.Model != "" {
		chatModel.SetPreferredModel(cfg.Model)
	}

	convID := config.NewConversationID()
	convTitle := ""
	if convs := config.ListConversations(); len(convs) > 0 {
		latest := convs[0]
		convID = latest.ID
		convTitle = latest.Title
		var msgs []chat.Message
		for _, m := range latest.Messages {
			msgs = append(msgs, chat.Message{
				Role:    m.Role,
				Content: m.Content,
				Time:    m.Time,
			})
		}
		chatModel.LoadMessages(msgs)
	}

	// Initialize tool system
	toolRegistry := llmtools.NewDefaultRegistry()
	toolPermissions := llmtools.NewPermissions()
	toolExecutor := llmtools.NewExecutor(toolRegistry, toolPermissions)

	// Wire up tool executor to chat
	chatModel.SetToolExecutor(toolExecutor)
	// Tools disabled by default - most Ollama models don't support function calling
	chatModel.EnableTools(false)

	// Set mesh client for mesh tools
	llmtools.SetMeshClient(c)

	// Create approval prompt for tool authorization
	approvalPrompt := ui.NewApprovalPrompt(t, s)

	return &App{
		client:            c,
		theme:             t,
		styles:            s,
		cfg:               cfg,
		conversationID:    convID,
		conversationTitle: convTitle,
		mode:              modes.Normal,
		chat:              chatModel,
		systemPrompt:      systemPrompt,
		statusBar:         sb,
		cmdInput:          ci,
		registry:          commands.NewRegistry(),
		toolExecutor:      toolExecutor,
		approvalPrompt:    approvalPrompt,
	}
}

// healthMsg carries daemon health check results.
type healthMsg struct {
	status string
}

// healthTickMsg triggers periodic health polling.
type healthTickMsg struct{}

// Init starts the app — fetch models, check health, start polling.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.chat.Init(),
		a.checkHealth,
		a.scheduleHealthTick(),
	)
}

func (a *App) scheduleHealthTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return healthTickMsg{}
	})
}

// Update is the main message loop.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)
		chatHeight := a.chatAreaHeight()
		a.chat.SetSize(msg.Width, chatHeight)
		if a.browseReady {
			a.browseView.SetSize(msg.Width, msg.Height) // Pass terminal size; modal calculates its own dimensions
		}
		if a.pairReady {
			a.pairView.SetSize(a.pairWidth(), a.pairHeight())
		}
		if a.editorReady {
			a.editorView.SetSize(a.width, a.editorHeight())
		}

	case tea.KeyMsg:
		cmd := a.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
		a.statusBar.Mode = a.mode
		a.statusBar.InputLen = a.chat.InputLen()
		a.statusBar.SessionTokens = a.chat.SessionTokenCount()

	case healthMsg:
		a.daemonStatus = msg.status
		a.statusBar.DaemonStatus = msg.status
		// Sync status bar model from chat state (don't override with first model from list)
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()

	case healthTickMsg:
		cmds = append(cmds, a.checkHealth, a.scheduleHealthTick())

	// Command system messages
	case commands.InjectSystemMsg:
		a.chat.InjectSystemMessage(msg.Content)

	case commands.ClearChatMsg:
		a.chat.ClearMessages()

	case commands.SwitchModelMsg:
		a.chat.SwitchModel(msg.Name)
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
		// Persist model selection
		a.cfg.Model = msg.Name
		_ = a.cfg.Save()

	case commands.SetModeMsg:
		cmd := a.enterMode(modes.Mode(msg.Mode))
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.SwitchThemeMsg:
		a.switchTheme(msg.Theme)

	case commands.EditFileMsg:
		cmd := a.openEditor(msg.Path)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.NewConversationMsg:
		a.startNewConversation()
		a.chat.InjectSystemMessage("Started new conversation.")

	case commands.LoadConversationMsg:
		if err := a.loadConversation(msg.ID); err != nil {
			a.chat.InjectSystemMessage("Failed to load: " + err.Error())
		} else {
			a.chat.InjectSystemMessage("Loaded conversation: " + msg.ID)
		}

	case commands.EnableToolsMsg:
		a.chat.EnableTools(msg.Enabled)
		status := "disabled"
		if msg.Enabled {
			status = "enabled"
		}
		a.chat.InjectSystemMessage("LLM function calling " + status)

	case browse.SelectModelMsg:
		// User selected an LLM model from browse
		a.chat.SwitchModel(msg.ModelName)
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
		a.setMode(modes.Normal)
		a.chat.InjectSystemMessage("Model switched to: " + msg.ModelName)
		// Persist model selection
		a.cfg.Model = msg.ModelName
		_ = a.cfg.Save()

	case commands.SwitchRoleMsg:
		a.cfg.Personality.ActiveRole = msg.Role
		_ = a.cfg.Save()
		// Rebuild and apply system prompt
		newPrompt := a.cfg.BuildSystemPrompt()
		a.systemPrompt = newPrompt
		a.chat.SetSystemPrompt(newPrompt)
		roleName := a.cfg.ActiveRoleDisplayName()
		if roleName == "" {
			roleName = msg.Role
		}
		a.chat.InjectSystemMessage("Role switched to: " + roleName)
	}

	// Forward to chat for streaming updates
	wasStreaming := a.chat.IsStreaming()
	var chatCmd tea.Cmd
	a.chat, chatCmd = a.chat.Update(msg)
	cmds = append(cmds, chatCmd)

	// Auto-save when streaming completes (assistant response finished)
	if wasStreaming && !a.chat.IsStreaming() {
		a.saveConversation()
	}

	// Forward to browse if in Browse mode
	if a.mode == modes.Browse && a.browseReady {
		var browseCmd tea.Cmd
		a.browseView, browseCmd = a.browseView.Update(msg)
		cmds = append(cmds, browseCmd)
	}

	// Forward to pair if in Pair mode
	if a.mode == modes.Pair && a.pairReady {
		var pairCmd tea.Cmd
		a.pairView, pairCmd = a.pairView.Update(msg)
		cmds = append(cmds, pairCmd)
	}

	// Forward to editor if in Edit mode (non-key msgs like blink)
	if a.mode == modes.Edit && a.editorReady {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			var edCmd tea.Cmd
			var model tea.Model
			model, edCmd = a.editorView.Update(msg)
			a.editorView = model.(editor.Model)
			cmds = append(cmds, edCmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// handleKey dispatches keys based on current mode.
func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if key == "ctrl+c" {
		return tea.Quit
	}

	switch a.mode {
	case modes.Normal:
		return a.handleNormalKey(key)
	case modes.Insert:
		return a.handleInsertKey(key)
	case modes.Command:
		return a.handleCommandKey(key, msg)
	case modes.Browse:
		return a.handleBrowseKey(key, msg)
	case modes.Pair:
		return a.handlePairKey(key, msg)
	case modes.Edit:
		return a.handleEditKey(key, msg)
	default:
		if key == "esc" {
			a.setMode(modes.Normal)
		}
		return nil
	}
}

func (a *App) handleNormalKey(key string) tea.Cmd {
	// Handle tool approval keys when there's a pending approval
	if a.chat.HasPendingApproval() {
		switch key {
		case "y": // Allow this time
			return a.chat.ApproveToolCall(false)
		case "n": // Deny
			return a.chat.DenyToolCall()
		case "a": // Allow all for session
			return a.chat.ApproveToolCall(true)
		case "esc": // Cancel (same as deny)
			return a.chat.DenyToolCall()
		}
		// Block other keys while approval is pending
		return nil
	}

	switch key {
	case "i":
		a.setMode(modes.Insert)
	case "/", ":":
		a.setMode(modes.Command)
		if key == ":" {
			a.cmdInput.Prompt = ":"
		} else {
			a.cmdInput.Prompt = "/"
		}
	case "j", "down":
		a.chat.ScrollDown(1)
	case "k", "up":
		a.chat.ScrollUp(1)
	case "ctrl+d":
		a.chat.HalfPageDown()
	case "ctrl+u":
		a.chat.HalfPageUp()
	case "g":
		a.chat.GotoTop()
	case "G":
		a.chat.GotoBottom()
	case "?":
		ctx := a.commandContext()
		return commands.ModeHelp(int(a.mode), ctx)
	case "r":
		return a.chat.RetryLast()
	case "y":
		return a.yankLastResponse()
	case "q":
		return tea.Quit
	}
	return nil
}

func (a *App) handleInsertKey(key string) tea.Cmd {
	// Handle tool approval keys when there's a pending approval
	if a.chat.HasPendingApproval() {
		switch key {
		case "y": // Allow this time
			return a.chat.ApproveToolCall(false)
		case "n": // Deny
			return a.chat.DenyToolCall()
		case "a": // Allow all for session
			return a.chat.ApproveToolCall(true)
		case "esc": // Cancel (same as deny)
			return a.chat.DenyToolCall()
		}
		// Block other keys while approval is pending
		return nil
	}

	switch key {
	case "esc":
		if a.chat.IsStreaming() {
			a.chat.CancelStreaming()
			return nil
		}
		a.setMode(modes.Normal)
	case "enter":
		cmd := a.chat.SendCurrentInput()
		if cmd != nil {
			a.saveConversation()
		}
		return cmd
	case "alt+enter":
		a.chat.InsertNewline()
	case "tab":
		a.chat.CycleModel()
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
	case "shift+tab":
		a.chat.CycleModelReverse()
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
	}
	return nil
}

func (a *App) handleCommandKey(key string, msg tea.KeyMsg) tea.Cmd {
	switch key {
	case "esc":
		a.cmdHistIdx = -1
		a.setMode(modes.Normal)
		return nil
	case "enter":
		input := a.cmdInput.Value()
		a.setMode(modes.Normal)
		if input != "" {
			// Save to history (avoid duplicates at top)
			if len(a.cmdHistory) == 0 || a.cmdHistory[len(a.cmdHistory)-1] != input {
				a.cmdHistory = append(a.cmdHistory, input)
				if len(a.cmdHistory) > 50 {
					a.cmdHistory = a.cmdHistory[1:]
				}
			}
			a.cmdHistIdx = -1
			ctx := a.commandContext()
			prefix := a.cmdInput.Prompt
			return a.registry.Dispatch(prefix+input, ctx)
		}
		return nil
	case "up":
		if len(a.cmdHistory) == 0 {
			return nil
		}
		if a.cmdHistIdx == -1 {
			// Save current draft and start browsing
			a.cmdDraft = a.cmdInput.Value()
			a.cmdHistIdx = len(a.cmdHistory) - 1
		} else if a.cmdHistIdx > 0 {
			a.cmdHistIdx--
		}
		a.cmdInput.SetValue(a.cmdHistory[a.cmdHistIdx])
		a.cmdInput.CursorEnd()
		return nil
	case "down":
		if a.cmdHistIdx == -1 {
			return nil
		}
		if a.cmdHistIdx < len(a.cmdHistory)-1 {
			a.cmdHistIdx++
			a.cmdInput.SetValue(a.cmdHistory[a.cmdHistIdx])
		} else {
			// Back to draft
			a.cmdHistIdx = -1
			a.cmdInput.SetValue(a.cmdDraft)
		}
		a.cmdInput.CursorEnd()
		return nil
	case "tab":
		prefix := a.cmdInput.Value()
		matches := a.registry.Complete(prefix)
		if len(matches) == 1 {
			a.cmdInput.SetValue(matches[0])
			a.cmdInput.CursorEnd()
		}
		return nil
	default:
		a.cmdHistIdx = -1 // reset history browsing on any other key
		var cmd tea.Cmd
		a.cmdInput, cmd = a.cmdInput.Update(msg)
		return cmd
	}
}

func (a *App) handleBrowseKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !a.browseReady {
		return nil
	}

	if key == "?" {
		ctx := a.commandContext()
		return commands.ModeHelp(int(a.mode), ctx)
	}

	// Esc in Browse returns to Normal (unless browse handles it internally)
	if key == "esc" {
		a.setMode(modes.Normal)
		return nil
	}

	consumed, cmd := a.browseView.HandleKey(key, msg)
	if consumed {
		return cmd
	}

	return nil
}

func (a *App) handlePairKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !a.pairReady {
		return nil
	}

	if key == "?" {
		ctx := a.commandContext()
		return commands.ModeHelp(int(a.mode), ctx)
	}

	consumed, cmd := a.pairView.HandleKey(key, msg)
	if consumed {
		return cmd
	}

	// Unconsumed Esc exits Pair mode
	if key == "esc" {
		a.setMode(modes.Normal)
		return nil
	}

	return nil
}

func (a *App) handleEditKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !a.editorReady {
		return nil
	}

	// Intercept Ctrl+Q and Esc to close editor instead of quitting
	switch key {
	case "ctrl+q", "esc":
		a.editorReady = false
		a.setMode(modes.Normal)
		return nil
	}

	// Forward to editor
	var cmd tea.Cmd
	var model tea.Model
	model, cmd = a.editorView.Update(msg)
	a.editorView = model.(editor.Model)

	// Catch tea.Quit from editor (e.g., save-then-quit flows) and convert to mode exit
	// Editor's own quit will be intercepted above via ctrl+q/esc
	return cmd
}

func (a *App) openEditor(path string) tea.Cmd {
	if path != "" {
		ed, err := editor.NewWithFile(path)
		if err != nil {
			a.chat.InjectSystemMessage("Could not open file: " + err.Error())
			return nil
		}
		a.editorView = ed
	} else {
		a.editorView = editor.New()
	}

	a.editorView.SetSize(a.width, a.editorHeight())
	a.editorView.Focus()
	a.editorReady = true
	a.setMode(modes.Edit)
	return a.editorView.Init()
}

func (a *App) editorHeight() int {
	return a.height - 2 // header + status bar
}

func (a *App) setMode(m modes.Mode) {
	if m == a.mode {
		return
	}
	a.prevMode = a.mode
	a.mode = m
	a.statusBar.Mode = m

	switch m {
	case modes.Normal:
		a.chat.SetInputVisible(false)
		a.cmdInput.Blur()
	case modes.Insert:
		a.chat.SetInputVisible(true)
		a.cmdInput.Blur()
	case modes.Command:
		a.chat.SetInputVisible(false)
		a.cmdInput.SetValue("")
		a.cmdInput.Focus()
	case modes.Browse:
		a.chat.SetInputVisible(false)
		a.cmdInput.Blur()
	case modes.Pair:
		a.chat.SetInputVisible(false)
		a.cmdInput.Blur()
	case modes.Edit:
		a.chat.SetInputVisible(false)
		a.cmdInput.Blur()
	}

	chatHeight := a.chatAreaHeight()
	a.chat.SetSize(a.width, chatHeight)
}

// enterMode switches mode with initialization (for modes that need setup).
func (a *App) enterMode(m modes.Mode) tea.Cmd {
	a.setMode(m)

	switch m {
	case modes.Browse:
		a.browseView = browse.New(a.client, a.theme, a.styles)
		a.browseView.SetSize(a.width, a.height) // Pass terminal size; modal calculates its own dimensions
		a.browseReady = true
		return a.browseView.Init()
	case modes.Pair:
		a.pairView = pair.New(a.client, a.theme, a.styles)
		a.pairView.SetSize(a.pairWidth(), a.pairHeight())
		a.pairReady = true
		return a.pairView.Init()
	}

	return nil
}

func (a *App) switchTheme(t *theme.Theme) {
	a.theme = t
	a.styles = t.ComputeStyles()

	// Rebuild components with new theme
	a.statusBar = statusbar.New(t, a.styles)
	a.statusBar.SetWidth(a.width)
	a.statusBar.Mode = a.mode
	a.statusBar.DaemonStatus = a.daemonStatus
	a.statusBar.ModelName = a.chat.ActiveModelName()
	a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
	a.statusBar.SessionTokens = a.chat.SessionTokenCount()

	// Rebuild chat with new theme (preserves messages via re-init)
	oldChat := a.chat
	a.chat = chat.New(a.client, t, a.styles)
	a.chat.SetSize(a.width, a.chatAreaHeight())
	// Transfer state — the new chat will re-fetch models via Init
	_ = oldChat

	// Update command input styling
	a.cmdInput.PromptStyle = lipgloss.NewStyle().Foreground(t.Warning).Bold(true)
	a.cmdInput.TextStyle = lipgloss.NewStyle().Foreground(t.Text)

	// Persist theme choice
	a.saveThemeToConfig(t)

	a.chat.InjectSystemMessage("Theme switched to: " + t.Name)
}

func (a *App) saveThemeToConfig(t *theme.Theme) {
	// Find the key for this theme
	for key, builtin := range theme.BuiltinThemes() {
		if builtin.Name == t.Name {
			a.cfg.Theme = key
			_ = a.cfg.Save()
			return
		}
	}
}

func (a *App) commandContext() *commands.Context {
	return &commands.Context{
		Client: a.client,
		Theme:  a.theme,
		Styles: a.styles,
		Width:  a.width,
		Height: a.height,
		SetMode: func(mode int) {
			a.setMode(modes.Mode(mode))
		},
		InjectChat: func(msg commands.ChatMessage) {
			a.chat.InjectSystemMessage(msg.Content)
		},
		GetMessages: func() []commands.ChatExportMsg {
			exported := a.chat.ExportMessages()
			var msgs []commands.ChatExportMsg
			for _, m := range exported {
				msgs = append(msgs, commands.ChatExportMsg{
					Role:    m.Role,
					Content: m.Content,
					Time:    m.Time,
				})
			}
			return msgs
		},
		GetSystemPrompt: func() string {
			return a.systemPrompt
		},
		SetSystemPrompt: func(prompt string) {
			a.systemPrompt = prompt
			a.chat.SetSystemPrompt(prompt)
			a.cfg.SystemPrompt = prompt
			_ = a.cfg.Save()
		},
		GetToolExecutor: func() *llmtools.Executor {
			return a.chat.ToolExecutor()
		},
		ToolsEnabled: func() bool {
			return a.chat.ToolsEnabled()
		},
		GetActiveRole: func() string {
			return a.cfg.Personality.ActiveRole
		},
		SetActiveRole: func(role string) error {
			a.cfg.Personality.ActiveRole = role
			return a.cfg.Save()
		},
		GetRoleNames: func() []string {
			return []string{"dna", "anp", "tni", "dno"}
		},
		RebuildPrompt: func() string {
			return a.cfg.BuildSystemPrompt()
		},
	}
}

func (a *App) chatAreaHeight() int {
	headerHeight := 2
	statusBarHeight := 1
	inputHeight := 0

	switch a.mode {
	case modes.Insert:
		inputHeight = 3 // 1 row + border
	case modes.Command:
		inputHeight = 1
	}

	statsHeight := 1
	h := a.height - headerHeight - statusBarHeight - inputHeight - statsHeight
	if h < 5 {
		h = 5
	}
	return h
}

// browseWidth returns the width for the browse overlay.
// Split pane on wide terminals (>= 100 cols), full width on narrow.
func (a *App) browseWidth() int {
	if a.width >= 100 {
		return a.width / 2
	}
	return a.width - 4
}

// browseHeight returns the height for the browse overlay.
func (a *App) browseHeight() int {
	return a.height - 4 // header + status bar + padding
}

// pairWidth returns the width for the pair overlay.
func (a *App) pairWidth() int {
	if a.width >= 100 {
		return a.width / 2
	}
	return a.width - 4
}

// pairHeight returns the height for the pair overlay.
func (a *App) pairHeight() int {
	return a.height - 4
}

// View renders the entire TUI.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	// Browse mode uses split or full layout
	if a.mode == modes.Browse && a.browseReady {
		return a.renderBrowseLayout()
	}

	// Pair mode uses split or full layout
	if a.mode == modes.Pair && a.pairReady {
		return a.renderPairLayout()
	}

	// Edit mode takes full screen
	if a.mode == modes.Edit && a.editorReady {
		return a.renderEditLayout()
	}

	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	// Chat area (always visible)
	sections = append(sections, a.chat.ViewChat())

	// Stats/streaming indicator
	if stats := a.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}

	// Error
	if errView := a.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}

	// Input area (mode-dependent)
	switch a.mode {
	case modes.Insert:
		sections = append(sections, a.chat.ViewInput())
	case modes.Command:
		sections = append(sections, a.renderCommandLine())
	}

	// Status bar (always at bottom)
	sections = append(sections, a.statusBar.View())

	content := strings.Join(sections, "\n")

	// Overlay tool approval prompt if there's a pending approval
	if a.chat.HasPendingApproval() && a.approvalPrompt != nil {
		content = a.renderWithApprovalOverlay(content)
	}

	return content
}

// renderWithApprovalOverlay overlays the approval prompt on top of the content.
func (a *App) renderWithApprovalOverlay(content string) string {
	call := a.chat.PendingToolCall()
	if call == nil {
		return content
	}

	// Get tool info from registry
	registry := a.toolExecutor.Registry()
	tool, _, ok := registry.Get(call.Name)
	if !ok {
		// Unknown tool, just show basic info
		tool = llmtools.Tool{
			Name:        call.Name,
			Description: "Unknown tool",
			Category:    llmtools.CategorySystem,
		}
	}

	// Set width based on terminal
	dialogWidth := 60
	if a.width > 80 {
		dialogWidth = 70
	}
	if a.width < 70 {
		dialogWidth = a.width - 4
	}
	a.approvalPrompt.SetWidth(dialogWidth)

	// Render the approval prompt
	prompt := a.approvalPrompt.Render(tool, *call)

	// Center the prompt on the screen
	promptLines := strings.Split(prompt, "\n")
	promptHeight := len(promptLines)

	// Calculate vertical position (center)
	contentLines := strings.Split(content, "\n")
	startLine := (len(contentLines) - promptHeight) / 2
	if startLine < 0 {
		startLine = 0
	}

	// Calculate horizontal padding to center
	maxPromptWidth := 0
	for _, line := range promptLines {
		if w := lipgloss.Width(line); w > maxPromptWidth {
			maxPromptWidth = w
		}
	}
	leftPad := (a.width - maxPromptWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	padding := strings.Repeat(" ", leftPad)

	// Overlay the prompt
	for i, line := range promptLines {
		lineIdx := startLine + i
		if lineIdx >= 0 && lineIdx < len(contentLines) {
			contentLines[lineIdx] = padding + line
		}
	}

	return strings.Join(contentLines, "\n")
}

func (a *App) renderBrowseLayout() string {
	// Render the normal chat view as the background
	var sections []string
	sections = append(sections, a.renderHeader())
	sections = append(sections, a.chat.ViewChat())
	if stats := a.chat.ViewStats(); stats != "" {
		sections = append(sections, stats)
	}
	if errView := a.chat.ViewError(); errView != "" {
		sections = append(sections, errView)
	}
	sections = append(sections, a.statusBar.View())
	background := strings.Join(sections, "\n")

	// Dim the background
	backgroundLines := strings.Split(background, "\n")
	for i, line := range backgroundLines {
		backgroundLines[i] = lipgloss.NewStyle().Foreground(a.theme.TextMuted).Render(line)
	}

	// The browse modal handles its own centering
	modal := a.browseView.View()
	modalLines := strings.Split(modal, "\n")

	// Overlay the modal on the dimmed background
	result := make([]string, len(backgroundLines))
	copy(result, backgroundLines)

	// Overlay modal lines onto background
	for i, line := range modalLines {
		if i < len(result) && strings.TrimSpace(line) != "" {
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

func (a *App) renderPairLayout() string {
	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	if a.width >= 100 {
		// Split pane: dimmed chat left, pair right
		chatWidth := a.width - a.pairWidth() - 1
		chatHeight := a.pairHeight()

		chatContent := a.chat.ViewChat()
		dimmedChat := lipgloss.NewStyle().
			Width(chatWidth).
			Height(chatHeight).
			Foreground(a.theme.TextMuted).
			Render(chatContent)

		pairPanel := a.pairView.View()

		sep := lipgloss.NewStyle().
			Foreground(a.theme.Border).
			Render("│")

		split := lipgloss.JoinHorizontal(lipgloss.Top, dimmedChat, sep, pairPanel)
		sections = append(sections, split)
	} else {
		// Full width pair
		sections = append(sections, a.pairView.View())
	}

	// Status bar
	sections = append(sections, a.statusBar.View())

	return strings.Join(sections, "\n")
}

func (a *App) renderEditLayout() string {
	var sections []string
	sections = append(sections, a.editorView.View())
	sections = append(sections, a.statusBar.View())
	return strings.Join(sections, "\n")
}

func (a *App) renderHeader() string {
	logo := lipgloss.NewStyle().Foreground(a.theme.Primary).Bold(true).Render("🔥🗝️🔥 Hecate")

	modelSection := ""
	if modelName := a.chat.ActiveModelName(); modelName != "" {
		modelSection = a.styles.Subtle.Render("  ·  " + modelName)
	}

	daemonSection := "  ·  "
	switch a.daemonStatus {
	case "healthy", "ok":
		daemonSection += a.styles.StatusOK.Render("●") + a.styles.Subtle.Render(" daemon")
	case "degraded":
		daemonSection += a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon")
	default:
		daemonSection += a.styles.Subtle.Render("○ daemon")
	}

	titleSection := ""
	if a.conversationTitle != "" {
		titleSection = a.styles.Subtle.Render("  ·  ") + a.styles.CardValue.Render(a.conversationTitle)
	}

	left := logo + modelSection + daemonSection + titleSection

	return lipgloss.NewStyle().Width(a.width).Padding(0, 1).Render(left)
}

func (a *App) renderCommandLine() string {
	return lipgloss.NewStyle().
		Width(a.width).
		Padding(0, 1).
		Background(a.theme.BgInput).
		Render(a.cmdInput.View())
}

func (a *App) checkHealth() tea.Msg {
	health, err := a.client.GetHealth()
	if err != nil {
		return healthMsg{status: "error"}
	}
	return healthMsg{status: health.Status}
}

func (a *App) saveConversation() {
	msgs := a.chat.Messages()
	if len(msgs) == 0 {
		return
	}

	var convMsgs []config.ConversationMsg
	for _, m := range msgs {
		if m.Role == "system" {
			continue // Don't persist command output
		}
		convMsgs = append(convMsgs, config.ConversationMsg{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}

	if len(convMsgs) == 0 {
		return
	}

	title := config.TitleFromMessages(convMsgs)
	a.conversationTitle = title

	conv := config.Conversation{
		ID:        a.conversationID,
		Title:     title,
		Model:     a.chat.ActiveModelName(),
		Messages:  convMsgs,
		CreatedAt: convMsgs[0].Time,
	}

	_ = config.SaveConversation(conv)
}

func (a *App) startNewConversation() {
	a.saveConversation()
	a.chat.ClearMessages()
	a.conversationID = config.NewConversationID()
	a.conversationTitle = ""
}

func (a *App) loadConversation(id string) error {
	conv, err := config.LoadConversation(id)
	if err != nil {
		return err
	}

	a.saveConversation() // save current first

	var msgs []chat.Message
	for _, m := range conv.Messages {
		msgs = append(msgs, chat.Message{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}

	a.chat.ClearMessages()
	a.chat.LoadMessages(msgs)
	a.conversationID = conv.ID
	a.conversationTitle = conv.Title
	return nil
}

func (a *App) yankLastResponse() tea.Cmd {
	content := a.chat.LastAssistantMessage()
	if content == "" {
		a.chat.InjectSystemMessage("No response to copy.")
		return nil
	}

	if err := clipboard.WriteAll(content); err != nil {
		a.chat.InjectSystemMessage("Clipboard unavailable: " + err.Error())
		return nil
	}

	// Truncate preview
	preview := content
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	a.chat.InjectSystemMessage("Copied to clipboard: " + preview)
	return nil
}
