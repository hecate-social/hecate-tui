package commands

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// AICmd manages LLM tool configuration and status.
type AICmd struct{}

func (c *AICmd) Name() string        { return "ai" }
func (c *AICmd) Aliases() []string   { return []string{"ai-tools", "llmtools"} }
func (c *AICmd) Description() string { return "Manage AI/LLM tools (/ai, /ai enable <tool>, /ai disable <tool>)" }

func (c *AICmd) Execute(args []string, ctx *Context) tea.Cmd {
	if len(args) == 0 {
		return c.listTools(ctx)
	}

	switch strings.ToLower(args[0]) {
	case "list", "ls":
		return c.listTools(ctx)
	case "enable", "on":
		return c.enableTool(args[1:], ctx)
	case "disable", "off":
		return c.disableTool(args[1:], ctx)
	case "status":
		return c.showStatus(ctx)
	case "help":
		return c.showHelp(ctx)
	default:
		// Assume it's a tool name - show info about it
		return c.showToolInfo(args[0], ctx)
	}
}

func (c *AICmd) listTools(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		// Get tool executor from chat model if available
		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: s.Error.Render("Tool system not initialized")}
		}

		registry := executor.Registry()
		permissions := executor.Permissions()
		tools := registry.All()

		if len(tools) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("No tools registered")}
		}

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("AI Tools"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("Function calling capabilities for the LLM"))
		b.WriteString("\n\n")

		// Group tools by category
		categories := []llmtools.ToolCategory{
			llmtools.CategoryFileSystem,
			llmtools.CategoryCodeExplore,
			llmtools.CategorySystem,
			llmtools.CategoryWeb,
			llmtools.CategoryMesh,
		}

		for _, cat := range categories {
			var catTools []llmtools.Tool
			for _, t := range tools {
				if t.Category == cat {
					catTools = append(catTools, t)
				}
			}

			if len(catTools) == 0 {
				continue
			}

			// Sort by name within category
			sort.Slice(catTools, func(i, j int) bool {
				return catTools[i].Name < catTools[j].Name
			})

			icon := categoryIcon(cat)
			name := llmtools.CategoryName(cat)
			b.WriteString(s.Bold.Render(icon + " " + name))
			b.WriteString("\n")

			for _, t := range catTools {
				// Check permission status
				perm := permissions.Check(t.Name, nil)
				status := permissionBadge(perm, s)

				// Approval indicator
				approval := ""
				if t.RequiresApproval {
					approval = s.StatusWarning.Render(" [asks]")
				}

				b.WriteString(fmt.Sprintf("  %s %s%s\n", status, t.Name, approval))
				b.WriteString(fmt.Sprintf("     %s\n", s.Subtle.Render(t.Description)))
			}
			b.WriteString("\n")
		}

		// Footer with commands
		b.WriteString(s.Subtle.Render(strings.Repeat("─", 40)))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  /ai <tool>          Show tool details\n"))
		b.WriteString(s.Subtle.Render("  /ai enable <tool>   Allow tool without asking\n"))
		b.WriteString(s.Subtle.Render("  /ai disable <tool>  Block tool execution\n"))
		b.WriteString("\n")
		b.WriteString(s.StatusOK.Render("●") + " allow  ")
		b.WriteString(s.StatusWarning.Render("●") + " ask  ")
		b.WriteString(s.StatusError.Render("●") + " deny  ")
		b.WriteString(s.StatusWarning.Render("[asks]") + " requires approval")

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *AICmd) enableTool(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("Usage: /ai enable <tool-name>")}
		}

		toolName := args[0]

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: s.Error.Render("Tool system not initialized")}
		}

		registry := executor.Registry()
		permissions := executor.Permissions()

		// Check if tool exists
		tool, _, ok := registry.Get(toolName)
		if !ok {
			return InjectSystemMsg{Content: s.Error.Render("Unknown tool: " + toolName)}
		}

		// Set permission to allow
		permissions.SetToolPermission(toolName, llmtools.PermissionAllow)

		msg := s.StatusOK.Render("Enabled: " + toolName)
		if tool.RequiresApproval {
			msg += "\n" + s.StatusWarning.Render("Note: This tool normally requires approval. It will now run without asking.")
		}

		return InjectSystemMsg{Content: msg}
	}
}

func (c *AICmd) disableTool(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		if len(args) == 0 {
			return InjectSystemMsg{Content: s.Subtle.Render("Usage: /ai disable <tool-name>")}
		}

		toolName := args[0]

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: s.Error.Render("Tool system not initialized")}
		}

		registry := executor.Registry()
		permissions := executor.Permissions()

		// Check if tool exists
		_, _, ok := registry.Get(toolName)
		if !ok {
			return InjectSystemMsg{Content: s.Error.Render("Unknown tool: " + toolName)}
		}

		// Set permission to deny
		permissions.SetToolPermission(toolName, llmtools.PermissionDeny)

		return InjectSystemMsg{Content: s.StatusOK.Render("Disabled: " + toolName + " (will be blocked)")}
	}
}

