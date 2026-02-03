# PRD-017: Pluggable Authentication & Authorization Middleware

**Status:** Complete
**Date:** 2026-02-03
**Priority:** P1 (High)
**Depends On:** PRD-016 (Interface Contracts) [archived]
**Proposal:** `../proposals/auth-middleware.md`
**Architecture:** `../proposals/persistence-boundary.md`

---

## Overview
Consolidate authentication and authorization behind **toolops/auth** and wire it into `metatools-mcp` as middleware. This replaces duplicate auth code and provides a single contract for identity, authorization, and request context propagation.

## Goals
- Use `toolops/auth` as the only auth source of truth.
- Enforce auth for discovery + execution in `metatools-mcp`.
- Propagate identity into request context for downstream policy checks.
- Provide a simple config surface for selecting auth strategies.

## Non-Goals
- Full IAM system or user management.
- Multi-tenancy (separate proposal).
- Long-term persistence of auth state (see persistence-boundary proposal).

## Scope
- Replace `metatools-mcp/internal/auth` usage with `toolops/auth`.
- Add auth middleware in `metatools-mcp/internal/middleware` using `toolops/auth`.
- Update config and docs to describe auth setup.

## Implementation Summary
- Replaced internal auth usage with `toolops/auth` (transport + middleware).
- Added auth middleware factory in `metatools-mcp/internal/middleware`.
- Updated middleware registry and docs/examples to reflect new config.
- Removed duplicated internal auth implementation.

## Acceptance Criteria
- `metatools-mcp` uses `toolops/auth` for all auth decisions.
- No duplicated auth implementation remains in the reference server.
- Tests cover allow/deny paths for discovery + execution.
- Docs show clear setup steps.

## Risks
- Divergent auth semantics between old and new implementations.
- Breaking change for users relying on internal auth types.

## References
- `toolops/auth`
- `metatools-mcp/internal/middleware`
- `ai-tools-stack/docs/roadmap.md`
