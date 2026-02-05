package commands

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/config"
)

// ConfigCmd shows current configuration.
type ConfigCmd struct{}

func (c *ConfigCmd) Name() string        { return "config" }
func (c *ConfigCmd) Aliases() []string   { return nil }
func (c *ConfigCmd) Description() string { return "Show current configuration" }

func (c *ConfigCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles
		var b strings.Builder

		b.WriteString(s.CardTitle.Render("Configuration"))
		b.WriteString("\n\n")

		// Connection
		b.WriteString(s.Bold.Render("Connection"))
		b.WriteString("\n")

		// Show socket or TCP connection info
		socketPath := os.Getenv("HECATE_SOCKET")
		if socketPath == "" {
			cfg := config.Load()
			socketPath = cfg.Connection.SocketPath
		}

		if socketPath != "" {
			b.WriteString(s.CardLabel.Render("  Socket:     "))
			if _, err := os.Stat(socketPath); err == nil {
				b.WriteString(s.StatusOK.Render(socketPath))
			} else {
				b.WriteString(s.Subtle.Render(socketPath + " (not found)"))
			}
			b.WriteString("\n")
		}

		hecateURL := os.Getenv("HECATE_URL")
		if hecateURL == "" {
			hecateURL = "http://localhost:4444"
		}
		b.WriteString(s.CardLabel.Render("  Daemon URL: "))
		b.WriteString(s.CardValue.Render(hecateURL))
		b.WriteString("\n")

		// Daemon health
		health, err := ctx.Client.GetHealth()
		if err != nil {
			b.WriteString(s.CardLabel.Render("  Status:     "))
			b.WriteString(s.Error.Render("unreachable"))
		} else {
			b.WriteString(s.CardLabel.Render("  Status:     "))
			b.WriteString(s.StatusOK.Render(health.Status))
			b.WriteString("\n")
			b.WriteString(s.CardLabel.Render("  Version:    "))
			b.WriteString(s.CardValue.Render(health.Version))
		}
		b.WriteString("\n\n")

		// Theme
		b.WriteString(s.Bold.Render("Theme"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Active: "))
		b.WriteString(s.CardValue.Render(ctx.Theme.Name))
		b.WriteString("\n\n")

		// Terminal
		b.WriteString(s.Bold.Render("Terminal"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Size:   "))
		b.WriteString(s.CardValue.Render(fmt.Sprintf("%dx%d", ctx.Width, ctx.Height)))
		b.WriteString("\n")
		term := os.Getenv("TERM")
		if term != "" {
			b.WriteString(s.CardLabel.Render("  TERM:   "))
			b.WriteString(s.CardValue.Render(term))
			b.WriteString("\n")
		}
		termProg := os.Getenv("TERM_PROGRAM")
		if termProg != "" {
			b.WriteString(s.CardLabel.Render("  App:    "))
			b.WriteString(s.CardValue.Render(termProg))
			b.WriteString("\n")
		}
		colorTerm := os.Getenv("COLORTERM")
		if colorTerm != "" {
			b.WriteString(s.CardLabel.Render("  Color:  "))
			b.WriteString(s.CardValue.Render(colorTerm))
			b.WriteString("\n")
		}

		// Config file
		b.WriteString("\n")
		b.WriteString(s.Bold.Render("Config File"))
		b.WriteString("\n")
		b.WriteString(s.CardLabel.Render("  Path:   "))
		b.WriteString(s.CardValue.Render(config.DefaultPath()))

		return InjectSystemMsg{Content: b.String()}
	}
}
