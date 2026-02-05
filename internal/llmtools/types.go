// Package llmtools provides a tool/function calling system for LLM agents.
// This enables AI to interact with the filesystem, search the web, explore codebases,
// and integrate with external services.
package llmtools

import "encoding/json"

// ToolCategory groups related tools by capability domain.
type ToolCategory string

const (
	CategoryFileSystem  ToolCategory = "filesystem"
	CategoryWeb         ToolCategory = "web"
	CategoryMesh        ToolCategory = "mesh"
	CategoryCodeExplore ToolCategory = "code_explore"
	CategorySystem      ToolCategory = "system"
)

// CategoryName returns a human-readable name for the category.
func CategoryName(cat ToolCategory) string {
	switch cat {
	case CategoryFileSystem:
		return "File System"
	case CategoryWeb:
		return "Web"
	case CategoryMesh:
		return "Mesh"
	case CategoryCodeExplore:
		return "Code Exploration"
	case CategorySystem:
		return "System"
	default:
		return string(cat)
	}
}

// Tool represents an available tool the LLM can invoke.
type Tool struct {
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Parameters       ToolParameters `json:"parameters"`
	Category         ToolCategory   `json:"category"`
	RequiresApproval bool           `json:"requires_approval"`
}

// ToolParameters describes the JSON Schema for tool arguments.
type ToolParameters struct {
	Type       string                   `json:"type"` // Always "object"
	Properties map[string]ParameterSpec `json:"properties"`
	Required   []string                 `json:"required,omitempty"`
}

// ParameterSpec describes a single parameter in the JSON Schema.
type ParameterSpec struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// ToolCall represents an LLM's request to use a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the output of executing a tool.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// NewObjectParameters creates a ToolParameters with type "object".
func NewObjectParameters() ToolParameters {
	return ToolParameters{
		Type:       "object",
		Properties: make(map[string]ParameterSpec),
	}
}

// AddProperty adds a parameter to the schema.
func (p *ToolParameters) AddProperty(name string, spec ParameterSpec) *ToolParameters {
	p.Properties[name] = spec
	return p
}

// AddRequired marks a parameter as required.
func (p *ToolParameters) AddRequired(names ...string) *ToolParameters {
	p.Required = append(p.Required, names...)
	return p
}

// String creates a string parameter spec.
func String(description string) ParameterSpec {
	return ParameterSpec{Type: "string", Description: description}
}

// Integer creates an integer parameter spec.
func Integer(description string) ParameterSpec {
	return ParameterSpec{Type: "integer", Description: description}
}

// Boolean creates a boolean parameter spec.
func Boolean(description string) ParameterSpec {
	return ParameterSpec{Type: "boolean", Description: description}
}

// Enum creates a string parameter with allowed values.
func Enum(description string, values ...string) ParameterSpec {
	return ParameterSpec{Type: "string", Description: description, Enum: values}
}

// WithDefault adds a default value to a parameter spec.
func (ps ParameterSpec) WithDefault(val any) ParameterSpec {
	ps.Default = val
	return ps
}
