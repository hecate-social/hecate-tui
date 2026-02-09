// Package alc provides Agent Lifecycle (ALC) context management.
// The ALC context represents what the user is working on (Chat/Torch/Cartwheel),
// which is orthogonal to the input mode (Normal/Insert/Command/etc.).
package alc

// Context represents the current ALC context level.
type Context int

const (
	// Chat mode - lightweight, no project context
	Chat Context = iota

	// Torch mode - project-level context, torch selected but no active cartwheel
	Torch

	// Cartwheel mode - active work unit with phase-specific behavior
	Cartwheel
)

// String returns the display name for the context.
func (c Context) String() string {
	switch c {
	case Chat:
		return "CHAT"
	case Torch:
		return "TORCH"
	case Cartwheel:
		return "CARTWHEEL"
	default:
		return "UNKNOWN"
	}
}

// HasTorchContext returns true if a torch is selected (Torch or Cartwheel mode).
func (c Context) HasTorchContext() bool {
	return c == Torch || c == Cartwheel
}

// HasCartwheelContext returns true if a cartwheel is active.
func (c Context) HasCartwheelContext() bool {
	return c == Cartwheel
}
