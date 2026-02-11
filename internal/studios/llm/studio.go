// Package llm implements the LLM Studio — the chat-with-AI workspace.
// This is the primary studio and wraps the existing chat, browse, pair,
// editor, form, and tool subsystems.
package llm

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/browse"
	"github.com/hecate-social/hecate-tui/internal/chat"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/config"
	"github.com/hecate-social/hecate-tui/internal/editor"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/pair"
	"github.com/hecate-social/hecate-tui/internal/studio"
	"github.com/hecate-social/hecate-tui/internal/theme"
	"github.com/hecate-social/hecate-tui/internal/ui"
)

// Studio is the LLM chat workspace.
type Studio struct {
	ctx *studio.Context

	// Mode (LLM studio manages Normal/Insert and overlay modes internally)
	mode     modes.Mode
	prevMode modes.Mode

	// Components
	chat       chat.Model
	browseView browse.Model
	pairView   pair.Model
	editorView editor.Model
	formView   *ui.FormModel

	// Tool system
	toolExecutor   *llmtools.Executor
	approvalPrompt *ui.ApprovalPrompt

	// Overlay states
	browseReady bool
	pairReady   bool
	editorReady bool
	formReady   bool

	// Chat input history
	msgHistory []string
	msgHistIdx int
	msgDraft   string

	// System prompt / personality
	systemPrompt string

	// Conversation
	conversationID    string
	conversationTitle string

	// ALC context
	alcState *alc.State

	// Dimensions
	width   int
	height  int
	focused bool

	// Cached config for save operations
	cfg config.Config
}

// txFlashDoneMsg resets the TX LED after flash duration.
type txFlashDoneMsg struct{}

// New creates a new LLM Studio.
func New(ctx *studio.Context) *Studio {
	chatModel := chat.New(ctx.Client, ctx.Theme, ctx.Styles)

	systemPrompt := ctx.Config.BuildSystemPrompt()
	if systemPrompt != "" {
		chatModel.SetSystemPrompt(systemPrompt)
	}

	if ctx.Config.Model != "" {
		chatModel.SetPreferredModel(ctx.Config.Model)
	}

	toolRegistry := llmtools.NewDefaultRegistry()
	toolPermissions := llmtools.NewPermissions()
	toolExecutor := llmtools.NewExecutor(toolRegistry, toolPermissions)
	chatModel.SetToolExecutor(toolExecutor)
	chatModel.EnableTools(false)
	llmtools.SetMeshClient(ctx.Client)

	approvalPrompt := ui.NewApprovalPrompt(ctx.Theme, ctx.Styles)

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

	return &Studio{
		ctx:               ctx,
		mode:              modes.Normal,
		chat:              chatModel,
		systemPrompt:      systemPrompt,
		toolExecutor:      toolExecutor,
		approvalPrompt:    approvalPrompt,
		alcState:          alc.NewState(),
		conversationID:    convID,
		conversationTitle: convTitle,
		cfg:               ctx.Config,
	}
}

func (s *Studio) Name() string      { return "LLM" }
func (s *Studio) ShortName() string { return "LLM" }
func (s *Studio) Icon() string      { return "🤖" }

func (s *Studio) Init() tea.Cmd {
	return tea.Batch(
		s.chat.Init(),
		s.detectVenture,
	)
}

func (s *Studio) Mode() modes.Mode { return s.mode }
func (s *Studio) Hints() string    { return s.mode.Hints() }
func (s *Studio) Focused() bool    { return s.focused }

func (s *Studio) SetFocused(focused bool) {
	s.focused = focused
}

func (s *Studio) SetSize(width, height int) {
	s.width = width
	s.height = height
	chatHeight := s.chatAreaHeight()
	s.chat.SetSize(width, chatHeight)
	if s.browseReady {
		s.browseView.SetSize(width, height)
	}
	if s.pairReady {
		s.pairView.SetSize(s.pairWidth(), s.pairHeight())
	}
	if s.editorReady {
		s.editorView.SetSize(width, s.editorHeight())
	}
}

