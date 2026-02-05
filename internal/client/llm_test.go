package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/llm/providers" {
			t.Errorf("Expected path '/api/llm/providers', got '%s'", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", r.Method)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"providers": {
					"ollama": {"type": "ollama", "url": "http://localhost:11434", "enabled": true},
					"anthropic": {"type": "anthropic", "url": "https://api.anthropic.com", "enabled": true}
				}
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	providers, err := c.ListProviders()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(providers) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(providers))
	}

	ollama, ok := providers["ollama"]
	if !ok {
		t.Fatal("Expected 'ollama' provider to exist")
	}
	if ollama.Type != "ollama" {
		t.Errorf("Expected ollama type 'ollama', got '%s'", ollama.Type)
	}
	if !ollama.Enabled {
		t.Error("Expected ollama to be enabled")
	}
	if ollama.URL != "http://localhost:11434" {
		t.Errorf("Expected ollama URL 'http://localhost:11434', got '%s'", ollama.URL)
	}

	anthropic, ok := providers["anthropic"]
	if !ok {
		t.Fatal("Expected 'anthropic' provider to exist")
	}
	if anthropic.Type != "anthropic" {
		t.Errorf("Expected anthropic type 'anthropic', got '%s'", anthropic.Type)
	}
	if !anthropic.Enabled {
		t.Error("Expected anthropic to be enabled")
	}
}

func TestListProvidersError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Ok:    false,
			Error: "internal error",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.ListProviders()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestAddProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/llm/providers/add" {
			t.Errorf("Expected path '/api/llm/providers/add', got '%s'", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		var reqBody map[string]string
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		if reqBody["name"] != "anthropic" {
			t.Errorf("Expected name 'anthropic', got '%s'", reqBody["name"])
		}
		if reqBody["type"] != "anthropic" {
			t.Errorf("Expected type 'anthropic', got '%s'", reqBody["type"])
		}
		if reqBody["api_key"] != "sk-test" {
			t.Errorf("Expected api_key 'sk-test', got '%s'", reqBody["api_key"])
		}

		resp := Response{
			Ok:     true,
			Result: json.RawMessage(`{}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.AddProvider("anthropic", "anthropic", "sk-test", "https://api.anthropic.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAddProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Ok:    false,
			Error: "invalid type",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.AddProvider("bad", "invalid", "", "")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestRemoveProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/llm/providers/anthropic/remove" {
			t.Errorf("Expected path '/api/llm/providers/anthropic/remove', got '%s'", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", r.Method)
		}

		resp := Response{
			Ok:     true,
			Result: json.RawMessage(`{}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.RemoveProvider("anthropic")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRemoveProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Ok:    false,
			Error: "cannot remove built-in provider: ollama",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.RemoveProvider("ollama")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/llm/models" {
			t.Errorf("Expected path '/api/llm/models', got '%s'", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", r.Method)
		}

		resp := Response{
			Ok: true,
			Result: json.RawMessage(`{
				"models": [
					{"name": "llama3.2", "family": "llama", "parameter_size": "3B", "context_length": 4096, "provider": "ollama"},
					{"name": "claude-sonnet-4-5-20250929", "family": "claude", "context_length": 200000, "provider": "anthropic"}
				]
			}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	models, err := c.ListModels()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	if models[0].Name != "llama3.2" {
		t.Errorf("Expected first model name 'llama3.2', got '%s'", models[0].Name)
	}
	if models[0].Family != "llama" {
		t.Errorf("Expected first model family 'llama', got '%s'", models[0].Family)
	}
	if models[0].ParameterSize != "3B" {
		t.Errorf("Expected first model parameter_size '3B', got '%s'", models[0].ParameterSize)
	}
	if models[0].ContextLength != 4096 {
		t.Errorf("Expected first model context_length 4096, got %d", models[0].ContextLength)
	}
	if models[0].Provider != "ollama" {
		t.Errorf("Expected first model provider 'ollama', got '%s'", models[0].Provider)
	}

	if models[1].Name != "claude-sonnet-4-5-20250929" {
		t.Errorf("Expected second model name 'claude-sonnet-4-5-20250929', got '%s'", models[1].Name)
	}
	if models[1].Family != "claude" {
		t.Errorf("Expected second model family 'claude', got '%s'", models[1].Family)
	}
	if models[1].ContextLength != 200000 {
		t.Errorf("Expected second model context_length 200000, got %d", models[1].ContextLength)
	}
	if models[1].Provider != "anthropic" {
		t.Errorf("Expected second model provider 'anthropic', got '%s'", models[1].Provider)
	}
}
