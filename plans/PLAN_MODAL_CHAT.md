# Plan: Modal Chat Interface

*The tab-based UI was the wrong metaphor. Hecate is a conversational gateway, not a dashboard. The TUI should be a dialogue with powers, not a collection of views.*

## Philosophy

**Chat is home.** Everything else is a command or a mode you visit and return from. Inspired by vim's modal editing — distinct keybindings per mode, but without composable counts or macros. Powerful enough for terminal natives, learnable for everyone else.

The agent herself teaches you: type `?` and she tells you what mode you're in and what keys work.

---

## Modes

### 1. Normal Mode (default)

The resting state. You see the chat history. Nothing is focused for input.

| Key | Action |
|-----|--------|
| `i` | Enter **Insert** mode (focus textarea) |
| `/` | Enter **Command** mode (slash command) |
| `:` | Enter **Command** mode (vim-style) |
| `j` / `k` | Scroll chat history (line) |
| `Ctrl+D` / `Ctrl+U` | Scroll half-page down/up |
| `g` | Jump to top of chat |
| `G` | Jump to bottom of chat |
| `?` | Show contextual help overlay |
| `q` | Quit |

**Visual indicator:** `-- NORMAL --` in status bar.

### 2. Insert Mode

You're typing a message to send. The textarea is focused.

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Esc` | Cancel, return to **Normal** |
| `Tab` | Cycle LLM model |
| `Ctrl+L` | Clear chat |
| Standard text editing | Type freely |

**Visual indicator:** `-- INSERT --` in status bar. Textarea border highlights.

### 3. Command Mode

Triggered by `/` or `:` from Normal mode. A command input line appears at the bottom (vim-style).

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `Esc` | Cancel, return to **Normal** |
| `Tab` | Autocomplete command |
| `↑` / `↓` | Command history |

**Visual indicator:** `-- COMMAND --` in status bar. Command line at bottom: `/` or `:` prompt.

### 4. Browse Mode

Entered via `/browse`. A panel overlays the chat showing a navigable list (capabilities, agents, etc.).

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate list items |
| `Enter` | View detail / select |
| `/` | Filter/search within list |
| `Esc` | Return to **Normal** |
| `?` | Help for browse mode |

**Visual indicator:** `-- BROWSE --` in status bar. List panel rendered.

### 5. Pair Mode

Entered via `/pair`. Shows pairing flow inline. Auto-returns to Normal on completion or cancel.

| Key | Action |
|-----|--------|
| `Esc` / `c` | Cancel pairing, return to **Normal** |

**Visual indicator:** `-- PAIR --` in status bar.

---

## Slash Commands

Commands are self-registering. Each implements a `Command` interface.

### Core Commands

| Command | Action | Output |
|---------|--------|--------|
| `/help` or `?` | Show available commands + current mode help | Inline help card |
| `/status` | Daemon health, identity, mesh, models | Inline status card |
| `/browse [type]` | Enter Browse mode. Optional: `llm`, `capability`, `agent` | Mode switch |
| `/models` | List available LLM models | Inline list |
| `/model <name>` | Switch active model | Confirmation |
| `/me` | Identity, realm, stats | Inline card |
| `/pair` | Start pairing flow | Mode switch |
| `/config [key] [value]` | View or set config | Inline display |
| `/project` | Current project context | Inline card |
| `/tools` | Detected dev tools | Inline list |
| `/health` | Quick daemon health check | Inline status |
| `/clear` | Clear chat history | - |
| `/quit` or `:q` | Quit TUI | - |
| `:w` | Save chat transcript | Confirmation |

### Future / Extensible

| Command | Action |
|---------|--------|
| `/mesh peers` | Show connected mesh peers |
| `/ucan grant <target>` | Grant UCAN capability |
| `/subscribe <mri>` | Subscribe to capability |
| `/rpc <mri> <action>` | Direct RPC call |
| `/log [level]` | View/set log level |

Commands registered from mesh capabilities could extend this further.

---

## Layout

```
┌─────────────────────────────────────────────┐
│  🔥🗝️🔥 Hecate  ·  llama3.2  ·  ● mesh    │  ← Header (always visible)
├─────────────────────────────────────────────┤
│                                             │
│  ▸ You                                      │
│  What LLM models do you have?               │
│                                             │
│  ◆ Hecate                                   │
│  I have 3 models available:                 │
│  • llama3.2 (3B) — fast, general purpose    │
│  • qwen2.5-coder (7B) — code optimized     │
│  • deepseek-r1 (8B) — chain of thought     │
│                                             │
│  Use /model <name> to switch.               │
│                                             │  ← Chat area (scrollable)
├─────────────────────────────────────────────┤
│ /browse llm█                                │  ← Command line (Command mode)
│                                             │  ← OR: textarea (Insert mode)
│                                             │  ← OR: empty (Normal mode)
├─────────────────────────────────────────────┤
│  -- COMMAND --          j/k:scroll  ?:help  │  ← Status bar
└─────────────────────────────────────────────┘
```

### Browse Mode Overlay

When in Browse mode, a panel appears over the right side (or full width on narrow terminals):

```
┌─────────────────────────────────────────────┐
│  🔥🗝️🔥 Hecate  ·  llama3.2  ·  ● mesh    │
├────────────────────┬────────────────────────┤
│                    │  Browse: Capabilities   │
│  (chat visible     │                        │
│   but dimmed)      │  ▸ llm/llama3.2       │
│                    │    llm/qwen2.5-coder   │
│                    │    llm/deepseek-r1     │
│                    │    translation/opus     │
│                    │                        │
│                    │  4 capabilities        │
├────────────────────┴────────────────────────┤
│  -- BROWSE --    j/k:nav  /:filter  esc:back│
└─────────────────────────────────────────────┘
```

On narrow terminals (< 100 cols), browse takes full width instead of split.

---

## Status Bar

Always visible at the bottom. Shows:

```
-- MODE --    [model: llama3.2]    [mesh: ●]    [hints]
```

- **Mode** — current mode name, highlighted
- **Model** — active LLM model (if any)
- **Mesh** — connection status (● connected, ○ disconnected)
- **Hints** — contextual keybinding hints for current mode

---

## Architecture

### Directory Structure

```
hecate-tui/
├── internal/
│   ├── app/
│   │   ├── app.go              # Root Bubble Tea model
│   │   ├── modes.go            # Mode enum + state machine
│   │   └── keymap.go           # Per-mode key dispatch
│   ├── chat/
│   │   ├── chat.go             # Chat renderer (messages, bubbles)
│   │   ├── styles.go           # Chat styling (reuse existing)
│   │   └── input.go            # Textarea for Insert mode
│   ├── commands/
│   │   ├── registry.go         # Command registry + dispatch
│   │   ├── command.go          # Command interface
│   │   ├── status.go           # /status
│   │   ├── browse.go           # /browse → enters Browse mode
│   │   ├── models.go           # /models, /model
│   │   ├── pair.go             # /pair → enters Pair mode
│   │   ├── me.go               # /me
│   │   ├── config.go           # /config
│   │   ├── project.go          # /project
│   │   ├── tools.go            # /tools
│   │   ├── health.go           # /health
│   │   └── help.go             # /help
│   ├── modes/
│   │   ├── browse.go           # Browse mode overlay
│   │   ├── pair.go             # Pair mode flow
│   │   └── visual.go           # Visual mode (future)
│   ├── statusbar/
│   │   └── statusbar.go        # Status bar renderer
│   ├── client/                 # API client (reuse as-is)
│   │   ├── client.go
│   │   └── llm.go
│   ├── llm/                    # LLM types (reuse as-is)
│   │   ├── types.go
│   │   └── stream.go
│   └── tools/                  # Tool detection (reuse as-is)
│       ├── detector.go
│       ├── config.go
│       └── launcher.go
```

### Command Interface

```go
type Command interface {
    Name() string                    // e.g. "browse"
    Aliases() []string               // e.g. ["b"]
    Description() string             // Short description for /help
    Execute(args []string, ctx *Context) tea.Cmd
}

