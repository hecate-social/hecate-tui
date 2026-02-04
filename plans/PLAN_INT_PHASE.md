# Plan: Implementation & Testing (InT) Phase

*The third phase of the Developer Studio workflow.*

---

## Overview

InT follows a **tests-first, docs-first, inside-out** approach. The domain model from AnD/AnP becomes the specification. AI generates tests and documentation upfront, then implements code to satisfy them.

**Philosophy:** Write what it should do before writing how it does it. Test the core first, expand outward.

---

## Inputs (from AnP)

| Artifact | Contains |
|----------|----------|
| Generated scaffolds | Vertical slice structure |
| `DOMAIN.yaml` | Aggregates, commands, events, policies |
| `CONTRACTS.yaml` | Fact schemas between contexts |
| `ESTIMATE.yaml` | Sized commands for tracking |

---

## The Inside-Out Testing Pyramid

```
┌─────────────────────────────────────────────────┐
│  3. INTEGRATION (outermost) — LAST             │
│     • Fact exchange between contexts            │
│     • Mesh communication (publish/subscribe)    │
│     • Store persistence (ReckonDB, SQLite)      │
│     • External service calls                    │
├─────────────────────────────────────────────────┤
│  2. DOMAIN INTERACTION — SECOND                 │
│     • Command → Handler → Event flow            │
│     • Policy: listens Event → dispatches Cmd    │
│     • Projection: Event → Read Model update     │
│     • Aggregate command validation              │
├─────────────────────────────────────────────────┤
│  1. STATE (innermost) — FIRST                   │
│     • Aggregate state transitions               │
│     • Given state + event = new state           │
│     • Pure functions, no side effects           │
│     • Value object validation                   │
└─────────────────────────────────────────────────┘
```

---

## Workflow

### Step 1: Generate Documentation Stubs

**From `DOMAIN.yaml`, generate:**

#### Context README
```markdown
# Loan Origination

Handles loan request lifecycle from application to decision.

## External API

### Responders (Incoming HOPES)
| Procedure | Input | Output | Description |
|-----------|-------|--------|-------------|
| `loan.apply` | ApplyRequest | LoanId | Submit new loan application |
| `loan.status` | LoanId | LoanStatus | Query loan status |

### Listeners (Incoming FACTS)
| Fact | Source | Action |
|------|--------|--------|
| `credit_score_received_v1` | credit_assessment | Updates risk profile |

## Diagrams
- [Context Diagram](./diagrams/context.svg)
- [Container Diagram](./diagrams/container.svg)
- [Event Flow](./diagrams/event-flow.svg)
```

#### C4 Context Diagram (SVG generation)
```
┌─────────────────┐
│   Applicant     │
│   [Person]      │
└────────┬────────┘
         │ applies for loan
         ▼
┌─────────────────┐      FACT: credit_score      ┌─────────────────┐
│ Loan Origination│◄─────────────────────────────│Credit Assessment│
│   [Context]     │                              │   [Context]     │
└────────┬────────┘                              └─────────────────┘
         │ 
         │ FACT: loan_approved
         ▼
┌─────────────────┐
│   Fulfillment   │
│   [Context]     │
└─────────────────┘
```

#### AsyncAPI Spec (for external API)
```yaml
asyncapi: 2.6.0
info:
  title: Loan Origination API
  version: 1.0.0
channels:
  loan/apply:
    publish:
      message:
        $ref: '#/components/messages/ApplyRequest'
    subscribe:
      message:
        $ref: '#/components/messages/LoanId'
```

---

### Step 2: Generate Test Stubs (Inside-Out)

#### Layer 1: State Tests (FIRST)

Test aggregate state transitions. Pure functions.

```erlang
%% loan_request_state_tests.erl
-module(loan_request_state_tests).
-include_lib("eunit/include/eunit.hrl").

%% Given: no state, When: initialized, Then: pending state
initialize_creates_pending_state_test() ->
    Event = loan_request_initialized_v1:new(#{
        id => <<"loan-001">>,
        applicant => <<"John Doe">>,
        amount => 50000
    }),
    State = loan_request_aggregate:apply_event(undefined, Event),
    ?assertEqual(pending, State#loan_request.status).

%% Given: pending, When: approved, Then: approved state
approve_transitions_to_approved_test() ->
    State = #loan_request{status = pending, credit_score = 720},
    Event = loan_approved_v1:new(#{approved_amount => 50000}),
    NewState = loan_request_aggregate:apply_event(State, Event),
    ?assertEqual(approved, NewState#loan_request.status).

%% Given: pending, When: rejected, Then: rejected state  
reject_transitions_to_rejected_test() ->
    State = #loan_request{status = pending},
    Event = loan_rejected_v1:new(#{reason => <<"Low credit">>}),
    NewState = loan_request_aggregate:apply_event(State, Event),
    ?assertEqual(rejected, NewState#loan_request.status).
```

#### Layer 2: Domain Interaction Tests (SECOND)

Test command handlers and policies.

