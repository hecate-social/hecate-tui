package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ViewChat renders just the chat area (messages + streaming).
func (m Model) ViewChat() string {
	return m.viewport.View()
}

// ViewInput renders the textarea (for Insert mode).
func (m Model) ViewInput() string {
	if !m.inputVisible {
		return ""
	}
	return m.input.View()
}

// ViewStats renders the stats line (streaming status or post-completion stats).
func (m Model) ViewStats() string {
	if m.streaming {
		// Show model name, elapsed time, and cancel hint (thinking animation is now in chat)
		subtleStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
		modelPart := ""
		if name := m.ActiveModelName(); name != "" {
			modelPart = subtleStyle.Render("  via " + name)
		}
		elapsed := time.Since(m.streamStart)
		elapsedPart := subtleStyle.Render(fmt.Sprintf("  %0.1fs", elapsed.Seconds()))
		cancelHint := subtleStyle.Render("  (Esc to cancel)")
		return modelPart + elapsedPart + cancelHint
	}
	if m.lastTokenCount > 0 {
		return m.renderStats()
	}
	return ""
}

// ViewError renders any error.
func (m Model) ViewError() string {
	if m.err != nil {
		return m.styles.Error.Render("  " + m.err.Error())
	}
	return ""
}

func (m Model) renderStreaming() string {
	frame := ThinkingFrames[m.thinkingFrame]
	sparkle := Sparkles[m.thinkingFrame%len(Sparkles)]

	streamStyle := lipgloss.NewStyle().Foreground(m.theme.StreamingColor).Bold(true)
	subtleStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)

	// Model name
	modelPart := ""
	if name := m.ActiveModelName(); name != "" {
		modelPart = subtleStyle.Render(" via " + name)
	}

	// Elapsed time
	elapsed := time.Since(m.streamStart)
	elapsedPart := subtleStyle.Render(fmt.Sprintf("  %0.1fs", elapsed.Seconds()))

	// Cancel hint
	cancelHint := subtleStyle.Render("  (Esc to cancel)")

	return streamStyle.Render("  "+sparkle+" "+frame+" "+sparkle) + modelPart + elapsedPart + cancelHint
}

func (m Model) renderStats() string {
	durationPart := lipgloss.NewStyle().Foreground(m.theme.TextMuted).
		Render(fmt.Sprintf("  %0.1fs", m.lastDuration.Seconds()))
	return "  " + FormatTokens(m.lastTokenCount, m.theme) + "  " + FormatSpeed(m.lastSpeed, m.theme) + durationPart
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		welcome := WelcomeArt(m.theme)
		return lipgloss.Place(
			m.viewport.Width,
			m.viewport.Height,
			lipgloss.Center,
			lipgloss.Center,
			welcome,
		)
	}

	var parts []string
	bubbleWidth := m.viewport.Width - 8
	if bubbleWidth < 30 {
		bubbleWidth = 30
	}

	timeStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)

	for _, msg := range m.messages {
		timestamp := ""
		if !msg.Time.IsZero() {
			timestamp = timeStyle.Render(" " + msg.Time.Format("15:04"))
		}

		switch msg.Role {
		case "user":
			// User messages: just the bullet + content, no header line
			bullet := m.styles.UserLabel.Render("▸ ")
			bubble := m.styles.UserBubble.Render(msg.Content) + timestamp
			parts = append(parts, bullet+bubble)

		case "assistant":
			label := m.styles.AssistantLabel.Render("◆ Hecate") + timestamp
			rendered := RenderMarkdown(msg.Content, m.theme, bubbleWidth-4)
			bubble := m.styles.AssistantBubble.Width(bubbleWidth).Render(rendered)
			parts = append(parts, label+"\n"+bubble)

		case "system":
			bubble := m.styles.SystemBubble.Width(bubbleWidth).Render(msg.Content)
			parts = append(parts, bubble)
		}
	}

	return strings.Join(parts, "\n\n")
}

func (m *Model) updateViewport() {
	content := m.renderMessages()
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) updateViewportPreserveScroll() {
	// Preserve scroll position as percentage
	oldTotal := m.viewport.TotalLineCount()
	oldOffset := m.viewport.YOffset
	scrollPercent := 0.0
	if oldTotal > 0 {
		scrollPercent = float64(oldOffset) / float64(oldTotal)
	}
	atBottom := m.viewport.AtBottom()

	content := m.renderMessages()
	m.viewport.SetContent(content)

	// Restore scroll position
	if atBottom {
		m.viewport.GotoBottom()
	} else {
		newTotal := m.viewport.TotalLineCount()
		newOffset := int(scrollPercent * float64(newTotal))
		m.viewport.SetYOffset(newOffset)
	}
}

func (m *Model) updateStreamingMessage() {
	content := m.renderMessages()
	// Always show assistant label when streaming
	content += "\n\n" + m.styles.AssistantLabel.Render("◆ Hecate") + "\n"
	if m.streamBuf.Len() > 0 {
		// Show streamed content with cursor
		bubble := m.styles.AssistantBubble.Width(m.viewport.Width - 8).Render(m.streamBuf.String() + "▊")
		content += bubble
	} else {
		// Show thinking animation in the chat area while waiting for content
		frame := ThinkingFrames[m.thinkingFrame]
		sparkle := Sparkles[m.thinkingFrame%len(Sparkles)]
		thinkingStyle := lipgloss.NewStyle().Foreground(m.theme.StreamingColor)
		thinking := thinkingStyle.Render(sparkle + " " + frame + " " + sparkle)
		bubble := m.styles.AssistantBubble.Width(m.viewport.Width - 8).Render(thinking)
		content += bubble
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m *Model) resize() {
	// m.height is already the chat area height (app subtracts header, statusbar, input, stats)
	vpHeight := m.height
	if vpHeight < 5 {
		vpHeight = 5
	}

	// Dynamic horizontal padding based on terminal width
	hPadding := 4
	if m.width < 80 {
		hPadding = 2
	} else if m.width < 60 {
		hPadding = 1
	}

	vpWidth := m.width - hPadding
	if vpWidth < 40 {
		vpWidth = 40
	}

	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.input.SetWidth(vpWidth - 2)

	m.updateViewportPreserveScroll()
}

// SetSize updates the model dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resize()
}
