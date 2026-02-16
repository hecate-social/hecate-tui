package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hecate-social/hecate-tui/internal/llm"
)

var debugLog *os.File

func init() {
	debugLog, _ = os.OpenFile("/tmp/hecate-tui-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func debugf(format string, args ...any) {
	if debugLog != nil {
		fmt.Fprintf(debugLog, format+"\n", args...)
		debugLog.Sync()
	}
}

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
		debugf("sendMessageWithToolResults: toolResults=%d messages=%d", len(toolResults), len(m.messages))
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
			lm := llm.Message{
				Role:    llm.Role(msg.Role),
				Content: msg.Content,
			}
			if len(msg.ToolCalls) > 0 {
				lm.ToolCalls = msg.ToolCalls
			}
			llmMsgs = append(llmMsgs, lm)
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
		debugf("pollStreamCmd: activeStream is nil")
		// Return no-op instead of streamDoneMsg so stale poll ticks during
		// tool execution don't kill the streaming state.
		return continueStreamMsg{}
	}

	select {
	case resp, ok := <-activeStream.respChan:
		if !ok {
			debugf("pollStreamCmd: respChan closed")
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			// Check errChan for a buffered error before reporting "channel closed".
			// This fixes a race where Go's select picks respChan closure over errChan.
			select {
			case err, eOk := <-activeStream.errChan:
				if eOk && err != nil && err != context.Canceled {
					activeStream = nil
					return streamErrorMsg{err: err}
				}
			default:
			}
			activeStream = nil
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "stream completed"}
		}
		// Debug: dump the raw response
		raw, _ := json.Marshal(resp)
		debugf("pollStreamCmd: got chunk: %s", string(raw))
		debugf("pollStreamCmd: Message=%v ToolUse=%v Done=%v", resp.Message != nil, resp.ToolUse != nil, resp.Done)
		if resp.Message != nil {
			debugf("pollStreamCmd: Message.ToolCalls=%d Content=%q", len(resp.Message.ToolCalls), resp.Message.Content)
		}

		if resp.EvalCount > 0 {
			activeStream.totalTokens = resp.EvalCount
		}

		// Check for tool use in the response (Anthropic streaming format)
		if resp.ToolUse != nil {
			debugf("pollStreamCmd: ToolUse detected (Anthropic)")
			// Clear activeStream so stale poll ticks don't read remaining chunks
			activeStream = nil
			return toolUseCompleteMsg{call: *resp.ToolUse}
		}

		// Check for tool calls in message (Ollama/OpenAI format).
		// Ollama sends tool_calls in a done:false chunk, so check regardless of Done.
		if resp.Message != nil && len(resp.Message.ToolCalls) > 0 {
			debugf("pollStreamCmd: ToolCalls detected in Message: %+v", resp.Message.ToolCalls[0])
			// Clear activeStream so stale poll ticks don't read remaining chunks
			activeStream = nil
			return toolUseCompleteMsg{call: resp.Message.ToolCalls[0]}
		}

		if resp.Done {
			duration := time.Since(activeStream.start)
			tokens := activeStream.totalTokens
			activeStream = nil
			debugf("pollStreamCmd: Done=true, tokens=%d duration=%v", tokens, duration)
			return streamDoneMsg{totalTokens: tokens, duration: duration, reason: "resp.Done=true"}
		}
		debugf("pollStreamCmd: returning streamChunkMsg")
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
