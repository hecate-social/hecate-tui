# PLAN: Hecate Developer Studio

**Version:** 1.0
**Date:** 2026-02-03
**Status:** Approved

---

## Vision

Hecate TUI evolves from a monitoring tool into a **full Developer Studio** for building Macula mesh services.

Two integrated experiences:

1. **Mesh Interface** — Chat, browse, pair, monitor (the mesh IS the computer)
2. **Developer Studio** — Project-based, AI-assisted development workflow

---

## Why Hecate? — Differentiators

Current AI coding tools (Claude Code, Cursor, Warp, etc.) are **chat-first**. You talk to AI, it responds with code. Hecate is **workflow-first**. The phases guide you through development. AI assists at each phase. The structure prevents chaos.

### The Problems We Solve

| Problem | Current Tools | Hecate |
|---------|---------------|--------|
| **Context Amnesia** | Each session starts fresh. Re-explain constantly. | Persistent `.hecate/memory/` per project. Phase context survives. |
| **No Architecture** | Generates `services/` folders. Apologizes. Repeats. | **Doctrine Coach** — real-time rules engine catches violations on write. |
| **Skip to Code** | Jump straight to implementation. No design. | **Phased workflow** — AnD → AnP → InT → DoO. Can't skip analysis. |
| **Reactive Only** | Waits for you to notice problems. | Filesystem watcher. Violations flagged immediately. |
| **Single File Focus** | Good at one file. Bad at coordination. | **Slice-aware** — vertical slices are the atomic unit, not files. |
| **Vendor Lock-in** | Tied to Claude/GPT. Data goes to cloud. | **Model agnostic** — Ollama, mesh models, any OpenAI-compatible. Fully local option. |
| **Tool Fragmentation** | Alt-tab between chat, git, deploy, test. | **Integrated** — lazygit, k9s, neovim one keypress away. `:q` returns to TUI. |
| **Generic Instructions** | Same prompts for every project. | **Skills per phase** + **HECATE.md per project**. Learns your patterns. |
| **Black Box** | Generates code. Doesn't explain why. | Reasoning in workflow. AnD/AnP decisions documented in `.hecate/state/`. |
| **Text Only** | Render diagrams elsewhere. | **Structured views** — Kanban, slice trees, event maps. Not just chat. |

### The Core Insight

> **Chat is ONE view. Not THE view.**

Current tools put chat at the center. Everything is a conversation. Hecate puts **workflow** at the center. Chat assists the workflow. Structure guides the chaos.

### Hecate's Unique Value

1. **Architecture-First Development**
   - Doctrine enforcement is automatic, not aspirational
   - HECATE.md defines YOUR patterns, not generic best practices
   - Coach catches violations before they compound

2. **Phased Workflow (AnD → AnP → InT → DoO)**
   - Analysis before architecture
   - Architecture before implementation
   - Implementation before deployment
   - Each phase has specialized AI guidance

3. **Mesh-Native Intelligence**
   - Discover LLM models from the mesh
   - Not just local Ollama — distributed AI
   - The mesh IS the computer

4. **Tool Integration, Not Replacement**
   - Your neovim config. Your lazygit workflows. Your k9s.
   - Hecate orchestrates. Tools execute.
   - One keypress away, `:q` returns home

5. **Project Memory**
   - Corrections remembered, not repeated
   - Domain discoveries persist across sessions
   - Architecture decisions tracked and explained

6. **No Vendor Lock-in**
   - Run fully local with Ollama
   - Or use cloud APIs
   - Or use mesh-discovered models
   - Your choice. Your data.

### The Tagline

> *"The AI coding assistant that understands architecture."*

Or:

> *"Workflow-first development. AI-assisted, not AI-dependent."*

Or:

> *"Chat is a feature. Structure is the product."*

---

The Studio follows **four phases** that mirror the software development lifecycle:

| Phase | Code | Focus |
|-------|------|-------|
| **Analysis & Discovery** | AnD | Event Storming, DDD, domain modeling |
| **Architecture & Planning** | AnP | Vertical slices, Division Architecture, Kanban |
| **Implementation & Testing** | InT | Code generation, doctrine, testing |
| **Deployment & Operations** | DoO | Deploy, publish to mesh, monitor |

Each phase is guided by AI using dedicated **Skills files**.

---

## Core Principle: The Mesh IS the Computer

Local and remote capabilities are equivalent. My node is just one node on the distributed mesh. The TUI treats them uniformly — where something runs is an implementation detail, not a primary concern.

---

## Navigation Structure

```
┌──────────────────────────────────────────────────────────────────────┐
│ [1]Chat [2]Browse [3]Projects [4]Monitor [5]Pair [6]Me               │
└──────────────────────────────────────────────────────────────────────┘

1. Chat      — Talk to AI (mesh LLMs, context-aware)
2. Browse    — Discover capabilities, agents, models
3. Projects  — Developer Studio (AnD → AnP → InT → DoO)
4. Monitor   — Daemon health, my services, logs
5. Pair      — Mesh connection management
6. Me        — Identity, social, permissions, settings
```

