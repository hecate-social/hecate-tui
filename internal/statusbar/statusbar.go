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

	Mode          modes.Mode
	ModelName     string
	ModelProvider string // "ollama", "openai", "anthropic", etc.
	MeshStatus    string // "connected", "disconnected", "unknown"
	DaemonStatus  string // "healthy", "degraded", "error", "unknown"
	ModelStatus   string // "ready", "loading", "error"
	ModelError    string // error message when ModelStatus is "error"
	InputLen      int    // character count for Insert mode
	SessionTokens int    // cumulative tokens for session
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

	// Model indicator with provider and status LED
	modelSection := ""
	if m.ModelName != "" {
		name := m.ModelName
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Model status LED (shows loading/ready/error)
		modelLED := ""
		switch m.ModelStatus {
		case "loading":
			modelLED = m.styles.StatusWarning.Render("◐") + " " // half-filled = loading
		case "error":
			modelLED = m.styles.StatusError.Render("●") + " " // red = error
		default:
			modelLED = m.styles.StatusOK.Render("●") + " " // green = ready
		}

		// Show provider in brackets, with $ for paid providers
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

	// Token count (only show if non-zero and using paid provider)
	tokenSection := ""
	if m.SessionTokens > 0 && m.isPaidProvider() {
		tokenSection = m.styles.Subtle.Render(fmt.Sprintf("  %s tok", formatTokenCount(m.SessionTokens)))
	}

	// Contextual hints (or error message if model failed)
	var hints string
	if m.ModelStatus == "error" && m.ModelError != "" {
		// Truncate long errors to fit in status bar
		errMsg := m.ModelError
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		hints = m.styles.StatusError.Render("  ✗ " + errMsg)
	} else if m.ModelStatus == "loading" {
		hints = m.styles.StatusWarning.Render("  ◐ Loading model...")
	} else {
		hintsText := m.Mode.Hints()
		if m.Mode == modes.Insert && m.InputLen > 0 {
			hintsText = fmt.Sprintf("%d chars  %s", m.InputLen, hintsText)
		}
		hints = m.styles.Subtle.Render("  " + hintsText)
	}

	// Left side: mode + model + daemon + tokens
	left := modeLabel + modelSection + daemonSection + tokenSection

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
