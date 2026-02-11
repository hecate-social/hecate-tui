package client

import (
	"encoding/json"
	"fmt"
)

// Department represents a bounded context in the Application Lifecycle.
type Department struct {
	DepartmentID          string `json:"division_id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	CurrentPhase          string `json:"current_phase"`
	Status                int    `json:"status"`
	FindingCount          int    `json:"finding_count"`
	TermCount             int    `json:"term_count"`
	DossierCount          int    `json:"dossier_count"`
	DeskCount             int    `json:"desk_count"`
	PlanApproved          bool   `json:"plan_approved"`
	SkeletonCreated       bool   `json:"skeleton_created"`
	ImplementedDeskCount  int    `json:"implemented_desk_count"`
	BuildVerified         bool   `json:"build_verified"`
	DeploymentCount       int    `json:"deployment_count"`
	ActiveIncidents       int    `json:"active_incidents"`
	InitiatedAt           int64  `json:"initiated_at"`
	PhaseStartedAt        int64  `json:"phase_started_at"`
	CompletedAt           int64  `json:"completed_at"`
}

// DepartmentFinding represents a discovery finding.
type DepartmentFinding struct {
	FindingID    string `json:"finding_id"`
	DepartmentID string `json:"division_id"`
	Category     string `json:"category"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Priority     string `json:"priority"`
	RecordedAt   int64  `json:"recorded_at"`
}

// DepartmentTerm represents a ubiquitous language term.
type DepartmentTerm struct {
	TermID       string `json:"term_id"`
	DepartmentID string `json:"division_id"`
	Term         string `json:"term"`
	Definition   string `json:"definition"`
	DefinedAt    int64  `json:"defined_at"`
}

// DepartmentDossier represents an architecture dossier (bounded context).
type DepartmentDossier struct {
	DossierID     string `json:"dossier_id"`
	DepartmentID  string `json:"division_id"`
	DossierName   string `json:"dossier_name"`
	StreamPattern string `json:"stream_pattern"`
	Description   string `json:"description"`
	DefinedAt     int64  `json:"defined_at"`
}

// DepartmentDesk represents an inventoried vertical slice.
type DepartmentDesk struct {
	DeskID       string `json:"desk_id"`
	DepartmentID string `json:"division_id"`
	DeskName     string `json:"desk_name"`
	DeskType     string `json:"desk_type"`
	Priority     string `json:"priority"`
	DossierID    string `json:"dossier_id"`
	Description  string `json:"description"`
	InventoriedAt int64 `json:"inventoried_at"`
}

// DepartmentImplementation represents a desk implementation record.
type DepartmentImplementation struct {
	ImplementationID    string `json:"implementation_id"`
	DepartmentID        string `json:"division_id"`
	DeskID              string `json:"desk_id"`
	ImplementationNotes string `json:"implementation_notes"`
	ImplementedAt       int64  `json:"implemented_at"`
}

// DepartmentBuild represents a build verification record.
type DepartmentBuild struct {
	BuildID      string `json:"build_id"`
	DepartmentID string `json:"division_id"`
	Result       string `json:"result"`
	Notes        string `json:"notes"`
	VerifiedAt   int64  `json:"verified_at"`
}

// DepartmentDeployment represents a deployment record.
type DepartmentDeployment struct {
	DeploymentID string `json:"deployment_id"`
	DepartmentID string `json:"division_id"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	Notes        string `json:"notes"`
	DeployedAt   int64  `json:"deployed_at"`
}

// DepartmentIncident represents an operational incident.
type DepartmentIncident struct {
	IncidentID   string `json:"incident_id"`
	DepartmentID string `json:"division_id"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	Resolution   string `json:"resolution"`
	ReportedAt   int64  `json:"reported_at"`
	ResolvedAt   int64  `json:"resolved_at"`
}

// divisionPath builds the API path prefix for a division under a venture.
func divisionPath(ventureID, divisionID string) string {
	return "/api/ventures/" + ventureID + "/divisions/" + divisionID
}

