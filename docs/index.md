# metatools-mcp

`metatools-mcp` is the MCP server that exposes the tool stack via a small,
progressive-disclosure tool surface. It composes toolmodel, toolindex, tooldocs,
toolrun, and optionally toolcode/toolruntime.

[![Docs](https://img.shields.io/badge/docs-ai--tools--stack-blue)](https://jonwraymond.github.io/ai-tools-stack/)

## Deep dives
- Design Notes: `design-notes.md`
- User Journey: `user-journey.md`

## Motivation

- **One MCP surface** for discovery, docs, and execution
- **Progressive disclosure** to keep tool context small
- **Pluggable design** for search, runtimes, and engines

## MCP tools exposed

- `search_tools`
- `list_namespaces`
- `describe_tool`
- `list_tool_examples`
- `run_tool`
- `run_chain`
- `execute_code` (optional)

## Quickstart

```go
idx := toolindex.NewInMemoryIndex()
docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
runner := toolrun.NewRunner(toolrun.WithIndex(idx))

cfg := adapters.NewConfig(idx, docs, runner, nil)
server, _ := server.New(cfg)

_ = server.Run(context.Background(), &mcp.StdioTransport{})
```

## Usability notes

- Fewer MCP tools means simpler agent prompts
- Outputs are structured and aligned to MCP schemas
- Search and execution behaviors are deterministic by default

## Next

- Server architecture: `architecture.md`
- Configuration and env vars: `usage.md`
- Examples: `examples.md`
- Design Notes: `design-notes.md`
- User Journey: `user-journey.md`

## Proposals

**Master Plan:**
- [ROADMAP](proposals/ROADMAP.md) - Master roadmap with all work streams, phases, and milestones

**Architecture:**
- [Pluggable Architecture](proposals/pluggable-architecture.md) - Extensible, modular design
- [Architecture Evaluation](proposals/architecture-evaluation.md) - Championship-level comparison
- [Component Library Analysis](proposals/component-library-analysis.md) - Tool* library ecosystem

**Features:**
- [Protocol-Agnostic Tools](proposals/protocol-agnostic-tools.md) - Composable toolsets and protocol adapters
- [Multi-Tenancy Extension](proposals/multi-tenancy.md) - Tenant isolation patterns
- Agent Skills - Higher-level capability composition (see [ROADMAP](proposals/ROADMAP.md#stream-e-agent-skills))

**Implementation:**
- [Implementation Phases](proposals/implementation-phases.md) - Phased rollout plan

