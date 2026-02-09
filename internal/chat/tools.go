package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llm"
	"github.com/hecate-social/hecate-tui/internal/llmtools"
)

// handleToolUseComplete processes a completed tool use request from the LLM.
func (m *Model) handleToolUseComplete(call llm.ToolCall) tea.Cmd {
	if m.toolExecutor == nil {
		// No executor, return error
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    "Tool execution not available",
					IsError:    true,
				},
			}
		}
	}

	registry := m.toolExecutor.Registry()
	tool, _, ok := registry.Get(call.Name)
	if !ok {
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
					IsError:    true,
				},
			}
		}
	}

	// Check permissions
	permissions := m.toolExecutor.Permissions()
	perm := permissions.Check(call.Name, call.Arguments)

	// For tools that require approval, always ask unless session-granted
	if tool.RequiresApproval && perm == llmtools.PermissionAllow {
		if !permissions.SessionGranted(call.Name) {
			perm = llmtools.PermissionAsk
		}
	}

	switch perm {
	case llmtools.PermissionDeny:
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Tool '%s' execution denied by policy", call.Name),
					IsError:    true,
				},
			}
		}

	case llmtools.PermissionAsk:
		// Store pending call and request approval
		m.pendingToolCall = &call
		return func() tea.Msg {
			return toolApprovalRequestMsg{tool: tool, call: call}
		}

	default: // PermissionAllow
		return m.executeToolCall(call)
	}
}

// executeToolCall runs a tool and returns the result message.
func (m *Model) executeToolCall(call llm.ToolCall) tea.Cmd {
	m.executingTool = true

	// Show that we're executing the tool
	m.showToolExecution(call)

	return func() tea.Msg {
		if m.toolExecutor == nil {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: call.ID,
					Content:    "Tool executor not configured",
					IsError:    true,
				},
			}
		}

		// Convert llm.ToolCall to llmtools.ToolCall
		toolCall := llmtools.ToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		}

		// Execute the tool
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result := m.toolExecutor.Registry().Execute(ctx, toolCall)

		return toolExecutionResultMsg{
			result: llm.ToolResult{
				ToolCallID: result.ToolCallID,
				Content:    result.Content,
				IsError:    result.IsError,
			},
		}
	}
}

// handleApprovalResponse processes the user's approval decision.
func (m *Model) handleApprovalResponse(msg toolApprovalResponseMsg) tea.Cmd {
	m.pendingToolCall = nil

	if !msg.approved {
		return func() tea.Msg {
			return toolExecutionResultMsg{
				result: llm.ToolResult{
					ToolCallID: msg.call.ID,
					Content:    "Tool execution denied by user",
					IsError:    true,
				},
			}
		}
	}

	// Grant session permission if requested
	if msg.grantForSession && m.toolExecutor != nil {
		m.toolExecutor.Permissions().GrantForSession(msg.call.Name)
	}

	return m.executeToolCall(msg.call)
}

// continueWithToolResults sends tool results back to the LLM to continue.
func (m *Model) continueWithToolResults() tea.Cmd {
	if len(m.toolResults) == 0 {
		return nil
	}

	results := m.toolResults
	m.toolResults = nil

	m.streaming = true
	m.streamBuf.Reset()
	m.streamStart = time.Now()

	return tea.Batch(
		m.sendMessageWithToolResults(results),
		m.thinkingTick(),
	)
}

// showToolExecution displays that a tool is being executed.
func (m *Model) showToolExecution(call llm.ToolCall) {
	var argsPreview string
	if len(call.Arguments) > 0 {
		var args map[string]any
		if err := json.Unmarshal(call.Arguments, &args); err == nil {
			// Show a brief preview of arguments
			parts := make([]string, 0, len(args))
			for k, v := range args {
				vs := fmt.Sprintf("%v", v)
				if len(vs) > 30 {
					vs = vs[:27] + "..."
				}
				parts = append(parts, fmt.Sprintf("%s=%s", k, vs))
			}
			argsPreview = strings.Join(parts, ", ")
		}
	}

	content := fmt.Sprintf("⚙️ Executing: %s", call.Name)
	if argsPreview != "" {
		content += fmt.Sprintf("\n   Args: %s", argsPreview)
	}

	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: content,
		Time:    time.Now(),
	})
	m.updateViewport()
}

// showToolResult displays the result of a tool execution.
func (m *Model) showToolResult(result llm.ToolResult) {
	status := "✓"
	if result.IsError {
		status = "✗"
	}

	// Truncate long results for display
	content := result.Content
	if len(content) > 500 {
		content = content[:500] + "\n... (truncated)"
	}

	msg := fmt.Sprintf("%s Tool result:\n%s", status, content)

	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: msg,
		Time:    time.Now(),
	})
	m.updateViewport()
}

// ApproveToolCall approves the pending tool call.
func (m *Model) ApproveToolCall(grantForSession bool) tea.Cmd {
	if m.pendingToolCall == nil {
		return nil
	}

	call := *m.pendingToolCall
	return func() tea.Msg {
		return toolApprovalResponseMsg{
			approved:        true,
			grantForSession: grantForSession,
			call:            call,
		}
	}
}

// DenyToolCall denies the pending tool call.
func (m *Model) DenyToolCall() tea.Cmd {
	if m.pendingToolCall == nil {
		return nil
	}

	call := *m.pendingToolCall
	return func() tea.Msg {
		return toolApprovalResponseMsg{
			approved: false,
			call:     call,
		}
	}
}

// ContinueAfterToolResult signals to continue the conversation after a tool result.
func (m *Model) ContinueAfterToolResult() tea.Cmd {
	return func() tea.Msg {
		return toolContinueMsg{}
	}
}
