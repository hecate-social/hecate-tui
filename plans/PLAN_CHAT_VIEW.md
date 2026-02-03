# Plan: Chat View + LLM Client

**Goal:** TUI can discover LLM models on the mesh and chat with them.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      hecate-tui                          │
│                                                         │
│  ┌─────────────────────────────────────────────────┐    │
│  │                   Chat View                      │    │
│  │  ┌─────────────────────────────────────────┐    │    │
│  │  │ [llama3.2 ▼] connected to hecate-dev    │    │    │
│  │  ├─────────────────────────────────────────┤    │    │
│  │  │                                         │    │    │
│  │  │  You: What is the Macula mesh?          │    │    │
│  │  │                                         │    │    │
│  │  │  🗝️: The Macula mesh is a decentralized │    │    │
│  │  │     network that uses HTTP/3 and QUIC   │    │    │
│  │  │     for NAT-friendly transport...       │    │    │
│  │  │     █                                   │    │    │
│  │  │                                         │    │    │
│  │  ├─────────────────────────────────────────┤    │    │
│  │  │ > Type a message...              [Send] │    │    │
│  │  └─────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────┘    │
│                           │                             │
│                           ▼                             │
│  ┌─────────────────────────────────────────────────┐    │
│  │                  LLM Client                      │    │
│  │   - ListModels() → local + mesh discovered      │    │
│  │   - Chat(model, messages) → stream response     │    │
│  └──────────────────────┬──────────────────────────┘    │
│                         │                               │
└─────────────────────────┼───────────────────────────────┘
                          │ HTTP
                          ▼
                 ┌─────────────────┐
                 │  hecate-daemon  │
                 │    :4444        │
                 └─────────────────┘
```

---

## File Structure

```
internal/
├── client/
│   ├── client.go          # Existing daemon client
│   └── llm.go             # LLM-specific methods (NEW)
│
├── llm/
│   ├── types.go           # Message, Model, ChatRequest, etc.
│   ├── stream.go          # SSE/chunked stream parser
│   └── discovery.go       # Model discovery logic
│
└── views/
    └── chat/
        ├── chat.go        # Main chat view (Bubble Tea model)
        ├── messages.go    # Message list component
        ├── input.go       # Input area component
        ├── selector.go    # Model selector dropdown
        └── styles.go      # Lip Gloss styles
```

---

## Types

```go
// internal/llm/types.go

package llm

type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
)

type Message struct {
    Role    Role   `json:"role"`
    Content string `json:"content"`
}

type Model struct {
    Name     string `json:"name"`      // "llama3.2"
    Provider string `json:"provider"`  // MRI of daemon serving it
    Local    bool   `json:"local"`     // true if from local daemon
    
    // Metadata
    ContextLength int    `json:"context_length,omitempty"`
    Description   string `json:"description,omitempty"`
}

type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Stream      bool      `json:"stream"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
}

type ChatResponse struct {
    Delta string `json:"delta,omitempty"` // Streaming chunk
    Done  bool   `json:"done,omitempty"`  // Stream complete
    
    // Final response (non-streaming or final chunk)
    Content string `json:"content,omitempty"`
    Usage   *Usage `json:"usage,omitempty"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
}
```

---

## LLM Client

```go
// internal/client/llm.go

package client

import (
    "bufio"
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/hecate-social/hecate-tui/internal/llm"
)

