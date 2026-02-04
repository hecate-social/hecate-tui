# Plan: Analysis & Discovery (AnD) Phase

*The first phase of the Developer Studio workflow.*

---

## Overview

The AnD phase guides developers through domain discovery using Event Storming principles, AI-assisted refinement, and structured artifact generation. The goal: transform a high-level process description into a validated domain model ready for architecture.

**Philosophy:** Discovery is conversational. The AI acts as a Socratic facilitator — asking questions, suggesting patterns, challenging assumptions — not as an oracle dictating solutions.

---

## Color Legend

| Color | Element | Symbol | Description |
|-------|---------|--------|-------------|
| 🟠 Orange | Domain Event | `event` | Something that happened (past tense verb) |
| 🔵 Blue | Command | `command` | Intent to change state (imperative verb) |
| 🟣 Purple | Integration Fact | `fact` | External system event (incoming from other contexts) |
| 🟢 Green | UI Fact | `ui` | User action that triggers flow (e.g., `form_submitted`) |
| 🟡 Yellow | Aggregate | `aggregate` | Consistency boundary / "dossier" that owns commands |
| ⬜ Grey | Policy | `policy` | Listens → gates → dispatches (process manager) |
| 🔴 Red | Hot Spot | `hotspot` | Confusion, risk, exception path, needs discussion |

---

## Workflow

### Phase 1: Process Brief

**Input:** Natural language description (200-500 words)

