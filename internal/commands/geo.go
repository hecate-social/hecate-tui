package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/geo"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// GeoCmd shows the current geo-restriction status.
type GeoCmd struct{}

func (c *GeoCmd) Name() string        { return "geo" }
func (c *GeoCmd) Aliases() []string   { return []string{"location"} }
func (c *GeoCmd) Description() string { return "Show geo-restriction status" }

func (c *GeoCmd) Execute(args []string, ctx *Context) tea.Cmd {
	return func() tea.Msg {
		s := ctx.Styles

		var b strings.Builder
		b.WriteString(s.Label.Render("Geo-Restriction Status"))
		b.WriteString("\n\n")

		// Try local check first
		checker, localErr := geo.NewChecker()
		if localErr == nil {
			defer func() { _ = checker.Close() }()
			result, err := checker.CheckPublicIP()
			if err == nil {
				b.WriteString(formatLabel(s, "Local Check", 14))
				if result.Allowed {
					b.WriteString(s.StatusOK.Render("allowed"))
				} else {
					b.WriteString(s.StatusError.Render("blocked"))
				}
				b.WriteString("\n")

				if result.IP != "" {
					b.WriteString(formatLabel(s, "Public IP", 14))
					b.WriteString(s.Value.Render(result.IP))
					b.WriteString("\n")
				}

				if result.CountryCode != "" {
					location := result.CountryCode
					if result.CountryName != "" {
						location = result.CountryName + " (" + result.CountryCode + ")"
					}
					b.WriteString(formatLabel(s, "Location", 14))
					b.WriteString(s.Value.Render(location))
					b.WriteString("\n")
				}
			}
		}

		// Also check daemon status
		b.WriteString("\n")
		daemonStatus, err := geo.CheckWithDaemon(ctx.SocketPath, ctx.HTTPUrl)
		if err != nil {
			b.WriteString(formatLabel(s, "Daemon Check", 14))
			b.WriteString(s.StatusError.Render("unavailable"))
			b.WriteString(" ")
			b.WriteString(s.Subtle.Render("(" + err.Error() + ")"))
		} else {
			b.WriteString(formatLabel(s, "Daemon Check", 14))
			if daemonStatus.Allowed {
				b.WriteString(s.StatusOK.Render("allowed"))
			} else {
				b.WriteString(s.StatusError.Render("blocked"))
				if daemonStatus.CountryCode != "" {
					b.WriteString(" ")
					b.WriteString(s.Subtle.Render("(" + daemonStatus.CountryCode + ")"))
				}
			}
		}

		return InjectSystemMsg{Content: b.String()}
	}
}

// formatLabel creates a right-aligned label with consistent width.
func formatLabel(s *theme.Styles, label string, width int) string {
	padded := strings.Repeat(" ", width-len(label)) + label + ": "
	return s.Label.Render(padded)
}