func (s *Studio) StatusInfo() studio.StatusInfo {
	return studio.StatusInfo{
		ModelName:     s.chat.ActiveModelName(),
		ModelProvider: s.chat.ActiveModelProvider(),
		ModelStatus:   s.modelStatus(),
		ModelError:    s.modelError(),
		InputLen:      s.chat.InputLen(),
		SessionTokens: s.chat.SessionTokenCount(),
	}
}

func (s *Studio) modelStatus() string {
	if s.chat.IsStreaming() {
		return "loading"
	}
	if s.chat.HasError() {
		return "error"
	}
	return "ready"
}

func (s *Studio) modelError() string {
	if s.chat.HasError() {
		return s.chat.LastError()
	}
	return ""
}

// Commands returns LLM-specific slash commands.
func (s *Studio) Commands() []commands.Command {
	return nil // Commands stay in global registry for now — migrated in future phase
}

// ALCState returns the ALC state for the shell to read.
func (s *Studio) ALCState() *alc.State {
	return s.alcState
}

// Update handles messages routed from the shell.
func (s *Studio) Update(msg tea.Msg) (studio.Studio, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			s.chat.ScrollUp(3)
		case tea.MouseButtonWheelDown:
			s.chat.ScrollDown(3)
		}

	case tea.KeyMsg:
		modeBefore := s.mode
		cmd := s.handleKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// If we just switched INTO Insert mode, don't forward the key to chat
		if modeBefore != modes.Insert && s.mode == modes.Insert {
			return s, tea.Batch(cmds...)
		}
		// Also skip forwarding if we switched OUT of Insert mode
		if modeBefore == modes.Insert && s.mode != modes.Insert {
			return s, tea.Batch(cmds...)
		}

	// Command system messages that affect LLM studio
	case commands.ClearChatMsg:
		s.chat.ClearMessages()

	case commands.SwitchModelMsg:
		s.chat.SwitchModel(msg.Name)
		s.cfg.Model = msg.Name
		_ = s.cfg.Save()

	case commands.SetModeMsg:
		cmd := s.enterMode(modes.Mode(msg.Mode))
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.EditFileMsg:
		cmd := s.openEditor(msg.Path)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.NewConversationMsg:
		s.startNewConversation()
		s.chat.InjectSystemMessage("Started new conversation.")

	case commands.LoadConversationMsg:
		if err := s.loadConversation(msg.ID); err != nil {
			s.chat.InjectSystemMessage("Failed to load: " + err.Error())
		} else {
			s.chat.InjectSystemMessage("Loaded conversation: " + msg.ID)
		}

	case commands.EnableToolsMsg:
		s.chat.EnableTools(msg.Enabled)
		status := "disabled"
		if msg.Enabled {
			status = "enabled"
		}
		s.chat.InjectSystemMessage("LLM function calling " + status)

	case browse.SelectModelMsg:
		s.chat.SwitchModel(msg.ModelName)
		s.setMode(modes.Normal)
		s.chat.InjectSystemMessage("Model switched to: " + msg.ModelName)
		s.cfg.Model = msg.ModelName
		_ = s.cfg.Save()

	case commands.SwitchRoleMsg:
		s.cfg.Personality.ActiveRole = msg.Role
		_ = s.cfg.Save()
		newPrompt := s.cfg.BuildSystemPrompt()
		s.systemPrompt = newPrompt
		s.chat.SetSystemPrompt(newPrompt)
		roleName := s.cfg.ActiveRoleDisplayName()
		if roleName == "" {
			roleName = msg.Role
		}
		s.chat.InjectSystemMessage("Role switched to: " + roleName)

	case commands.SetALCContextMsg:
		s.handleALCContextChange(msg)

	case commands.ShowFormMsg:
		cmd := s.showForm(msg.FormType)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.ChangeDirMsg:
		if err := os.Chdir(msg.Path); err != nil {
			s.chat.InjectSystemMessage(s.ctx.Styles.Error.Render("Failed to change directory: " + err.Error()))
		} else {
			s.chat.InjectSystemMessage(s.ctx.Styles.Subtle.Render("Changed to: " + msg.Path))
			cmds = append(cmds, s.detectVenture)
		}

	case commands.VentureCreatedMsg:
		s.chat.InjectSystemMessage(msg.Message)
		if err := os.Chdir(msg.Path); err != nil {
			s.chat.InjectSystemMessage(s.ctx.Styles.Error.Render("Failed to cd to new venture: " + err.Error()))
		} else {
			cmds = append(cmds, s.detectVenture)
		}

	case ui.FormResult:
		cmd := s.handleFormResult(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case commands.InjectSystemMsg:
		s.chat.InjectSystemMessage(msg.Content)

	case ventureDetectedMsg:
		if msg.venture != nil {
			s.alcState.SetVenture(msg.venture, msg.source)
			s.chat.InjectSystemMessage("Resuming venture: " + msg.venture.Name + " (detected from " + msg.source + ")")
		}

	case txFlashDoneMsg:
		// handled by shell
	}

	// Forward to chat for streaming updates
	wasStreaming := s.chat.IsStreaming()
	var chatCmd tea.Cmd
	s.chat, chatCmd = s.chat.Update(msg)
	cmds = append(cmds, chatCmd)

	// Auto-save on streaming completion
	nowStreaming := s.chat.IsStreaming()
	if wasStreaming && !nowStreaming {
		s.saveConversation()
	}

	// Forward to browse if in Browse mode
	if s.mode == modes.Browse && s.browseReady {
		var browseCmd tea.Cmd
		s.browseView, browseCmd = s.browseView.Update(msg)
		cmds = append(cmds, browseCmd)
	}

	// Forward to pair if in Pair mode
	if s.mode == modes.Pair && s.pairReady {
		var pairCmd tea.Cmd
		s.pairView, pairCmd = s.pairView.Update(msg)
		cmds = append(cmds, pairCmd)
	}

	// Forward to editor if in Edit mode (non-key msgs)
	if s.mode == modes.Edit && s.editorReady {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			var edCmd tea.Cmd
			var model tea.Model
			model, edCmd = s.editorView.Update(msg)
			s.editorView = model.(editor.Model)
			cmds = append(cmds, edCmd)
		}
	}

	// Forward to form if in Form mode (non-key msgs)
	if s.mode == modes.Form && s.formReady && s.formView != nil {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			var formCmd tea.Cmd
			s.formView, formCmd = s.formView.Update(msg)
			cmds = append(cmds, formCmd)
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *Studio) setMode(m modes.Mode) {
	if m == s.mode {
		return
	}
	s.prevMode = s.mode
	s.mode = m

	switch m {
	case modes.Normal:
		s.chat.SetInputVisible(false)
	case modes.Insert:
		s.chat.SetInputVisible(true)
	case modes.Browse, modes.Pair, modes.Edit, modes.Form:
		s.chat.SetInputVisible(false)
	}

	chatHeight := s.chatAreaHeight()
	s.chat.SetSize(s.width, chatHeight)
}

