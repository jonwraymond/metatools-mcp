# Component Library Analysis for Pluggable Architecture

**Status:** Draft
**Date:** 2026-01-27
**Related:** [Pluggable Architecture Proposal](./pluggable-architecture.md), [Implementation Phases](./implementation-phases.md)

## Overview

This document analyzes the metatools component library ecosystem and identifies changes needed to support the pluggable architecture. The analysis follows Go Architect principles: layered architecture, clean interfaces, dependency injection, and proper error handling.

---

## Component Library Ecosystem

### Dependency Graph

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         METATOOLS COMPONENT ECOSYSTEM                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│                           ┌─────────────────┐                                │
│                           │  metatools-mcp  │                                │
│                           │    (v0.1.x)     │                                │
│                           └────────┬────────┘                                │
│                                    │                                          │
│         ┌──────────────────────────┼──────────────────────────┐              │
│         │                          │                          │              │
│         ▼                          ▼                          ▼              │
│  ┌─────────────┐           ┌─────────────┐           ┌─────────────┐        │
│  │  toolcode   │           │  tooldocs   │           │  toolrun    │        │
│  │  (v0.1.10)  │           │  (v0.1.10)  │           │  (v0.1.9)   │        │
│  └──────┬──────┘           └──────┬──────┘           └──────┬──────┘        │
│         │                         │                         │                │
│         │    ┌────────────────────┼─────────────────────────┘                │
│         │    │                    │                                          │
│         ▼    ▼                    ▼                                          │
│  ┌─────────────┐           ┌─────────────┐                                  │
│  │ toolruntime │           │  toolindex  │◄─────────────────────────┐       │
│  │  (v0.1.10)  │           │  (v0.1.8)   │                          │       │
│  └──────┬──────┘           └──────┬──────┘                   ┌──────┴─────┐ │
│         │                         │                          │ toolsearch │ │
│         │                         │                          │  (v0.1.9)  │ │
│         │                         ▼                          └────────────┘ │
│         │                  ┌─────────────┐                                  │
│         └─────────────────▶│  toolmodel  │                                  │
│                            │  (v0.1.2)   │                                  │
│                            └──────┬──────┘                                  │
│                                   │                                          │
│                                   ▼                                          │
│                     ┌─────────────────────────┐                             │
│                     │ modelcontextprotocol/   │                             │
│                     │       go-sdk            │                             │
│                     │       (v1.2.0)          │                             │
│                     └─────────────────────────┘                             │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Library Summary Matrix

| Library | Version | Purpose | Key Interfaces | Dependencies |
|---------|---------|---------|----------------|--------------|
| **toolmodel** | v0.1.2 | Core data models | `Tool`, `ToolBackend`, `SchemaValidator` | mcp go-sdk |
| **toolindex** | v0.1.8 | Tool registry | `Index`, `Searcher`, `BackendSelector` | toolmodel |
| **tooldocs** | v0.1.10 | Documentation | `Store`, `DetailLevel` | toolindex, toolmodel |
| **toolrun** | v0.1.9 | Execution | `Runner`, `MCPExecutor`, `ProviderExecutor`, `LocalRegistry` | toolindex, toolmodel |
| **toolcode** | v0.1.10 | Code execution | `Executor`, `Engine`, `Tools` | toolrun, tooldocs, toolindex |
| **toolsearch** | v0.1.9 | BM25 search | `BM25Searcher` (implements `Searcher`) | toolindex, bleve |
| **toolruntime** | v0.1.10 | Sandbox runtime | `Runtime`, `Backend`, `ToolGateway` | toolcode, toolrun, tooldocs |

---

## Layer 1: toolmodel (Foundation)

### Current State

The foundational library defining what a "tool" is. Zero networking dependencies, safe for embedding.

**Exported Types:**
- `Tool` - Embeds `mcp.Tool` with extensions (Namespace, Version, Tags)
- `ToolBackend` - Backend binding (Kind, MCP/Provider/Local configs)
- `BackendKind` - Enum: `mcp`, `provider`, `local`
- `MCPBackend`, `ProviderBackend`, `LocalBackend` - Backend configs
- `SchemaValidator` - Interface for JSON Schema validation
- `DefaultValidator` - Default implementation using jsonschema-go