---

## View Specifications

### 1. Chat View

**Purpose:** Communicate with LLMs on the mesh.

**Features:**
- Model selector (local + mesh-discovered models)
- Streaming responses with token stats
- Context loading (HECATE.md, project files)
- Conversation persistence

**Status:** ✅ Phase 1 Complete (local chat)

**Pending:**
- [ ] Mesh model discovery
- [ ] HECATE.md context loading
- [ ] Conversation save/load

---

### 2. Browse View

**Purpose:** Discover capabilities, agents, and models on the mesh.

**Sub-views:**

| Sub-view | Description |
|----------|-------------|
| Capabilities | Search/filter available services |
| Agents | Browse agent profiles |
| Models | LLM models specifically |

**Features:**
- Unified search across local and mesh
- Filter by tags, type, rating
- Detail view with test capability
- Actions: Subscribe, Endorse, Test Call

**Key Insight:** No "Local vs Remote" distinction in primary UI. The mesh is one computer.

---

### 3. Projects View (Developer Studio)

**Purpose:** AI-assisted development workflow for Macula services.

**Structure:**
```
Projects View
├── Project List (recent projects, add new)
├── Project Selected
│   ├── [AnD] Analysis & Discovery
│   ├── [AnP] Architecture & Planning
│   ├── [InT] Implementation & Testing
│   └── [DoO] Deployment & Operations
```

Each phase is detailed below.

---

### 4. Monitor View

**Purpose:** Observe daemon and service health.

**Sub-views:**

| Sub-view | Description |
|----------|-------------|
| Daemon | Health, version, uptime, connection |
| Services | My announced capabilities, their status |
| Logs | Tail logs for daemon or specific service |
| Reputation | My ratings, endorsements, disputes |

---

### 5. Pair View

**Purpose:** Manage mesh connection.

**Features:**
- Pairing flow (QR code, confirmation code)
- Connection status (bootstrap nodes, peers)
- Re-pair / Unpair actions
- Multi-realm support (future)

---

### 6. Me View

**Purpose:** Identity, social, permissions, settings.

**Sub-views:**

| Sub-view | Description |
|----------|-------------|
| Profile | MRI, display name, description |
| Social | Followers, following, endorsements |
| UCAN | Granted/received permissions |
| Settings | LLM config, preferences, theme |

---

## Developer Studio — Phase Details

### AnD: Analysis & Discovery

**Purpose:** Domain modeling using Event Storming and DDD practices.

**AI Skills:** `~/.hecate/skills/AnD_SKILLS.md`

**Capabilities:**
- Scan existing codebase for domain events
- Identify aggregates and bounded contexts
- Discover commands and queries
- Visualize domain model
- Chat with AI about domain concepts

**UI Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ AnD: Analysis & Discovery                          [project-name]   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Domain Events:                      Aggregates:                     │
│  ┌────────────────────────────┐     ┌────────────────────────────┐  │
│  │ 🟠 capability_announced    │     │ ▪ capability_aggregate     │  │
│  │ 🟠 capability_retracted    │     │ ▪ identity_aggregate       │  │
│  │ 🟠 agent_paired            │     │ ▪ serve_llm_aggregate      │  │
│  │ 🟠 follower_recorded       │     └────────────────────────────┘  │
│  │ [+ Add Event]              │                                      │
│  └────────────────────────────┘     Commands:                        │
│                                      ┌────────────────────────────┐  │
│  Bounded Contexts:                   │ ▸ announce_capability      │  │
│  ┌────────────────────────────┐     │ ▸ pair_agent               │  │
│  │ ▪ capabilities (manage_*)  │     │ ▸ follow_agent             │  │
│  │ ▪ identity (manage_*)      │     └────────────────────────────┘  │
│  │ ▪ social (manage_social)   │                                      │
│  └────────────────────────────┘                                      │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ [Scan] Analyze codebase  [Chat] Ask AI  [Export] → AnP         │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**Actions:**
- `Scan` — Parse codebase for events, commands, aggregates
- `Chat` — Discuss domain with AI (uses AnD_SKILLS.md)
- `Export` — Carry discovered model to AnP phase
- `Diagram` — Generate Event Storming board (mermaid)

**Outputs:**
- Domain event catalog
- Aggregate map
- Bounded context boundaries
- Exported context for AnP

---

### AnP: Architecture & Planning

**Purpose:** Design vertical slices and plan implementation.

**AI Skills:** `~/.hecate/skills/AnP_SKILLS.md`

**Capabilities:**
- Design Division Architecture vertical slices
- Define desk/supervisor structure
- Generate Kanban task board
- Export tasks to external tools

