# metatools-mcp

MCP server that exposes the tool stack via standardized MCP tools.

## What this repo provides

- MCP tool wiring for search, docs, run, and code execution
- Official MCP Go SDK integration

## Example

```go
srv := metatools.NewServer(cfg)
_ = srv.Run(context.Background(), &mcp.StdioTransport{})
```
