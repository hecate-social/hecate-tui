package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RegisterFileSystemTools adds filesystem tools to the registry.
func RegisterFileSystemTools(r *Registry) {
	r.Register(readFileTool(), readFileHandler)
	r.Register(writeFileTool(), writeFileHandler)
	r.Register(editFileTool(), editFileHandler)
	r.Register(listDirectoryTool(), listDirectoryHandler)
	r.Register(globSearchTool(), globSearchHandler)
}

// --- read_file ---

func readFileTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("path", String("Absolute path to the file to read"))
	params.AddProperty("offset", Integer("Line number to start reading from (1-indexed, default: 1)"))
	params.AddProperty("limit", Integer("Maximum number of lines to read (default: 500)"))
	params.AddRequired("path")

	return Tool{
		Name:             "read_file",
		Description:      "Read the contents of a file. Returns file content with line numbers.",
		Parameters:       params,
		Category:         CategoryFileSystem,
		RequiresApproval: false,
	}
}

type readFileArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func readFileHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a readFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Expand ~ to home directory
	path := expandHomePath(a.Path)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Apply offset and limit
	offset := a.Offset
	if offset < 1 {
		offset = 1
	}
	limit := a.Limit
	if limit <= 0 {
		limit = 500
	}

	start := offset - 1
	if start >= len(lines) {
		return fmt.Sprintf("File has only %d lines, offset %d is out of range", len(lines), offset), nil
	}

	end := start + limit
	if end > len(lines) {
		end = len(lines)
	}

	// Format with line numbers
	var sb strings.Builder
	fmt.Fprintf(&sb, "File: %s (%d lines total)\n", a.Path, len(lines))
	if start > 0 || end < len(lines) {
		fmt.Fprintf(&sb, "Showing lines %d-%d\n", start+1, end)
	}
	sb.WriteString("\n")

	for i := start; i < end; i++ {
		fmt.Fprintf(&sb, "%6d│ %s\n", i+1, lines[i])
	}

	return sb.String(), nil
}

// --- write_file ---

func writeFileTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("path", String("Absolute path to the file to write"))
	params.AddProperty("content", String("Content to write to the file"))
	params.AddRequired("path", "content")

	return Tool{
		Name:             "write_file",
		Description:      "Write content to a file, creating it if it doesn't exist or overwriting if it does.",
		Parameters:       params,
		Category:         CategoryFileSystem,
		RequiresApproval: true,
	}
}

type writeFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func writeFileHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a writeFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	path := expandHomePath(a.Path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(a.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	lines := strings.Count(a.Content, "\n") + 1
	return fmt.Sprintf("Successfully wrote %d bytes (%d lines) to %s", len(a.Content), lines, a.Path), nil
}

// --- edit_file ---

func editFileTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("path", String("Absolute path to the file to edit"))
	params.AddProperty("old_string", String("The exact text to find and replace"))
	params.AddProperty("new_string", String("The replacement text"))
	params.AddProperty("replace_all", Boolean("Replace all occurrences instead of just the first (default: false)"))
	params.AddRequired("path", "old_string", "new_string")

	return Tool{
		Name:             "edit_file",
		Description:      "Find and replace text in a file. The old_string must match exactly (including whitespace).",
		Parameters:       params,
		Category:         CategoryFileSystem,
		RequiresApproval: true,
	}
}

type editFileArgs struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all"`
}

func editFileHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a editFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Path == "" || a.OldString == "" {
		return "", fmt.Errorf("path and old_string are required")
	}

	path := expandHomePath(a.Path)

	// Read current content
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// Check if old_string exists
	count := strings.Count(content, a.OldString)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file")
	}

	// Perform replacement
	var newContent string
	var replacements int
	if a.ReplaceAll {
		newContent = strings.ReplaceAll(content, a.OldString, a.NewString)
		replacements = count
	} else {
		newContent = strings.Replace(content, a.OldString, a.NewString, 1)
		replacements = 1
	}

	// Write back
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if count > 1 && !a.ReplaceAll {
		return fmt.Sprintf("Replaced 1 of %d occurrences in %s", count, a.Path), nil
	}
	return fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replacements, a.Path), nil
}

// --- list_directory ---

func listDirectoryTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("path", String("Absolute path to the directory to list"))
	params.AddProperty("recursive", Boolean("List contents recursively (default: false)"))
	params.AddProperty("show_hidden", Boolean("Show hidden files (default: false)"))
	params.AddRequired("path")

	return Tool{
		Name:             "list_directory",
		Description:      "List files and directories in a path.",
		Parameters:       params,
		Category:         CategoryFileSystem,
		RequiresApproval: false,
	}
}

type listDirectoryArgs struct {
	Path       string `json:"path"`
	Recursive  bool   `json:"recursive"`
	ShowHidden bool   `json:"show_hidden"`
}

func listDirectoryHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a listDirectoryArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	path := expandHomePath(a.Path)

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("path not found: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Directory: %s\n\n", a.Path))

	if a.Recursive {
		_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			rel, _ := filepath.Rel(path, p)
			if rel == "." {
				return nil
			}

			// Skip hidden files unless requested
			if !a.ShowHidden && strings.HasPrefix(filepath.Base(p), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			prefix := ""
			if info.IsDir() {
				prefix = "[dir]  "
			} else {
				prefix = "[file] "
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix, rel))
			return nil
		})
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}

		dirs := []string{}
		files := []string{}

		for _, entry := range entries {
			name := entry.Name()

			// Skip hidden unless requested
			if !a.ShowHidden && strings.HasPrefix(name, ".") {
				continue
			}

			if entry.IsDir() {
				dirs = append(dirs, name+"/")
			} else {
				files = append(files, name)
			}
		}

		// List directories first
		for _, d := range dirs {
			sb.WriteString(fmt.Sprintf("[dir]  %s\n", d))
		}
		for _, f := range files {
			sb.WriteString(fmt.Sprintf("[file] %s\n", f))
		}
	}

	return sb.String(), nil
}

// --- glob_search ---

func globSearchTool() Tool {
	params := NewObjectParameters()
	params.AddProperty("pattern", String("Glob pattern to match (e.g., '**/*.go', 'src/**/*.ts')"))
	params.AddProperty("path", String("Base directory to search in (default: current directory)"))
	params.AddProperty("limit", Integer("Maximum number of results (default: 100)"))
	params.AddRequired("pattern")

	return Tool{
		Name:             "glob_search",
		Description:      "Find files matching a glob pattern. Supports ** for recursive matching.",
		Parameters:       params,
		Category:         CategoryFileSystem,
		RequiresApproval: false,
	}
}

type globSearchArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
	Limit   int    `json:"limit"`
}

func globSearchHandler(ctx context.Context, args json.RawMessage) (string, error) {
	var a globSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if a.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	basePath := a.Path
	if basePath == "" {
		if p, err := os.Getwd(); err == nil {
			basePath = p
		} else {
			basePath = "/"
		}
	}
	basePath = expandHomePath(basePath)

	limit := a.Limit
	if limit <= 0 {
		limit = 100
	}

	var matches []string

	// Simple glob implementation using filepath.Walk
	// Handle ** pattern by walking recursively
	if strings.Contains(a.Pattern, "**") {
		parts := strings.Split(a.Pattern, "**")
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := ""
		if len(parts) > 1 {
			suffix = strings.TrimPrefix(parts[1], "/")
		}

		searchPath := basePath
		if prefix != "" {
			searchPath = filepath.Join(basePath, prefix)
		}

		_ = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			if len(matches) >= limit {
				return filepath.SkipAll
			}

			rel, _ := filepath.Rel(basePath, path)

			// Match suffix pattern
			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				if !matched {
					return nil
				}
			}

			matches = append(matches, rel)
			return nil
		})
	} else {
		// Simple glob
		pattern := filepath.Join(basePath, a.Pattern)
		found, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid glob pattern: %w", err)
		}

		for i, m := range found {
			if i >= limit {
				break
			}
			rel, _ := filepath.Rel(basePath, m)
			matches = append(matches, rel)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No files found matching pattern: %s", a.Pattern), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d files matching '%s':\n\n", len(matches), a.Pattern))
	for _, m := range matches {
		sb.WriteString(m + "\n")
	}

	if len(matches) == limit {
		sb.WriteString(fmt.Sprintf("\n(limited to %d results)", limit))
	}

	return sb.String(), nil
}

// --- Helpers ---

func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
