// Package alc provides Agent Lifecycle (ALC) context management.
// The ALC context represents what the user is working on (Chat/Venture/Department),
// which is orthogonal to the input mode (Normal/Insert/Command/etc.).
package alc

// Context represents the current ALC context level.
type Context int

const (
	// Chat mode - lightweight, no project context
	Chat Context = iota

	// Venture mode - project-level context, venture selected but no active department
	Venture

	// Department mode - active work unit with phase-specific behavior
	Department
)

// String returns the display name for the context.
func (c Context) String() string {
	switch c {
	case Chat:
		return "CHAT"
	case Venture:
		return "VENTURE"
	case Department:
		return "DEPARTMENT"
	default:
		return "UNKNOWN"
	}
}

// HasVentureContext returns true if a venture is selected (Venture or Department mode).
func (c Context) HasVentureContext() bool {
	return c == Venture || c == Department
}

// HasDepartmentContext returns true if a department is active.
func (c Context) HasDepartmentContext() bool {
	return c == Department
}
