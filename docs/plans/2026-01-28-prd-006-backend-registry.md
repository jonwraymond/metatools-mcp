# PRD-006: Backend Registry

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a Backend Registry enabling multi-source tool aggregation, where tools can be discovered and executed from local handlers, MCP servers, HTTP APIs, and other external sources through a unified interface.

**Architecture:** Define a Backend interface that abstracts tool sources. Create a BackendRegistry that manages backend lifecycle, discovery, and routing. Implement LocalBackend for in-process handlers. Enable configuration-driven backend loading with automatic tool aggregation.

**Tech Stack:** Go interfaces, toolmodel.Tool integration, config integration from PRD-003

**Priority:** P1 - Stream A, Phase 4 (completes core exposure)

**Scope:** Backend interface + registry + local backend + tool aggregation

**Dependencies:** PRD-002 (CLI), PRD-003 (Config), PRD-005 (Provider Registry)

---

## Context

The current architecture supports only local tool handlers. The pluggable architecture requires:
1. Multiple tool sources (local, MCP, HTTP, custom)
2. Unified tool discovery across backends
3. Transparent execution routing
4. Configuration-driven backend management

**Current State:**
```go
// All tools are local handlers
handlers := map[string]Handler{
    "search_tools": searchHandler,
    "run_tool": runHandler,
}
```

**Target State:**
```go
// Backend interface for any tool source
type Backend interface {
    Kind() string
    Name() string
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
}

// Registry aggregates all backends
registry.Register("local", &LocalBackend{...})
registry.Register("github", &MCPBackend{...})  // Future PRD
```

---

## Tasks

### Task 1: Define Backend Interface

**Files:**
- Create: `internal/backend/backend.go`
- Test: `internal/backend/backend_test.go`

**Step 1: Write failing test for Backend interface**

```go
// internal/backend/backend_test.go
package backend

import (
    "context"
    "testing"

    "github.com/jraymond/toolmodel"
)

// mockBackend implements Backend for testing
type mockBackend struct {
    kind     string
    name     string
    enabled  bool
    tools    []toolmodel.Tool
    execFn   func(ctx context.Context, tool string, args map[string]any) (any, error)
}

func (m *mockBackend) Kind() string    { return m.kind }
func (m *mockBackend) Name() string    { return m.name }
func (m *mockBackend) Enabled() bool   { return m.enabled }

func (m *mockBackend) ListTools(ctx context.Context) ([]toolmodel.Tool, error) {
    return m.tools, nil
}

func (m *mockBackend) Execute(ctx context.Context, tool string, args map[string]any) (any, error) {
    if m.execFn != nil {
        return m.execFn(ctx, tool, args)
    }
    return nil, nil
}

func (m *mockBackend) Start(ctx context.Context) error { return nil }
func (m *mockBackend) Stop() error                     { return nil }

func TestBackend_Interface(t *testing.T) {
    // Verify interface is implemented correctly
    var _ Backend = (*mockBackend)(nil)
}

func TestBackend_Methods(t *testing.T) {
    backend := &mockBackend{
        kind:    "local",
        name:    "test-backend",
        enabled: true,
        tools: []toolmodel.Tool{
            {Name: "test_tool", Description: "A test tool"},
        },
        execFn: func(ctx context.Context, tool string, args map[string]any) (any, error) {
            return "executed", nil
        },
    }

    if backend.Kind() != "local" {
        t.Errorf("Kind() = %q, want %q", backend.Kind(), "local")
    }

    if backend.Name() != "test-backend" {
        t.Errorf("Name() = %q, want %q", backend.Name(), "test-backend")
    }

    if !backend.Enabled() {
        t.Error("Enabled() = false, want true")
    }

    tools, err := backend.ListTools(context.Background())
    if err != nil {
        t.Fatalf("ListTools() error = %v", err)
    }
    if len(tools) != 1 {
        t.Errorf("ListTools() returned %d tools, want 1", len(tools))
    }

    result, err := backend.Execute(context.Background(), "test_tool", nil)
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }
    if result != "executed" {
        t.Errorf("Execute() = %v, want %v", result, "executed")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -v`
Expected: FAIL - Backend type doesn't exist

**Step 3: Implement Backend interface**

