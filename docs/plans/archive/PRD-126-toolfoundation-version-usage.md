# PRD-126: toolfoundation Version Package Usage Docs

**Phase:** 2 - Foundation Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-122  
**Status:** Done (2026-01-31)

---

## Objective

Publish usage guidance for `toolfoundation/version`:

- Semantic version parsing and comparison
- Constraints (`>=`, `^`, `~`)
- Compatibility matrix + negotiation

---

## Scope

**In scope**
- `toolfoundation/docs/index.md`
- `toolfoundation/docs/user-journey.md`

**Out of scope**
- Changing version semantics

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Version example | `docs/index.md` | Parse/Compare/Compatible usage |
| Compatibility example | `docs/user-journey.md` | Matrix + negotiate example |

---

## Acceptance Criteria

- Version package is documented in both index and user journey.
- Examples compile against current API.

---

## Completion Evidence

- `version` examples added to `toolfoundation/docs/index.md`.
- Compatibility example added to `toolfoundation/docs/user-journey.md`.

---

## Next Steps

- PRD-127: Contract verification