func (s *Studio) enterMode(m modes.Mode) tea.Cmd {
	s.setMode(m)

	switch m {
	case modes.Browse:
		s.browseView = browse.New(s.ctx.Client, s.ctx.Theme, s.ctx.Styles)
		s.browseView.SetSize(s.width, s.height)
		s.browseReady = true
		return s.browseView.Init()
	case modes.Pair:
		s.pairView = pair.New(s.ctx.Client, s.ctx.Theme, s.ctx.Styles)
		s.pairView.SetSize(s.pairWidth(), s.pairHeight())
		s.pairReady = true
		return s.pairView.Init()
	}

	return nil
}

// InjectSystemMessage injects a system message into chat.
func (s *Studio) InjectSystemMessage(content string) {
	s.chat.InjectSystemMessage(content)
}

// Chat returns the chat model for the shell to read streaming state.
func (s *Studio) Chat() *chat.Model {
	return &s.chat
}

// handleALCContextChange processes ALC context switch messages.
func (s *Studio) handleALCContextChange(msg commands.SetALCContextMsg) {
	switch msg.Context {
	case alc.Chat:
		s.alcState.ClearVenture()
		s.chat.InjectSystemMessage("Returned to chat mode.")

	case alc.Venture:
		if msg.Venture != nil {
			source := msg.Source
			if source == "" {
				source = "manual"
			}
			s.alcState.SetVenture(msg.Venture, source)
			s.chat.InjectSystemMessage("Venture selected: " + msg.Venture.Name)
		}

	case alc.Department:
		if msg.Department != nil {
			s.alcState.SetDepartment(msg.Department)
			s.chat.InjectSystemMessage("Department active: " + msg.Department.Name + " (" + string(msg.Department.CurrentPhase) + ")")
		}
	}
}

