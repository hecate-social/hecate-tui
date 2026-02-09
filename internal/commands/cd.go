package commands

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ChangeDirMsg tells the app to change its working directory.
type ChangeDirMsg struct {
	Path string
}

// CdCmd handles /cd - change working directory.
type CdCmd struct{}

func (c *CdCmd) Name() string        { return "cd" }
func (c *CdCmd) Aliases() []string   { return nil }
func (c *CdCmd) Description() string { return "Change working directory" }

// Complete implements Completable for directory completion.
func (c *CdCmd) Complete(args []string, ctx *Context) []string {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}

	return c.completeDirs(prefix)
}

func (c *CdCmd) Execute(args []string, ctx *Context) tea.Cmd {
	s := ctx.Styles

	if len(args) == 0 {
		// No args - show current directory
		cwd, err := os.Getwd()
		if err != nil {
			return func() tea.Msg {
				return InjectSystemMsg{Content: s.Error.Render("Failed to get current directory: " + err.Error())}
			}
		}
		return func() tea.Msg {
			return InjectSystemMsg{Content: s.Subtle.Render("Current directory: " + cwd)}
		}
	}

	// Expand ~ to home directory
	path := args[0]
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	// Make absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return func() tea.Msg {
			return InjectSystemMsg{Content: s.Error.Render("Invalid path: " + err.Error())}
		}
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return func() tea.Msg {
				return InjectSystemMsg{Content: s.Error.Render("Directory does not exist: " + absPath)}
			}
		}
		return func() tea.Msg {
			return InjectSystemMsg{Content: s.Error.Render("Cannot access: " + err.Error())}
		}
	}

	if !info.IsDir() {
		return func() tea.Msg {
			return InjectSystemMsg{Content: s.Error.Render("Not a directory: " + absPath)}
		}
	}

	// Send message to change directory
	return func() tea.Msg {
		return ChangeDirMsg{Path: absPath}
	}
}

// completeDirs returns directory completions for the given prefix.
func (c *CdCmd) completeDirs(prefix string) []string {
	// Expand ~
	searchPath := prefix
	if strings.HasPrefix(searchPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			searchPath = filepath.Join(home, searchPath[1:])
		}
	}

	// If empty, use current directory
	if searchPath == "" {
		searchPath = "."
	}

	// Get directory to search in
	dir := searchPath
	base := ""

	info, err := os.Stat(searchPath)
	if err != nil || !info.IsDir() {
		// Path doesn't exist or isn't a dir - search parent
		dir = filepath.Dir(searchPath)
		base = filepath.Base(searchPath)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue // Skip hidden unless prefix starts with .
		}
		if base != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(base)) {
			continue
		}

		// Build the full completion path
		fullPath := filepath.Join(dir, name)

		// If original prefix used ~, convert back
		if strings.HasPrefix(prefix, "~") {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(fullPath, home) {
				fullPath = "~" + fullPath[len(home):]
			}
		}

		matches = append(matches, fullPath+"/")
	}

	return matches
}
