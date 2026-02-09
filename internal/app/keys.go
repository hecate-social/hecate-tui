package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/commands"
	"github.com/hecate-social/hecate-tui/internal/editor"
	"github.com/hecate-social/hecate-tui/internal/modes"
)

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
	case modes.Form:
		return a.handleFormKey(key, msg)
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
	case "t":
		a.chat.ToggleThinking()
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
		a.msgHistIdx = -1
		a.setMode(modes.Normal)
	case "enter":
		// Save message to history before sending
		content := a.chat.InputValue()
		if content != "" {
			a.msgHistory = append(a.msgHistory, content)
		}
		a.msgHistIdx = -1
		a.msgDraft = ""
		cmd := a.chat.SendCurrentInput()
		if cmd != nil {
			// Set model status to loading and clear any previous error
			a.statusBar.ModelStatus = "loading"
			a.statusBar.ModelError = ""
			a.chat.ClearError()
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
	case "up":
		// Navigate back in message history
		if len(a.msgHistory) == 0 {
			return nil
		}
		if a.msgHistIdx == -1 {
			// Save current input as draft before entering history
			a.msgDraft = a.chat.InputValue()
			a.msgHistIdx = len(a.msgHistory) - 1
		} else if a.msgHistIdx > 0 {
			a.msgHistIdx--
		}
		a.chat.SetInputValue(a.msgHistory[a.msgHistIdx])
	case "down":
		// Navigate forward in message history
		if a.msgHistIdx == -1 {
			return nil
		}
		if a.msgHistIdx < len(a.msgHistory)-1 {
			a.msgHistIdx++
			a.chat.SetInputValue(a.msgHistory[a.msgHistIdx])
		} else {
			// Reached end of history, restore draft
			a.msgHistIdx = -1
			a.chat.SetInputValue(a.msgDraft)
		}
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
		input := a.cmdInput.Value()
		ctx := a.commandContext()
		matches := a.registry.CompleteWithArgs(input, ctx)
		if len(matches) == 1 {
			// Single match - complete it
			parts := strings.Fields(input)
			if len(parts) <= 1 && !strings.Contains(input, " ") {
				// Completing command name
				a.cmdInput.SetValue(matches[0])
			} else {
				// Completing argument - replace last part
				if len(parts) > 0 {
					parts[len(parts)-1] = matches[0]
					a.cmdInput.SetValue(strings.Join(parts, " "))
				} else {
					a.cmdInput.SetValue(matches[0])
				}
			}
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

func (a *App) handleFormKey(key string, msg tea.KeyMsg) tea.Cmd {
	if !a.formReady || a.formView == nil {
		return nil
	}

	// Forward all keys to the form - it handles esc internally
	var cmd tea.Cmd
	a.formView, cmd = a.formView.Update(msg)
	return cmd
}
