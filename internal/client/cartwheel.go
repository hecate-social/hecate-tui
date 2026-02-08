package client

import (
	"encoding/json"
	"fmt"
)

// Cartwheel represents a bounded context in the Application Lifecycle.
type Cartwheel struct {
	CartwheelID           string `json:"project_id"` // API still uses project_id
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

// CartwheelFinding represents a discovery finding.
type CartwheelFinding struct {
	FindingID    string `json:"finding_id"`
	CartwheelID  string `json:"project_id"`
	Category     string `json:"category"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Priority     string `json:"priority"`
	RecordedAt   int64  `json:"recorded_at"`
}

// CartwheelTerm represents a ubiquitous language term.
type CartwheelTerm struct {
	TermID      string `json:"term_id"`
	CartwheelID string `json:"project_id"`
	Term        string `json:"term"`
	Definition  string `json:"definition"`
	DefinedAt   int64  `json:"defined_at"`
}

// CartwheelDossier represents an architecture dossier (bounded context).
type CartwheelDossier struct {
	DossierID     string `json:"dossier_id"`
	CartwheelID   string `json:"project_id"`
	DossierName   string `json:"dossier_name"`
	StreamPattern string `json:"stream_pattern"`
	Description   string `json:"description"`
	DefinedAt     int64  `json:"defined_at"`
}

// CartwheelSpoke represents an inventoried vertical slice.
type CartwheelSpoke struct {
	SpokeID       string `json:"spoke_id"`
	CartwheelID   string `json:"project_id"`
	SpokeName     string `json:"spoke_name"`
	SpokeType     string `json:"spoke_type"`
	Priority      string `json:"priority"`
	DossierID     string `json:"dossier_id"`
	Description   string `json:"description"`
	InventoriedAt int64  `json:"inventoried_at"`
}

// CartwheelImplementation represents a spoke implementation record.
type CartwheelImplementation struct {
	ImplementationID    string `json:"implementation_id"`
	CartwheelID         string `json:"project_id"`
	SpokeID             string `json:"spoke_id"`
	ImplementationNotes string `json:"implementation_notes"`
	ImplementedAt       int64  `json:"implemented_at"`
}

// CartwheelBuild represents a build verification record.
type CartwheelBuild struct {
	BuildID     string `json:"build_id"`
	CartwheelID string `json:"project_id"`
	Result      string `json:"result"`
	Notes       string `json:"notes"`
	VerifiedAt  int64  `json:"verified_at"`
}

// CartwheelDeployment represents a deployment record.
type CartwheelDeployment struct {
	DeploymentID string `json:"deployment_id"`
	CartwheelID  string `json:"project_id"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	Notes        string `json:"notes"`
	DeployedAt   int64  `json:"deployed_at"`
}

// CartwheelIncident represents an operational incident.
type CartwheelIncident struct {
	IncidentID  string `json:"incident_id"`
	CartwheelID string `json:"project_id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Resolution  string `json:"resolution"`
	ReportedAt  int64  `json:"reported_at"`
	ResolvedAt  int64  `json:"resolved_at"`
}

// ListCartwheels returns all cartwheels.
func (c *Client) ListCartwheels() ([]Cartwheel, error) {
	resp, err := c.get("/api/cartwheels")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list cartwheels failed: %s", resp.Error)
	}
	var cartwheels []Cartwheel
	if err := json.Unmarshal(resp.Result, &cartwheels); err != nil {
		return nil, fmt.Errorf("failed to parse cartwheels: %w", err)
	}
	return cartwheels, nil
}

// GetCartwheel returns a single cartwheel by ID.
func (c *Client) GetCartwheel(cartwheelID string) (*Cartwheel, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("get cartwheel failed: %s", resp.Error)
	}
	var cartwheel Cartwheel
	if err := json.Unmarshal(resp.Result, &cartwheel); err != nil {
		return nil, fmt.Errorf("failed to parse cartwheel: %w", err)
	}
	return &cartwheel, nil
}

// ListCartwheelFindings returns findings for a cartwheel's discovery phase.
func (c *Client) ListCartwheelFindings(cartwheelID string) ([]CartwheelFinding, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/discovery/findings")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list findings failed: %s", resp.Error)
	}
	var findings []CartwheelFinding
	if err := json.Unmarshal(resp.Result, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse findings: %w", err)
	}
	return findings, nil
}

// ListCartwheelTerms returns terms for a cartwheel's discovery phase.
func (c *Client) ListCartwheelTerms(cartwheelID string) ([]CartwheelTerm, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/discovery/terms")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list terms failed: %s", resp.Error)
	}
	var terms []CartwheelTerm
	if err := json.Unmarshal(resp.Result, &terms); err != nil {
		return nil, fmt.Errorf("failed to parse terms: %w", err)
	}
	return terms, nil
}

// ListCartwheelDossiers returns dossiers for a cartwheel's architecture phase.
func (c *Client) ListCartwheelDossiers(cartwheelID string) ([]CartwheelDossier, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/architecture/dossiers")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list dossiers failed: %s", resp.Error)
	}
	var dossiers []CartwheelDossier
	if err := json.Unmarshal(resp.Result, &dossiers); err != nil {
		return nil, fmt.Errorf("failed to parse dossiers: %w", err)
	}
	return dossiers, nil
}

// ListCartwheelSpokes returns spokes for a cartwheel's architecture phase.
func (c *Client) ListCartwheelSpokes(cartwheelID string) ([]CartwheelSpoke, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/architecture/spokes")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list spokes failed: %s", resp.Error)
	}
	var spokes []CartwheelSpoke
	if err := json.Unmarshal(resp.Result, &spokes); err != nil {
		return nil, fmt.Errorf("failed to parse spokes: %w", err)
	}
	return spokes, nil
}

// ListCartwheelImplementations returns implementations for a cartwheel's testing phase.
func (c *Client) ListCartwheelImplementations(cartwheelID string) ([]CartwheelImplementation, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/testing/implementations")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list implementations failed: %s", resp.Error)
	}
	var impls []CartwheelImplementation
	if err := json.Unmarshal(resp.Result, &impls); err != nil {
		return nil, fmt.Errorf("failed to parse implementations: %w", err)
	}
	return impls, nil
}

// ListCartwheelBuilds returns builds for a cartwheel's testing phase.
func (c *Client) ListCartwheelBuilds(cartwheelID string) ([]CartwheelBuild, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/testing/builds")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list builds failed: %s", resp.Error)
	}
	var builds []CartwheelBuild
	if err := json.Unmarshal(resp.Result, &builds); err != nil {
		return nil, fmt.Errorf("failed to parse builds: %w", err)
	}
	return builds, nil
}

// ListCartwheelDeployments returns deployments for a cartwheel's deployment phase.
func (c *Client) ListCartwheelDeployments(cartwheelID string) ([]CartwheelDeployment, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/deployment/deployments")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list deployments failed: %s", resp.Error)
	}
	var deployments []CartwheelDeployment
	if err := json.Unmarshal(resp.Result, &deployments); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %w", err)
	}
	return deployments, nil
}

// ListCartwheelIncidents returns incidents for a cartwheel's deployment phase.
func (c *Client) ListCartwheelIncidents(cartwheelID string) ([]CartwheelIncident, error) {
	resp, err := c.get("/api/cartwheels/" + cartwheelID + "/deployment/incidents")
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("list incidents failed: %s", resp.Error)
	}
	var incidents []CartwheelIncident
	if err := json.Unmarshal(resp.Result, &incidents); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}
	return incidents, nil
}

// CartwheelCommand sends a generic POST command to a cartwheel endpoint.
// This covers all mutation endpoints (initiate, start phases, record artifacts,
// complete phases, transition). The caller constructs the right path and body.
func (c *Client) CartwheelCommand(path string, body map[string]interface{}) error {
	resp, err := c.post(path, body)
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
