# PRD-143: Extract toolbackend

**Phase:** 4 - Execution Layer
**Priority:** High
**Effort:** 6 hours
**Dependencies:** PRD-120

---

## Objective

Extract backend management code from `metatools-mcp` into `toolexec/backend/` as the fourth package in the consolidated execution layer.

---

## Source Analysis

**Current Location:** `metatools-mcp/internal/backend/` (embedded in MCP server)
**Target Location:** `github.com/ApertureStack/toolexec/backend`

**Code to Extract:**
- Backend registry and management
- Provider interface for tool backends
- Multi-backend aggregation
- Backend health checking
- ~600 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Backend Package | `toolexec/backend/` | Backend management |
| Tests | `toolexec/backend/*_test.go` | Comprehensive tests |
| Documentation | `toolexec/backend/doc.go` | Package documentation |

---

## Tasks

### Task 1: Analyze Source Code

```bash
cd /Users/jraymond/Documents/Projects/ApertureStack/metatools-mcp

# Find backend-related code
find . -name "*.go" -exec grep -l "Backend\|Provider" {} \;

# Count lines in relevant files
wc -l internal/backend/*.go 2>/dev/null || echo "Check actual path"
```

### Task 2: Create Package Structure

```bash
cd /tmp/migration/toolexec

mkdir -p backend
```

### Task 3: Define Core Interfaces

**File:** `toolexec/backend/backend.go`

```go
package backend

import (
    "context"
    "github.com/ApertureStack/toolfoundation/model"
)

// Backend represents a tool execution backend.
type Backend interface {
    // Name returns the backend identifier.
    Name() string

    // Type returns the backend type (local, mcp, http, grpc).
    Type() string

    // ListTools returns available tools from this backend.
    ListTools(ctx context.Context) ([]model.Tool, error)

    // GetTool retrieves a specific tool by ID.
    GetTool(ctx context.Context, id string) (*model.Tool, error)

    // Execute runs a tool with the given input.
    Execute(ctx context.Context, toolID string, input map[string]any) (any, error)

    // Health returns the backend health status.
    Health(ctx context.Context) (*HealthStatus, error)

    // Close releases backend resources.
    Close() error
}

// HealthStatus represents backend health.
type HealthStatus struct {
    Healthy     bool
    Message     string
    LastChecked time.Time
    Latency     time.Duration
}
```

### Task 4: Implement Registry

**File:** `toolexec/backend/registry.go`

```go
package backend

import (
    "context"
    "fmt"
    "sync"

    "github.com/ApertureStack/toolfoundation/model"
)

// Registry manages multiple backends.
type Registry struct {
    backends map[string]Backend
    mu       sync.RWMutex
}

// NewRegistry creates a new backend registry.
func NewRegistry() *Registry {
    return &Registry{
        backends: make(map[string]Backend),
    }
}

// Register adds a backend to the registry.
func (r *Registry) Register(backend Backend) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := backend.Name()
    if _, exists := r.backends[name]; exists {
        return fmt.Errorf("backend %q already registered", name)
    }
    r.backends[name] = backend
    return nil
}

// Unregister removes a backend from the registry.
func (r *Registry) Unregister(name string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    backend, exists := r.backends[name]
    if !exists {
        return fmt.Errorf("backend %q not found", name)
    }

    if err := backend.Close(); err != nil {
        return fmt.Errorf("closing backend %q: %w", name, err)
    }

    delete(r.backends, name)
    return nil
}

// Get retrieves a backend by name.
func (r *Registry) Get(name string) (Backend, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    b, ok := r.backends[name]
    return b, ok
}

// List returns all registered backend names.
func (r *Registry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    names := make([]string, 0, len(r.backends))
    for name := range r.backends {
        names = append(names, name)
    }
    return names
}

// ListAllTools aggregates tools from all backends.
func (r *Registry) ListAllTools(ctx context.Context) ([]model.Tool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var allTools []model.Tool
    for _, backend := range r.backends {
        tools, err := backend.ListTools(ctx)
        if err != nil {
            continue // Skip failing backends
        }
        allTools = append(allTools, tools...)
    }
    return allTools, nil
}

// FindTool searches all backends for a tool.
func (r *Registry) FindTool(ctx context.Context, toolID string) (*model.Tool, Backend, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, backend := range r.backends {
        tool, err := backend.GetTool(ctx, toolID)
        if err == nil && tool != nil {
            return tool, backend, nil
        }
    }
    return nil, nil, fmt.Errorf("tool %q not found in any backend", toolID)
}

// HealthCheck returns health status for all backends.
func (r *Registry) HealthCheck(ctx context.Context) map[string]*HealthStatus {
    r.mu.RLock()
    defer r.mu.RUnlock()

    results := make(map[string]*HealthStatus)
    for name, backend := range r.backends {
        status, err := backend.Health(ctx)
        if err != nil {
            results[name] = &HealthStatus{
                Healthy: false,
                Message: err.Error(),
            }
        } else {
            results[name] = status
        }
    }
    return results
}

// Close closes all backends.
func (r *Registry) Close() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    var errs []error
    for name, backend := range r.backends {
        if err := backend.Close(); err != nil {
            errs = append(errs, fmt.Errorf("closing %s: %w", name, err))
        }
    }
    r.backends = make(map[string]Backend)

    if len(errs) > 0 {
        return fmt.Errorf("errors closing backends: %v", errs)
    }
    return nil
}
```

