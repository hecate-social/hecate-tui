# PLAN: Hecate Studio UX

The Hecate TUI evolves from a simple monitoring tool into a full **Development Studio** for building Macula services.

## Vision

Three pillars:

1. **Social** — Browse the mesh community (agents, capabilities, reputation)
2. **Workflow** — AI-guided service development (Discover → Architect → Implement → Deploy)
3. **Coach** — Agent trainer with doctrine enforcement ("Hecate on the Node")

Plus: **Identity** — Profile, pairing status, settings.

---

## Navigation Structure

```
[1] Social     - Browse agents, capabilities, local vs remote services
[2] Discover   - Find gaps in the mesh, identify opportunities
[3] Architect  - Cartwheel scaffolding for new services
[4] Implement  - Guided coding with templates and AI assistance
[5] Deploy     - Pre-flight checks, announce to mesh
[6] Coach      - Agent training, doctrine enforcement, live monitoring
[7] Identity   - Profile, pairing, daemon status, settings
```

---

## 1. Social View

### Purpose
Browse the Hecate Social community — agents, capabilities, reputation.

### Sub-tabs
- **Local Services** — Capabilities I provide (my daemon)
- **Remote Services** — Capabilities discovered from mesh
- **Agents** — Browse agent profiles
- **Following** — Agents I follow

### Local vs Remote Distinction

| Aspect | Local Service | Remote Service |
|--------|---------------|----------------|
| Ownership | I own it | Someone else owns it |
| Actions | Edit, Pause, Deprecate | Test, Subscribe, Endorse, Report |
| Metrics | Full logs, detailed stats | Public reputation only |
| Config | Editable | Read-only |

### Local Service List
- Service name and MRI
- Status indicator (LIVE, PAUSED, ERROR)
- Calls today / total
- Star rating
- Quick actions: View, Pause, Edit

### Local Service Detail
- Identity: Agent MRI, registration date, procedure, tags
- Reputation: Stars, endorsements, disputes
- Metrics: Calls, latency (avg/p99), success rate, uptime
- Recent activity log
- Configuration (editable)
- Actions: Edit config, Pause, Deprecate, View logs

### Remote Service List
- Service name and MRI
- Provider agent
- Star rating and call count
- Tags
- Quick actions: View, Test Call, Subscribe

### Remote Service Detail
- Provider: Agent MRI, owner, realm, since date
- Capability: Procedure, tags, description
- Reputation: Stars, endorsements, disputes, total calls
- My interaction: Subscribed?, My calls, My rating, Endorsed?
- Test call input (JSON editor)
- Actions: Test, Subscribe, Rate, Endorse, Report

### Agent List
- Agent name and MRI
- Capability count
- Calls served
- Star rating
- Actions: View, Follow

### Agent Detail
- Profile info
- Capabilities provided
- Reputation summary
- Badges earned
- Actions: Follow, View capabilities

---

## 2. Discover View

### Purpose
AI-assisted gap analysis — find opportunities for new services.

### Flow
1. User describes what they want to build (free text input)
2. TUI queries mesh for related capabilities
3. AI analyzes gaps and opportunities
4. Presents recommendation

### Sections
- **Input**: Text field for service idea
- **Mesh Analysis**: Related capabilities found, their ratings and usage
- **Gap Analysis**: What's missing, opportunities identified
- **Recommendation**: Suggested service scope and differentiation

### Actions
- Continue to Architect (with context carried forward)
- Refine search
- View related capability details

---

## 3. Architect View

### Purpose
Generate Cartwheel Architecture scaffolding for the service.

### Context
Receives service concept from Discover (or manual entry).

### Sections
- **Service metadata**: Name, realm, description
- **Domain structure**: Visual tree of planned slices

### Domain Structure Display