```go
// internal/backend/backend.go
package backend

import (
    "context"
    "errors"

    "github.com/jraymond/toolmodel"
)

// Common errors for backend operations
var (
    ErrBackendNotFound    = errors.New("backend not found")
    ErrBackendDisabled    = errors.New("backend disabled")
    ErrToolNotFound       = errors.New("tool not found in backend")
    ErrBackendUnavailable = errors.New("backend unavailable")
)

// Backend defines a source of tools.
// Backends can be local handlers, MCP servers, HTTP APIs, or custom implementations.
type Backend interface {
    // Kind returns the backend type (e.g., "local", "mcp", "http")
    Kind() string

    // Name returns the unique instance name for this backend
    Name() string

    // Enabled returns whether this backend is currently enabled
    Enabled() bool

    // ListTools returns all tools available from this backend
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)

    // Execute invokes a tool on this backend
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)

    // Start initializes the backend (connect to remote, start subprocess, etc.)
    Start(ctx context.Context) error

    // Stop gracefully shuts down the backend
    Stop() error
}

// ConfigurableBackend is a backend that can be configured from raw bytes (YAML/JSON)
type ConfigurableBackend interface {
    Backend

    // Configure applies configuration from raw bytes
    Configure(raw []byte) error
}

// StreamingBackend supports streaming responses
type StreamingBackend interface {
    Backend

    // ExecuteStream returns a channel of response chunks
    ExecuteStream(ctx context.Context, tool string, args map[string]any) (<-chan any, error)
}

// BackendFactory creates backend instances from configuration
type BackendFactory func(name string) (Backend, error)

// BackendInfo contains metadata about a backend
type BackendInfo struct {
    Kind        string
    Name        string
    Enabled     bool
    ToolCount   int
    Streaming   bool
    Configurable bool
}

// GetInfo returns metadata about a backend
func GetInfo(b Backend) BackendInfo {
    info := BackendInfo{
        Kind:    b.Kind(),
        Name:    b.Name(),
        Enabled: b.Enabled(),
    }

    // Check capabilities
    _, info.Streaming = b.(StreamingBackend)
    _, info.Configurable = b.(ConfigurableBackend)

    // Get tool count (best effort)
    if tools, err := b.ListTools(context.Background()); err == nil {
        info.ToolCount = len(tools)
    }

    return info
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backend/backend.go internal/backend/backend_test.go
git commit -m "$(cat <<'EOF'
feat(backend): define Backend interface

- Backend with Kind, Name, Enabled, ListTools, Execute methods
- Start/Stop for lifecycle management
- ConfigurableBackend for YAML/JSON config
- StreamingBackend for streaming responses
- BackendFactory and BackendInfo types

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Implement Backend Registry

**Files:**
- Create: `internal/backend/registry.go`
- Test: `internal/backend/registry_test.go`

**Step 1: Write failing test for registry**

```go
// internal/backend/registry_test.go
package backend

import (
    "context"
    "testing"
)

func TestRegistry_Register(t *testing.T) {
    registry := NewRegistry()

    backend := &mockBackend{
        kind:    "local",
        name:    "test",
        enabled: true,
    }

    err := registry.Register(backend)
    if err != nil {
        t.Fatalf("Register() error = %v", err)
    }

    // Duplicate registration should fail
    err = registry.Register(backend)
    if err == nil {
        t.Error("Register() should fail on duplicate")
    }
}

func TestRegistry_Get(t *testing.T) {
    registry := NewRegistry()

    backend := &mockBackend{
        kind:    "local",
        name:    "test",
        enabled: true,
    }
    registry.Register(backend)

    got, ok := registry.Get("test")
    if !ok {
        t.Fatal("Get() returned false")
    }
    if got.Name() != "test" {
        t.Errorf("Get().Name() = %q, want %q", got.Name(), "test")
    }

    _, ok = registry.Get("nonexistent")
    if ok {
        t.Error("Get() should return false for nonexistent backend")
    }
}

func TestRegistry_List(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockBackend{kind: "local", name: "a", enabled: true})
    registry.Register(&mockBackend{kind: "mcp", name: "b", enabled: true})
    registry.Register(&mockBackend{kind: "http", name: "c", enabled: false})

    // List all
    all := registry.List()
    if len(all) != 3 {
        t.Errorf("List() returned %d backends, want 3", len(all))
    }

    // List enabled only
    enabled := registry.ListEnabled()
    if len(enabled) != 2 {
        t.Errorf("ListEnabled() returned %d backends, want 2", len(enabled))
    }
}

func TestRegistry_ListByKind(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockBackend{kind: "local", name: "local1", enabled: true})
    registry.Register(&mockBackend{kind: "local", name: "local2", enabled: true})
    registry.Register(&mockBackend{kind: "mcp", name: "mcp1", enabled: true})

    locals := registry.ListByKind("local")
    if len(locals) != 2 {
        t.Errorf("ListByKind(local) returned %d backends, want 2", len(locals))
    }

    mcps := registry.ListByKind("mcp")
    if len(mcps) != 1 {
        t.Errorf("ListByKind(mcp) returned %d backends, want 1", len(mcps))
    }
}

