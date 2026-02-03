package tools

import (
	"os/exec"
	"strings"
)

// Tool represents an external tool
type Tool struct {
	Name        string
	Command     string
	Args        []string
	Description string
	Category    ToolCategory
	Installed   bool
	Version     string
}

// ToolCategory groups related tools
type ToolCategory string

const (
	CategoryEditor    ToolCategory = "editor"
	CategoryTerminal  ToolCategory = "terminal"
	CategoryVCS       ToolCategory = "vcs"
	CategoryBuild     ToolCategory = "build"
	CategoryContainer ToolCategory = "container"
	CategoryLLM       ToolCategory = "llm"
)

// KnownTools defines tools we look for
var KnownTools = []Tool{
	// Editors
	{Name: "Neovim", Command: "nvim", Args: []string{}, Description: "Hyperextensible Vim-based editor", Category: CategoryEditor},
	{Name: "Vim", Command: "vim", Args: []string{}, Description: "Vi Improved editor", Category: CategoryEditor},
	{Name: "VS Code", Command: "code", Args: []string{}, Description: "Visual Studio Code", Category: CategoryEditor},
	{Name: "Helix", Command: "hx", Args: []string{}, Description: "Post-modern modal editor", Category: CategoryEditor},
	{Name: "Emacs", Command: "emacs", Args: []string{}, Description: "Extensible text editor", Category: CategoryEditor},
	{Name: "Nano", Command: "nano", Args: []string{}, Description: "Simple terminal editor", Category: CategoryEditor},

	// Terminals
	{Name: "Kitty", Command: "kitty", Args: []string{}, Description: "GPU-based terminal", Category: CategoryTerminal},
	{Name: "Alacritty", Command: "alacritty", Args: []string{}, Description: "GPU-accelerated terminal", Category: CategoryTerminal},
	{Name: "WezTerm", Command: "wezterm", Args: []string{}, Description: "GPU-accelerated terminal", Category: CategoryTerminal},

	// VCS
	{Name: "Git", Command: "git", Args: []string{}, Description: "Distributed version control", Category: CategoryVCS},
	{Name: "GitHub CLI", Command: "gh", Args: []string{}, Description: "GitHub on the command line", Category: CategoryVCS},
	{Name: "LazyGit", Command: "lazygit", Args: []string{}, Description: "Terminal UI for git", Category: CategoryVCS},

	// Build
	{Name: "Go", Command: "go", Args: []string{}, Description: "Go programming language", Category: CategoryBuild},
	{Name: "Rust", Command: "rustc", Args: []string{}, Description: "Rust compiler", Category: CategoryBuild},
	{Name: "Node.js", Command: "node", Args: []string{}, Description: "JavaScript runtime", Category: CategoryBuild},
	{Name: "Python", Command: "python3", Args: []string{}, Description: "Python interpreter", Category: CategoryBuild},
	{Name: "Elixir", Command: "elixir", Args: []string{}, Description: "Elixir language", Category: CategoryBuild},
	{Name: "Erlang", Command: "erl", Args: []string{}, Description: "Erlang runtime", Category: CategoryBuild},

	// Containers
	{Name: "Docker", Command: "docker", Args: []string{}, Description: "Container runtime", Category: CategoryContainer},
	{Name: "Podman", Command: "podman", Args: []string{}, Description: "Daemonless containers", Category: CategoryContainer},
	{Name: "kubectl", Command: "kubectl", Args: []string{}, Description: "Kubernetes CLI", Category: CategoryContainer},

	// LLM
	{Name: "Ollama", Command: "ollama", Args: []string{}, Description: "Local LLM runner", Category: CategoryLLM},
	{Name: "Claude Code", Command: "claude", Args: []string{}, Description: "Claude CLI", Category: CategoryLLM},
}

// Detector checks for installed tools
type Detector struct {
	tools []Tool
}

// NewDetector creates a tool detector
func NewDetector() *Detector {
	return &Detector{
		tools: make([]Tool, len(KnownTools)),
	}
}

// Detect checks which tools are installed
func (d *Detector) Detect() []Tool {
	copy(d.tools, KnownTools)

	for i := range d.tools {
		d.tools[i].Installed = d.isInstalled(d.tools[i].Command)
		if d.tools[i].Installed {
			d.tools[i].Version = d.getVersion(d.tools[i].Command)
		}
	}

	return d.tools
}

// DetectByCategory returns tools in a category
func (d *Detector) DetectByCategory(cat ToolCategory) []Tool {
	all := d.Detect()
	var result []Tool
	for _, t := range all {
		if t.Category == cat {
			result = append(result, t)
		}
	}
	return result
}

// GetInstalledEditors returns installed editors in preference order
func (d *Detector) GetInstalledEditors() []Tool {
	editors := d.DetectByCategory(CategoryEditor)
	var installed []Tool
	for _, e := range editors {
		if e.Installed {
			installed = append(installed, e)
		}
	}
	return installed
}

// GetPreferredEditor returns the first installed editor
func (d *Detector) GetPreferredEditor() *Tool {
	editors := d.GetInstalledEditors()
	if len(editors) > 0 {
		return &editors[0]
	}
	return nil
}

func (d *Detector) isInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (d *Detector) getVersion(cmd string) string {
	// Try common version flags
	for _, flag := range []string{"--version", "-version", "version"} {
		out, err := exec.Command(cmd, flag).Output()
		if err == nil {
			// Return first line, trimmed
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				v := strings.TrimSpace(lines[0])
				// Truncate if too long
				if len(v) > 50 {
					v = v[:47] + "..."
				}
				return v
			}
		}
	}
	return ""
}

// CategoryName returns human-readable category name
func CategoryName(cat ToolCategory) string {
	switch cat {
	case CategoryEditor:
		return "Editors"
	case CategoryTerminal:
		return "Terminals"
	case CategoryVCS:
		return "Version Control"
	case CategoryBuild:
		return "Build Tools"
	case CategoryContainer:
		return "Containers"
	case CategoryLLM:
		return "AI/LLM"
	default:
		return string(cat)
	}
}

// CategoryIcon returns an icon for the category
func CategoryIcon(cat ToolCategory) string {
	switch cat {
	case CategoryEditor:
		return "📝"
	case CategoryTerminal:
		return "💻"
	case CategoryVCS:
		return "📊"
	case CategoryBuild:
		return "🔧"
	case CategoryContainer:
		return "📦"
	case CategoryLLM:
		return "🤖"
	default:
		return "🔹"
	}
}
