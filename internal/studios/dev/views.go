package dev

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// View renders the Dev Studio content area.
func (s *Studio) View() string {
	if s.width == 0 {
		return ""
	}

	if s.noVenture {
		return s.viewNoVenture()
	}
	if s.loading {
		return s.viewLoading()
	}
	if s.loadErr != nil {
		return s.viewError()
	}
	return s.viewTaskList()
}

func (s *Studio) viewNoVenture() string {
	t := s.ctx.Theme
	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Development Studio")

	msg := lipgloss.NewStyle().
		Foreground(t.TextMuted).
		Render("No active venture found.")

	hint := lipgloss.NewStyle().
		Foreground(t.TextDim).Italic(true).
		Render("Use /venture init in the LLM Studio to get started.")

	content := title + "\n\n" + msg + "\n\n" + hint
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *Studio) viewLoading() string {
	t := s.ctx.Theme
	msg := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Loading venture tasks...")
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, msg)
}

func (s *Studio) viewError() string {
	t := s.ctx.Theme
	title := lipgloss.NewStyle().
		Foreground(t.Error).Bold(true).
		Render("Failed to load tasks")

	detail := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render(s.loadErr.Error())

	hint := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("Press r to retry")

	content := title + "\n\n" + detail + "\n\n" + hint
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *Studio) viewTaskList() string {
	t := s.ctx.Theme
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("Venture: " + s.ventureName)
	b.WriteString(header)
	b.WriteString("\n")

	// Separator
	sep := lipgloss.NewStyle().
		Foreground(t.Border).
		Render(strings.Repeat("\u2501", min(s.width, 60)))
	b.WriteString(sep)
	b.WriteString("\n")

	// Task rows
	visible := s.taskList.VisibleItems()
	if len(visible) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(t.TextDim).Italic(true).
			Render("  No tasks yet")
		b.WriteString(empty)
		return b.String()
	}

	offset := s.taskList.ScrollOffset()
	maxRows := s.height - 2 // subtract header lines
	if maxRows < 1 {
		maxRows = 1
	}

	for i := offset; i < len(visible) && i-offset < maxRows; i++ {
		item := visible[i]
		selected := i == s.taskList.Cursor()
		b.WriteString(s.renderRow(item, selected))
		if i < len(visible)-1 && i-offset < maxRows-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (s *Studio) renderRow(item TaskItem, selected bool) string {
	t := s.ctx.Theme

	if item.IsHeader {
		return s.renderDivisionHeader(item, selected)
	}

	// Indent
	indent := ""
	switch item.Depth {
	case 0:
		indent = " "
	case 2:
		indent = "    "
	}

	// State symbol
	symbol := stateSymbol(item.State)
	symbolStyle := stateStyle(t, item.State)

	// Label
	label := item.Label
	labelStyle := lipgloss.NewStyle().Foreground(t.Text)
	if item.State == "blocked" {
		labelStyle = lipgloss.NewStyle().Foreground(t.TextMuted)
	} else if item.State == "done" {
		labelStyle = lipgloss.NewStyle().Foreground(t.TextDim)
	}

	// AI role badge
	roleBadge := ""
	if item.AIRole != "" {
		roleBadge = lipgloss.NewStyle().
			Foreground(t.Secondary).
			Render(" [" + item.AIRole + "]")
	}

	// Cursor marker
	cursor := " "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8")
	}

	// Build the row
	row := fmt.Sprintf("%s%s%s %s%s",
		cursor, indent,
		symbolStyle.Render(symbol),
		labelStyle.Render(label),
		roleBadge,
	)

	if selected {
		// Highlight the entire row
		row = lipgloss.NewStyle().Bold(true).Render(row)
	}

	return row
}

func (s *Studio) renderDivisionHeader(item TaskItem, selected bool) string {
	t := s.ctx.Theme

	// Collapse indicator
	arrow := "\u25b8" // right-pointing triangle (collapsed)
	if !item.Collapsed {
		arrow = "\u25be" // down-pointing triangle (expanded)
	}

	// Dotted separator before division groups
	dotSep := lipgloss.NewStyle().
		Foreground(t.Border).
		Render(strings.Repeat("\u2500 ", min(s.width/2, 30)))

	cursor := " "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("\u25b8")
	}

	arrowStyle := lipgloss.NewStyle().Foreground(t.TextDim)
	nameStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)

	row := fmt.Sprintf("%s %s %s",
		cursor,
		arrowStyle.Render(arrow),
		nameStyle.Render(item.Label),
	)

	if selected {
		row = lipgloss.NewStyle().Bold(true).Render(row)
	}

	return dotSep + "\n" + row
}

// stateSymbol returns the Unicode symbol for a task state.
func stateSymbol(state string) string {
	switch state {
	case "done":
		return "\u2713" // check mark
	case "active":
		return "\u25cf" // filled circle
	case "paused":
		return "\u25d1" // half circle
	case "running":
		return "\u25d0" // other half circle
	case "blocked":
		return "\u25cb" // open circle
	case "pending":
		return "\u25cb" // open circle
	default:
		return "\u25cb"
	}
}

// stateStyle returns the lipgloss style for a task state symbol.
func stateStyle(t *theme.Theme, state string) lipgloss.Style {
	switch state {
	case "done":
		return lipgloss.NewStyle().Foreground(t.Success)
	case "active":
		return lipgloss.NewStyle().Foreground(t.Primary)
	case "paused":
		return lipgloss.NewStyle().Foreground(t.TextDim)
	case "running":
		return lipgloss.NewStyle().Foreground(t.Accent)
	case "blocked":
		return lipgloss.NewStyle().Foreground(t.TextMuted)
	case "pending":
		return lipgloss.NewStyle().Foreground(t.Text)
	default:
		return lipgloss.NewStyle().Foreground(t.TextDim)
	}
}
