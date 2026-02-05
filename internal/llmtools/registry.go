package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Handler executes a tool and returns the result content.
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

// Registry manages available tools and their handlers.
type Registry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	handlers map[string]Handler
	order    []string // Preserve registration order
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]Handler),
	}
}

// Register adds a tool with its handler to the registry.
func (r *Registry) Register(tool Tool, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; !exists {
		r.order = append(r.order, tool.Name)
	}
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// Get retrieves a tool and its handler by name.
func (r *Registry) Get(name string) (Tool, Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return Tool{}, nil, false
	}
	handler := r.handlers[name]
	return tool, handler, true
}

// All returns all registered tools in registration order.
func (r *Registry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		if tool, ok := r.tools[name]; ok {
			result = append(result, tool)
		}
	}
	return result
}

// ByCategory returns tools grouped by category.
func (r *Registry) ByCategory() map[ToolCategory][]Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[ToolCategory][]Tool)
	for _, name := range r.order {
		if tool, ok := r.tools[name]; ok {
			result[tool.Category] = append(result[tool.Category], tool)
		}
	}
	return result
}

// Names returns all tool names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]string(nil), r.order...)
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// ToAnthropicSchema converts all tools to Anthropic's tool format.
func (r *Registry) ToAnthropicSchema() []map[string]any {
	tools := r.All()
	result := make([]map[string]any, len(tools))
	for i, tool := range tools {
		result[i] = map[string]any{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.Parameters,
		}
	}
	return result
}

// ToOpenAISchema converts all tools to OpenAI's function calling format.
func (r *Registry) ToOpenAISchema() []map[string]any {
	tools := r.All()
	result := make([]map[string]any, len(tools))
	for i, tool := range tools {
		result[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	return result
}

// ToJSON returns tools as a JSON array (generic format).
func (r *Registry) ToJSON() ([]byte, error) {
	return json.Marshal(r.All())
}

// Execute runs a tool by name with the given arguments.
func (r *Registry) Execute(ctx context.Context, call ToolCall) ToolResult {
	tool, handler, ok := r.Get(call.Name)
	if !ok {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
			IsError:    true,
		}
	}

	_ = tool // Could be used for logging, metrics, etc.

	content, err := handler(ctx, call.Arguments)
	if err != nil {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    err.Error(),
			IsError:    true,
		}
	}

	return ToolResult{
		ToolCallID: call.ID,
		Content:    content,
	}
}
