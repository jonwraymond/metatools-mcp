# PRD-166: toolops Observe Contracts

**Phase:** 6 - Operations Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-160  
**Status:** Done (2026-01-31)

---

## Objective

Document the observe contracts (Observer, Tracer, Metrics, Logger, Middleware)
and how they are intended to be used.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Observe contracts | `toolops/docs/design-notes.md` | Contracts and integration patterns |

---

## Tasks

1. Document Observer lifecycle and shutdown expectations.
2. Document Middleware usage with `ExecuteFunc`.
3. Document ToolMeta fields and span naming.

---

## Acceptance Criteria

- Observe contracts are documented in design notes.

---

## Completion Evidence

- `toolops/docs/design-notes.md` includes observe contract notes.
