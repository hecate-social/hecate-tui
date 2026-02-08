package client

import (
	"encoding/json"
	"fmt"
)

// CostSummary represents LLM usage costs.
type CostSummary struct {
	TotalCost      float64 `json:"total_cost"`
	TotalTokensIn  int64   `json:"total_tokens_in"`
	TotalTokensOut int64   `json:"total_tokens_out"`
	CallCount      int64   `json:"call_count"`
}

// GetTotalCost returns the total LLM cost summary.
func (c *Client) GetTotalCost() (*CostSummary, error) {
	resp, err := c.get("/api/telemetry/cost")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("get cost failed: %s", resp.Error)
	}

	var cost CostSummary
	if err := json.Unmarshal(resp.Result, &cost); err != nil {
		return nil, fmt.Errorf("failed to parse cost response: %w", err)
	}

	return &cost, nil
}

// GetCostByTorch returns LLM cost for a specific torch.
func (c *Client) GetCostByTorch(torchID string) (*CostSummary, error) {
	resp, err := c.get("/api/telemetry/cost/" + torchID)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("get cost by torch failed: %s", resp.Error)
	}

	var cost CostSummary
	if err := json.Unmarshal(resp.Result, &cost); err != nil {
		return nil, fmt.Errorf("failed to parse cost response: %w", err)
	}

	return &cost, nil
}