**UI Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ AnP: Architecture & Planning                       [project-name]   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Vertical Slices:                    Kanban Board:                   │
│  ┌────────────────────────────┐     ┌──────┬───────┬──────────────┐ │
│  │ serve_llm/                 │     │ TODO │ DOING │ DONE         │ │
│  │ ├── announce_llm_cap/      │     ├──────┼───────┼──────────────┤ │
│  │ │   ├── command            │     │ ▪ P3 │ ▪ P2  │ ▪ P1 API     │ │
│  │ │   ├── event              │     │ ▪ P4 │       │ ▪ Tests      │ │
│  │ │   ├── handler            │     │      │       │ ▪ Docs       │ │
│  │ │   └── emitter            │     │      │       │              │ │
│  │ ├── retract_llm_cap/       │     └──────┴───────┴──────────────┘ │
│  │ └── handle_llm_rpc/        │                                      │
│  │                            │     Task Details:                    │
│  │ [+ Add Slice]              │     ┌────────────────────────────┐  │
│  └────────────────────────────┘     │ P2: Mesh announcement      │  │
│                                      │ Slice: announce_llm_cap/   │  │
│  Mesh Integration:                   │ Files: 4  Tests: 0         │  │
│  ┌────────────────────────────┐     └────────────────────────────┘  │
│  │ ▪ Emitters (EVENT → FACT)  │                                      │
│  │ ▪ Listeners (FACT → CMD)   │                                      │
│  │ ▪ Responders (RPC)         │                                      │
│  └────────────────────────────┘                                      │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ [Generate] Scaffold  [Chat] Ask AI  [Export] → taskwarrior     │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**Actions:**
- `Generate` — Create directory structure and boilerplate
- `Chat` — Discuss architecture with AI (uses AnP_SKILLS.md)
- `Export` — Sync tasks to taskwarrior or GitHub Issues
- `Diagram` — Generate architecture diagram (mermaid)

**Division Architecture Patterns:**
- CMD slices (command → event → handler → aggregate)
- QRY slices (queries on projections)
- Projections (event → read model)
- Mesh integration (emitters, listeners, responders)

**Outputs:**
- Slice directory structure
- Kanban task list
- Architecture diagrams
- Generated boilerplate (→ InT)

---

### InT: Implementation & Testing

**Purpose:** Code generation, doctrine enforcement, testing.

**AI Skills:** `~/.hecate/skills/InT_SKILLS.md`

**Capabilities:**
- Generate Division Architecture code from templates
- Real-time doctrine violation detection
- Test generation and execution
- Integration with external editors

**UI Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ InT: Implementation & Testing                      [project-name]   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Current Slice: serve_llm/announce_llm_capability/                   │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ Status │ File                              │ Type     │ Notes │  │
│  ├────────┼───────────────────────────────────┼──────────┼───────┤  │
│  │   ✅   │ announce_llm_capability_v1.erl    │ command  │       │  │
│  │   ✅   │ llm_capability_announced_v1.erl   │ event    │       │  │
│  │   ⚠️   │ maybe_announce_llm_capability.erl │ handler  │ 2 TODO│  │
│  │   ✅   │ llm_capability_announced_to_mesh  │ emitter  │       │  │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  Doctrine Coach:                     Tests:                          │
│  ┌────────────────────────────┐     ┌────────────────────────────┐  │
│  │ Violations: 0              │     │ Total:    14               │  │
│  │ Warnings:   1              │     │ Passing:  14  ✅           │  │
│  │ ─────────────────────────  │     │ Failing:  0               │  │
│  │ ⚠️ Nested case in handler  │     │ Coverage: 78%             │  │
│  └────────────────────────────┘     └────────────────────────────┘  │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ [Edit] neovim  [Git] lazygit  [Test] Run  [Chat] Ask AI        │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**Actions:**
- `Edit` — Open file in neovim (or configured editor)
- `Git` — Launch lazygit for version control
- `Test` — Run tests for current slice
- `Chat` — Get AI help (uses InT_SKILLS.md)
- `Generate` — Create file from template

**Doctrine Coach (Built-in):**
Real-time filesystem watcher that detects violations:

| Rule | Detection |
|------|-----------|
| Horizontal directories | Path regex: `/services/`, `/helpers/`, `/utils/` |
| Central supervisors | Path regex: `_listeners_sup.erl`, `_handlers_sup.erl` |
| CRUD events | Content regex: `_created_v`, `_updated_v`, `_deleted_v` |
| God modules | Path regex: `_manager.erl` |

Violations shown inline with explanation and suggested fix.

**External Tool Integration:**
- **neovim** — Code editing
- **lazygit** — Git operations
- **rebar3/mix/go** — Build and test

**Outputs:**
- Implemented code
- Passing tests
- Clean doctrine report
- Ready for deployment (→ DoO)

---

### DoO: Deployment & Operations

**Purpose:** Deploy to local/cluster/mesh, announce capabilities, monitor.

**AI Skills:** `~/.hecate/skills/DoO_SKILLS.md`

**Capabilities:**
- Pre-flight checks (compile, dialyzer, tests)
- Deploy to multiple targets
- Announce capabilities to mesh
- Monitor deployed services

