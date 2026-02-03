# PRD-151: Complete toolskill

**Phase:** 5 - Composition Layer
**Priority:** High
**Effort:** 8 hours
**Dependencies:** PRD-150, PRD-140
**Status:** Done (2026-01-31)

---

## Objective

Migrate the partial `toolskill` implementation and complete it as `toolcompose/skill/` for agent skills management.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/toolskill` (partial implementation)
**Target Location:** `github.com/jonwraymond/toolcompose/skill`

**Current State:**
- Minimal declarative skill model (Skill/Step) implemented
- Planner provides deterministic ordering
- Guards + Execute flow exist via Runner interface

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Skill Package | `toolcompose/skill/` | Declarative skill model |
| Planner | `skill/planner.go` | Deterministic plan generation |
| Guards | `skill/guard.go` | Policy validation helpers |
| Executor | `skill/execute.go` | Step execution via Runner |
| Tests | `skill/*_test.go` | Contract + behavior tests |

---

## Tasks

1. Ensure `Skill`, `Step`, `Planner`, `Guard`, and `Execute` are implemented in `toolcompose/skill`.
2. Validate deterministic planning (sorted by step ID).
3. Provide guard helpers for max steps and allowed tool IDs.
4. Ensure runner interface is documented and tested.
5. Update docs/examples to reflect the minimal skill model.

## Verification Checklist

- [ ] Core interfaces defined
- [ ] Planner produces deterministic ordering
- [ ] Guard helpers enforce constraints
- [ ] Execute uses Runner
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Error handling works
- [ ] Package documentation complete

---

## Acceptance Criteria

1. `toolcompose/skill` package builds successfully
2. Skill + Step validation behaves correctly
3. Planner produces deterministic ordering
4. Execute uses Runner and returns StepResults
5. Guards enforce max steps and allowed tool IDs

## Completion Notes

- Skill package provides `Skill`, `Step`, `Planner`, `Guard`, and `Execute` primitives.
- Runner integration is explicit to keep tool execution decoupled.

---

## Rollback Plan

```bash
cd /tmp/migration/toolcompose
rm -rf skill/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- Gate G4: Composition layer complete (both packages)
- PRD-160: Migrate toolobserve