func TestRegistry_Unregister(t *testing.T) {
    registry := NewRegistry()

    backend := &mockBackend{kind: "local", name: "test", enabled: true}
    registry.Register(backend)

    registry.Unregister("test")

    _, ok := registry.Get("test")
    if ok {
        t.Error("Get() should return false after Unregister()")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestRegistry -v`
Expected: FAIL - Registry doesn't exist

**Step 3: Implement Registry**

```go
// internal/backend/registry.go
package backend

import (
    "fmt"
    "sort"
    "sync"
)

// ErrBackendExists is returned when registering a duplicate backend
var ErrBackendExists = errors.New("backend already registered")

// Registry manages backend instances
type Registry struct {
    backends  map[string]Backend
    factories map[string]BackendFactory
    mu        sync.RWMutex
}

// NewRegistry creates a new backend registry
func NewRegistry() *Registry {
    return &Registry{
        backends:  make(map[string]Backend),
        factories: make(map[string]BackendFactory),
    }
}

// RegisterFactory registers a factory for creating backends of a given kind
func (r *Registry) RegisterFactory(kind string, factory BackendFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.factories[kind] = factory
}

// Register adds a backend to the registry
func (r *Registry) Register(b Backend) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := b.Name()
    if _, exists := r.backends[name]; exists {
        return fmt.Errorf("%w: %s", ErrBackendExists, name)
    }

    r.backends[name] = b
    return nil
}

// MustRegister adds a backend or panics
func (r *Registry) MustRegister(b Backend) {
    if err := r.Register(b); err != nil {
        panic(err)
    }
}

// Unregister removes a backend from the registry
func (r *Registry) Unregister(name string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    if b, exists := r.backends[name]; exists {
        b.Stop() // Best effort cleanup
        delete(r.backends, name)
    }
}

// Get retrieves a backend by name
func (r *Registry) Get(name string) (Backend, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    b, ok := r.backends[name]
    return b, ok
}

// List returns all registered backends
func (r *Registry) List() []Backend {
    r.mu.RLock()
    defer r.mu.RUnlock()

    backends := make([]Backend, 0, len(r.backends))
    for _, b := range r.backends {
        backends = append(backends, b)
    }

    // Sort by name for consistency
    sort.Slice(backends, func(i, j int) bool {
        return backends[i].Name() < backends[j].Name()
    })

    return backends
}

// ListEnabled returns only enabled backends
func (r *Registry) ListEnabled() []Backend {
    r.mu.RLock()
    defer r.mu.RUnlock()

    backends := make([]Backend, 0, len(r.backends))
    for _, b := range r.backends {
        if b.Enabled() {
            backends = append(backends, b)
        }
    }

    sort.Slice(backends, func(i, j int) bool {
        return backends[i].Name() < backends[j].Name()
    })

    return backends
}

// ListByKind returns backends of a specific kind
func (r *Registry) ListByKind(kind string) []Backend {
    r.mu.RLock()
    defer r.mu.RUnlock()

    backends := make([]Backend, 0)
    for _, b := range r.backends {
        if b.Kind() == kind {
            backends = append(backends, b)
        }
    }

    sort.Slice(backends, func(i, j int) bool {
        return backends[i].Name() < backends[j].Name()
    })

    return backends
}

// Count returns the number of registered backends
func (r *Registry) Count() int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.backends)
}

// StartAll starts all enabled backends
func (r *Registry) StartAll(ctx context.Context) error {
    for _, b := range r.ListEnabled() {
        if err := b.Start(ctx); err != nil {
            return fmt.Errorf("start backend %s: %w", b.Name(), err)
        }
    }
    return nil
}

// StopAll stops all backends
func (r *Registry) StopAll() error {
    var firstErr error
    for _, b := range r.List() {
        if err := b.Stop(); err != nil && firstErr == nil {
            firstErr = fmt.Errorf("stop backend %s: %w", b.Name(), err)
        }
    }
    return firstErr
}

// Clear removes all backends
func (r *Registry) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, b := range r.backends {
        b.Stop()
    }
    r.backends = make(map[string]Backend)
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestRegistry -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backend/registry.go internal/backend/registry_test.go
git commit -m "$(cat <<'EOF'
feat(backend): implement Backend Registry

- Register/Unregister backends by name
- Get/List/ListEnabled/ListByKind for lookup
- RegisterFactory for kind-based creation
- StartAll/StopAll for lifecycle management
- Thread-safe with RWMutex

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement Local Backend

**Files:**
- Create: `internal/backend/local/local.go`
- Test: `internal/backend/local/local_test.go`

**Step 1: Write failing test for LocalBackend**

