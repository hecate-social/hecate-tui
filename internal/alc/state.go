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

// TorchInfo holds information about the current torch (project).
type TorchInfo struct {
	ID          string    // e.g., "torch-abc123"
	Name        string    // e.g., "auth-system"
	Brief       string    // Short description
	InitiatedAt time.Time // When the torch was created
}

// CartwheelInfo holds information about the current cartwheel (bounded context).
type CartwheelInfo struct {
	ID           string // e.g., "prj-xyz789"
	Name         string // e.g., "user-registration"
	Description  string
	CurrentPhase Phase
	TorchID      string // Parent torch
}

// State holds the current ALC context state.
type State struct {
	Context   Context
	Torch     *TorchInfo
	Cartwheel *CartwheelInfo

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

// SetTorch switches to Torch context with the given torch.
func (s *State) SetTorch(torch *TorchInfo, source string) {
	s.Torch = torch
	s.Cartwheel = nil
	s.Context = Torch
	s.DetectionSource = source
}

// SetCartwheel switches to Cartwheel context with the given cartwheel.
// The cartwheel's parent torch must already be set.
func (s *State) SetCartwheel(cartwheel *CartwheelInfo) {
	s.Cartwheel = cartwheel
	s.Context = Cartwheel
}

// ClearCartwheel returns to Torch context, keeping the torch.
func (s *State) ClearCartwheel() {
	s.Cartwheel = nil
	if s.Torch != nil {
		s.Context = Torch
	} else {
		s.Context = Chat
	}
}

// ClearTorch returns to Chat context.
func (s *State) ClearTorch() {
	s.Torch = nil
	s.Cartwheel = nil
	s.Context = Chat
	s.DetectionSource = "manual"
}

// ActivePhase returns the current phase, or empty if no cartwheel is active.
func (s *State) ActivePhase() Phase {
	if s.Cartwheel != nil {
		return s.Cartwheel.CurrentPhase
	}
	return PhaseNone
}

// TorchName returns the current torch name, or empty string.
func (s *State) TorchName() string {
	if s.Torch != nil {
		return s.Torch.Name
	}
	return ""
}

// CartwheelName returns the current cartwheel name, or empty string.
func (s *State) CartwheelName() string {
	if s.Cartwheel != nil {
		return s.Cartwheel.Name
	}
	return ""
}
