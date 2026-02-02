package client

import (
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
	path := "/capabilities/discover"
	if realm != "" || tag != "" || limit > 0 {
		path += "?"
		if realm != "" {
			path += "realm=" + realm + "&"
		}
		if tag != "" {
			path += "tag=" + tag + "&"
		}
		if limit > 0 {
			path += fmt.Sprintf("limit=%d", limit)
		}
	}

	resp, err := c.get(path)
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
func (c *Client) ListProcedures() ([]Procedure, error) {
	resp, err := c.get("/rpc/procedures")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list procedures failed: %s", resp.Error)
	}

	var result struct {
		Procedures []Procedure `json:"procedures"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse procedures response: %w", err)
	}

	return result.Procedures, nil
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