type Context struct {
    Client    *client.Client
    Width     int
    Height    int
    Mode      *Mode                  // Can trigger mode switch
    Chat      *ChatState             // Can inject messages
}
```

### Mode State Machine

```go
type Mode int

const (
    ModeNormal Mode = iota
    ModeInsert
    ModeCommand
    ModeBrowse
    ModePair
    ModeVisual  // future
)
```

Transitions are explicit:
- `Normal` → `Insert` (press `i`)
- `Normal` → `Command` (press `/` or `:`)
- `Command` → `Browse` (execute `/browse`)
- `Command` → `Pair` (execute `/pair`)
- `*` → `Normal` (press `Esc`)

---

## Migration from Tab-Based

### What to keep (reuse directly)
- `internal/client/` — all API calls
- `internal/llm/` — types and streaming
- `internal/tools/` — detection, config, launcher
- `internal/views/chat/styles.go` — bubble styles, colors
- `internal/views/chat/chat.go` — message rendering logic (extract into `chat/chat.go`)
- `internal/editor/` — built-in editor (for `/edit` command, future)

### What to restructure
- `internal/views/browse/` → `internal/modes/browse.go` (list logic, not a full view)
- `internal/views/monitor/` → `internal/commands/status.go` (inline card renderer)
- `internal/views/pair/` → `internal/modes/pair.go` (flow logic)
- `internal/views/me/` → `internal/commands/me.go` (inline card)
- `internal/views/projects/` → `internal/commands/project.go` (inline card)
- `internal/ui/app.go` → `internal/app/app.go` (modal state machine, not tab switcher)

### What to delete
- `internal/views/views.go` (Tab enum, View interface — replaced by modes)
- Tab navigation logic in app.go

---

## Implementation Phases

### Phase 1: Modal Foundation
- [ ] Mode state machine (`app/modes.go`)
- [ ] Per-mode key dispatch (`app/keymap.go`)
- [ ] Root app model with Normal/Insert/Command modes
- [ ] Status bar with mode indicator
- [ ] Command line input (bottom of screen, vim-style)
- [ ] Chat renderer extracted from existing code
- [ ] `i` to insert, `Esc` to normal, `j/k` scroll

### Phase 2: Command System
- [ ] Command registry + interface
- [ ] Command autocomplete (Tab)
- [ ] Command history (↑/↓)
- [ ] Implement: `/help`, `/status`, `/health`, `/clear`, `/quit`
- [ ] Implement: `/models`, `/model <name>`
- [ ] Implement: `/me`, `/config`, `/tools`
- [ ] Inline card rendering (status cards, lists, identity)

### Phase 3: Browse Mode
- [ ] Browse mode overlay (split or full-width)
- [ ] `/browse` command → enter mode
- [ ] `j/k` navigation, `Enter` detail, `/` filter, `Esc` back
- [ ] Browse types: capabilities, agents, models

### Phase 4: Pair Mode + Project
- [ ] `/pair` command → pair mode
- [ ] Pairing flow (reuse existing state machine)
- [ ] `/project` command → inline project card

### Phase 5: Polish
- [ ] Contextual `?` help per mode
- [ ] Smooth transitions between modes (optional animation)
- [ ] Terminal width adaptation (split vs full browse)
- [ ] Chat transcript save (`:w`)
- [ ] Welcome screen with the Threshold Guardian

---

## Open Questions

1. **Command output rendering** — Should command output (e.g., `/status`) appear as a "system message" in the chat stream, or as a temporary overlay that disappears? Chat stream is simpler and preserves history.

2. **Inline vs overlay for Browse** — Split pane (chat + browse side by side) or full takeover? Split for wide terminals, full for narrow?

3. **Command prefix** — `/` for commands, `:` for vim-style (`:q`, `:w`). Both enter command mode. Should we differentiate or treat them identically?

---

*The goddess speaks through dialogue, not dashboards.* 🔥🗝️🔥
