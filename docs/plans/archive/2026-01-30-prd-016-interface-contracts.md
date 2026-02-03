# PRD-016 Interface Contracts — metatools-mcp

**Status:** Done
**Date:** 2026-01-30


## Overview
Define explicit interface contracts (GoDoc + documented semantics) for all interfaces in this repo. Contracts must state concurrency guarantees, error semantics, ownership of inputs/outputs, and context handling.


## Goals
- Every interface has explicit GoDoc describing behavioral contract.
- Contract behavior is codified in tests (contract tests).
- Docs/README updated where behavior is user-facing.


## Non-Goals
- No API shape changes unless required to satisfy the contract tests.
- No new features beyond contract clarity and tests.


## Interface Inventory
| Interface | File | Methods |
| --- | --- | --- |
| `MetricsCollector` | `metatools-mcp/internal/middleware/metrics.go:22` | Start(tool string)<br/>Finish(tool string, err error, duration time.Duration) |
| `Server` | `metatools-mcp/internal/transport/transport.go:10` | Run(ctx context.Context, transport mcp.Transport) error<br/>MCPServer() *mcp.Server |
| `Transport` | `metatools-mcp/internal/transport/transport.go:23` | Name() string<br/>Info() Info<br/>Serve(ctx context.Context, server Server) error<br/>Close() error |
| `ToolProvider` | `metatools-mcp/internal/provider/provider.go:11` | Name() string<br/>Enabled() bool<br/>Tool() mcp.Tool<br/>Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) |
| `ConfigurableProvider` | `metatools-mcp/internal/provider/provider.go:27` | ToolProvider<br/>Configure(cfg map[string]any) error |
| `StreamingProvider` | `metatools-mcp/internal/provider/provider.go:35` | ToolProvider<br/>HandleStream(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (<-chan any, error) |
| `Backend` | `toolexec/backend/backend.go:20` | Kind() string<br/>Name() string<br/>Enabled() bool<br/>ListTools(ctx context.Context) ([]toolmodel.Tool, error)<br/>Execute(ctx context.Context, tool string, args map[string]any) (any, error)<br/>Start(ctx context.Context) error<br/>Stop() error |
| `ConfigurableBackend` | `toolexec/backend/backend.go:44` | Backend<br/>Configure(raw []byte) error |
| `StreamingBackend` | `toolexec/backend/backend.go:51` | Backend<br/>ExecuteStream(ctx context.Context, tool string, args map[string]any) (<-chan any, error) |
| `Index` | `metatools-mcp/internal/handlers/interfaces.go:10` | SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error)<br/>ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error) |
| `Store` | `metatools-mcp/internal/handlers/interfaces.go:26` | DescribeTool(ctx context.Context, id string, level string) (ToolDoc, error)<br/>ListExamples(ctx context.Context, id string, maxExamples int) ([]metatools.ToolExample, error) |
| `Runner` | `metatools-mcp/internal/handlers/interfaces.go:64` | Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error)<br/>RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error) |
| `ProgressRunner` | `metatools-mcp/internal/handlers/interfaces.go:70` | RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(ProgressEvent)) (RunResult, error)<br/>RunChainWithProgress(ctx context.Context, steps []ChainStep, onProgress func(ProgressEvent)) (RunResult, []StepResult, error) |
| `Executor` | `metatools-mcp/internal/handlers/interfaces.go:92` | ExecuteCode(ctx context.Context, params ExecuteParams) (ExecuteResult, error) |

## Contract Template (apply per interface)
- **Thread-safety:** explicitly state if safe for concurrent use.
- **Context:** cancellation/deadline handling (if context is a parameter).
- **Errors:** classification, retryability, and wrapping expectations.
- **Ownership:** who owns/allocates inputs/outputs; mutation expectations.
- **Determinism/order:** ordering guarantees for returned slices/maps/streams.
- **Nil/zero handling:** behavior for nil inputs or empty values.


## Acceptance Criteria
- All interfaces have GoDoc with explicit behavioral contract.
- Contract tests exist and pass.
- No interface contract contradictions across repos.
