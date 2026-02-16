package social

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// timeNow is a variable for testability.
var timeNow = time.Now

// View renders the Social Studio content area.
func (s *Studio) View() string {
	if s.width == 0 {
		return ""
	}

	if s.activeApp == "irc" && s.irc != nil {
		return s.irc.view()
	}

	return s.viewHome()
}

// viewHome renders the app explorer grid.
func (s *Studio) viewHome() string {
	t := s.ctx.Theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).Bold(true).
		Render("\U0001F4AC Social Studio")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Chat, connect, share")

	cols := 2
	gap := 2
	cardWidth := 28

	var rows []string
	for i := 0; i < len(s.apps); i += cols {
		var rowCards []string
		for j := 0; j < cols && i+j < len(s.apps); j++ {
			idx := i + j
			selected := idx == s.appIndex
			rowCards = append(rowCards, renderAppCard(t, s.apps[idx], selected, cardWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			strings.Join(rowCards, strings.Repeat(" ", gap))))
	}

	grid := strings.Join(rows, "\n")

	hints := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("\u2191\u2193\u2190\u2192:navigate  Enter:open")

	content := title + "\n" + subtitle + "\n\n" + grid + "\n\n" + hints
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

// renderAppCard renders a single app card for the home screen.
func renderAppCard(t *theme.Theme, app socialApp, selected bool, width int) string {
	borderColor := t.Border
	if selected && app.active {
		borderColor = t.Primary
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	iconStyle := lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(t.TextDim)

	if !app.active {
		iconStyle = iconStyle.Foreground(t.TextMuted)
		nameStyle = nameStyle.Foreground(t.TextMuted)
		descStyle = descStyle.Foreground(t.TextMuted)
	}

	var content strings.Builder
	content.WriteString(iconStyle.Render(app.icon) + " " + nameStyle.Render(app.name) + "\n")
	content.WriteString(descStyle.Render(app.description))

	if !app.active {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render("Coming Soon"))
	}

	return cardStyle.Render(content.String())
}

// ─── IRC Sub-App Views ───────────────────────────────────────────────

// view renders the full IRC interface.
func (m *ircModel) view() string {
	if m.loading {
		return m.viewLoading()
	}
	if m.loadErr != nil {
		return m.viewError()
	}

	t := m.ctx.Theme

	// Header bar
	header := m.viewHeader()

	// Sidebar (channels + peers)
	sidebarWidth := 20
	sidebar := m.viewSidebar(sidebarWidth)

	// Messages area + input
	mainWidth := m.width - sidebarWidth - 1
	if mainWidth < 20 {
		mainWidth = 20
	}
	main := m.viewMain(mainWidth)

	// Vertical separator
	bodyHeight := m.height - 1
	sepLines := make([]string, bodyHeight)
	sepChar := lipgloss.NewStyle().Foreground(t.Border).Render("\u2502")
	for i := range sepLines {
		sepLines[i] = sepChar
	}
	sep := strings.Join(sepLines, "\n")

	sidebarBlock := lipgloss.NewStyle().Width(sidebarWidth).Height(bodyHeight).Render(sidebar)
	sepBlock := lipgloss.NewStyle().Width(1).Height(bodyHeight).Render(sep)
	mainBlock := lipgloss.NewStyle().Width(mainWidth).Height(bodyHeight).Render(main)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBlock, sepBlock, mainBlock)
	return header + "\n" + body
}

// viewLoading renders a loading state.
func (m *ircModel) viewLoading() string {
	t := m.ctx.Theme
	msg := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render("Connecting to IRC...")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
}

// viewError renders an error state.
func (m *ircModel) viewError() string {
	t := m.ctx.Theme
	title := lipgloss.NewStyle().
		Foreground(t.Error).Bold(true).
		Render("Failed to load channels")

	detail := lipgloss.NewStyle().
		Foreground(t.TextDim).
		Render(m.loadErr.Error())

	hint := lipgloss.NewStyle().
		Foreground(t.TextMuted).Italic(true).
		Render("Press r to retry  |  Esc to go back")

	content := title + "\n\n" + detail + "\n\n" + hint
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewHeader renders the top header bar.
func (m *ircModel) viewHeader() string {
	t := m.ctx.Theme
	headerStyle := lipgloss.NewStyle().Foreground(t.TextDim)

	back := lipgloss.NewStyle().Foreground(t.TextMuted).Render("\u2190 Esc")
	parts := []string{back}

	if m.activeChannel != "" {
		sep := lipgloss.NewStyle().Foreground(t.Border).Render(" \u2502 ")
		chanName := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).
			Render("#" + m.activeChannelName())
		parts = append(parts, sep, chanName)

		topic := m.activeChannelTopic()
		if topic != "" {
			parts = append(parts, sep, headerStyle.Render(topic))
		}
	}

	online := lipgloss.NewStyle().Foreground(t.Success).
		Render(itoa(m.onlineCount()) + " online")

	left := strings.Join(parts, "")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(online)
	if gap < 2 {
		gap = 2
	}

	return left + strings.Repeat(" ", gap) + online
}

