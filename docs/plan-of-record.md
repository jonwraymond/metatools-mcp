# Plan of Record (Ordered Execution)

This page consolidates **all proposals and PRDs** into a single, ordered execution sequence. It is intentionally Go‑architecture focused: concurrency safety, context propagation, and production operational boundaries are explicit in each phase.

## Principles (Go Architect Evaluation)

- **Context propagation is a contract**: every long‑running or remote execution path must honor `context.Context` for cancellation and timeouts.
- **Concurrency safety by default**: all registries and caches must be thread‑safe under concurrent reads/writes.
- **Errors are data**: tool errors should be structured and preserved end‑to‑end rather than surfaced as raw panics.
- **Stable seams first**: prioritize interfaces, wiring, and deterministic behaviors before adding new feature surface.

## Current Baseline (already in place)

These libraries and contracts are the foundation and must remain stable:

- `toolmodel` – core types and MCP schema compatibility
- `toolindex` – registry + discovery
- `tooldocs` – progressive disclosure docs/examples
- `toolrun` – execution and chaining
- `toolcode` – code orchestration (optional)
- `toolruntime` – sandbox/runtime isolation (optional)
- `toolsearch` – BM25 search implementation (optional)

See the master roadmap for the current version matrix: [ROADMAP](proposals/ROADMAP.md)

---

## Phase 0 — Spec Alignment & Server Correctness (P1)

**Goal:** Ensure the MCP server edge is protocol‑correct before expanding capability.

1. **MCP spec alignment (PRD‑015)**
   - `notifications/tools/list_changed`
   - pagination/cursor consistency
   - cancellation propagation
   - optional progress forwarding

Docs:
- Proposal: [MCP Spec Alignment](proposals/mcp-spec-alignment.md)
- PRD: [PRD‑015](plans/2026-01-28-prd-015-mcp-spec-alignment.md)

---

## Phase 1 — Core Exposure (MVP Foundation)

**Goal:** Provide CLI, configuration, transport, and provider/backends for production use.

2. **CLI foundation (PRD‑002)**
3. **Configuration layer (PRD‑003)**
4. **Transport layer (PRD‑004)**
5. **Tool provider registry (PRD‑005)**
6. **Backend registry (PRD‑006)**
7. **Middleware chain (PRD‑007)**

Docs:
- PRDs: [PRD‑002](plans/2026-01-28-prd-002-cli-cobra-foundation.md), [PRD‑003](plans/2026-01-28-prd-003-koanf-config.md),
  [PRD‑004](plans/2026-01-28-prd-004-transport-sse.md), [PRD‑005](plans/2026-01-28-prd-005-tool-provider-registry.md),
  [PRD‑006](plans/2026-01-28-prd-006-backend-registry.md), [PRD‑007](plans/2026-01-28-prd-007-middleware-chain.md)

---

## Phase 2 — Protocol Layer

**Goal:** Normalize tools into composable, protocol‑agnostic sets without changing core semantics.

8. **tooladapter (PRD‑008)**
9. **toolset (PRD‑009)**

Docs:
- [PRD‑008](plans/2026-01-28-prd-008-tooladapter-library.md)
- [PRD‑009](plans/2026-01-28-prd-009-toolset-composition.md)

---

## Phase 3 — Cross‑Cutting Observability & Caching

**Goal:** Make the system operationally measurable and resilient.

10. **toolobserve (PRD‑010)**
11. **toolcache (PRD‑011)**

Docs:
- [PRD‑010](plans/2026-01-28-prd-010-toolobserve-library.md)
- [PRD‑011](plans/2026-01-28-prd-011-toolcache-library.md)

---

## Phase 4 — Enterprise Extensions

**Goal:** Enable scale, isolation, and advanced discovery without destabilizing core APIs.

12. **Multi‑tenancy core (PRD‑012)**
13. **toolsemantic (PRD‑013)**

Docs:
- [PRD‑012](plans/2026-01-28-prd-012-multi-tenancy-core.md)
- [PRD‑013](plans/2026-01-28-prd-013-toolsemantic-library.md)

---

## Phase 5 — Agent Skills

**Goal:** Higher‑level capability composition for reusable workflows.

14. **toolskill (PRD‑014)**

Docs:
- [PRD‑014](plans/2026-01-28-prd-014-toolskill-library.md)

---

## Phase 6 — Runtime Expansion

**Goal:** Expand sandbox options and isolation strategies.

15. **toolruntime Docker backend (PRD‑001)**

Docs:
- [PRD‑001](plans/2026-01-28-prd-001-toolruntime-docker-backend.md)

---

## Go Architecture Review (Summary)

- **Context propagation:** enforce in all public execution APIs; cancellation must be honored by toolrun/toolruntime to avoid leaked goroutines.
- **Concurrency safety:** all registries must be RW‑safe; avoid maps without guards under write paths.
- **Pagination correctness:** use stable cursors and cap limits across list endpoints.
- **Error semantics:** preserve tool errors as structured data; avoid panics in runtime paths.
- **Observability:** add tracing hooks before multi‑tenant and semantic layers to avoid blind spots.

---

## Reference Docs

- [ROADMAP](proposals/ROADMAP.md)
- [Pluggable Architecture](proposals/pluggable-architecture.md)
- [Implementation Phases](proposals/implementation-phases.md)
- [Architecture Evaluation](proposals/architecture-evaluation.md)
- [Protocol‑Agnostic Tools](proposals/protocol-agnostic-tools.md)
- [Multi‑Tenancy](proposals/multi-tenancy.md)
- [Architecture Review](proposals/ARCHITECTURE-REVIEW.md)
