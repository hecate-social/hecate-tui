package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	c := New("http://localhost:4444")
	if c == nil {
		t.Fatal("Expected non-nil client")
	}
	if c.baseURL != "http://localhost:4444" {
		t.Errorf("Expected baseURL 'http://localhost:4444', got '%s'", c.baseURL)
	}
}

func TestGetHealth(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("Expected path '/health', got '%s'", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", r.Method)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"status": "healthy",
				"uptime_seconds": 3600,
				"version": "0.1.0"
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	health, err := c.GetHealth()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}
	if health.UptimeSeconds != 3600 {
		t.Errorf("Expected uptime 3600, got %d", health.UptimeSeconds)
	}
	if health.Version != "0.1.0" {
		t.Errorf("Expected version '0.1.0', got '%s'", health.Version)
	}
}

func TestGetHealthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Ok:    false,
			Error: "daemon_unavailable",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.GetHealth()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity" {
			t.Errorf("Expected path '/identity', got '%s'", r.URL.Path)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"identity": "mri:agent:io.macula/test-agent",
				"public_key": "base64key",
				"created_at": "2026-02-01T12:00:00Z"
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	identity, err := c.GetIdentity()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if identity.Identity != "mri:agent:io.macula/test-agent" {
		t.Errorf("Expected identity 'mri:agent:io.macula/test-agent', got '%s'", identity.Identity)
	}
}

func TestDiscoverCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/capabilities/discover" {
			t.Errorf("Expected path '/capabilities/discover', got '%s'", r.URL.Path)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"capabilities": [
					{
						"mri": "mri:capability:io.macula/weather",
						"agent_identity": "mri:agent:io.macula/weather-service",
						"tags": ["weather", "forecast"],
						"description": "Weather service",
						"announced_at": "2026-02-01T12:00:00Z"
					}
				]
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	caps, err := c.DiscoverCapabilities("", "", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(caps) != 1 {
		t.Fatalf("Expected 1 capability, got %d", len(caps))
	}
	if caps[0].MRI != "mri:capability:io.macula/weather" {
		t.Errorf("Expected MRI 'mri:capability:io.macula/weather', got '%s'", caps[0].MRI)
	}
	if len(caps[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(caps[0].Tags))
	}
}

func TestListProcedures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc/procedures" {
			t.Errorf("Expected path '/rpc/procedures', got '%s'", r.URL.Path)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"procedures": [
					{
						"name": "echo",
						"mri": "mri:rpc:io.macula/echo",
						"endpoint": "http://localhost:8080/echo",
						"registered_at": "2026-02-01T12:00:00Z"
					}
				]
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	procs, err := c.ListProcedures()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(procs) != 1 {
		t.Fatalf("Expected 1 procedure, got %d", len(procs))
	}
	if procs[0].Name != "echo" {
		t.Errorf("Expected name 'echo', got '%s'", procs[0].Name)
	}
}

func TestListSubscriptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions" {
			t.Errorf("Expected path '/subscriptions', got '%s'", r.URL.Path)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"subscriptions": [
					{
						"subscription_id": "sub-123",
						"service_mri": "mri:service:io.macula/updates",
						"subscribed_at": "2026-02-01T12:00:00Z"
					}
				]
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	subs, err := c.ListSubscriptions()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(subs) != 1 {
		t.Fatalf("Expected 1 subscription, got %d", len(subs))
	}
	if subs[0].ServiceMRI != "mri:service:io.macula/updates" {
		t.Errorf("Expected service MRI 'mri:service:io.macula/updates', got '%s'", subs[0].ServiceMRI)
	}
}

func TestConnectionError(t *testing.T) {
	c := New("http://localhost:99999") // Invalid port
	_, err := c.GetHealth()
	if err == nil {
		t.Fatal("Expected connection error, got nil")
	}
}
