package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/modes"
)

// handleKey dispatches keys at the shell level.
func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if key == "ctrl+c" {
		return tea.Quit
	}

	// Home screen keys
	if a.showHome {
		return a.handleHomeKey(key)
	}

	// Command mode: shell handles all keys
	if a.inCommandMode {
		return a.handleCommandKey(key, msg)
	}

	// Active studio's mode determines which keys the shell intercepts
	activeMode := a.studios[a.activeStudio].Mode()

	if activeMode == modes.Normal {
		// Studio switching (Ctrl+1-5)
		switch key {
		case "ctrl+1":
			return a.switchStudio(0)
		case "ctrl+2":
			return a.switchStudio(1)
		case "ctrl+3":
			return a.switchStudio(2)
		case "ctrl+4":
			return a.switchStudio(3)
		case "ctrl+5":
			return a.switchStudio(4)
		case "q":
			return tea.Quit
		case "/", ":":
			a.enterCommandMode(key)
			return nil
		}
	}

	// All other keys go to the studio via Update() forwarding
	return nil
}

func (a *App) handleHomeKey(key string) tea.Cmd {
	switch key {
	case "1":
		return a.switchStudio(0)
	case "2":
		return a.switchStudio(1)
	case "3":
		return a.switchStudio(2)
	case "4":
		return a.switchStudio(3)
	case "5":
		return a.switchStudio(4)
	case "q":
		return tea.Quit
	}
	return nil
}

func (a *App) enterCommandMode(prefix string) {
	a.inCommandMode = true
	a.cmdInput.SetValue("")
	if prefix == ":" {
		a.cmdInput.Prompt = ":"
	} else {
		a.cmdInput.Prompt = "/"
	}
	a.cmdInput.Focus()
}

func (a *App) handleCommandKey(key string, msg tea.KeyMsg) tea.Cmd {
	switch key {
	case "esc":
		a.inCommandMode = false
		a.cmdHistIdx = -1
		a.cmdInput.Blur()
		return nil
	case "enter":
		input := a.cmdInput.Value()
		prefix := a.cmdInput.Prompt
		a.inCommandMode = false
		a.cmdInput.Blur()
		if input != "" {
			if len(a.cmdHistory) == 0 || a.cmdHistory[len(a.cmdHistory)-1] != input {
				a.cmdHistory = append(a.cmdHistory, input)
				if len(a.cmdHistory) > 50 {
					a.cmdHistory = a.cmdHistory[1:]
				}
			}
			a.cmdHistIdx = -1
			ctx := a.commandContext()
			return a.registry.Dispatch(prefix+input, ctx)
		}
		return nil
	case "up":
		if len(a.cmdHistory) == 0 {
			return nil
		}
		if a.cmdHistIdx == -1 {
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
			parts := strings.Fields(input)
			if len(parts) <= 1 && !strings.Contains(input, " ") {
				a.cmdInput.SetValue(matches[0])
			} else {
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
		a.cmdHistIdx = -1
		var cmd tea.Cmd
		a.cmdInput, cmd = a.cmdInput.Update(msg)
		return cmd
	}
}
