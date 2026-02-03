# Design Notes

This page documents the tradeoffs and error semantics behind `metatools-mcp`.

## Design tradeoffs

- **MCP-native surface.** All metatools (search, describe, run, chain, execute_code) are exposed via the official MCP Go SDK types to keep wire compatibility.
- **Adapters, not re-implementation.** The server delegates to tooldiscovery/index, tooldiscovery/tooldoc, toolexec/run, and toolexec/code via thin adapters so the libraries remain the source of truth.
- **Structured error objects.** Tool-level errors are returned in a consistent `ErrorObject` shape rather than raw Go errors, preserving the MCP tool contract.
- **Explicit limits.** Inputs such as `limit` and `max` are capped for safe defaults (e.g., search limit cap 100, examples cap 5).
- **Opaque pagination.** Cursor tokens are opaque and validated against index mutations to prevent stale paging.
- **Pluggable search.** BM25 is optional via build tags (`toolsearch`) and runtime config via env vars.
- **Change notifications.** Tool list updates emit `notifications/tools/list_changed` with a debounce window; notifications can be disabled and are emitted as a single list change per debounce window.
- **Transport abstraction.** The `Transport` interface decouples protocol handling from server logic, enabling stdio, SSE, and Streamable HTTP without code changes.
- **Runtime isolation.** `execute_code` is optional; the `toolruntime` build tag enables sandboxed execution via toolexec/runtime with runtime profile selection.

## Transport layer

The transport layer abstracts how clients connect to the MCP server. All transports
implement the same `Transport` interface, ensuring identical behavior regardless of
protocol:

```go
type Transport interface {
    Name() string
    Info() Info
    Serve(ctx context.Context, server Server) error
    Close() error
}
```

### Transport selection rationale

| Transport | Status | Use Case | Rationale |
|-----------|--------|----------|-----------|
| `stdio` | **Recommended** (local) | Claude Desktop, local CLIs | Zero config, implicit session, lowest latency |
| `streamable` | **Recommended** (HTTP) | Web apps, remote clients | MCP spec 2025-11-25 compliant, session management, bidirectional |
| `sse` | **Deprecated** | Legacy web clients | Superseded by streamable per MCP spec |

### Streamable HTTP design decisions

1. **Single endpoint (`/mcp`):** Follows MCP spec 2025-11-25 with POST/GET/DELETE methods
   on one path, simplifying routing and CORS configuration.

2. **Session management via header:** Uses `Mcp-Session-Id` header (not cookies) for
   stateless load balancing compatibility and explicit session lifecycle.

3. **Stateless mode option:** Enables serverless/FaaS deployments where session
   persistence is impractical. Trade-off: no server-initiated requests.

4. **JSON vs SSE response modes:** Default SSE streaming supports long-running tools;
   `JSONResponse=true` option for simpler request/response patterns.

5. **TLS built-in:** Direct TLS support avoids reverse proxy requirements for simple
   deployments while allowing proxy termination for complex setups.

6. **Graceful shutdown:** 5-second timeout allows in-flight requests to complete,
   balancing responsiveness with reliability.

## Error semantics

`metatools-mcp` distinguishes protocol errors from tool errors:

- **Protocol errors** (invalid input) return a non-nil error from handlers.
- **Tool errors** are wrapped into `ErrorObject` and returned with `isError = true` so MCP clients treat them as tool failures.

Key error behaviors:

- `run_tool` rejects `stream=true` and `backend_override` in the default handler (not supported yet).
- `run_chain` stops on first error and returns partial results with an `ErrorObject`.
- `describe_tool`/`list_tool_examples` return validation errors when required fields are missing.
- Invalid cursors return JSON-RPC invalid params.
- Cancellation and timeouts map to `cancelled` and `timeout` error codes.

## Extension points

- **Transport:** implement `Transport` interface to add new protocols (e.g., WebSocket, gRPC).
- **Search strategy:** enable BM25 via the `toolsearch` build tag and env vars.
- **Tool execution:** swap `toolexec/run` runner implementation or configure different backends.
- **Code execution:** plug in a different `toolexec/code` engine (e.g., toolexec/runtime-backed).
- **Progress:** when a progress token is provided, `run_tool`, `run_chain`, and `execute_code` emit progress notifications. If the runner supports progress callbacks, step-level updates are forwarded; otherwise a coarse start/end signal is sent.

## Runtime profile selection

When built with `-tags toolruntime`, metatools-mcp wires `toolexec/runtime` into
`toolexec/code`:

- **Dev profile (`dev`)** uses the unsafe subprocess backend for fast iteration.
- **Standard profile (`standard`)** uses Docker by default or WASM when selected.
- Set `METATOOLS_RUNTIME_PROFILE=standard` to opt into standard isolation.
- Use `METATOOLS_RUNTIME_BACKEND=wasm` with `METATOOLS_WASM_ENABLED=true` to
  prefer the WASM backend (wazero) for standard profile.
- If Docker is unavailable and WASM is enabled, the server falls back to WASM
  for the standard profile.

## Operational guidance

### Transport configuration

- Use `--transport=stdio` (default) for Claude Desktop and local CLI integration.
- Use `--transport=streamable --port=8080` for HTTP-based clients and web applications.
- Configure TLS for production HTTP deployments: `--tls --tls-cert=cert.pem --tls-key=key.pem`
- Use `--stateless` for serverless/FaaS where session persistence is unavailable.
- Set `METATOOLS_TRANSPORT_STREAMABLE_SESSION_TIMEOUT` to control idle session cleanup.

### Search configuration

- Use environment variables to configure search strategy:
  - `METATOOLS_SEARCH_STRATEGY=lexical|bm25`
  - `METATOOLS_SEARCH_BM25_*` for weighting and caps

### General guidance

- Keep tool schemas in `toolfoundation/model` to preserve MCP compatibility end-to-end.
- Treat metatools as the stable surface; update libraries behind it as needed.