```erlang
%% maybe_approve_loan_tests.erl
-module(maybe_approve_loan_tests).
-include_lib("eunit/include/eunit.hrl").

%% Command validation: valid command produces event
valid_approval_produces_event_test() ->
    Cmd = approve_loan_v1:new(#{
        loan_id => <<"loan-001">>,
        approved_amount => 50000,
        approved_by => <<"underwriter-1">>
    }),
    %% Mock aggregate state
    State = #loan_request{status = pending, credit_score = 720},
    {ok, Event} = maybe_approve_loan:handle(Cmd, State),
    ?assertEqual(loan_approved_v1, element(1, Event)).

%% Command validation: low credit score rejected
low_credit_rejected_test() ->
    Cmd = approve_loan_v1:new(#{loan_id => <<"loan-001">>}),
    State = #loan_request{status = pending, credit_score = 520},
    {error, credit_score_too_low} = maybe_approve_loan:handle(Cmd, State).

%% Policy: auto-reject on low score
auto_reject_policy_test() ->
    Event = credit_score_received_v1:new(#{
        loan_id => <<"loan-001">>,
        score => 520
    }),
    {dispatch, Cmd} = auto_reject_low_credit:on_event(Event),
    ?assertEqual(reject_loan_v1, element(1, Cmd)).
```

#### Layer 3: Integration Tests (LAST)

Test fact exchange and infrastructure.

```erlang
%% loan_origination_integration_tests.erl
-module(loan_origination_integration_tests).
-include_lib("eunit/include/eunit.hrl").

%% Fact received from credit_assessment triggers policy
credit_fact_integration_test() ->
    %% Simulate incoming mesh fact
    Fact = #{
        type => <<"credit_score_received_v1">>,
        payload => #{loan_id => <<"loan-001">>, score => 750}
    },
    %% Listener receives and dispatches
    ok = credit_score_listener:handle_fact(Fact),
    %% Verify command was dispatched
    ?assert(was_command_dispatched(update_credit_score_v1)).

%% Event published to mesh as fact
approval_published_to_mesh_test() ->
    Event = loan_approved_v1:new(#{loan_id => <<"loan-001">>}),
    %% Emitter transforms and publishes
    ok = loan_approved_v1_to_mesh:publish(Event),
    %% Verify fact on mesh
    ?assert(mesh_received_fact(<<"loan_approved_v1">>)).
```

---

### Step 3: Implement to Green

AI generates implementation code to satisfy the tests.

**Order:**
1. Value objects (validation, equality)
2. Aggregate state transitions (`apply_event/2`)
3. Command records and getters
4. Event records and getters
5. Handlers (`handle/1`, `handle/2`)
6. Policies (listeners + dispatch)
7. Emitters (event → mesh fact)
8. Projections (event → read model)

**AI workflow:**
```
User: "Implement loan_request_aggregate to pass state tests"

AI: Looking at the state tests, I need:
- #loan_request record with status, credit_score, amount fields
- apply_event/2 that pattern matches each event type
- State transitions: undefined→pending, pending→approved, pending→rejected

Generating...

[Shows code, applies on confirmation]

Running tests... 3/3 passing ✓
```

---

### Step 4: Documentation Finalization

After implementation, finalize docs:

1. **Update diagrams** — Actual event flows from code
2. **Generate API reference** — From module attributes
3. **Verify README** — Matches implementation

---

## Output Artifacts

| Artifact | Purpose |
|----------|---------|
| `README.md` | Context overview, external API |
| `diagrams/*.svg` | C4 Context, Container, Event Flow |
| `test/*_tests.erl` | Inside-out test suites |
| Implemented modules | Passing all tests |
| `PROGRESS.yaml` | Completion tracking |

---

## Documentation Structure

```
apps/{context}/
├── README.md                    # Context overview + external API
├── diagrams/
│   ├── context.svg              # C4 Context (who uses this)
│   ├── container.svg            # C4 Container (what's inside)
│   └── event-flow.svg           # Event/command flow
├── src/
│   └── {slices}/                # Code screams, no extra docs
└── test/
    ├── state/                   # Layer 1: state tests
    ├── domain/                  # Layer 2: domain tests  
    └── integration/             # Layer 3: integration tests
```

**Internal code needs no documentation** — the vertical slice structure and screaming names ARE the documentation.

---

## TUI Implementation Notes

### Views Required

1. **Test Runner** — Run tests by layer, show results
2. **Doc Viewer** — Render README, view diagrams
3. **Coverage Map** — Which slices have tests, which pass
4. **AI Chat** — Context-aware implementation assistance

### Key Interactions

- `1/2/3` — Run layer 1/2/3 tests
- `t` — Run all tests for slice
- `T` — Run all tests for context
- `d` — View/edit documentation
- `g` — Generate (tests/docs/impl)
- `c` — Chat with AI about current slice

### Progress Display

```
┌─────────────────────────────────────────────────┐
│  LOAN ORIGINATION — InT Progress                │
├─────────────────────────────────────────────────┤
│                                                 │
│  Tests:     ██████████░░ 24/30 (80%)            │
│  ├─ State:  ████████████ 12/12 ✓                │
│  ├─ Domain: ████████░░░░  8/12                  │
│  └─ Integ:  ████░░░░░░░░  4/6                   │
│                                                 │
│  Docs:      ████████████ Complete ✓             │
│                                                 │
│  Slices:    ████████░░░░ 8/12 implemented       │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## Transition to DoO

When InT is complete:
- ✅ All tests passing (inside-out)
- ✅ Documentation complete (README, diagrams, API)
- ✅ All slices implemented
- ✅ Code screams its intent

The **Deployment & Operations (DoO)** phase ships it.

*AnD discovers WHAT. AnP plans HOW. InT builds & tests IT.*

---

*Tests first. Docs first. Inside out.* 🔥🗝️🔥