**UI Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ DoO: Deployment & Operations                       [project-name]   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Deploy Target:                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  ○ Local (this machine)                                      │   │
│  │  ● Cluster (beam00-03.lab)                                   │   │
│  │  ○ Container (Docker/Podman)                                 │   │
│  │  ○ Kubernetes (k3s cluster)                                  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  Pre-flight Checks:                  Capabilities to Announce:       │
│  ┌────────────────────────────┐     ┌────────────────────────────┐  │
│  │ ✅ Compiles cleanly        │     │ ☑ serve_llm                │  │
│  │ ✅ Dialyzer passes         │     │   ├─ llama3.2             │  │
│  │ ✅ Tests pass (14/14)      │     │   ├─ qwen2.5-coder        │  │
│  │ ✅ No doctrine violations  │     │   └─ deepseek-r1          │  │
│  │ ⚠️ 2 TODOs remaining       │     │ ☐ query_capabilities      │  │
│  └────────────────────────────┘     │ ☐ manage_social           │  │
│                                      └────────────────────────────┘  │
│  Announcement Preview:                                               │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ MRI: mri:capability:io.macula/hecate-dev/llm/llama3.2        │   │
│  │ Type: llm                                                     │   │
│  │ Tags: [ai, chat, llm, local]                                  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │ [Deploy] Execute  [Docker] lazydocker  [K8s] k9s  [Chat] AI    │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**Actions:**
- `Deploy` — Execute deployment to selected target
- `Docker` — Launch lazydocker for container management
- `K8s` — Launch k9s for Kubernetes management
- `Announce` — Publish capabilities to mesh
- `Chat` — Get AI help (uses DoO_SKILLS.md)

**Deployment Targets:**
1. **Local** — Run on this machine
2. **Cluster** — SSH to beam cluster nodes
3. **Container** — Build and run Docker image
4. **Kubernetes** — Deploy to k3s/k8s cluster

**Post-deployment:**
- Automatic capability announcement (if selected)
- Health check verification
- Link to Monitor view for ongoing observation

---

## External Tool Integration

### Tool Categories

The Developer Studio integrates with external TUI tools. Users choose their preferred tool per category.

#### 🔍 Search & Find
| Tool | Description | Phase |
|------|-------------|-------|
| **fzf** | Fuzzy finder (recommended) | All |
| **ripgrep (rg)** | Fast grep | AnD |
| **fd** | Fast find | AnD |

#### 📝 Editor
| Tool | Description | Phase |
|------|-------------|-------|
| **neovim** | Extensible vim (recommended) | InT |
| **helix** | Post-modern, modal | InT |
| **micro** | Simple, intuitive | InT |
| **kakoune** | Selection-based | InT |

#### 📂 File Manager
| Tool | Description | Phase |
|------|-------------|-------|
| **yazi** | Blazing fast, async (recommended) | All |
| **ranger** | Vim-style, classic | All |
| **lf** | Lightweight, Go-based | All |
| **superfile** | Fancy, modern | All |
| **broot** | Tree navigation | All |

#### 🔀 Git Interface
| Tool | Description | Phase |
|------|-------------|-------|
| **lazygit** | Simple, powerful (recommended) | InT |
| **tig** | Classic, ncurses | InT |
| **gitui** | Rust-based, fast | InT |
| **serie** | Rich commit graph | InT |

#### 🐳 Containers
| Tool | Description | Phase |
|------|-------------|-------|
| **lazydocker** | Docker TUI (recommended) | DoO |
| **oxker** | Lightweight | DoO |
| **dry** | Docker manager | DoO |
| **dive** | Image layer explorer | DoO |

#### ☸️ Kubernetes
| Tool | Description | Phase |
|------|-------------|-------|
| **k9s** | K8s TUI (recommended) | DoO |
| **kdash** | Dashboard | DoO |
| **kubetui** | Monitoring focused | DoO |

#### 🗄️ Database
| Tool | Description | Phase |
|------|-------------|-------|
| **lazysql** | Multi-DB client | InT |
| **harlequin** | SQL IDE | InT |
| **rainfrog** | Postgres/MySQL/SQLite | InT |
| **dblab** | Database browser | InT |

#### 🌐 API/HTTP
| Tool | Description | Phase |
|------|-------------|-------|
| **posting** | HTTP client TUI | InT |
| **ATAC** | Full API client | InT |

#### 📊 JSON/Data
| Tool | Description | Phase |
|------|-------------|-------|
| **fx** | JSON viewer/processor | InT |
| **jqp** | jq playground | InT |
| **visidata** | Data exploration | AnD |

#### 📋 Logs
| Tool | Description | Phase |
|------|-------------|-------|
| **lazyjournal** | journalctl TUI | DoO |
| **nerdlog** | Multi-host logs | DoO |

#### 📈 System Monitor
| Tool | Description | Phase |
|------|-------------|-------|
| **btop** | Resource monitor (recommended) | DoO |
| **bottom** | Customizable | DoO |
| **glances** | Cross-platform | DoO |

