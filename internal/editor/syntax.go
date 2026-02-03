package editor

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Language represents a programming language
type Language string

const (
	LangPlain    Language = "plain"
	LangGo       Language = "go"
	LangRust     Language = "rust"
	LangPython   Language = "python"
	LangJS       Language = "javascript"
	LangTS       Language = "typescript"
	LangElixir   Language = "elixir"
	LangErlang   Language = "erlang"
	LangMarkdown Language = "markdown"
	LangYAML     Language = "yaml"
	LangTOML     Language = "toml"
	LangJSON     Language = "json"
	LangShell    Language = "shell"
)

// Syntax colors
var (
	KeywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")) // Purple
	TypeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")) // Yellow
	StringStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")) // Green
	CommentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")) // Gray
	NumberStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D19A66")) // Orange
	FunctionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")) // Blue
	OperatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#56B6C2")) // Cyan
)

// Keywords by language
var keywords = map[Language][]string{
	LangGo:     {"func", "package", "import", "var", "const", "type", "struct", "interface", "map", "chan", "if", "else", "for", "range", "switch", "case", "default", "return", "defer", "go", "select", "break", "continue", "fallthrough", "nil", "true", "false", "make", "new", "append", "len", "cap", "error"},
	LangRust:   {"fn", "let", "mut", "const", "static", "struct", "enum", "impl", "trait", "pub", "mod", "use", "if", "else", "match", "for", "while", "loop", "return", "break", "continue", "self", "Self", "true", "false", "None", "Some", "Ok", "Err"},
	LangPython: {"def", "class", "import", "from", "as", "if", "elif", "else", "for", "while", "try", "except", "finally", "with", "return", "yield", "raise", "pass", "break", "continue", "and", "or", "not", "in", "is", "None", "True", "False", "lambda", "async", "await"},
	LangJS:     {"function", "const", "let", "var", "class", "extends", "import", "export", "from", "if", "else", "for", "while", "switch", "case", "default", "return", "break", "continue", "try", "catch", "finally", "throw", "new", "this", "true", "false", "null", "undefined", "async", "await", "yield"},
	LangTS:     {"function", "const", "let", "var", "class", "extends", "implements", "interface", "type", "import", "export", "from", "if", "else", "for", "while", "switch", "case", "default", "return", "break", "continue", "try", "catch", "finally", "throw", "new", "this", "true", "false", "null", "undefined", "async", "await", "yield", "public", "private", "protected", "readonly"},
	LangElixir: {"def", "defp", "defmodule", "defstruct", "defimpl", "defprotocol", "do", "end", "if", "else", "unless", "case", "cond", "with", "for", "fn", "receive", "after", "try", "catch", "rescue", "raise", "import", "alias", "require", "use", "true", "false", "nil", "when", "and", "or", "not", "in"},
	LangErlang: {"module", "export", "import", "define", "include", "record", "if", "case", "of", "end", "receive", "after", "try", "catch", "throw", "fun", "when", "true", "false", "undefined", "ok", "error"},
	LangShell:  {"if", "then", "else", "elif", "fi", "case", "esac", "for", "while", "do", "done", "in", "function", "return", "exit", "local", "export", "source", "true", "false"},
}

// DetectLanguage guesses language from filename
func DetectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.ToLower(filepath.Base(filename))

	switch ext {
	case ".go":
		return LangGo
	case ".rs":
		return LangRust
	case ".py":
		return LangPython
	case ".js", ".jsx", ".mjs":
		return LangJS
	case ".ts", ".tsx":
		return LangTS
	case ".ex", ".exs":
		return LangElixir
	case ".erl", ".hrl":
		return LangErlang
	case ".md", ".markdown":
		return LangMarkdown
	case ".yaml", ".yml":
		return LangYAML
	case ".toml":
		return LangTOML
	case ".json":
		return LangJSON
	case ".sh", ".bash", ".zsh":
		return LangShell
	}

	// Check basename
	switch base {
	case "makefile", "dockerfile":
		return LangShell
	}

	return LangPlain
}

// Highlighter provides syntax highlighting
type Highlighter struct {
	lang Language
}

// NewHighlighter creates a highlighter for a language
func NewHighlighter(lang Language) *Highlighter {
	return &Highlighter{lang: lang}
}

// HighlightLine applies syntax highlighting to a line
func (h *Highlighter) HighlightLine(line string) string {
	if h.lang == LangPlain {
		return line
	}

	// Check for comments first
	trimmed := strings.TrimSpace(line)
	if h.isComment(trimmed) {
		return CommentStyle.Render(line)
	}

	// Simple token-based highlighting
	return h.highlightTokens(line)
}

func (h *Highlighter) isComment(line string) bool {
	switch h.lang {
	case LangGo, LangRust, LangJS, LangTS:
		return strings.HasPrefix(line, "//")
	case LangPython, LangShell, LangYAML, LangTOML:
		return strings.HasPrefix(line, "#")
	case LangElixir:
		return strings.HasPrefix(line, "#")
	case LangErlang:
		return strings.HasPrefix(line, "%")
	}
	return false
}

func (h *Highlighter) highlightTokens(line string) string {
	kws, ok := keywords[h.lang]
	if !ok {
		return line
	}

	result := line

	// Highlight strings first (simple approach)
	inString := false
	stringChar := byte(0)
	var highlighted strings.Builder
	i := 0

	for i < len(line) {
		ch := line[i]

		// Handle string literals
		if !inString && (ch == '"' || ch == '\'' || ch == '`') {
			inString = true
			stringChar = ch
			start := i
			i++
			for i < len(line) && (line[i] != stringChar || (i > 0 && line[i-1] == '\\')) {
				i++
			}
			if i < len(line) {
				i++ // include closing quote
			}
			highlighted.WriteString(StringStyle.Render(line[start:i]))
			continue
		}

		// Handle identifiers/keywords
		if isAlpha(ch) {
			start := i
			for i < len(line) && (isAlpha(line[i]) || isDigit(line[i])) {
				i++
			}
			word := line[start:i]
			if contains(kws, word) {
				highlighted.WriteString(KeywordStyle.Render(word))
			} else {
				highlighted.WriteString(word)
			}
			continue
		}

		// Handle numbers
		if isDigit(ch) {
			start := i
			for i < len(line) && (isDigit(line[i]) || line[i] == '.' || line[i] == 'x' || line[i] == 'X') {
				i++
			}
			highlighted.WriteString(NumberStyle.Render(line[start:i]))
			continue
		}

		// Regular character
		highlighted.WriteByte(ch)
		i++
	}

	if highlighted.Len() > 0 {
		result = highlighted.String()
	}

	return result
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
