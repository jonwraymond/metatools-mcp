# PRD-123: toolfoundation Docs + README Alignment

**Phase:** 2 - Foundation Layer  
**Priority:** High  
**Effort:** 3 hours  
**Dependencies:** PRD-120, PRD-121, PRD-122  
**Status:** Done (2026-01-31)

---

## Objective

Align public-facing documentation with the consolidated `toolfoundation` API:

- Remove placeholders (README `TBD` entries).
- Ensure docs reflect the **actual** API (validator constructors, signatures).
- Add `version` package coverage in docs and user journey.

---

## Scope

**In scope**
- `toolfoundation/README.md`
- `toolfoundation/docs/index.md`
- `toolfoundation/docs/user-journey.md`

**Out of scope**
- API changes or behavior changes in code
- New functionality beyond documentation

---

## Deliverables

| Deliverable | Location | Description |
|---|---|---|
| Updated README | `toolfoundation/README.md` | Package table + quick usage |
| Updated index docs | `toolfoundation/docs/index.md` | Include `version` package + accurate API snippets |
| Updated user journey | `toolfoundation/docs/user-journey.md` | Correct validator usage + version walkthrough |

---

## Tasks

### Task 1: README package table

- Replace `TBD` with concrete package list:
  - `model`
  - `adapter`
  - `version`
- Add short descriptions + link to `docs/`.

### Task 2: docs/index.md

- Add `version` package to Packages table.
- Add a short version example (Parse / Compare / Compatible).
- Ensure schema validation example uses `NewDefaultValidator`.

### Task 3: docs/user-journey.md

- Fix schema validation snippet to use `NewDefaultValidator` and `ValidateInput(&tool, input)`.
- Add a section showing version negotiation (compatibility matrix and constraints).

### Task 4: Verification

```bash
cd toolfoundation
go test ./...
```

---

## Acceptance Criteria

1. README has concrete package table (no `TBD`).
2. Docs show `version` package in index + user journey.
3. All snippets compile against current API.
4. `go test ./...` passes.

---

## Completion Evidence

- `toolfoundation/README.md` package table updated.
- `toolfoundation/docs/index.md` includes `version` package + examples.
- `toolfoundation/docs/user-journey.md` uses correct validator API.
- `go test ./...` passes in `toolfoundation`.

---

## Next Steps

- PRD-124: Schema validation policy documentation
- PRD-125: Adapter feature matrix documentation
