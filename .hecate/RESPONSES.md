# Apprentice Responses

*Write here when you need Hecate's attention.*

---

## How To Use

When you:
- Complete a task → Report it here
- Have a question → Ask it here
- Hit a blocker → Describe it here
- Need a decision → Request it here

**Format:**
```markdown
## [DATE] [TYPE]: Brief Title

[Your message]

---
```

Types: `COMPLETE`, `QUESTION`, `BLOCKED`, `DECISION`, `UPDATE`

---

## Messages

*(Write below this line)*

---

## 2026-02-03 UPDATE [tui]: Endpoint Mismatch Identified

### Cross-Repo Verification Results

Verified `internal/client/client.go` against `hecate_api_app.erl`. Found issues:

**TUI calls that don't match daemon:**
- `GET /rpc/procedures` - endpoint doesn't exist (daemon has `POST /rpc/track`)
- `GET /capabilities/discover` - should be `POST /capabilities/discover`

**Will fix after completing [node] SKILLS.md rewrite** (same source of truth).

Related: Working on [node] SKILLS.md rewrite first.

---

## 2026-02-03 COMPLETE [tui]: Fixed Endpoint Mismatch

### Changes to `internal/client/client.go`

1. **Fixed `/capabilities/discover`** - Changed from GET to POST with JSON body
   - Added `post()` method for JSON POST requests
   - Added `bytes` import

2. **Fixed `/rpc/procedures`** - Endpoint doesn't exist in daemon
   - `ListProcedures()` now returns empty list with explanatory comment
   - Daemon only has `/rpc/track` for reputation tracking

### Code Changes

```go
// DiscoverCapabilities - now uses POST with JSON body
resp, err := c.post("/capabilities/discover", reqBody)

// ListProcedures - returns empty (endpoint doesn't exist)
func (c *Client) ListProcedures() ([]Procedure, error) {
    return []Procedure{}, nil
}
```

### Note for Future

If daemon adds a `/rpc/procedures` endpoint, update `ListProcedures()` to call it.

---
