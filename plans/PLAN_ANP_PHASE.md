# Plan: Architecture & Planning (AnP) Phase

*The second phase of the Developer Studio workflow.*

---

## Overview

AnP transforms the validated domain model from AnD into executable structure and project plan. If AnD was done properly, this phase is largely mechanical — the architecture writes itself from the artifacts.

**Philosophy:** Architecture is not invention, it's translation. The commands dictate the structure. The budget is arithmetic.

---

## Inputs (from AnD)

| Artifact | Contains |
|----------|----------|
| `CONTEXT_MAP.yaml` | Bounded contexts, relationships |
| `{context}/DOMAIN.yaml` | Aggregates, commands, events, policies |

---

## Workflow

### Step 1: Scaffold Generation

**For each context in CONTEXT_MAP.yaml:**

```
apps/{context}/
├── src/
│   ├── {context}_app.erl
│   ├── {context}_sup.erl
│   ├── {context}_store.erl          # ReckonDB instance
│   └── {aggregate}_aggregate.erl    # Aggregate module
```

**For each command in DOMAIN.yaml:**

```
apps/{context}/src/{command}/
├── {command}_v1.erl                 # Command record
├── {event}_v1.erl                   # Domain event record
├── maybe_{command}.erl              # Handler (validate, dispatch)
└── {event}_v1_to_mesh.erl           # Emitter (if integrates externally)
```

**For each policy in DOMAIN.yaml:**

```
apps/{context}/src/{policy}/
├── {policy}_sup.erl                 # Desk supervisor
└── {policy}.erl                     # Listener + dispatcher
```

**Output:** Generated directory structure with stub files.

---

### Step 2: Estimation

Each command gets a t-shirt size:

| Size | Criteria | Effort |
|------|----------|--------|
| **S** | Simple command, one event, no policy, no integration | 2h |
| **M** | Validation logic, triggers policy, or updates projection | 4h |
| **L** | External integration, multiple events, complex invariants | 1d |
| **XL** | Saga/orchestration, compensation flows, multiple aggregates | 2d |

**AI-Assisted Sizing:**
- Parse `DOMAIN.yaml`
- Count preconditions → adds complexity
- Check if command triggers policies → +size
- Check if event has mesh emitter → +size
- Suggest size, developer confirms/adjusts

**Output:** `ESTIMATE.yaml`

```yaml
# ESTIMATE.yaml
context: loan_origination
generated: 2026-02-04T02:30:00Z

commands:
  - name: initialize_loan_request
    size: S
    effort_hours: 2
    mvp: true
    notes: "Simple initialization, no validation"
    
  - name: approve_loan
    size: M
    effort_hours: 4
    mvp: true
    notes: "Has preconditions, triggers notification policy"
    
  - name: reject_loan
    size: M
    effort_hours: 4
    mvp: true
    notes: "Has preconditions, triggers notification policy"

summary:
  total_commands: 3
  mvp_commands: 3
  total_hours: 10
  mvp_hours: 10
  
  by_size:
    S: 1
    M: 2
    L: 0
    XL: 0
```

---

### Step 3: Sequencing

Determine implementation order based on:

1. **Dependencies** — Which commands must exist before others?
2. **MVP priority** — MVP commands first
3. **Integration needs** — Contexts that publish facts before contexts that consume them

**Output:** `SEQUENCE.yaml`

```yaml
# SEQUENCE.yaml
phases:
  - phase: 1
    name: "Core Origination"
    commands:
      - initialize_loan_request
    rationale: "Entry point, no dependencies"
    
  - phase: 2
    name: "Decision Flow"
    commands:
      - approve_loan
      - reject_loan
    rationale: "Depends on initialized request"
    dependencies: [initialize_loan_request]
```

---

### Step 4: API Contracts (if integrating)

For contexts that exchange facts, define the schema:

**Output:** `CONTRACTS.yaml`

```yaml
# CONTRACTS.yaml
facts:
  - name: credit_score_received_v1
    publisher: credit_assessment
    subscribers: [loan_origination]
    schema:
      request_id: string
      score: integer
      provider: string
      checked_at: datetime
    guarantees:
      - "Score is between 300-850"
      - "Delivered within 30s of request"
```

