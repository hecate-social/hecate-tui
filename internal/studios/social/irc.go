package social

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/client"
	"github.com/hecate-social/hecate-tui/internal/irc"
	"github.com/hecate-social/hecate-tui/internal/modes"
	"github.com/hecate-social/hecate-tui/internal/studio"
)

const maxMessagesPerChannel = 500

// panel identifies which panel is focused.
type panel int

const (
	panelChannels panel = iota
	panelMessages
)

// ircMessage is a single chat message.
type ircMessage struct {
	Nick      string
	Content   string
	Timestamp time.Time
}

// peerInfo tracks an online peer.
type peerInfo struct {
	NodeID   string
	Nick     string
	LastSeen time.Time
}

// ircModel holds all IRC sub-app state.
type ircModel struct {
	ctx *studio.Context

	// Connection
	stream    *irc.Connection
	connected bool

	// Channels
	channels       []client.IrcChannel
	joinedChannels map[string]bool
	activeChannel  string // channel_id of selected channel
	channelCursor  int    // cursor in channel list

	// Messages (ring buffer per channel)
	messages     map[string][]ircMessage
	scrollOffset int

	// Presence
	peers map[string]peerInfo
	nick  string

	// Input
	input textarea.Model
	mode  modes.Mode
	focus panel

	// Layout
	width  int
	height int

	// Signal to parent to close
	wantsBack bool

	// Loading
	loading bool
	loadErr error
}

// newIrcModel creates a new IRC model.
func newIrcModel(ctx *studio.Context) *ircModel {
	ti := textarea.New()
	ti.Placeholder = "Type a message..."
	ti.ShowLineNumbers = false
	ti.SetHeight(1)
	ti.CharLimit = 500

	// Derive nick from identity
	nick := "anon"
	if identity, err := ctx.Client.GetIdentity(); err == nil && identity != nil {
		nick = extractNick(identity.Identity)
	}

	return &ircModel{
		ctx:            ctx,
		joinedChannels: make(map[string]bool),
		messages:       make(map[string][]ircMessage),
		peers:          make(map[string]peerInfo),
		nick:           nick,
		input:          ti,
		mode:           modes.Normal,
		focus:          panelChannels,
		loading:        true,
	}
}

// init starts the IRC model — fetches channels and connects SSE.
func (m *ircModel) init() tea.Cmd {
	return tea.Batch(
		m.fetchChannels,
		m.connectStream(),
	)
}

// close tears down the IRC model.
func (m *ircModel) close() {
	if m.stream != nil {
		m.stream.Close()
	}
}

// setSize updates the layout dimensions.
func (m *ircModel) setSize(width, height int) {
	m.width = width
	m.height = height
}

// update handles all Bubble Tea messages for the IRC model.
func (m *ircModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case channelsFetchedMsg:
		m.loading = false
		m.loadErr = nil
		m.channels = msg.channels
		return nil

	case channelsFetchErrMsg:
		m.loading = false
		m.loadErr = msg.err
		return nil

	case channelOpenedMsg:
		m.channels = append(m.channels, msg.channel)
		return nil

	case channelOpenErrMsg:
		// Could show error, for now just log
		return nil

	case irc.IrcEventMsg:
		return m.handleStreamEvent(msg.Event)

	case irc.IrcContinueMsg:
		return m.pollStream()

	case irc.IrcDisconnectedMsg:
		m.connected = false
		return m.connectStream()

	case ircPollTickMsg:
		return m.pollStream()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward to textarea when in Insert mode
	if m.mode == modes.Insert {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return cmd
	}

	return nil
}

// handleStreamEvent processes a typed event from the IRC SSE stream.
func (m *ircModel) handleStreamEvent(evt irc.StreamEvent) tea.Cmd {
	m.connected = true

	switch evt.Type {
	case "message":
		m.appendMessage(evt.ChannelID, ircMessage{
			Nick:      evt.Nick,
			Content:   evt.Content,
			Timestamp: time.UnixMilli(evt.Timestamp),
		})

	case "presence":
		m.peers[evt.NodeID] = peerInfo{
			NodeID:   evt.NodeID,
			Nick:     evt.Nick,
			LastSeen: time.Now(),
		}

	case "joined":
		m.joinedChannels[evt.ChannelID] = true

	case "parted":
		delete(m.joinedChannels, evt.ChannelID)
	}

	return m.pollStream()
}

// appendMessage adds a message to the ring buffer for a channel.
func (m *ircModel) appendMessage(channelID string, msg ircMessage) {
	msgs := m.messages[channelID]
	msgs = append(msgs, msg)
	if len(msgs) > maxMessagesPerChannel {
		msgs = msgs[len(msgs)-maxMessagesPerChannel:]
	}
	m.messages[channelID] = msgs
}

// connectStream creates and subscribes to the IRC SSE stream.
func (m *ircModel) connectStream() tea.Cmd {
	m.stream = irc.NewConnection(
		m.ctx.Client.SocketPath(),
		m.ctx.Client.BaseURL(),
	)
	return m.stream.Subscribe()
}

// pollStream returns a command to poll the IRC stream.
func (m *ircModel) pollStream() tea.Cmd {
	if m.stream == nil {
		return nil
	}
	// Use a small tick to avoid busy-spinning
	return tea.Tick(50*time.Millisecond, func(_ time.Time) tea.Msg {
		return ircPollTickMsg{}
	})
}

