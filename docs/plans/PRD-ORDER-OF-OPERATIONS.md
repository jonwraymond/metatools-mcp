# PRD Order of Operations

**Date:** 2026-01-30
**Purpose:** Exact execution order for all consolidation PRDs

---

## Critical Path

```
PRD-100 → PRD-110 → PRD-120 → PRD-130 → PRD-150 → PRD-180 → PRD-190
   ↓         ↓         ↓         ↓         ↓
PRD-101   PRD-111   PRD-121   PRD-131   PRD-151
PRD-102   PRD-112   PRD-122   PRD-132           → PRD-181 → PRD-191
          PRD-113           PRD-133                       → PRD-192
                              ↓
                           PRD-140 → PRD-143
                           PRD-141
                           PRD-142
                              ↓
                           PRD-160 → PRD-162
                           PRD-161
                           PRD-163
                           PRD-164
                              ↓
                           PRD-170 → PRD-171 → ... → PRD-179
                              ↓
                           PRD-182
```

---

## Execution Order (Sequential)

### Week 1: Planning & Infrastructure

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 1 | PRD-100 | Master Plan | 4h | — |
| 2 | PRD-101 | Architecture Diagrams | 4h | PRD-100 |
| 3 | PRD-102 | Schema Definitions | 4h | PRD-100 |
| 4 | PRD-110 | Repository Creation | 8h | PRD-100 |
| 5 | PRD-111 | CI/CD Templates | 4h | PRD-110 |
| 6 | PRD-112 | GitHub Org Config | 2h | PRD-110 |
| 7 | PRD-113 | Release Automation | 2h | PRD-111 |

### Week 2: Foundation Layer

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 8 | PRD-120 | Migrate toolmodel | 4h | PRD-110 |
| 9 | PRD-121 | Migrate tooladapter | 4h | PRD-120 |
| 10 | PRD-122 | Create toolversion | 8h | PRD-120 |

### Week 3: Discovery Layer

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 11 | PRD-130 | Migrate toolindex | 4h | PRD-120 |
| 12 | PRD-131 | Migrate toolsearch | 4h | PRD-130 |
| 13 | PRD-132 | Migrate toolsemantic | 6h | PRD-131 |
| 14 | PRD-133 | Migrate tooldocs | 4h | PRD-120 |

### Week 4: Execution Layer

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 15 | PRD-140 | Migrate toolrun | 4h | PRD-120 |
| 16 | PRD-141 | Migrate toolruntime | 4h | PRD-120 |
| 17 | PRD-142 | Migrate toolcode | 4h | PRD-140 |
| 18 | PRD-143 | Extract toolbackend | 6h | PRD-120 |

### Week 5: Composition + Operations Layers

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 19 | PRD-150 | Migrate toolset | 4h | PRD-121, PRD-130 |
| 20 | PRD-151 | Complete toolskill | 8h | PRD-150, PRD-140 |
| 21 | PRD-160 | Migrate toolobserve | 4h | PRD-120 |
| 22 | PRD-161 | Migrate toolcache | 4h | PRD-120 |

### Week 6: Operations Layer (continued)

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 23 | PRD-162 | Extract toolauth | 8h | PRD-120 |
| 24 | PRD-163 | Create toolresilience | 8h | PRD-120 |
| 25 | PRD-164 | Create toolhealth | 6h | PRD-120 |

### Week 7-8: Protocol Layer

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 26 | PRD-170 | Create tooltransport | 8h | PRD-120 |
| 27 | PRD-171 | Create toolwire | 12h | PRD-170 |
| 28 | PRD-172 | Create tooldiscover | 8h | PRD-171 |
| 29 | PRD-173 | Create toolcontent | 8h | PRD-120 |
| 30 | PRD-174 | Create tooltask | 10h | PRD-140, PRD-173 |
| 31 | PRD-175 | Create toolstream | 8h | PRD-170 |
| 32 | PRD-176 | Create toolsession | 6h | PRD-120 |
| 33 | PRD-177 | Create toolelicit | 6h | PRD-173 |
| 34 | PRD-178 | Create toolresource | 10h | PRD-130 |
| 35 | PRD-179 | Create toolprompt | 8h | PRD-173 |

### Week 9: Integration & Cleanup

