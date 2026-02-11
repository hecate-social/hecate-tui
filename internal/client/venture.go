package client

import (
	"encoding/json"
	"fmt"
	"os"
)

// Venture represents a business endeavor in the Hecate system.
type Venture struct {
	VentureID           string `json:"torch_id"`
	Name                string `json:"name"`
	Brief               string `json:"brief"`
	Status              int    `json:"status"`
	StatusLabel         string `json:"status_label"`
	ActiveDepartmentID  string `json:"active_cartwheel_id"`
	InitiatedAt         int64  `json:"initiated_at"`
	InitiatedBy         string `json:"initiated_by"`
}

// GetVenture returns the current (active) venture.
func (c *Client) GetVenture() (*Venture, error) {
	resp, err := c.get("/api/torch")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get venture failed: %s", resp.Error)
	}
	var venture Venture
	if err := json.Unmarshal(resp.Result, &venture); err != nil {
		return nil, fmt.Errorf("failed to parse venture: %w", err)
	}
	return &venture, nil
}

// GetVentureByID returns a specific venture by its ID.
func (c *Client) GetVentureByID(ventureID string) (*Venture, error) {
	resp, err := c.get("/api/torches/" + ventureID)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get venture failed: %s", resp.Error)
	}
	var venture Venture
	if err := json.Unmarshal(resp.Result, &venture); err != nil {
		return nil, fmt.Errorf("failed to parse venture: %w", err)
	}
	return &venture, nil
}

// ListVentures returns active (non-archived) ventures.
func (c *Client) ListVentures() ([]Venture, error) {
	return c.listVenturesInternal(false)
}

// ListAllVentures returns all ventures including archived ones.
func (c *Client) ListAllVentures() ([]Venture, error) {
	return c.listVenturesInternal(true)
}

func (c *Client) listVenturesInternal(includeArchived bool) ([]Venture, error) {
	path := "/api/torches"
	if includeArchived {
		path += "?include_archived=true"
	}
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list ventures failed: %s", resp.Error)
	}
	// Daemon returns {"ok": true, "torches": [...]}
	var result struct {
		Torches []Venture `json:"torches"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ventures: %w", err)
	}
	return result.Torches, nil
}

// InitiateVenture creates a new venture with the given name and brief.
func (c *Client) InitiateVenture(name, brief string) (*Venture, error) {
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
		return nil, fmt.Errorf("initiate venture failed: %s", resp.Error)
	}
	var venture Venture
	if err := json.Unmarshal(resp.Result, &venture); err != nil {
		return nil, fmt.Errorf("failed to parse venture: %w", err)
	}
	return &venture, nil
}

// ArchiveVenture archives a venture (soft delete).
func (c *Client) ArchiveVenture(ventureID, reason string) error {
	body := map[string]interface{}{
		"reason":      reason,
		"archived_by": "tui",
	}
	resp, err := c.post("/api/torches/"+ventureID+"/archive", body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("archive venture failed: %s", resp.Error)
	}
	return nil
}

// RefineVision refines the vision of a venture (updates brief, repos, etc.).
func (c *Client) RefineVision(ventureID string, params map[string]interface{}) error {
	resp, err := c.post("/api/torches/"+ventureID+"/vision/refine", params)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("refine vision failed: %s", resp.Error)
	}
	return nil
}

// VentureTask represents a single task in the venture task list.
type VentureTask struct {
	Verb   string `json:"verb"`
	Scope  string `json:"scope,omitempty"`
	State  string `json:"state"`
	Phase  string `json:"phase"`
	AIRole string `json:"ai_role"`
}

// VentureDivisionTasks groups tasks for a single division.
type VentureDivisionTasks struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Tasks []VentureTask `json:"tasks"`
}

// VentureTaskSummary is the venture metadata in a task list response.
type VentureTaskSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// VentureTaskList is the full response from the venture tasks endpoint.
type VentureTaskList struct {
	Venture   VentureTaskSummary     `json:"venture"`
	Tasks     []VentureTask          `json:"tasks"`
	Divisions []VentureDivisionTasks `json:"divisions"`
}

// GetVentureTasks returns the task list for a venture.
func (c *Client) GetVentureTasks(ventureID string) (*VentureTaskList, error) {
	resp, err := c.get("/api/ventures/" + ventureID + "/tasks")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get venture tasks failed: %s", resp.Error)
	}
	var taskList VentureTaskList
	if err := json.Unmarshal(resp.Result, &taskList); err != nil {
		return nil, fmt.Errorf("failed to parse venture tasks: %w", err)
	}
	return &taskList, nil
}

// SubmitVision submits the venture vision, completing the DnA phase.
func (c *Client) SubmitVision(ventureID, submittedBy string) error {
	body := map[string]interface{}{
		"submitted_by": submittedBy,
	}
	resp, err := c.post("/api/torches/"+ventureID+"/vision/submit", body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("submit vision failed: %s", resp.Error)
	}
	return nil
}
