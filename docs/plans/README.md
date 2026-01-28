# metatools-mcp Implementation Plans

This directory contains Product Requirement Documents (PRDs) for incremental improvements to the metatools-mcp codebase. Each PRD follows TDD methodology with bite-sized tasks.

## PRD Index

### Stream A: Core Exposure (MVP Foundation)

| PRD | Title | Priority | Status | Dependencies |
|-----|-------|----------|--------|--------------|
| [PRD-001](./2026-01-28-prd-001-toolruntime-docker-backend.md) | toolruntime Docker Backend | P0 | Ready | None |
| [PRD-002](./2026-01-28-prd-002-cli-cobra-foundation.md) | CLI Foundation with Cobra | P0 | Ready | None |
| [PRD-003](./2026-01-28-prd-003-koanf-config.md) | Configuration Layer with Koanf | P0 | Ready | PRD-002 |
| [PRD-004](./2026-01-28-prd-004-transport-sse.md) | SSE Transport Layer | P1 | Ready | PRD-002, PRD-003 |
| [PRD-005](./2026-01-28-prd-005-tool-provider-registry.md) | Tool Provider Registry | P0 | Ready | PRD-002, PRD-003 |
| [PRD-006](./2026-01-28-prd-006-backend-registry.md) | Backend Registry | P1 | Ready | PRD-002, PRD-003, PRD-005 |
| [PRD-007](./2026-01-28-prd-007-middleware-chain.md) | Middleware Chain | P2 | Ready | PRD-002, PRD-003, PRD-005 |
| [PRD-015](./2026-01-28-prd-015-mcp-spec-alignment.md) | MCP Spec Alignment (Tools) | P1 | Ready | PRD-005 |

### Stream B: Protocol Layer

| PRD | Title | Priority | Status | Dependencies |
|-----|-------|----------|--------|--------------|
| [PRD-008](./2026-01-28-prd-008-tooladapter-library.md) | tooladapter Library | P1 | Ready | toolmodel |
| [PRD-009](./2026-01-28-prd-009-toolset-composition.md) | toolset Composition | P1 | Ready | PRD-008, toolindex |

### Stream C: Cross-Cutting Concerns

| PRD | Title | Priority | Status | Dependencies |
|-----|-------|----------|--------|--------------|
| [PRD-010](./2026-01-28-prd-010-toolobserve-library.md) | toolobserve Library | P1 | Ready | None |
| [PRD-011](./2026-01-28-prd-011-toolcache-library.md) | toolcache Library | P2 | Ready | None |

### Stream D: Enterprise Features

| PRD | Title | Priority | Status | Dependencies |
|-----|-------|----------|--------|--------------|
| [PRD-012](./2026-01-28-prd-012-multi-tenancy-core.md) | Multi-tenancy Core | P2 | Ready | PRD-005, PRD-007 |
| [PRD-013](./2026-01-28-prd-013-toolsemantic-library.md) | toolsemantic Library | P2 | Ready | toolindex, toolsearch |

### Stream E: Agent Skills

| PRD | Title | Priority | Status | Dependencies |
|-----|-------|----------|--------|--------------|
| [PRD-014](./2026-01-28-prd-014-toolskill-library.md) | toolskill Library | P3 | Ready | PRD-009, toolrun, PRD-010 (opt) |

## Execution Guidelines

### For Claude: Use superpowers:executing-plans

When implementing any PRD:
1. Read the PRD completely
2. Execute tasks sequentially
3. Run tests after each step
4. Commit after each task

### Task Structure

Each task follows this pattern:
1. **Write failing test** - TDD red phase
2. **Run test to verify it fails** - Confirm test setup
3. **Implement minimal code** - TDD green phase
4. **Run test to verify it passes** - Confirm implementation
5. **Commit** - Atomic commits per task

### Verification

Before marking a PRD complete:
- [ ] All tests pass
- [ ] Code coverage > 80%
- [ ] Documentation updated
- [ ] No breaking changes (unless noted)

## Architecture Alignment

These PRDs implement the pluggable architecture defined in:
- [ROADMAP.md](../proposals/ROADMAP.md) - Master plan
- [pluggable-architecture.md](../proposals/pluggable-architecture.md) - Architecture spec
- [implementation-phases.md](../proposals/implementation-phases.md) - Phase timeline

## Work Streams

| Stream | Focus | PRDs | Timeline |
|--------|-------|------|----------|
| **A** | Core Exposure | 001-007 | Weeks 1-8 |
| **B** | Protocol Layer | 008-009 | Weeks 5-10 |
| **C** | Cross-Cutting | 010-011 | Weeks 3-18 |
| **D** | Enterprise | 012-013 | Weeks 8-17 |
| **E** | Agent Skills | 014 | Weeks 17-21 |

## Principles

1. **Incremental** - Each PRD delivers a workable outcome
2. **TDD** - Tests before implementation
3. **DRY/YAGNI** - No unnecessary code
4. **Stable** - No breaking changes unless essential
5. **Pluggable** - Maintain extensibility
