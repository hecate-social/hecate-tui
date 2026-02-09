package llmtools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RegisterCodeExploreTools adds code exploration tools to the registry.
func RegisterCodeExploreTools(r *Registry) {
	r.Register(grepSearchTool(), grepSearchHandler)
	r.Register(symbolSearchTool(), symbolSearchHandler)
	r.Register(codeContextTool(), codeContextHandler)
}

// --- grep_search ---

func grepSearchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("pattern", String("Regular expression pattern to search for"))
	params.AddProperty("path", String("File or directory to search in (default: current directory)"))
	params.AddProperty("glob", String("Glob pattern to filter files (e.g., '*.go', '*.{ts,tsx}')"))
	params.AddProperty("context_lines", Integer("Number of lines to show before/after each match (default: 0)"))
	params.AddProperty("case_insensitive", Boolean("Perform case-insensitive search (default: false)"))
	params.AddProperty("limit", Integer("Maximum number of matches to return (default: 50)"))
	params.AddRequired("pattern")

	return Tool{
		Name:             "grep_search",
		Description:      "Search for a pattern in files using regular expressions. Similar to grep/ripgrep.",
		Parameters:       params,
		Category:         CategoryCodeExplore,
		RequiresApproval: false,
	}
}

type grepSearchArgs struct {
	Pattern         string `json:"pattern"`
	Path            string `json:"path"`
	Glob            string `json:"glob"`
	ContextLines    int    `json:"context_lines"`
	CaseInsensitive bool   `json:"case_insensitive"`
	Limit           int    `json:"limit"`
}

func grepSearchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a grepSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	searchPath := a.Path
	if searchPath == "" {
		if p, err := os.Getwd(); err == nil {
			searchPath = p
		} else {
			searchPath = "/"
		}
	}
	searchPath = expandHomePath(searchPath)

	limit := a.Limit
	if limit <= 0 {
		limit = 50
	}

	// Try ripgrep first, fall back to manual search
	if rgPath, err := exec.LookPath("rg"); err == nil {
		return grepWithRipgrep(ctx, rgPath, a, searchPath, limit)
	}

	return grepManual(ctx, a, searchPath, limit)
}

func grepWithRipgrep(ctx context.Context, rgPath string, a grepSearchArgs, searchPath string, limit int) (string, error) {
	args := []string{
		"--line-number",
		"--no-heading",
		"--max-count", fmt.Sprintf("%d", limit),
	}

	if a.CaseInsensitive {
		args = append(args, "-i")
	}

	if a.ContextLines > 0 {
		args = append(args, "-C", fmt.Sprintf("%d", a.ContextLines))
	}

	if a.Glob != "" {
		args = append(args, "-g", a.Glob)
	}

	args = append(args, a.Pattern, searchPath)

	cmd := exec.CommandContext(ctx, rgPath, args...)
	output, err := cmd.Output()
	if err != nil {
		// Exit code 1 means no matches, which is fine
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return fmt.Sprintf("No matches found for pattern: %s", a.Pattern), nil
		}
		return "", fmt.Errorf("ripgrep error: %w", err)
	}

	if len(output) == 0 {
		return fmt.Sprintf("No matches found for pattern: %s", a.Pattern), nil
	}

	return fmt.Sprintf("Search results for '%s':\n\n%s", a.Pattern, string(output)), nil
}

