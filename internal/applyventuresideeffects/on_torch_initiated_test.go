package applyventuresideeffects

import (
	"encoding/json"
	"testing"
)

func TestParsesValidJSON(t *testing.T) {
	data := json.RawMessage(`{
		"torch_id": "torch-abc-123",
		"name": "My Torch",
		"brief": "A test torch",
		"initiated_by": "test-user@localhost",
		"status": 1,
		"initiated_at": 1700000000000
	}`)

	msg, err := HandleVentureInitiated(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.VentureID != "torch-abc-123" {
		t.Errorf("Expected VentureID 'torch-abc-123', got '%s'", msg.VentureID)
	}
	if msg.Name != "My Torch" {
		t.Errorf("Expected Name 'My Torch', got '%s'", msg.Name)
	}
}

func TestAllFieldsMapped(t *testing.T) {
	data := json.RawMessage(`{
		"torch_id": "torch-full",
		"name": "Full Fields",
		"brief": "Everything filled",
		"initiated_by": "admin@host",
		"status": 3,
		"initiated_at": 1700000000999
	}`)

	msg, err := HandleVentureInitiated(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.VentureID != "torch-full" {
		t.Errorf("Expected VentureID 'torch-full', got '%s'", msg.VentureID)
	}
	if msg.Name != "Full Fields" {
		t.Errorf("Expected Name 'Full Fields', got '%s'", msg.Name)
	}
	if msg.Brief != "Everything filled" {
		t.Errorf("Expected Brief 'Everything filled', got '%s'", msg.Brief)
	}
	if msg.InitiatedBy != "admin@host" {
		t.Errorf("Expected InitiatedBy 'admin@host', got '%s'", msg.InitiatedBy)
	}
	if msg.Status != 3 {
		t.Errorf("Expected Status 3, got %d", msg.Status)
	}
	if msg.InitiatedAt != 1700000000999 {
		t.Errorf("Expected InitiatedAt 1700000000999, got %d", msg.InitiatedAt)
	}
}

func TestHandlesMissingOptionalFields(t *testing.T) {
	// Only required fields — optional ones should get zero values
	data := json.RawMessage(`{"torch_id": "minimal", "name": "Minimal Torch"}`)

	msg, err := HandleVentureInitiated(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.VentureID != "minimal" {
		t.Errorf("Expected VentureID 'minimal', got '%s'", msg.VentureID)
	}
	if msg.Name != "Minimal Torch" {
		t.Errorf("Expected Name 'Minimal Torch', got '%s'", msg.Name)
	}
	if msg.Brief != "" {
		t.Errorf("Expected empty Brief, got '%s'", msg.Brief)
	}
	if msg.InitiatedBy != "" {
		t.Errorf("Expected empty InitiatedBy, got '%s'", msg.InitiatedBy)
	}
	if msg.Status != 0 {
		t.Errorf("Expected Status 0, got %d", msg.Status)
	}
	if msg.InitiatedAt != 0 {
		t.Errorf("Expected InitiatedAt 0, got %d", msg.InitiatedAt)
	}
}

func TestReturnsErrorForInvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)

	_, err := HandleVentureInitiated(data)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestReturnsErrorForNil(t *testing.T) {
	_, err := HandleVentureInitiated(nil)
	if err == nil {
		t.Fatal("Expected error for nil input, got nil")
	}
}
