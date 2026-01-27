# metatools-mcp

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

## Dependency versions (current module)

From `go.mod`:

- `github.com/jonwraymond/toolmodel` `v0.1.0`
- `github.com/jonwraymond/toolindex` `v0.1.2`
- `github.com/jonwraymond/tooldocs` `v0.1.2`
- `github.com/jonwraymond/toolrun` `v0.1.1`
- `github.com/jonwraymond/toolcode` `v0.1.1`
- `github.com/jonwraymond/toolruntime` `v0.1.1` (when built with `-tags toolruntime`)
- `github.com/jonwraymond/toolsearch` `v0.1.1` (optional)
- `github.com/modelcontextprotocol/go-sdk` `v1.2.0`

## CI

CI runs on pushes and pull requests to `main` and enforces:

- `go mod download`
- `go vet ./...`
- `go test ./...`
