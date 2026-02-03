# PRD-146: toolexec Backend Matrix Documentation

**Phase:** 4 - Execution Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-141, PRD-143  
**Status:** Done (2026-01-31)

---

## Objective

Provide a clear matrix of runtime backend kinds, isolation levels, and
environment requirements so operators can choose appropriate backends.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Backend matrix | `toolexec/docs/design-notes.md` | Table of backend kinds + requirements |

---

## Tasks

1. List all runtime backend kinds in a single table.
2. Capture isolation level and key requirements (Docker, containerd, k8s, etc.).
3. Note dev-only and strongest-isolation options.

---

## Acceptance Criteria

- Backend kinds are enumerated.
- Requirements and isolation levels are documented.

---

## Completion Evidence

- `toolexec/docs/design-notes.md` includes a Runtime Backend Matrix table.
