package tools

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Launcher handles launching external tools
type Launcher struct {
	config *Config
}

// NewLauncher creates a tool launcher
func NewLauncher(cfg *Config) *Launcher {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Launcher{config: cfg}
}

// LaunchResult is returned after a tool exits
type LaunchResult struct {
	Tool     string
	ExitCode int
	Err      error
}

// LaunchEditor opens a file in the preferred editor
func (l *Launcher) LaunchEditor(path string) tea.Cmd {
	editor := l.getEditor()
	if editor == "" {
		return func() tea.Msg {
			return LaunchResult{Tool: "editor", ExitCode: 1, Err: os.ErrNotExist}
		}
	}

	args := append(l.config.Editor.Args, path)
	return l.launch(editor, args...)
}

// LaunchEditorAtLine opens a file at a specific line
func (l *Launcher) LaunchEditorAtLine(path string, line int) tea.Cmd {
	editor := l.getEditor()
	if editor == "" {
		return func() tea.Msg {
			return LaunchResult{Tool: "editor", ExitCode: 1, Err: os.ErrNotExist}
		}
	}

	// Build line argument based on editor
	var args []string
	args = append(args, l.config.Editor.Args...)

	switch editor {
	case "nvim", "vim", "vi":
		args = append(args, "+"+itoa(line), path)
	case "code":
		args = append(args, "--goto", path+":"+itoa(line))
	case "hx":
		args = append(args, path+":"+itoa(line))
	case "emacs":
		args = append(args, "+"+itoa(line), path)
	default:
		args = append(args, path)
	}

	return l.launch(editor, args...)
}

// LaunchTool opens a specific tool
func (l *Launcher) LaunchTool(tool Tool, args ...string) tea.Cmd {
	cmd := tool.Command
	if override, ok := l.config.Tools.Paths[tool.Name]; ok {
		cmd = override
	}
	return l.launch(cmd, args...)
}

// LaunchCommand runs an arbitrary command
func (l *Launcher) LaunchCommand(cmd string, args ...string) tea.Cmd {
	return l.launch(cmd, args...)
}

func (l *Launcher) launch(cmd string, args ...string) tea.Cmd {
	c := exec.Command(cmd, args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}
		return LaunchResult{Tool: cmd, ExitCode: exitCode, Err: err}
	})
}

func (l *Launcher) getEditor() string {
	// Check config preference first
	if l.config.Editor.Preferred != "" {
		if _, err := exec.LookPath(l.config.Editor.Preferred); err == nil {
			return l.config.Editor.Preferred
		}
	}

	// Check EDITOR env var
	if editor := os.Getenv("EDITOR"); editor != "" {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	// Check VISUAL env var
	if visual := os.Getenv("VISUAL"); visual != "" {
		if _, err := exec.LookPath(visual); err == nil {
			return visual
		}
	}

	// Fall back to detection
	detector := NewDetector()
	if editor := detector.GetPreferredEditor(); editor != nil {
		return editor.Command
	}

	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
