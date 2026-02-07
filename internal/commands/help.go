package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpCmd shows available commands.
type HelpCmd struct {
	registry *Registry
}

func (c *HelpCmd) Name() string        { return "help" }
func (c *HelpCmd) Aliases() []string   { return []string{"h", "?"} }
func (c *HelpCmd) Description() string { return "Show available commands" }

func (c *HelpCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		var b strings.Builder
		s := ctx.Styles
		t := ctx.Theme

		b.WriteString(s.CardTitle.Render("Hecate Commands"))
		b.WriteString("\n\n")

		// Styling helpers
		cmdStyle := lipgloss.NewStyle().Foreground(t.Secondary)
		descStyle := lipgloss.NewStyle().Foreground(t.Text)

		row := func(cmd, aliases, desc string) string {
			// Build command string with alias
			cmdStr := cmd
			if aliases != "" {
				cmdStr += " " + aliases
			}
			// Pad to 30 chars for alignment
			for len(cmdStr) < 30 {
				cmdStr += " "
			}
			return cmdStyle.Render(cmdStr) + descStyle.Render(desc) + "\n"
		}

		section := func(emoji, title string) string {
			return s.Bold.Render(emoji+" "+title) + "\n"
		}

		// 📋 General
		b.WriteString(section("📋", "General"))
		b.WriteString(row("/help", "(h, ?)", "Show this help"))
		b.WriteString(row("/clear", "", "Clear the screen"))
		b.WriteString(row("/quit", "(q, exit)", "Exit Hecate"))
		b.WriteString("\n")

		// 💬 Chat
		b.WriteString(section("💬", "Chat"))
		b.WriteString(row("/new", "", "Start new conversation"))
		b.WriteString(row("/history", "", "Show conversation history"))
		b.WriteString(row("/delete", "(del)", "Delete messages"))
		b.WriteString(row("/save", "", "Save conversation"))
		b.WriteString(row("/edit", "", "Edit a message"))
		b.WriteString(row("/system", "(sys)", "Set system prompt"))
		b.WriteString("\n")

		// 🤖 LLM & Models
		b.WriteString(section("🤖", "LLM & Models"))
		b.WriteString(row("/models", "", "List available models"))
		b.WriteString(row("/model", "", "Show/select current model"))
		b.WriteString(row("/load", "", "Load a model"))
		b.WriteString(row("/provider", "", "Manage LLM providers"))
		b.WriteString(row("/browse", "", "Browse capabilities"))
		b.WriteString("\n")

		// 🌐 Mesh & Network
		b.WriteString(section("🌐", "Mesh & Network"))
		b.WriteString(row("/status", "", "Show daemon status"))
		b.WriteString(row("/health", "", "Health check"))
		b.WriteString(row("/call", "(rpc)", "Call mesh procedure"))
		b.WriteString(row("/subscriptions", "(subs)", "Show subscriptions"))
		b.WriteString(row("/me", "", "Show identity"))
		b.WriteString("\n")

		// 🛠️ Project & Tools
		b.WriteString(section("🛠️", "Project & Tools"))
		b.WriteString(row("/project", "(proj)", "Show workspace info"))
		b.WriteString(row("/config", "", "Show configuration"))
		b.WriteString(row("/alc", "(lifecycle, lc)", "Application lifecycle"))
		b.WriteString(row("/pair", "", "Pair programming mode"))
		b.WriteString(row("/find", "", "Find in codebase"))
		b.WriteString(row("/tools", "", "Detect developer tools"))
		b.WriteString(row("/fn", "(on|off)", "LLM function calling"))
		b.WriteString("\n")

		// 🎨 Appearance
		b.WriteString(section("🎨", "Appearance"))
		b.WriteString(row("/theme", "", "Change theme"))
		b.WriteString("\n")

		b.WriteString(s.Subtle.Render("Type / or : to enter command mode"))

		return InjectSystemMsg{Content: b.String()}
	}
}
