# PRD-145: toolexec Runtime Security Profiles Documentation

**Phase:** 4 - Execution Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-141  
**Status:** Done (2026-01-31)

---

## Objective

Document the runtime security profiles and their contract so consumers know
which isolation guarantees they receive for each profile.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Profile contract | `toolexec/docs/design-notes.md` | Profile definitions + usage guidance |
| Runtime doc update | `toolexec/runtime/doc.go` | Package name + contract summary |

---

## Tasks

1. Document `ProfileDev`, `ProfileStandard`, `ProfileHardened` semantics.
2. Clarify Gateway requirement and limit enforcement expectations.
3. Ensure package doc comment matches `package runtime`.

---

## Acceptance Criteria

- Profiles are described with intended isolation level.
- Gateway requirement is explicit.
- Package documentation is consistent with `package runtime`.

---

## Completion Evidence

- `toolexec/docs/design-notes.md` updated with profile definitions and Gateway requirement.
- `toolexec/runtime/doc.go` comment updated to `Package runtime`.