**Key Functions:**
- `Tool.ToolID()` - Returns canonical ID (`namespace:name`)
- `ParseToolID()` - Parses ID string into components
- `Tool.Validate()` - Validates tool invariants
- `NormalizeTags()` - Tag normalization for indexing

### Changes Needed for Pluggable Architecture

#### Priority: LOW (mostly stable)

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add `ToolMetadata` field** | Support arbitrary metadata for middleware/backends | Minor - additive |
| **Add `BackendKindHTTP`** | Support HTTP backend type for remote APIs | Minor - additive |
| **Add `BackendKindGRPC`** | Support gRPC backend type | Minor - additive |
| **Add `HTTPBackend` struct** | Config for HTTP backends | Minor - additive |

#### Proposed Additions

```go
// New backend kinds for multi-backend architecture
const (
    BackendKindMCP      BackendKind = "mcp"
    BackendKindProvider BackendKind = "provider"
    BackendKindLocal    BackendKind = "local"
    BackendKindHTTP     BackendKind = "http"      // NEW
    BackendKindGRPC     BackendKind = "grpc"      // NEW
)

// HTTPBackend metadata for HTTP API backends
type HTTPBackend struct {
    BaseURL   string            `json:"baseUrl"`
    AuthType  string            `json:"authType,omitempty"`  // bearer, oauth2, apikey
    Headers   map[string]string `json:"headers,omitempty"`
    Timeout   time.Duration     `json:"timeout,omitempty"`
}

// GRPCBackend metadata for gRPC backends
type GRPCBackend struct {
    Address   string `json:"address"`
    TLS       bool   `json:"tls,omitempty"`
    CACert    string `json:"caCert,omitempty"`
}

// ToolBackend - add HTTP and GRPC fields
type ToolBackend struct {
    Kind     BackendKind      `json:"kind"`
    MCP      *MCPBackend      `json:"mcp,omitempty"`
    Provider *ProviderBackend `json:"provider,omitempty"`
    Local    *LocalBackend    `json:"local,omitempty"`
    HTTP     *HTTPBackend     `json:"http,omitempty"`      // NEW
    GRPC     *GRPCBackend     `json:"grpc,omitempty"`      // NEW
}

// Tool - add metadata field
type Tool struct {
    mcp.Tool
    Namespace string         `json:"namespace,omitempty"`
    Version   string         `json:"version,omitempty"`
    Tags      []string       `json:"tags,omitempty"`
    Metadata  map[string]any `json:"metadata,omitempty"`   // NEW
}
```

---

## Layer 2: toolindex (Registry)

### Current State

Global registry and discovery layer for tools. Thread-safe, supports pluggable search.

**Exported Types:**
- `Index` - Interface for tool registry
- `InMemoryIndex` - Default implementation
- `Summary` - Lightweight search result
- `SearchDoc` - Document for search indexing
- `Searcher` - Interface for search implementations
- `BackendSelector` - Function type for backend selection
- `ToolRegistration` - Tool + Backend pair

**Key Methods:**
- `RegisterTool()`, `RegisterTools()`, `RegisterToolsFromMCP()`
- `GetTool()`, `GetAllBackends()`
- `Search()`, `ListNamespaces()`
- `UnregisterBackend()`

### Changes Needed for Pluggable Architecture

#### Priority: MEDIUM

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add `ListTools()` method** | Backend aggregation needs full tool list | Minor - additive |
| **Add `BackendSource` tracking** | Track which backend registered a tool | Medium - structural |
| **Add `Refresh()` capability** | Support hot-reload from config changes | Medium - additive |
| **Add `OnChange` callback** | Notify listeners of registry changes | Minor - additive |

#### Proposed Interface Extensions

