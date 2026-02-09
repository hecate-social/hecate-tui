package chat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

func newTestModelWithTools() Model {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	registry := llmtools.NewRegistry()
	registry.Register(llmtools.Tool{
		Name:             "read_file",
		Description:      "Read a file",
		Category:         llmtools.CategoryFileSystem,
		RequiresApproval: true,
		Parameters: llmtools.ToolParameters{
			Type: "object",
			Properties: map[string]llmtools.ParameterSpec{
				"path": {Type: "string", Description: "File path"},
			},
			Required: []string{"path"},
		},
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "file contents here", nil
	})

	registry.Register(llmtools.Tool{
		Name:             "echo",
		Description:      "Echo a message",
		Category:         "test",
		RequiresApproval: false,
		Parameters: llmtools.ToolParameters{
			Type: "object",
			Properties: map[string]llmtools.ParameterSpec{
				"message": {Type: "string", Description: "Message to echo"},
			},
		},
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "echoed", nil
	})

	permissions := llmtools.NewPermissions()
	executor := llmtools.NewExecutor(registry, permissions)
	m.SetToolExecutor(executor)
	m.EnableTools(true)

	return m
}

func TestToolsEnabled(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	if m.ToolsEnabled() {
		t.Error("ToolsEnabled() should be false by default")
	}

	m.EnableTools(true)
	// Still false because no executor
	if m.ToolsEnabled() {
		t.Error("ToolsEnabled() should be false without executor")
	}

	registry := llmtools.NewRegistry()
	permissions := llmtools.NewPermissions()
	executor := llmtools.NewExecutor(registry, permissions)
	m.SetToolExecutor(executor)
	m.EnableTools(true)

	if !m.ToolsEnabled() {
		t.Error("ToolsEnabled() should be true with executor and enabled")
	}
}

func TestHasPendingApproval_Default(t *testing.T) {
	m := newTestModelWithTools()
	if m.HasPendingApproval() {
		t.Error("HasPendingApproval() should be false initially")
	}
	if m.PendingToolCall() != nil {
		t.Error("PendingToolCall() should be nil initially")
	}
}

func TestApproveToolCall_NoPending(t *testing.T) {
	m := newTestModelWithTools()
	cmd := m.ApproveToolCall(false)
	if cmd != nil {
		t.Error("ApproveToolCall() with no pending should return nil")
	}
}

func TestDenyToolCall_NoPending(t *testing.T) {
	m := newTestModelWithTools()
	cmd := m.DenyToolCall()
	if cmd != nil {
		t.Error("DenyToolCall() with no pending should return nil")
	}
}

func TestApproveToolCall_WithPending(t *testing.T) {
	m := newTestModelWithTools()
	m.pendingToolCall = &llm.ToolCall{
		ID:   "call_123",
		Name: "read_file",
	}

	cmd := m.ApproveToolCall(false)
	if cmd == nil {
		t.Fatal("ApproveToolCall() with pending should return non-nil cmd")
	}

	msg := cmd()
	resp, ok := msg.(toolApprovalResponseMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolApprovalResponseMsg", msg)
	}
	if !resp.approved {
		t.Error("response should be approved")
	}
	if resp.grantForSession {
		t.Error("grantForSession should be false")
	}
	if resp.call.ID != "call_123" {
		t.Errorf("call.ID = %q, want %q", resp.call.ID, "call_123")
	}
}

func TestApproveToolCall_WithSessionGrant(t *testing.T) {
	m := newTestModelWithTools()
	m.pendingToolCall = &llm.ToolCall{
		ID:   "call_456",
		Name: "read_file",
	}

	cmd := m.ApproveToolCall(true)
	if cmd == nil {
		t.Fatal("ApproveToolCall(true) should return non-nil cmd")
	}

	msg := cmd()
	resp, ok := msg.(toolApprovalResponseMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolApprovalResponseMsg", msg)
	}
	if !resp.grantForSession {
		t.Error("grantForSession should be true")
	}
}

func TestDenyToolCall_WithPending(t *testing.T) {
	m := newTestModelWithTools()
	m.pendingToolCall = &llm.ToolCall{
		ID:   "call_789",
		Name: "read_file",
	}

	cmd := m.DenyToolCall()
	if cmd == nil {
		t.Fatal("DenyToolCall() with pending should return non-nil cmd")
	}

	msg := cmd()
	resp, ok := msg.(toolApprovalResponseMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolApprovalResponseMsg", msg)
	}
	if resp.approved {
		t.Error("response should not be approved")
	}
	if resp.call.ID != "call_789" {
		t.Errorf("call.ID = %q, want %q", resp.call.ID, "call_789")
	}
}

