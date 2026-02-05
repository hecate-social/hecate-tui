package llmtools

import (
	"context"
	"encoding/json"
	"fmt"
)

// ApprovalRequest represents a request for user approval to execute a tool.
type ApprovalRequest struct {
	Tool     Tool
	Args     json.RawMessage
	ResultCh chan ApprovalResult
}

// ApprovalResult contains the user's decision on a tool approval request.
type ApprovalResult struct {
	Approved       bool
	GrantForSession bool // If true, grant permission for this tool for the session
}

// ApprovalHandler is called when a tool requires user approval.
// It should present the request to the user and return their decision.
type ApprovalHandler func(req ApprovalRequest) ApprovalResult

// Executor manages tool execution with permission checking.
type Executor struct {
	registry        *Registry
	permissions     *Permissions
	approvalHandler ApprovalHandler
}

// NewExecutor creates a new tool executor.
func NewExecutor(registry *Registry, permissions *Permissions) *Executor {
	return &Executor{
		registry:    registry,
		permissions: permissions,
	}
}

// SetApprovalHandler sets the callback for tool approval requests.
func (e *Executor) SetApprovalHandler(h ApprovalHandler) {
	e.approvalHandler = h
}

// Registry returns the underlying tool registry.
func (e *Executor) Registry() *Registry {
	return e.registry
}

// Permissions returns the underlying permissions.
func (e *Executor) Permissions() *Permissions {
	return e.permissions
}

// Execute runs a tool call, checking permissions first.
func (e *Executor) Execute(ctx context.Context, call ToolCall) ToolResult {
	tool, handler, ok := e.registry.Get(call.Name)
	if !ok {
		return ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
			IsError:    true,
		}
	}

	// Check base permission for the tool
	perm := e.permissions.Check(call.Name, call.Arguments)

	// For tools that require approval, override to Ask if not explicitly denied
	if tool.RequiresApproval && perm == PermissionAllow {
		// Session grants can still allow
		if !e.permissions.sessionGrants[call.Name] {
			perm = PermissionAsk
		}
	}

	// Additional path-based checks for filesystem tools
	if tool.Category == CategoryFileSystem {
		pathPerm := e.checkPathPermission(call.Arguments)
		if pathPerm == PermissionDeny {
			perm = PermissionDeny
		} else if pathPerm == PermissionAsk && perm == PermissionAllow {
			perm = PermissionAsk
		}
	}

	// Additional command checks for run_command
	if call.Name == "run_command" {
		cmdPerm := e.checkCommandPermission(call.Arguments)
		if cmdPerm == PermissionDeny {
			perm = PermissionDeny
		}
	}

	switch perm {
	case PermissionDeny:
		return ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Tool '%s' execution denied by policy", call.Name),
			IsError:    true,
		}

	case PermissionAsk:
		if e.approvalHandler == nil {
			return ToolResult{
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Tool '%s' requires approval but no approval handler is configured", call.Name),
				IsError:    true,
			}
		}

		// Request approval
		req := ApprovalRequest{
			Tool:     tool,
			Args:     call.Arguments,
			ResultCh: make(chan ApprovalResult, 1),
		}

		result := e.approvalHandler(req)

		if !result.Approved {
			return ToolResult{
				ToolCallID: call.ID,
				Content:    "Tool execution denied by user",
				IsError:    true,
			}
		}

		if result.GrantForSession {
			e.permissions.GrantForSession(call.Name)
		}
	}

	// Execute the tool
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

// ExecuteAll runs multiple tool calls and returns all results.
func (e *Executor) ExecuteAll(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, len(calls))
	for i, call := range calls {
		results[i] = e.Execute(ctx, call)
	}
	return results
}

// checkPathPermission extracts path from arguments and checks permission.
func (e *Executor) checkPathPermission(args json.RawMessage) PermissionLevel {
	var pathArgs struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &pathArgs); err != nil || pathArgs.Path == "" {
		return PermissionAllow // Can't extract path, let the tool handle it
	}

	return e.permissions.CheckPath(pathArgs.Path)
}

// checkCommandPermission extracts command from arguments and checks permission.
func (e *Executor) checkCommandPermission(args json.RawMessage) PermissionLevel {
	var cmdArgs struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(args, &cmdArgs); err != nil || cmdArgs.Command == "" {
		return PermissionDeny // Can't extract command
	}

	return e.permissions.CheckCommand(cmdArgs.Command)
}

// FormatToolCallDescription creates a human-readable description of a tool call.
func FormatToolCallDescription(tool Tool, args json.RawMessage) string {
	var argsMap map[string]any
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return fmt.Sprintf("%s (invalid arguments)", tool.Name)
	}

	switch tool.Name {
	case "read_file":
		if path, ok := argsMap["path"].(string); ok {
			return fmt.Sprintf("Read file: %s", path)
		}
	case "write_file":
		if path, ok := argsMap["path"].(string); ok {
			return fmt.Sprintf("Write file: %s", path)
		}
	case "edit_file":
		if path, ok := argsMap["path"].(string); ok {
			return fmt.Sprintf("Edit file: %s", path)
		}
	case "list_directory":
		if path, ok := argsMap["path"].(string); ok {
			return fmt.Sprintf("List directory: %s", path)
		}
	case "glob_search":
		if pattern, ok := argsMap["pattern"].(string); ok {
			return fmt.Sprintf("Search for files: %s", pattern)
		}
	case "grep_search":
		if pattern, ok := argsMap["pattern"].(string); ok {
			return fmt.Sprintf("Search for: %s", pattern)
		}
	case "run_command":
		if cmd, ok := argsMap["command"].(string); ok {
			if len(cmd) > 50 {
				cmd = cmd[:50] + "..."
			}
			return fmt.Sprintf("Run: %s", cmd)
		}
	case "ask_user":
		if q, ok := argsMap["question"].(string); ok {
			if len(q) > 50 {
				q = q[:50] + "..."
			}
			return fmt.Sprintf("Ask: %s", q)
		}
	}

	// Default: show tool name and argument keys
	keys := make([]string, 0, len(argsMap))
	for k := range argsMap {
		keys = append(keys, k)
	}
	return fmt.Sprintf("%s(%v)", tool.Name, keys)
}
