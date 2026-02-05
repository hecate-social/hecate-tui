# Apprentice Status

*Current state of the apprentice's work.*

---

## Current Task

**COMPLETE: Connector Architecture — Unix Socket Transport + Config Consolidation**

## Last Active

**2026-02-05**

---

## Session Log

### 2026-02-05 Session (Connector Architecture — Phases 4-5)

**Status:** Complete

**Completed:**
- **Phase 4: Unix Socket Transport**
  - Added `NewWithSocket(socketPath)` to `internal/client/client.go`:
    - Custom `http.Transport` with Unix socket `DialContext`
    - `Transport()` accessor for SSE streaming reuse
  - Updated `internal/client/llm.go`:
    - SSE streaming client reuses socket transport
  - Rewrote `cmd/hecate-tui/main.go`:
    - 4-level connection priority: HECATE_SOCKET env → default socket → HECATE_URL env → TCP default
    - Version bumped to 0.3.0
    - Updated help text with socket documentation
  - Added `NewWithSocket()` to `internal/app/app.go`

- **Phase 5: Config Consolidation + Migration**
  - Rewrote `internal/config/config.go` to consolidated TOML:
    - New path: `~/.config/hecate-tui/config.toml`
    - Config struct: Theme, SystemPrompt, Connection (SocketPath, DaemonURL, Timeout), Editor, UI
    - Migration: reads old JSON + old TOML, merges, writes new, renames old to `.migrated`
  - Updated `internal/config/history.go`:
    - New path: `~/.config/hecate-tui/conversations/`
    - Auto-migration of old directory
  - Updated `internal/commands/config.go`:
    - Shows socket path with existence status in /config output
  - Updated `internal/app/app.go`:
    - Changed `cfg.DaemonURL` field access to `cfg.DaemonURL()` method call

**Build:** `go build ./...` — Clean

**Modified files:**
```
internal/client/client.go           # NewWithSocket, Transport accessor
internal/client/llm.go              # Socket transport reuse for SSE
cmd/hecate-tui/main.go              # Connection resolution, help text, v0.3.0
internal/app/app.go                 # NewWithSocket, DaemonURL() method call
internal/config/config.go           # TOML consolidation, migration
internal/config/history.go          # New conversations path, migration
internal/commands/config.go         # Socket path display
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 10: Housekeeping & Workflow)

**Status:** Complete

**Completed:**
- Created `/delete` command (`/del` alias) in `conversation.go`:
  - Delete conversations by ID or numeric index from `/history`
  - Uses `config.DeleteConversation()` for filesystem removal
- Created `DeleteConversation()` in `config/history.go`:
  - Removes conversation JSON file by ID
  - Returns error if conversation not found
- Created `/find` command (`/search`, `/f` aliases) in `find.go`:
  - Case-insensitive search through current chat messages
  - Shows role, timestamp, and up to 3 matching lines per message
  - Match count summary at bottom
- Conversation title in header bar:
  - `conversationTitle` field on App struct
  - `renderHeader()` displays title after daemon status indicator
  - Title set on auto-load at startup
  - Title set when loading via `/load`
  - Title cleared when starting new conversation via `/new`
  - Title auto-updated when saving conversation (derives from first user message)
- Updated main.go help text with `/delete` and `/find` commands
- Registered `/delete` and `/find` in command registry

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/commands/find.go              # /find command
```