This is the *only* coupling between contexts. No shared code. Schema is the contract.

---

## Output Artifacts

| Artifact | Purpose |
|----------|---------|
| Generated directories | Scaffold with stub files |
| `ESTIMATE.yaml` | Time/effort per command |
| `SEQUENCE.yaml` | Implementation order |
| `CONTRACTS.yaml` | Fact schemas between contexts |

---

## TUI Implementation Notes

### Views Required

1. **Scaffold Preview** — Tree view of what will be generated
2. **Estimate Editor** — List of commands with size selector
3. **Sequence Planner** — Drag/reorder or phase assignment
4. **Contract Editor** — Fact schema definition

### Key Interactions

- `g` — Generate scaffolds (with confirmation)
- `s` — Set size for selected command (S/M/L/XL)
- `↑↓` — Reorder sequence
- `p` — Assign to phase
- `Enter` — Edit contract schema

### AI Integration

- "Suggest sizes based on complexity"
- "What's blocking this command?"
- "Generate contract for this fact"

---

## Scaffold Templates

The generator uses templates for each file type:

### Command Template (`{command}_v1.erl`)
```erlang
-module({command}_v1).
-export([new/1, to_map/1, from_map/1]).
-export([{field}/1 || field <- fields]).  %% Getters

-record({command}_v1, {
    %% Fields from DOMAIN.yaml payload
}).

new(Params) ->
    #?MODULE{
        %% Map params to record fields
    }.

to_map(#?MODULE{} = Cmd) ->
    #{}.

from_map(Map) ->
    #?MODULE{}.
```

### Handler Template (`maybe_{command}.erl`)
```erlang
-module(maybe_{command}).
-export([handle/1, dispatch/1]).

handle(#{{command}_v1{} = Cmd) ->
    %% Validate preconditions from DOMAIN.yaml
    case validate(Cmd) of
        ok -> {ok, create_event(Cmd)};
        {error, _} = Err -> Err
    end.

dispatch(Cmd) ->
    case handle(Cmd) of
        {ok, Event} ->
            evoq:dispatch({context}_store, Event);
        Error ->
            Error
    end.

validate(_Cmd) ->
    %% TODO: Implement preconditions
    ok.

create_event(Cmd) ->
    %% TODO: Build event from command
    ok.
```

### Emitter Template (`{event}_v1_to_mesh.erl`)
```erlang
-module({event}_v1_to_mesh).
-behaviour(gen_server).
%% ... standard gen_server boilerplate

init([]) ->
    reckon:subscribe({context}_store, ?MODULE, self()),
    {ok, #{}}.

handle_info({event, #{type := {event}_v1} = Event}, State) ->
    Fact = to_mesh_fact(Event),
    hecate_mesh_publisher:publish(Fact),
    {noreply, State}.

to_mesh_fact(Event) ->
    %% Transform domain event to integration fact
    #{}.
```

---

## Budget Summary View

```
┌─────────────────────────────────────────────────┐
│  LOAN ORIGINATION - BUDGET SUMMARY              │
├─────────────────────────────────────────────────┤
│                                                 │
│  Commands:  12 total (8 MVP, 4 extra)           │
│                                                 │
│  MVP Effort:                                    │
│    S (2h) × 3  =   6h                           │
│    M (4h) × 4  =  16h                           │
│    L (1d) × 1  =   8h                           │
│    ─────────────────                            │
│    Total:         30h (~4 days)                 │
│                                                 │
│  Extra:                                         │
│    M (4h) × 2  =   8h                           │
│    L (1d) × 2  =  16h                           │
│    ─────────────────                            │
│    Total:         24h (~3 days)                 │
│                                                 │
│  Full Scope:      54h (~7 days)                 │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## Transition to InT

When AnP is complete, the developer has:
- ✅ Generated scaffold structure
- ✅ Sized and estimated all commands
- ✅ Sequenced implementation phases
- ✅ Defined integration contracts

The **Implementation & Testing (InT)** phase:
- Fills in the stub files with actual logic
- Runs tests per slice
- Validates against contracts

*AnD discovers WHAT. AnP plans HOW. InT builds IT.*

---

*The architecture is not debated. It is derived.* 🔥🗝️🔥
