package commands

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ProjectCmd shows project/workspace information.
type ProjectCmd struct{}

func (c *ProjectCmd) Name() string        { return "project" }
func (c *ProjectCmd) Aliases() []string   { return []string{"proj"} }
func (c *ProjectCmd) Description() string { return "Show workspace and project info" }

func (c *ProjectCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		var b strings.Builder

		b.WriteString(s.CardTitle.Render("Project"))
		b.WriteString("\n\n")

		// Working directory
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		}

		b.WriteString(s.Bold.Render("Workspace"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("Directory: "))
		b.WriteString(s.CardValue.Render(cwd))
		b.WriteString("\n")

		// Check for common project markers
		markers := []struct {
			file string
			kind string
		}{
			{"go.mod", "Go module"},
			{"Cargo.toml", "Rust crate"},
			{"package.json", "Node.js project"},
			{"mix.exs", "Elixir project"},
			{"rebar.config", "Erlang project"},
			{"pyproject.toml", "Python project"},
			{"requirements.txt", "Python project"},
			{"Makefile", "Makefile"},
			{"Dockerfile", "Docker"},
			{"docker-compose.yml", "Docker Compose"},
			{"docker-compose.yaml", "Docker Compose"},
			{".git", "Git repository"},
		}

		var detected []string
		for _, m := range markers {
			path := filepath.Join(cwd, m.file)
			if _, err := os.Stat(path); err == nil {
				detected = append(detected, m.kind)
			}
		}

		if len(detected) > 0 {
			b.WriteString(s.CardLabel.Render("Type: "))
			b.WriteString(s.CardValue.Render(detected[0]))
			b.WriteString("\n")

			if len(detected) > 1 {
				b.WriteString(s.CardLabel.Render("Also: "))
				b.WriteString(s.Subtle.Render(strings.Join(detected[1:], ", ")))
				b.WriteString("\n")
			}
		}

		// Read go.mod for module name if available
		gomod := filepath.Join(cwd, "go.mod")
		if data, err := os.ReadFile(gomod); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					module := strings.TrimPrefix(line, "module ")
					b.WriteString(s.CardLabel.Render("Module: "))
					b.WriteString(s.CardValue.Render(module))
					b.WriteString("\n")
					break
				}
			}
		}

		// Read mix.exs for app name if available
		mixexs := filepath.Join(cwd, "mix.exs")
		if data, err := os.ReadFile(mixexs); err == nil {
			content := string(data)
			if idx := strings.Index(content, "app:"); idx != -1 {
				after := content[idx+4:]
				after = strings.TrimLeft(after, " :")
				end := strings.IndexAny(after, ",\n])")
				if end > 0 {
					appName := strings.TrimSpace(after[:end])
					b.WriteString(s.CardLabel.Render("App: "))
					b.WriteString(s.CardValue.Render(appName))
					b.WriteString("\n")
				}
			}
		}

		b.WriteString("\n")

		// Daemon connection
		b.WriteString(s.Bold.Render("Daemon"))
		b.WriteString("\n")

		health, err := ctx.Client.GetHealth()
		if err != nil {
			b.WriteString(s.CardLabel.Render("Status: "))
			b.WriteString(s.Error.Render("unreachable"))
		} else {
			b.WriteString(s.CardLabel.Render("Status: "))
			b.WriteString(s.StatusOK.Render(health.Status))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("Version: "))
			b.WriteString(s.CardValue.Render(health.Version))
		}

		return InjectSystemMsg{Content: b.String()}
	}
}