func TestHandleToolUseComplete_NilExecutor(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	call := llm.ToolCall{ID: "call_1", Name: "test"}
	cmd := m.handleToolUseComplete(call)
	if cmd == nil {
		t.Fatal("handleToolUseComplete with nil executor should return non-nil cmd")
	}

	msg := cmd()
	result, ok := msg.(toolExecutionResultMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolExecutionResultMsg", msg)
	}
	if !result.result.IsError {
		t.Error("result should be an error")
	}
	if result.result.Content != "Tool execution not available" {
		t.Errorf("content = %q, want %q", result.result.Content, "Tool execution not available")
	}
}

func TestHandleToolUseComplete_UnknownTool(t *testing.T) {
	m := newTestModelWithTools()

	call := llm.ToolCall{ID: "call_1", Name: "nonexistent_tool"}
	cmd := m.handleToolUseComplete(call)
	if cmd == nil {
		t.Fatal("handleToolUseComplete with unknown tool should return non-nil cmd")
	}

	msg := cmd()
	result, ok := msg.(toolExecutionResultMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolExecutionResultMsg", msg)
	}
	if !result.result.IsError {
		t.Error("result should be an error")
	}
	if result.result.Content != "Unknown tool: nonexistent_tool" {
		t.Errorf("content = %q, want %q", result.result.Content, "Unknown tool: nonexistent_tool")
	}
}

func TestHandleToolUseComplete_RequiresApproval(t *testing.T) {
	m := newTestModelWithTools()

	call := llm.ToolCall{ID: "call_1", Name: "read_file"}
	cmd := m.handleToolUseComplete(call)
	if cmd == nil {
		t.Fatal("handleToolUseComplete with approval-required tool should return non-nil cmd")
	}

	msg := cmd()
	_, ok := msg.(toolApprovalRequestMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolApprovalRequestMsg", msg)
	}

	if m.pendingToolCall == nil {
		t.Error("pendingToolCall should be set")
	}
	if m.pendingToolCall.Name != "read_file" {
		t.Errorf("pendingToolCall.Name = %q, want %q", m.pendingToolCall.Name, "read_file")
	}
}

func TestContinueAfterToolResult(t *testing.T) {
	m := newTestModelWithTools()
	cmd := m.ContinueAfterToolResult()
	if cmd == nil {
		t.Fatal("ContinueAfterToolResult() should return non-nil cmd")
	}

	msg := cmd()
	_, ok := msg.(toolContinueMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolContinueMsg", msg)
	}
}

func TestHandleApprovalResponse_Denied(t *testing.T) {
	m := newTestModelWithTools()
	m.pendingToolCall = &llm.ToolCall{ID: "call_1", Name: "read_file"}

	cmd := m.handleApprovalResponse(toolApprovalResponseMsg{
		approved: false,
		call:     llm.ToolCall{ID: "call_1", Name: "read_file"},
	})

	if m.pendingToolCall != nil {
		t.Error("pendingToolCall should be cleared after handling response")
	}

	if cmd == nil {
		t.Fatal("handleApprovalResponse should return non-nil cmd")
	}

	msg := cmd()
	result, ok := msg.(toolExecutionResultMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want toolExecutionResultMsg", msg)
	}
	if !result.result.IsError {
		t.Error("denied tool should produce error result")
	}
	if result.result.Content != "Tool execution denied by user" {
		t.Errorf("content = %q, want %q", result.result.Content, "Tool execution denied by user")
	}
}

func TestShowToolResult_Truncation(t *testing.T) {
	m := newTestModelWithTools()

	longContent := ""
	for i := 0; i < 600; i++ {
		longContent += "x"
	}

	m.showToolResult(llm.ToolResult{
		ToolCallID: "call_1",
		Content:    longContent,
		IsError:    false,
	})

	if len(m.messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(m.messages))
	}

	msg := m.messages[0]
	if msg.Role != "system" {
		t.Errorf("message role = %q, want %q", msg.Role, "system")
	}
	if len(msg.Content) >= len(longContent) {
		t.Error("message content should be truncated")
	}
}

func TestShowToolResult_Success(t *testing.T) {
	m := newTestModelWithTools()

	m.showToolResult(llm.ToolResult{
		ToolCallID: "call_1",
		Content:    "result data",
		IsError:    false,
	})

	if len(m.messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(m.messages))
	}
	// Success marker
	if msg := m.messages[0].Content; msg[0] != 0xe2 { // UTF-8 for checkmark
		// Just verify it contains the content
		if !contains(msg, "result data") {
			t.Errorf("message should contain result data, got %q", msg)
		}
	}
}

func TestShowToolResult_Error(t *testing.T) {
	m := newTestModelWithTools()

	m.showToolResult(llm.ToolResult{
		ToolCallID: "call_1",
		Content:    "something failed",
		IsError:    true,
	})

	if len(m.messages) != 1 {
		t.Fatalf("messages count = %d, want 1", len(m.messages))
	}
	msg := m.messages[0].Content
	if !contains(msg, "something failed") {
		t.Errorf("message should contain error content, got %q", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
