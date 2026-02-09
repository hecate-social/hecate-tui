package app

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/alc"
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
	formView   *ui.FormModel

	// Tool system
	toolExecutor   *llmtools.Executor
	approvalPrompt *ui.ApprovalPrompt

	// Overlay modes initialized
	browseReady bool
	pairReady   bool
	editorReady bool
	formReady   bool

	// Command history
	cmdHistory []string
	cmdHistIdx int    // -1 = new input, 0..N = browsing history
	cmdDraft   string // save draft when browsing history

	// Chat input history (for Insert mode up/down arrow)
	msgHistory []string
	msgHistIdx int    // -1 = new input, 0..N = browsing history
	msgDraft   string // save draft when browsing history

	// System prompt for LLM
	systemPrompt string

	// Persistent config
	cfg               config.Config
	conversationID    string
	conversationTitle string

	// Health polling
	daemonStatus string

	// ALC context (Chat/Torch/Cartwheel)
	alcState *alc.State
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

	chatModel := chat.New(c, t, s)

	systemPrompt := cfg.BuildSystemPrompt()
	if systemPrompt != "" {
		chatModel.SetSystemPrompt(systemPrompt)
	}

	if cfg.Model != "" {
		chatModel.SetPreferredModel(cfg.Model)
	}

	toolRegistry := llmtools.NewDefaultRegistry()
	toolPermissions := llmtools.NewPermissions()
	toolExecutor := llmtools.NewExecutor(toolRegistry, toolPermissions)
	chatModel.SetToolExecutor(toolExecutor)
	chatModel.EnableTools(false)
	llmtools.SetMeshClient(c)

	approvalPrompt := ui.NewApprovalPrompt(t, s)

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
		alcState:          alc.NewState(),
	}
}