func (c *AICmd) showStatus(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: s.Error.Render("Tool system not initialized")}
		}

		registry := executor.Registry()
		permissions := executor.Permissions()
		tools := registry.All()

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Tool Status"))
		b.WriteString("\n\n")

		// Tools enabled status
		toolsEnabled := ctx.ToolsEnabled()
		if toolsEnabled {
			b.WriteString(s.StatusOK.Render("● Tools enabled"))
		} else {
			b.WriteString(s.StatusError.Render("● Tools disabled"))
		}
		b.WriteString("\n\n")

		// Count by permission
		allowCount := 0
		askCount := 0
		denyCount := 0
		for _, t := range tools {
			perm := permissions.Check(t.Name, nil)
			switch perm {
			case llmtools.PermissionAllow:
				allowCount++
			case llmtools.PermissionAsk:
				askCount++
			case llmtools.PermissionDeny:
				denyCount++
			}
		}

		b.WriteString(fmt.Sprintf("  %s %d tools allowed (no prompt)\n", s.StatusOK.Render("●"), allowCount))
		b.WriteString(fmt.Sprintf("  %s %d tools ask before running\n", s.StatusWarning.Render("●"), askCount))
		b.WriteString(fmt.Sprintf("  %s %d tools blocked\n", s.StatusError.Render("●"), denyCount))
		b.WriteString(fmt.Sprintf("\n  Total: %d tools registered", len(tools)))

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *AICmd) showToolInfo(name string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		executor := ctx.GetToolExecutor()
		if executor == nil {
			return InjectSystemMsg{Content: s.Error.Render("Tool system not initialized")}
		}

		registry := executor.Registry()
		permissions := executor.Permissions()

		tool, _, ok := registry.Get(name)
		if !ok {
			return InjectSystemMsg{Content: s.Error.Render("Unknown tool: " + name + "\nUse /ai to list available tools.")}
		}

		perm := permissions.Check(name, nil)

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("Tool: " + tool.Name))
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("Description"))
		b.WriteString("\n")
		b.WriteString("  " + tool.Description)
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("Category"))
		b.WriteString("\n")
		b.WriteString("  " + categoryIcon(tool.Category) + " " + llmtools.CategoryName(tool.Category))
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("Permission"))
		b.WriteString("\n")
		b.WriteString("  " + permissionBadge(perm, s) + " " + permissionText(perm))
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("Requires Approval"))
		b.WriteString("\n")
		if tool.RequiresApproval {
			b.WriteString("  " + s.StatusWarning.Render("Yes") + " - will ask before running (unless explicitly allowed)")
		} else {
			b.WriteString("  " + s.StatusOK.Render("No") + " - runs without prompting")
		}
		b.WriteString("\n\n")

		// Parameters
		if len(tool.Parameters.Properties) > 0 {
			b.WriteString(s.Bold.Render("Parameters"))
			b.WriteString("\n")

			// Sort parameter names
			var paramNames []string
			for name := range tool.Parameters.Properties {
				paramNames = append(paramNames, name)
			}
			sort.Strings(paramNames)

			for _, pname := range paramNames {
				spec := tool.Parameters.Properties[pname]
				required := ""
				for _, r := range tool.Parameters.Required {
					if r == pname {
						required = s.StatusWarning.Render("*")
						break
					}
				}
				b.WriteString(fmt.Sprintf("  %s%s (%s)\n", pname, required, spec.Type))
				b.WriteString(fmt.Sprintf("    %s\n", s.Subtle.Render(spec.Description)))
			}
			b.WriteString("\n")
			b.WriteString(s.Subtle.Render("  * = required"))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

func (c *AICmd) showHelp(ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		var b strings.Builder
		b.WriteString(s.CardTitle.Render("AI Tools Help"))
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("What are AI Tools?"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  AI tools (function calling) let the LLM interact with your system -"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  reading files, searching code, executing commands, and more."))
		b.WriteString("\n\n")

		b.WriteString(s.Bold.Render("Commands"))
		b.WriteString("\n")
		b.WriteString("  /ai                List all available tools\n")
		b.WriteString("  /ai <tool>         Show details for a specific tool\n")
		b.WriteString("  /ai enable <tool>  Allow tool to run without asking\n")
		b.WriteString("  /ai disable <tool> Block tool from running\n")
		b.WriteString("  /ai status         Show overall tool status\n")
		b.WriteString("\n")

		b.WriteString(s.Bold.Render("Permission Levels"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s Allow  - Tool runs without prompting\n", s.StatusOK.Render("●")))
		b.WriteString(fmt.Sprintf("  %s Ask    - You'll be asked before each use\n", s.StatusWarning.Render("●")))
		b.WriteString(fmt.Sprintf("  %s Deny   - Tool is blocked entirely\n", s.StatusError.Render("●")))
		b.WriteString("\n")

		b.WriteString(s.Bold.Render("Safety"))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Tools that modify files or run commands require approval by default."))
		b.WriteString("\n")
		b.WriteString(s.Subtle.Render("  Press [y] to approve, [n] to deny, [a] to allow for the session."))

		return InjectSystemMsg{Content: b.String()}
	}
}

// --- Helpers ---

func categoryIcon(cat llmtools.ToolCategory) string {
	switch cat {
	case llmtools.CategoryFileSystem:
		return "📁"
	case llmtools.CategoryCodeExplore:
		return "🔍"
	case llmtools.CategorySystem:
		return "⚙️"
	case llmtools.CategoryWeb:
		return "🌐"
	case llmtools.CategoryMesh:
		return "🔗"
	default:
		return "📦"
	}
}

func permissionBadge(perm llmtools.PermissionLevel, s *theme.Styles) string {
	switch perm {
	case llmtools.PermissionAllow:
		return s.StatusOK.Render("●")
	case llmtools.PermissionAsk:
		return s.StatusWarning.Render("●")
	case llmtools.PermissionDeny:
		return s.StatusError.Render("●")
	default:
		return s.Subtle.Render("○")
	}
}

func permissionText(perm llmtools.PermissionLevel) string {
	switch perm {
	case llmtools.PermissionAllow:
		return "Allow (runs without prompting)"
	case llmtools.PermissionAsk:
		return "Ask (will prompt before running)"
	case llmtools.PermissionDeny:
		return "Deny (blocked)"
	default:
		return "Unknown"
	}
}
