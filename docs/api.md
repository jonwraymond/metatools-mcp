# API Reference

## Server

```go
func New(cfg config.Config) (*Server, error)
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error
func (s *Server) MCPServer() *mcp.Server
func (s *Server) ListTools() []*mcp.Tool
```

## Config

```go
type Config struct {
  Index    index.Index
  Docs     tooldoc.Store
  Runner   run.Runner
  Executor code.Executor // optional

  NotifyToolListChanged           bool
  NotifyToolListChangedDebounceMs int
}
```

## Interface Contracts

- **Transport**: thread-safe; `Serve` honors context; `Close` is idempotent.
- **Backend**: thread-safe; `ListTools`/`Execute` honor context; streaming backends must return non-nil channel when err is nil.
- **ToolProvider**: thread-safe; `Handle` honors context; streaming providers must return non-nil channel when err is nil.

## Transport

The transport layer is in `internal/transport`:

```go
// Transport interface for MCP protocol transports
type Transport interface {
  Name() string
  Info() Info
  Serve(ctx context.Context, server Server) error
  Close() error
}

// Info describes a transport instance
type Info struct {
  Name string
  Addr string
  Path string
}
```

### Available transports

| Type | Struct | Use Case |
|------|--------|----------|
| stdio | `StdioTransport` | Claude Desktop, local CLI |
| streamable | `StreamableHTTPTransport` | Web apps, HTTP clients (MCP 2025-11-25) |
| sse | `SSETransport` | Legacy HTTP clients (deprecated) |

### StreamableHTTPConfig

```go
type StreamableHTTPConfig struct {
  Host              string        // Network interface (default: "0.0.0.0")
  Port              int           // TCP port (required)
  Path              string        // Endpoint path (default: "/mcp")
  ReadHeaderTimeout time.Duration // Header read timeout (default: 10s)
  TLS               TLSConfig     // HTTPS configuration
  Stateless         bool          // Disable session management
  JSONResponse      bool          // Prefer JSON over SSE streaming
  SessionTimeout    time.Duration // Idle session cleanup
}

type TLSConfig struct {
  Enabled  bool
  CertFile string
  KeyFile  string
}
```

## Toolruntime integration

When built with `-tags toolruntime`, `execute_code` is backed by a
`toolexec/runtime` runtime. The runtime selects a profile at startup:

- `dev` profile: unsafe subprocess backend (default).
- `standard` profile: Docker sandbox by default or WASM when selected.
- `METATOOLS_DOCKER_IMAGE` overrides the sandbox image name.
- `METATOOLS_WASM_ENABLED=true` enables the WASM backend (wazero).
- `METATOOLS_RUNTIME_BACKEND=wasm` selects WASM for the standard profile.

## MCP tool I/O types

These are exported in `pkg/metatools`:

- `SearchToolsInput` / `SearchToolsOutput`
- `ListNamespacesInput` / `ListNamespacesOutput`
- `DescribeToolInput` / `DescribeToolOutput`
- `ListToolExamplesInput` / `ListToolExamplesOutput`
- `RunToolInput` / `RunToolOutput`
- `RunChainInput` / `RunChainOutput`
- `ExecuteCodeInput` / `ExecuteCodeOutput`

Notes:
- `SearchToolsInput`/`ListNamespacesInput` accept `limit` + `cursor`.
- `SearchToolsOutput`/`ListNamespacesOutput` return `nextCursor` when more data exists.

## Error codes

Metatools surfaces standardized error codes (strings), including:

- `tool_not_found`
- `no_backends`
- `validation_input`
- `validation_output`
- `execution_failed`
- `stream_not_supported`
- `chain_step_failed`
- `cancelled`
- `timeout`
- `internal`
