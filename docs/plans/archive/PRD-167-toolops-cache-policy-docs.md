# PRD-167: toolops Cache Policy Docs

**Phase:** 6 - Operations Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-161  
**Status:** Done (2026-01-31)

---

## Objective

Document cache keying, policy semantics, and unsafe tag handling.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Cache policy docs | `toolops/docs/design-notes.md` | Keying + policy semantics |

---

## Tasks

1. Document deterministic keying (canonical JSON + SHA‑256).
2. Document policy behavior (TTL defaults, clamping, unsafe tags).
3. Document middleware behavior (no cache on errors).

---

## Acceptance Criteria

- Cache policy and keying are documented.

---

## Completion Evidence

- `toolops/docs/design-notes.md` includes cache policy section.
