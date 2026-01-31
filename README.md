# metatools-mcp

[![Docs](https://img.shields.io/badge/docs-ai--tools--stack-blue)](https://jonwraymond.github.io/ai-tools-stack/)

MCP-first "metatools" server that composes the consolidated tool libraries:

- `toolfoundation/model`: canonical MCP-aligned tool definitions and IDs
- `toolfoundation/adapter`: protocol-agnostic format conversion
- `tooldiscovery/index`: global registry + progressive discovery (search/namespaces)
- `tooldiscovery/tooldoc`: progressive documentation tiers + examples
- `tooldiscovery/search`: optional search strategies (e.g., BM25)
- `toolexec/run`: backend-agnostic execution + chaining
- `toolexec/code`: optional code-style orchestration (engine/runtime backed)
- `toolexec/runtime`: recommended sandbox/runtime boundary for any code execution
- `toolops/observe`: optional observability middleware
- `toolops/cache`: optional caching middleware

This server exposes a small, opinionated MCP tool surface that optimizes
progressive disclosure:

1) discover tools cheaply,
2) inspect only what you need, then
3) execute with consistent error semantics.

## MCP tools exposed

- `search_tools`
- `list_namespaces`
- `describe_tool`
- `list_tool_examples`
- `run_tool`
- `run_chain`
- `execute_code` (only when an executor is injected)

## Changelog

See `CHANGELOG.md` for release notes.

## Transport selection

metatools-mcp supports multiple transports for different deployment scenarios:

| Transport | Command | Use Case |
|-----------|---------|----------|
| `stdio` | `metatools serve` | Claude Desktop, local CLI (default) |
| `streamable` | `metatools serve --transport=streamable --port=8080` | Web apps, remote clients (recommended for HTTP) |
| `sse` | `metatools serve --transport=sse --port=8080` | Legacy web clients (deprecated) |

**Streamable HTTP** implements MCP spec 2025-03-26 with session management via
`Mcp-Session-Id` header, supporting both SSE streaming and JSON responses.

```bash
# Basic HTTP server
metatools serve --transport=streamable --port=8080

# With TLS
metatools serve --transport=streamable --port=443 --tls --tls-cert=cert.pem --tls-key=key.pem

# Stateless mode (serverless/FaaS)
metatools serve --transport=streamable --port=8080 --stateless
```

See `docs/usage.md` for full configuration options.

## Search Strategy

By default, metatools-mcp uses lexical search. For BM25 ranking:

1. Build with the toolsearch tag:
   ```bash
   go build -tags toolsearch ./cmd/metatools
   ```

2. Set the environment variable:
   ```bash
   METATOOLS_SEARCH_STRATEGY=bm25 ./metatools
   ```

Notes:

- BM25 requires the `toolsearch` build tag. If you set
  `METATOOLS_SEARCH_STRATEGY=bm25` without it, the server now fails fast.
- `METATOOLS_SEARCH_STRATEGY` is case-insensitive (for example, `BM25` works).

### Environment Variables

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

### Transport Environment (CLI defaults)

These are applied by `metatools serve` when flags are not explicitly set:

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_PORT` | `8080` | HTTP port for `streamable`/`sse` |
| `METATOOLS_HOST` | `0.0.0.0` | HTTP bind host |
| `METATOOLS_CONFIG` | "" | Path to config file |

### Transport Environment (Koanf config)

These are consumed by `config.Load` via Koanf (file/env/flags precedence):

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT_TYPE` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_TRANSPORT_HTTP_HOST` | `0.0.0.0` | HTTP bind host |
| `METATOOLS_TRANSPORT_HTTP_PORT` | `8080` | HTTP port |
| `METATOOLS_TRANSPORT_HTTP_TLS_ENABLED` | `false` | Enable TLS for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_TLS_CERT` | "" | TLS certificate path |
| `METATOOLS_TRANSPORT_HTTP_TLS_KEY` | "" | TLS key path |
| `METATOOLS_TRANSPORT_STREAMABLE_STATELESS` | `false` | Disable session management |
| `METATOOLS_TRANSPORT_STREAMABLE_JSON_RESPONSE` | `false` | Prefer JSON over SSE streaming |
| `METATOOLS_TRANSPORT_STREAMABLE_SESSION_TIMEOUT` | `30m` | Idle session cleanup duration |

