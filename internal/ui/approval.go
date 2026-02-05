package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// ApprovalPrompt renders a tool approval dialog.
type ApprovalPrompt struct {
	theme  *theme.Theme
	styles *theme.Styles
	width  int
}

// NewApprovalPrompt creates a new approval prompt renderer.
func NewApprovalPrompt(t *theme.Theme, s *theme.Styles) *ApprovalPrompt {
	return &ApprovalPrompt{
		theme:  t,
		styles: s,
		width:  60,
	}
}

// SetWidth sets the dialog width.
func (p *ApprovalPrompt) SetWidth(w int) {
	p.width = w
	if p.width < 40 {
		p.width = 40
	}
	if p.width > 100 {
		p.width = 100
	}
}

// Render renders the approval prompt for a tool call.
func (p *ApprovalPrompt) Render(tool llmtools.Tool, call llm.ToolCall) string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Warning).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.theme.Warning).
		Padding(1, 2).
		Width(p.width)

	labelStyle := lipgloss.NewStyle().
		Foreground(p.theme.Primary).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(p.theme.Text)

	dimStyle := lipgloss.NewStyle().
		Foreground(p.theme.TextDim)

	keyStyle := lipgloss.NewStyle().
		Foreground(p.theme.Success).
		Bold(true)

	// Format title
	title := titleStyle.Render(fmt.Sprintf("🔧 Tool Request: %s", tool.Name))

	// Format description
	desc := valueStyle.Render(tool.Description)

	// Format arguments
	var argsDisplay string
	if len(call.Arguments) > 0 {
		var args map[string]interface{}
		if err := json.Unmarshal(call.Arguments, &args); err == nil {
			argsDisplay = p.formatArgs(args, labelStyle, valueStyle)
		} else {
			// Fallback: show raw JSON truncated
			raw := string(call.Arguments)
			if len(raw) > 200 {
				raw = raw[:200] + "..."
			}
			argsDisplay = dimStyle.Render(raw)
		}
	} else {
		argsDisplay = dimStyle.Render("(no arguments)")
	}

	// Build category badge
	categoryBadge := p.categoryBadge(tool.Category)

	// Format keybindings
	keybindings := fmt.Sprintf(
		"%s Allow  %s Deny  %s Allow all (session)",
		keyStyle.Render("[y]"),
		keyStyle.Render("[n]"),
		keyStyle.Render("[a]"),
	)

	// Assemble content
	var parts []string
	parts = append(parts, title)
	parts = append(parts, "")
	parts = append(parts, categoryBadge)
	parts = append(parts, "")
	parts = append(parts, desc)
	parts = append(parts, "")
	parts = append(parts, labelStyle.Render("Arguments:"))
	parts = append(parts, argsDisplay)
	parts = append(parts, "")
	parts = append(parts, keybindings)

	content := strings.Join(parts, "\n")
	return borderStyle.Render(content)
}

// formatArgs formats the arguments map for display.
func (p *ApprovalPrompt) formatArgs(args map[string]interface{}, labelStyle, valueStyle lipgloss.Style) string {
	var lines []string
	for key, val := range args {
		valStr := p.formatValue(val)
		// Truncate long values
		if len(valStr) > 60 {
			valStr = valStr[:57] + "..."
		}
		line := fmt.Sprintf("  %s %s",
			labelStyle.Render(key+":"),
			valueStyle.Render(valStr))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// formatValue formats a single value for display.
func (p *ApprovalPrompt) formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		var items []string
		for _, item := range v {
			items = append(items, p.formatValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// categoryBadge returns a styled badge for the tool category.
func (p *ApprovalPrompt) categoryBadge(category llmtools.ToolCategory) string {
	var color lipgloss.Color
	var icon string

	switch category {
	case llmtools.CategoryFileSystem:
		color = lipgloss.Color("#4a9eff") // blue
		icon = "📁"
	case llmtools.CategoryCodeExplore:
		color = lipgloss.Color("#9b59b6") // purple
		icon = "🔍"
	case llmtools.CategorySystem:
		color = lipgloss.Color("#e74c3c") // red
		icon = "⚙️"
	case llmtools.CategoryWeb:
		color = lipgloss.Color("#3498db") // light blue
		icon = "🌐"
	case llmtools.CategoryMesh:
		color = lipgloss.Color("#2ecc71") // green
		icon = "🔗"
	default:
		color = p.theme.TextDim
		icon = "🔧"
	}

	badge := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(fmt.Sprintf("%s %s", icon, string(category)))

	return badge
}
