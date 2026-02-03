package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the REST client for hecate daemon API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new hecate client
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Response is the standard hecate API response
type Response struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Health represents the health check response
type Health struct {
	Status        string `json:"status"`
	UptimeSeconds int    `json:"uptime_seconds"`
	Version       string `json:"version"`
}

// Identity represents the current agent identity
type Identity struct {
	Identity  string `json:"identity"`
	PublicKey string `json:"public_key"`
	CreatedAt string `json:"created_at"`
}

// Capability represents a discovered capability
type Capability struct {
	MRI           string            `json:"mri"`
	AgentIdentity string            `json:"agent_identity"`
	Tags          []string          `json:"tags"`
	Description   string            `json:"description"`
	DemoProcedure string            `json:"demo_procedure,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	InputSchema   string            `json:"input_schema,omitempty"`
	OutputSchema  string            `json:"output_schema,omitempty"`
	AnnouncedAt   string            `json:"announced_at"`
}

// Procedure represents a registered procedure
type Procedure struct {
	Name         string `json:"name"`
	MRI          string `json:"mri"`
	Endpoint     string `json:"endpoint"`
	RegisteredAt string `json:"registered_at"`
}

// Subscription represents an active subscription
type Subscription struct {
	SubscriptionID string `json:"subscription_id"`
	ServiceMRI     string `json:"service_mri"`
	SubscribedAt   string `json:"subscribed_at"`
}

// GetHealth checks daemon health
func (c *Client) GetHealth() (*Health, error) {
	resp, err := c.get("/health")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("health check failed: %s", resp.Error)
	}

	var health Health
	if err := json.Unmarshal(resp.Result, &health); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
	}

	return &health, nil
}

// GetIdentity returns the current agent identity
func (c *Client) GetIdentity() (*Identity, error) {
	resp, err := c.get("/identity")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("identity request failed: %s", resp.Error)
	}

	var identity Identity
	if err := json.Unmarshal(resp.Result, &identity); err != nil {
		return nil, fmt.Errorf("failed to parse identity response: %w", err)
	}

	return &identity, nil
}

// DiscoverCapabilities returns discovered capabilities
func (c *Client) DiscoverCapabilities(realm, tag string, limit int) ([]Capability, error) {
	// Build request body
	reqBody := make(map[string]interface{})
	if realm != "" {
		reqBody["realm"] = realm
	}
	if tag != "" {
		reqBody["tags"] = []string{tag}
	}
	if limit > 0 {
		reqBody["limit"] = limit
	}

	resp, err := c.post("/capabilities/discover", reqBody)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("discover capabilities failed: %s", resp.Error)
	}

	var result struct {
		Capabilities []Capability `json:"capabilities"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities response: %w", err)
	}

	return result.Capabilities, nil
}

// ListProcedures returns registered procedures
// NOTE: The daemon does not have a /rpc/procedures endpoint.
// RPC tracking is done via POST /rpc/track for reputation.
// This returns empty until the daemon implements procedure listing.
func (c *Client) ListProcedures() ([]Procedure, error) {
	// Daemon doesn't have this endpoint - return empty list
	return []Procedure{}, nil
}

// PairingStatus represents the pairing status response
type PairingStatus struct {
	Status     string `json:"status"`     // "idle", "waiting", "paired", "error"
	Code       string `json:"code"`       // Pairing code to enter on realm
	ExpiresAt  string `json:"expires_at"` // When the pairing session expires
	RealmURL   string `json:"realm_url"`  // URL to complete pairing
	Message    string `json:"message"`    // Status message
}

// StartPairing initiates a pairing session
func (c *Client) StartPairing() (*PairingStatus, error) {
	resp, err := c.post("/api/pairing/start", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("start pairing failed: %s", resp.Error)
	}

	var status PairingStatus
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return nil, fmt.Errorf("failed to parse pairing response: %w", err)
	}

	return &status, nil
}

// GetPairingStatus returns the current pairing status
func (c *Client) GetPairingStatus() (*PairingStatus, error) {
	resp, err := c.get("/api/pairing/status")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("get pairing status failed: %s", resp.Error)
	}

	var status PairingStatus
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		return nil, fmt.Errorf("failed to parse pairing status: %w", err)
	}

	return &status, nil
}

// CancelPairing cancels an active pairing session
func (c *Client) CancelPairing() error {
	resp, err := c.post("/api/pairing/cancel", nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("cancel pairing failed: %s", resp.Error)
	}

	return nil
}

// ListSubscriptions returns active subscriptions
func (c *Client) ListSubscriptions() ([]Subscription, error) {
	resp, err := c.get("/subscriptions")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list subscriptions failed: %s", resp.Error)
	}

	var result struct {
		Subscriptions []Subscription `json:"subscriptions"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions response: %w", err)
	}

	return result.Subscriptions, nil
}

// get performs a GET request
func (c *Client) get(path string) (*Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// post performs a POST request with JSON body
func (c *Client) post(path string, body interface{}) (*Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}