### Task 5: Implement Local Backend

**File:** `toolexec/backend/local.go`

```go
package backend

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/ApertureStack/toolfoundation/model"
)

// LocalBackend provides in-process tool execution.
type LocalBackend struct {
    name     string
    tools    map[string]model.Tool
    handlers map[string]Handler
    mu       sync.RWMutex
}

// Handler executes a tool.
type Handler func(ctx context.Context, input map[string]any) (any, error)

// NewLocalBackend creates a new local backend.
func NewLocalBackend(name string) *LocalBackend {
    return &LocalBackend{
        name:     name,
        tools:    make(map[string]model.Tool),
        handlers: make(map[string]Handler),
    }
}

func (b *LocalBackend) Name() string { return b.name }
func (b *LocalBackend) Type() string { return "local" }

// RegisterTool adds a tool with its handler.
func (b *LocalBackend) RegisterTool(tool model.Tool, handler Handler) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    if _, exists := b.tools[tool.ID]; exists {
        return fmt.Errorf("tool %q already registered", tool.ID)
    }
    b.tools[tool.ID] = tool
    b.handlers[tool.ID] = handler
    return nil
}

func (b *LocalBackend) ListTools(ctx context.Context) ([]model.Tool, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    tools := make([]model.Tool, 0, len(b.tools))
    for _, tool := range b.tools {
        tools = append(tools, tool)
    }
    return tools, nil
}

func (b *LocalBackend) GetTool(ctx context.Context, id string) (*model.Tool, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    tool, ok := b.tools[id]
    if !ok {
        return nil, fmt.Errorf("tool %q not found", id)
    }
    return &tool, nil
}

func (b *LocalBackend) Execute(ctx context.Context, toolID string, input map[string]any) (any, error) {
    b.mu.RLock()
    handler, ok := b.handlers[toolID]
    b.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("tool %q not found", toolID)
    }

    return handler(ctx, input)
}

func (b *LocalBackend) Health(ctx context.Context) (*HealthStatus, error) {
    return &HealthStatus{
        Healthy:     true,
        Message:     "local backend healthy",
        LastChecked: time.Now(),
        Latency:     0,
    }, nil
}

func (b *LocalBackend) Close() error {
    return nil
}
```

### Task 6: Create Package Documentation

**File:** `toolexec/backend/doc.go`

```go
// Package backend provides backend management for tool execution.
//
// This package implements the registry pattern for managing multiple tool
// execution backends. It supports local, MCP, HTTP, and gRPC backends with
// unified discovery and execution APIs.
//
// # Registry
//
// The Registry aggregates multiple backends:
//
//	registry := backend.NewRegistry()
//	registry.Register(localBackend)
//	registry.Register(mcpBackend)
//	registry.Register(httpBackend)
//
//	// List tools from all backends
//	tools, _ := registry.ListAllTools(ctx)
//
//	// Find and execute a tool
//	tool, backend, _ := registry.FindTool(ctx, "calculator")
//	result, _ := backend.Execute(ctx, tool.ID, input)
//
// # Backend Types
//
// Built-in backend implementations:
//
//   - LocalBackend: In-process tool execution
//   - MCPBackend: MCP server tool provider
//   - HTTPBackend: HTTP API tool provider
//   - GRPCBackend: gRPC tool provider
//
// # Health Checking
//
// Monitor backend health:
//
//	status := registry.HealthCheck(ctx)
//	for name, health := range status {
//	    fmt.Printf("%s: healthy=%v latency=%v\n",
//	        name, health.Healthy, health.Latency)
//	}
//
// # Custom Backends
//
// Implement the Backend interface for custom backends:
//
//	type MyBackend struct {}
//	func (b *MyBackend) Name() string { return "my-backend" }
//	func (b *MyBackend) Type() string { return "custom" }
//	func (b *MyBackend) ListTools(ctx context.Context) ([]model.Tool, error) { ... }
//	func (b *MyBackend) Execute(ctx context.Context, toolID string, input map[string]any) (any, error) { ... }
//
// # Extraction Note
//
// This package was extracted from metatools-mcp/internal/backend as part of
// the ApertureStack consolidation to enable reuse across projects.
package backend
```