#### 🔧 Process Manager
| Tool | Description | Phase |
|------|-------------|-------|
| **process-compose** | Multi-process runner | DoO |

#### 🐙 GitHub
| Tool | Description | Phase |
|------|-------------|-------|
| **gh-dash** | PR/issue dashboard | InT |

---

### Install-Time Selection (Workstation Role)

When `workstation` role is selected during install, show interactive checklist:

```
━━━ Developer Tools (optional) ━━━

The Developer Studio integrates with these tools.
Select which to install (space to toggle, enter to confirm):

  🔍 Search & Find:
    [x] fzf              fuzzy finder (recommended)
    [ ] ripgrep (rg)     fast grep
    [ ] fd               fast find

  📝 Editor:
    [x] neovim           (recommended)
    [ ] helix            post-modern editor

  📂 File Manager:
    [x] yazi             blazing fast (recommended)
    [ ] ranger           vim-style classic

  🔀 Git Interface:
    [x] lazygit          (recommended)
    [ ] tig              classic

  🐳 Containers:
    [x] lazydocker       (recommended)

  ☸️  Kubernetes:
    [x] k9s              (recommended)

  ... (more categories)

  [ ] Skip all - I'll configure in TUI later
```

---

### TUI Settings Configuration

Users can change tool preferences in `Me → Settings → Developer Tools`:

```
━━━ Developer Tools ━━━

Configure which tools the Studio launches.
Blank = disabled. "custom" = use custom command.

┌──────────────────┬────────────────┬─────────────────────────────┐
│  Category        │  Tool          │  Status                     │
├──────────────────┼────────────────┼─────────────────────────────┤
│  Search/Fuzzy    │  [fzf       ]  │  ✅ installed               │
│  Search/Grep     │  [rg        ]  │  ✅ installed               │
│  Search/Find     │  [fd        ]  │  ✅ installed               │
├──────────────────┼────────────────┼─────────────────────────────┤
│  Editor          │  [nvim      ]  │  ✅ installed               │
│  Files           │  [yazi      ]  │  ❌ not found               │
│  Git             │  [lazygit   ]  │  ✅ installed               │
├──────────────────┼────────────────┼─────────────────────────────┤
│  Containers      │  [lazydocker]  │  ✅ installed               │
│  Kubernetes      │  [k9s       ]  │  ✅ installed               │
│  Database        │  [lazysql   ]  │  ❌ not found               │
├──────────────────┼────────────────┼─────────────────────────────┤
│  API             │  [posting   ]  │  ❌ not found               │
│  JSON            │  [fx        ]  │  ✅ installed               │
│  Logs            │  [          ]  │  (disabled)                 │
│  Monitor         │  [btop      ]  │  ✅ installed               │
│  GitHub          │  [gh-dash   ]  │  ❌ not found               │
└──────────────────┴────────────────┴─────────────────────────────┘

  [Install Missing] via package manager    [Verify All] check paths
```

---

### Configuration File

```toml
# ~/.hecate/config.toml

[tools.search]
fuzzy = "fzf"           # fzf, skim, custom
grep = "rg"             # rg, grep, ag, custom
find = "fd"             # fd, find, custom

[tools.editor]
default = "nvim"        # nvim, hx, micro, code, custom
diff = "delta"          # delta, diff, custom

[tools.files]
manager = "yazi"        # yazi, ranger, lf, nnn, superfile, broot, custom

[tools.git]
client = "lazygit"      # lazygit, tig, gitui, custom
graph = "serie"         # serie, git-log, custom

[tools.containers]
docker = "lazydocker"   # lazydocker, oxker, dry, custom
images = "dive"         # dive, custom

[tools.kubernetes]
client = "k9s"          # k9s, kdash, kubetui, custom

[tools.database]
client = "lazysql"      # lazysql, harlequin, rainfrog, dblab, custom

[tools.api]
client = "posting"      # posting, ATAC, custom

[tools.data]
json = "fx"             # fx, jq, custom
playground = "jqp"      # jqp, play, custom
tables = "visidata"     # visidata, custom

[tools.logs]
viewer = "lazyjournal"  # lazyjournal, nerdlog, custom

[tools.monitor]
system = "btop"         # btop, bottom, htop, glances, custom

[tools.github]
dashboard = "gh-dash"   # gh-dash, custom

[tools.custom]
# Override any tool with custom command
editor = ""             # e.g., "emacsclient -t"
```

---

### Integration Approach

Tools are launched externally (not embedded). Hecate TUI:
1. Detects if tool is installed
2. Provides keybinding to launch
3. **Returns focus to TUI after tool exits** (`:q` in neovim → back to TUI)

```go
// Example: Launch configured editor
func launchEditor(cfg *config.Tools, filepath string) tea.Cmd {
    tool := cfg.Editor.Default // "nvim", "hx", etc.
    if tool == "" || tool == "disabled" {
        return nil
    }
    cmd := exec.Command(tool, filepath)
    return tea.ExecProcess(cmd, func(err error) tea.Msg {
        return editorExitMsg{err: err, file: filepath}
    })
}
```