### Runtime Environment (toolruntime build tag)

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_RUNTIME_PROFILE` | `dev` | `dev` (unsafe) or `standard` (Docker/WASM) |
| `METATOOLS_DOCKER_IMAGE` | `toolruntime-sandbox:latest` | Docker image for standard profile |
| `METATOOLS_WASM_ENABLED` | `false` | Enable WASM backend (wazero) |
| `METATOOLS_RUNTIME_BACKEND` | `docker` | Preferred standard backend: `docker` or `wasm` |

## Optional toolexec/runtime integration

`execute_code` is wired behind a build tag so the server stays minimal by
default.

Enable it locally with:

```bash
go get github.com/jonwraymond/toolexec@latest
go run -tags toolruntime ./cmd/metatools
```

If you are developing `toolexec` locally:

```bash
go mod edit -replace github.com/jonwraymond/toolexec=../toolexec
go run -tags toolruntime ./cmd/metatools
```

Notes:

- The build tag enables a `toolexec/runtime`-backed `toolexec/code.Executor`.
- Default profile is `dev` (unsafe subprocess backend).
- If Docker is available, set `METATOOLS_RUNTIME_PROFILE=standard` to enable
  the hardened Docker backend.
- Override the Docker image with `METATOOLS_DOCKER_IMAGE` (default:
  `toolruntime-sandbox:latest`).
- Enable the WASM backend with `METATOOLS_WASM_ENABLED=true` and select it
  with `METATOOLS_RUNTIME_BACKEND=wasm` (uses wazero).

## Quickstart (server wiring)

Minimal wiring uses the adapter layer plus the internal transport package:

```go
package main

import (
  "context"

  "github.com/jonwraymond/metatools-mcp/internal/adapters"
  "github.com/jonwraymond/metatools-mcp/internal/server"
  "github.com/jonwraymond/metatools-mcp/internal/transport"
  "github.com/jonwraymond/tooldiscovery/index"
  "github.com/jonwraymond/tooldiscovery/tooldoc"
  "github.com/jonwraymond/toolexec/run"
)

func main() {
  idx := index.NewInMemoryIndex()
  docs := tooldoc.NewInMemoryStore(tooldoc.StoreOptions{Index: idx})
  runner := run.NewRunner(run.WithIndex(idx))

  cfg := adapters.NewConfig(idx, docs, runner, nil) // executor optional
  srv, err := server.New(cfg)
  if err != nil {
    panic(err)
  }

  // Use stdio for local/CLI clients
  tr := &transport.StdioTransport{}
  _ = tr.Serve(context.Background(), srv)

  // Or use streamable HTTP for web clients
  // tr := &transport.StreamableHTTPTransport{
  //   Config: transport.StreamableHTTPConfig{Port: 8080},
  // }
  // _ = tr.Serve(context.Background(), srv)
}
```

See a full working example (including a local tool + docs registration) in
`examples/basic/main.go`.

## How it fits together

- The MCP surface is implemented via the official SDK (`mcp.NewServer`,
  `mcp.AddTool`, `mcp.CallToolResult{IsError: ...}`).
- `internal/adapters` bridges the public tool libraries into handler-facing
  interfaces without leaking protocol details into the libraries.
- The server does not bypass library policy:
  - tool IDs come from `toolfoundation/model.Tool.ToolID()`
  - backend selection follows `tooldiscovery/index.DefaultBackendSelector`
  - docs caps and schema derivation come from `tooldiscovery/tooldoc`
  - execution/chain semantics come from `toolexec/run`

## Documentation

- `docs/index.md` — overview
- `docs/design-notes.md` — tradeoffs and error semantics
- `docs/user-journey.md` — end-to-end agent workflow

## Version compatibility

See `VERSIONS.md` for the authoritative, auto-generated compatibility matrix.

## CI

CI runs on pushes and pull requests to `main` and enforces:

- `go mod download`
- `go vet ./...`
- `go test ./...`