// CommandContext builds a commands.Context for command dispatch.
func (s *Studio) CommandContext() *commands.Context {
	return &commands.Context{
		Client: s.ctx.Client,
		Theme:  s.ctx.Theme,
		Styles: s.ctx.Styles,
		Width:  s.width,
		Height: s.height,
		SetMode: func(mode int) {
			s.setMode(modes.Mode(mode))
		},
		InjectChat: func(msg commands.ChatMessage) {
			s.chat.InjectSystemMessage(msg.Content)
		},
		GetMessages: func() []commands.ChatExportMsg {
			exported := s.chat.ExportMessages()
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
			return s.systemPrompt
		},
		SetSystemPrompt: func(prompt string) {
			s.systemPrompt = prompt
			s.chat.SetSystemPrompt(prompt)
			s.cfg.SystemPrompt = prompt
			_ = s.cfg.Save()
		},
		GetToolExecutor: func() *llmtools.Executor {
			return s.chat.ToolExecutor()
		},
		ToolsEnabled: func() bool {
			return s.chat.ToolsEnabled()
		},
		GetActiveRole: func() string {
			return s.cfg.Personality.ActiveRole
		},
		SetActiveRole: func(role string) error {
			s.cfg.Personality.ActiveRole = role
			return s.cfg.Save()
		},
		GetRoleNames: func() []string {
			return []string{"dna", "anp", "tni", "dno"}
		},
		RebuildPrompt: func() string {
			return s.cfg.BuildSystemPrompt()
		},
		GetALCContext: func() *alc.State {
			return s.alcState
		},
	}
}

// conversation management

