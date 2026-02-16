package client

import (
	"encoding/json"
	"fmt"
)

// Agent represents an active agent in the swarm.
type Agent struct {
	AgentID       string `json:"agent_id"`
	VentureID     string `json:"venture_id"`
	AgentType     string `json:"agent_type"`      // "specialist" or "generalist"
	Role          string `json:"role"`            // "dna", "anp", "tni", "dno"
	Status        int    `json:"status"`
	CurrentTaskID string `json:"current_task_id"`
	ActivatedAt   int64  `json:"activated_at"`
}

// ListAgents returns all active agents in the swarm.
func (c *Client) ListAgents() ([]Agent, error) {
	resp, err := c.get("/api/agents")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list agents failed: %s", resp.Error)
	}

	var result struct {
		Agents []Agent `json:"agents"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse agents response: %w", err)
	}

	return result.Agents, nil
}

// GetAgent returns a specific agent by ID.
func (c *Client) GetAgent(agentID string) (*Agent, error) {
	resp, err := c.get("/api/agents/" + agentID)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("get agent failed: %s", resp.Error)
	}

	var agent Agent
	if err := json.Unmarshal(resp.Result, &agent); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	return &agent, nil
}