**Modified files:**
```
internal/app/app.go                     # Header title, load/new title wiring, save title update
internal/commands/conversation.go       # /delete command added
internal/commands/registry.go           # Register /delete, /find
internal/config/history.go             # DeleteConversation()
cmd/hecate-tui/main.go                 # Help text
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 9: Chat Persistence & Conversations)

**Status:** Complete

**Completed:**
- Created conversation persistence layer (`internal/config/history.go`):
  - Conversations saved as JSON in `~/.config/hecate/conversations/`
  - Each conversation: ID (timestamp-based), title (from first user message), model, messages, timestamps
  - `SaveConversation`, `LoadConversation`, `ListConversations` (sorted newest first)
  - `TitleFromMessages` auto-derives title from first user message
  - System messages excluded from persistence (command outputs are ephemeral)
- Auto-save on every user message send and stream completion:
  - Detects streaming -> not-streaming transition in Update loop
  - Also saves immediately when user presses Enter in Insert mode
  - `saveConversation()` converts chat messages to persistent format
- Auto-load most recent conversation on startup:
  - `New()` checks `ListConversations()` and loads the latest
  - Messages restored with full content and timestamps
  - If no conversations exist, starts fresh with new ID
- Created `/new` command (`/n` alias):
  - Saves current conversation, clears chat, starts fresh with new ID
- Created `/history` command (`/hist` alias):
  - Lists up to 10 most recent conversations
  - Shows title, date, message count, model name
  - Shows conversation ID for `/load`
  - "...N more" indicator for overflow
- Created `/load <id|number>` command:
  - Load by conversation ID string: `/load 20260205-143022`
  - Load by index from /history: `/load 1` (most recent)
  - Saves current conversation before loading
- Updated main.go help text with /new, /history, /load
- `LoadMessages()` and `Messages()` public methods on chat.Model

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/config/history.go              # Conversation persistence
internal/commands/conversation.go       # /new, /history, /load commands
```

**Modified files:**
```
internal/app/app.go                     # Auto-save/load, conversation lifecycle
internal/chat/chat.go                   # LoadMessages, Messages methods
internal/commands/command.go            # NewConversationMsg, LoadConversationMsg
internal/commands/registry.go           # Register new commands
cmd/hecate-tui/main.go                  # Help text
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 8: Input & Conversation Polish)

**Status:** Complete

**Completed:**
- Multiline input in Insert mode:
  - `Alt+Enter` inserts a newline at cursor position
  - `Enter` still sends the message (unchanged behavior)
  - `InsertNewline()` method on chat model via textarea.InsertString
  - Updated mode hints and help text
- Retry last message (`r` in Normal mode):
  - `RetryLast()` finds last user message, removes any assistant response after it
  - Re-triggers streaming with the same conversation context
  - No-op if no user messages exist or if already streaming
- Yank/Copy last response (`y` in Normal mode):
  - Copies last assistant message to system clipboard via `atotto/clipboard`
  - Shows truncated preview as system message on success
  - Graceful fallback if clipboard unavailable or no response exists
- Character count in Insert mode status bar:
  - `InputLen` field on status bar model, updated on every keypress
  - Shows "N chars" prefix in hints when input has content
  - `InputLen()` method on chat model returns current input length
- Updated all help text (modehelp.go, modes.go hints, main.go --help)

**Build:** `go build ./...` + `go vet ./...` — Clean

**Modified files:**
```
internal/app/app.go                 # Alt+Enter, retry, yank, InputLen wiring
internal/chat/chat.go               # InsertNewline, InputLen, RetryLast, LastAssistantMessage
internal/commands/modehelp.go        # Updated Insert + Normal mode help
internal/modes/modes.go             # Updated hints for Normal + Insert
internal/statusbar/statusbar.go     # InputLen display in Insert mode
cmd/hecate-tui/main.go              # Updated --help text
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 7: Session Persistence & Live Status)

**Status:** Complete

**Completed:**
- Created persistent config system (`internal/config/config.go`):
  - Saves/loads to `~/.config/hecate/config.json` (XDG-compliant)
  - Persists theme choice, system prompt, and daemon URL
  - Auto-creates config directory on first save
  - Loaded at startup in `app.New()` — restores saved preferences
  - Theme changes auto-save via `saveThemeToConfig()`
  - System prompt changes auto-save via `SetSystemPrompt` callback
  - Config can override daemon URL (env var takes precedence if non-default)
- Added periodic health polling (30-second interval):
  - `healthTickMsg` triggers re-poll of daemon health and models
  - Status bar stays fresh without manual refresh
  - Polls daemon health + model list each tick
  - Scheduled via `tea.Tick(30*time.Second, ...)`
- Added `?` help key to Browse and Pair modes:
  - `?` now shows contextual mode help in Browse and Pair modes (not just Normal)
  - Edit mode intentionally excluded — `?` is a typeable character in the editor
- Updated `/config` command to show config file path
- Updated `/subscriptions` command (Phase 6 leftover: removed duplicate `itoa`)

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/config/config.go           # Persistent config (load/save JSON)
```

**Modified files:**
```
internal/app/app.go                 # Config loading, health polling, ? in Browse/Pair
internal/commands/config.go         # Shows config file path
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 6: Cleanup & Polish)

**Status:** Complete

