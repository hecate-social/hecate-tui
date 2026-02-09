package chat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

func TestBuildToolSchemas_NilExecutor(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	schemas := m.buildToolSchemas()
	if schemas != nil {
		t.Errorf("buildToolSchemas() with nil executor = %v, want nil", schemas)
	}
}

func TestBuildToolSchemas_WithTools(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	registry := llmtools.NewRegistry()
	registry.Register(llmtools.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Category:    "test",
		Parameters: llmtools.ToolParameters{
			Type: "object",
			Properties: map[string]llmtools.ParameterSpec{
				"path": {
					Type:        "string",
					Description: "File path",
				},
			},
			Required: []string{"path"},
		},
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "ok", nil
	})

	permissions := llmtools.NewPermissions()
	executor := llmtools.NewExecutor(registry, permissions)
	m.SetToolExecutor(executor)

	schemas := m.buildToolSchemas()
	if len(schemas) != 1 {
		t.Fatalf("buildToolSchemas() returned %d schemas, want 1", len(schemas))
	}

	schema := schemas[0]
	if schema.Name != "test_tool" {
		t.Errorf("schema.Name = %q, want %q", schema.Name, "test_tool")
	}
	if schema.Description != "A test tool" {
		t.Errorf("schema.Description = %q, want %q", schema.Description, "A test tool")
	}

	// Check input schema structure
	inputSchema := schema.InputSchema
	if inputSchema["type"] != "object" {
		t.Errorf("inputSchema type = %v, want %q", inputSchema["type"], "object")
	}

	props, ok := inputSchema["properties"].(map[string]llmtools.ParameterSpec)
	if !ok {
		t.Fatalf("inputSchema properties type = %T, want map[string]llmtools.ParameterSpec", inputSchema["properties"])
	}
	if _, exists := props["path"]; !exists {
		t.Error("inputSchema properties missing 'path'")
	}

	required, ok := inputSchema["required"].([]string)
	if !ok {
		t.Fatalf("inputSchema required type = %T, want []string", inputSchema["required"])
	}
	if len(required) != 1 || required[0] != "path" {
		t.Errorf("inputSchema required = %v, want [path]", required)
	}
}

func TestBuildToolSchemas_NoRequired(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	registry := llmtools.NewRegistry()
	registry.Register(llmtools.Tool{
		Name:        "simple_tool",
		Description: "No required params",
		Category:    "test",
		Parameters: llmtools.ToolParameters{
			Type:       "object",
			Properties: map[string]llmtools.ParameterSpec{},
		},
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "ok", nil
	})

	permissions := llmtools.NewPermissions()
	executor := llmtools.NewExecutor(registry, permissions)
	m.SetToolExecutor(executor)

	schemas := m.buildToolSchemas()
	if len(schemas) != 1 {
		t.Fatalf("buildToolSchemas() returned %d schemas, want 1", len(schemas))
	}

	inputSchema := schemas[0].InputSchema
	if _, exists := inputSchema["required"]; exists {
		t.Error("inputSchema should not have 'required' when no required fields")
	}
}

func TestPollStreamCmd_NilActiveStream(t *testing.T) {
	activeStream = nil

	msg := pollStreamCmd()
	done, ok := msg.(streamDoneMsg)
	if !ok {
		t.Fatalf("pollStreamCmd() with nil activeStream = %T, want streamDoneMsg", msg)
	}
	if done.reason != "activeStream was nil" {
		t.Errorf("reason = %q, want %q", done.reason, "activeStream was nil")
	}
}

func TestSendMessageWithToolResults_NoModels(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	cmd := m.sendMessageWithToolResults(nil)
	if cmd == nil {
		t.Fatal("sendMessageWithToolResults() returned nil cmd, want non-nil")
	}

	msg := cmd()
	errMsg, ok := msg.(streamErrorMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want streamErrorMsg", msg)
	}
	if errMsg.err.Error() != "no models available" {
		t.Errorf("err = %q, want %q", errMsg.err.Error(), "no models available")
	}
}

func TestPreferredModelApplied(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	m.SetPreferredModel("gpt-4o")

	// Simulate models arriving
	m.models = []llm.Model{
		{Name: "llama3:latest"},
		{Name: "gpt-4o"},
		{Name: "claude-3-opus"},
	}

	// Trigger the preferred model selection via Update
	m, _ = m.Update(modelsMsg{models: m.models})

	if got := m.ActiveModelName(); got != "gpt-4o" {
		t.Errorf("ActiveModelName() after preferred = %q, want %q", got, "gpt-4o")
	}
}
