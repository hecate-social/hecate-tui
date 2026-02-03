package projects

import "time"

// Phase represents a development phase
type Phase string

const (
	PhaseAnD Phase = "and" // Analysis & Discovery
	PhaseAnP Phase = "anp" // Architecture & Planning
	PhaseInT Phase = "int" // Implementation & Testing
	PhaseDoO Phase = "doo" // Deployment & Operations
)

// PhaseInfo contains metadata about a phase
type PhaseInfo struct {
	Phase       Phase
	Name        string
	ShortName   string
	Icon        string
	Description string
}

// Phases returns all phase info in order
func Phases() []PhaseInfo {
	return []PhaseInfo{
		{PhaseAnD, "Analysis & Discovery", "AnD", "📊", "Understand the problem and explore solutions"},
		{PhaseAnP, "Architecture & Planning", "AnP", "🏗️", "Design the system and plan implementation"},
		{PhaseInT, "Implementation & Testing", "InT", "⚡", "Build and verify the solution"},
		{PhaseDoO, "Deployment & Operations", "DoO", "🚀", "Ship and maintain in production"},
	}
}

// ProjectType indicates how the project was detected
type ProjectType int

const (
	ProjectTypeUnknown ProjectType = iota
	ProjectTypeGit                 // Has .git directory
	ProjectTypeHecate              // Has HECATE.md file
	ProjectTypeBoth                // Has both git and HECATE.md
)

// Project represents a detected project
type Project struct {
	Name        string
	Path        string
	Type        ProjectType
	CurrentPhase Phase
	HasWorkspace bool      // Has .hecate/ directory
	DetectedAt   time.Time

	// Git info (if available)
	GitBranch   string
	GitRemote   string

	// Hecate info (if available)
	HecateTitle string
	HecatePhase string
}

// TypeString returns a human-readable type
func (p Project) TypeString() string {
	switch p.Type {
	case ProjectTypeGit:
		return "Git"
	case ProjectTypeHecate:
		return "Hecate"
	case ProjectTypeBoth:
		return "Git+Hecate"
	default:
		return "Unknown"
	}
}

// TypeIcon returns an icon for the type
func (p Project) TypeIcon() string {
	switch p.Type {
	case ProjectTypeGit:
		return "󰊢"
	case ProjectTypeHecate:
		return "🗝️"
	case ProjectTypeBoth:
		return "⚡"
	default:
		return "📁"
	}
}

// PhaseIcon returns the icon for current phase
func (p Project) PhaseIcon() string {
	for _, info := range Phases() {
		if info.Phase == p.CurrentPhase {
			return info.Icon
		}
	}
	return "📊"
}
