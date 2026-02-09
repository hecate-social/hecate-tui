package client

import (
	"encoding/json"
	"fmt"
)

// Torch represents a business endeavor in the Hecate system.
type Torch struct {
	TorchID           string `json:"torch_id"`
	Name              string `json:"name"`
	Brief             string `json:"brief"`
	Status            int    `json:"status"`
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

// ListTorches returns all torches.
func (c *Client) ListTorches() ([]Torch, error) {
	resp, err := c.get("/api/torches")
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
	body := map[string]interface{}{
		"name":  name,
		"brief": brief,
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
