# MCP Spec Alignment Proposal

**Status:** Draft
**Date:** 2026-01-28
**Related:** [pluggable-architecture](./pluggable-architecture.md), [architecture-evaluation](./architecture-evaluation.md), [ROADMAP](./ROADMAP.md)

## Summary

`metatools-mcp` already exposes a clean MCP-native tool surface, but several MCP-spec features are only partially covered or are not yet surfaced at the server boundary. This proposal defines a focused alignment path with the **latest MCP spec (2025-11-25)** and the official Go SDK, while keeping the current tool-first architecture stable.

The goal is to improve protocol completeness and operational correctness without redesigning the existing libraries.

## Goals

1. **Spec compliance at the server edge**: align handler behavior with MCP method contracts (tools list, tool calls, tool error semantics, pagination).
2. **Dynamic tool updates**: support `tools/list_changed` notifications tied to toolindex change events.
3. **Operational robustness**: ensure cancellation/progress signals propagate through toolrun and toolruntime where supported.
4. **Clear expansion path** for MCP resources and prompts without blocking current tool-focused roadmap.

## Observations (Current State)

- `metatools-mcp` exposes MCP tools via the official Go SDK and maps errors into MCP tool error payloads.
- Tool schemas are derived from `toolmodel` and remain MCP-native, preserving compatibility end-to-end.
- Tool registration is static at startup; runtime updates now emit `tools/list_changed` via toolindex change notifications (debounced and optionally disabled).
- Resources and prompts are not yet part of the metatools surface.
- Cancellation behavior now propagates via `context.Context`; progress notifications remain dependent on downstream runner support.

## MCP Spec Alignment Targets

### 1) Tool list change notifications

When tool availability changes (new tools, removed tools, updated schemas), the MCP server should notify connected clients via `notifications/tools/list_changed`. This can be wired to `toolindex` change callbacks.

**Implementation idea:**
- Subscribe to toolindex OnChange/Refresh hooks.
- Emit `tools/list_changed` to all MCP sessions.
- Expose a config flag + debounce window to disable or dampen notifications for static deployments.

### 2) Pagination and list contracts

The MCP spec defines `tools/list` pagination and page sizes. `metatools-mcp` already sets a default page size in the Go SDK; align `search_tools` and `list_namespaces` with consistent paging semantics and cursor shapes.

**Implementation idea:**
- Cap and validate list size consistently.
- Use stable cursor generation (opaque tokens with validation) for pagination of search results and namespace listing.

### 3) Cancellation and progress signals

MCP supports cancellation and progress notifications for long-running operations. For `run_tool`, `run_chain`, and `execute_code`, cancellation should be propagated to toolrun/toolruntime if supported.

**Implementation idea:**
- Use `ctx` cancellation to interrupt tool execution.
- Surface progress events from toolruntime or toolrun (where available).
- If progress events are unavailable, document that progress notifications are not emitted.

### 4) Resources and prompts expansion (future PRD)

MCP defines resources and prompts alongside tools. The architecture should leave a clear integration path (e.g., new `toolresource` and `toolprompt` libraries or adapters in metatools-mcp).

**Implementation idea:**
- Draft a separate PRD to introduce a resources/prompt provider interface.
- Keep current tool-only server behavior stable and backwards compatible.

## Corrected Assumptions

- **MCP versioning:** target **2025-11-25** as the default protocol version in metatools-mcp and toolmodel. Back-compat shims should be explicit rather than implicit.
- **Go SDK alignment:** the official MCP Go SDK should be pinned to a version that explicitly supports the 2025-11-25 spec and its tool schema fields (icons, outputSchema, etc.).

## Risks

- **Notification spam:** emitting list change notifications on every small index update could overwhelm clients. Mitigate with debouncing.
- **Partial cancellation support:** some backends may not support cancellation; document and expose that limitation.
- **Scope creep:** adding resources/prompts too early could delay core tool stability; keep those as a later PRD.

## Recommended Next Steps

1. Implement tool list change notifications tied to toolindex changes.
2. Ensure pagination and cursor semantics are consistent across tool listing and search.
3. Add cancellation/progress wiring and document limitations.
4. Create a dedicated PRD for resources/prompts support.

---

## Appendix: Related PRD

See `docs/plans/2026-01-28-prd-015-mcp-spec-alignment.md` for an incremental implementation plan.