```go
// internal/backend/local/local_test.go
package local

import (
    "context"
    "testing"

    "github.com/your-org/metatools-mcp/internal/backend"
)

func TestLocalBackend_Interface(t *testing.T) {
    // Verify it implements Backend
    var _ backend.Backend = (*Backend)(nil)
}

func TestLocalBackend_Kind(t *testing.T) {
    b := New("test")
    if b.Kind() != "local" {
        t.Errorf("Kind() = %q, want %q", b.Kind(), "local")
    }
}

func TestLocalBackend_Name(t *testing.T) {
    b := New("my-local")
    if b.Name() != "my-local" {
        t.Errorf("Name() = %q, want %q", b.Name(), "my-local")
    }
}

func TestLocalBackend_RegisterHandler(t *testing.T) {
    b := New("test")

    handler := func(ctx context.Context, args map[string]any) (any, error) {
        return "handled", nil
    }

    b.RegisterHandler("my_tool", ToolDef{
        Name:        "my_tool",
        Description: "A test tool",
        Handler:     handler,
    })

    tools, err := b.ListTools(context.Background())
    if err != nil {
        t.Fatalf("ListTools() error = %v", err)
    }

    if len(tools) != 1 {
        t.Fatalf("ListTools() returned %d tools, want 1", len(tools))
    }

    if tools[0].Name != "my_tool" {
        t.Errorf("Tool.Name = %q, want %q", tools[0].Name, "my_tool")
    }
}

func TestLocalBackend_Execute(t *testing.T) {
    b := New("test")

    b.RegisterHandler("echo", ToolDef{
        Name:        "echo",
        Description: "Echo input",
        Handler: func(ctx context.Context, args map[string]any) (any, error) {
            return args["message"], nil
        },
    })

    result, err := b.Execute(context.Background(), "echo", map[string]any{
        "message": "hello",
    })

    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    if result != "hello" {
        t.Errorf("Execute() = %v, want %v", result, "hello")
    }
}

func TestLocalBackend_ExecuteNotFound(t *testing.T) {
    b := New("test")

    _, err := b.Execute(context.Background(), "nonexistent", nil)
    if err == nil {
        t.Error("Execute() should fail for nonexistent tool")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/local/... -v`
Expected: FAIL - Backend doesn't exist

**Step 3: Implement LocalBackend**

```go
// internal/backend/local/local.go
package local

import (
    "context"
    "fmt"
    "sync"

    "github.com/jraymond/toolmodel"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/backend"
)

// HandlerFunc is the function signature for tool handlers
type HandlerFunc func(ctx context.Context, args map[string]any) (any, error)

// ToolDef defines a local tool with its handler
type ToolDef struct {
    Name        string
    Description string
    InputSchema map[string]any
    Handler     HandlerFunc
}

// Backend implements the Backend interface for local tool handlers
type Backend struct {
    name     string
    enabled  bool
    handlers map[string]ToolDef
    mu       sync.RWMutex
}

// New creates a new local backend
func New(name string) *Backend {
    return &Backend{
        name:     name,
        enabled:  true,
        handlers: make(map[string]ToolDef),
    }
}

// Kind returns "local"
func (b *Backend) Kind() string {
    return "local"
}

// Name returns the backend instance name
func (b *Backend) Name() string {
    return b.name
}

// Enabled returns whether the backend is enabled
func (b *Backend) Enabled() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.enabled
}

// SetEnabled enables or disables the backend
func (b *Backend) SetEnabled(enabled bool) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.enabled = enabled
}

// RegisterHandler adds a tool handler
func (b *Backend) RegisterHandler(name string, def ToolDef) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers[name] = def
}

// UnregisterHandler removes a tool handler
func (b *Backend) UnregisterHandler(name string) {
    b.mu.Lock()
    defer b.mu.Unlock()
    delete(b.handlers, name)
}

// ListTools returns all registered tools
func (b *Backend) ListTools(ctx context.Context) ([]toolmodel.Tool, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    tools := make([]toolmodel.Tool, 0, len(b.handlers))
    for _, def := range b.handlers {
        tool := toolmodel.Tool{
            Tool: mcp.Tool{
                Name:        def.Name,
                Description: def.Description,
                InputSchema: mcp.ToolInputSchema{
                    Type:       "object",
                    Properties: def.InputSchema,
                },
            },
            Namespace: b.name,
        }
        tools = append(tools, tool)
    }

    return tools, nil
}

// Execute invokes a tool handler
func (b *Backend) Execute(ctx context.Context, tool string, args map[string]any) (any, error) {
    b.mu.RLock()
    def, ok := b.handlers[tool]
    b.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("%w: %s", backend.ErrToolNotFound, tool)
    }

    if def.Handler == nil {
        return nil, fmt.Errorf("tool %s has no handler", tool)
    }

    return def.Handler(ctx, args)
}

// Start is a no-op for local backends
func (b *Backend) Start(ctx context.Context) error {
    return nil
}

// Stop is a no-op for local backends
func (b *Backend) Stop() error {
    return nil
}

// Ensure Backend implements backend.Backend
var _ backend.Backend = (*Backend)(nil)
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/local/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backend/local/local.go internal/backend/local/local_test.go
git commit -m "$(cat <<'EOF'
feat(backend): implement LocalBackend for in-process handlers

- RegisterHandler/UnregisterHandler for tool management
- ListTools returns toolmodel.Tool with namespace
- Execute invokes handler with context and args
- Thread-safe handler access

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Implement Tool Aggregator

**Files:**
- Create: `internal/backend/aggregator.go`
- Test: `internal/backend/aggregator_test.go`

**Step 1: Write failing test for Aggregator**

```go
// internal/backend/aggregator_test.go
package backend

