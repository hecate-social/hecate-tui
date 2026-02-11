package statusbar

import (
	"fmt"
	"os"
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

	Mode          modes.Mode
	Cwd           string // current working directory
	ModelName     string
	ModelProvider string // "ollama", "openai", "anthropic", etc.
	MeshStatus    string // "connected", "disconnected", "unknown"
	DaemonStatus  string // "healthy", "degraded", "error", "unknown"
	ModelStatus   string // "ready", "loading", "error"
	ModelError    string // error message when ModelStatus is "error"
	InputLen      int    // character count for Insert mode
	SessionTokens int    // cumulative tokens for session

	// Venture context
	VentureName string // current venture name (empty if none)
	ActivePhase string // current ALC phase: "dna", "anp", "tni", "dno"
	AgentCount  int    // number of active agents
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

// Height returns the number of lines the status bar occupies.
func (m Model) Height() int {
	return 2
}

// View renders the status bar as two lines with consistent styling.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	barStyle := m.styles.StatusBar.Width(m.width)

	// ── Line 1: mode + model + tokens ──
	modeStyle := m.modeStyle()
	modeLabel := modeStyle.Render(" " + m.Mode.String() + " ")

	// Model indicator with provider and status LED
	modelSection := ""
	if m.ModelName != "" {
		name := m.ModelName
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		modelLED := ""
		switch m.ModelStatus {
		case "loading":
			modelLED = m.styles.StatusWarning.Render("◐") + " "
		case "error":
			modelLED = m.styles.StatusError.Render("●") + " "
		default:
			modelLED = m.styles.StatusOK.Render("●") + " "
		}

		providerLabel := ""
		if m.ModelProvider != "" {
			if m.isPaidProvider() {
				providerLabel = m.styles.StatusWarning.Render(" [" + m.ModelProvider + " $]")
			} else {
				providerLabel = m.styles.Subtle.Render(" [" + m.ModelProvider + "]")
			}
		}
		modelSection = "  " + modelLED + m.styles.Subtle.Render(name) + providerLabel
	}

	// Token count (only show if non-zero and using paid provider)
	tokenSection := ""
	if m.SessionTokens > 0 && m.isPaidProvider() {
		tokenSection = m.styles.Subtle.Render(fmt.Sprintf("  %s tok", formatTokenCount(m.SessionTokens)))
	}

	line1 := modeLabel + modelSection + tokenSection

	// ── Line 2: cwd + hints ──
	cwdSection := ""
	if m.Cwd != "" {
		cwd := shortenPath(m.Cwd, 40)
		cwdSection = " " + m.styles.Subtle.Render(cwd)
	}

	// Contextual hints (or error message if model failed)
	var hints string
	if m.ModelStatus == "error" && m.ModelError != "" {
		errMsg := m.ModelError
		if len(errMsg) > 50 {
			errMsg = errMsg[:47] + "..."
		}
		hints = m.styles.StatusError.Render(" ✗ " + errMsg)
	} else if m.ModelStatus == "loading" {
		hints = m.styles.StatusWarning.Render(" ◐ Loading model...")
	} else {
		hintsText := m.Mode.Hints()
		if m.Mode == modes.Insert && m.InputLen > 0 {
			hintsText = fmt.Sprintf("%d chars  %s", m.InputLen, hintsText)
		}
		hints = m.styles.Subtle.Render(" " + hintsText)
	}

	line2Left := cwdSection
	line2Right := hints
	line2LeftWidth := lipgloss.Width(line2Left)
	line2RightWidth := lipgloss.Width(line2Right)
	spacer2 := m.width - line2LeftWidth - line2RightWidth
	if spacer2 < 1 {
		spacer2 = 1
	}
	line2 := line2Left + strings.Repeat(" ", spacer2) + line2Right

	return barStyle.Render(line1) + "\n" + barStyle.Render(line2)
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
	case modes.Form:
		return m.styles.CommandMode // Reuse command style for forms
	default:
		return m.styles.NormalMode
	}
}

// phaseStyle returns a style for ALC phase badges.
func (m Model) phaseStyle(phase string) lipgloss.Style {
	base := lipgloss.NewStyle().Padding(0, 1).Bold(true)
	switch strings.ToLower(phase) {
	case "dna":
		return base.Foreground(lipgloss.Color("#7C3AED")) // Purple - Discovery
	case "anp":
		return base.Foreground(lipgloss.Color("#2563EB")) // Blue - Architecture
	case "tni":
		return base.Foreground(lipgloss.Color("#059669")) // Green - Testing
	case "dno":
		return base.Foreground(lipgloss.Color("#DC2626")) // Red - Deployment
	default:
		return base.Foreground(lipgloss.Color("#6B7280")) // Gray
	}
}

// isPaidProvider returns true if the current model uses a commercial provider.
func (m Model) isPaidProvider() bool {
	switch m.ModelProvider {
	case "anthropic", "openai", "google", "mistral", "groq", "together":
		return true
	default:
		return false
	}
}

// formatTokenCount formats token count with K suffix for thousands.
func formatTokenCount(count int) string {
	if count >= 1000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000)
	}
	return fmt.Sprintf("%d", count)
}

// shortenPath shortens a path for display, replacing home dir with ~ and truncating if needed.
func shortenPath(path string, maxLen int) string {
	// Replace home directory with ~
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	// If still too long, show .../<last-two-dirs>
	if len(path) > maxLen {
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			// Keep last 2 parts
			short := ".../" + strings.Join(parts[len(parts)-2:], "/")
			if len(short) <= maxLen {
				return short
			}
		}
		// Still too long, just truncate
		return "..." + path[len(path)-(maxLen-3):]
	}

	return path
}
