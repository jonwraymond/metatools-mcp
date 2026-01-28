# metatools-mcp

`metatools-mcp` is the MCP server that exposes the tool stack via a small,
progressive-disclosure tool surface. It composes toolmodel, toolindex, tooldocs,
toolrun, and optionally toolcode/toolruntime.

## What this server provides

- MCP tools: `search_tools`, `list_namespaces`, `describe_tool`, `list_tool_examples`,
  `run_tool`, `run_chain`, and optional `execute_code`
- Official MCP Go SDK integration
- Configurable search strategy (lexical or BM25)

## Quickstart

```go
idx := toolindex.NewInMemoryIndex()
docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
runner := toolrun.NewRunner(toolrun.WithIndex(idx))

cfg := adapters.NewConfig(idx, docs, runner, nil)
server, _ := server.New(cfg)

_ = server.Run(context.Background(), &mcp.StdioTransport{})
```

## Next

- Server architecture: `architecture.md`
- Configuration and env vars: `usage.md`
- Examples: `examples.md`