import (
    "context"
    "testing"

    "github.com/jraymond/toolmodel"
)

func TestAggregator_ListAllTools(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockBackend{
        kind:    "local",
        name:    "local1",
        enabled: true,
        tools: []toolmodel.Tool{
            {Name: "tool_a", Namespace: "local1"},
            {Name: "tool_b", Namespace: "local1"},
        },
    })

    registry.Register(&mockBackend{
        kind:    "mcp",
        name:    "github",
        enabled: true,
        tools: []toolmodel.Tool{
            {Name: "create_issue", Namespace: "github"},
        },
    })

    registry.Register(&mockBackend{
        kind:    "local",
        name:    "disabled",
        enabled: false,
        tools: []toolmodel.Tool{
            {Name: "should_not_appear"},
        },
    })

    agg := NewAggregator(registry)

    tools, err := agg.ListAllTools(context.Background())
    if err != nil {
        t.Fatalf("ListAllTools() error = %v", err)
    }

    if len(tools) != 3 {
        t.Errorf("ListAllTools() returned %d tools, want 3", len(tools))
    }
}

func TestAggregator_Execute(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockBackend{
        kind:    "local",
        name:    "local",
        enabled: true,
        execFn: func(ctx context.Context, tool string, args map[string]any) (any, error) {
            if tool == "echo" {
                return args["msg"], nil
            }
            return nil, ErrToolNotFound
        },
    })

    agg := NewAggregator(registry)

    // Execute with backend prefix
    result, err := agg.Execute(context.Background(), "local/echo", map[string]any{
        "msg": "hello",
    })

    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    if result != "hello" {
        t.Errorf("Execute() = %v, want %v", result, "hello")
    }
}

func TestAggregator_ExecuteNotFound(t *testing.T) {
    registry := NewRegistry()
    agg := NewAggregator(registry)

    _, err := agg.Execute(context.Background(), "nonexistent/tool", nil)
    if err == nil {
        t.Error("Execute() should fail for nonexistent backend")
    }
}

func TestAggregator_ParseToolID(t *testing.T) {
    tests := []struct {
        id              string
        wantBackend     string
        wantTool        string
        wantErr         bool
    }{
        {"local/echo", "local", "echo", false},
        {"github/create_issue", "github", "create_issue", false},
        {"my-backend/my_tool", "my-backend", "my_tool", false},
        {"no_slash", "", "", true},
        {"", "", "", true},
    }

    for _, tt := range tests {
        backend, tool, err := ParseToolID(tt.id)
        if (err != nil) != tt.wantErr {
            t.Errorf("ParseToolID(%q) error = %v, wantErr = %v", tt.id, err, tt.wantErr)
            continue
        }
        if backend != tt.wantBackend {
            t.Errorf("ParseToolID(%q) backend = %q, want %q", tt.id, backend, tt.wantBackend)
        }
        if tool != tt.wantTool {
            t.Errorf("ParseToolID(%q) tool = %q, want %q", tt.id, tool, tt.wantTool)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestAggregator -v`
Expected: FAIL - Aggregator doesn't exist

**Step 3: Implement Aggregator**

```go
// internal/backend/aggregator.go
package backend

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "sync"

    "github.com/jraymond/toolmodel"
)

// ErrInvalidToolID is returned for malformed tool IDs
var ErrInvalidToolID = errors.New("invalid tool ID format (expected backend/tool)")

// Aggregator combines tools from multiple backends
type Aggregator struct {
    registry *Registry
}

// NewAggregator creates a new tool aggregator
func NewAggregator(registry *Registry) *Aggregator {
    return &Aggregator{registry: registry}
}

// ListAllTools returns tools from all enabled backends
func (a *Aggregator) ListAllTools(ctx context.Context) ([]toolmodel.Tool, error) {
    backends := a.registry.ListEnabled()

    var allTools []toolmodel.Tool
    var mu sync.Mutex
    var wg sync.WaitGroup
    var firstErr error

    for _, b := range backends {
        wg.Add(1)
        go func(backend Backend) {
            defer wg.Done()

            tools, err := backend.ListTools(ctx)
            if err != nil {
                mu.Lock()
                if firstErr == nil {
                    firstErr = fmt.Errorf("backend %s: %w", backend.Name(), err)
                }
                mu.Unlock()
                return
            }

            // Add backend prefix to tool IDs
            for i := range tools {
                if tools[i].Namespace == "" {
                    tools[i].Namespace = backend.Name()
                }
            }

            mu.Lock()
            allTools = append(allTools, tools...)
            mu.Unlock()
        }(b)
    }

    wg.Wait()

    if firstErr != nil {
        return nil, firstErr
    }

    return allTools, nil
}

// Execute invokes a tool on the appropriate backend
func (a *Aggregator) Execute(ctx context.Context, toolID string, args map[string]any) (any, error) {
    backendName, toolName, err := ParseToolID(toolID)
    if err != nil {
        return nil, err
    }

    backend, ok := a.registry.Get(backendName)
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrBackendNotFound, backendName)
    }

    if !backend.Enabled() {
        return nil, fmt.Errorf("%w: %s", ErrBackendDisabled, backendName)
    }

    return backend.Execute(ctx, toolName, args)
}

