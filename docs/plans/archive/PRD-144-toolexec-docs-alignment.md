# PRD-144: toolexec Docs + README Alignment

**Phase:** 4 - Execution Layer  
**Priority:** High  
**Effort:** 2 hours  
**Dependencies:** PRD-140–143  
**Status:** Done (2026-01-31)

---

## Objective

Align public-facing documentation with the consolidated `toolexec` API:

- Replace README placeholders.
- Ensure docs show correct package names and usage.

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Updated README | `toolexec/README.md` | Package table + descriptions |
| Updated docs | `toolexec/docs/*` | Accurate usage examples |

---

## Tasks

1. Update README package table (run/runtime/code/backend).
2. Verify docs/index.md examples match current API.
3. Update docs/user-journey.md runtime example to use `runtime.ExecuteRequest`.
4. Update docs/design-notes.md with current runtime semantics.
5. Run tests.

```bash
cd toolexec
go test ./...
```

---

## Acceptance Criteria

- README has no `TBD` entries.
- Docs reflect current package names and usage.
- Tests pass.

---

## Completion Evidence

- `toolexec/README.md` updated with final package list.
- `toolexec/docs/user-journey.md` runtime example updated to `runtime.ExecuteRequest`.
- `toolexec/docs/design-notes.md` updated with current runtime profile semantics.
