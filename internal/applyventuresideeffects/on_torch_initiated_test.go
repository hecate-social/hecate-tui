package applyventuresideeffects

import (
	"encoding/json"
	"testing"
)

func TestParsesValidJSON(t *testing.T) {
	data := json.RawMessage(`{
		"venture_id": "venture-abc-123",
		"name": "My Venture",
		"brief": "A test venture",
		"initiated_by": "test-user@localhost",
		"status": 1,
		"initiated_at": 1700000000000
	}`)

	msg, err := HandleVentureInitiated(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.VentureID != "venture-abc-123" {
		t.Errorf("Expected VentureID 'venture-abc-123', got '%s'", msg.VentureID)
	}
	if msg.Name != "My Venture" {
		t.Errorf("Expected Name 'My Venture', got '%s'", msg.Name)
	}
}

func TestAllFieldsMapped(t *testing.T) {
	data := json.RawMessage(`{
		"venture_id": "venture-full",
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

	if msg.VentureID != "venture-full" {
		t.Errorf("Expected VentureID 'venture-full', got '%s'", msg.VentureID)
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
	data := json.RawMessage(`{"venture_id": "minimal", "name": "Minimal Venture"}`)

	msg, err := HandleVentureInitiated(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.VentureID != "minimal" {
		t.Errorf("Expected VentureID 'minimal', got '%s'", msg.VentureID)
	}
	if msg.Name != "Minimal Venture" {
		t.Errorf("Expected Name 'Minimal Venture', got '%s'", msg.Name)
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
