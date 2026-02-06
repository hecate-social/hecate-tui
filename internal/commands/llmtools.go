package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
)

// LLMToolsCmd lists and manages LLM tools for function calling.
type LLMToolsCmd struct{}

func (c *LLMToolsCmd) Name() string        { return "llmtools" }
func (c *LLMToolsCmd) Aliases() []string   { return []string{"lt", "agentic"} }
func (c *LLMToolsCmd) Description() string { return "List/manage LLM tools (enable|disable <tool>)" }

func (c *LLMToolsCmd) Execute(args []string, ctx *Context) tea.Cmd {
	// Handle subcommands
	if len(args) >= 1 {
		switch strings.ToLower(args[0]) {
		case "enable":
			return c.enableTool(args[1:], ctx)
		case "disable":
			return c.disableTool(args[1:], ctx)
		case "list", "ls":
			// Fall through to list
		default:
			// Unknown subcommand, show help
			return func() tea.Msg {
				return InjectSystemMsg{Content: "Usage: /llmtools [enable|disable <tool>]\n\nExamples:\n  /llmtools              - List all tools\n  /llmtools disable bash - Disable the bash tool\n  /llmtools enable bash  - Re-enable the bash tool"}
			}
		}
	}

	return c.listTools(ctx)
}

func (c *LLMToolsCmd) enableTool(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			return InjectSystemMsg{Content: "Usage: /llmtools enable <tool_name>"}
		}

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: "Tool system not initialized"}
		}

		toolName := args[0]
		registry := executor.Registry()

		// Verify tool exists
		if _, _, ok := registry.Get(toolName); !ok {
			return InjectSystemMsg{Content: fmt.Sprintf("Unknown tool: %s\nUse /llmtools to see available tools.", toolName)}
		}

		permissions := executor.Permissions()
		permissions.EnableTool(toolName)

		return InjectSystemMsg{Content: fmt.Sprintf("✓ Enabled tool: %s", toolName)}
	}
}

func (c *LLMToolsCmd) disableTool(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			return InjectSystemMsg{Content: "Usage: /llmtools disable <tool_name>"}
		}

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: "Tool system not initialized"}
		}

		toolName := args[0]
		registry := executor.Registry()

		// Verify tool exists
		if _, _, ok := registry.Get(toolName); !ok {
			return InjectSystemMsg{Content: fmt.Sprintf("Unknown tool: %s\nUse /llmtools to see available tools.", toolName)}
		}

		permissions := executor.Permissions()
		permissions.DisableTool(toolName)

		return InjectSystemMsg{Content: fmt.Sprintf("✗ Disabled tool: %s", toolName)}
	}
}

func (c *LLMToolsCmd) listTools(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: "Tool system not initialized"}
		}

		registry := executor.Registry()
		if registry == nil {
			return InjectSystemMsg{Content: "Tool registry not available"}
		}

		permissions := executor.Permissions()
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
				isDisabled := permissions.IsDisabled(tool.Name)

				// Tool name with status indicator
				b.WriteString("  ")
				if isDisabled {
					b.WriteString(s.StatusError.Render("○"))
					b.WriteString(" ")
					b.WriteString(s.Subtle.Render(tool.Name + " (disabled)"))
				} else {
					b.WriteString(s.StatusOK.Render("•"))
					b.WriteString(" ")
					b.WriteString(tool.Name)
					if tool.RequiresApproval {
						b.WriteString(s.StatusWarning.Render(" ⚠"))
					}
				}
				b.WriteString("\n")

				// Description (indented, subtle) - skip for disabled tools
				if tool.Description != "" && !isDisabled {
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
		disabledCount := len(permissions.DisabledTools())
		for _, tools := range byCategory {
			for _, t := range tools {
				if t.RequiresApproval && !permissions.IsDisabled(t.Name) {
					requiresApproval++
				}
			}
		}

		b.WriteString(s.Subtle.Render(strings.Repeat("─", 40)))
		b.WriteString("\n")
		activeCount := total - disabledCount
		b.WriteString(fmt.Sprintf("  %d tools active", activeCount))
		if disabledCount > 0 {
			b.WriteString(fmt.Sprintf(", %d disabled", disabledCount))
		}
		if requiresApproval > 0 {
			b.WriteString(fmt.Sprintf(", %d require approval", requiresApproval))
		}
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  ⚠ = requires approval  ○ = disabled"))

		return InjectSystemMsg{Content: b.String()}
	}
}
