package chat

import (
	"testing"
	"time"

	"github.com/hecate-social/hecate-tui/internal/theme"
)

func TestNew_Defaults(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	if m.IsStreaming() {
		t.Error("new model should not be streaming")
	}
	if m.HasError() {
		t.Error("new model should not have error")
	}
	if m.LastError() != "" {
		t.Errorf("LastError() = %q, want empty", m.LastError())
	}
	if m.SessionTokenCount() != 0 {
		t.Errorf("SessionTokenCount() = %d, want 0", m.SessionTokenCount())
	}
	if len(m.Messages()) != 0 {
		t.Errorf("Messages() = %d, want 0", len(m.Messages()))
	}
	if m.ToolsEnabled() {
		t.Error("tools should be disabled by default")
	}
	if m.HasPendingApproval() {
		t.Error("should not have pending approval")
	}
}

func TestInjectSystemMessage(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	m.InjectSystemMessage("Hello system")

	msgs := m.Messages()
	if len(msgs) != 1 {
		t.Fatalf("Messages() = %d, want 1", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Errorf("message role = %q, want %q", msgs[0].Role, "system")
	}
	if msgs[0].Content != "Hello system" {
		t.Errorf("message content = %q, want %q", msgs[0].Content, "Hello system")
	}
}

func TestClearMessages(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	m.InjectSystemMessage("Message 1")
	m.InjectSystemMessage("Message 2")
	if len(m.Messages()) != 2 {
		t.Fatalf("Messages() = %d, want 2", len(m.Messages()))
	}

	m.ClearMessages()
	if len(m.Messages()) != 0 {
		t.Errorf("Messages() after clear = %d, want 0", len(m.Messages()))
	}
}

func TestLoadMessages(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	msgs := []Message{
		{Role: "user", Content: "Hello", Time: time.Now()},
		{Role: "assistant", Content: "Hi there!", Time: time.Now()},
	}

	m.LoadMessages(msgs)

	got := m.Messages()
	if len(got) != 2 {
		t.Fatalf("Messages() after load = %d, want 2", len(got))
	}
	if got[0].Role != "user" || got[0].Content != "Hello" {
		t.Errorf("message[0] = %+v, want user/Hello", got[0])
	}
	if got[1].Role != "assistant" || got[1].Content != "Hi there!" {
		t.Errorf("message[1] = %+v, want assistant/Hi there!", got[1])
	}
}

func TestExportMessages(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	now := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
	m.LoadMessages([]Message{
		{Role: "user", Content: "Hello", Time: now},
		{Role: "assistant", Content: "Hi", Time: now},
	})

	exported := m.ExportMessages()
	if len(exported) != 2 {
		t.Fatalf("ExportMessages() = %d, want 2", len(exported))
	}
	if exported[0].Role != "user" {
		t.Errorf("exported[0].Role = %q, want %q", exported[0].Role, "user")
	}
	if exported[0].Time != "2026-02-09 12:00:00" {
		t.Errorf("exported[0].Time = %q, want %q", exported[0].Time, "2026-02-09 12:00:00")
	}
}

func TestExportMessages_ZeroTime(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	m.LoadMessages([]Message{
		{Role: "user", Content: "Hello"},
	})

	exported := m.ExportMessages()
	if exported[0].Time != "" {
		t.Errorf("exported time for zero time = %q, want empty", exported[0].Time)
	}
}

func TestLastAssistantMessage(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	if got := m.LastAssistantMessage(); got != "" {
		t.Errorf("LastAssistantMessage() with no messages = %q, want empty", got)
	}

	m.LoadMessages([]Message{
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello!"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm great!"},
	})

	if got := m.LastAssistantMessage(); got != "I'm great!" {
		t.Errorf("LastAssistantMessage() = %q, want %q", got, "I'm great!")
	}
}

func TestClearError(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	if m.HasError() {
		t.Error("should not have error initially")
	}

	// Simulate an error via streamErrorMsg
	m, _ = m.Update(streamErrorMsg{err: errTest("test error")})

	if !m.HasError() {
		t.Error("should have error after streamErrorMsg")
	}
	if m.LastError() != "test error" {
		t.Errorf("LastError() = %q, want %q", m.LastError(), "test error")
	}

	m.ClearError()
	if m.HasError() {
		t.Error("should not have error after ClearError()")
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestRetryLast_NoMessages(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	cmd := m.RetryLast()
	if cmd != nil {
		t.Error("RetryLast() with no messages should return nil")
	}
}

func TestRetryLast_NoUserMessage(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)
	m.InjectSystemMessage("system only")

	cmd := m.RetryLast()
	if cmd != nil {
		t.Error("RetryLast() with no user messages should return nil")
	}
}

func TestRetryLast_WhileStreaming(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)
	m.streaming = true

	cmd := m.RetryLast()
	if cmd != nil {
		t.Error("RetryLast() while streaming should return nil")
	}
}

func TestSystemPromptRoundTrip(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	if got := m.GetSystemPrompt(); got != "" {
		t.Errorf("GetSystemPrompt() initially = %q, want empty", got)
	}

	m.SetSystemPrompt("Be helpful")
	if got := m.GetSystemPrompt(); got != "Be helpful" {
		t.Errorf("GetSystemPrompt() = %q, want %q", got, "Be helpful")
	}
}

func TestPreferredModel(t *testing.T) {
	th := theme.HecateDark()
	s := th.ComputeStyles()
	m := New(nil, th, s)

	m.SetPreferredModel("claude-3-opus")
	if m.preferredModel != "claude-3-opus" {
		t.Errorf("preferredModel = %q, want %q", m.preferredModel, "claude-3-opus")
	}
}
