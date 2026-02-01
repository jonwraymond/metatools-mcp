# Examples

## Register a local tool + expose MCP server

```go
type localRegistry struct {
  handlers map[string]run.LocalHandler
}

func newLocalRegistry() *localRegistry {
  return &localRegistry{handlers: make(map[string]run.LocalHandler)}
}

func (r *localRegistry) Get(name string) (run.LocalHandler, bool) {
  h, ok := r.handlers[name]
  return h, ok
}

func (r *localRegistry) Register(name string, h run.LocalHandler) {
  r.handlers[name] = h
}

idx := index.NewInMemoryIndex()

local := newLocalRegistry()
local.Register("ping", func(ctx context.Context, args map[string]any) (any, error) {
  return map[string]any{"ok": true}, nil
})

_ = idx.RegisterTool(model.Tool{
  Namespace: "local",
  Tool: mcp.Tool{
    Name:        "ping",
    Description: "Simple health check",
    InputSchema: map[string]any{"type": "object"},
  },
}, model.ToolBackend{
  Kind:  model.BackendKindLocal,
  Local: &model.LocalBackend{Name: "ping"},
})

runner := run.NewRunner(run.WithIndex(idx), run.WithLocalRegistry(local))

cfg := adapters.NewConfig(idx, tooldoc.NewInMemoryStore(tooldoc.StoreOptions{Index: idx}), runner, nil)
server, _ := server.New(cfg)
_ = server.Run(context.Background(), &mcp.StdioTransport{})
```

## Tool search and execution

```go
summaries, _ := idx.Search("ping", 3)

res, _ := runner.Run(ctx, summaries[0].ID, map[string]any{})
fmt.Println(res.Structured)
```
