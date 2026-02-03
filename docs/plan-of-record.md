# Plan of Record (Ordered Execution)

**Status:** Superseded  
**Canonical Roadmap:** `ai-tools-stack/docs/roadmap.md`

This page consolidates **all proposals and PRDs** into a single, ordered execution sequence. It is intentionally Go‑architecture focused: concurrency safety, context propagation, and production operational boundaries are explicit in each phase.

## Principles (Go Architect Evaluation)

- **Context propagation is a contract**: every long‑running or remote execution path must honor `context.Context` for cancellation and timeouts.
- **Concurrency safety by default**: all registries and caches must be thread‑safe under concurrent reads/writes.
- **Errors are data**: tool errors should be structured and preserved end‑to‑end rather than surfaced as raw panics.
- **Stable seams first**: prioritize interfaces, wiring, and deterministic behaviors before adding new feature surface.

## Current Baseline (already in place)

These libraries and contracts are the foundation and must remain stable:

- `toolfoundation` – core types + adapters + versioning
- `tooldiscovery` – registry, docs, search strategies
- `toolexec` – execution, orchestration, runtime isolation
- `toolcompose` – toolsets + skills
- `toolops` – observability, cache, auth, resilience, health
- `toolprotocol` – transport, wire, content, session, task primitives

See the canonical stack roadmap for current status and version matrix: `ai-tools-stack/docs/roadmap.md`

---

## Phase 0 — Spec Alignment & Server Correctness (P1)

**Goal:** Ensure the MCP server edge is protocol‑correct before expanding capability.

1. **MCP spec alignment**
   - `notifications/tools/list_changed`
   - pagination/cursor consistency
   - cancellation propagation
   - optional progress forwarding

Docs:
- Proposal: [MCP Spec Alignment](proposals/mcp-spec-alignment.md)
- PRD: [PRD‑180](plans/PRD-180-update-metatools-mcp.md)

---

## Phase 1 — Core Exposure (MVP Foundation)

**Goal:** Provide CLI, configuration, transport, and provider/backends for production use.

2. **Repo scaffolding + CLI surface**
3. **Configuration layer**
4. **Transport layer**
5. **Tool provider registry**
6. **Backend registry**
7. **Middleware chain**

Docs:
- PRDs: [PRD‑110](plans/PRD-110-repo-creation.md), [PRD‑111](plans/PRD-111-cicd-templates.md),
  [PRD‑112](plans/PRD-112-github-org-config.md), [PRD‑113](plans/PRD-113-release-automation.md)

---

## Phase 2 — Protocol Layer

**Goal:** Normalize tools into composable, protocol‑agnostic sets without changing core semantics.

8. **tooladapter** → now `toolfoundation/adapter`
9. **toolset** → now `toolcompose/set`

Docs:
- [PRD‑121](plans/PRD-121-migrate-tooladapter.md)
- [PRD‑150](plans/PRD-150-migrate-toolset.md)

---

## Phase 3 — Cross‑Cutting Observability & Caching

**Goal:** Make the system operationally measurable and resilient.

10. **toolobserve** → now `toolops/observe`
11. **toolcache** → now `toolops/cache`

Docs:
- [PRD‑160](plans/PRD-160-migrate-toolobserve.md)
- [PRD‑161](plans/PRD-161-migrate-toolcache.md)

---

## Phase 4 — Enterprise Extensions

**Goal:** Enable scale, isolation, and advanced discovery without destabilizing core APIs.

12. **Multi‑tenancy core**
13. **toolsemantic** → now `tooldiscovery/semantic`

Docs:
- Proposal: [Multi‑Tenancy](proposals/multi-tenancy.md)
- [PRD‑132](plans/PRD-132-migrate-toolsemantic.md)

---

## Phase 5 — Agent Skills

**Goal:** Higher‑level capability composition for reusable workflows.

14. **toolskill** → now `toolcompose/skill`

Docs:
- [PRD‑151](plans/PRD-151-complete-toolskill.md)

---

## Phase 6 — Runtime Expansion

**Goal:** Expand sandbox options and isolation strategies.

15. **toolruntime Docker backend** → now `toolexec/runtime`

Docs:
- [PRD‑141](plans/PRD-141-migrate-toolruntime.md)

---

## Go Architecture Review (Summary)

- **Context propagation:** enforce in all public execution APIs; cancellation must be honored by toolexec/run and toolexec/runtime to avoid leaked goroutines.
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
