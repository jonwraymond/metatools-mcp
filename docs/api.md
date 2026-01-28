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
  Index    toolindex.Index
  Docs     tooldocs.Store
  Runner   toolrun.Runner
  Executor toolcode.Executor // optional
}
```

## MCP tool I/O types

These are exported in `pkg/metatools`:

- `SearchToolsInput` / `SearchToolsOutput`
- `ListNamespacesInput` / `ListNamespacesOutput`
- `DescribeToolInput` / `DescribeToolOutput`
- `ListToolExamplesInput` / `ListToolExamplesOutput`
- `RunToolInput` / `RunToolOutput`
- `RunChainInput` / `RunChainOutput`
- `ExecuteCodeInput` / `ExecuteCodeOutput`

## Error codes

Metatools surfaces standardized error codes (strings), including:

- `tool_not_found`
- `no_backends`
- `validation_input`
- `validation_output`
- `execution_failed`
- `stream_not_supported`
- `chain_step_failed`
- `internal`
