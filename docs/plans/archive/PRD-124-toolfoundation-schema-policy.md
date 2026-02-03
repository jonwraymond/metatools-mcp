# PRD-124: toolfoundation Schema Validation Policy

**Phase:** 2 - Foundation Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-120  
**Status:** Done (2026-01-31)

---

## Objective

Document the JSON Schema validation contract for `toolfoundation/model`:

- Supported dialects (2020-12, draft-07)
- External `$ref` handling policy (blocked)
- Validation limitations (format/content keywords)

---

## Scope

**In scope**
- `toolfoundation/docs/design-notes.md`
- `toolfoundation/docs/index.md`

**Out of scope**
- Changing validation behavior
- Adding new validation libraries

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Schema policy notes | `docs/design-notes.md` | Dialects + external ref policy + limitations |
| Public summary | `docs/index.md` | Short section linking to design notes |

---

## Tasks

1. Add a “Schema Validation Policy” section to `docs/design-notes.md`.
2. Add a short summary in `docs/index.md` with a link to the policy section.
3. Verify examples remain accurate.

---

## Acceptance Criteria

- Supported dialects and limitations are explicitly documented.
- External `$ref` policy is stated.
- Docs are consistent with `model/validator.go`.

---

## Completion Evidence

- Schema policy documented in `toolfoundation/docs/design-notes.md`.
- Summary added to `toolfoundation/docs/index.md`.

---

## Next Steps

- PRD-125: Adapter feature matrix documentation