---

### Editor Integration: Full vs Quick Edit

Two modes for editing files:

#### Full Edit (External Editor)

Launch user's configured editor (neovim, helix, etc.):
- Full editor experience with user's config/plugins
- `:q` / `Ctrl+Q` returns to TUI
- Used for: implementation work, complex edits

```
Keybinding: [e] Edit in neovim
            [E] Edit in $EDITOR (fallback)
```

#### Quick Edit (Inline)

Built-in lightweight editor for small changes:
- Single-file, basic editing
- No external process
- Used for: config tweaks, template fills, commit messages

```
┌─────────────────────────────────────────────────────────────────────┐
│ Quick Edit: announce_llm_capability_v1.erl                          │
├─────────────────────────────────────────────────────────────────────┤
│   1 │ -module(announce_llm_capability_v1).                          │
│   2 │ -export([new/3, to_map/1, from_map/1]).                       │
│   3 │                                                                │
│   4 │ -record(announce_llm_capability_v1, {                         │
│   5 │     model_name,                                                │
│   6 │     agent_identity,█                                          │
│   7 │     metadata                                                   │
│   8 │ }).                                                            │
│   9 │                                                                │
├─────────────────────────────────────────────────────────────────────┤
│ [Ctrl+S] Save  [Ctrl+Q] Cancel  [Ctrl+G] Go to line                 │
└─────────────────────────────────────────────────────────────────────┘
```

**Implementation:** Use `github.com/charmbracelet/bubbles/textarea` with line numbers and syntax highlighting (via `github.com/alecthomas/chroma`).

```go
// Quick edit for simple changes
type QuickEditModel struct {
    textarea textarea.Model
    filepath string
    modified bool
}

// Full edit launches external editor
func (m Model) fullEdit(filepath string) tea.Cmd {
    editor := m.cfg.Tools.Editor.Default
    if editor == "" {
        editor = os.Getenv("EDITOR")
    }
    if editor == "" {
        editor = "vi" // fallback
    }
    return tea.ExecProcess(
        exec.Command(editor, filepath),
        func(err error) tea.Msg { return editorExitMsg{err, filepath} },
    )
}
```

#### When to Use Which

| Scenario | Mode | Keybinding |
|----------|------|------------|
| Implement a full slice | Full Edit | `e` |
| Fill in a template placeholder | Quick Edit | `q` |
| Edit config file | Quick Edit | `q` |
| Write commit message | Quick Edit | (auto) |
| Complex refactoring | Full Edit | `e` |
| View-only with small fix | Quick Edit | `q` |

```go
// Example: Launch configured git tool
func launchGitTool(cfg *config.Tools, projectDir string) tea.Cmd {
    tool := cfg.Git.Client // "lazygit", "tig", etc.
    if tool == "" || tool == "disabled" {
        return nil
    }
    return tea.ExecProcess(
        exec.Command(tool),
        func(err error) tea.Msg { return toolExitMsg{err} },
    )
}
```

### Tool Detection

```go
type ToolStatus struct {
    Configured string // what user configured
    Installed  bool   // whether it exists on PATH
    Path       string // resolved path
}

type ToolAvailability struct {
    Search struct {
        Fuzzy ToolStatus
        Grep  ToolStatus
        Find  ToolStatus
    }
    Editor    ToolStatus
    Files     ToolStatus
    Git       ToolStatus
    Containers ToolStatus
    Kubernetes ToolStatus
    Database  ToolStatus
    API       ToolStatus
    JSON      ToolStatus
    Logs      ToolStatus
    Monitor   ToolStatus
    GitHub    ToolStatus
}

func detectTools(cfg *config.Tools) ToolAvailability {
    return ToolAvailability{
        Editor: checkTool(cfg.Editor.Default),
        Git:    checkTool(cfg.Git.Client),
        // ... etc
    }
}

func checkTool(name string) ToolStatus {
    if name == "" || name == "disabled" {
        return ToolStatus{Configured: name, Installed: false}
    }
    path, err := exec.LookPath(name)
    return ToolStatus{
        Configured: name,
        Installed:  err == nil,
        Path:       path,
    }
}
```

---

## Skills Files

AI guidance for each phase lives in Skills files:

```
~/.hecate/skills/
├── AnD_SKILLS.md     # Analysis & Discovery
├── AnP_SKILLS.md     # Architecture & Planning
├── InT_SKILLS.md     # Implementation & Testing
└── DoO_SKILLS.md     # Deployment & Operations
```

### ⚠️ Skills = Quality

**Each skills file is a separate project.** These determine the quality of AI assistance in each phase. They deserve focused attention and iteration.