```
service_name/
├── CMD SLICES (write path)
│   ├── command_one/
│   │   ├── command_one_v1.erl        (command)
│   │   ├── event_happened_v1.erl     (event)
│   │   ├── maybe_command_one.erl     (handler)
│   │   └── aggregate.erl             (aggregate)
│   └── command_two/
│       └── ...
├── QRY SLICES (read path)
│   ├── query_one/
│   │   └── query_one.erl             (queries projection)
│   └── query_two/
│       └── ...
├── PROJECTIONS
│   ├── projection_one.erl            (events → read model)
│   └── projection_two.erl
└── MESH INTEGRATION
    ├── fact_listener.erl             (FACT → CMD)
    ├── event_emitter.erl             (EVENT → FACT)
    └── rpc_responder.erl             (HOPE → CMD → FEEDBACK)
```

### Important
- CMD/QRY/PRJ are **domain slices** (business logic)
- LISTENER/EMITTER/RESPONDER are **mesh components** (infrastructure within domains)
- Listeners are NOT parallel to CMD slices — they live inside domains and feed into CMD slices

### Actions
- Generate scaffold (creates directory structure and boilerplate)
- Edit structure (add/remove slices)
- Explain pattern (AI explains Cartwheel concepts)

---

## 4. Implement View

### Purpose
Guided coding with templates, progress tracking, and AI assistance.

### Sections
- **Current file**: Which file is being worked on
- **Template pane**: Boilerplate code for current file type
- **Guidance pane**: Explanation of the pattern being implemented
- **Progress tracker**: Checklist of files in current slice

### File Types and Templates
- Command modules: `new/N`, `to_map/1`, `from_map/1`
- Event modules: Same pattern as commands
- Handler modules: `handle/1`, `dispatch/1`
- Aggregate modules: State management
- Query modules: `execute/N`
- Projection modules: Event handlers that update read models
- Listener modules: FACT → CMD dispatch
- Emitter modules: EVENT → FACT publish
- Responder modules: HOPE → CMD → FEEDBACK

### Actions
- Copy template to clipboard
- Next file in sequence
- Ask AI for help
- Run tests for current slice

---

## 5. Deploy View

### Purpose
Pre-flight checks and mesh deployment.

### Pre-flight Checks
- Compiles cleanly
- Dialyzer passes
- Tests pass
- No doctrine violations detected
- Event schemas versioned

### Capability Announcement Preview
- MRI to be announced
- Procedure name
- Tags
- Description

### Deployment Target
- Local daemon only
- Mesh realm (with realm selection)

### Actions
- Deploy now
- Preview announcement
- View logs

---

## 6. Coach View

### Purpose
Agent training, doctrine enforcement, live monitoring. "Hecate on the Node."

### Agent Selection
- Select which AI agent to coach (claude-code, etc.)
- Display session duration and correction count

### Doctrine Memory
Progress bar showing internalization level.

Per-doctrine status:
- Vertical slicing — correct applications count
- CMD/Event/Handler pattern — understanding level
- Spoke supervisors — learned/needs reinforcement
- FACTS ≠ EVENTS — status
- Horizontal temptation — violation count

### Live Monitor
Filesystem watcher on project directory.

Event log showing:
- Files/directories created
- Doctrine violations detected (with alerts)
- Timestamps

### Intervention System
When violation detected:
- Alert popup with explanation
- What the agent did wrong
- What they should do instead
- Actions: Send correction to agent, Delete offending file/dir, Dismiss

### History
- Session history
- Correction log
- Patterns of mistakes

### Memory Report
- Export agent's doctrine understanding
- Identify areas needing reinforcement

---

## 7. Identity View

### Purpose
Profile, pairing status, daemon health, settings.

### Sections

#### My Identity
- Agent MRI
- Display name
- Description (editable)
- Created date
- Public key fingerprint

#### Pairing Status
- Paired realm(s)
- Pairing date
- Certificate expiry
- Actions: Pair new realm, Unpair

#### Daemon Status
- Running/stopped
- Version
- Uptime
- Port
- Data directory
- Actions: Restart, View logs

#### Mesh Connection
- Connected/disconnected
- Bootstrap nodes
- Peers count
- Last sync

#### Settings
- Default realm
- Notification preferences
- Coach strictness level
- Theme (dark/light)

---

## Technical Notes

### State Management
Each view maintains its own state. Navigation preserves state within session.

### Data Sources
- **Local**: Daemon HTTP API at `:4444`
- **Remote**: Mesh queries via daemon
- **Coach**: Filesystem watcher + Git integration