```go
// Index interface additions
type Index interface {
    // Existing methods...
    RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error
    RegisterTools(regs []ToolRegistration) error
    RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error
    UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error
    GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)
    GetAllBackends(id string) ([]toolmodel.ToolBackend, error)
    Search(query string, limit int) ([]Summary, error)
    ListNamespaces() ([]string, error)

    // NEW: Support for multi-backend architecture
    ListTools() ([]toolmodel.Tool, error)                    // List all registered tools
    ListToolsFromBackend(backendName string) ([]Summary, error) // Filter by source

    // NEW: Support for dynamic updates
    OnChange(callback func(event RegistryEvent)) func()      // Subscribe to changes
    Refresh() error                                          // Trigger refresh from sources
}

// RegistryEvent for change notifications
type RegistryEvent struct {
    Type    RegistryEventType  // added, updated, removed
    ToolID  string
    Backend *toolmodel.ToolBackend
}

type RegistryEventType string

const (
    RegistryEventAdded   RegistryEventType = "added"
    RegistryEventUpdated RegistryEventType = "updated"
    RegistryEventRemoved RegistryEventType = "removed"
)
```

#### BackendSelector Enhancement

```go
// Enhanced backend selection with context
type BackendSelectorFunc func(
    tool toolmodel.Tool,
    backends []toolmodel.ToolBackend,
    hints *SelectionHints,
) toolmodel.ToolBackend

// SelectionHints provides context for backend selection
type SelectionHints struct {
    PreferredKind    toolmodel.BackendKind // Caller preference
    PreferredBackend string                 // Specific backend name
    LatencyBudget    time.Duration          // Latency requirement
    CostSensitive    bool                   // Prefer cheaper backends
}
```

---

## Layer 3: tooldocs (Documentation)

### Current State

Progressive disclosure documentation layer. Three detail levels: summary, schema, full.

**Exported Types:**
- `Store` - Interface for documentation storage
- `InMemoryStore` - Default implementation
- `DetailLevel` - Enum: summary, schema, full
- `ToolDoc` - Documentation at various levels
- `ToolExample` - Usage example
- `SchemaInfo` - Derived schema information
- `DocEntry` - Input for registration

**Key Methods:**
- `DescribeTool(id, level)` - Get documentation at level
- `ListExamples(id, max)` - Get examples
- `RegisterDoc()`, `RegisterExamples()`

### Changes Needed for Pluggable Architecture

#### Priority: LOW (minimal changes needed)

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add `BulkRegister()` method** | Backend aggregation may load many docs | Minor - additive |
| **Add source tracking** | Track which backend provided docs | Minor - structural |

#### Proposed Additions

```go
// Store interface additions
type Store interface {
    // Existing methods...
    DescribeTool(id string, level DetailLevel) (ToolDoc, error)
    ListExamples(id string, maxExamples int) ([]ToolExample, error)

    // NEW: Bulk operations for backend aggregation
    BulkRegisterDocs(entries map[string]DocEntry) error

    // NEW: Source tracking
    GetDocSource(id string) (string, error)  // Returns backend name
}

// DocEntry - add source field
type DocEntry struct {
    Summary      string
    Notes        string
    Examples     []ToolExample
    ExternalRefs []string
    Source       string  // NEW: Backend that provided this doc
}
```

---

## Layer 4: toolrun (Execution)

### Current State

Tool execution layer supporting MCP, Provider, and Local backends.

**Exported Types:**
- `Runner` - Interface for tool execution
- `DefaultRunner` - Default implementation
- `MCPExecutor` - Interface for MCP backend calls
- `ProviderExecutor` - Interface for provider backend calls
- `LocalRegistry` - Interface for local handler lookup
- `RunResult`, `StepResult`, `ChainStep` - Execution results
- `StreamEvent` - Streaming event type
- `ToolError` - Contextual error wrapper

**Key Methods:**
- `Run()` - Execute single tool
- `RunStream()` - Execute with streaming
- `RunChain()` - Execute tool sequence

### Changes Needed for Pluggable Architecture

