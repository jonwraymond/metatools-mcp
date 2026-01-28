# Examples

## Register a local tool + expose MCP server

```go
idx := toolindex.NewInMemoryIndex()

local := toolrun.NewLocalRegistry()
local.Register("ping", func(ctx context.Context, args map[string]any) (any, error) {
  return map[string]any{"ok": true}, nil
})

_ = idx.RegisterTool(toolmodel.Tool{
  Namespace: "local",
  Tool: mcp.Tool{
    Name:        "ping",
    Description: "Simple health check",
    InputSchema: map[string]any{"type": "object"},
  },
}, toolmodel.ToolBackend{
  Kind:  toolmodel.BackendKindLocal,
  Local: &toolmodel.LocalBackend{Name: "ping"},
})

runner := toolrun.NewRunner(toolrun.WithIndex(idx), toolrun.WithLocalRegistry(local))

cfg := adapters.NewConfig(idx, tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx}), runner, nil)
server, _ := server.New(cfg)
_ = server.Run(context.Background(), &mcp.StdioTransport{})
```

## Tool search and execution

```go
summaries, _ := idx.Search("ping", 3)

res, _ := runner.Run(ctx, summaries[0].ID, map[string]any{})
fmt.Println(res.Structured)
```