**Completed:**
- Deleted dead tab-based UI code:
  - Removed `internal/ui/` (old tab-based app + styles)
  - Removed `internal/views/` (old browse, chat, me, monitor, pair, projects views)
  - Verified no imports from new code reference these packages
- Fixed editor save feedback bug:
  - `saveResultMsg` was returned by `save()` but never handled in `Update()`
  - Added handler: shows "Saved: filename" on success, error message on failure
  - Clears `modified` flag on successful save (prevents false "unsaved changes" warning)
- Created `/subscriptions` command (`internal/commands/subscriptions.go`):
  - Lists active mesh subscriptions via `client.ListSubscriptions()`
  - Shows MRI, subscription ID, and subscribed-at timestamp per subscription
  - Graceful empty state: "No active subscriptions."
  - Alias: `/subs`
- Updated main.go help text with `/subs` command

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/commands/subscriptions.go  # /subscriptions command
```

**Modified files:**
```
internal/editor/editor.go          # saveResultMsg handler
internal/commands/registry.go      # Register /subscriptions
cmd/hecate-tui/main.go             # Help text update
```

**Deleted:**
```
internal/ui/                       # Old tab-based app (dead code)
internal/views/                    # Old tab-based views (dead code)
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 5)

**Status:** Complete

**Completed:**
- Command history with up/down navigation in Command mode:
  - Up arrow browses backwards through history (up to 50 entries)
  - Down arrow browses forward, restoring draft on exit
  - Draft input preserved when browsing (saved on first up, restored on down past end)
  - History deduplicates consecutive identical entries
  - Any non-nav key resets history browsing index
- Created `/save` command (`internal/commands/save.go`):
  - Exports chat transcript to markdown file
  - Default filename: `hecate-chat-{timestamp}.md`
  - Custom filename: `/save mychat.md`
  - Includes role labels, timestamps, and full message content
  - Alias: `/w` (vim-style)
- Created `/system` command (`internal/commands/system.go`):
  - `/system` — view current system prompt
  - `/system clear` — remove system prompt
  - `/system <text>` — set system prompt for LLM
  - Alias: `/sys`
- Created `/edit` command (`internal/commands/edit.go`):
  - `/edit` — open scratch buffer
  - `/edit <path>` — open file in built-in editor
  - Emits EditFileMsg handled by app
  - Alias: `/e`
- Wired Edit mode into app:
  - handleEditKey intercepts Ctrl+Q/Esc to close editor (prevents app quit)
  - Non-key messages (blink cursor) forwarded to editor separately
  - openEditor initializes editor with file or scratch buffer
  - renderEditLayout: full-screen editor + status bar
  - EditMode added to modes.go (mode 5), theme, and statusbar
- Created contextual help system (`internal/commands/modehelp.go`):
  - `ModeHelp(mode, ctx)` returns mode-specific help as system message
  - Detailed keybinding help for all 6 modes (Normal, Insert, Command, Browse, Pair, Edit)
  - `?` key in Normal mode now shows contextual help instead of /help
- Extended `commands.Context` with message access callbacks:
  - `GetMessages()` — export chat messages for /save
  - `GetSystemPrompt()` / `SetSystemPrompt()` — system prompt access for /system
- Updated main.go help text with all Phase 5 commands and Edit mode

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/commands/save.go         # /save command
internal/commands/system.go       # /system command
internal/commands/edit.go         # /edit command
internal/commands/modehelp.go     # Contextual mode help
```

**Modified files:**
```
internal/app/app.go               # Command history, editor wiring, system prompt, contextual help
internal/chat/chat.go             # System prompt, ExportMessages, SetSystemPrompt
internal/commands/command.go      # Context callbacks (GetMessages, Get/SetSystemPrompt)
internal/commands/registry.go     # Register /save, /system, /edit, modehelp
internal/modes/modes.go           # Edit mode (5)
internal/theme/theme.go           # EditMode style
internal/statusbar/statusbar.go   # Edit mode in modeStyle()
cmd/hecate-tui/main.go            # Help text update
```

**Next:** Phase 6 — TBD (persistent config, keybindings, polish)

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 4)

**Status:** Complete

**Completed:**
- Created lightweight markdown renderer (`internal/chat/markdown.go`):
  - Code blocks with language label and themed background
  - `**bold**`, `*italic*`, `` `inline code` `` formatting
  - `# h1`, `## h2`, `### h3` headers with distinct theme colors
  - Bullet lists (`-`, `*`) with themed markers
  - Numbered lists with styled numbers
  - Horizontal rules (`---`)
  - No external dependencies — uses lipgloss + theme colors
  - Integrated into assistant message rendering