// ExecuteResult contains execution result with metadata
type ExecuteResult struct {
    Result    any
    Backend   string
    Tool      string
    Duration  int64 // milliseconds
}

// ExecuteWithInfo invokes a tool and returns detailed result
func (a *Aggregator) ExecuteWithInfo(ctx context.Context, toolID string, args map[string]any) (*ExecuteResult, error) {
    backendName, toolName, err := ParseToolID(toolID)
    if err != nil {
        return nil, err
    }

    backend, ok := a.registry.Get(backendName)
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrBackendNotFound, backendName)
    }

    if !backend.Enabled() {
        return nil, fmt.Errorf("%w: %s", ErrBackendDisabled, backendName)
    }

    result, err := backend.Execute(ctx, toolName, args)
    if err != nil {
        return nil, err
    }

    return &ExecuteResult{
        Result:  result,
        Backend: backendName,
        Tool:    toolName,
    }, nil
}

// ParseToolID splits a qualified tool ID into backend and tool names
func ParseToolID(id string) (backend, tool string, err error) {
    if id == "" {
        return "", "", ErrInvalidToolID
    }

    parts := strings.SplitN(id, "/", 2)
    if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
        return "", "", fmt.Errorf("%w: %s", ErrInvalidToolID, id)
    }

    return parts[0], parts[1], nil
}

