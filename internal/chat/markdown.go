package chat

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// RenderMarkdown applies lightweight markdown formatting to text.
// Handles: code blocks, inline code, bold, italic, headers, bullet lists.
// Designed for LLM output — no external dependencies.
func RenderMarkdown(text string, t *theme.Theme, width int) string {
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false
	var codeLines []string
	codeLang := ""

	codeBlockStyle := lipgloss.NewStyle().
		Foreground(t.CodeText).
		Background(t.CodeBg).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(t.Primary).
		PaddingLeft(1).
		PaddingRight(1)

	codeLabelStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted).
		Bold(true)

	inlineCodeStyle := lipgloss.NewStyle().
		Foreground(t.Secondary).
		Background(t.BgCard)

	headerStyle := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	h2Style := lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	h3Style := lipgloss.NewStyle().
		Foreground(t.PrimaryLight).
		Bold(true)

	boldStyle := lipgloss.NewStyle().
		Foreground(t.Text).
		Bold(true)

	italicStyle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Italic(true)

	bulletStyle := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	hrStyle := lipgloss.NewStyle().
		Foreground(t.Border)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block boundaries
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// End of code block
				inCodeBlock = false
				code := strings.Join(codeLines, "\n")

				codeW := width - 4
				if codeW < 20 {
					codeW = 20
				}

				// Combine label and code inside single styled block
				var content string
				if codeLang != "" {
					content = codeLabelStyle.Render(codeLang) + "\n" + code
				} else {
					content = code
				}
				block := codeBlockStyle.Width(codeW).Render(content)
				result = append(result, block)
				codeLines = nil
				codeLang = ""
			} else {
				// Start of code block
				inCodeBlock = true
				codeLang = strings.TrimPrefix(trimmed, "```")
				codeLang = strings.TrimSpace(codeLang)
				codeLines = nil
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			ruleW := width - 4
			if ruleW < 10 {
				ruleW = 10
			}
			result = append(result, hrStyle.Render(strings.Repeat("─", ruleW)))
			continue
		}

		// Headers
		if strings.HasPrefix(trimmed, "### ") {
			result = append(result, h3Style.Render(strings.TrimPrefix(trimmed, "### ")))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			result = append(result, h2Style.Render(strings.TrimPrefix(trimmed, "## ")))
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			result = append(result, headerStyle.Render(strings.TrimPrefix(trimmed, "# ")))
			continue
		}

		// Bullet lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			indent := leadingSpaces(line)
			content := trimmed[2:]
			content = formatInline(content, inlineCodeStyle, boldStyle, italicStyle)
			result = append(result, strings.Repeat(" ", indent)+bulletStyle.Render("*")+" "+content)
			continue
		}

		// Numbered lists
		if isNumberedList(trimmed) {
			num, content := parseNumberedList(trimmed)
			indent := leadingSpaces(line)
			content = formatInline(content, inlineCodeStyle, boldStyle, italicStyle)
			result = append(result, strings.Repeat(" ", indent)+bulletStyle.Render(num)+" "+content)
			continue
		}

		// Regular text with inline formatting
		formatted := formatInline(trimmed, inlineCodeStyle, boldStyle, italicStyle)
		result = append(result, formatted)
	}

	// Handle unclosed code block
	if inCodeBlock && len(codeLines) > 0 {
		code := strings.Join(codeLines, "\n")
		codeW := width - 4
		if codeW < 20 {
			codeW = 20
		}
		result = append(result, codeBlockStyle.Width(codeW).Render(code))
	}

	return strings.Join(result, "\n")
}

// formatInline handles inline formatting: `code`, **bold**, *italic*.
func formatInline(text string, codeStyle, boldStyle, italicStyle lipgloss.Style) string {
	// Process inline code first (backticks)
	text = processDelimited(text, "`", "`", codeStyle)
	// Then bold (double asterisk)
	text = processDelimited(text, "**", "**", boldStyle)
	// Then italic (single asterisk, but not double)
	text = processItalic(text, italicStyle)
	return text
}

// processDelimited finds and styles text between delimiter pairs.
func processDelimited(text, open, close string, style lipgloss.Style) string {
	var result strings.Builder
	remaining := text

	for {
		start := strings.Index(remaining, open)
		if start == -1 {
			result.WriteString(remaining)
			break
		}

		after := remaining[start+len(open):]
		end := strings.Index(after, close)
		if end == -1 {
			result.WriteString(remaining)
			break
		}

		// Write text before the delimiter
		result.WriteString(remaining[:start])
		// Write styled content
		content := after[:end]
		result.WriteString(style.Render(content))
		// Move past the closing delimiter
		remaining = after[end+len(close):]
	}

	return result.String()
}

// processItalic handles single asterisk italic, avoiding double asterisks.
func processItalic(text string, style lipgloss.Style) string {
	var result strings.Builder
	remaining := text

	for {
		start := strings.Index(remaining, "*")
		if start == -1 {
			result.WriteString(remaining)
			break
		}

		// Skip if this is a double asterisk (part of bold)
		if start+1 < len(remaining) && remaining[start+1] == '*' {
			result.WriteString(remaining[:start+2])
			remaining = remaining[start+2:]
			continue
		}

		after := remaining[start+1:]
		end := strings.Index(after, "*")
		if end == -1 {
			result.WriteString(remaining)
			break
		}

		// Skip if the closing asterisk is a double
		if end > 0 && after[end-1] == '*' {
			result.WriteString(remaining[:start+1+end+1])
			remaining = after[end+1:]
			continue
		}

		result.WriteString(remaining[:start])
		content := after[:end]
		result.WriteString(style.Render(content))
		remaining = after[end+1:]
	}

	return result.String()
}

func leadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

func isNumberedList(s string) bool {
	for i, ch := range s {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '.' && i > 0 && i < len(s)-1 && s[i+1] == ' ' {
			return true
		}
		return false
	}
	return false
}

func parseNumberedList(s string) (string, string) {
	dotIdx := strings.Index(s, ".")
	if dotIdx == -1 {
		return "", s
	}
	num := s[:dotIdx+1]
	content := strings.TrimSpace(s[dotIdx+1:])
	return num, content
}
