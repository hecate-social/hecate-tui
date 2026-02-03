package llm

// Role represents a message role
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a chat message
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Model represents an available LLM model
type Model struct {
	Name          string  `json:"name"`
	Size          string  `json:"size,omitempty"`
	ModifiedAt    string  `json:"modified_at,omitempty"`
	Digest        string  `json:"digest,omitempty"`
	ParameterSize string  `json:"parameter_size,omitempty"`
	Family        string  `json:"family,omitempty"`
	Format        string  `json:"format,omitempty"`
	ContextLength int     `json:"context_length,omitempty"`
	Quantization  string  `json:"quantization_level,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// ChatResponse represents a chat completion response chunk
type ChatResponse struct {
	Model     string `json:"model,omitempty"`
	Message   *Message `json:"message,omitempty"`
	Done      bool   `json:"done"`

	// Usage stats (only present when done=true)
	TotalDuration    int64 `json:"total_duration,omitempty"`
	LoadDuration     int64 `json:"load_duration,omitempty"`
	PromptEvalCount  int   `json:"prompt_eval_count,omitempty"`
	EvalCount        int   `json:"eval_count,omitempty"`
	EvalDuration     int64 `json:"eval_duration,omitempty"`
}

// LLMHealth represents the LLM backend health status
type LLMHealth struct {
	Status  string `json:"status"`
	Backend string `json:"backend"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ModelsResponse represents the list models response
type ModelsResponse struct {
	Models []Model `json:"models"`
}