// Init starts the app — fetch models, check health, start polling, detect torch.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.chat.Init(),
		a.checkHealth,
		a.scheduleHealthTick(),
		a.detectTorch,
	)
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
		// Track mode before handling key to detect mode-switching keys
		modeBefore := a.mode
		cmd := a.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()
		a.statusBar.Mode = a.mode
		a.statusBar.InputLen = a.chat.InputLen()
		a.statusBar.SessionTokens = a.chat.SessionTokenCount()
		// If we just switched INTO Insert mode, don't forward the key to chat
		// (prevents 'i' from appearing in the input box when entering Insert mode)
		if modeBefore != modes.Insert && a.mode == modes.Insert {
			return a, tea.Batch(cmds...)
		}
		// Also skip forwarding if we switched OUT of Insert mode (esc key)
		if modeBefore == modes.Insert && a.mode != modes.Insert {
			return a, tea.Batch(cmds...)
		}

	case healthMsg:
		a.daemonStatus = msg.status
		a.statusBar.DaemonStatus = msg.status
		// Sync status bar model from chat state (don't override with first model from list)
		a.statusBar.ModelName = a.chat.ActiveModelName()
		a.statusBar.ModelProvider = a.chat.ActiveModelProvider()

	case torchDetectedMsg:
		if msg.torch != nil {
			a.alcState.SetTorch(msg.torch, msg.source)
			a.statusBar.TorchName = msg.torch.Name
			a.chat.InjectSystemMessage("Resuming torch: " + msg.torch.Name + " (detected from " + msg.source + ")")
		}

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
		if err := a.cfg.Save(); err != nil {
			a.chat.InjectSystemMessage("Warning: failed to save config: " + err.Error())
		}

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
		if err := a.cfg.Save(); err != nil {
			a.chat.InjectSystemMessage("Warning: failed to save config: " + err.Error())
		}

	case commands.SwitchRoleMsg:
		a.cfg.Personality.ActiveRole = msg.Role
		if err := a.cfg.Save(); err != nil {
			a.chat.InjectSystemMessage("Warning: failed to save config: " + err.Error())
		}
		// Rebuild and apply system prompt
		newPrompt := a.cfg.BuildSystemPrompt()
		a.systemPrompt = newPrompt
		a.chat.SetSystemPrompt(newPrompt)
		roleName := a.cfg.ActiveRoleDisplayName()
		if roleName == "" {
			roleName = msg.Role
		}
		a.chat.InjectSystemMessage("Role switched to: " + roleName)

	case commands.SetALCContextMsg:
		a.handleALCContextChange(msg)

	case commands.ShowFormMsg:
		a.chat.InjectSystemMessage("DEBUG: ShowFormMsg received, type=" + msg.FormType)
		cmd := a.showForm(msg.FormType)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.ChangeDirMsg:
		if err := os.Chdir(msg.Path); err != nil {
			a.chat.InjectSystemMessage(a.styles.Error.Render("Failed to change directory: " + err.Error()))
		} else {
			a.statusBar.Cwd = msg.Path
			a.chat.InjectSystemMessage(a.styles.Subtle.Render("Changed to: " + msg.Path))
			// Re-detect torch in new directory
			cmds = append(cmds, a.detectTorch)
		}

	case commands.TorchCreatedMsg:
		// Show the creation message
		a.chat.InjectSystemMessage(msg.Message)
		// cd into the new torch directory
		if err := os.Chdir(msg.Path); err != nil {
			a.chat.InjectSystemMessage(a.styles.Error.Render("Failed to cd to new torch: " + err.Error()))
		} else {
			a.statusBar.Cwd = msg.Path
			// Re-detect torch (will load the newly created one)
			cmds = append(cmds, a.detectTorch)
		}

	case ui.FormResult:
		cmd := a.handleFormResult(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Forward to chat for streaming updates
	wasStreaming := a.chat.IsStreaming()
	var chatCmd tea.Cmd
	a.chat, chatCmd = a.chat.Update(msg)
	cmds = append(cmds, chatCmd)

	// Update model status LED based on streaming state
	if a.chat.IsStreaming() {
		a.statusBar.ModelStatus = "loading"
		a.statusBar.ModelError = ""
	} else if wasStreaming && !a.chat.IsStreaming() {
		// Streaming just completed
		if a.chat.HasError() {
			a.statusBar.ModelStatus = "error"
			a.statusBar.ModelError = a.chat.LastError()
		} else {
			a.statusBar.ModelStatus = "ready"
			a.statusBar.ModelError = ""
		}
		// Auto-save conversation
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

	// Forward to form if in Form mode (non-key msgs like blink)
	if a.mode == modes.Form && a.formReady && a.formView != nil {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			var formCmd tea.Cmd
			a.formView, formCmd = a.formView.Update(msg)
			cmds = append(cmds, formCmd)
		}
	}

	return a, tea.Batch(cmds...)
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
	case modes.Form:
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
	a.chat = chat.New(a.client, t, a.styles)
	a.chat.SetSize(a.width, a.chatAreaHeight())

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
			if err := a.cfg.Save(); err != nil {
				a.chat.InjectSystemMessage("Warning: failed to save config: " + err.Error())
			}
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
			if err := a.cfg.Save(); err != nil {
				a.chat.InjectSystemMessage("Warning: failed to save config: " + err.Error())
			}
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
		GetALCContext: func() *alc.State {
			return a.alcState
		},
	}
}

// handleALCContextChange processes ALC context switch messages.
func (a *App) handleALCContextChange(msg commands.SetALCContextMsg) {
	switch msg.Context {
	case alc.Chat:
		a.alcState.ClearTorch()
		a.statusBar.TorchName = ""
		a.statusBar.ActivePhase = ""
		a.chat.InjectSystemMessage("Returned to chat mode.")

	case alc.Torch:
		if msg.Torch != nil {
			source := msg.Source
			if source == "" {
				source = "manual"
			}
			a.alcState.SetTorch(msg.Torch, source)
			a.statusBar.TorchName = msg.Torch.Name
			a.statusBar.ActivePhase = ""
			a.chat.InjectSystemMessage("Torch selected: " + msg.Torch.Name)
		}

	case alc.Cartwheel:
		if msg.Cartwheel != nil {
			a.alcState.SetCartwheel(msg.Cartwheel)
			a.statusBar.ActivePhase = string(msg.Cartwheel.CurrentPhase)
			a.chat.InjectSystemMessage("Cartwheel active: " + msg.Cartwheel.Name + " (" + string(msg.Cartwheel.CurrentPhase) + ")")
		}
	}
}
