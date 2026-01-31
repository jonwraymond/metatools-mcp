# PRD-125: toolfoundation Adapter Feature Matrix Docs

**Phase:** 2 - Foundation Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-121  
**Status:** Done (2026-01-31)

---

## Objective

Document adapter feature support and loss semantics for `toolfoundation/adapter`:

- Supported schema features by target format
- How `FeatureLossWarning` is generated
- Guidance for consumers on safe conversions

---

## Scope

**In scope**
- `toolfoundation/docs/design-notes.md`
- `toolfoundation/docs/index.md`

**Out of scope**
- Changing adapter behavior
- Adding new adapters

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Feature matrix | `docs/design-notes.md` | Table of features by adapter |
| Warnings guidance | `docs/design-notes.md` | How to interpret `FeatureLossWarning` |
| Public summary | `docs/index.md` | Short summary with link |

---

## Acceptance Criteria

- Feature matrix is present and aligns with adapter implementation.
- Warning semantics are documented with an example.

---

## Completion Evidence

- Feature matrix + warning semantics documented in `toolfoundation/docs/design-notes.md`.
- Summary link added in `toolfoundation/docs/index.md`.

---

## Next Steps

- PRD-126: Version package usage docs
