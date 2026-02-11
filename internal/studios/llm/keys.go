package llm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/editor"
	"github.com/hecate-social/hecate-tui/internal/modes"

	"github.com/atotto/clipboard"
)

// handleKey dispatches keys based on the studio's current mode.
func (s *Studio) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch s.mode {
	case modes.Normal:
		return s.handleNormalKey(key)
	case modes.Insert:
		return s.handleInsertKey(key)
	case modes.Browse:
		return s.handleBrowseKey(key, msg)
	case modes.Pair:
		return s.handlePairKey(key, msg)
	case modes.Edit:
		return s.handleEditKey(key, msg)
	case modes.Form:
		return s.handleFormKey(key, msg)
	default:
		if key == "esc" {
			s.setMode(modes.Normal)
		}
		return nil
	}
}

func (s *Studio) handleNormalKey(key string) tea.Cmd {
	// Handle tool approval keys when pending
	if s.chat.HasPendingApproval() {
		switch key {
		case "y":
			return s.chat.ApproveToolCall(false)
		case "n":
			return s.chat.DenyToolCall()
		case "a":
			return s.chat.ApproveToolCall(true)
		case "esc":
			return s.chat.DenyToolCall()
		}
		return nil
	}

	switch key {
	case "i":
		s.setMode(modes.Insert)
	case "j", "down":
		s.chat.ScrollDown(1)
	case "k", "up":
		s.chat.ScrollUp(1)
	case "ctrl+d":
		s.chat.HalfPageDown()
	case "ctrl+u":
		s.chat.HalfPageUp()
	case "g":
		s.chat.GotoTop()
	case "G":
		s.chat.GotoBottom()
	case "?":
		ctx := s.CommandContext()
		return commands.ModeHelp(int(s.mode), ctx)
	case "t":
		s.chat.ToggleThinking()
	case "r":
		return s.chat.RetryLast()
	case "y":
		return yankLastResponse(s)
	}
	return nil
}

func (s *Studio) handleInsertKey(key string) tea.Cmd {
	// Handle tool approval keys when pending
	if s.chat.HasPendingApproval() {
		switch key {
		case "y":
			return s.chat.ApproveToolCall(false)
		case "n":
			return s.chat.DenyToolCall()
		case "a":
			return s.chat.ApproveToolCall(true)
		case "esc":
			return s.chat.DenyToolCall()
		}
		return nil
	}

	switch key {
	case "esc":
		if s.chat.IsStreaming() {
			s.chat.CancelStreaming()
			return nil
		}
		s.msgHistIdx = -1
		s.setMode(modes.Normal)
	case "enter":
		content := s.chat.InputValue()
		if content != "" {
			s.msgHistory = append(s.msgHistory, content)
		}
		s.msgHistIdx = -1
		s.msgDraft = ""
		cmd := s.chat.SendCurrentInput()
		if cmd != nil {
			s.chat.ClearError()
			s.saveConversation()
		}
		return cmd
	case "alt+enter":
		s.chat.InsertNewline()
	case "tab":
		s.chat.CycleModel()
	case "shift+tab":
		s.chat.CycleModelReverse()
	case "up":
		if len(s.msgHistory) == 0 {
			return nil
		}
		if s.msgHistIdx == -1 {
			s.msgDraft = s.chat.InputValue()
			s.msgHistIdx = len(s.msgHistory) - 1
		} else if s.msgHistIdx > 0 {
			s.msgHistIdx--
		}
		s.chat.SetInputValue(s.msgHistory[s.msgHistIdx])
	case "down":
		if s.msgHistIdx == -1 {
			return nil
		}
		if s.msgHistIdx < len(s.msgHistory)-1 {
			s.msgHistIdx++
			s.chat.SetInputValue(s.msgHistory[s.msgHistIdx])
		} else {
			s.msgHistIdx = -1
			s.chat.SetInputValue(s.msgDraft)
		}
	}
	return nil
}

func (s *Studio) handleBrowseKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !s.browseReady {
		return nil
	}

	if key == "?" {
		ctx := s.CommandContext()
		return commands.ModeHelp(int(s.mode), ctx)
	}

	if key == "esc" {
		s.setMode(modes.Normal)
		return nil
	}

	consumed, cmd := s.browseView.HandleKey(key, msg)
	if consumed {
		return cmd
	}

	return nil
}

func (s *Studio) handlePairKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !s.pairReady {
		return nil
	}

	if key == "?" {
		ctx := s.CommandContext()
		return commands.ModeHelp(int(s.mode), ctx)
	}

	consumed, cmd := s.pairView.HandleKey(key, msg)
	if consumed {
		return cmd
	}

	if key == "esc" {
		s.setMode(modes.Normal)
		return nil
	}

	return nil
}

func (s *Studio) handleEditKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !s.editorReady {
		return nil
	}

	switch key {
	case "ctrl+q", "esc":
		s.editorReady = false
		s.setMode(modes.Normal)
		return nil
	}

	var cmd tea.Cmd
	var model tea.Model
	model, cmd = s.editorView.Update(msg)
	s.editorView = model.(editor.Model)

	return cmd
}

func (s *Studio) handleFormKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !s.formReady || s.formView == nil {
		return nil
	}

	var cmd tea.Cmd
	s.formView, cmd = s.formView.Update(msg)
	return cmd
}

func yankLastResponse(s *Studio) tea.Cmd {
	content := s.chat.LastAssistantMessage()
	if content == "" {
		s.chat.InjectSystemMessage("No response to copy.")
		return nil
	}

	if err := clipboard.WriteAll(content); err != nil {
		s.chat.InjectSystemMessage("Clipboard unavailable: " + err.Error())
		return nil
	}

	preview := content
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	s.chat.InjectSystemMessage("Copied to clipboard: " + preview)
	return nil
}
