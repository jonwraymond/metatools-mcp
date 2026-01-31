# PRD-147: toolexec Toolcode ↔ Runtime Contract

**Phase:** 4 - Execution Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-142, PRD-141  
**Status:** Done (2026-01-31)

---

## Objective

Document the contract between `code` orchestration and the runtime layer,
including how `ExecuteParams` map to `runtime.ExecuteRequest` and how tool
calls flow through the Gateway.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Contract section | `toolexec/docs/design-notes.md` | Toolcode ↔ Runtime mapping |

---

## Tasks

1. Describe how `code` delegates to `runtime/toolcodeengine`.
2. Document mapping of profile, limits, and Gateway injection.
3. Note that the runtime enforces tool call limits and returns ToolCall records.

---

## Acceptance Criteria

- Contract is documented in design notes.
- Mapping of parameters is explicit.

---

## Completion Evidence

- `toolexec/docs/design-notes.md` includes Toolcode ↔ Runtime Contract section.
