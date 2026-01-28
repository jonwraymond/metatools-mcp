# metatools-mcp

[![Docs](https://img.shields.io/badge/docs-ai--tools--stack-blue)](https://jonwraymond.github.io/ai-tools-stack/)

MCP-first "metatools" server that composes the core tool libraries:

- `toolmodel`: canonical MCP-aligned tool definitions and IDs
- `toolindex`: global registry + progressive discovery (search/namespaces)
- `tooldocs`: progressive documentation tiers + examples
- `toolrun`: backend-agnostic execution + chaining
- `toolcode`: optional code-style orchestration (engine/runtime backed)
- `toolruntime`: recommended sandbox/runtime boundary for any code execution
- `toolsearch`: optional search strategies (for example, BM25)

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

## Optional toolruntime integration

`execute_code` is wired behind a build tag so the server stays minimal by
default.

Enable it locally with:

```bash
go get github.com/jonwraymond/toolruntime@v0.1.1
go run -tags toolruntime ./cmd/metatools
```

If you are developing `toolruntime` locally:

```bash
go mod edit -replace github.com/jonwraymond/toolruntime=../toolruntime
go run -tags toolruntime ./cmd/metatools
```

Notes:

- The build tag enables a `toolruntime`-backed `toolcode.Executor`.
- The default profile is `dev` and uses the unsafe host backend.
- This is intentionally dev-only until the runtime backends are hardened.

## Quickstart (server wiring)

Minimal wiring uses the adapter layer plus the official MCP Go SDK transport:

```go
package main

import (
  "context"

  "github.com/jonwraymond/metatools-mcp/internal/adapters"
  "github.com/jonwraymond/metatools-mcp/internal/server"
  "github.com/jonwraymond/tooldocs"
  "github.com/jonwraymond/toolindex"
  "github.com/jonwraymond/toolrun"
  "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
  idx := toolindex.NewInMemoryIndex()
  docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
  runner := toolrun.NewRunner(toolrun.WithIndex(idx))

  cfg := adapters.NewConfig(idx, docs, runner, nil) // executor optional
  srv, err := server.New(cfg)
  if err != nil {
    panic(err)
  }

  _ = srv.Run(context.Background(), &mcp.StdioTransport{})
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
  - tool IDs come from `toolmodel.Tool.ToolID()`
  - backend selection follows `toolindex.DefaultBackendSelector`
  - docs caps and schema derivation come from `tooldocs`
  - execution/chain semantics come from `toolrun`

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
