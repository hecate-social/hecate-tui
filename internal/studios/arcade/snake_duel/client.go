package snake_duel

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Bubble Tea message types for match lifecycle.
type MatchStateMsg struct{ State GameState }
type MatchContinueMsg struct{}
type MatchDoneMsg struct{}
type MatchErrorMsg struct{ Err error }
type MatchStartedMsg struct{ MatchID string }
type MatchStartFailedMsg struct{ Err error }

// MatchStream manages an SSE connection to a single match.
// Unlike the IRC stream, matches are ephemeral — no auto-reconnect.
type MatchStream struct {
	socketPath string
	baseURL    string
	eventChan  chan GameState
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMatchStream creates a stream for a specific match.
func NewMatchStream(socketPath, baseURL string) *MatchStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &MatchStream{
		socketPath: socketPath,
		baseURL:    baseURL,
		eventChan:  make(chan GameState, 20),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Connect starts the SSE goroutine for the given match and returns
// a Bubble Tea command that kicks off polling.
func (s *MatchStream) Connect(matchID string) tea.Cmd {
	go s.readLoop(matchID)
	return s.PollCmd()
}

// PollCmd returns a Bubble Tea command that non-blocking checks for a state update.
func (s *MatchStream) PollCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case state, ok := <-s.eventChan:
			if !ok {
				return MatchDoneMsg{}
			}
			return MatchStateMsg{State: state}
		default:
			return MatchContinueMsg{}
		}
	}
}

// Close cancels the SSE connection.
func (s *MatchStream) Close() {
	s.cancel()
}

// readLoop establishes the SSE connection and reads until disconnect.
func (s *MatchStream) readLoop(matchID string) {
	defer close(s.eventChan)

	url := s.baseURL + "/api/arcade/snake-duel/matches/" + matchID + "/stream"

	req, err := http.NewRequestWithContext(s.ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "text/event-stream")

	transport := &http.Transport{
		IdleConnTimeout:       0,
		ResponseHeaderTimeout: 0,
		ExpectContinueTimeout: 0,
	}
	if s.socketPath != "" {
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", s.socketPath)
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	s.readSSE(resp.Body)
}

// readSSE parses SSE data: lines from the response body.
func (s *MatchStream) readSSE(body io.Reader) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
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

			var state GameState
			if err := json.Unmarshal([]byte(data), &state); err != nil {
				continue
			}

			select {
			case s.eventChan <- state:
			case <-s.ctx.Done():
				return
			}
		}
	}
}

// StartMatch sends POST to the daemon to create a new match.
// Returns a tea.Cmd for async execution.
func StartMatch(socketPath, baseURL string, af1, af2, tickMs int) tea.Cmd {
	return func() tea.Msg {
		body := map[string]interface{}{
			"af1":     af1,
			"af2":     af2,
			"tick_ms": tickMs,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return MatchStartFailedMsg{Err: err}
		}

		url := baseURL + "/api/arcade/snake-duel/matches"
		req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
		if err != nil {
			return MatchStartFailedMsg{Err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		transport := &http.Transport{}
		if socketPath != "" {
			transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socketPath)
			}
		}

		httpClient := &http.Client{Transport: transport}
		resp, err := httpClient.Do(req)
		if err != nil {
			return MatchStartFailedMsg{Err: err}
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return MatchStartFailedMsg{Err: err}
		}

		var result StartMatchResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return MatchStartFailedMsg{Err: err}
		}

		if result.MatchID == "" {
			return MatchStartFailedMsg{Err: errEmptyMatchID}
		}

		return MatchStartedMsg{MatchID: result.MatchID}
	}
}

type matchError string

func (e matchError) Error() string { return string(e) }

const errEmptyMatchID = matchError("daemon returned empty match_id")
