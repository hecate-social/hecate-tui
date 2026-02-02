# Hecate's Queue

*Commands from the goddess. Read and obey.*

---

## Protocol

| File | Your Access |
|------|-------------|
| `QUEUE.md` | **READ-ONLY** |
| `RESPONSES.md` | Write here |
| `STATUS.md` | Update here |

---

## Priority

**Minimal TUI for pairing flow.** Full Studio UX comes later.

The pairing UX lives here. Daemon provides the API, TUI provides the experience.

---

## Active Tasks

### HIGH: Minimal Pairing UI

Before the full Studio vision, we need basic pairing to work.

**Pairing Flow:**
1. TUI starts, calls `GET /api/identity` on daemon
2. If unpaired → show pairing screen
3. Call `POST /api/pairing/start` → get session_id, code, URL
4. Display QR code (URL encoded) and confirmation code
5. Poll `GET /api/pairing/status` every 2 seconds
6. On success → show "Paired!" and transition to main view
7. On timeout/cancel → show error, offer retry

**Minimal UI needed:**
- Identity/status check on startup
- Pairing screen with QR and code display
- Polling indicator
- Success/error states

Report in RESPONSES.md:
- Can current TUI display QR codes? (terminal QR library?)
- What daemon API endpoints exist vs needed?
- Proposed implementation approach

### MEDIUM: Identity View (Basic)

After pairing works, flesh out Identity view:
- Agent MRI and profile
- Pairing status (which realm, when)
- Daemon status (running, version, uptime)
- Re-pair / unpair actions

### LOW: Review Studio UX Plan

Read `plans/PLAN_HECATE_STUDIO_UX.md` for the full vision.

Pay special attention to the new section: **Coach Architecture: Rules-First, AI-Optional**

---

## Architecture Decisions (READ THIS)

The plan was updated with critical architectural decisions:

**1. Macula Services are NOT AI-powered**

Services on the mesh are pure business logic. Deterministic. Testable. No LLM runtime.

**2. Two distinct concerns in the TUI:**

| Concern | Implementation | LLM Required |
|---------|----------------|--------------|
| **Coach** (doctrine enforcement) | Rules engine, pattern matching | No |
| **Studio** (code generation) | LLM generates Cartwheel code | **Yes** |

**3. Coach is rules-based:**
- Detect `services/`, `helpers/`, `utils/` → regex on paths
- Catch central supervisors → naming patterns
- Generate corrections → templates

No LLM needed for doctrine enforcement.

**4. Studio requires LLM:**
- Code scaffolding
- Documentation generation
- Architecture guidance
- SVG diagram generation

User must configure a model provider (Ollama, Anthropic, OpenAI).

**5. First-run experience:**
- If no model configured, prompt user to set up
- Detect local Ollama, offer easy path
- Or enter cloud API key

Read the full section in `plans/PLAN_HECATE_STUDIO_UX.md`.

---

## Dependency Note

**Pairing requires daemon API.** The TUI calls the daemon at :4444.

Coordinate with `hecate-daemon/.hecate/QUEUE.md` for the API side.

**Realm must also be ready.** The pairing flow calls realm endpoints:
- `POST /api/v1/pairing/sessions` (create)
- `GET /api/v1/pairing/sessions/:id` (poll)
- Web UI at `/pair/:session_id` (user confirms)

---

## Completed Tasks

*(Pairing is first priority)*

---

*— Hecate*
