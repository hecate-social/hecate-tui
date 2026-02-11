package client

import (
	"encoding/json"
	"fmt"
)

// Department represents a bounded context in the Application Lifecycle.
type Department struct {
	DepartmentID          string `json:"project_id"` // API still uses project_id
	Name                  string `json:"name"`
	Description           string `json:"description"`
	CurrentPhase          string `json:"current_phase"`
	Status                int    `json:"status"`
	FindingCount          int    `json:"finding_count"`
	TermCount             int    `json:"term_count"`
	DossierCount          int    `json:"dossier_count"`
	SpokeCount            int    `json:"spoke_count"`
	PlanApproved          bool   `json:"plan_approved"`
	SkeletonCreated       bool   `json:"skeleton_created"`
	ImplementedSpokeCount int    `json:"implemented_spoke_count"`
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
	DepartmentID string `json:"project_id"`
	Category     string `json:"category"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Priority     string `json:"priority"`
	RecordedAt   int64  `json:"recorded_at"`
}

// DepartmentTerm represents a ubiquitous language term.
type DepartmentTerm struct {
	TermID       string `json:"term_id"`
	DepartmentID string `json:"project_id"`
	Term         string `json:"term"`
	Definition   string `json:"definition"`
	DefinedAt    int64  `json:"defined_at"`
}

// DepartmentDossier represents an architecture dossier (bounded context).
type DepartmentDossier struct {
	DossierID     string `json:"dossier_id"`
	DepartmentID  string `json:"project_id"`
	DossierName   string `json:"dossier_name"`
	StreamPattern string `json:"stream_pattern"`
	Description   string `json:"description"`
	DefinedAt     int64  `json:"defined_at"`
}

// DepartmentSpoke represents an inventoried vertical slice.
type DepartmentSpoke struct {
	SpokeID       string `json:"spoke_id"`
	DepartmentID  string `json:"project_id"`
	SpokeName     string `json:"spoke_name"`
	SpokeType     string `json:"spoke_type"`
	Priority      string `json:"priority"`
	DossierID     string `json:"dossier_id"`
	Description   string `json:"description"`
	InventoriedAt int64  `json:"inventoried_at"`
}

// DepartmentImplementation represents a spoke implementation record.
type DepartmentImplementation struct {
	ImplementationID    string `json:"implementation_id"`
	DepartmentID        string `json:"project_id"`
	SpokeID             string `json:"spoke_id"`
	ImplementationNotes string `json:"implementation_notes"`
	ImplementedAt       int64  `json:"implemented_at"`
}

// DepartmentBuild represents a build verification record.
type DepartmentBuild struct {
	BuildID      string `json:"build_id"`
	DepartmentID string `json:"project_id"`
	Result       string `json:"result"`
	Notes        string `json:"notes"`
	VerifiedAt   int64  `json:"verified_at"`
}

// DepartmentDeployment represents a deployment record.
type DepartmentDeployment struct {
	DeploymentID string `json:"deployment_id"`
	DepartmentID string `json:"project_id"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	Notes        string `json:"notes"`
	DeployedAt   int64  `json:"deployed_at"`
}

// DepartmentIncident represents an operational incident.
type DepartmentIncident struct {
	IncidentID   string `json:"incident_id"`
	DepartmentID string `json:"project_id"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	Resolution   string `json:"resolution"`
	ReportedAt   int64  `json:"reported_at"`
	ResolvedAt   int64  `json:"resolved_at"`
}

// ListDepartments returns all departments.
func (c *Client) ListDepartments() ([]Department, error) {
	resp, err := c.get("/api/cartwheels")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list departments failed: %s", resp.Error)
	}
	var departments []Department
	if err := json.Unmarshal(resp.Result, &departments); err != nil {
		return nil, fmt.Errorf("failed to parse departments: %w", err)
	}
	return departments, nil
}

// GetDepartment returns a single department by ID.
func (c *Client) GetDepartment(departmentID string) (*Department, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID)
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
func (c *Client) ListDepartmentFindings(departmentID string) ([]DepartmentFinding, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/discovery/findings")
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
func (c *Client) ListDepartmentTerms(departmentID string) ([]DepartmentTerm, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/discovery/terms")
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
func (c *Client) ListDepartmentDossiers(departmentID string) ([]DepartmentDossier, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/architecture/dossiers")
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

// ListDepartmentSpokes returns spokes for a department's architecture phase.
func (c *Client) ListDepartmentSpokes(departmentID string) ([]DepartmentSpoke, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/architecture/spokes")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list spokes failed: %s", resp.Error)
	}
	var spokes []DepartmentSpoke
	if err := json.Unmarshal(resp.Result, &spokes); err != nil {
		return nil, fmt.Errorf("failed to parse spokes: %w", err)
	}
	return spokes, nil
}

// ListDepartmentImplementations returns implementations for a department's testing phase.
func (c *Client) ListDepartmentImplementations(departmentID string) ([]DepartmentImplementation, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/testing/implementations")
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
func (c *Client) ListDepartmentBuilds(departmentID string) ([]DepartmentBuild, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/testing/builds")
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
func (c *Client) ListDepartmentDeployments(departmentID string) ([]DepartmentDeployment, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/deployment/deployments")
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
func (c *Client) ListDepartmentIncidents(departmentID string) ([]DepartmentIncident, error) {
	resp, err := c.get("/api/cartwheels/" + departmentID + "/deployment/incidents")
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