// ListModels returns available models (local + discovered)
func (c *Client) ListModels(ctx context.Context) ([]llm.Model, error) {
    resp, err := c.get(ctx, "/api/llm/models")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Models []llm.Model `json:"models"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result.Models, nil
}

// ChatStream sends a chat request and streams responses to a channel
func (c *Client) ChatStream(ctx context.Context, req llm.ChatRequest) (<-chan llm.ChatResponse, error) {
    req.Stream = true
    
    resp, err := c.post(ctx, "/api/llm/chat", req)
    if err != nil {
        return nil, err
    }
    
    ch := make(chan llm.ChatResponse)
    
    go func() {
        defer close(ch)
        defer resp.Body.Close()
        
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            
            // SSE format: "data: {...}"
            if len(line) > 6 && line[:6] == "data: " {
                var chunk llm.ChatResponse
                if err := json.Unmarshal([]byte(line[6:]), &chunk); err != nil {
                    continue
                }
                
                select {
                case ch <- chunk:
                case <-ctx.Done():
                    return
                }
                
                if chunk.Done {
                    return
                }
            }
        }
    }()
    
    return ch, nil
}
```

---

## Chat View (Bubble Tea)

```go
// internal/views/chat/chat.go

package chat

import (
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    
    "github.com/hecate-social/hecate-tui/internal/client"
    "github.com/hecate-social/hecate-tui/internal/llm"
)

type Model struct {
    client   *client.Client
    
    // UI components
    viewport viewport.Model  // Message history
    input    textarea.Model  // User input
    
    // State
    messages     []llm.Message
    models       []llm.Model
    activeModel  int
    streaming    bool
    streamBuf    string  // Current streaming response
    
    // Dimensions
    width  int
    height int
}

func New(c *client.Client) Model {
    ta := textarea.New()
    ta.Placeholder = "Type a message..."
    ta.Focus()
    ta.CharLimit = 4096
    ta.SetHeight(3)
    
    vp := viewport.New(80, 20)
    
    return Model{
        client:   c,
        input:    ta,
        viewport: vp,
        messages: []llm.Message{},
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        textarea.Blink,
        m.fetchModels(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            return m, tea.Quit
            
        case "enter":
            if !m.streaming && m.input.Value() != "" {
                return m, m.sendMessage()
            }
            
        case "tab":
            // Cycle through models
            if len(m.models) > 0 {
                m.activeModel = (m.activeModel + 1) % len(m.models)
            }
        }
        
    case modelsMsg:
        m.models = msg.models
        
    case streamChunkMsg:
        m.streamBuf += msg.delta
        m.updateViewport()
        
    case streamDoneMsg:
        // Finalize assistant message
        m.messages = append(m.messages, llm.Message{
            Role:    llm.RoleAssistant,
            Content: m.streamBuf,
        })
        m.streamBuf = ""
        m.streaming = false
        m.updateViewport()
        
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 6  // Leave room for input + header
        m.input.SetWidth(msg.Width - 2)
    }
    
    // Update sub-components
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    cmds = append(cmds, cmd)
    
    m.viewport, cmd = m.viewport.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    // Header: model selector
    header := m.renderHeader()
    
    // Messages viewport
    messages := m.viewport.View()
    
    // Input area
    input := m.input.View()
    
    return header + "\n" + messages + "\n" + input
}

// Commands

func (m Model) fetchModels() tea.Cmd {
    return func() tea.Msg {
        models, err := m.client.ListModels(context.Background())
        if err != nil {
            return errMsg{err}
        }
        return modelsMsg{models}
    }
}

func (m *Model) sendMessage() tea.Cmd {
    content := m.input.Value()
    m.input.Reset()
    
    // Add user message
    m.messages = append(m.messages, llm.Message{
        Role:    llm.RoleUser,
        Content: content,
    })
    m.streaming = true
    m.updateViewport()
    
    return func() tea.Msg {
        model := ""
        if len(m.models) > 0 {
            model = m.models[m.activeModel].Name
        }
        
        ch, err := m.client.ChatStream(context.Background(), llm.ChatRequest{
            Model:    model,
            Messages: m.messages,
        })
        if err != nil {
            return errMsg{err}
        }
        
        // Return command that reads from stream
        return streamStartMsg{ch: ch}
    }
}

// Messages

type modelsMsg struct{ models []llm.Model }
type streamStartMsg struct{ ch <-chan llm.ChatResponse }
type streamChunkMsg struct{ delta string }
type streamDoneMsg struct{ usage *llm.Usage }
type errMsg struct{ err error }
```

---

## Model Discovery

```go
// internal/llm/discovery.go

package llm

// Discovery priority:
// 1. Local daemon models (fastest)
// 2. LAN mesh models (low latency)  
// 3. WAN mesh models (fallback)

type DiscoverySource int

const (
    SourceLocal DiscoverySource = iota
    SourceLAN
    SourceWAN
)

func (m Model) Source() DiscoverySource {
    if m.Local {
        return SourceLocal
    }
    // TODO: Determine LAN vs WAN from provider MRI
    return SourceWAN
}

// SortByLatency sorts models by expected latency
func SortByLatency(models []Model) []Model {
    // Local first, then by discovery source
    sort.Slice(models, func(i, j int) bool {
        return models[i].Source() < models[j].Source()
    })
    return models
}
```

---

## Phases

### Phase 1: Local Only
- [ ] LLM client methods in `internal/client/llm.go`
- [ ] Basic chat view with message history
- [ ] Streaming response display
- [ ] Model selector (local models only)

### Phase 2: Mesh Discovery
- [ ] Discovery logic for mesh capabilities
- [ ] Display model source (local/LAN/WAN)
- [ ] Route requests to remote daemons via mesh

### Phase 3: Polish
- [ ] Syntax highlighting for code blocks
- [ ] Message editing/regeneration
- [ ] Conversation persistence
- [ ] System prompt configuration

---

## Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Tab` | Cycle through models |
| `Ctrl+N` | New conversation |
| `Ctrl+S` | Save conversation |
| `Ctrl+C` / `Esc` | Exit |
| `↑/↓` | Scroll history |

---

*The TUI becomes a window into the mesh's collective intelligence.* 🗝️
