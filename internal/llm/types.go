package llm

import "encoding/json"

// Role represents a message role
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool" // Tool result message
)

// Message represents a chat message.
// For assistant messages with tool calls, Content may be empty and ToolCalls populated.
// For tool result messages, ToolCallID identifies which call this result is for.
type Message struct {
	Role       Role        `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`   // Assistant requesting tools
	ToolCallID string      `json:"tool_call_id,omitempty"` // Tool result reference
}

// ToolCall represents an LLM's request to invoke a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the output of a tool execution.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// ContentBlock represents a block in multi-part content (Anthropic format).
type ContentBlock struct {
	Type    string          `json:"type"` // "text", "tool_use", "tool_result"
	Text    string          `json:"text,omitempty"`
	ID      string          `json:"id,omitempty"`      // tool_use
	Name    string          `json:"name,omitempty"`    // tool_use
	Input   json.RawMessage `json:"input,omitempty"`   // tool_use
	Content string          `json:"content,omitempty"` // tool_result
	IsError bool            `json:"is_error,omitempty"`
}

// ToolSchema represents a tool definition sent to the LLM.
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"` // JSON Schema
}

// Model represents an available LLM model
type Model struct {
	Name          string `json:"name"`
	Size          string `json:"size,omitempty"`
	ModifiedAt    string `json:"modified_at,omitempty"`
	Digest        string `json:"digest,omitempty"`
	ParameterSize string `json:"parameter_size,omitempty"`
	Family        string `json:"family,omitempty"`
	Format        string `json:"format,omitempty"`
	ContextLength int    `json:"context_length,omitempty"`
	Quantization  string `json:"quantization_level,omitempty"`
	Provider      string `json:"provider,omitempty"`
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model       string       `json:"model"`
	Messages    []Message    `json:"messages"`
	Stream      bool         `json:"stream"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	Tools       []ToolSchema `json:"tools,omitempty"` // Available tools for function calling
}

// ChatResponse represents a chat completion response chunk.
type ChatResponse struct {
	Model   string   `json:"model,omitempty"`
	Message *Message `json:"message,omitempty"`
	Done    bool     `json:"done"`

	// Tool use events (streaming)
	ToolUse *ToolCall `json:"tool_use,omitempty"` // When LLM wants to call a tool

	// Stop reason (when done=true)
	StopReason string `json:"stop_reason,omitempty"` // "end_turn", "tool_use", etc.

	// Usage stats (only present when done=true)
	TotalDuration   int64 `json:"total_duration,omitempty"`
	LoadDuration    int64 `json:"load_duration,omitempty"`
	PromptEvalCount int   `json:"prompt_eval_count,omitempty"`
	EvalCount       int   `json:"eval_count,omitempty"`
	EvalDuration    int64 `json:"eval_duration,omitempty"`
}

// LLMHealth represents the LLM backend health status
type LLMHealth struct {
	Status    string            `json:"status"`
	Backend   string            `json:"backend,omitempty"`
	URL       string            `json:"url,omitempty"`
	Error     string            `json:"error,omitempty"`
	Providers map[string]string `json:"providers,omitempty"`
}

// Provider represents a configured LLM provider
type Provider struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type"`
	URL     string `json:"url,omitempty"`
	Enabled bool   `json:"enabled"`
}

// ProvidersResponse represents the list providers response
type ProvidersResponse struct {
	Providers map[string]Provider `json:"providers"`
}

// ModelsResponse represents the list models response
type ModelsResponse struct {
	Models []Model `json:"models"`
}