// FormatToolID creates a qualified tool ID from backend and tool names
func FormatToolID(backend, tool string) string {
    return fmt.Sprintf("%s/%s", backend, tool)
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestAggregator -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backend/aggregator.go internal/backend/aggregator_test.go
git commit -m "$(cat <<'EOF'
feat(backend): implement Tool Aggregator

- ListAllTools aggregates from all enabled backends (parallel)
- Execute routes to correct backend by tool ID prefix
- ParseToolID/FormatToolID for qualified IDs (backend/tool)
- ExecuteWithInfo returns result with metadata

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Add Backend Configuration Support

**Files:**
- Create: `internal/backend/config.go`
- Test: `internal/backend/config_test.go`

**Step 1: Write failing test for config loading**

```go
// internal/backend/config_test.go
package backend

import (
    "testing"
)

func TestBackendConfig_Unmarshal(t *testing.T) {
    yaml := `
backends:
  local:
    enabled: true
    paths:
      - ~/.config/metatools/tools
  github:
    enabled: true
    kind: mcp
    config:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-github"]
`
    cfg, err := ParseBackendsConfig([]byte(yaml))
    if err != nil {
        t.Fatalf("ParseBackendsConfig() error = %v", err)
    }

    if len(cfg.Backends) != 2 {
        t.Errorf("Expected 2 backends, got %d", len(cfg.Backends))
    }

    local, ok := cfg.Backends["local"]
    if !ok {
        t.Fatal("local backend not found")
    }
    if !local.Enabled {
        t.Error("local backend should be enabled")
    }

    github, ok := cfg.Backends["github"]
    if !ok {
        t.Fatal("github backend not found")
    }
    if github.Kind != "mcp" {
        t.Errorf("github.Kind = %q, want %q", github.Kind, "mcp")
    }
}

func TestBackendConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        cfg     BackendConfig
        wantErr bool
    }{
        {
            name:    "valid local",
            cfg:     BackendConfig{Kind: "local", Enabled: true},
            wantErr: false,
        },
        {
            name:    "valid mcp",
            cfg:     BackendConfig{Kind: "mcp", Enabled: true},
            wantErr: false,
        },
        {
            name:    "empty kind defaults to local",
            cfg:     BackendConfig{Enabled: true},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.cfg.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestBackendConfig -v`
Expected: FAIL - Config types don't exist

**Step 3: Implement config types**

```go
// internal/backend/config.go
package backend

import (
    "fmt"

    "gopkg.in/yaml.v3"
)

// BackendsConfig is the top-level backends configuration
type BackendsConfig struct {
    Backends map[string]BackendConfig `yaml:"backends" koanf:"backends"`
}

// BackendConfig configures a single backend
type BackendConfig struct {
    // Kind is the backend type (local, mcp, http, custom)
    // Defaults to "local" if not specified
    Kind string `yaml:"kind" koanf:"kind"`

    // Enabled determines if the backend is active
    Enabled bool `yaml:"enabled" koanf:"enabled"`

    // Config contains backend-specific configuration
    // This is passed as raw bytes to ConfigurableBackend.Configure()
    Config map[string]any `yaml:"config" koanf:"config"`

    // Common fields extracted for convenience
    Paths   []string `yaml:"paths,omitempty" koanf:"paths"`
    Command string   `yaml:"command,omitempty" koanf:"command"`
    Args    []string `yaml:"args,omitempty" koanf:"args"`
    BaseURL string   `yaml:"base_url,omitempty" koanf:"base_url"`
}

// ParseBackendsConfig parses YAML configuration
func ParseBackendsConfig(data []byte) (*BackendsConfig, error) {
    var cfg BackendsConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse backends config: %w", err)
    }

    // Set defaults
    for name, bc := range cfg.Backends {
        if bc.Kind == "" {
            bc.Kind = "local"
            cfg.Backends[name] = bc
        }
    }

    return &cfg, nil
}

// Validate checks the configuration is valid
func (c *BackendConfig) Validate() error {
    // Kind defaults to local
    if c.Kind == "" {
        c.Kind = "local"
    }

    switch c.Kind {
    case "local", "mcp", "http", "custom":
        // Valid kinds
    default:
        return fmt.Errorf("unknown backend kind: %s", c.Kind)
    }

    return nil
}

// RawConfig returns the config section as YAML bytes for ConfigurableBackend
func (c *BackendConfig) RawConfig() ([]byte, error) {
    if c.Config == nil {
        return nil, nil
    }
    return yaml.Marshal(c.Config)
}

// LoadFromConfig creates backends from configuration
func (r *Registry) LoadFromConfig(cfg *BackendsConfig) error {
    for name, backendCfg := range cfg.Backends {
        if err := backendCfg.Validate(); err != nil {
            return fmt.Errorf("backend %s: %w", name, err)
        }

        if !backendCfg.Enabled {
            continue
        }

        // Get factory for this kind
        factory, ok := r.factories[backendCfg.Kind]
        if !ok {
            // For now, only local is built-in; others require explicit registration
            if backendCfg.Kind != "local" {
                return fmt.Errorf("no factory for backend kind: %s", backendCfg.Kind)
            }
            continue // Local backend handled separately
        }

        // Create backend
        backend, err := factory(name)
        if err != nil {
            return fmt.Errorf("create backend %s: %w", name, err)
        }

        // Configure if possible
        if configurable, ok := backend.(ConfigurableBackend); ok {
            raw, err := backendCfg.RawConfig()
            if err != nil {
                return fmt.Errorf("serialize config for %s: %w", name, err)
            }
            if raw != nil {
                if err := configurable.Configure(raw); err != nil {
                    return fmt.Errorf("configure backend %s: %w", name, err)
                }
            }
        }

        // Register
        if err := r.Register(backend); err != nil {
            return fmt.Errorf("register backend %s: %w", name, err)
        }
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/backend/... -run TestBackendConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/backend/config.go internal/backend/config_test.go
git commit -m "$(cat <<'EOF'
feat(backend): add configuration support

- BackendsConfig and BackendConfig types
- ParseBackendsConfig for YAML parsing
- Validate() for config validation
- LoadFromConfig to create backends from config
- RawConfig() for passing to ConfigurableBackend

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Integrate with Server

**Files:**
- Create: `internal/server/backend_adapter.go`
- Test: `internal/server/backend_adapter_test.go`

**Step 1: Write integration test**

```go
// internal/server/backend_adapter_test.go
package server

import (
    "context"
    "testing"

    "github.com/your-org/metatools-mcp/internal/backend"
    "github.com/your-org/metatools-mcp/internal/backend/local"
)

func TestBackendAdapter_GetTools(t *testing.T) {
    registry := backend.NewRegistry()

    // Create local backend with a test tool
    localBackend := local.New("local")
    localBackend.RegisterHandler("echo", local.ToolDef{
        Name:        "echo",
        Description: "Echo input",
        Handler: func(ctx context.Context, args map[string]any) (any, error) {
            return args["message"], nil
        },
    })

    registry.Register(localBackend)

    adapter := NewBackendAdapter(registry)

    tools, err := adapter.GetTools(context.Background())
    if err != nil {
        t.Fatalf("GetTools() error = %v", err)
    }

    if len(tools) == 0 {
        t.Error("GetTools() returned empty tools")
    }
}

func TestBackendAdapter_Execute(t *testing.T) {
    registry := backend.NewRegistry()

    localBackend := local.New("local")
    localBackend.RegisterHandler("echo", local.ToolDef{
        Name:        "echo",
        Description: "Echo input",
        Handler: func(ctx context.Context, args map[string]any) (any, error) {
            return args["message"], nil
        },
    })

    registry.Register(localBackend)

    adapter := NewBackendAdapter(registry)

    result, err := adapter.Execute(context.Background(), "local/echo", map[string]any{
        "message": "hello world",
    })

    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    if result != "hello world" {
        t.Errorf("Execute() = %v, want %v", result, "hello world")
    }
}
```

**Step 2: Implement adapter**

```go
// internal/server/backend_adapter.go
package server

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/backend"
)

