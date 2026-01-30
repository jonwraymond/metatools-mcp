# Usage

## Build and run (stdio)

```bash
go run ./cmd/metatools serve
```

## CLI overview

```bash
metatools serve --transport=stdio                        # Local/Claude Desktop (default)
metatools serve --transport=streamable --port=8080       # HTTP clients (recommended)
metatools serve --transport=sse --port=8080              # Legacy HTTP clients (deprecated)
metatools version
metatools config validate --config examples/metatools.yaml
```

## Transport selection

| Transport | Use Case | Protocol |
|-----------|----------|----------|
| `stdio` | Claude Desktop, local CLI clients | stdin/stdout JSON-RPC |
| `streamable` | Web apps, REST APIs, remote clients | HTTP POST/GET/DELETE (MCP 2025-03-26) |
| `sse` | Legacy web clients | HTTP + Server-Sent Events (deprecated) |

### Streamable HTTP (recommended for HTTP)

Streamable HTTP is the MCP spec (2025-03-26) transport replacing SSE:

```bash
# Basic HTTP server
metatools serve --transport=streamable --port=8080

# With TLS
metatools serve --transport=streamable --port=443 \
  --tls --tls-cert=cert.pem --tls-key=key.pem

# Stateless mode (no session tracking)
metatools serve --transport=streamable --port=8080 --stateless
```

**Protocol flow:**
1. Client POSTs JSON-RPC to `/mcp` with `initialize` request
2. Server responds with `Mcp-Session-Id` header
3. Client includes session ID in subsequent requests
4. Client may GET `/mcp` for server notification stream
5. Client DELETEs `/mcp` to terminate session

**YAML configuration:**
```yaml
transport:
  type: streamable
  http:
    host: 0.0.0.0
    port: 8080
    tls:
      enabled: true
      cert: /path/to/cert.pem
      key: /path/to/key.pem
  streamable:
    stateless: false        # Enable session management
    json_response: false    # Use SSE streaming (default)
    session_timeout: 30m    # Clean up idle sessions
```

## Configuration files (Koanf)

Config precedence:
1. Defaults
2. Config file (`--config`)
3. Environment variables (`METATOOLS_` prefix)
4. CLI flags

Example file: `examples/metatools.yaml`

## Provider toggles

Built-in metatools can be enabled/disabled via `providers.*.enabled` in the
config file (see the `providers` block in `examples/metatools.yaml`). This
controls which MCP tools are registered at startup.

## Middleware chain

Configure optional middleware in `middleware.chain` (ordered) with per-middleware
settings under `middleware.configs`. Built-in middleware: `logging`, `metrics`.

## Enable BM25 search (build tag + env)

```bash
go build -tags toolsearch ./cmd/metatools
METATOOLS_SEARCH_STRATEGY=bm25 ./metatools
```

## Environment variables

### CLI defaults (serve command)

These map directly to `metatools serve` flags when the flags are not set:

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_PORT` | `8080` | Port for HTTP transports |
| `METATOOLS_HOST` | `0.0.0.0` | Host/interface for HTTP transports |
| `METATOOLS_CONFIG` | "" | Path to config file |

### Transport configuration (Koanf config)

These map to the config schema loaded by `config.Load`:

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT_TYPE` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_TRANSPORT_HTTP_HOST` | `0.0.0.0` | Host/interface for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_PORT` | `8080` | Port for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_TLS_ENABLED` | `false` | Enable TLS for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_TLS_CERT` | "" | TLS certificate path |
| `METATOOLS_TRANSPORT_HTTP_TLS_KEY` | "" | TLS key path |
| `METATOOLS_TRANSPORT_STREAMABLE_STATELESS` | `false` | Disable session management |
| `METATOOLS_TRANSPORT_STREAMABLE_JSON_RESPONSE` | `false` | Prefer JSON over SSE streaming |
| `METATOOLS_TRANSPORT_STREAMABLE_SESSION_TIMEOUT` | `30m` | Idle session cleanup duration |

### Runtime configuration (toolruntime build tag)

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_RUNTIME_PROFILE` | `dev` | `dev` (unsafe) or `standard` (Docker) |
| `METATOOLS_DOCKER_IMAGE` | `toolruntime-sandbox:latest` | Docker image for standard profile |
| `METATOOLS_WASM_ENABLED` | `false` | Enable WASM backend (wazero) |
| `METATOOLS_RUNTIME_BACKEND` | `docker` | Preferred standard backend: `docker` or `wasm` |

### Search configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_SEARCH_STRATEGY` | `lexical` | `lexical` or `bm25` |
| `METATOOLS_SEARCH_BM25_NAME_BOOST` | `3` | BM25 name field boost |
| `METATOOLS_SEARCH_BM25_NAMESPACE_BOOST` | `2` | BM25 namespace field boost |
| `METATOOLS_SEARCH_BM25_TAGS_BOOST` | `2` | BM25 tags field boost |
| `METATOOLS_SEARCH_BM25_MAX_DOCS` | `0` | Max docs to index (0=unlimited) |
| `METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN` | `0` | Max doc text length (0=unlimited) |
| `METATOOLS_NOTIFY_TOOL_LIST_CHANGED` | `true` | Emit `notifications/tools/list_changed` on index updates |
| `METATOOLS_NOTIFY_TOOL_LIST_CHANGED_DEBOUNCE_MS` | `150` | Debounce window for list change notifications |

## Pagination and cursors

- `search_tools` and `list_namespaces` accept `limit` (default 20, max 100) and `cursor`.
- Responses include `nextCursor` when more results are available.
- Cursor tokens are opaque and invalid cursors return JSON-RPC invalid params.

## Tool list change notifications

- `notifications/tools/list_changed` is emitted when the underlying toolindex changes.
- Notifications are debounced to avoid client spam and can be disabled with `METATOOLS_NOTIFY_TOOL_LIST_CHANGED=false`.

## Progress notifications

When callers supply a progress token, `run_tool`, `run_chain`, and `execute_code`
emit progress notifications. If the runner exposes progress callbacks, step-level
updates are forwarded; otherwise a coarse start/end signal is emitted.

## Optional toolruntime support

```bash
go run -tags toolruntime ./cmd/metatools
```

This enables `execute_code` backed by a `toolruntime` engine.
By default it uses the `dev` (unsafe) profile; set
`METATOOLS_RUNTIME_PROFILE=standard` to enable the Docker backend when
available. To use WASM instead, set `METATOOLS_WASM_ENABLED=true` and
`METATOOLS_RUNTIME_BACKEND=wasm`.
