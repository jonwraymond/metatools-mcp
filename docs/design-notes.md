# Design Notes

This page documents the tradeoffs and error semantics behind `metatools-mcp`.

## Design tradeoffs

- **MCP-native surface.** All metatools (search, describe, run, chain, execute_code) are exposed via the official MCP Go SDK types to keep wire compatibility.
- **Adapters, not re-implementation.** The server delegates to toolindex/tooldocs/toolrun/toolcode via thin adapters so the libraries remain the source of truth.
- **Structured error objects.** Tool-level errors are returned in a consistent `ErrorObject` shape rather than raw Go errors, preserving the MCP tool contract.
- **Explicit limits.** Inputs such as `limit` and `max` are capped for safe defaults (e.g., search limit cap 100, examples cap 5).
- **Opaque pagination.** Cursor tokens are opaque and validated against index mutations to prevent stale paging.
- **Pluggable search.** BM25 is optional via build tags (`toolsearch`) and runtime config via env vars.
- **Change notifications.** Tool list updates emit `notifications/tools/list_changed` with a debounce window and can be disabled.

## Error semantics

`metatools-mcp` distinguishes protocol errors from tool errors:

- **Protocol errors** (invalid input) return a non-nil error from handlers.
- **Tool errors** are wrapped into `ErrorObject` and returned with `isError = true` so MCP clients treat them as tool failures.

Key error behaviors:

- `run_tool` rejects `stream=true` and `backend_override` in the default handler (not supported yet).
- `run_chain` stops on first error and returns partial results with an `ErrorObject`.
- `describe_tool`/`list_tool_examples` return validation errors when required fields are missing.
- Invalid cursors return JSON-RPC invalid params.

## Extension points

- **Search strategy:** enable BM25 via the `toolsearch` build tag and env vars.
- **Tool execution:** swap `toolrun` runner implementation or configure different backends.
- **Code execution:** plug in a different `toolcode.Engine` (e.g., toolruntime-backed).
- **Progress:** progress notifications are not emitted yet for non-streaming calls (runner limitation).

## Operational guidance

- Use environment variables to configure search strategy:
  - `METATOOLS_SEARCH_STRATEGY=lexical|bm25`
  - `METATOOLS_SEARCH_BM25_*` for weighting and caps
- Keep tool schemas in `toolmodel` to preserve MCP compatibility end-to-end.
- Treat metatools as the stable surface; update libraries behind it as needed.