// BackendAdapter adapts the backend registry for the MCP server
type BackendAdapter struct {
    registry   *backend.Registry
    aggregator *backend.Aggregator
}

// NewBackendAdapter creates a new adapter
func NewBackendAdapter(registry *backend.Registry) *BackendAdapter {
    return &BackendAdapter{
        registry:   registry,
        aggregator: backend.NewAggregator(registry),
    }
}

// GetTools returns all tools from all enabled backends
func (a *BackendAdapter) GetTools(ctx context.Context) ([]mcp.Tool, error) {
    tools, err := a.aggregator.ListAllTools(ctx)
    if err != nil {
        return nil, err
    }

    mcpTools := make([]mcp.Tool, 0, len(tools))
    for _, t := range tools {
        // Use qualified ID (backend/tool)
        qualifiedName := backend.FormatToolID(t.Namespace, t.Name)
        mcpTool := mcp.Tool{
            Name:        qualifiedName,
            Description: t.Description,
            InputSchema: t.InputSchema,
        }
        mcpTools = append(mcpTools, mcpTool)
    }

    return mcpTools, nil
}

// Execute invokes a tool through the backend registry
func (a *BackendAdapter) Execute(ctx context.Context, toolID string, args map[string]any) (any, error) {
    return a.aggregator.Execute(ctx, toolID, args)
}

// CreateToolHandler creates an MCP tool handler for backend tools
func (a *BackendAdapter) CreateToolHandler() mcp.ToolHandler {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        result, err := a.Execute(ctx, req.Params.Name, req.Params.Arguments)
        if err != nil {
            return nil, err
        }

        // Format result as JSON
        text, err := formatAsText(result)
        if err != nil {
            return nil, err
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{
                mcp.TextContent{
                    Type: "text",
                    Text: text,
                },
            },
        }, nil
    }
}

// formatAsText converts a result to string
func formatAsText(result any) (string, error) {
    switch v := result.(type) {
    case string:
        return v, nil
    case []byte:
        return string(v), nil
    default:
        data, err := json.MarshalIndent(result, "", "  ")
        if err != nil {
            return fmt.Sprintf("%v", result), nil
        }
        return string(data), nil
    }
}

// Start starts all backends
func (a *BackendAdapter) Start(ctx context.Context) error {
    return a.registry.StartAll(ctx)
}

// Stop stops all backends
func (a *BackendAdapter) Stop() error {
    return a.registry.StopAll()
}
```

**Step 3: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/server/... -run TestBackendAdapter -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/server/backend_adapter.go internal/server/backend_adapter_test.go
git commit -m "$(cat <<'EOF'
feat(server): add BackendAdapter for registry integration

- GetTools aggregates tools from all backends
- Execute routes to correct backend by tool ID
- CreateToolHandler for MCP server integration
- Start/Stop for backend lifecycle

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Backend interface defined with Kind, Name, Enabled, ListTools, Execute
- [ ] ConfigurableBackend interface for YAML config
- [ ] StreamingBackend interface for streaming responses
- [ ] Registry with Register/Get/List/ListByKind
- [ ] LocalBackend implementation
- [ ] Aggregator for multi-backend tool listing and execution
- [ ] ParseToolID/FormatToolID for qualified IDs
- [ ] BackendsConfig for YAML configuration
- [ ] BackendAdapter for server integration
- [ ] All tests pass

## Definition of Done

1. All tests pass: `go test ./internal/backend/...`
2. Backend interface enables multiple tool sources
3. LocalBackend works for in-process handlers
4. Aggregator combines tools from all backends
5. Configuration supports YAML-based backend definition
6. Server can use backends for tool discovery and execution

## Next PRD

PRD-007 will implement the Middleware Chain for cross-cutting concerns (logging, auth, rate limiting).