**Contents:**
- What is the process? (high-level purpose)
- Who are the actors? (human roles, external systems)
- What are the boundaries? (what's in/out of scope)
- What triggers the process? (entry points)
- What are the outcomes? (success states, failure states)

**AI Role:**
- Ask clarifying questions if brief is ambiguous
- Identify missing actors or boundaries
- Suggest related processes that might integrate

**Output:** `PROCESS_BRIEF.md`

---

### Phase 2: Domain Event Discovery

**Input:** Process brief from Phase 1

**AI Generates:**
- Initial list of domain events (🟠 orange)
- Suggested sub-processes (potential bounded contexts)
- Integration points (🟣 purple facts from outside)
- UI triggers (🟢 green facts from users)
- Hot spots (🔴 red) — areas needing clarification

**Interactive Refinement:**
- Groom events (add, remove, rename, reorder)
- Discuss timeline ordering (before/after relationships)
- Capture "why" for each event (importance, MVP vs extra)
- Identify exception paths and alerts

**Ordering via Policies:**
- Grey stickers (⬜) represent policies
- Arrow from event → policy implies ordering
- Policy may dispatch command that causes next event
- This captures causality without explicit sequence numbers

**Output:** `EVENTS.yaml`

```yaml
# EVENTS.yaml
process: loan_application
version: 1
events:
  - name: loan_request_initialized_v1
    type: event
    color: orange
    description: New loan request dossier created
    why: Entry point for all loan processing
    mvp: true
    triggers: []  # UI fact or integration fact that caused this
    
  - name: credit_check_requested_v1
    type: event
    color: orange
    description: External credit check initiated
    why: Required for risk assessment
    mvp: true
    triggers:
      - event: loan_request_initialized_v1
        via_policy: initiate_credit_check
        
  - name: credit_score_received_v1
    type: fact
    color: purple
    description: Credit bureau responded with score
    source: external/credit_bureau
    why: Determines approval threshold
    mvp: true

policies:
  - name: initiate_credit_check
    type: policy
    color: grey
    listens_to: loan_request_initialized_v1
    dispatches: request_credit_check  # command
    gate: "loan amount > $10,000"
    
hotspots:
  - description: "What if credit bureau is unavailable?"
    related_events: [credit_check_requested_v1]
    resolution: null  # TBD
```

---

### Phase 3: Context Map (Big Picture)

**Input:** Events from Phase 2

**AI Analysis:**
- Cluster events by cohesion (what changes together)
- Identify bounded context boundaries
- Map integration facts to source contexts
- Suggest context relationships

**Output:** `CONTEXT_MAP.yaml`

```yaml
# CONTEXT_MAP.yaml
process: loan_application
contexts:
  - name: loan_origination
    description: Handles initial loan request and decision
    events:
      - loan_request_initialized_v1
      - loan_approved_v1
      - loan_rejected_v1
    commands:
      - initialize_loan_request
      - approve_loan
      - reject_loan
    integrates_with:
      - context: credit_assessment
        receives: [credit_score_received_v1]
        sends: [credit_check_requested_v1]
        
  - name: credit_assessment
    description: Manages credit checks and risk scoring
    events:
      - credit_check_requested_v1
      - credit_score_calculated_v1
    external_dependencies:
      - credit_bureau (purple facts)

relationships:
  - upstream: credit_assessment
    downstream: loan_origination
    type: customer_supplier  # or conformist, ACL, etc.
```

**Context = Process = Microservice = OTP App**

This is "World-Level Slicing" — the macro structure before diving into command-level slices.

---

### Phase 4: Context Refinement

**Input:** One context from the Context Map

**AI-Assisted Deep Dive:**

#### 4a. Aggregate Discovery (🟡 Yellow)

The aggregate is the "dossier" — the consistency boundary that:
- Accepts commands
- Emits events  
- Maintains invariants

**Initialization Pattern:**
Every aggregate starts with `{aggregate}_initialized_v1`:
```yaml
aggregate: loan_request
initialized_by: loan_request_initialized_v1
payload:
  id: LoanRequestId  # Value Object
  applicant: Applicant  # Entity or VO
  amount: Money  # Value Object
  term_months: PositiveInteger  # Value Object
  status: LoanStatus  # Value Object (enum)
```

**Class Diagram Elements:**
- Aggregate Root (with ID as Value Object)
- Entities (have identity, mutable)
- Value Objects (no identity, immutable)
- Invariants (business rules the aggregate enforces)

#### 4b. Command Discovery (🔵 Blue)

For each event, identify the command that caused it:

| Event | Command | Actor |
|-------|---------|-------|
| `loan_request_initialized_v1` | `initialize_loan_request` | Applicant (via UI) |
| `loan_approved_v1` | `approve_loan` | Underwriter |
| `loan_rejected_v1` | `reject_loan` | Underwriter / Policy |

Commands are imperative verbs. Events are past-tense verbs.

#### 4c. Policy Refinement (⬜ Grey)

Formalize policies discovered in Phase 2:

```yaml
policies:
  - name: auto_reject_low_credit
    listens_to: credit_score_received_v1
    gate: "score < 580"
    dispatches: reject_loan
    reason: "Automatic rejection for credit score below threshold"
    
  - name: escalate_high_value
    listens_to: loan_request_initialized_v1
    gate: "amount > $500,000"
    dispatches: request_manual_review
    reason: "High-value loans require senior underwriter"
```

#### 4d. MVP Determination

Mark each element:
- **MVP:** Required for first release
- **Extra:** Deferred to later iteration

**Output:** `{context}/DOMAIN.yaml`

```yaml
# loan_origination/DOMAIN.yaml
context: loan_origination
aggregate:
  name: loan_request
  root: LoanRequestId
  entities: []
  value_objects:
    - LoanRequestId
    - Applicant
    - Money
    - LoanStatus
  invariants:
    - "Amount must be positive"
    - "Term must be 12, 24, 36, 48, or 60 months"

commands:
  - name: initialize_loan_request
    actor: applicant
    mvp: true
    payload: {applicant, amount, term_months}
    emits: loan_request_initialized_v1
    
  - name: approve_loan
    actor: underwriter
    mvp: true
    payload: {loan_id, approved_amount, rate}
    emits: loan_approved_v1
    preconditions:
      - "credit_score >= 580"
      - "status == pending_decision"

events:
  - name: loan_request_initialized_v1
    mvp: true
    payload: {full dossier structure}
    
  - name: loan_approved_v1
    mvp: true
    payload: {loan_id, approved_amount, rate, approved_by, approved_at}

policies:
  - name: auto_reject_low_credit
    mvp: true
    listens_to: credit_score_received_v1
    dispatches: reject_loan
```

---

## Artifact Summary

| Phase | Artifact | Purpose |
|-------|----------|---------|
| 1 | `PROCESS_BRIEF.md` | High-level context, actors, boundaries |
| 2 | `EVENTS.yaml` | Discovered events, policies, ordering, hotspots |
| 3 | `CONTEXT_MAP.yaml` | Bounded contexts, relationships, integration points |
| 4 | `{context}/DOMAIN.yaml` | Aggregate, commands, events, policies per context |

---

## AI Behavior Guidelines

### Session Continuity
- Context must survive across sessions and time
- Store working state in `.hecate/state/and/`
- Allow resumption: "Continue where we left off"

### Socratic Facilitation
- Ask clarifying questions, don't assume
- Challenge: "What happens if X fails?"
- Suggest: "Have you considered Y?"
- Never dictate — propose and validate

### Pattern Recognition
- Identify common patterns: Saga, Choreography, Orchestration
- Suggest: "This looks like a compensation flow"
- But stay in discovery — implementation is AnP phase

### Hotspot Management
- Track unresolved questions
- Prompt resolution before advancing phases
- Allow "park it for later" with explicit marker

---

## Key Patterns

### 1. Initialization Event
Every aggregate begins with `{aggregate}_initialized_v1`:
- Contains full dossier structure
- Is the "birth certificate" of the aggregate
- All subsequent events are deltas

### 2. Facts Between Contexts
- Contexts share **nothing** — no code, no schemas
- Communication is via **facts only** (purple)
- Subscriber interprets the fact autonomously
- Publisher doesn't know/care who subscribes

### 3. UI Facts as Triggers
- Green facts represent user actions: `form_submitted`, `button_clicked`
- These trigger commands, not events directly
- Flow: UI Fact → Command → Domain Event

### 4. Policy as Glue
- Policies connect events to commands
- They encode business rules ("if X then Y")
- They are the "horizontal" connectors between vertical slices

---

## Transition to AnP

When AnD is complete, the developer has:
- ✅ Validated domain model
- ✅ Clear bounded contexts
- ✅ Defined aggregates with structure
- ✅ Commands and events mapped
- ✅ Policies formalized
- ✅ MVP scope determined

The **Architecture & Planning (AnP)** phase takes these artifacts and:
- Generates vertical slice scaffolds
- Plans technical implementation
- Designs API contracts
- Structures the codebase

*AnD discovers WHAT. AnP plans HOW.*

---

## TUI Implementation Notes

### Views Required
1. **Brief Editor** — Markdown editor for process brief
2. **Event Canvas** — List/board view of events with colors
3. **Context Map** — Visual or list representation of contexts
4. **Domain Editor** — Aggregate/command/event editing per context

### Key Interactions
- `Tab` — Cycle through phases
- `/` — Search/filter events
- `Enter` — Edit selected element
- `c` — Add command to selected event
- `p` — Add policy
- `h` — Mark as hotspot
- `m` — Toggle MVP flag

### AI Chat Integration
- Side panel or split view
- AI responds to: "What events might follow X?"
- AI suggests: "You haven't defined what triggers Y"
- Context-aware based on current phase and selection

---

*The goddess guides through the fog of discovery.* 🔥🗝️🔥
