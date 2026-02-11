package alc

import "time"

// Phase represents an ALC lifecycle phase.
type Phase string

const (
	PhaseNone         Phase = ""
	PhaseDiscovery    Phase = "dna"
	PhaseArchitecture Phase = "anp"
	PhaseTesting      Phase = "tni"
	PhaseDeployment   Phase = "dno"
)

// PhaseDisplayName returns the full display name for a phase.
func (p Phase) DisplayName() string {
	switch p {
	case PhaseDiscovery:
		return "Discovery & Analysis"
	case PhaseArchitecture:
		return "Architecture & Planning"
	case PhaseTesting:
		return "Testing & Implementation"
	case PhaseDeployment:
		return "Deployment & Operations"
	default:
		return ""
	}
}

// VentureInfo holds information about the current venture (project).
type VentureInfo struct {
	ID          string    // e.g., "venture-abc123"
	Name        string    // e.g., "auth-system"
	Brief       string    // Short description
	InitiatedAt time.Time // When the venture was created
}

// DepartmentInfo holds information about the current department (bounded context).
type DepartmentInfo struct {
	ID           string // e.g., "prj-xyz789"
	Name         string // e.g., "user-registration"
	Description  string
	CurrentPhase Phase
	VentureID    string // Parent venture
}

// State holds the current ALC context state.
type State struct {
	Context    Context
	Venture    *VentureInfo
	Department *DepartmentInfo

	// Source indicates how the context was detected
	// "manual" = user selected, "git" = auto-detected from git remote, "config" = from .hecate/torch.json
	DetectionSource string
}

// NewState creates a new ALC state in Chat context.
func NewState() *State {
	return &State{
		Context:         Chat,
		DetectionSource: "manual",
	}
}

// SetVenture switches to Venture context with the given venture.
func (s *State) SetVenture(venture *VentureInfo, source string) {
	s.Venture = venture
	s.Department = nil
	s.Context = Venture
	s.DetectionSource = source
}

// SetDepartment switches to Department context with the given department.
// The department's parent venture must already be set.
func (s *State) SetDepartment(department *DepartmentInfo) {
	s.Department = department
	s.Context = Department
}

// ClearDepartment returns to Venture context, keeping the venture.
func (s *State) ClearDepartment() {
	s.Department = nil
	if s.Venture != nil {
		s.Context = Venture
	} else {
		s.Context = Chat
	}
}

// ClearVenture returns to Chat context.
func (s *State) ClearVenture() {
	s.Venture = nil
	s.Department = nil
	s.Context = Chat
	s.DetectionSource = "manual"
}

// ActivePhase returns the current phase, or empty if no department is active.
func (s *State) ActivePhase() Phase {
	if s.Department != nil {
		return s.Department.CurrentPhase
	}
	return PhaseNone
}

// VentureName returns the current venture name, or empty string.
func (s *State) VentureName() string {
	if s.Venture != nil {
		return s.Venture.Name
	}
	return ""
}

// DepartmentName returns the current department name, or empty string.
func (s *State) DepartmentName() string {
	if s.Department != nil {
		return s.Department.Name
	}
	return ""
}
