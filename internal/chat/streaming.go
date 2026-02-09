package chat

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llm"
)

type streamState struct {
	ctx         context.Context
	cancel      context.CancelFunc
	respChan    <-chan llm.ChatResponse
	errChan     <-chan error
	start       time.Time
	totalTokens int
}

var activeStream *streamState

func (m *Model) sendMessage() tea.Cmd {
	return m.sendMessageWithToolResults(nil)
}

func (m *Model) sendMessageWithToolResults(toolResults []llm.ToolResult) tea.Cmd {
	return func() tea.Msg {
		if len(m.models) == 0 {
			return streamErrorMsg{err: fmt.Errorf("no models available")}
		}

		modelName := m.models[m.activeModel].Name
		ctx, cancel := context.WithCancel(context.Background())

		// Convert our messages to llm.Message
		var llmMsgs []llm.Message

		// Prepend system prompt if set
		if m.systemPrompt != "" {
			llmMsgs = append(llmMsgs, llm.Message{
				Role:    llm.RoleSystem,
				Content: m.systemPrompt,
			})
		}

		for _, msg := range m.messages {
			if msg.Role == "system" {
				continue // Don't send system messages to LLM
			}
			llmMsgs = append(llmMsgs, llm.Message{
				Role:    llm.Role(msg.Role),
				Content: msg.Content,
			})
		}

		// Add tool results if any
		for _, result := range toolResults {
			llmMsgs = append(llmMsgs, llm.Message{
				Role:       llm.RoleTool,
				Content:    result.Content,
				ToolCallID: result.ToolCallID,
			})
		}

		req := llm.ChatRequest{
			Model:    modelName,
			Messages: llmMsgs,
			Stream:   true,
		}

		// Add tool schemas if tools are enabled
		if m.toolsEnabled && m.toolExecutor != nil {
			req.Tools = m.buildToolSchemas()
		}

		start := time.Now()
		respChan, errChan := m.client.ChatStream(ctx, req)

		activeStream = &streamState{
			ctx:      ctx,
			cancel:   cancel,
			respChan: respChan,
			errChan:  errChan,
			start:    start,
		}

		return pollStreamCmd()
	}
}

// buildToolSchemas converts the tool registry to LLM tool schemas.
func (m *Model) buildToolSchemas() []llm.ToolSchema {
	if m.toolExecutor == nil {
		return nil
	}

	registry := m.toolExecutor.Registry()
	tools := registry.All()
	schemas := make([]llm.ToolSchema, len(tools))

	for i, tool := range tools {
		// Convert ToolParameters to map[string]any for JSON schema
		inputSchema := map[string]any{
			"type":       tool.Parameters.Type,
			"properties": tool.Parameters.Properties,
		}
		if len(tool.Parameters.Required) > 0 {
			inputSchema["required"] = tool.Parameters.Required
		}

		schemas[i] = llm.ToolSchema{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: inputSchema,
		}
	}

	return schemas
}

func pollStreamCmd() tea.Msg {
	if activeStream == nil {
		return streamDoneMsg{totalTokens: 0, duration: 0, reason: "activeStream was nil"}
	}

	select {
	case resp, ok := <-activeStream.respChan:
		if !ok {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "channel closed"}
		}
		if resp.EvalCount > 0 {
			activeStream.totalTokens = resp.EvalCount
		}

		// Check for tool use in the response
		if resp.ToolUse != nil {
			// Complete tool call received
			return toolUseCompleteMsg{call: *resp.ToolUse}
		}

		// Check for tool calls - Anthropic uses stop_reason="tool_use",
		// but Ollama uses "stop" with tool_calls present
		if resp.Done && resp.Message != nil && len(resp.Message.ToolCalls) > 0 {
			// Return the first tool call (we'll handle multiple later)
			return toolUseCompleteMsg{call: resp.Message.ToolCalls[0]}
		}

		if resp.Done {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "resp.Done=true"}
		}
		return streamChunkMsg{chunk: resp}

	case err, ok := <-activeStream.errChan:
		if !ok {
			// errChan closed without error - stream ended normally, keep polling respChan
			return continueStreamMsg{}
		}
		duration := time.Since(activeStream.start)
		tokens := activeStream.totalTokens
		activeStream = nil
		if err != nil && err != context.Canceled {
			return streamErrorMsg{err: err}
		}
		return streamDoneMsg{totalTokens: tokens, duration: duration, reason: fmt.Sprintf("errChan: %v", err)}

	default:
		return continueStreamMsg{}
	}
}

func (m Model) thinkingTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}

func (m Model) fetchModels() tea.Msg {
	models, err := m.client.ListModels()
	return modelsMsg{models: models, err: err}
}