- Polished streaming UX:
  - Shows active model name during streaming ("Channeling via model...")
  - Elapsed time counter during streaming (e.g., "3.2s")
  - "Esc to cancel" hint during streaming
  - Stats cleared when starting a new response (no stale stats)
  - Duration shown in completion stats
  - FormatTokens/FormatSpeed now use theme colors instead of hardcoded hex
- Created `/call` command (`internal/commands/call.go`):
  - Invokes RPC procedures on the mesh by MRI
  - Optional JSON args: `/call mri:proc:io.macula/echo {"msg":"hello"}`
  - Pretty-prints JSON result with duration
  - Error display for failed calls
  - Alias: `/rpc`
- Added RPC client method (`internal/client/rpc.go`):
  - `RPCCall(procedure, args)` — POST /api/rpc/call
  - Returns RPCResult with parsed JSON result, error, duration
- Added message timestamps:
  - All messages now carry a `time.Time` field
  - User and assistant messages show "HH:MM" next to the role label
  - System messages remain clean (no timestamp clutter)

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/chat/markdown.go         # Lightweight markdown renderer
internal/client/rpc.go            # RPC call client method
internal/commands/call.go         # /call command
```

**Modified files:**
```
internal/chat/chat.go             # Timestamps, markdown integration, streaming polish
internal/chat/styles.go           # Theme-based colors for FormatTokens/FormatSpeed
internal/commands/registry.go     # Register /call
cmd/hecate-tui/main.go            # Help text update
```

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 3)

**Status:** Complete

**Completed:**
- Created Pair mode overlay (`internal/pair/pair.go`):
  - Full pairing state machine (Idle, Starting, Waiting, Paired, Error)
  - HandleKey pattern matching Browse mode (returns consumed, cmd)
  - Themed with `*theme.Theme` — no hardcoded colors
  - Spinner + polling for pairing confirmation (2s interval)
  - Code display with double-border box for realm pairing code
  - Step-by-step instructions for the pairing flow
  - Connected info box when paired
  - Split pane on wide terminals (>=100 cols)
- Created `/pair` command (`internal/commands/pair.go`):
  - Triggers Pair mode via SetModeMsg
  - Alias: `/p`
- Created `/tools` command (`internal/commands/tools.go`):
  - Detects installed developer tools using `internal/tools/detector.go`
  - Groups by category (editors, terminals, VCS, build, containers, AI/LLM)
  - Shows installed/not-found with versions
  - Alias: `/t`
- Created `/config` command (`internal/commands/config.go`):
  - Shows daemon URL, status, version
  - Shows active theme
  - Shows terminal info (size, TERM, TERM_PROGRAM, COLORTERM)
- Created `/project` command (`internal/commands/project.go`):
  - Detects project type from markers (go.mod, Cargo.toml, mix.exs, etc.)
  - Shows module name from go.mod
  - Shows workspace directory
  - Alias: `/proj`
- Wired Pair mode into app:
  - handlePairKey dispatches to pair model
  - renderPairLayout for split/full layout
  - enterMode initializes pair on demand
  - Esc in Pair (when not waiting) returns to Normal
- Registered new commands in registry (pair, tools, config, project)
- Updated main.go help text with new modes and commands

**Build:** `go build ./...` + `go vet ./...` — Clean

**New files:**
```
internal/pair/pair.go              # Pair mode overlay
internal/commands/pair.go          # /pair command
internal/commands/tools.go         # /tools command
internal/commands/config.go        # /config command
internal/commands/project.go       # /project command
```

**Next:** Phase 4 — Streaming polish, /call command, response formatting

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 1)

**Status:** Complete

**Completed:**
- **Architectural pivot from tab-based UI to modal chat interface (vim-inspired)**
- Created theme system (`internal/theme/`):
  - `theme.go` — Theme struct with semantic color mapping
  - `builtin.go` — 3 built-in themes (Hecate Dark, Light, Monochrome)
  - Every component receives `*theme.Theme` for consistent styling
- Created mode state machine (`internal/modes/`):
  - `modes.go` — Mode enum (Normal, Insert, Command, Browse, Pair)
  - Explicit transitions with contextual hints per mode
- Created command system (`internal/commands/`):
  - `command.go` — Command interface + Context struct
  - `registry.go` — Registry with dispatch + Tab autocomplete
  - Built-in commands: /help, /clear, /quit, /status, /health, /models, /model, /me
  - All command output appears as system messages in chat
- Created standalone chat renderer (`internal/chat/`):
  - `chat.go` — Chat as always-visible canvas (not a "view")
  - `styles.go` — Themed styles, welcome art, formatting
  - Public API for mode-driven input visibility, scrolling, streaming
  - System message injection for command output
- Created status bar (`internal/statusbar/`):
  - Mode indicator (color-coded per mode)
  - Active model name, daemon status, contextual key hints
- Created root app model (`internal/app/`):
  - `app.go` — Modal state machine replacing tab-based `ui/app.go`
  - Per-mode key dispatch: i→Insert, /→Command, Esc→Normal
  - j/k scrolling, Ctrl+D/U half-page, g/G top/bottom
  - ? for help, q to quit
  - Command line with / or : prefix (vim-style)
  - Tab autocomplete for commands
- Updated `cmd/hecate-tui/main.go` — Points to new app, v0.2.0

**Layout:**
```
┌──────────────────────────────────────┐
│ 🔥🗝️🔥 Hecate  ·  model  ·  ● daemon │  ← Header
├──────────────────────────────────────┤
│                                      │
│ Chat area (always visible)           │
│                                      │
├──────────────────────────────────────┤
│ (mode-dependent input area)          │
├──────────────────────────────────────┤
│ NORMAL   model  ●  i:chat /:cmd ... │  ← Status bar
└──────────────────────────────────────┘
```

**New Directory Structure:**
```
internal/
├── app/app.go          # Root modal model (replaces ui/app.go)
├── modes/modes.go      # Mode enum + transitions
├── theme/              # Theme system
│   ├── theme.go
│   └── builtin.go
├── chat/               # Chat renderer (standalone)
│   ├── chat.go
│   └── styles.go
├── commands/           # Slash command framework
│   ├── command.go
│   ├── registry.go
│   ├── help.go
│   ├── clear.go
│   ├── quit.go
│   ├── status.go
│   ├── health.go
│   ├── models.go
│   └── me.go
├── statusbar/statusbar.go
├── client/             # (unchanged)
├── llm/                # (unchanged)
├── tools/              # (unchanged)
└── views/              # (preserved, old tab-based views)
```

**Build:** `go build ./...` — Clean, zero errors
**Version:** 0.2.0

**What was preserved (old views still compile):**
- `internal/views/` — All 6 tab-based views still exist and compile
- `internal/ui/app.go` — Old tab-based app still exists
- `internal/client/` — Untouched
- `internal/llm/` — Untouched
- `internal/tools/` — Untouched

**Next:** Phase 3 — Pair Mode, /config, /tools, /project commands

---

### 2026-02-05 Session (Modal Chat Pivot — Phase 2)

**Status:** Complete

**Completed:**
- Created Browse mode overlay (`internal/browse/browse.go`):
  - Navigable capability list with j/k navigation
  - Enter for detail view, / for search/filter, r to refresh
  - Split pane on wide terminals (>=100 cols), full width on narrow
  - Themed using `*theme.Theme`
- Created `/browse` command (`internal/commands/browse.go`):
  - Triggers Browse mode via SetModeMsg
  - Alias: `/b`
- Created `/theme` command (`internal/commands/theme.go`):
  - `/theme list` — shows available themes with active indicator
  - `/theme <name>` — switches theme at runtime
  - SwitchThemeMsg rebuilds all styled components
- Wired Browse mode into app:
  - handleBrowseKey dispatches to browse model
  - renderBrowseLayout for split/full layout
  - Esc in Browse returns to Normal
  - enterMode initializes browse on demand

**Build:** `go build ./...` + `go vet ./...` — Clean

---

### 2026-02-03 Session (Me & Pair View Enhancements)

**Status:** Complete

**Completed:**
- Enhanced `internal/views/me/me.go` with:
  - Settings panel (press 's')
  - ViewMode state machine (Profile/Settings)
  - Profile card with avatar art
  - Stats fetching (capabilities, subscriptions)
  - Settings navigation and toggling
- Created `internal/views/me/styles.go` with dedicated styling
- Enhanced `internal/views/pair/pair.go` with:
  - Proper pairing flow state machine
  - Code display during pairing
  - Polling for confirmation
  - Paired/unpaired/waiting/error states
- Created `internal/views/pair/styles.go` with dedicated styling
- Added pairing client methods:
  - `StartPairing()` - POST /api/pairing/start
  - `GetPairingStatus()` - GET /api/pairing/status
  - `CancelPairing()` - POST /api/pairing/cancel

**Build:** Successful, go vet clean

---

### 2026-02-03 Session (Browse & Monitor Enhancements)

**Status:** Complete

**Completed:**
- Enhanced `internal/views/browse/browse.go` with:
  - Search mode (`/` to activate, live filtering)
  - Detail view (`Enter` to view capability details)
  - Scroll indicator for long lists
  - ViewMode state machine (List/Search/Detail)
- Created `internal/views/browse/styles.go` with dedicated browse styling
- Enhanced `internal/views/monitor/monitor.go` with:
  - Stats cards row (Uptime, Subscriptions, Capabilities, Status)
  - Two-column layout for Daemon/Mesh sections
  - Error state with helpful hints
  - Last refresh timestamp
- Created `internal/views/monitor/styles.go` with dedicated monitor styling
- Added InputSchema/OutputSchema fields to Capability struct

**Build:** Successful, go vet clean

---

### 2026-02-03 Session (Navigation Refactor)

**Status:** Complete

**Completed:**
- Created `internal/views/views.go` — View interface + Tab enum
- Created `internal/views/browse/browse.go` — Capability discovery list
- Created `internal/views/projects/projects.go` — Projects placeholder (phases preview)
- Created `internal/views/monitor/monitor.go` — Daemon health/status view
- Created `internal/views/pair/pair.go` — Pairing flow view
- Created `internal/views/me/me.go` — Identity/profile view
- Updated `internal/views/chat/chat.go` — Added Name(), ShortHelp(), IsStreaming()
- Rewrote `internal/ui/app.go` — New 6-tab navigation with View interface

**New Tab Order:**
```
[1]Chat [2]Browse [3]Projects [4]Monitor [5]Pair [6]Me
```

**View Interface:**
```go
type View interface {
    tea.Model
    Name() string
    ShortHelp() string
    SetSize(width, height int)
    Focus()
    Blur()
}
```

**Build:** Successful, go vet clean

---

### 2026-02-03 Session (Chat View Implementation)

**Status:** Complete

**Completed:**
- Created `internal/llm/types.go` — LLM types (Message, Model, ChatRequest, ChatResponse, etc.)
- Created `internal/llm/stream.go` — SSE/NDJSON stream parser for streaming responses
- Created `internal/client/llm.go` — Client methods:
  - `ListModels()` — GET /api/llm/models
  - `GetLLMHealth()` — GET /api/llm/health
  - `ChatStream()` — POST /api/llm/chat with SSE streaming
  - `Chat()` — POST /api/llm/chat non-streaming
- Created `internal/views/chat/styles.go` — Beautiful chat styling:
  - Extended color palette (purple, emerald, cyan, pink gradients)
  - Message bubbles (user purple, assistant gray, system muted)
  - Model selector with active/inactive states
  - Streaming animation with sparkles
  - Token count and speed display styles
  - Welcome art for empty state
- Created `internal/views/chat/chat.go` — Main chat view:
  - Bubble Tea model with viewport + textarea
  - Model selector (Tab to cycle)
  - Real-time streaming display with thinking animation
  - Token/speed stats after completion
  - Scroll history with ↑↓
  - Cancel streaming with Esc
  - Clear chat with Ctrl+L
- Updated `internal/ui/app.go`:
  - Added TabChat to navigation
  - Integrated chat view model
  - Added 'c' key to jump to chat
  - Updated help text per view

**Features:**
- 🎨 Beautiful message bubbles with role labels
- ✨ Animated streaming indicator with sparkles
- 📊 Token count and tok/s speed display
- 🔄 Model selector with Tab cycling
- 📜 Scrollable message history
- 🗝️ Welcome art for empty chat state

**Build:** Successful, go vet clean

---
