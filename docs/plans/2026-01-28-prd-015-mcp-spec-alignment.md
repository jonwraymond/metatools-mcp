# PRD-015: MCP Spec Alignment (Tools)

**Status:** Ready
**Priority:** P1
**Owner:** metatools-mcp
**Date:** 2026-01-28

## Objective

Align `metatools-mcp` with the MCP 2025-11-25 tool semantics by adding dynamic tool list change notifications, consistent pagination/cursor behavior, and explicit cancellation/progress propagation for long-running calls.

## Non-Goals

- Adding MCP resources or prompts (separate PRD).
- Introducing new execution backends.
- Changing toolmodel schemas.

## Requirements

### R1 — Tool list change notifications

- Emit `notifications/tools/list_changed` when tools are added, removed, or updated.
- Wire to `toolindex` change hooks (OnChange/Refresh).
- Provide a config flag to disable notifications.

### R2 — Pagination & cursor consistency

- Enforce consistent `limit` caps for `search_tools` and `list_namespaces`.
- Use opaque cursor tokens for stable pagination (even if in-memory).

### R3 — Cancellation propagation

- Ensure `ctx` cancellation aborts `run_tool`, `run_chain`, and `execute_code`.
- Document behavior for backends that do not support cancellation.

### R4 — Progress wiring (optional v1)

- If `toolrun` or `toolruntime` exposes progress events, surface them as MCP progress notifications.
- If not available, document that progress is unsupported.

## TDD Plan

### Task 1 — Tool list change notifications

1. **Write failing test** for `tools/list_changed` emission when toolindex changes.
2. **Run test** and confirm failure.
3. **Implement** hook wiring and notification emission.
4. **Run test** and confirm pass.
5. **Commit** `feat(metatools): emit tool list change notifications`.

### Task 2 — Pagination and cursor consistency

1. **Write failing tests** for limit caps and cursor semantics.
2. **Run tests** and confirm failure.
3. **Implement** cursor helper and caps.
4. **Run tests** and confirm pass.
5. **Commit** `feat(metatools): normalize search pagination`.

### Task 3 — Cancellation propagation

1. **Write failing tests** that cancel context mid-call.
2. **Run tests** and confirm failure.
3. **Implement** cancellation handling in handlers.
4. **Run tests** and confirm pass.
5. **Commit** `feat(metatools): propagate cancellation`.

### Task 4 — Progress wiring (optional)

1. **Write failing test** for progress event forwarding (if supported by runner).
2. **Implement** forwarding or explicitly skip with documented limitation.
3. **Commit** `feat(metatools): progress forwarding (optional)`.

## Acceptance Criteria

- Tools change notifications are emitted when toolindex changes.
- Search/list pagination behaves consistently and caps are enforced.
- Cancellation of context stops tool execution where possible.
- Documentation updated to reflect any limitations.

## Dependencies

- `toolindex` OnChange/Refresh hooks.
- `toolrun` cancellation semantics.
- MCP Go SDK notification APIs.

## Notes

- Keep feature gated by config to preserve static deployments.
- Debounce notifications to avoid client spam.
