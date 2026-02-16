package social

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/modes"
)

// handleHomeKey processes key events on the Studio Home screen.
func (s *Studio) handleHomeKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Count active apps for grid navigation
	cols := 2
	totalApps := len(s.apps)

	switch key {
	case "enter":
		if s.appIndex < totalApps && s.apps[s.appIndex].active {
			return s.openIrc()
		}
		return nil

	case "j", "down":
		if s.appIndex+cols < totalApps {
			s.appIndex += cols
		}
	case "k", "up":
		if s.appIndex-cols >= 0 {
			s.appIndex -= cols
		}
	case "l", "right":
		if s.appIndex+1 < totalApps && (s.appIndex+1)%cols != 0 {
			s.appIndex++
		}
	case "h", "left":
		if s.appIndex > 0 && s.appIndex%cols != 0 {
			s.appIndex--
		}
	}

	return nil
}

// handleKey processes key events in the IRC sub-app.
func (m *ircModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Insert mode: typing a message
	if m.mode == modes.Insert {
		return m.handleInsertKey(msg, key)
	}

	// Normal mode
	return m.handleNormalKey(key)
}

// handleInsertKey processes keys in Insert mode.
func (m *ircModel) handleInsertKey(msg tea.KeyMsg, key string) tea.Cmd {
	switch key {
	case "enter":
		return m.sendMessage()
	case "esc":
		m.mode = modes.Normal
		m.input.Blur()
		return nil
	}

	// Forward to textarea
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

// handleNormalKey processes keys in Normal mode.
func (m *ircModel) handleNormalKey(key string) tea.Cmd {
	switch key {
	case "i":
		if m.activeChannel != "" && m.joinedChannels[m.activeChannel] {
			m.mode = modes.Insert
			m.input.Focus()
		}
		return nil

	case "esc":
		m.wantsBack = true
		return nil

	case "tab":
		if m.focus == panelChannels {
			m.focus = panelMessages
		} else {
			m.focus = panelChannels
		}
		return nil

	case "o":
		// Quick open: create a channel named after a timestamp
		// In a full implementation this would show a form
		return m.openChannel("chan-" + strings.ReplaceAll(
			strings.Split(strings.Split(
				formatTimestamp(timeNow()), ":")[0], " ")[0],
			":", ""))

	case "p":
		return m.partActiveChannel()

	case "r":
		m.loading = true
		return tea.Cmd(m.fetchChannels)
	}

	// Panel-specific navigation
	if m.focus == panelChannels {
		return m.handleChannelNav(key)
	}
	return m.handleMessageNav(key)
}

// handleChannelNav handles j/k navigation in the channel sidebar.
func (m *ircModel) handleChannelNav(key string) tea.Cmd {
	total := len(m.channels)
	if total == 0 {
		return nil
	}

	switch key {
	case "j", "down":
		if m.channelCursor < total-1 {
			m.channelCursor++
		}
	case "k", "up":
		if m.channelCursor > 0 {
			m.channelCursor--
		}
	case "enter":
		return m.joinSelectedChannel()
	}

	return nil
}

// handleMessageNav handles j/k scrolling in the messages panel.
func (m *ircModel) handleMessageNav(key string) tea.Cmd {
	msgs := m.activeMessages()
	if len(msgs) == 0 {
		return nil
	}

	switch key {
	case "j", "down":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "k", "up":
		maxScroll := len(msgs) - 1
		if m.scrollOffset < maxScroll {
			m.scrollOffset++
		}
	case "g", "home":
		m.scrollOffset = len(msgs) - 1
	case "G", "end":
		m.scrollOffset = 0
	}

	return nil
}
