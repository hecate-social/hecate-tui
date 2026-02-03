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
2. `plans/PLAN_DEVELOPER_STUDIO.md` — **The master plan (APPROVED)**

---

## 🎯 Current Focus: Build the TUI

**Skills files come later.** Build the structure first, refine AI guidance iteratively.

---

## ✅ Completed

- [x] Chat View (local LLM) — `b8da1b7`
- [x] Basic navigation (tabs)
- [x] Daemon client
- [x] Endpoint mismatch fix
- [x] Phase 1.1 Navigation refactor — `bae9309`
- [x] Phase 1.2-1.3 Browse & Monitor — `c555ca6`
- [x] Phase 1.4-1.5 Me & Pair — `14b3100`

---

## 🎨 NEW: Chat Welcome Avatar

**Update chat view welcome screen with Hecate ASCII avatar.**

Source: `hecate-social/hecate-artwork/banners/CHAT_AVATAR.md`

Use the **Threshold Guardian** (Option 5):

```go
const hecateAvatar = `
    ╭─╮           ╭─╮
    │█│   ▄███▄   │█│
    │▓│  █▒◉▒◉▒█  │▓│
    ╰┬╯  █▒╰─╯▒█  ╰┬╯
     │  █▒▒▒▒▒▒▒█  │
     │  █▒╭───╮▒█  │
     │  █▒│ ⚷ │▒█  │
     │  █▒╰─┬─╯▒█  │
    ╭┴╮  ▀█▄│▄█▀  ╭┴╮
    ╚═╝     │     ╚═╝
       
       🔥  🗝️  🔥`
```

**Style with Lip Gloss:**
- Avatar/hood: Purple `#7C3AED`
- Eyes: Amber `#F59E0B`  
- Torches: Orange gradient
- Key: Gold `#FCD34D`

Replace the current simple welcome box in `internal/views/chat/chat.go`.

---

## 🔴 Phase 1: Foundation (NOW)

### 1.1 Navigation Refactor

Current tabs are placeholder. Refactor to match plan:

```
[1]Chat [2]Browse [3]Projects [4]Monitor [5]Pair [6]Me
```

**Files:**
```
internal/
├── app/
│   └── app.go             # Main model, tab switching
└── views/
    ├── chat/              # ✅ EXISTS
    ├── browse/            # NEW
    ├── projects/          # NEW
    ├── monitor/           # NEW
    ├── pair/              # NEW (refactor from existing)
    └── me/                # NEW
```

Each view is a Bubble Tea model implementing:
```go
type View interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
    Name() string      // Tab label
    ShortHelp() string // Status bar hint
}
```

---

### 1.2 Browse View (Basic)

Show capabilities from daemon. Start simple.

**Endpoints:**
- `POST /capabilities/discover` — list capabilities

**UI:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ Browse                                                    [2]       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Capabilities on mesh:                                               │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ ● serve_llm/llama3.2        local     llm, chat              │   │
│  │   serve_llm/qwen2.5-coder   local     llm, code              │   │
│  │   weather.forecast          remote    weather, api           │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  [Enter] View details  [/] Search  [r] Refresh                      │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Files:**
```
internal/views/browse/
├── browse.go          # Main model
├── capabilities.go    # Capability list component
└── styles.go
```

---

### 1.3 Monitor View (Basic)

Daemon health and status.

**Endpoints:**
- `GET /health`
- `GET /identity`

**UI:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ Monitor                                                   [4]       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Daemon Status:                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ Status:    ● Running                                         │   │
│  │ Version:   0.1.1                                              │   │
│  │ Uptime:    2h 34m                                             │   │
│  │ Port:      4444                                               │   │
│  │ Identity:  mri:agent:io.macula/hecate-dev                     │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  Mesh Connection:                                                    │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ Status:    ● Connected                                        │   │
│  │ Bootstrap: boot.macula.io:443                                 │   │
│  │ Peers:     3                                                  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Files:**
```
internal/views/monitor/
├── monitor.go
├── daemon.go
└── styles.go
```

---

### 1.4 Me View (Basic)

Identity and basic settings.

**UI:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ Me                                                        [6]       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Identity:                                                           │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ MRI:       mri:agent:io.macula/hecate-dev                     │   │
│  │ Realm:     io.macula                                          │   │
│  │ Paired:    ✅ Yes (since 2026-02-03)                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  [p] Re-pair  [s] Settings                                          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Files:**
```
internal/views/me/
├── me.go
├── profile.go
└── styles.go
```

---

### 1.5 Pair View (Refactor)

Move existing pairing logic into proper view structure.

**Files:**
```
internal/views/pair/
├── pair.go
├── qr.go
└── styles.go
```

---

## 🟡 Phase 2: Projects Shell

After Phase 1, build the Projects view structure.

### 2.1 Project Detection

```
internal/projects/
├── detector.go        # Find projects (git, HECATE.md, etc.)
├── project.go         # Project type/state
└── workspace.go       # .hecate/ management
```

### 2.2 Projects View Shell

```
internal/views/projects/
├── projects.go        # Project list + selection
├── phases.go          # AnD/AnP/InT/DoO tab bar
└── styles.go
```

### 2.3 Phase Placeholder Views

Empty shells that say "Coming soon" — structure first:

```
internal/views/projects/
├── and/
│   └── and.go         # "Analysis & Discovery - Coming Soon"
├── anp/
│   └── anp.go         # "Architecture & Planning - Coming Soon"
├── int/
│   └── int.go         # "Implementation & Testing - Coming Soon"
└── doo/
    └── doo.go         # "Deployment & Operations - Coming Soon"
```

---

## 🟢 Phase 3: Tool Integration

### 3.1 Tool Detection

```
internal/tools/
├── detector.go        # Check which tools are installed
├── config.go          # Load ~/.hecate/config.toml
└── launcher.go        # tea.ExecProcess wrappers
```

### 3.2 Quick Edit

Built-in lightweight editor:

```
internal/editor/
├── editor.go          # textarea-based editor
├── syntax.go          # chroma highlighting
└── styles.go
```

---

## 🟢 Phase 4: Flesh Out Phases

Build actual functionality for each phase. Order TBD based on needs.

---

## Architecture Notes

### View Interface

All views implement:

```go
package views

type View interface {
    tea.Model
    Name() string
    ShortHelp() string
}
```

### Navigation

```go
// Global keybindings (work in any view)
"1" → Chat
"2" → Browse
"3" → Projects
"4" → Monitor
"5" → Pair
"6" → Me
"?" → Help overlay
"q" → Quit (with confirm if unsaved state)
```

### Shared Styles

```
internal/ui/styles/
└── styles.go          # Shared colors, borders, etc.
```

Use consistent Hecate colors:
- Purple: `#7C3AED` (primary)
- Amber: `#F59E0B` (accent)
- Gray scale for text

---

## Test Flow

After Phase 1:

```bash
# Terminal 1
ollama run llama3.2

# Terminal 2  
cd hecate-daemon && rebar3 shell

# Terminal 3
cd hecate-tui && go run ./cmd/hecate-tui

# Should see:
# - 6 tabs: Chat, Browse, Projects, Monitor, Pair, Me
# - Number keys switch tabs
# - Each view shows basic content
```

---

*Build the structure. Refine the soul later.* 🔥🗝️🔥
