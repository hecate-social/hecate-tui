# Hecate's Queue

*Commands from the goddess. Read and obey.*

---

## 📍 CHANNEL TAGGING (MANDATORY)

This is the **[tui]** channel. Tag all RESPONSES.md entries:
- `## 2026-02-03 COMPLETE [tui]: Task Name`
- Cross-refs: `Related: [daemon] endpoint`

---

## 📖 READ FIRST

1. `cat ~/work/github.com/CLAUDE.md` — Re-read every session
2. `plans/PLAN_MODAL_CHAT.md` — **The new master plan (APPROVED). This REPLACES PLAN_DEVELOPER_STUDIO.md**

---

## ⚠️ ARCHITECTURAL PIVOT: Modal Chat Interface

**The tab-based UI is DEAD.** We're pivoting to a **modal chat interface** inspired by vim.

**Why:** Hecate is a conversational gateway, not a dashboard. Chat is the primary interface. Everything else is a command or a mode you visit and return from.

**Read `plans/PLAN_MODAL_CHAT.md` carefully.** It defines modes, keybindings, commands, layout, and architecture.

**Key principles:**
- Chat is home. Always visible. Everything returns here.
- Modes: Normal → Insert → Command → Browse → Pair
- `/` commands for actions. Self-registering. Extensible.
- `j/k` navigation. `Esc` always returns to Normal.
- Status bar shows current mode + contextual hints.

---

## ✅ Completed (Tab-Based Era — ARCHIVED)

These are done but the tab-based UI is being replaced:

- [x] Chat View (local LLM) — streaming, bubbles, model selector
- [x] Chat Welcome Avatar (Threshold Guardian)
- [x] Tab navigation (6 tabs)
- [x] Browse view (search, detail)
- [x] Monitor view (stats cards, two-column)
- [x] Me view (profile, settings)
- [x] Pair view (flow state machine)
- [x] Projects view (phase navigation)
- [x] Tool integration (detector, config, launcher, editor)
- [x] Endpoint mismatch fixes

**What SURVIVES the pivot (reuse directly):**
- `internal/client/` — all API calls (untouched)
- `internal/llm/` — types and streaming (untouched)
- `internal/tools/` — detection, config, launcher (untouched)
- Chat rendering logic (bubbles, styles, streaming animation)
- Built-in editor

**What gets RESTRUCTURED:**
- `internal/views/*` → extracted into `commands/` and `modes/`
- `internal/ui/app.go` → rewritten as modal state machine

---

## 🔴 Phase 1: Modal Foundation (NOW)

### 1.1 Mode State Machine

Create the modal core. This is THE fundamental change.

**Files:**
```
internal/app/
├── app.go              # Root Bubble Tea model (replaces ui/app.go)
├── modes.go            # Mode enum + transition rules
└── keymap.go           # Per-mode key dispatch
```

**Mode enum:**
```go
type Mode int
const (
    ModeNormal Mode = iota
    ModeInsert
    ModeCommand
    ModeBrowse
    ModePair
)
```

**Transitions:**
- `Normal` → `Insert` (press `i`)
- `Normal` → `Command` (press `/` or `:`)
- `Command` → `Browse` (execute `/browse`)
- `Command` → `Pair` (execute `/pair`)
- `*` → `Normal` (press `Esc`)

**Key dispatch:** Each mode has its own key handler. `app.go` delegates to the current mode's handler.

### 1.2 Chat Renderer

Extract chat rendering from existing `views/chat/chat.go` into standalone renderer.

**Files:**
```
internal/chat/
├── chat.go             # Message list renderer + viewport
├── input.go            # Textarea for Insert mode
└── styles.go           # Reuse existing bubble styles
```

The chat is NOT a "view" anymore — it's the canvas. Always visible. Messages render in the main area. The textarea appears/disappears based on Insert mode.

### 1.3 Command System

The slash command framework.

**Files:**
```
internal/commands/
├── command.go          # Command interface
├── registry.go         # Command registry + dispatch + autocomplete
├── help.go             # /help
├── status.go           # /status (inline card)
├── health.go           # /health
├── models.go           # /models, /model <name>
├── me.go               # /me (inline card)
├── clear.go            # /clear
└── quit.go             # /quit, :q
```

**Command interface:**
```go
type Command interface {
    Name() string
    Aliases() []string
    Description() string
    Execute(args []string, ctx *Context) tea.Cmd
}
```

**Context provides:** client, terminal size, mode setter, chat message injector.

Command output appears as "system messages" in the chat stream. This preserves history and is simple to implement.

### 1.4 Status Bar

Always visible at the bottom. Mode indicator + model + mesh status + hints.

**Files:**
```
internal/statusbar/
└── statusbar.go
```

### 1.5 Layout Assembly

Wire it all together:
```
┌──────────────────────────────────┐
│ Header (title, model, mesh)      │
├──────────────────────────────────┤
│                                  │
│ Chat area (scrollable)           │
│                                  │
├──────────────────────────────────┤
│ Input area (mode-dependent)      │
│ - Normal: empty                  │
│ - Insert: textarea               │
│ - Command: command line (/...)   │
├──────────────────────────────────┤
│ Status bar (mode, model, hints)  │
└──────────────────────────────────┘
```

**Acceptance criteria:**
- `i` enters Insert mode, textarea appears, `Esc` returns to Normal
- `j/k` scrolls chat in Normal mode
- `/` opens command line, typing a command + Enter executes it
- `/help` shows available commands as a system message in chat
- `/status` shows daemon status as an inline card
- Status bar updates with current mode
- LLM chat works (type in Insert mode, send, see streaming response)

---

## 🟡 Phase 2: Browse Mode

### 2.1 Browse Mode Overlay

Entered via `/browse`. Shows capability list as an overlay panel.

**Files:**
```
internal/modes/
└── browse.go           # Browse mode: list, navigation, filter, detail
```

**Keybindings in Browse mode:** `j/k` navigate, `Enter` detail, `/` filter, `Esc` back to Normal.

**Layout:** Split pane (chat dimmed left, browse right) on wide terminals. Full-width browse on narrow terminals.

### 2.2 Browse Command

```
internal/commands/
└── browse.go           # /browse [type] → enters Browse mode
```

---

## 🟡 Phase 3: Pair Mode + Utilities

### 3.1 Pair Mode

Entered via `/pair`. Reuses existing pairing state machine logic.

```
internal/modes/
└── pair.go
```

### 3.2 Utility Commands

```
internal/commands/
├── config.go           # /config [key] [value]
├── tools.go            # /tools
└── project.go          # /project
```

---

## 🟢 Phase 4: Polish

- Contextual `?` help per mode
- Command autocomplete with Tab
- Command history with ↑/↓
- Terminal width adaptation
- `:w` to save chat transcript
- Welcome screen with Threshold Guardian
- Smooth mode transitions

---

## Architecture Notes

### Command Output = System Messages

When `/status` runs, its output becomes a system message in the chat:

```
◆ System
┌─────────────────────────────┐
│ Status                       │
│ Daemon: ● Running            │
│ Mesh:   ● Connected          │
│ Models: 3 available          │
│ Tests:  85 passing           │
└─────────────────────────────┘
```

This is simpler than overlays and preserves history. You can scroll up to see past command results.

### Mode Transitions Are Explicit

Never silently switch modes. The status bar MUST update. The input area MUST change. The user should always know where they are.

### Esc Is Sacred

`Esc` ALWAYS returns toward Normal mode. No exceptions. No "Esc does different things in different contexts." It goes home.

---

*The goddess speaks through dialogue, not dashboards.* 🔥🗝️🔥
