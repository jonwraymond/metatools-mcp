# Persistence Boundary Architecture

**Status:** Draft (Rewritten)
**Date:** 2026-02-02

## Principle
Separate **in-memory orchestration** from **durable state** to keep the stack portable and testable.

## Boundary Rules
- Core packages should accept storage interfaces, not concrete DBs.
- Durable state lives behind optional interfaces in `toolprotocol` or per-layer packages.
- The reference server (`metatools-mcp`) may include concrete persistence adapters, but should not force them into core libs.

## Candidates for Persistence
- Tool registry snapshots
- Execution history and audit logs
- Cached tool results
- Long-running task state

## References
- `toolprotocol` (interfaces)
- `toolops/cache`, `toolops/observe`