| Order | PRD | Title | Est. Hours | Depends On |
|-------|-----|-------|------------|------------|
| 36 | PRD-180 | Update metatools-mcp | 12h | All Phase 2-7 |
| 37 | PRD-181 | Update ai-tools-stack | 4h | PRD-180 |
| 38 | PRD-182 | Documentation Site | 6h | PRD-181 |
| 39 | PRD-190 | Archive Old Repos | 2h | PRD-180 |
| 40 | PRD-191 | Update Submodules | 2h | PRD-190 |
| 41 | PRD-192 | Validation | 4h | PRD-191 |

---

## Parallel Execution Opportunities

These PRDs can run in parallel:

### Parallel Group 1 (After PRD-110)
- PRD-111, PRD-112 (can start together)

### Parallel Group 2 (After PRD-120)
- PRD-121, PRD-122
- PRD-130, PRD-133
- PRD-140, PRD-141, PRD-143
- PRD-160, PRD-161, PRD-162

### Parallel Group 3 (During Phase 7)
- PRD-170 and PRD-173 (independent)
- PRD-175, PRD-176, PRD-177 (independent)
- PRD-178, PRD-179 (independent)

---

## Checkpoint Gates

| Gate | After PRD | Validation |
|------|-----------|------------|
| **G1** | PRD-113 | All repos created, CI working |
| **G2** | PRD-122 | Foundation layer complete, tests pass |
| **G3** | PRD-143 | Discovery + Execution layers complete |
| **G4** | PRD-164 | Composition + Operations layers complete |
| **G5** | PRD-179 | Protocol layer complete |
| **G6** | PRD-182 | Integration complete |
| **G7** | PRD-192 | Full validation, cleanup complete |

---

## PRD File Naming Convention

All PRDs stored in: `docs/plans/`

Format: `PRD-{number}-{short-name}.md`

Examples:
- `PRD-100-master-plan.md`
- `PRD-110-repo-creation.md`
- `PRD-120-migrate-toolmodel.md`
- `PRD-170-create-tooltransport.md`

---

## Quick Reference: What Each PRD Delivers

### Infrastructure PRDs (100s)

| PRD | Deliverable |
|-----|-------------|
| 100 | Master plan document |
| 101 | D2 diagrams, architecture.d2 |
| 102 | schemas/*.json files |
| 110 | 6 empty repos with structure |
| 111 | .github/workflows/*.yml templates |
| 112 | GitHub org secrets configured |
| 113 | release-please-config.json for each |

### Migration PRDs (120-160s)

| PRD | Deliverable |
|-----|-------------|
| 120-122 | toolfoundation/ with 3 packages |
| 130-133 | tooldiscovery/ with 4 packages |
| 140-143 | toolexec/ with 4 packages |
| 150-151 | toolcompose/ with 2 packages |
| 160-164 | toolops/ with 5 packages |

### New Development PRDs (170s)

| PRD | Deliverable |
|-----|-------------|
| 170-179 | toolprotocol/ with 10 packages |

### Integration PRDs (180s)

| PRD | Deliverable |
|-----|-------------|
| 180 | metatools-mcp using new imports |
| 181 | Updated VERSIONS.md, go.mod |
| 182 | Updated MkDocs site |

### Cleanup PRDs (190s)

| PRD | Deliverable |
|-----|-------------|
| 190 | 13 repos archived |
| 191 | New .gitmodules |
| 192 | All smoke tests passing |

---

## Total Effort Summary

| Phase | PRDs | Hours | Days (8h) |
|-------|------|-------|-----------|
| 0 | 100-102 | 12h | 1.5 |
| 1 | 110-113 | 16h | 2 |
| 2 | 120-122 | 16h | 2 |
| 3 | 130-133 | 18h | 2.25 |
| 4 | 140-143 | 18h | 2.25 |
| 5 | 150-151 | 12h | 1.5 |
| 6 | 160-164 | 30h | 3.75 |
| 7 | 170-179 | 84h | 10.5 |
| 8 | 180-182 | 22h | 2.75 |
| 9 | 190-192 | 8h | 1 |
| **Total** | **41 PRDs** | **236h** | **29.5 days** |

With parallelization, estimated calendar time: **6-8 weeks**
