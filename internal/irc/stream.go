// Package irc provides an SSE connection to the daemon's /api/irc/stream endpoint.
// Follows the factbus.Connection pattern: auto-reconnecting goroutine that
// pushes typed messages into a buffered channel consumed by Bubble Tea commands.
package irc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// StreamEvent is a raw SSE event from the IRC stream.
type StreamEvent struct {
	Type      string          `json:"type"`       // "message", "presence", "joined", "parted"
	ChannelID string          `json:"channel_id"` // relevant channel (empty for presence)
	Nick      string          `json:"nick"`
	Content   string          `json:"content"`
	NodeID    string          `json:"node_id"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"` // raw fallback
}

// IrcEventMsg wraps a stream event for Bubble Tea.
type IrcEventMsg struct{ Event StreamEvent }

// IrcContinueMsg signals re-poll (no event available yet).
type IrcContinueMsg struct{}

// IrcDisconnectedMsg signals the SSE connection was lost.
type IrcDisconnectedMsg struct{}

// Connection manages the SSE subscription to /api/irc/stream.
type Connection struct {
	socketPath string
	baseURL    string
	eventChan  chan StreamEvent
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConnection creates an IRC stream connection.
func NewConnection(socketPath, baseURL string) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	return &Connection{
		socketPath: socketPath,
		baseURL:    baseURL,
		eventChan:  make(chan StreamEvent, 50),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Subscribe starts the SSE goroutine and returns a Bubble Tea command
// that kicks off the polling loop.
func (c *Connection) Subscribe() tea.Cmd {
	go c.connectLoop()
	return c.PollCmd()
}

// PollCmd returns a Bubble Tea command that non-blocking checks for an event.
func (c *Connection) PollCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case evt, ok := <-c.eventChan:
			if !ok {
				return IrcDisconnectedMsg{}
			}
			return IrcEventMsg{Event: evt}
		default:
			return IrcContinueMsg{}
		}
	}
}

// Close cancels the SSE connection.
func (c *Connection) Close() {
	c.cancel()
}

// connectLoop runs the SSE connection with automatic reconnection.
func (c *Connection) connectLoop() {
	defer close(c.eventChan)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.runSSE()

		select {
		case <-c.ctx.Done():
			return
		default:
		}

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

// runSSE establishes one SSE connection and reads until disconnect.
func (c *Connection) runSSE() {
	url := c.baseURL + "/api/irc/stream"

	httpReq, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return
	}
	httpReq.Header.Set("Accept", "text/event-stream")

	transport := &http.Transport{
		IdleConnTimeout:       0,
		ResponseHeaderTimeout: 0,
		ExpectContinueTimeout: 0,
	}
	if c.socketPath != "" {
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", c.socketPath)
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return
	}

	c.readSSE(resp.Body)
}

// readSSE parses SSE lines from the response body.
func (c *Connection) readSSE(body io.Reader) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if data == "" || data == "[DONE]" {
				continue
			}

			var evt StreamEvent
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}

			select {
			case c.eventChan <- evt:
			case <-c.ctx.Done():
				return
			}
		}
	}
}