// viewSidebar renders the channel list and online peers.
func (m *ircModel) viewSidebar(width int) string {
	t := m.ctx.Theme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true)
	b.WriteString(sectionStyle.Render("# Channels") + "\n")

	if len(m.channels) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("  No channels") + "\n")
	}

	for i, ch := range m.channels {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(t.TextDim)

		if i == m.channelCursor && m.focus == panelChannels {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		} else if ch.ChannelID == m.activeChannel {
			style = lipgloss.NewStyle().Foreground(t.Text)
		}

		name := ch.Name
		if len(name) > width-3 {
			name = name[:width-3]
		}

		indicator := ""
		if m.joinedChannels[ch.ChannelID] {
			indicator = lipgloss.NewStyle().Foreground(t.Success).Render("\u25cf ")
		}

		b.WriteString(prefix + indicator + style.Render("#"+name) + "\n")
	}

	peers := m.onlinePeers()
	if len(peers) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render("\u2500 Online") + "\n")
		for _, nick := range peers {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(t.Success).Render("\u25cf") + " ")
			b.WriteString(lipgloss.NewStyle().Foreground(t.Text).Render(nick) + "\n")
		}
	}

	return b.String()
}

// viewMain renders the messages area and input bar.
func (m *ircModel) viewMain(width int) string {
	t := m.ctx.Theme

	inputHeight := 1
	if m.mode == modes.Insert {
		inputHeight = 3
	}
	msgsHeight := m.height - inputHeight - 2
	if msgsHeight < 1 {
		msgsHeight = 1
	}

	msgs := m.activeMessages()
	var msgLines []string

	if m.activeChannel == "" {
		placeholder := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("Select a channel to start chatting")
		msgLines = append(msgLines, lipgloss.Place(width, msgsHeight,
			lipgloss.Center, lipgloss.Center, placeholder))
	} else if len(msgs) == 0 {
		placeholder := lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).
			Render("No messages yet. Say hello!")
		msgLines = append(msgLines, lipgloss.Place(width, msgsHeight,
			lipgloss.Center, lipgloss.Center, placeholder))
	} else {
		visibleMsgs := msgs
		if len(visibleMsgs) > msgsHeight {
			start := len(visibleMsgs) - msgsHeight - m.scrollOffset
			if start < 0 {
				start = 0
			}
			end := start + msgsHeight
			if end > len(visibleMsgs) {
				end = len(visibleMsgs)
			}
			visibleMsgs = visibleMsgs[start:end]
		}

		nickColors := []lipgloss.Color{
			lipgloss.Color("6"),
			lipgloss.Color("3"),
			lipgloss.Color("2"),
			lipgloss.Color("5"),
			lipgloss.Color("4"),
			lipgloss.Color("1"),
			lipgloss.Color("14"),
			lipgloss.Color("11"),
		}

		for _, msg := range visibleMsgs {
			ts := lipgloss.NewStyle().Foreground(t.TextMuted).
				Render(formatTimestamp(msg.Timestamp))

			colorIdx := 0
			for _, c := range msg.Nick {
				colorIdx += int(c)
			}
			nickColor := nickColors[colorIdx%len(nickColors)]
			nick := lipgloss.NewStyle().Foreground(nickColor).Bold(true).
				Render(msg.Nick)

			content := lipgloss.NewStyle().Foreground(t.Text).
				Render(msg.Content)

			line := ts + " " + nick + " " + content
			msgLines = append(msgLines, line)
		}
	}

	for len(msgLines) < msgsHeight {
		msgLines = append([]string{""}, msgLines...)
	}

	messagesView := strings.Join(msgLines, "\n")

	sep := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("\u2500", width))

	var inputView string
	if m.mode == modes.Insert && m.activeChannel != "" {
		m.input.SetWidth(width - 4)
		inputView = "> " + m.input.View()
	} else if m.activeChannel != "" && m.joinedChannels[m.activeChannel] {
		inputView = lipgloss.NewStyle().Foreground(t.TextMuted).
			Render("> Press i to type...")
	} else {
		inputView = lipgloss.NewStyle().Foreground(t.TextMuted).
			Render("  Join a channel to chat")
	}

	return messagesView + "\n" + sep + "\n" + inputView
}

// itoa converts an int to a string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