func (s *Studio) saveConversation() {
	msgs := s.chat.Messages()
	if len(msgs) == 0 {
		return
	}

	var convMsgs []config.ConversationMsg
	for _, m := range msgs {
		if m.Role == "system" {
			continue
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
	s.conversationTitle = title

	conv := config.Conversation{
		ID:        s.conversationID,
		Title:     title,
		Model:     s.chat.ActiveModelName(),
		Messages:  convMsgs,
		CreatedAt: convMsgs[0].Time,
	}

	_ = config.SaveConversation(conv)
}

func (s *Studio) startNewConversation() {
	s.saveConversation()
	s.chat.ClearMessages()
	s.conversationID = config.NewConversationID()
	s.conversationTitle = ""
}

func (s *Studio) loadConversation(id string) error {
	conv, err := config.LoadConversation(id)
	if err != nil {
		return err
	}

	s.saveConversation()

	var msgs []chat.Message
	for _, m := range conv.Messages {
		msgs = append(msgs, chat.Message{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}

	s.chat.ClearMessages()
	s.chat.LoadMessages(msgs)
	s.conversationID = conv.ID
	s.conversationTitle = conv.Title
	return nil
}

// venture detection

type ventureDetectedMsg struct {
	venture *alc.VentureInfo
	source  string
}

func (s *Studio) detectVenture() tea.Msg {
	result := alc.DetectVenture()
	if !result.Found {
		return nil
	}

	if result.Source == "config" && result.Config != nil && result.Config.VentureID != "" {
		venture, err := s.ctx.Client.GetVentureByID(result.Config.VentureID)
		if err == nil && venture != nil {
			return ventureDetectedMsg{
				venture: &alc.VentureInfo{
					ID:    venture.VentureID,
					Name:  venture.Name,
					Brief: venture.Brief,
				},
				source: "config",
			}
		}
		return ventureDetectedMsg{
			venture: &alc.VentureInfo{
				ID:    result.Config.VentureID,
				Name:  result.Config.Name,
				Brief: result.Config.Brief,
			},
			source: "config",
		}
	}

	if result.Source == "git" && result.Config != nil {
		ventures, err := s.ctx.Client.ListVentures()
		if err == nil {
			for _, v := range ventures {
				if containsIgnoreCase(result.Config.Name, v.Name) {
					return ventureDetectedMsg{
						venture: &alc.VentureInfo{
							ID:    v.VentureID,
							Name:  v.Name,
							Brief: v.Brief,
						},
						source: "git",
					}
				}
			}
		}
	}

	return nil
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 &&
		containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// forms

func (s *Studio) showForm(formType string) tea.Cmd {
	switch formType {
	case "venture_init":
		cwd, _ := os.Getwd()
		s.formView = ui.NewVentureForm(s.ctx.Theme, s.ctx.Styles, cwd)
		formWidth := 60
		if s.width > 0 && s.width < 70 {
			formWidth = s.width - 4
		}
		s.formView.SetWidth(formWidth)
		s.formReady = true
		s.setMode(modes.Form)
		return s.formView.Init()
	default:
		s.chat.InjectSystemMessage("Unknown form type: " + formType)
		return nil
	}
}

func (s *Studio) handleFormResult(result ui.FormResult) tea.Cmd {
	s.formReady = false
	s.setMode(modes.Normal)

	if !result.Submitted {
		s.chat.InjectSystemMessage("Cancelled.")
		return nil
	}

	switch result.FormID {
	case "venture_init":
		return s.handleVentureFormResult(result)
	default:
		s.chat.InjectSystemMessage("Unknown form: " + result.FormID)
		return nil
	}
}

func (s *Studio) handleVentureFormResult(result ui.FormResult) tea.Cmd {
	pathInput := result.Values["path"]
	name := result.Values["name"]
	brief := result.Values["brief"]

	if len(pathInput) == 0 || isBlank(pathInput) {
		s.chat.InjectSystemMessage(s.ctx.Styles.Error.Render("Path is required"))
		return nil
	}

	cwd, _ := os.Getwd()
	path := ui.ExpandPath(pathInput, cwd)

	if isBlank(name) {
		name = ui.InferName(path)
	}

	return s.createVentureFromForm(path, name, brief)
}

func isBlank(s string) bool {
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

func (s *Studio) createVentureFromForm(path, name, brief string) tea.Cmd {
	return func() tea.Msg {
		st := s.ctx.Styles

		if err := os.MkdirAll(path, 0755); err != nil {
			return commands.InjectSystemMsg{Content: st.Error.Render("Failed to create directory: " + err.Error())}
		}

		venture, err := s.ctx.Client.InitiateVenture(name, brief)
		if err != nil {
			return commands.InjectSystemMsg{Content: st.Error.Render("Failed to initiate venture: " + err.Error())}
		}

		return buildVentureScaffoldMsg(st, venture.VentureID, venture.Name, venture.Brief,
			venture.InitiatedAt, venture.InitiatedBy, path)
	}
}

func (s *Studio) openEditor(path string) tea.Cmd {
	if path != "" {
		ed, err := editor.NewWithFile(path)
		if err != nil {
			s.chat.InjectSystemMessage("Could not open file: " + err.Error())
			return nil
		}
		s.editorView = ed
	} else {
		s.editorView = editor.New()
	}

	s.editorView.SetSize(s.width, s.editorHeight())
	s.editorView.Focus()
	s.editorReady = true
	s.setMode(modes.Edit)
	return s.editorView.Init()
}

// SwitchTheme updates the studio's components for a new theme.
func (s *Studio) SwitchTheme(t *theme.Theme, styles *theme.Styles) {
	s.ctx.Theme = t
	s.ctx.Styles = styles
	s.chat = chat.New(s.ctx.Client, t, styles)
	s.chat.SetSize(s.width, s.chatAreaHeight())
	s.approvalPrompt = ui.NewApprovalPrompt(t, styles)
	s.chat.InjectSystemMessage("Theme switched to: " + t.Name)
}

// IsStreaming returns whether the chat is currently streaming a response.
func (s *Studio) IsStreaming() bool {
	return s.chat.IsStreaming()
}

// ScheduleTxFlash returns a command to flash the TX LED.
func (s *Studio) ScheduleTxFlash() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return txFlashDoneMsg{}
	})
}

// YankLastResponse copies the last assistant message to clipboard.
func (s *Studio) YankLastResponse() tea.Cmd {
	return yankLastResponse(s)
}
