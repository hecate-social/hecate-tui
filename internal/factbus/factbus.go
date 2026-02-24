package factbus

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

// FactMsg is a domain fact received from the daemon over SSE.
type FactMsg struct {
	FactType string          `json:"fact_type"`
	Data     json.RawMessage `json:"data"`
}

// FactContinueMsg signals the App to re-poll the fact channel.
type FactContinueMsg struct{}

// FactDisconnectedMsg signals the SSE connection was lost.
type FactDisconnectedMsg struct{}

// FactErrorMsg carries a transport error.
type FactErrorMsg struct{ Err error }

// Connection manages an SSE subscription to the daemon's /api/facts/stream endpoint.
type Connection struct {
	socketPath string
	baseURL    string
	factChan   chan FactMsg
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConnection creates a factbus connection. Pass socketPath for Unix socket,
// or empty string + baseURL for TCP.
func NewConnection(socketPath, baseURL string) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	return &Connection{
		socketPath: socketPath,
		baseURL:    baseURL,
		factChan:   make(chan FactMsg, 50),
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

// PollCmd returns a Bubble Tea command that non-blocking checks for a fact.
func (c *Connection) PollCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case fact, ok := <-c.factChan:
			if !ok {
				return FactDisconnectedMsg{}
			}
			return fact
		default:
			return FactContinueMsg{}
		}
	}
}

// Close cancels the SSE connection.
func (c *Connection) Close() {
	c.cancel()
}

// connectLoop runs the SSE connection with automatic reconnection.
func (c *Connection) connectLoop() {
	defer close(c.factChan)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.runSSE()

		// If context cancelled, exit
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Wait before reconnecting
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

// runSSE establishes one SSE connection and reads until disconnect/error.
func (c *Connection) runSSE() {
	url := c.baseURL + "/api/facts/stream"

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

		// Skip empty lines and comments (heartbeats)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data lines
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if data == "" || data == "[DONE]" {
				continue
			}

			var fact FactMsg
			if err := json.Unmarshal([]byte(data), &fact); err != nil {
				continue
			}

			select {
			case c.factChan <- fact:
			case <-c.ctx.Done():
				return
			}
		}
	}
}