### Keyboard Navigation
- Number keys (1-7) jump to views
- Tab cycles through panes
- Arrow keys navigate lists
- Enter selects/expands
- Esc goes back
- ? shows help

### Responsive Layout
- Minimum terminal size: 80x24
- Adapts to larger terminals
- Panes resize proportionally

---

## Implementation Phases

### Phase 1: Foundation
- Navigation framework (7 tabs)
- Identity view (daemon status, basic profile)
- Social view — Local services list and detail

### Phase 2: Discovery
- Social view — Remote services list and detail
- Discover view — Gap analysis with AI

### Phase 3: Workflow
- Architect view — Cartwheel scaffolding
- Implement view — Templates and guidance

### Phase 4: Operations
- Deploy view — Pre-flight and deployment
- Social view — Full agent browsing

### Phase 5: Coach
- Coach view — Live monitoring
- Filesystem watcher integration
- Intervention system

---

## Design Principles

1. **Information density**: Show relevant data without clutter
2. **Progressive disclosure**: Summary → Detail on demand
3. **Keyboard-first**: Everything accessible via keyboard
4. **Offline-capable**: Core features work without mesh connection
5. **AI-assisted, not AI-dependent**: Guidance helps, but manual override always available

---

## Coach Architecture: Rules-First, AI-Optional

### Key Distinction

**Macula Agents/Services are NOT AI-powered.** They are traditional business process services — deterministic, testable, reliable. Weather forecasting, data aggregation, payment processing, whatever. No LLM in the runtime.

**LLMs are development tools**, not runtime dependencies. We use AI to:
- Generate code scaffolding
- Write documentation
- Create SVG diagrams
- Explain architecture concepts

The services themselves are pure business logic.

### Coach Implementation

The Coach enforces Cartwheel doctrine. Most of this is **rule-based, not AI-powered**:

| Function | Implementation | LLM Required |
|----------|----------------|--------------|
| Detect violations | Pattern matching on paths/content | No |
| Identify `services/`, `helpers/`, `utils/` | Regex on directory names | No |
| Catch central supervisors | AST analysis or naming patterns | No |
| Generate correction message | Templated responses | No |
| Cartwheel scaffolding | Code templates with variable substitution | No |
| Explain "why is this wrong?" | Small local model OR curated FAQ | Optional |
| Dynamic code generation | Larger model helps | Optional |
| Documentation generation | Model-assisted | Optional |

### Rules Engine (Core)

The Coach core is a rules engine:

```
Rule: horizontal_directory
Match: path contains /services/ OR /helpers/ OR /utils/ OR /handlers/
Action: alert "Horizontal directory detected. Each {type} belongs to its domain."

Rule: central_supervisor
Match: filename matches *_listeners_sup.erl OR *_handlers_sup.erl
Action: alert "Central supervisor detected. Each domain supervises its own."

Rule: crud_event
Match: event name contains _created OR _updated OR _deleted
Action: alert "CRUD event detected. Use business-meaningful event names."
```

No LLM needed. Pure pattern matching.

### Optional AI Integration

For users who want AI-enhanced guidance:

**Supported providers (user brings their own):**
- Ollama (local, free)
- Anthropic API (Claude)
- OpenAI API (GPT)
- Any OpenAI-compatible endpoint

**Use cases:**
- Richer explanations of violations
- Code generation beyond templates
- Documentation drafting
- Architecture Q&A

**Configuration:**
```
~/.hecate/config.toml

[coach]
mode = "rules"  # or "ai-enhanced"

[coach.ai]
provider = "ollama"  # or "anthropic", "openai"
model = "llama3:8b"
endpoint = "http://localhost:11434"
# api_key = "..." (for cloud providers)
```

### Summary

- **Services on the mesh**: Pure business logic, no AI
- **Development tooling**: AI-assisted, AI-optional
- **Coach core**: Rules engine, works offline, no dependencies
- **Coach enhanced**: Bring your own model for richer experience

The daemon stays lean. The AI is a plugin, not a requirement.

---

*This document defines the UX vision. Implementation details live in code.*

🔥🗝️🔥
