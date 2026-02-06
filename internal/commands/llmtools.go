package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
)

// LLMToolsCmd lists available LLM tools for function calling.
type LLMToolsCmd struct{}

func (c *LLMToolsCmd) Name() string        { return "llmtools" }
func (c *LLMToolsCmd) Aliases() []string   { return []string{"lt", "agentic"} }
func (c *LLMToolsCmd) Description() string { return "List available LLM tools for function calling" }

func (c *LLMToolsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: "Tool system not initialized"}
		}

		registry := executor.Registry()
		if registry == nil {
			return InjectSystemMsg{Content: "Tool registry not available"}
		}

		s := ctx.Styles
		var b strings.Builder

		// Header
		enabled := ctx.ToolsEnabled()
		status := "enabled"
		if !enabled {
			status = "disabled"
		}
		b.WriteString(s.CardTitle.Render("LLM Tools"))
		b.WriteString("  ")
		if enabled {
			b.WriteString(s.StatusOK.Render("● " + status))
		} else {
			b.WriteString(s.Subtle.Render("○ " + status))
		}
		b.WriteString("\n\n")

		// Group by category
		byCategory := registry.ByCategory()
		categories := []llmtools.ToolCategory{
			llmtools.CategoryFileSystem,
			llmtools.CategoryCodeExplore,
			llmtools.CategorySystem,
			llmtools.CategoryWeb,
			llmtools.CategoryMesh,
		}

		for _, cat := range categories {
			tools, ok := byCategory[cat]
			if !ok || len(tools) == 0 {
				continue
			}

			// Category header
			catName := llmtools.CategoryName(cat)
			b.WriteString(s.Bold.Render(catName))
			b.WriteString("\n")

			for _, tool := range tools {
				// Tool name and approval indicator
				approval := ""
				if tool.RequiresApproval {
					approval = s.StatusWarning.Render(" ⚠")
				}
				b.WriteString("  ")
				b.WriteString(s.StatusOK.Render("•"))
				b.WriteString(" ")
				b.WriteString(tool.Name)
				b.WriteString(approval)
				b.WriteString("\n")

				// Description (indented, subtle)
				if tool.Description != "" {
					desc := tool.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					b.WriteString("    ")
					b.WriteString(s.Subtle.Render(desc))
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
		}

		// Summary
		total := registry.Count()
		requiresApproval := 0
		for _, tools := range byCategory {
			for _, t := range tools {
				if t.RequiresApproval {
					requiresApproval++
				}
			}
		}

		b.WriteString(s.Subtle.Render(strings.Repeat("─", 40)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %d tools available", total))
		if requiresApproval > 0 {
			b.WriteString(fmt.Sprintf(", %d require approval", requiresApproval))
		}
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  ⚠ = requires user approval before execution"))

		return InjectSystemMsg{Content: b.String()}
	}
}