### Task 7: Create Tests

**File:** `toolexec/backend/backend_test.go`

```go
package backend

import (
    "context"
    "testing"

    "github.com/ApertureStack/toolfoundation/model"
)

func TestRegistry(t *testing.T) {
    ctx := context.Background()
    registry := NewRegistry()

    // Create local backend with a tool
    local := NewLocalBackend("local")
    local.RegisterTool(
        model.Tool{ID: "test", Name: "Test Tool"},
        func(ctx context.Context, input map[string]any) (any, error) {
            return "success", nil
        },
    )

    // Register
    if err := registry.Register(local); err != nil {
        t.Fatal(err)
    }

    // List
    names := registry.List()
    if len(names) != 1 || names[0] != "local" {
        t.Errorf("expected [local], got %v", names)
    }

    // Get
    b, ok := registry.Get("local")
    if !ok || b.Name() != "local" {
        t.Error("failed to get backend")
    }

    // Find tool
    tool, backend, err := registry.FindTool(ctx, "test")
    if err != nil {
        t.Fatal(err)
    }
    if tool.ID != "test" || backend.Name() != "local" {
        t.Error("wrong tool or backend")
    }

    // Execute
    result, err := backend.Execute(ctx, "test", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result != "success" {
        t.Errorf("expected 'success', got %v", result)
    }
}

func TestLocalBackend(t *testing.T) {
    ctx := context.Background()
    backend := NewLocalBackend("test")

    // Register tool
    tool := model.Tool{ID: "calc", Name: "Calculator"}
    handler := func(ctx context.Context, input map[string]any) (any, error) {
        a := input["a"].(float64)
        b := input["b"].(float64)
        return a + b, nil
    }

    if err := backend.RegisterTool(tool, handler); err != nil {
        t.Fatal(err)
    }

    // List tools
    tools, err := backend.ListTools(ctx)
    if err != nil {
        t.Fatal(err)
    }
    if len(tools) != 1 {
        t.Errorf("expected 1 tool, got %d", len(tools))
    }

    // Execute
    result, err := backend.Execute(ctx, "calc", map[string]any{"a": 1.0, "b": 2.0})
    if err != nil {
        t.Fatal(err)
    }
    if result != 3.0 {
        t.Errorf("expected 3.0, got %v", result)
    }
}
```

### Task 8: Build and Test

```bash
cd /tmp/migration/toolexec

go mod tidy
go build ./...
go test -v -coverprofile=backend_coverage.out ./backend/...

go tool cover -func=backend_coverage.out | grep total
```

### Task 9: Commit and Push

```bash
cd /tmp/migration/toolexec

git add -A
git commit -m "feat(backend): extract backend management package

Extract backend registry and management from metatools-mcp.

Package contents:
- Backend interface for tool providers
- Registry for multi-backend management
- LocalBackend for in-process execution
- Health checking for all backends
- Tool aggregation across backends

Features:
- Register/unregister backends dynamically
- Find tools across all backends
- Execute tools via appropriate backend
- Health monitoring with latency tracking

Dependencies:
- github.com/ApertureStack/toolfoundation/model

This extraction enables backend reuse across projects.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [ ] Backend interface defined
- [ ] Registry implemented
- [ ] LocalBackend implemented
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Health checking works
- [ ] Tool aggregation works
- [ ] Package documentation complete

---

## Acceptance Criteria

1. `toolexec/backend` package builds successfully
2. All tests pass with >= 80% coverage
3. Registry manages multiple backends
4. Tools can be found across backends
5. Health status is accurate

---

## Rollback Plan

```bash
cd /tmp/migration/toolexec
rm -rf backend/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- Gate G3: Execution layer complete (all 4 packages)
- PRD-150: Migrate toolset
