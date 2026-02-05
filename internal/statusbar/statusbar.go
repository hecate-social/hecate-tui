package statusbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// Model is the status bar — always visible at the bottom.
type Model struct {
	theme  *theme.Theme
	styles *theme.Styles
	width  int

	Mode         modes.Mode
	ModelName    string
	MeshStatus   string // "connected", "disconnected", "unknown"
	DaemonStatus string // "healthy", "degraded", "error", "unknown"
	InputLen     int    // character count for Insert mode
}

// New creates a new status bar.
func New(t *theme.Theme, s *theme.Styles) Model {
	return Model{
		theme:        t,
		styles:       s,
		MeshStatus:   "unknown",
		DaemonStatus: "unknown",
	}
}

// SetWidth updates the status bar width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// View renders the status bar.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Mode label — each mode gets its own color
	modeStyle := m.modeStyle()
	modeLabel := modeStyle.Render(" " + m.Mode.String() + " ")

	// Model indicator
	modelSection := ""
	if m.ModelName != "" {
		name := m.ModelName
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		modelSection = m.styles.Subtle.Render("  " + name)
	}

	// Daemon status
	daemonSection := "  "
	switch m.DaemonStatus {
	case "healthy", "ok":
		daemonSection += m.styles.StatusOK.Render("●")
	case "degraded":
		daemonSection += m.styles.StatusWarning.Render("●")
	case "error", "unhealthy":
		daemonSection += m.styles.StatusError.Render("●")
	default:
		daemonSection += m.styles.Subtle.Render("○")
	}

	// Contextual hints
	hintsText := m.Mode.Hints()
	if m.Mode == modes.Insert && m.InputLen > 0 {
		hintsText = fmt.Sprintf("%d chars  %s", m.InputLen, hintsText)
	}
	hints := m.styles.Subtle.Render("  " + hintsText)

	// Left side: mode + model + daemon
	left := modeLabel + modelSection + daemonSection

	// Right side: hints
	right := hints

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacerWidth := m.width - leftWidth - rightWidth
	if spacerWidth < 1 {
		spacerWidth = 1
	}

	bar := left + strings.Repeat(" ", spacerWidth) + right

	return m.styles.StatusBar.Width(m.width).Render(bar)
}

func (m Model) modeStyle() lipgloss.Style {
	switch m.Mode {
	case modes.Normal:
		return m.styles.NormalMode
	case modes.Insert:
		return m.styles.InsertMode
	case modes.Command:
		return m.styles.CommandMode
	case modes.Browse:
		return m.styles.BrowseMode
	case modes.Pair:
		return m.styles.PairMode
	case modes.Edit:
		return m.styles.EditMode
	default:
		return m.styles.NormalMode
	}
}
