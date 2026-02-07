package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LLMToolsCmd manages LLM function calling (tool use).
type LLMToolsCmd struct{}

func (c *LLMToolsCmd) Name() string        { return "fn" }
func (c *LLMToolsCmd) Aliases() []string   { return []string{"functions", "fc"} }
func (c *LLMToolsCmd) Description() string { return "Manage LLM function calling" }

func (c *LLMToolsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) == 0 {
			// Show status
			status := "disabled"
			statusStyle := s.Error
			if ctx.ToolsEnabled() {
				status = "enabled"
				statusStyle = s.StatusOK
			}

			var b strings.Builder
			b.WriteString(s.CardTitle.Render("LLM Function Calling"))
			b.WriteString("\n\n")
			b.WriteString("  Status: ")
			b.WriteString(statusStyle.Render(status))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("  /fn on   - Enable function calling"))
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("  /fn off  - Disable function calling"))
			b.WriteString("\n\n")
			b.WriteString(s.Subtle.Render("  Note: Enable only for models that support tools"))
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("  (Claude, GPT-4, etc). Ollama models don't support this."))

			return InjectSystemMsg{Content: b.String()}
		}

		arg := strings.ToLower(args[0])
		switch arg {
		case "on", "enable", "1", "true":
			return EnableToolsMsg{Enabled: true}
		case "off", "disable", "0", "false":
			return EnableToolsMsg{Enabled: false}
		default:
			return InjectSystemMsg{
				Content: s.Error.Render(fmt.Sprintf("Unknown argument: %s (use 'on' or 'off')", arg)),
			}
		}
	}
}

// EnableToolsMsg tells the app to enable/disable LLM tools.
type EnableToolsMsg struct {
	Enabled bool
}
