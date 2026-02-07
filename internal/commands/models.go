package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ModelsCmd lists available LLM models.
type ModelsCmd struct{}

func (c *ModelsCmd) Name() string        { return "models" }
func (c *ModelsCmd) Aliases() []string   { return nil }
func (c *ModelsCmd) Description() string { return "List available LLM models" }

func (c *ModelsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		models, err := ctx.Client.ListModels()
		if err != nil {
			return InjectSystemMsg{Content: s.Error.Render("Failed to list models: " + err.Error())}
		}

		if len(models) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No models available. Is Ollama running?\nUse /provider add <type> <key> to add a cloud provider.")}
		}

		// Calculate column widths
		maxName := 4 // "Name"
		maxSize := 4 // "Size"
		maxProvider := 8 // "Provider"
		for _, m := range models {
			if len(m.Name) > maxName {
				maxName = len(m.Name)
			}
			if len(m.ParameterSize) > maxSize {
				maxSize = len(m.ParameterSize)
			}
			if len(m.Provider) > maxProvider {
				maxProvider = len(m.Provider)
			}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Available Models"))
		b.WriteString("\n\n")

		// Header
		header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			maxName, "Name",
			maxSize, "Size",
			maxProvider, "Provider",
			"Family")
		b.WriteString(s.Subtle.Render(header))
		b.WriteString("\n")

		// Separator
		b.WriteString(s.Subtle.Render("  " + strings.Repeat("─", maxName+maxSize+maxProvider+20)))
		b.WriteString("\n")

		// Rows
		for _, m := range models {
			size := m.ParameterSize
			if size == "" {
				size = "-"
			}
			provider := m.Provider
			if provider == "" {
				provider = "local"
			}
			family := m.Family
			if family == "" {
				family = "-"
			}

			b.WriteString("  ")
			b.WriteString(s.Bold.Render(fmt.Sprintf("%-*s", maxName, m.Name)))
			b.WriteString("  ")
			b.WriteString(s.CardValue.Render(fmt.Sprintf("%-*s", maxSize, size)))
			b.WriteString("  ")
			b.WriteString(s.Subtle.Render(fmt.Sprintf("%-*s", maxProvider, provider)))
			b.WriteString("  ")
			b.WriteString(s.Subtle.Render(family))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Use /model <name> to switch"))

		return InjectSystemMsg{Content: b.String()}
	}
}

// ModelCmd switches the active LLM model.
type ModelCmd struct{}

func (c *ModelCmd) Name() string        { return "model" }
func (c *ModelCmd) Aliases() []string   { return nil }
func (c *ModelCmd) Description() string { return "Switch LLM model (/model <name>)" }

// SwitchModelMsg tells the chat to switch its active model.
type SwitchModelMsg struct {
	Name string
}

func (c *ModelCmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return func() tea.Msg {
			return InjectSystemMsg{Content: ctx.Styles.Subtle.Render("Usage: /model <name>")}
		}
	}

	modelName := strings.Join(args, " ")
	return func() tea.Msg {
		return SwitchModelMsg{Name: modelName}
	}
}
