package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/alc"
	"github.com/hecate-social/hecate-tui/internal/version"
)

// View renders the entire TUI.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	if a.showHome {
		return a.renderHome()
	}

	var sections []string

	// Header (brand + context + tab bar + separator)
	sections = append(sections, a.renderHeader())

	// Active studio content
	if a.activeStudio < len(a.studios) {
		sections = append(sections, a.studios[a.activeStudio].View())
	}

	// Command line (if in command mode)
	if a.inCommandMode {
		sections = append(sections, a.renderCommandLine())
	}

	// Status bar (always at bottom)
	sections = append(sections, a.statusBar.View())

	return strings.Join(sections, "\n")
}

func (a *App) renderBrandRow() string {
	logo := lipgloss.NewStyle().Foreground(a.theme.Primary).Bold(true).Render("🔥🗝️🔥 Hecate")
	versionSection := a.styles.Subtle.Render(" v" + version.Version)

	daemonSection := "  "
	switch a.daemonStatus {
	case "healthy", "ok":
		daemonSection += a.styles.StatusOK.Render("●") + a.styles.Subtle.Render(" daemon")
	case "starting":
		daemonSection += a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon starting")
	case "degraded":
		daemonSection += a.styles.StatusWarning.Render("●") + a.styles.Subtle.Render(" daemon")
	default:
		daemonSection += a.styles.Subtle.Render("○ daemon")
	}

	rxLED := "  "
	if !a.factStreamConnected {
		rxLED += a.styles.Subtle.Render("▽ rx")
	} else if a.rxActive {
		rxLED += a.styles.StatusOK.Render("▼ rx")
	} else {
		rxLED += a.styles.Subtle.Render("▼ rx")
	}

	txLED := " "
	if !a.factStreamConnected {
		txLED += a.styles.Subtle.Render("△ tx")
	} else if a.txActive {
		txLED += a.styles.StatusOK.Render("▲ tx")
	} else {
		txLED += a.styles.Subtle.Render("▲ tx")
	}

	row1Left := logo + versionSection + daemonSection + rxLED + txLED

	donateURL := "https://" + version.DonateURL
	donateText := a.styles.Subtle.Render("☕ donate")
	donateLink := fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", donateURL, donateText)

	row1LeftWidth := lipgloss.Width(row1Left)
	row1RightWidth := lipgloss.Width(donateText)
	spacer := a.width - row1LeftWidth - row1RightWidth - 2
	if spacer < 1 {
		spacer = 1
	}
	return " " + row1Left + strings.Repeat(" ", spacer) + donateLink + " "
}

func (a *App) renderContextRow() string {
	llm := a.llmStudio()
	if llm == nil {
		return ""
	}
	alcState := llm.ALCState()
	if alcState == nil || alcState.Context == alc.Chat {
		return ""
	}

	rowStyle := lipgloss.NewStyle().Width(a.width).Padding(0, 1)
	var parts []string

	if alcState.Venture != nil {
		ventureStyle := lipgloss.NewStyle().Foreground(a.theme.Warning).Bold(true)
		parts = append(parts, ventureStyle.Render("🔥 "+alcState.Venture.Name))
	}

	if alcState.Context == alc.Department && alcState.Department != nil {
		departmentStyle := lipgloss.NewStyle().Foreground(a.theme.Secondary)
		parts = append(parts, departmentStyle.Render("🏢 "+alcState.Department.Name))

		if phase := alcState.Department.CurrentPhase; phase != "" {
			phaseStyle := a.phaseStyle(string(phase))
			parts = append(parts, phaseStyle.Render("📍 "+strings.ToUpper(string(phase))))
		}
	}

	return rowStyle.Render(strings.Join(parts, a.styles.Subtle.Render(" › ")))
}

func (a *App) renderTabBar() string {
	var tabs []string

	for i, s := range a.studios {
		label := s.Icon() + " " + s.ShortName()
		if i == a.activeStudio {
			style := lipgloss.NewStyle().
				Foreground(a.theme.Text).
				Bold(true).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		} else {
			style := lipgloss.NewStyle().
				Foreground(a.theme.TextMuted).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		}
	}

	bar := strings.Join(tabs, a.styles.Subtle.Render("│"))
	return lipgloss.NewStyle().Width(a.width).Padding(0, 1).Render(bar)
}

func (a *App) renderCommandLine() string {
	return lipgloss.NewStyle().
		Width(a.width).
		Padding(0, 1).
		Background(a.theme.BgInput).
		Render(a.cmdInput.View())
}

func (a *App) phaseStyle(phase string) lipgloss.Style {
	switch strings.ToLower(phase) {
	case "dna":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	case "anp":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Bold(true)
	case "tni":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#059669")).Bold(true)
	case "dno":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Bold(true)
	}
}
