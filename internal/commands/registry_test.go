package commands

import (
	"testing"
)

func TestNewRegistry_HasCommands(t *testing.T) {
	r := NewRegistry()
	cmds := r.List()
	if len(cmds) == 0 {
		t.Fatal("NewRegistry() should register built-in commands")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}

	r.Register(&ClearCmd{})

	cmds := r.List()
	if len(cmds) != 1 {
		t.Fatalf("List() after Register = %d, want 1", len(cmds))
	}
	if cmds[0].Name() != "clear" {
		t.Errorf("command name = %q, want %q", cmds[0].Name(), "clear")
	}
}

func TestRegistry_DispatchByName(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("/clear", nil)
	if cmd == nil {
		t.Fatal("Dispatch(/clear) should return non-nil cmd")
	}

	msg := cmd()
	_, ok := msg.(ClearChatMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ClearChatMsg", msg)
	}
}

func TestRegistry_DispatchByAlias(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("/cls", nil)
	if cmd == nil {
		t.Fatal("Dispatch(/cls) should return non-nil cmd (alias for clear)")
	}

	msg := cmd()
	_, ok := msg.(ClearChatMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ClearChatMsg", msg)
	}
}

func TestRegistry_DispatchWithColonPrefix(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch(":clear", nil)
	if cmd == nil {
		t.Fatal("Dispatch(:clear) should return non-nil cmd")
	}

	msg := cmd()
	_, ok := msg.(ClearChatMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ClearChatMsg", msg)
	}
}

func TestRegistry_DispatchUnknown(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("/nonexistent", nil)
	if cmd == nil {
		t.Fatal("Dispatch with unknown command should return non-nil cmd (error message)")
	}

	msg := cmd()
	sysMsg, ok := msg.(InjectSystemMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want InjectSystemMsg", msg)
	}
	if sysMsg.Content == "" {
		t.Error("error message should not be empty")
	}
}

func TestRegistry_DispatchEmpty(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("", nil)
	if cmd != nil {
		t.Error("Dispatch('') should return nil")
	}

	cmd = r.Dispatch("  ", nil)
	if cmd != nil {
		t.Error("Dispatch('  ') should return nil")
	}
}

func TestRegistry_DispatchCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("/CLEAR", nil)
	if cmd == nil {
		t.Fatal("Dispatch(/CLEAR) should work (case insensitive)")
	}

	msg := cmd()
	_, ok := msg.(ClearChatMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want ClearChatMsg", msg)
	}
}

func TestRegistry_Complete(t *testing.T) {
	r := NewRegistry()

	// Empty prefix returns all commands
	all := r.Complete("")
	if len(all) == 0 {
		t.Fatal("Complete('') should return all commands")
	}

	// Prefix matching
	matches := r.Complete("cl")
	found := false
	for _, m := range matches {
		if m == "clear" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Complete('cl') = %v, should contain 'clear'", matches)
	}

	// No matches
	none := r.Complete("zzz")
	if len(none) != 0 {
		t.Errorf("Complete('zzz') = %v, want empty", none)
	}
}

func TestRegistry_CompleteStripsPrefix(t *testing.T) {
	r := NewRegistry()

	// Should strip / prefix
	matches := r.Complete("/cl")
	found := false
	for _, m := range matches {
		if m == "clear" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Complete('/cl') = %v, should contain 'clear'", matches)
	}

	// Should strip : prefix
	matches = r.Complete(":cl")
	found = false
	for _, m := range matches {
		if m == "clear" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Complete(':cl') = %v, should contain 'clear'", matches)
	}
}

func TestRegistry_CompleteWithArgs_CommandName(t *testing.T) {
	r := NewRegistry()

	// Only command name, no space: complete the command name
	matches := r.CompleteWithArgs("cl", nil)
	found := false
	for _, m := range matches {
		if m == "clear" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CompleteWithArgs('cl') = %v, should contain 'clear'", matches)
	}
}

func TestRegistry_CompleteWithArgs_Empty(t *testing.T) {
	r := NewRegistry()

	matches := r.CompleteWithArgs("", nil)
	if len(matches) == 0 {
		t.Fatal("CompleteWithArgs('') should return all commands")
	}
}

func TestRegistry_ListSorted(t *testing.T) {
	r := NewRegistry()
	cmds := r.List()

	for i := 1; i < len(cmds); i++ {
		if cmds[i].Name() < cmds[i-1].Name() {
			t.Errorf("List() not sorted: %q comes after %q", cmds[i].Name(), cmds[i-1].Name())
			break
		}
	}
}

func TestRegistry_AliasCompletion(t *testing.T) {
	r := &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
	r.Register(&ClearCmd{}) // aliases: ["cls"]

	matches := r.Complete("cls")
	if len(matches) == 0 {
		t.Fatal("Complete('cls') should match the alias")
	}
}
