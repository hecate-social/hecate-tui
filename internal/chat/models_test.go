package chat

import (
	"testing"

	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

func newTestModel(models []llm.Model) Model {
	t := theme.HecateDark()
	s := t.ComputeStyles()
	m := New(nil, t, s)
	m.models = models
	return m
}

var testModels = []llm.Model{
	{Name: "llama3:latest", Provider: "ollama"},
	{Name: "gpt-4o", Provider: "openai"},
	{Name: "claude-3-opus", Provider: "anthropic"},
}

func TestActiveModelName_Empty(t *testing.T) {
	m := newTestModel(nil)
	if got := m.ActiveModelName(); got != "" {
		t.Errorf("ActiveModelName() with no models = %q, want empty", got)
	}
}

func TestActiveModelName_Default(t *testing.T) {
	m := newTestModel(testModels)
	if got := m.ActiveModelName(); got != "llama3:latest" {
		t.Errorf("ActiveModelName() = %q, want %q", got, "llama3:latest")
	}
}

func TestActiveModelProvider(t *testing.T) {
	m := newTestModel(testModels)
	if got := m.ActiveModelProvider(); got != "ollama" {
		t.Errorf("ActiveModelProvider() = %q, want %q", got, "ollama")
	}
}

func TestCycleModel(t *testing.T) {
	m := newTestModel(testModels)

	m.CycleModel()
	if got := m.ActiveModelName(); got != "gpt-4o" {
		t.Errorf("after CycleModel() = %q, want %q", got, "gpt-4o")
	}

	m.CycleModel()
	if got := m.ActiveModelName(); got != "claude-3-opus" {
		t.Errorf("after 2nd CycleModel() = %q, want %q", got, "claude-3-opus")
	}

	// Wrap around
	m.CycleModel()
	if got := m.ActiveModelName(); got != "llama3:latest" {
		t.Errorf("after wrap CycleModel() = %q, want %q", got, "llama3:latest")
	}
}

func TestCycleModelReverse(t *testing.T) {
	m := newTestModel(testModels)

	// Reverse from 0 wraps to last
	m.CycleModelReverse()
	if got := m.ActiveModelName(); got != "claude-3-opus" {
		t.Errorf("after CycleModelReverse() = %q, want %q", got, "claude-3-opus")
	}

	m.CycleModelReverse()
	if got := m.ActiveModelName(); got != "gpt-4o" {
		t.Errorf("after 2nd CycleModelReverse() = %q, want %q", got, "gpt-4o")
	}
}

func TestCycleModel_Empty(t *testing.T) {
	m := newTestModel(nil)
	// Should not panic
	m.CycleModel()
	m.CycleModelReverse()
}

func TestSwitchModel_ExactMatch(t *testing.T) {
	m := newTestModel(testModels)

	m.SwitchModel("gpt-4o")
	if got := m.ActiveModelName(); got != "gpt-4o" {
		t.Errorf("SwitchModel(gpt-4o) = %q, want %q", got, "gpt-4o")
	}
}

func TestSwitchModel_PrefixMatch(t *testing.T) {
	m := newTestModel(testModels)

	m.SwitchModel("claude")
	if got := m.ActiveModelName(); got != "claude-3-opus" {
		t.Errorf("SwitchModel(claude) = %q, want %q", got, "claude-3-opus")
	}
}

func TestSwitchModel_CaseInsensitive(t *testing.T) {
	m := newTestModel(testModels)

	m.SwitchModel("GPT-4O")
	if got := m.ActiveModelName(); got != "gpt-4o" {
		t.Errorf("SwitchModel(GPT-4O) = %q, want %q", got, "gpt-4o")
	}
}

func TestSwitchModel_NotFound(t *testing.T) {
	m := newTestModel(testModels)

	m.SwitchModel("nonexistent")
	// Should stay on first model
	if got := m.ActiveModelName(); got != "llama3:latest" {
		t.Errorf("SwitchModel(nonexistent) = %q, want %q", got, "llama3:latest")
	}
}

func TestIsPaidProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     bool
	}{
		{"anthropic", true},
		{"openai", true},
		{"google", true},
		{"groq", true},
		{"together", true},
		{"ollama", false},
		{"", false},
	}

	for _, tt := range tests {
		m := newTestModel([]llm.Model{{Name: "test", Provider: tt.provider}})
		if got := m.IsPaidProvider(); got != tt.want {
			t.Errorf("IsPaidProvider() with provider=%q = %v, want %v", tt.provider, got, tt.want)
		}
	}
}

func TestActiveModelProvider_Empty(t *testing.T) {
	m := newTestModel(nil)
	if got := m.ActiveModelProvider(); got != "" {
		t.Errorf("ActiveModelProvider() with no models = %q, want empty", got)
	}
}