#### Priority: HIGH (core execution layer)

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add `HTTPExecutor` interface** | Support HTTP backend execution | Medium - additive |
| **Add `GRPCExecutor` interface** | Support gRPC backend execution | Medium - additive |
| **Add `ExecutorRegistry`** | Dynamic executor registration | Medium - structural |
| **Add execution hooks** | Middleware integration point | Medium - additive |
| **Add `BackendRouter`** | Route execution to correct executor | Medium - structural |

#### Proposed Interface Additions

```go
// HTTPExecutor for HTTP backend execution
type HTTPExecutor interface {
    CallTool(ctx context.Context, baseURL string, tool string, args map[string]any) (any, error)
    CallToolStream(ctx context.Context, baseURL string, tool string, args map[string]any) (<-chan StreamEvent, error)
}

// GRPCExecutor for gRPC backend execution
type GRPCExecutor interface {
    CallTool(ctx context.Context, address string, tool string, args map[string]any) (any, error)
    CallToolStream(ctx context.Context, address string, tool string, args map[string]any) (<-chan StreamEvent, error)
}

// ExecutorRegistry manages execution backends
type ExecutorRegistry interface {
    Register(kind toolmodel.BackendKind, executor any) error
    Get(kind toolmodel.BackendKind) (any, bool)
    List() []toolmodel.BackendKind
}

// ExecutionHook for middleware integration
type ExecutionHook interface {
    BeforeExecution(ctx context.Context, toolID string, args map[string]any) (context.Context, error)
    AfterExecution(ctx context.Context, toolID string, result RunResult, err error) error
}

// Config additions
type Config struct {
    // Existing fields...
    Index            toolindex.Index
    ToolResolver     func(id string) (*toolmodel.Tool, error)
    BackendsResolver func(id string) ([]toolmodel.ToolBackend, error)
    BackendSelector  toolindex.BackendSelector
    Validator        toolmodel.SchemaValidator
    ValidateInput    bool
    ValidateOutput   bool
    MCP              MCPExecutor
    Provider         ProviderExecutor
    Local            LocalRegistry

    // NEW: Additional executors
    HTTP             HTTPExecutor        // NEW
    GRPC             GRPCExecutor        // NEW
    ExecutorRegistry ExecutorRegistry    // NEW: Dynamic registration
    Hooks            []ExecutionHook     // NEW: Middleware hooks
}
```

#### Enhanced Runner Interface

```go
// Runner interface with execution options
type Runner interface {
    Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error)
    RunStream(ctx context.Context, toolID string, args map[string]any) (<-chan StreamEvent, error)
    RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error)

    // NEW: Execution with options
    RunWithOptions(ctx context.Context, toolID string, args map[string]any, opts RunOptions) (RunResult, error)
}

// RunOptions for execution customization
type RunOptions struct {
    PreferredBackend string                 // Override backend selection
    Timeout          time.Duration          // Per-call timeout
    Metadata         map[string]any         // Pass-through metadata
    SkipValidation   bool                   // Skip input/output validation
}
```

---

## Layer 5: toolcode (Code Execution)

### Current State

Code execution orchestration layer with pluggable engines.

**Exported Types:**
- `Executor` - Interface for code execution
- `DefaultExecutor` - Default implementation
- `Engine` - Interface for language-specific execution
- `Tools` - Metatool environment exposed to code
- `Config`, `ExecuteParams`, `ExecuteResult`
- `ToolCallRecord`, `CodeError`

### Changes Needed for Pluggable Architecture

#### Priority: LOW (mostly stable)

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add engine registry** | Support multiple language engines | Minor - structural |
| **Add execution context** | Pass middleware context through | Minor - additive |

#### Proposed Additions

```go
// EngineRegistry for multiple language support
type EngineRegistry interface {
    Register(language string, engine Engine) error
    Get(language string) (Engine, bool)
    List() []string
}

// Config additions
type Config struct {
    // Existing fields...
    Index           toolindex.Index
    Docs            tooldocs.Store
    Run             toolrun.Runner
    Engine          Engine
    DefaultTimeout  time.Duration
    DefaultLanguage string
    MaxToolCalls    int
    MaxChainSteps   int
    Logger          Logger

    // NEW: Multiple engine support
    EngineRegistry  EngineRegistry  // NEW
}

// ExecuteParams additions
type ExecuteParams struct {
    Language     string
    Code         string
    Timeout      time.Duration
    MaxToolCalls int

    // NEW: Execution context
    Metadata     map[string]any  // NEW: Pass-through metadata
}
```