func grepManual(ctx context.Context, a grepSearchArgs, searchPath string, limit int) (string, error) {
	flags := ""
	if a.CaseInsensitive {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + a.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	var matches []string
	matchCount := 0

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if matchCount >= limit {
			return filepath.SkipAll
		}

		// Apply glob filter if specified
		if a.Glob != "" {
			matched, _ := filepath.Match(a.Glob, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// Skip binary files (simple check)
		if isBinaryFile(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		rel, _ := filepath.Rel(searchPath, path)
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if re.MatchString(line) {
				matches = append(matches, fmt.Sprintf("%s:%d: %s", rel, lineNum, line))
				matchCount++

				if matchCount >= limit {
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("search error: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for pattern: %s", a.Pattern), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches for '%s':\n\n", len(matches), a.Pattern))
	for _, m := range matches {
		sb.WriteString(m + "\n")
	}

	return sb.String(), nil
}

// --- symbol_search ---

func symbolSearchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("symbol", String("Symbol name to search for (function, type, variable)"))
	params.AddProperty("type", Enum("Type of symbol to search for", "function", "type", "variable", "any"))
	params.AddProperty("path", String("Directory to search in (default: current directory)"))
	params.AddProperty("language", Enum("Programming language hint", "go", "rust", "typescript", "javascript", "python", "erlang", "elixir"))
	params.AddRequired("symbol")

	return Tool{
		Name:             "symbol_search",
		Description:      "Find function, type, or variable definitions in code. Uses language-aware patterns.",
		Parameters:       params,
		Category:         CategoryCodeExplore,
		RequiresApproval: false,
	}
}

type symbolSearchArgs struct {
	Symbol   string `json:"symbol"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

func symbolSearchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a symbolSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	searchPath := a.Path
	if searchPath == "" {
		if p, err := os.Getwd(); err == nil {
			searchPath = p
		} else {
			searchPath = "/"
		}
	}
	searchPath = expandHomePath(searchPath)

	symbolType := a.Type
	if symbolType == "" {
		symbolType = "any"
	}

	// Build regex patterns based on language and type
	patterns := buildSymbolPatterns(a.Symbol, symbolType, a.Language)

	var matches []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Filter by language extension
		ext := filepath.Ext(path)
		if a.Language != "" && !matchesLanguage(ext, a.Language) {
			return nil
		}

		// Skip non-code files
		if !isCodeFile(ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		lines := strings.Split(content, "\n")
		rel, _ := filepath.Rel(searchPath, path)

		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			for i, line := range lines {
				if re.MatchString(line) {
					matches = append(matches, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(line)))
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("search error: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No definitions found for symbol: %s", a.Symbol), nil
	}

	// Deduplicate matches
	seen := make(map[string]bool)
	unique := []string{}
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d definitions for '%s':\n\n", len(unique), a.Symbol))
	for _, m := range unique {
		sb.WriteString(m + "\n")
	}

	return sb.String(), nil
}

func buildSymbolPatterns(symbol, symbolType, language string) []string {
	var patterns []string
	escaped := regexp.QuoteMeta(symbol)

	switch symbolType {
	case "function":
		patterns = functionPatterns(escaped, language)
	case "type":
		patterns = typePatterns(escaped, language)
	case "variable":
		patterns = variablePatterns(escaped, language)
	default:
		patterns = append(patterns, functionPatterns(escaped, language)...)
		patterns = append(patterns, typePatterns(escaped, language)...)
		patterns = append(patterns, variablePatterns(escaped, language)...)
	}

	return patterns
}

func functionPatterns(symbol, language string) []string {
	switch language {
	case "go":
		return []string{`func\s+` + symbol + `\s*\(`}
	case "rust":
		return []string{`fn\s+` + symbol + `\s*[<(]`}
	case "typescript", "javascript":
		return []string{
			`function\s+` + symbol + `\s*[<(]`,
			`(const|let|var)\s+` + symbol + `\s*=\s*(async\s+)?\([^)]*\)\s*=>`,
			symbol + `\s*:\s*(async\s+)?\([^)]*\)\s*=>`,
		}
	case "python":
		return []string{`def\s+` + symbol + `\s*\(`}
	case "erlang":
		return []string{`^` + symbol + `\s*\(`}
	case "elixir":
		return []string{`def(p)?\s+` + symbol + `[(\s]`}
	default:
		return []string{
			`func\s+` + symbol + `\s*\(`,
			`fn\s+` + symbol + `\s*[<(]`,
			`function\s+` + symbol + `\s*[<(]`,
			`def\s+` + symbol + `\s*\(`,
		}
	}
}

func typePatterns(symbol, language string) []string {
	switch language {
	case "go":
		return []string{`type\s+` + symbol + `\s+(struct|interface|=)`}
	case "rust":
		return []string{
			`(struct|enum|trait|type)\s+` + symbol + `[<\s{]`,
		}
	case "typescript", "javascript":
		return []string{
			`(interface|type|class)\s+` + symbol + `[<\s{]`,
		}
	case "python":
		return []string{`class\s+` + symbol + `[(\s:]`}
	case "erlang":
		return []string{`-record\(` + symbol + `,`}
	case "elixir":
		return []string{`defmodule\s+` + symbol + `\s+do`}
	default:
		return []string{
			`type\s+` + symbol,
			`(struct|enum|trait|interface|class)\s+` + symbol,
		}
	}
}

func variablePatterns(symbol, language string) []string {
	switch language {
	case "go":
		return []string{
			`var\s+` + symbol + `\s+`,
			symbol + `\s*:=`,
		}
	case "rust":
		return []string{
			`(let|const|static)\s+(mut\s+)?` + symbol + `\s*[=:]`,
		}
	case "typescript", "javascript":
		return []string{
			`(const|let|var)\s+` + symbol + `\s*[=:]`,
		}
	case "python":
		return []string{
			`^` + symbol + `\s*=`,
		}
	default:
		return []string{
			`(var|let|const)\s+` + symbol + `\s*[=:]`,
		}
	}
}

func matchesLanguage(ext, language string) bool {
	extMap := map[string][]string{
		"go":         {".go"},
		"rust":       {".rs"},
		"typescript": {".ts", ".tsx"},
		"javascript": {".js", ".jsx", ".mjs"},
		"python":     {".py"},
		"erlang":     {".erl", ".hrl"},
		"elixir":     {".ex", ".exs"},
	}

	if exts, ok := extMap[language]; ok {
		for _, e := range exts {
			if ext == e {
				return true
			}
		}
	}
	return false
}

func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go": true, ".rs": true, ".ts": true, ".tsx": true,
		".js": true, ".jsx": true, ".mjs": true, ".py": true,
		".erl": true, ".hrl": true, ".ex": true, ".exs": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".java": true, ".kt": true, ".scala": true, ".rb": true,
		".php": true, ".cs": true, ".swift": true, ".m": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true,
	}
	return codeExts[ext]
}

// --- code_context ---

func codeContextTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("path", String("Absolute path to the file"))
	params.AddProperty("line", Integer("Line number to center the context on"))
	params.AddProperty("context_lines", Integer("Number of lines before/after to include (default: 10)"))
	params.AddRequired("path", "line")

	return Tool{
		Name:             "code_context",
		Description:      "Get code surrounding a specific line. Useful for understanding context around a search match.",
		Parameters:       params,
		Category:         CategoryCodeExplore,
		RequiresApproval: false,
	}
}

type codeContextArgs struct {
	Path         string `json:"path"`
	Line         int    `json:"line"`
	ContextLines int    `json:"context_lines"`
}

func codeContextHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a codeContextArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Path == "" || a.Line < 1 {
		return "", fmt.Errorf("path and line are required")
	}

	path := expandHomePath(a.Path)

	contextLines := a.ContextLines
	if contextLines <= 0 {
		contextLines = 10
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	if a.Line > len(lines) {
		return "", fmt.Errorf("line %d is out of range (file has %d lines)", a.Line, len(lines))
	}

	start := a.Line - contextLines - 1
	if start < 0 {
		start = 0
	}

	end := a.Line + contextLines
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s (lines %d-%d)\n\n", a.Path, start+1, end))

	for i := start; i < end; i++ {
		marker := " "
		if i+1 == a.Line {
			marker = ">" // Mark the target line
		}
		sb.WriteString(fmt.Sprintf("%s%5d│ %s\n", marker, i+1, lines[i]))
	}

	return sb.String(), nil
}

// --- Helpers ---

func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return true
	}

	// Check for null bytes (simple binary detection)
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}