| Skills File | Focus | Status |
|-------------|-------|--------|
| `AnD_SKILLS.md` | Event Storming, DDD, domain discovery | 📋 TODO |
| `AnP_SKILLS.md` | Division Architecture Architecture, vertical slices, Kanban | 📋 TODO |
| `InT_SKILLS.md` | Code generation, doctrine enforcement, testing | 📋 TODO |
| `DoO_SKILLS.md` | Deployment, mesh publishing, monitoring | 📋 TODO |

**Will be developed separately with dedicated planning.**

### Skills File Structure (Template)

Each file follows this template:

```markdown
# [Phase] Skills

## Context
What this phase is about.

## Patterns
Specific patterns the AI should follow.

## Templates
Code/document templates for this phase.

## Checklist
What must be complete before moving to next phase.

## Anti-patterns
What to avoid.

## Examples
Worked examples demonstrating correct approach.
```

Skills files are:
- Shipped with hecate-install installer
- User-customizable
- Loaded as context when Chat is invoked in that phase
- Version-controlled in `hecate-social/hecate-skills` repo (future)

---

## Project Detection

### Detection Logic

A directory is recognized as a project using this priority:

| Priority | Signal | Meaning |
|----------|--------|---------|
| 1 | `HECATE.md` | Explicit Hecate project with AI instructions |
| 2 | `.hecate/` | Has Hecate workspace (config, memory) |
| 3 | `rebar.config` / `mix.exs` / `go.mod` / `Cargo.toml` / `package.json` | Language project |
| 4 | `.git/` | Any git repository (fallback) |

### Philosophy: Low Friction, Rich Enhancement

**Any git repo is a project.** No explicit opt-in required. Users can browse and work with any repository.

**`HECATE.md` unlocks richer AI context.** When present, the AI loads project-specific instructions, patterns, and constraints.

### Auto-Create `.hecate/` on First Use

When a user opens a project in Studio for the first time:

```
┌─────────────────────────────────────────────────────────────────────┐
│ Initialize Hecate Workspace?                                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  This project doesn't have a Hecate workspace yet.                   │
│                                                                      │
│  Creating .hecate/ will enable:                                      │
│    • Project-specific settings                                       │
│    • Conversation memory                                             │
│    • AnD/AnP phase state                                             │
│                                                                      │
│  Also create HECATE.md for AI context?                               │
│    [x] Yes, create HECATE.md with template                           │
│                                                                      │
│              [Create]  [Skip for now]  [Never ask]                   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Workspace Structure

```
project/
├── HECATE.md              # AI instructions (optional but recommended)
├── .hecate/
│   ├── config.toml        # Project-specific settings
│   ├── memory/            # Conversation history per phase
│   │   ├── and.md         # AnD phase notes
│   │   ├── anp.md         # AnP phase notes
│   │   └── int.md         # InT phase notes
│   ├── state/
│   │   ├── events.json    # Discovered domain events (AnD)
│   │   ├── slices.json    # Designed slices (AnP)
│   │   └── kanban.json    # Task board state
│   └── .gitignore         # Ignore memory, keep config
└── ...
```

### HECATE.md Template

When user opts to create `HECATE.md`:

```markdown
# Project Name

Brief description of what this project does.

## Architecture

Key patterns and conventions used in this project.

## Commands

Common commands for building, testing, deploying.

## Guidelines

Do's and don'ts for AI assistance in this project.

## Context

Any additional context the AI should know.
```

### Detection Implementation

```go
type Project struct {
    Path         string
    Name         string
    Type         ProjectType  // Hecate, Language, Git
    HasHecateMD  bool
    HasWorkspace bool
    Language     string       // erlang, elixir, go, rust, etc.
}

type ProjectType int

const (
    ProjectTypeGit ProjectType = iota
    ProjectTypeLanguage
    ProjectTypeHecate
)

func detectProject(dir string) (*Project, error) {
    p := &Project{Path: dir, Name: filepath.Base(dir)}
    
    // Check for HECATE.md (highest priority)
    if fileExists(filepath.Join(dir, "HECATE.md")) {
        p.Type = ProjectTypeHecate
        p.HasHecateMD = true
    }
    
    // Check for .hecate/ workspace
    if dirExists(filepath.Join(dir, ".hecate")) {
        p.HasWorkspace = true
        if p.Type != ProjectTypeHecate {
            p.Type = ProjectTypeHecate
        }
    }
    
    // Check for language markers
    if p.Type != ProjectTypeHecate {
        if lang := detectLanguage(dir); lang != "" {
            p.Type = ProjectTypeLanguage
            p.Language = lang
        }
    }
    
    // Fallback to git
    if p.Type == 0 && dirExists(filepath.Join(dir, ".git")) {
        p.Type = ProjectTypeGit
    }
    
    return p, nil
}