---

## Layer 6: toolsearch (BM25 Search)

### Current State

BM25 search implementation using Bleve. Implements `toolindex.Searcher`.

**Exported Types:**
- `BM25Config` - Configuration (boosts, limits)
- `BM25Searcher` - Searcher implementation

### Changes Needed for Pluggable Architecture

#### Priority: NONE (stable, well-designed)

The library already implements the pluggable pattern correctly via `toolindex.Searcher` interface. No changes needed for the pluggable architecture.

---

## Layer 7: toolruntime (Sandbox)

### Current State

Runtime and trust boundary layer with multiple isolation backends.

**Exported Types:**
- `Runtime` - Interface for code execution runtime
- `DefaultRuntime` - Default implementation
- `Backend` - Interface for isolation backends
- `ToolGateway` - Tool access interface for sandboxed code
- `SecurityProfile` - dev, standard, hardened
- `BackendKind` - Isolation mechanism (Docker, K8s, gVisor, etc.)
- `ExecuteRequest`, `ExecuteResult`, `Limits`

### Changes Needed for Pluggable Architecture

#### Priority: LOW (mostly independent)

| Change | Rationale | Impact |
|--------|-----------|--------|
| **Add security profile config** | Configure profiles via YAML | Minor - additive |
| **Add backend discovery** | Auto-discover available backends | Minor - additive |

---

## Cross-Cutting Concerns

### Error Handling Improvements

All libraries should adopt a consistent error taxonomy:

```go
// Proposed: pkg/errors/errors.go (new shared package)

// ToolError is the base error type for all tool operations
type ToolError struct {
    Code       ErrorCode
    Message    string
    ToolID     string
    Backend    string
    Op         string
    Cause      error
    Retryable  bool
    Metadata   map[string]any
}

type ErrorCode string

const (
    ErrCodeNotFound       ErrorCode = "not_found"
    ErrCodeValidation     ErrorCode = "validation"
    ErrCodeExecution      ErrorCode = "execution"
    ErrCodeTimeout        ErrorCode = "timeout"
    ErrCodeBackendFailure ErrorCode = "backend_failure"
    ErrCodeRateLimit      ErrorCode = "rate_limit"
    ErrCodeAuth           ErrorCode = "auth"
)

func (e *ToolError) Error() string
func (e *ToolError) Unwrap() error
func (e *ToolError) Is(target error) bool
func (e *ToolError) IsRetryable() bool
```

### Context Propagation

All libraries should propagate context consistently:

```go
// Proposed: Context keys for cross-cutting concerns
type contextKey string

const (
    ContextKeyRequestID  contextKey = "request_id"
    ContextKeyUserID     contextKey = "user_id"
    ContextKeyTraceID    contextKey = "trace_id"
    ContextKeyBackend    contextKey = "backend"
    ContextKeyMetadata   contextKey = "metadata"
)

// Helper functions
func WithRequestID(ctx context.Context, id string) context.Context
func RequestIDFromContext(ctx context.Context) string
// ... etc
```

---

## Implementation Priority

### Phase 1: Foundation (with metatools-mcp Phase 1)

1. **toolmodel**: Add HTTP/GRPC backend kinds (1 day)
2. **toolrun**: Add HTTPExecutor, GRPCExecutor interfaces (2 days)

### Phase 2: Registry Enhancements (with metatools-mcp Phase 4)

1. **toolindex**: Add ListTools, OnChange, source tracking (2 days)
2. **toolrun**: Add ExecutorRegistry, BackendRouter (2 days)

### Phase 3: Middleware Integration (with metatools-mcp Phase 5)

1. **toolrun**: Add ExecutionHook interface (1 day)
2. **All libraries**: Consistent error taxonomy (2 days)
3. **All libraries**: Context propagation (1 day)

---

## Version Compatibility Matrix

Current ecosystem versions:

