# Hecate's Queue

*Commands from the goddess. Read and obey.*

---

## 📍 CHANNEL TAGGING (NEW RULE — READ THIS)

**Effective immediately:** All RESPONSES.md entries MUST include channel tags.

This is the **[tui]** channel. When writing to RESPONSES.md:
- Prefix your title: `## 2026-02-03 COMPLETE [tui]: Task Name`
- Cross-references: `Related: Depends on [daemon] endpoint`

**All channels:**
- `[tui]` — hecate-tui (this repo)
- `[daemon]` — hecate-daemon
- `[node]` — hecate-node
- `[realm]` — hecate-realm
- `[macula-realm]` — macula-io/macula-realm

**Why:** I monitor multiple repos. Without tags, context is lost. Be explicit.

---

## ⚠️ MANDATORY: Re-read CLAUDE.md NOW

**Before doing anything else this session:**

```bash
cat ~/work/github.com/CLAUDE.md
```

New rules have been added. Pay special attention to:
- **"NEVER DELETE FEATURES"** section
- Read the whole file before editing
- Extend, don't replace

**Acknowledge in RESPONSES.md that you've read it.**

---

## Protocol

| File | Your Access |
|------|-------------|
| `QUEUE.md` | **READ-ONLY** |
| `RESPONSES.md` | Write here |
| `STATUS.md` | Update here |

---

## 🔴 HIGH: Chat View + LLM Client

**TOP PRIORITY. The TUI becomes a window into intelligence.**

Read `plans/PLAN_CHAT_VIEW.md` for the full design.

**Phase 1: Local Chat Only**

Create these files:

```
internal/
├── client/
│   └── llm.go             # LLM methods on existing client
├── llm/
│   ├── types.go           # Message, Model, ChatRequest, etc.
│   └── stream.go          # SSE stream parser
└── views/
    └── chat/
        ├── chat.go        # Main Bubble Tea model
        ├── messages.go    # Message list component
        ├── input.go       # Input textarea
        └── styles.go      # Lip Gloss styles
```

**Implement:**
1. `internal/llm/types.go` — Message, Model, ChatRequest, ChatResponse
2. `internal/llm/stream.go` — SSE parser for streaming responses
3. `internal/client/llm.go` — `ListModels()`, `ChatStream()`
4. `internal/views/chat/` — Bubble Tea chat view
   - Model selector (Tab to cycle)
   - Message history viewport
   - Input textarea
   - Streaming response display

**Key bindings:**
- `Enter` — send message
- `Tab` — cycle models  
- `Ctrl+C` / `Esc` — exit chat view
- `↑/↓` — scroll history

**Depends on:** Daemon `GET /api/llm/models` and `POST /api/llm/chat` endpoints.

The daemon is building the backend: `hecate-daemon/.hecate/QUEUE.md`

**Test flow:**
```bash
# 1. Start Ollama
ollama run llama3.2

# 2. Start daemon
./hecate-daemon

# 3. Start TUI, navigate to Chat view
./hecate-tui
# Press 'c' for chat (or whatever key you assign)
```

**Phase 2 (later):** Mesh discovery, remote model routing.

---

## Active Tasks

### ✅ DONE [tui]: Fix Endpoint Mismatch

Fixed in this session. **COMMIT AND PUSH NOW.**

---

### 🔴 HIGH [tui]: Chat View + LLM Client — UNBLOCKED

**The daemon LLM API is DONE.** You built it yourself:
- Phase 1: `d604efb` — serve_llm app
- Phase 2: `6e40a5b` — mesh announcement
- Phase 3: `3a8278f` — RPC listener

**Endpoints ready:**
```
GET  /api/llm/models   → list Ollama models
POST /api/llm/chat     → chat completion (SSE streaming)
GET  /api/llm/health   → backend status
```

**Proceed with TUI chat view implementation.** See `plans/PLAN_CHAT_VIEW.md`.

---

### 🟡 MEDIUM [tui]: Project Context Support (HECATE.md)

**Hecate TUI is THE AI interface. Not Claude. Not anything else.**

The TUI should read project context files, just like other AI coding tools.

**Context files to support:**

| File | Scope | Purpose |
|------|-------|---------|
| `HECATE.md` | Project root | Project-specific instructions |
| `SKILLS.md` | Project root | Specialized capabilities |
| `.hecate/config.yaml` | Workspace | TUI settings, preferences |
| `.hecate/memory/` | Workspace | Conversation history, context |

**Implementation:**

```
internal/
├── context/
│   ├── loader.go          # Find and load context files
│   ├── hecate_md.go       # Parse HECATE.md
│   ├── skills.go          # Parse SKILLS.md
│   └── memory.go          # Load/save conversation memory
```

**Behavior:**

1. On startup, walk up from cwd looking for `HECATE.md`
2. Load project context into system prompt
3. Load any `SKILLS.md` as additional capabilities
4. Include context when sending chat requests to LLM

**HECATE.md format (same as CLAUDE.md):**

```markdown
# Project Name

Brief description.

## Architecture

Key patterns, conventions.

## Commands

Common tasks, how to run them.

## Guidelines

Do's and don'ts for this project.
```

**This makes Hecate TUI a first-class AI coding assistant.**

---

### 🟡 MEDIUM: Pairing UI Polish

Basic pairing works. Polish it:
- Better QR code display
- Progress indicator during polling
- Nicer success/error states
- Timeout handling

### 🟡 MEDIUM: Identity View

After pairing works, flesh out Identity view:
- Agent MRI and profile
- Pairing status (which realm, when)
- Daemon status (running, version, uptime)
- Re-pair / unpair actions

### 🟢 LOW: Coach Rules Engine

Read the Architecture Decisions in `plans/PLAN_HECATE_STUDIO_UX.md`.

**Coach is rules-based, NOT LLM-based:**
- Detect `services/`, `helpers/`, `utils/` → regex on paths
- Catch central supervisors → naming patterns
- Generate corrections → templates

No LLM needed for doctrine enforcement.

---

## Architecture Decisions (READ THIS)

**1. Macula Services are NOT AI-powered**

Services on the mesh are pure business logic. Deterministic. Testable. No LLM runtime.

**2. Two distinct concerns in the TUI:**

| Concern | Implementation | LLM Required |
|---------|----------------|--------------|
| **Coach** (doctrine enforcement) | Rules engine, pattern matching | No |
| **Studio** (code generation) | LLM generates Cartwheel code | **Yes** |

**3. First-run experience:**
- If no model configured, prompt user to set up
- Detect local Ollama, offer easy path
- Or enter cloud API key

---

## Completed Tasks

### ✅ Basic TUI Structure
- Views: Status, Mesh, Capabilities, RPC, Logs
- Tab navigation
- Daemon client connection

### ✅ Basic Pairing Flow
- QR code display
- Polling logic
- Success/error handling

---

*Ship it.* 🔥🗝️🔥
