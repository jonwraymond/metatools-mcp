# PRD-015 Execution Plan — MCP Spec Alignment (Strict TDD)

**Status:** Ready
**Scope:** Phase 0 — Spec Alignment & Server Correctness (P1)
**Date:** 2026-01-28

This plan is a **strict TDD execution guide** for PRD‑015. It includes all tasks, sub‑tasks, ordering, and rationale. **No assumptions** are made beyond what is stated here.

---

## Why this ordering (Go architecture rationale)

1) **Cancellation propagation first** ensures we do not build correctness on top of handlers that may leak goroutines or ignore context timeouts.
2) **List change notifications second** depend on stable behavior and should not spam clients (debounce must be built in).
3) **Pagination/cursor** requires deterministic ordering; once exposed at MCP edge it must remain stable.
4) **Optional progress forwarding last** because it is conditional on backend support and introduces async complexity.

---

## Global TDD Rules (mandatory)

For every task:

1) **Red:** write a failing test
2) **Red verification:** run the test, confirm it fails
3) **Green:** implement the smallest change to pass
4) **Green verification:** run the test, confirm it passes
5) **Refactor:** optional, only if test remains green
6) **Commit:** one atomic commit per task

---

## Task 1 — Cancellation Propagation (Go correctness baseline)

**Repos:**
- `toolrun`
- `toolruntime` (only if gateway/sandbox ignores ctx)
- `metatools-mcp`

### Sub‑task 1.1 — toolrun cancellation behavior

**Goal:** `Run`, `RunStream`, `RunChain` must honor context cancellation.

**Red (test):**
- Add a long‑running local tool handler that blocks until context is cancelled.
- Start `Run` with `ctx`, cancel it, expect error consistent with cancellation.
- For `RunChain`, ensure cancellation mid‑chain aborts subsequent steps.

**Green (implementation):**
- Ensure `ctx` is passed through to local handler calls and MCP/provider runners.
- Ensure `RunChain` checks `ctx.Done()` between steps.

**Commit:** `test/toolrun: enforce ctx cancellation` + `fix(toolrun): honor ctx cancellation`

### Sub‑task 1.2 — toolruntime cancellation (if applicable)

**Goal:** runtime gateways and sandbox backends honor ctx.

**Red (test):**
- Use a sandbox backend with a long‑running execution and cancel the ctx.
- Assert backend stops and returns cancellation error.

**Green (implementation):**
- Propagate ctx through gateway methods and backend calls.

**Commit:** `test(toolruntime): ctx cancellation` + `fix(toolruntime): propagate ctx`

### Sub‑task 1.3 — metatools-mcp cancellation

**Goal:** MCP handlers pass request context to toolrun/toolcode and surface cancellation properly.

**Red (test):**
- Invoke `run_tool` with a cancellable ctx; cancel mid‑run; assert handler returns error.
- Invoke `run_chain` with cancellation; ensure partial results and error semantics match.

**Green (implementation):**
- Ensure handler uses incoming context without background replacement.

**Commit:** `test(metatools): cancellation propagation` + `fix(metatools): honor ctx cancellation`

---

## Task 2 — `notifications/tools/list_changed`

**Repos:**
- `metatools-mcp`
- (Optional) `toolindex` if change hooks are insufficient

### Sub‑task 2.1 — notification emission

**Goal:** emit MCP `notifications/tools/list_changed` when toolindex changes.

**Red (test):**
- Create server with index stub capable of triggering OnChange/Refresh.
- Assert MCP server emits `tools/list_changed` notification.

**Green (implementation):**
- Subscribe to index change hooks.
- Emit notifications to all MCP sessions.

**Commit:** `test(metatools): tools/list_changed` + `feat(metatools): emit list_changed`

### Sub‑task 2.2 — debounce

**Goal:** prevent notification storms on bulk tool registration.

**Red (test):**
- Trigger 10 rapid changes; assert only 1 notification within debounce window.

**Green (implementation):**
- Add configurable debounce (default 100–250ms).

