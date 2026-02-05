package client

import (
	"encoding/json"
	"fmt"
)

// RPCResult represents the result of an RPC call.
type RPCResult struct {
	Result   json.RawMessage `json:"result"`
	Error    string          `json:"error,omitempty"`
	Duration string          `json:"duration,omitempty"`
}

// RPCCall invokes a procedure on the mesh by MRI.
func (c *Client) RPCCall(procedure string, args interface{}) (*RPCResult, error) {
	body := map[string]interface{}{
		"procedure": procedure,
	}
	if args != nil {
		body["args"] = args
	}

	resp, err := c.post("/api/rpc/call", body)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("RPC call failed: %s", resp.Error)
	}

	var result RPCResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	return &result, nil
}
