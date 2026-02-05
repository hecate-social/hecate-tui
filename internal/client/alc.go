package client

import (
	"encoding/json"
	"fmt"
)

// ALCProject represents a project in the Application Lifecycle.
type ALCProject struct {
	ProjectID             string `json:"project_id"`
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

// ALCFinding represents a discovery finding.
type ALCFinding struct {
	FindingID  string `json:"finding_id"`
	ProjectID  string `json:"project_id"`
	Category   string `json:"category"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Priority   string `json:"priority"`
	RecordedAt int64  `json:"recorded_at"`
}

// ALCTerm represents a ubiquitous language term.
type ALCTerm struct {
	TermID     string `json:"term_id"`
	ProjectID  string `json:"project_id"`
	Term       string `json:"term"`
	Definition string `json:"definition"`
	DefinedAt  int64  `json:"defined_at"`
}

// ALCDossier represents an architecture dossier (bounded context).
type ALCDossier struct {
	DossierID     string `json:"dossier_id"`
	ProjectID     string `json:"project_id"`
	DossierName   string `json:"dossier_name"`
	StreamPattern string `json:"stream_pattern"`
	Description   string `json:"description"`
	DefinedAt     int64  `json:"defined_at"`
}

// ALCSpoke represents an inventoried vertical slice.
type ALCSpoke struct {
	SpokeID       string `json:"spoke_id"`
	ProjectID     string `json:"project_id"`
	SpokeName     string `json:"spoke_name"`
	SpokeType     string `json:"spoke_type"`
	Priority      string `json:"priority"`
	DossierID     string `json:"dossier_id"`
	Description   string `json:"description"`
	InventoriedAt int64  `json:"inventoried_at"`
}

// ALCImplementation represents a spoke implementation record.
type ALCImplementation struct {
	ImplementationID    string `json:"implementation_id"`
	ProjectID           string `json:"project_id"`
	SpokeID             string `json:"spoke_id"`
	ImplementationNotes string `json:"implementation_notes"`
	ImplementedAt       int64  `json:"implemented_at"`
}

// ALCBuild represents a build verification record.
type ALCBuild struct {
	BuildID    string `json:"build_id"`
	ProjectID  string `json:"project_id"`
	Result     string `json:"result"`
	Notes      string `json:"notes"`
	VerifiedAt int64  `json:"verified_at"`
}

// ALCDeployment represents a deployment record.
type ALCDeployment struct {
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	Notes        string `json:"notes"`
	DeployedAt   int64  `json:"deployed_at"`
}

// ALCIncident represents an operational incident.
type ALCIncident struct {
	IncidentID  string `json:"incident_id"`
	ProjectID   string `json:"project_id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Resolution  string `json:"resolution"`
	ReportedAt  int64  `json:"reported_at"`
	ResolvedAt  int64  `json:"resolved_at"`
}

// ListProjects returns all ALC projects.
func (c *Client) ListProjects() ([]ALCProject, error) {
	resp, err := c.get("/alc/projects")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list projects failed: %s", resp.Error)
	}
	var projects []ALCProject
	if err := json.Unmarshal(resp.Result, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	return projects, nil
}

// GetProject returns a single ALC project by ID.
func (c *Client) GetProject(projectID string) (*ALCProject, error) {
	resp, err := c.get("/alc/projects/" + projectID)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get project failed: %s", resp.Error)
	}
	var project ALCProject
	if err := json.Unmarshal(resp.Result, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	return &project, nil
}

// ListFindings returns findings for a project's discovery phase.
func (c *Client) ListFindings(projectID string) ([]ALCFinding, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/discovery/findings")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list findings failed: %s", resp.Error)
	}
	var findings []ALCFinding
	if err := json.Unmarshal(resp.Result, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse findings: %w", err)
	}
	return findings, nil
}

// ListTerms returns terms for a project's discovery phase.
func (c *Client) ListTerms(projectID string) ([]ALCTerm, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/discovery/terms")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list terms failed: %s", resp.Error)
	}
	var terms []ALCTerm
	if err := json.Unmarshal(resp.Result, &terms); err != nil {
		return nil, fmt.Errorf("failed to parse terms: %w", err)
	}
	return terms, nil
}

// ListDossiers returns dossiers for a project's architecture phase.
func (c *Client) ListDossiers(projectID string) ([]ALCDossier, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/architecture/dossiers")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list dossiers failed: %s", resp.Error)
	}
	var dossiers []ALCDossier
	if err := json.Unmarshal(resp.Result, &dossiers); err != nil {
		return nil, fmt.Errorf("failed to parse dossiers: %w", err)
	}
	return dossiers, nil
}

// ListSpokes returns spokes for a project's architecture phase.
func (c *Client) ListSpokes(projectID string) ([]ALCSpoke, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/architecture/spokes")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list spokes failed: %s", resp.Error)
	}
	var spokes []ALCSpoke
	if err := json.Unmarshal(resp.Result, &spokes); err != nil {
		return nil, fmt.Errorf("failed to parse spokes: %w", err)
	}
	return spokes, nil
}

// ListImplementations returns implementations for a project's testing phase.
func (c *Client) ListImplementations(projectID string) ([]ALCImplementation, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/testing/implementations")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list implementations failed: %s", resp.Error)
	}
	var impls []ALCImplementation
	if err := json.Unmarshal(resp.Result, &impls); err != nil {
		return nil, fmt.Errorf("failed to parse implementations: %w", err)
	}
	return impls, nil
}

// ListBuilds returns builds for a project's testing phase.
func (c *Client) ListBuilds(projectID string) ([]ALCBuild, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/testing/builds")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list builds failed: %s", resp.Error)
	}
	var builds []ALCBuild
	if err := json.Unmarshal(resp.Result, &builds); err != nil {
		return nil, fmt.Errorf("failed to parse builds: %w", err)
	}
	return builds, nil
}

// ListDeployments returns deployments for a project's deployment phase.
func (c *Client) ListDeployments(projectID string) ([]ALCDeployment, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/deployment/deployments")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list deployments failed: %s", resp.Error)
	}
	var deployments []ALCDeployment
	if err := json.Unmarshal(resp.Result, &deployments); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %w", err)
	}
	return deployments, nil
}

// ListIncidents returns incidents for a project's deployment phase.
func (c *Client) ListIncidents(projectID string) ([]ALCIncident, error) {
	resp, err := c.get("/alc/projects/" + projectID + "/deployment/incidents")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list incidents failed: %s", resp.Error)
	}
	var incidents []ALCIncident
	if err := json.Unmarshal(resp.Result, &incidents); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}
	return incidents, nil
}

// ALCCommand sends a generic POST command to an ALC endpoint.
// This covers all mutation endpoints (initiate, start phases, record artifacts,
// complete phases, transition). The caller constructs the right path and body.
func (c *Client) ALCCommand(path string, body map[string]interface{}) error {
	resp, err := c.post(path, body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
