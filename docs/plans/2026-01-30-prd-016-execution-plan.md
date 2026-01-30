# PRD-016 Execution Plan — metatools-mcp (TDD)

**Status:** Done
**Date:** 2026-01-30
**PRD:** `2026-01-30-prd-016-interface-contracts.md`


## TDD Workflow (required)
1. Red — write failing contract tests
2. Red verification — run tests
3. Green — minimal code/doc changes
4. Green verification — run tests
5. Commit — one commit per task


## Tasks
### Task 0 — Inventory + contract outline
- Confirm interface list and method signatures.
- Draft explicit contract bullets for each interface.
- Update docs/plans/README.md with this PRD + plan.
### Task 1 — Contract tests (Red/Green)
- Add `*_contract_test.go` with tests for each interface listed below.
- Use stub implementations where needed.
### Task 2 — GoDoc contracts
- Add/expand GoDoc on each interface with explicit contract clauses (thread-safety, errors, context, ownership).
- Update README/design-notes if user-facing.
### Task 3 — Verification
- Run `go test ./...`
- Run linters if configured (golangci-lint / gosec).


## Test Skeletons (contract_test.go)
### MetricsCollector
```go
func TestMetricsCollector_Contract(t *testing.T) {
    // Methods:
    // - Start(tool string)
    // - Finish(tool string, err error, duration time.Duration)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Server
```go
func TestServer_Contract(t *testing.T) {
    // Methods:
    // - Run(ctx context.Context, transport mcp.Transport) error
    // - MCPServer() *mcp.Server
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Transport
```go
func TestTransport_Contract(t *testing.T) {
    // Methods:
    // - Name() string
    // - Info() Info
    // - Serve(ctx context.Context, server Server) error
    // - Close() error
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### ToolProvider
```go
func TestToolProvider_Contract(t *testing.T) {
    // Methods:
    // - Name() string
    // - Enabled() bool
    // - Tool() mcp.Tool
    // - Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### ConfigurableProvider
```go
func TestConfigurableProvider_Contract(t *testing.T) {
    // Methods:
    // - ToolProvider
    // - Configure(cfg map[string]any) error
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### StreamingProvider
```go
func TestStreamingProvider_Contract(t *testing.T) {
    // Methods:
    // - ToolProvider
    // - HandleStream(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (<-chan any, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Backend
```go
func TestBackend_Contract(t *testing.T) {
    // Methods:
    // - Kind() string
    // - Name() string
    // - Enabled() bool
    // - ListTools(ctx context.Context) ([]toolmodel.Tool, error)
    // - Execute(ctx context.Context, tool string, args map[string]any) (any, error)
    // - Start(ctx context.Context) error
    // - Stop() error
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### ConfigurableBackend
```go
func TestConfigurableBackend_Contract(t *testing.T) {
    // Methods:
    // - Backend
    // - Configure(raw []byte) error
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### StreamingBackend
```go
func TestStreamingBackend_Contract(t *testing.T) {
    // Methods:
    // - Backend
    // - ExecuteStream(ctx context.Context, tool string, args map[string]any) (<-chan any, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Index
```go
func TestIndex_Contract(t *testing.T) {
    // Methods:
    // - SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error)
    // - ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Store
```go
func TestStore_Contract(t *testing.T) {
    // Methods:
    // - DescribeTool(ctx context.Context, id string, level string) (ToolDoc, error)
    // - ListExamples(ctx context.Context, id string, maxExamples int) ([]metatools.ToolExample, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Runner
```go
func TestRunner_Contract(t *testing.T) {
    // Methods:
    // - Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error)
    // - RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### ProgressRunner
```go
func TestProgressRunner_Contract(t *testing.T) {
    // Methods:
    // - RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(ProgressEvent)) (RunResult, error)
    // - RunChainWithProgress(ctx context.Context, steps []ChainStep, onProgress func(ProgressEvent)) (RunResult, []StepResult, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Executor
```go
func TestExecutor_Contract(t *testing.T) {
    // Methods:
    // - ExecuteCode(ctx context.Context, params ExecuteParams) (ExecuteResult, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
