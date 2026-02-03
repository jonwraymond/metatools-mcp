# PRD-135: tooldiscovery Search Strategy Policy Docs

**Phase:** 3 - Discovery Layer  
**Priority:** Medium  
**Effort:** 2 hours  
**Dependencies:** PRD-131  
**Status:** Done (2026-01-31)

---

## Objective

Document search strategy behavior and configuration for the `search` package:

- BM25 configuration (field boosts)
- Indexing scope and caching behavior
- Guidance for choosing strategies

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Search policy section | `tooldiscovery/docs/design-notes.md` | BM25 config + behavior |
| Summary note | `tooldiscovery/docs/index.md` | Short guidance + link |

---

## Acceptance Criteria

- BM25 configuration is explicitly documented.
- Guidance for selecting lexical vs BM25 vs semantic is present.

---

## Completion Evidence

- Search policy documented in `tooldiscovery/docs/design-notes.md`.
- Summary guidance added in `tooldiscovery/docs/index.md`.