| Library | Current | After Phase 1 | After Phase 2 | After Phase 3 |
|---------|---------|---------------|---------------|---------------|
| toolmodel | v0.1.2 | v0.2.0 | v0.2.0 | v0.2.0 |
| toolindex | v0.1.8 | v0.1.8 | v0.2.0 | v0.2.0 |
| tooldocs | v0.1.10 | v0.1.10 | v0.1.11 | v0.1.11 |
| toolrun | v0.1.9 | v0.2.0 | v0.2.0 | v0.3.0 |
| toolcode | v0.1.10 | v0.1.10 | v0.1.10 | v0.1.11 |
| toolsearch | v0.1.9 | v0.1.9 | v0.1.9 | v0.1.9 |
| toolruntime | v0.1.10 | v0.1.10 | v0.1.10 | v0.1.11 |

---

## Summary: Required Library Changes

### Breaking Changes: NONE

All proposed changes are additive and backward-compatible.

### High Priority Changes

| Library | Change | Files Affected |
|---------|--------|----------------|
| **toolmodel** | Add BackendKindHTTP, BackendKindGRPC | types.go |
| **toolmodel** | Add HTTPBackend, GRPCBackend structs | types.go |
| **toolrun** | Add HTTPExecutor interface | executor.go (new) |
| **toolrun** | Add GRPCExecutor interface | executor.go (new) |
| **toolrun** | Add ExecutorRegistry | registry.go (new) |

### Medium Priority Changes

| Library | Change | Files Affected |
|---------|--------|----------------|
| **toolindex** | Add ListTools method | index.go |
| **toolindex** | Add OnChange callback | index.go |
| **toolrun** | Add ExecutionHook interface | hooks.go (new) |
| **toolrun** | Add RunOptions, RunWithOptions | runner.go |

### Low Priority Changes

| Library | Change | Files Affected |
|---------|--------|----------------|
| **tooldocs** | Add BulkRegisterDocs | store.go |
| **toolcode** | Add EngineRegistry | engine.go (new) |
| **All** | Consistent error taxonomy | errors.go (new per lib) |

---

## Key Discovery: Architecture Already Pluggable

Comprehensive analysis of all 8 component libraries (40+ source files, 12,000+ lines) revealed:

> **The metatools ecosystem is NOT a monolith to be refactored.** It is a mature, layered, pluggable architecture with **13 extension points** already implemented as Go interfaces.

### 13 Extension Points Catalogued

| # | Interface | Library | Status |
|---|-----------|---------|--------|
| 1 | `SchemaValidator` | toolmodel | ✅ Interface-based |
| 2 | `Searcher` | toolindex | ✅ Interface-based |
| 3 | `BackendSelector` | toolindex | ✅ Function-based |
| 4 | `Store` | tooldocs | ✅ Interface-based |
| 5 | `ToolResolver` | tooldocs | ✅ Function-based |
| 6 | `Runner` | toolrun | ✅ Interface-based |
| 7 | `MCPExecutor` | toolrun | ✅ Interface-based |
| 8 | `ProviderExecutor` | toolrun | ✅ Interface-based |
| 9 | `LocalRegistry` | toolrun | ✅ Interface-based |
| 10 | `Backend` | toolruntime | ✅ Interface-based (10 implementations!) |
| 11 | `ToolGateway` | toolruntime | ✅ Interface-based |
| 12 | `Logger` | toolcode | ✅ Interface-based |
| 13 | `Engine` | toolcode | ✅ Interface-based |

### Implications for Implementation

The "pluggable architecture" work is primarily:
1. **Exposure** - Make internal extension points accessible via configuration
2. **Configuration** - Add CLI + config layer (Cobra + Koanf)
3. **Documentation** - Catalog the 13 extension points with examples

This reduces the implementation timeline from 9 weeks to **6-7 weeks** (25% reduction).

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-27 | Initial analysis of all 7 component libraries |
| 2026-01-28 | Added Key Discovery section documenting 13 existing extension points |
| 2026-01-28 | Updated analysis to reflect that architecture is already pluggable |
