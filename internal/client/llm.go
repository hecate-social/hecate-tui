package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hecate-social/hecate-tui/internal/llm"
)

// ListModels returns available LLM models
func (c *Client) ListModels() ([]llm.Model, error) {
	resp, err := c.get("/api/llm/models")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list models failed: %s", resp.Error)
	}

	var result llm.ModelsResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	return result.Models, nil
}

// GetLLMHealth checks LLM backend health
func (c *Client) GetLLMHealth() (*llm.LLMHealth, error) {
	resp, err := c.get("/api/llm/health")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("LLM health check failed: %s", resp.Error)
	}

	var health llm.LLMHealth
	if err := json.Unmarshal(resp.Result, &health); err != nil {
		return nil, fmt.Errorf("failed to parse LLM health response: %w", err)
	}

	return &health, nil
}

// ChatStream sends a chat request and returns a channel of streaming responses
func (c *Client) ChatStream(ctx context.Context, req llm.ChatRequest) (<-chan llm.ChatResponse, <-chan error) {
	respChan := make(chan llm.ChatResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		// Ensure streaming is enabled
		req.Stream = true

		jsonBody, err := json.Marshal(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/llm/chat", bytes.NewReader(jsonBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")

		// Use a client without timeout for streaming, but reuse socket transport
		streamClient := &http.Client{}
		if c.transport != nil {
			streamClient.Transport = c.transport
		}
		httpResp, err := streamClient.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("request failed: %w", err)
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(httpResp.Body)
			errChan <- fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, string(body))
			return
		}

		parser := llm.NewStreamParser(httpResp.Body)
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				resp, err := parser.Next()
				if err == io.EOF {
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				select {
				case respChan <- *resp:
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}

				if resp.Done {
					return
				}
			}
		}
	}()

	return respChan, errChan
}

// ListProviders returns configured LLM providers
func (c *Client) ListProviders() (map[string]llm.Provider, error) {
	resp, err := c.get("/api/llm/providers")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list providers failed: %s", resp.Error)
	}

	var result llm.ProvidersResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse providers response: %w", err)
	}

	return result.Providers, nil
}

// AddProvider adds a new LLM provider configuration
func (c *Client) AddProvider(name, pType, apiKey, url string) error {
	body := map[string]string{
		"name": name,
		"type": pType,
	}
	if apiKey != "" {
		body["api_key"] = apiKey
	}
	if url != "" {
		body["url"] = url
	}

	resp, err := c.post("/api/llm/providers/add", body)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("add provider failed: %s", resp.Error)
	}

	return nil
}

// RemoveProvider removes an LLM provider by name
func (c *Client) RemoveProvider(name string) error {
	resp, err := c.post("/api/llm/providers/"+name+"/remove", nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("remove provider failed: %s", resp.Error)
	}

	return nil
}

// Chat sends a non-streaming chat request
func (c *Client) Chat(req llm.ChatRequest) (*llm.ChatResponse, error) {
	req.Stream = false

	resp, err := c.post("/api/llm/chat", req)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("chat failed: %s", resp.Error)
	}

	var chatResp llm.ChatResponse
	if err := json.Unmarshal(resp.Result, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}

	return &chatResp, nil
}