func detectLanguage(dir string) string {
    markers := map[string]string{
        "rebar.config":  "erlang",
        "mix.exs":       "elixir",
        "go.mod":        "go",
        "Cargo.toml":    "rust",
        "package.json":  "javascript",
        "pyproject.toml": "python",
        "Makefile":      "make",
    }
    for file, lang := range markers {
        if fileExists(filepath.Join(dir, file)) {
            return lang
        }
    }
    return ""
}
```

---

## File Structure

```
internal/views/
├── chat/                    # Chat with mesh LLMs
│   ├── chat.go
│   ├── styles.go
│   ├── model_selector.go
│   └── context_loader.go
│
├── browse/                  # Discover mesh capabilities
│   ├── browse.go
│   ├── capabilities.go
│   ├── agents.go
│   ├── models.go
│   └── test_call.go
│
├── projects/                # Developer Studio
│   ├── projects.go          # Project list/selection
│   ├── detector.go          # Project detection logic
│   │
│   ├── and/                 # Analysis & Discovery
│   │   ├── and.go
│   │   ├── scanner.go       # Codebase scanner
│   │   ├── events.go        # Event discovery UI
│   │   ├── aggregates.go    # Aggregate visualization
│   │   └── contexts.go      # Bounded contexts
│   │
│   ├── anp/                 # Architecture & Planning
│   │   ├── anp.go
│   │   ├── slices.go        # Slice designer
│   │   ├── kanban.go        # Task board
│   │   ├── generator.go     # Scaffold generator
│   │   └── export.go        # Taskwarrior/GH export
│   │
│   ├── int/                 # Implementation & Testing
│   │   ├── int.go
│   │   ├── files.go         # File checklist
│   │   ├── coach.go         # Doctrine enforcer
│   │   ├── templates.go     # Code templates
│   │   └── tests.go         # Test runner
│   │
│   └── doo/                 # Deployment & Operations
│       ├── doo.go
│       ├── preflight.go     # Pre-flight checks
│       ├── deploy.go        # Deployment execution
│       ├── announce.go      # Capability announcement
│       └── targets.go       # Deploy target config
│
├── monitor/                 # Daemon & service health
│   ├── monitor.go
│   ├── daemon.go
│   ├── services.go
│   ├── logs.go
│   └── reputation.go
│
├── pair/                    # Mesh connection
│   ├── pair.go
│   ├── qr.go
│   └── status.go
│
└── me/                      # Identity & settings
    ├── me.go
    ├── profile.go
    ├── social.go
    ├── ucan.go
    └── settings.go
```

---

## Implementation Phases

### Phase 1: Foundation ✅ (Partial)
- [x] Chat view (local LLM)
- [ ] Browse view (basic capability list)
- [ ] Monitor view (daemon health)
- [ ] Me view (identity display)
- [ ] Pair view (pairing flow)

### Phase 2: Projects Shell
- [ ] Project list/detection
- [ ] Project selection
- [ ] Phase navigation (AnD/AnP/InT/DoO tabs)
- [ ] Tool detection

### Phase 3: AnD — Analysis & Discovery
- [ ] Codebase scanner
- [ ] Event/aggregate discovery
- [ ] Domain visualization
- [ ] AI chat integration

### Phase 4: AnP — Architecture & Planning
- [ ] Slice designer
- [ ] Kanban board
- [ ] Scaffold generator
- [ ] Task export

### Phase 5: InT — Implementation & Testing
- [ ] File checklist
- [ ] Doctrine coach (filesystem watcher)
- [ ] Editor integration (neovim)
- [ ] Test runner

### Phase 6: DoO — Deployment & Operations
- [ ] Pre-flight checks
- [ ] Multi-target deployment
- [ ] Capability announcement
- [ ] Container/K8s integration

### Phase 7: Polish
- [ ] Mesh model discovery (Chat)
- [ ] Full Browse functionality
- [ ] Social features (Me)
- [ ] UCAN management (Me)

---

## Resolved Questions

1. **Sidecar installation** ✅
   - Workstation role shows interactive checklist of suggested tools
   - User selects which to install (opinionated TUI users choose their own)
   - TUI Settings allows changing tools later
   - Config stored in `~/.hecate/config.toml`

2. **Neovim integration** ✅
   - **Full Edit:** Launch externally, `:q` returns to TUI
   - **Quick Edit:** Built-in lightweight editor for small changes
   - Keybindings: `[e]` full edit, `[q]` quick edit

3. **Skills files** ✅
   - Each is a separate project (AnD, AnP, InT, DoO)
   - Will be developed with dedicated planning
   - Quality of AI assistance depends on these

4. **Project detection** ✅
   - Any git repo is a project (low friction)
   - `HECATE.md` unlocks richer AI context
   - Auto-create `.hecate/` on first Studio use (with confirmation)

---

## Design Principles

1. **Task-based, not data-based** — Views are activities, not data types
2. **Screaming architecture** — Folder names describe what users DO
3. **AI-assisted, not AI-dependent** — Guidance helps, manual override always available
4. **Tools compose** — Integrate external tools, don't reinvent them
5. **Offline-capable** — Core features work without mesh connection
6. **The mesh is the computer** — Local/remote is an implementation detail

---

*The goddess guides developers through the crossroads of creation.* 🔥🗝️🔥