// ircPollTickMsg triggers a stream poll after a small delay.
type ircPollTickMsg struct{}

// hints returns contextual keybinding hints.
func (m *ircModel) hints() string {
	if m.mode == modes.Insert {
		return "Enter:send  Esc:normal"
	}
	if m.focus == panelChannels {
		return "i:chat  j/k:channels  Enter:join  o:open  p:part  Esc:back"
	}
	return "i:chat  Tab:channels  j/k:scroll  Esc:back"
}

// statusInfo returns data for the shared status bar.
func (m *ircModel) statusInfo() studio.StatusInfo {
	info := studio.StatusInfo{}
	if m.activeChannel != "" {
		for _, ch := range m.channels {
			if ch.ChannelID == m.activeChannel {
				info.ChannelName = "#" + ch.Name
				break
			}
		}
	}
	info.OnlineCount = m.onlineCount()
	if m.mode == modes.Insert {
		info.InputLen = len(m.input.Value())
	}
	return info
}

// onlineCount returns the number of peers seen in the last 45 seconds.
func (m *ircModel) onlineCount() int {
	cutoff := time.Now().Add(-45 * time.Second)
	count := 0
	for _, p := range m.peers {
		if p.LastSeen.After(cutoff) {
			count++
		}
	}
	return count
}

// activeChannelName returns the display name of the active channel.
func (m *ircModel) activeChannelName() string {
	for _, ch := range m.channels {
		if ch.ChannelID == m.activeChannel {
			return ch.Name
		}
	}
	return ""
}

// activeChannelTopic returns the topic of the active channel.
func (m *ircModel) activeChannelTopic() string {
	for _, ch := range m.channels {
		if ch.ChannelID == m.activeChannel {
			return ch.Topic
		}
	}
	return ""
}

// activeMessages returns messages for the currently selected channel.
func (m *ircModel) activeMessages() []ircMessage {
	if m.activeChannel == "" {
		return nil
	}
	return m.messages[m.activeChannel]
}

// sendMessage sends the current input content to the active channel.
func (m *ircModel) sendMessage() tea.Cmd {
	content := m.input.Value()
	if content == "" || m.activeChannel == "" {
		return nil
	}

	channelID := m.activeChannel
	nick := m.nick
	cl := m.ctx.Client

	m.input.Reset()

	return func() tea.Msg {
		_ = cl.SendIrcMessage(channelID, content, nick)
		return nil
	}
}

// joinSelectedChannel joins the channel at the current cursor position.
func (m *ircModel) joinSelectedChannel() tea.Cmd {
	if m.channelCursor >= len(m.channels) {
		return nil
	}

	ch := m.channels[m.channelCursor]
	m.activeChannel = ch.ChannelID
	channelID := ch.ChannelID
	cl := m.ctx.Client

	return func() tea.Msg {
		_ = cl.JoinChannel(channelID)
		return nil
	}
}

// partActiveChannel parts the currently active channel.
func (m *ircModel) partActiveChannel() tea.Cmd {
	if m.activeChannel == "" {
		return nil
	}

	channelID := m.activeChannel
	cl := m.ctx.Client

	delete(m.joinedChannels, channelID)
	m.activeChannel = ""

	return func() tea.Msg {
		_ = cl.PartChannel(channelID)
		return nil
	}
}

// openChannel creates a new channel with the given name.
func (m *ircModel) openChannel(name string) tea.Cmd {
	cl := m.ctx.Client
	return func() tea.Msg {
		ch, err := cl.OpenChannel(name, "")
		if err != nil {
			return channelOpenErrMsg{err: err}
		}
		return channelOpenedMsg{channel: *ch}
	}
}

// fetchChannels fetches the channel list from the daemon.
func (m *ircModel) fetchChannels() tea.Msg {
	channels, err := m.ctx.Client.ListChannels()
	if err != nil {
		return channelsFetchErrMsg{err: err}
	}
	return channelsFetchedMsg{channels: channels}
}

// onlinePeers returns a sorted list of online peer nicks.
func (m *ircModel) onlinePeers() []string {
	cutoff := time.Now().Add(-45 * time.Second)
	var peers []string
	for _, p := range m.peers {
		if p.LastSeen.After(cutoff) {
			nick := p.Nick
			if nick == "" {
				nick = p.NodeID
			}
			peers = append(peers, nick)
		}
	}
	return peers
}

// extractNick extracts a short display name from an identity string.
func extractNick(identity string) string {
	if identity == "" {
		return "anon"
	}
	// Identity might be "mri:agent:io.macula/hecate@beam00" or similar
	// Try to extract the last meaningful segment
	for i := len(identity) - 1; i >= 0; i-- {
		if identity[i] == '@' || identity[i] == '/' {
			return identity[i+1:]
		}
	}
	// Truncate if too long
	if len(identity) > 16 {
		return identity[:16]
	}
	return identity
}

// Message types for async operations
type channelsFetchedMsg struct{ channels []client.IrcChannel }
type channelsFetchErrMsg struct{ err error }
type channelOpenedMsg struct{ channel client.IrcChannel }
type channelOpenErrMsg struct{ err error }

// openChannelPromptMsg signals the user wants to open a new channel.
type openChannelPromptMsg struct{}

// Ensure formatDuration helper for view use
func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}