**Commit:** `test(metatools): debounce list_changed` + `feat(metatools): debounce notifications`

### Sub‑task 2.3 — config toggle

**Goal:** allow disabling notifications for static deployments.

**Red (test):**
- Set `METATOOLS_NOTIFY_TOOL_LIST_CHANGED=false` and assert no notification.

**Green (implementation):**
- Add config flag and gate the subscriber.

**Commit:** `test(metatools): notification toggle` + `feat(metatools): config flag for list_changed`

---

## Task 3 — Pagination & Cursor Consistency

**Repos:**
- `toolindex` (paging helpers)
- `metatools-mcp` (cursor passthrough + output contract)

### Sub‑task 3.1 — toolindex paging helpers

**Goal:** deterministic, stable paging with opaque cursors.

**Red (test):**
- Register tools and assert search results are stable and page correctly.
- Cursor token must produce identical next page given same input.

**Green (implementation):**
- Stable ordering: `namespace:name` ascending.
- Cursor encoding: base64 JSON `{offset, checksum}` (opaque to caller).
- If invalid cursor: return error.

**Commit:** `test(toolindex): cursor paging` + `feat(toolindex): SearchPage/ListNamespacesPage`

### Sub‑task 3.2 — metatools cursor passthrough

**Goal:** expose cursor in `search_tools` and `list_namespaces` output.

**Red (test):**
- Call MCP `search_tools` with cursor; expect `nextCursor` to be non‑empty.

**Green (implementation):**
- Add cursor handling to handlers and schemas.
- Enforce limit caps consistently.

**Commit:** `test(metatools): cursor support` + `feat(metatools): cursor paging`

---

## Task 4 — Optional Progress Forwarding

**Repos:**
- `toolrun` (optional interface)
- `toolruntime` (if progress available)
- `metatools-mcp`

### Sub‑task 4.1 — progress interface

**Goal:** optional progress stream without hard coupling.

**Red (test):**
- Fake runner implements `ProgressEmitter` and emits events.

**Green (implementation):**
- Introduce `ProgressEmitter` interface and no‑op default.

**Commit:** `test(toolrun): progress emitter` + `feat(toolrun): progress interface`

### Sub‑task 4.2 — MCP progress forwarding

**Goal:** forward progress events to MCP clients.

**Red (test):**
- Invoke `run_tool` and assert progress notifications are emitted.

**Green (implementation):**
- If runner implements `ProgressEmitter`, forward events.
- Otherwise, no progress behavior.

**Commit:** `test(metatools): progress forwarding` + `feat(metatools): progress notifications`

---

## Documentation Updates (after each task)

- `docs/usage.md`: env vars and pagination/cursor semantics
- `docs/api.md`: schema updates for `search_tools`, `list_namespaces`
- `docs/design-notes.md`: cancellation + notifications + progress behavior
- `docs/proposals/mcp-spec-alignment.md`: mark implemented sections

---

## Verification Checklist (end of Phase 0)

- [x] All tests pass in `toolrun`, `toolruntime` (if touched), `toolindex`, `metatools-mcp`
- [x] MCP `tools/list_changed` notifications observed under test
- [x] Cursor paging deterministic and opaque
- [x] Cancellation honored for all tool execution paths
- [x] Progress forwarding documented and conditional
- [x] Docs updated and consistent

---

## Commit Order (strict)

1) toolrun (cancellation)
2) toolruntime (cancellation if needed)
3) metatools-mcp (cancellation)
4) metatools-mcp (notifications + debounce + config)
5) toolindex (cursor paging)
6) metatools-mcp (cursor passthrough)
7) toolrun/toolruntime (progress interface if desired)
8) metatools-mcp (progress forwarding)
9) docs update

---

## Notes

- If any repo change introduces a new API, bump version in that repo and update `ai-tools-stack` matrix accordingly.
- Do **not** merge partial PRs that break MCP tool schemas.
- Use smallest change per PR to keep CI green and revertable.
