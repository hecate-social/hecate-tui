package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ModeHelpCmd shows contextual help for the current mode.
type ModeHelpCmd struct{}

func (c *ModeHelpCmd) Name() string        { return "modehelp" }
func (c *ModeHelpCmd) Aliases() []string   { return nil }
func (c *ModeHelpCmd) Description() string { return "Show help for current mode (internal)" }

func (c *ModeHelpCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return nil // Not called directly — used via ModeHelp()
}

// ModeHelp returns a tea.Cmd that injects mode-specific help into chat.
func ModeHelp(mode int, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		var b strings.Builder

		switch mode {
		case 0: // Normal
			b.WriteString(s.CardTitle.Render("Normal Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Navigation"))
			b.WriteString("\n")
			b.WriteString("  j/k       Scroll chat up/down\n")
			b.WriteString("  Ctrl+D/U  Half-page scroll\n")
			b.WriteString("  g/G       Jump to top/bottom\n")
			b.WriteString("\n")
			b.WriteString(s.Bold.Render("Mode Switching"))
			b.WriteString("\n")
			b.WriteString("  i         Enter Insert mode (type messages)\n")
			b.WriteString("  /         Enter Command mode\n")
			b.WriteString("  :         Enter Command mode (vim-style)\n")
			b.WriteString("\n")
			b.WriteString(s.Bold.Render("Actions"))
			b.WriteString("\n")
			b.WriteString("  ?         Show this help\n")
			b.WriteString("  r         Retry last message\n")
			b.WriteString("  y         Copy last response to clipboard\n")
			b.WriteString("  q         Quit\n")
			b.WriteString("  Ctrl+C    Force quit\n")
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("Type / to see available commands"))

		case 1: // Insert
			b.WriteString(s.CardTitle.Render("Insert Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Messaging"))
			b.WriteString("\n")
			b.WriteString("  Enter       Send message to LLM\n")
			b.WriteString("  Alt+Enter   Insert newline (multiline)\n")
			b.WriteString("  Tab         Cycle through available models\n")
			b.WriteString("  Esc         Return to Normal (or cancel streaming)\n")
			b.WriteString("\n")
			b.WriteString(s.Bold.Render("During Streaming"))
			b.WriteString("\n")
			b.WriteString("  Esc       Cancel the current response\n")

		case 2: // Command
			b.WriteString(s.CardTitle.Render("Command Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Input"))
			b.WriteString("\n")
			b.WriteString("  Enter     Execute command\n")
			b.WriteString("  Tab       Autocomplete command name\n")
			b.WriteString("  Up/Down   Browse command history\n")
			b.WriteString("  Esc       Cancel and return to Normal\n")

		case 3: // Browse
			b.WriteString(s.CardTitle.Render("Browse Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Navigation"))
			b.WriteString("\n")
			b.WriteString("  j/k       Navigate capability list\n")
			b.WriteString("  g/G       Jump to top/bottom\n")
			b.WriteString("  Enter     View capability details\n")
			b.WriteString("  /         Search/filter capabilities\n")
			b.WriteString("  r         Refresh list\n")
			b.WriteString("  Esc       Return to Normal\n")

		case 4: // Pair
			b.WriteString(s.CardTitle.Render("Pair Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Actions"))
			b.WriteString("\n")
			b.WriteString("  p         Start pairing / re-pair\n")
			b.WriteString("  c         Cancel pairing\n")
			b.WriteString("  r         Refresh identity\n")
			b.WriteString("  Esc       Return to Normal\n")

		case 5: // Edit
			b.WriteString(s.CardTitle.Render("Edit Mode"))
			b.WriteString("\n\n")
			b.WriteString(s.Bold.Render("Actions"))
			b.WriteString("\n")
			b.WriteString("  Ctrl+S    Save file\n")
			b.WriteString("  Ctrl+Q    Close editor\n")
			b.WriteString("  Esc       Close editor\n")

		default:
			b.WriteString(s.CardTitle.Render("Help"))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("No help available for this mode."))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}
