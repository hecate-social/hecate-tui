package stables

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

// Bubble Tea message types for stables lifecycle.
type StablesListMsg struct{ Stables []Stable }
type StablesListErrMsg struct{ Err error }
type StableDetailMsg struct{ Stable Stable }
type StableDetailErrMsg struct{ Err error }
type ChampionMsg struct{ Champion Champion }
type ChampionErrMsg struct{ Err error }
type GenerationsMsg struct{ Generations []GenerationStats }
type GenerationsErrMsg struct{ Err error }
type StableCreatedMsg struct{ StableID string }
type StableCreateErrMsg struct{ Err error }
type TrainingHaltedMsg struct{ StableID string }
type TrainingHaltErrMsg struct{ Err error }
type DuelStartedMsg struct{ MatchID string }
type DuelStartErrMsg struct{ Err error }

// Training SSE stream messages
type TrainingUpdateMsg struct{ Progress TrainingProgress }
type TrainingStreamContinueMsg struct{}
type TrainingStreamDoneMsg struct{}

// FetchStables retrieves all stables from the daemon.
func FetchStables(socketPath, baseURL string) tea.Cmd {
	return func() tea.Msg {
		body, err := doGet(socketPath, baseURL, "/api/arcade/gladiators/stables")
		if err != nil {
			return StablesListErrMsg{Err: err}
		}
		var resp StablesListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return StablesListErrMsg{Err: err}
		}
		return StablesListMsg{Stables: resp.Stables}
	}
}

// FetchStable retrieves a single stable by ID.
func FetchStable(socketPath, baseURL, stableID string) tea.Cmd {
	return func() tea.Msg {
		body, err := doGet(socketPath, baseURL, "/api/arcade/gladiators/stables/"+stableID)
		if err != nil {
			return StableDetailErrMsg{Err: err}
		}
		var resp StableResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return StableDetailErrMsg{Err: err}
		}
		return StableDetailMsg{Stable: resp.Stable}
	}
}

// FetchChampion retrieves the champion for a stable.
func FetchChampion(socketPath, baseURL, stableID string) tea.Cmd {
	return func() tea.Msg {
		body, err := doGet(socketPath, baseURL, "/api/arcade/gladiators/stables/"+stableID+"/champion")
		if err != nil {
			return ChampionErrMsg{Err: err}
		}
		var resp ChampionResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return ChampionErrMsg{Err: err}
		}
		return ChampionMsg{Champion: resp.Champion}
	}
}

// FetchGenerations retrieves training history for a stable.
func FetchGenerations(socketPath, baseURL, stableID string) tea.Cmd {
	return func() tea.Msg {
		body, err := doGet(socketPath, baseURL, "/api/arcade/gladiators/stables/"+stableID+"/generations")
		if err != nil {
			return GenerationsErrMsg{Err: err}
		}
		var resp GenerationsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return GenerationsErrMsg{Err: err}
		}
		return GenerationsMsg{Generations: resp.Generations}
	}
}

// InitiateStable creates a new training stable.
func InitiateStable(socketPath, baseURL string, req InitiateStableRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := doPost(socketPath, baseURL, "/api/arcade/gladiators/stables", req)
		if err != nil {
			return StableCreateErrMsg{Err: err}
		}
		var resp InitiateStableResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return StableCreateErrMsg{Err: err}
		}
		if resp.StableID == "" {
			return StableCreateErrMsg{Err: stableErr("daemon returned empty stable_id")}
		}
		return StableCreatedMsg{StableID: resp.StableID}
	}
}

// HaltTraining stops a running training process.
func HaltTraining(socketPath, baseURL, stableID string) tea.Cmd {
	return func() tea.Msg {
		_, err := doPost(socketPath, baseURL, "/api/arcade/gladiators/stables/"+stableID+"/halt", nil)
		if err != nil {
			return TrainingHaltErrMsg{Err: err}
		}
		return TrainingHaltedMsg{StableID: stableID}
	}
}

// StartChampionDuel starts a match between the stable's champion and an AI opponent.
func StartChampionDuel(socketPath, baseURL, stableID string, opponentAF, tickMs int) tea.Cmd {
	return func() tea.Msg {
		payload := map[string]int{"opponent_af": opponentAF, "tick_ms": tickMs}
		body, err := doPost(socketPath, baseURL, "/api/arcade/gladiators/stables/"+stableID+"/duel", payload)
		if err != nil {
			return DuelStartErrMsg{Err: err}
		}
		var resp DuelResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return DuelStartErrMsg{Err: err}
		}
		if resp.MatchID == "" {
			return DuelStartErrMsg{Err: stableErr("daemon returned empty match_id")}
		}
		return DuelStartedMsg{MatchID: resp.MatchID}
	}
}

// TrainingStream manages an SSE connection to a training progress stream.
type TrainingStream struct {
	socketPath string
	baseURL    string
	eventChan  chan TrainingProgress
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewTrainingStream creates a new SSE stream for training progress.
func NewTrainingStream(socketPath, baseURL string) *TrainingStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &TrainingStream{
		socketPath: socketPath,
		baseURL:    baseURL,
		eventChan:  make(chan TrainingProgress, 20),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Connect starts the SSE goroutine and returns a Bubble Tea command for polling.
func (s *TrainingStream) Connect(stableID string) tea.Cmd {
	go s.readLoop(stableID)
	return s.PollCmd()
}

// PollCmd returns a Bubble Tea command that non-blocking checks for an update.
func (s *TrainingStream) PollCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case progress, ok := <-s.eventChan:
			if !ok {
				return TrainingStreamDoneMsg{}
			}
			return TrainingUpdateMsg{Progress: progress}
		default:
			return TrainingStreamContinueMsg{}
		}
	}
}

// Close cancels the SSE connection.
func (s *TrainingStream) Close() {
	s.cancel()
}

func (s *TrainingStream) readLoop(stableID string) {
	defer close(s.eventChan)

	url := s.baseURL + "/api/arcade/gladiators/stables/" + stableID + "/stream"

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

	scanner := bufio.NewScanner(resp.Body)
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

			var progress TrainingProgress
			if err := json.Unmarshal([]byte(data), &progress); err != nil {
				continue
			}

			select {
			case s.eventChan <- progress:
			case <-s.ctx.Done():
				return
			}
		}
	}
}

// HTTP helpers

func newHTTPClient(socketPath string) *http.Client {
	transport := &http.Transport{}
	if socketPath != "" {
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}
	}
	return &http.Client{Transport: transport}
}

func doGet(socketPath, baseURL, path string) ([]byte, error) {
	httpClient := newHTTPClient(socketPath)
	resp, err := httpClient.Get(baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func doPost(socketPath, baseURL, path string, payload interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		jsonBody, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequest("POST", baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := newHTTPClient(socketPath)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type stableErr string

func (e stableErr) Error() string { return string(e) }