// ListDepartments returns all divisions for a venture.
func (c *Client) ListDepartments(ventureID string) ([]Department, error) {
	resp, err := c.get("/api/ventures/" + ventureID + "/divisions")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list departments failed: %s", resp.Error)
	}
	var result struct {
		Divisions []Department `json:"divisions"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse departments: %w", err)
	}
	return result.Divisions, nil
}

// GetDepartment returns a single division by ID.
func (c *Client) GetDepartment(ventureID, departmentID string) (*Department, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID))
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get department failed: %s", resp.Error)
	}
	var department Department
	if err := json.Unmarshal(resp.Result, &department); err != nil {
		return nil, fmt.Errorf("failed to parse department: %w", err)
	}
	return &department, nil
}

// ListDepartmentFindings returns findings for a department's discovery phase.
func (c *Client) ListDepartmentFindings(ventureID, departmentID string) ([]DepartmentFinding, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/discovery/findings")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list findings failed: %s", resp.Error)
	}
	var findings []DepartmentFinding
	if err := json.Unmarshal(resp.Result, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse findings: %w", err)
	}
	return findings, nil
}

// ListDepartmentTerms returns terms for a department's discovery phase.
func (c *Client) ListDepartmentTerms(ventureID, departmentID string) ([]DepartmentTerm, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/discovery/terms")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list terms failed: %s", resp.Error)
	}
	var terms []DepartmentTerm
	if err := json.Unmarshal(resp.Result, &terms); err != nil {
		return nil, fmt.Errorf("failed to parse terms: %w", err)
	}
	return terms, nil
}

// ListDepartmentDossiers returns dossiers for a department's architecture phase.
func (c *Client) ListDepartmentDossiers(ventureID, departmentID string) ([]DepartmentDossier, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/design/dossiers")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list dossiers failed: %s", resp.Error)
	}
	var dossiers []DepartmentDossier
	if err := json.Unmarshal(resp.Result, &dossiers); err != nil {
		return nil, fmt.Errorf("failed to parse dossiers: %w", err)
	}
	return dossiers, nil
}

// ListDepartmentDesks returns desks for a department's architecture phase.
func (c *Client) ListDepartmentDesks(ventureID, departmentID string) ([]DepartmentDesk, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/design/desks")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list desks failed: %s", resp.Error)
	}
	var desks []DepartmentDesk
	if err := json.Unmarshal(resp.Result, &desks); err != nil {
		return nil, fmt.Errorf("failed to parse desks: %w", err)
	}
	return desks, nil
}

// ListDepartmentImplementations returns implementations for a department's testing phase.
func (c *Client) ListDepartmentImplementations(ventureID, departmentID string) ([]DepartmentImplementation, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/test/implementations")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list implementations failed: %s", resp.Error)
	}
	var impls []DepartmentImplementation
	if err := json.Unmarshal(resp.Result, &impls); err != nil {
		return nil, fmt.Errorf("failed to parse implementations: %w", err)
	}
	return impls, nil
}

// ListDepartmentBuilds returns builds for a department's testing phase.
func (c *Client) ListDepartmentBuilds(ventureID, departmentID string) ([]DepartmentBuild, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/test/builds")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list builds failed: %s", resp.Error)
	}
	var builds []DepartmentBuild
	if err := json.Unmarshal(resp.Result, &builds); err != nil {
		return nil, fmt.Errorf("failed to parse builds: %w", err)
	}
	return builds, nil
}

// ListDepartmentDeployments returns deployments for a department's deployment phase.
func (c *Client) ListDepartmentDeployments(ventureID, departmentID string) ([]DepartmentDeployment, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/deploy/deployments")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list deployments failed: %s", resp.Error)
	}
	var deployments []DepartmentDeployment
	if err := json.Unmarshal(resp.Result, &deployments); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %w", err)
	}
	return deployments, nil
}

// ListDepartmentIncidents returns incidents for a department's deployment phase.
func (c *Client) ListDepartmentIncidents(ventureID, departmentID string) ([]DepartmentIncident, error) {
	resp, err := c.get(divisionPath(ventureID, departmentID) + "/deploy/incidents")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list incidents failed: %s", resp.Error)
	}
	var incidents []DepartmentIncident
	if err := json.Unmarshal(resp.Result, &incidents); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}
	return incidents, nil
}

// DepartmentCommand sends a generic POST command to a department endpoint.
// This covers all mutation endpoints (initiate, start phases, record artifacts,
// complete phases, transition). The caller constructs the right path and body.
func (c *Client) DepartmentCommand(path string, body map[string]interface{}) error {
	resp, err := c.post(path, body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
