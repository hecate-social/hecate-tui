package client

import (
	"encoding/json"
	"fmt"
	"os"
)

// Torch represents a business endeavor in the Hecate system.
type Torch struct {
	TorchID           string `json:"torch_id"`
	Name              string `json:"name"`
	Brief             string `json:"brief"`
	Status            int    `json:"status"`
	StatusLabel       string `json:"status_label"`
	ActiveCartwheelID string `json:"active_cartwheel_id"`
	InitiatedAt       int64  `json:"initiated_at"`
	InitiatedBy       string `json:"initiated_by"`
}

// GetTorch returns the current (active) torch.
func (c *Client) GetTorch() (*Torch, error) {
	resp, err := c.get("/api/torch")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get torch failed: %s", resp.Error)
	}
	var torch Torch
	if err := json.Unmarshal(resp.Result, &torch); err != nil {
		return nil, fmt.Errorf("failed to parse torch: %w", err)
	}
	return &torch, nil
}

// GetTorchByID returns a specific torch by its ID.
func (c *Client) GetTorchByID(torchID string) (*Torch, error) {
	resp, err := c.get("/api/torches/" + torchID)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get torch failed: %s", resp.Error)
	}
	var torch Torch
	if err := json.Unmarshal(resp.Result, &torch); err != nil {
		return nil, fmt.Errorf("failed to parse torch: %w", err)
	}
	return &torch, nil
}

// ListTorches returns active (non-archived) torches.
func (c *Client) ListTorches() ([]Torch, error) {
	return c.listTorchesInternal(false)
}

// ListAllTorches returns all torches including archived ones.
func (c *Client) ListAllTorches() ([]Torch, error) {
	return c.listTorchesInternal(true)
}

func (c *Client) listTorchesInternal(includeArchived bool) ([]Torch, error) {
	path := "/api/torches"
	if includeArchived {
		path += "?include_archived=true"
	}
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list torches failed: %s", resp.Error)
	}
	// Daemon returns {"ok": true, "torches": [...]}
	var result struct {
		Torches []Torch `json:"torches"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse torches: %w", err)
	}
	return result.Torches, nil
}

// InitiateTorch creates a new torch with the given name and brief.
func (c *Client) InitiateTorch(name, brief string) (*Torch, error) {
	// Get user@hostname for initiated_by
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	initiatedBy := user + "@" + hostname

	body := map[string]interface{}{
		"name":         name,
		"brief":        brief,
		"initiated_by": initiatedBy,
	}
	resp, err := c.post("/api/torch/initiate", body)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("initiate torch failed: %s", resp.Error)
	}
	var torch Torch
	if err := json.Unmarshal(resp.Result, &torch); err != nil {
		return nil, fmt.Errorf("failed to parse torch: %w", err)
	}
	return &torch, nil
}

// ArchiveTorch archives a torch (soft delete).
func (c *Client) ArchiveTorch(torchID, reason string) error {
	body := map[string]interface{}{
		"reason":      reason,
		"archived_by": "tui",
	}
	resp, err := c.post("/api/torches/"+torchID+"/archive", body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("archive torch failed: %s", resp.Error)
	}
	return nil
}

// RefineVision refines the vision of a torch (updates brief, repos, etc.).
func (c *Client) RefineVision(torchID string, params map[string]interface{}) error {
	resp, err := c.post("/api/torches/"+torchID+"/vision/refine", params)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("refine vision failed: %s", resp.Error)
	}
	return nil
}

// SubmitVision submits the torch vision, completing the DnA phase.
func (c *Client) SubmitVision(torchID, submittedBy string) error {
	body := map[string]interface{}{
		"submitted_by": submittedBy,
	}
	resp, err := c.post("/api/torches/"+torchID+"/vision/submit", body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("submit vision failed: %s", resp.Error)
	}
	return nil
}
