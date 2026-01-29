# PRD-007: Middleware Chain

> **Implementation note:** In code, `MiddlewareConfig` is named `Config` and `MiddlewareEntry` is named `Entry` to avoid lint stutter.

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a Middleware Chain enabling cross-cutting concerns (logging, auth, rate limiting, metrics) to be applied to tool providers through composable, configuration-driven middleware functions.

**Architecture:** Define a Middleware type as a function that wraps a ToolProvider. Create a MiddlewareRegistry for managing available middleware. Implement core middleware (logging, metrics). Enable configuration-driven chain construction with per-tool customization.

**Tech Stack:** Go interfaces, provider integration from PRD-005, config integration from PRD-003

**Priority:** P2 - Stream A, Phase 5 (completes core exposure)

**Scope:** Middleware interface + registry + chain builder + logging middleware + metrics middleware

**Dependencies:** PRD-002 (CLI), PRD-003 (Config), PRD-005 (Provider Registry)

---

## Context

The current architecture lacks cross-cutting concerns that are essential for production deployments:
1. Request/response logging for debugging and audit
2. Metrics collection for observability
3. Rate limiting for protection
4. Authentication for security

**Current State:**
```go
// No middleware - each provider handles concerns independently
provider.Handle(ctx, args)  // No logging, no metrics
```

**Target State:**
```go
// Middleware wraps providers with cross-cutting concerns
type Middleware func(provider.ToolProvider) provider.ToolProvider

// Chain applies multiple middleware in order
wrapped := chain.Apply(
    LoggingMiddleware,
    MetricsMiddleware,
    provider,
)
```

---

## Tasks

### Task 1: Define Middleware Type and Registry

**Files:**
- Create: `internal/middleware/middleware.go`
- Test: `internal/middleware/middleware_test.go`

**Step 1: Write failing test for Middleware type**

```go
// internal/middleware/middleware_test.go
package middleware

import (
    "context"
    "testing"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// mockProvider for testing
type mockProvider struct {
    name     string
    enabled  bool
    handleFn func(ctx context.Context, args map[string]any) (any, error)
}

func (m *mockProvider) Name() string       { return m.name }
func (m *mockProvider) Enabled() bool      { return m.enabled }
func (m *mockProvider) Tool() mcp.Tool     { return mcp.Tool{Name: m.name} }
func (m *mockProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    if m.handleFn != nil {
        return m.handleFn(ctx, args)
    }
    return nil, nil
}

func TestMiddleware_Wrapping(t *testing.T) {
    called := false

    // Simple middleware that sets a flag
    mw := func(next provider.ToolProvider) provider.ToolProvider {
        return &wrappedProvider{
            ToolProvider: next,
            beforeFn: func() {
                called = true
            },
        }
    }

    original := &mockProvider{name: "test", enabled: true}
    wrapped := mw(original)

    wrapped.Handle(context.Background(), nil)

    if !called {
        t.Error("Middleware was not invoked")
    }
}

func TestMiddleware_ChainOrder(t *testing.T) {
    var order []string

    mw1 := func(next provider.ToolProvider) provider.ToolProvider {
        return &wrappedProvider{
            ToolProvider: next,
            beforeFn:     func() { order = append(order, "mw1-before") },
            afterFn:      func() { order = append(order, "mw1-after") },
        }
    }

    mw2 := func(next provider.ToolProvider) provider.ToolProvider {
        return &wrappedProvider{
            ToolProvider: next,
            beforeFn:     func() { order = append(order, "mw2-before") },
            afterFn:      func() { order = append(order, "mw2-after") },
        }
    }

    original := &mockProvider{
        name:    "test",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            order = append(order, "handler")
            return nil, nil
        },
    }

    // Chain: mw1 -> mw2 -> handler
    chain := NewChain(mw1, mw2)
    wrapped := chain.Apply(original)

    wrapped.Handle(context.Background(), nil)

    expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
    if len(order) != len(expected) {
        t.Errorf("Order = %v, want %v", order, expected)
    }
    for i, v := range expected {
        if order[i] != v {
            t.Errorf("Order[%d] = %q, want %q", i, order[i], v)
        }
    }
}

// wrappedProvider for testing middleware behavior
type wrappedProvider struct {
    provider.ToolProvider
    beforeFn func()
    afterFn  func()
}

func (w *wrappedProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    if w.beforeFn != nil {
        w.beforeFn()
    }
    result, err := w.ToolProvider.Handle(ctx, args)
    if w.afterFn != nil {
        w.afterFn()
    }
    return result, err
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -v`
Expected: FAIL - Middleware types don't exist

**Step 3: Implement Middleware types**

```go
// internal/middleware/middleware.go
package middleware

import (
    "github.com/your-org/metatools-mcp/internal/provider"
)

// Middleware wraps a ToolProvider to add cross-cutting concerns.
// Middleware functions receive the next provider in the chain and return
// a wrapped provider that adds behavior before/after the next provider.
type Middleware func(provider.ToolProvider) provider.ToolProvider

// Chain holds an ordered list of middleware to apply.
type Chain struct {
    middleware []Middleware
}

// NewChain creates a new middleware chain.
func NewChain(middleware ...Middleware) *Chain {
    return &Chain{middleware: middleware}
}

// Use adds middleware to the chain.
func (c *Chain) Use(mw Middleware) *Chain {
    c.middleware = append(c.middleware, mw)
    return c
}

// Apply wraps a provider with all middleware in the chain.
// Middleware is applied in order: first middleware wraps outermost.
// Request flow: mw1 -> mw2 -> ... -> provider
// Response flow: provider -> ... -> mw2 -> mw1
func (c *Chain) Apply(p provider.ToolProvider) provider.ToolProvider {
    wrapped := p
    // Apply in reverse order so first middleware is outermost
    for i := len(c.middleware) - 1; i >= 0; i-- {
        wrapped = c.middleware[i](wrapped)
    }
    return wrapped
}

// ApplyToRegistry wraps all providers in a registry with the chain.
func (c *Chain) ApplyToRegistry(registry *provider.Registry) {
    for _, p := range registry.List() {
        wrapped := c.Apply(p)
        registry.Unregister(p.Name())
        registry.Register(wrapped)
    }
}

// Len returns the number of middleware in the chain.
func (c *Chain) Len() int {
    return len(c.middleware)
}

// Clear removes all middleware from the chain.
func (c *Chain) Clear() {
    c.middleware = nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/middleware.go internal/middleware/middleware_test.go
git commit -m "$(cat <<'EOF'
feat(middleware): define Middleware type and Chain

- Middleware type wraps ToolProvider
- Chain holds ordered middleware list
- Apply wraps provider with all middleware
- ApplyToRegistry for bulk wrapping

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Implement Middleware Registry

**Files:**
- Create: `internal/middleware/registry.go`
- Test: `internal/middleware/registry_test.go`

**Step 1: Write failing test for registry**

```go
// internal/middleware/registry_test.go
package middleware

import (
    "testing"
)

func TestRegistry_Register(t *testing.T) {
    registry := NewRegistry()

    factory := func(cfg map[string]any) (Middleware, error) {
        return func(next provider.ToolProvider) provider.ToolProvider {
            return next
        }, nil
    }

    registry.Register("logging", factory)

    if !registry.Has("logging") {
        t.Error("Has(logging) = false, want true")
    }
}

func TestRegistry_Get(t *testing.T) {
    registry := NewRegistry()

    factory := func(cfg map[string]any) (Middleware, error) {
        return func(next provider.ToolProvider) provider.ToolProvider {
            return next
        }, nil
    }

    registry.Register("test", factory)

    got, ok := registry.Get("test")
    if !ok {
        t.Fatal("Get() returned false")
    }
    if got == nil {
        t.Error("Get() returned nil factory")
    }

    _, ok = registry.Get("nonexistent")
    if ok {
        t.Error("Get() should return false for nonexistent middleware")
    }
}

func TestRegistry_Create(t *testing.T) {
    registry := NewRegistry()

    called := false
    factory := func(cfg map[string]any) (Middleware, error) {
        called = true
        return func(next provider.ToolProvider) provider.ToolProvider {
            return next
        }, nil
    }

    registry.Register("test", factory)

    mw, err := registry.Create("test", nil)
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    if mw == nil {
        t.Error("Create() returned nil middleware")
    }
    if !called {
        t.Error("Factory was not called")
    }
}

func TestRegistry_List(t *testing.T) {
    registry := NewRegistry()

    registry.Register("a", nil)
    registry.Register("b", nil)
    registry.Register("c", nil)

    names := registry.List()
    if len(names) != 3 {
        t.Errorf("List() returned %d names, want 3", len(names))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestRegistry -v`
Expected: FAIL - Registry doesn't exist

**Step 3: Implement Registry**

```go
// internal/middleware/registry.go
package middleware

import (
    "errors"
    "fmt"
    "sort"
    "sync"
)

// ErrMiddlewareNotFound is returned when middleware is not registered.
var ErrMiddlewareNotFound = errors.New("middleware not found")

// Factory creates a middleware instance from configuration.
type Factory func(cfg map[string]any) (Middleware, error)

// Registry manages middleware factories.
type Registry struct {
    factories map[string]Factory
    mu        sync.RWMutex
}

// NewRegistry creates a new middleware registry.
func NewRegistry() *Registry {
    return &Registry{
        factories: make(map[string]Factory),
    }
}

// Register adds a middleware factory.
func (r *Registry) Register(name string, factory Factory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.factories[name] = factory
}

// Has checks if a middleware is registered.
func (r *Registry) Has(name string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    _, ok := r.factories[name]
    return ok
}

// Get retrieves a middleware factory by name.
func (r *Registry) Get(name string) (Factory, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    f, ok := r.factories[name]
    return f, ok
}

// Create instantiates middleware with configuration.
func (r *Registry) Create(name string, cfg map[string]any) (Middleware, error) {
    factory, ok := r.Get(name)
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrMiddlewareNotFound, name)
    }

    return factory(cfg)
}

// List returns all registered middleware names.
func (r *Registry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    names := make([]string, 0, len(r.factories))
    for name := range r.factories {
        names = append(names, name)
    }
    sort.Strings(names)
    return names
}

// Clear removes all registered middleware.
func (r *Registry) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.factories = make(map[string]Factory)
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestRegistry -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/registry.go internal/middleware/registry_test.go
git commit -m "$(cat <<'EOF'
feat(middleware): implement Middleware Registry

- Factory type for middleware creation
- Register/Get/Has/Create for factory management
- List returns all registered middleware names
- Thread-safe with RWMutex

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement Logging Middleware

**Files:**
- Create: `internal/middleware/logging.go`
- Test: `internal/middleware/logging_test.go`

**Step 1: Write failing test for LoggingMiddleware**

```go
// internal/middleware/logging_test.go
package middleware

import (
    "bytes"
    "context"
    "log/slog"
    "strings"
    "testing"
)

func TestLoggingMiddleware(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))

    mw := NewLoggingMiddleware(LoggingConfig{
        Logger: logger,
    })

    original := &mockProvider{
        name:    "test_tool",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            return "result", nil
        },
    }

    wrapped := mw(original)

    _, err := wrapped.Handle(context.Background(), map[string]any{"key": "value"})
    if err != nil {
        t.Fatalf("Handle() error = %v", err)
    }

    output := buf.String()
    if !strings.Contains(output, "test_tool") {
        t.Errorf("Log output missing tool name: %s", output)
    }
}

func TestLoggingMiddleware_Error(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))

    mw := NewLoggingMiddleware(LoggingConfig{
        Logger: logger,
    })

    original := &mockProvider{
        name:    "failing_tool",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            return nil, errors.New("tool failed")
        },
    }

    wrapped := mw(original)

    _, err := wrapped.Handle(context.Background(), nil)
    if err == nil {
        t.Fatal("Handle() should return error")
    }

    output := buf.String()
    if !strings.Contains(output, "error") || !strings.Contains(output, "tool failed") {
        t.Errorf("Log output missing error: %s", output)
    }
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewTextHandler(&buf, nil))

    mw := NewLoggingMiddleware(LoggingConfig{
        Logger: logger,
    })

    original := &mockProvider{name: "test", enabled: true}
    wrapped := mw(original)

    wrapped.Handle(context.Background(), nil)

    output := buf.String()
    if !strings.Contains(output, "request_id") {
        t.Errorf("Log output missing request_id: %s", output)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestLoggingMiddleware -v`
Expected: FAIL - LoggingMiddleware doesn't exist

**Step 3: Implement LoggingMiddleware**

```go
// internal/middleware/logging.go
package middleware

import (
    "context"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// LoggingConfig configures the logging middleware.
type LoggingConfig struct {
    Logger              *slog.Logger
    Level               slog.Level
    IncludeRequestBody  bool
    IncludeResponseBody bool
}

// loggingProvider wraps a provider with logging.
type loggingProvider struct {
    provider.ToolProvider
    cfg LoggingConfig
}

// NewLoggingMiddleware creates a middleware that logs requests and responses.
func NewLoggingMiddleware(cfg LoggingConfig) Middleware {
    if cfg.Logger == nil {
        cfg.Logger = slog.Default()
    }

    return func(next provider.ToolProvider) provider.ToolProvider {
        return &loggingProvider{
            ToolProvider: next,
            cfg:          cfg,
        }
    }
}

// Handle logs before and after calling the wrapped provider.
func (p *loggingProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    requestID := uuid.New().String()
    start := time.Now()

    // Log request
    logAttrs := []any{
        slog.String("request_id", requestID),
        slog.String("tool", p.ToolProvider.Name()),
    }

    if p.cfg.IncludeRequestBody {
        logAttrs = append(logAttrs, slog.Any("args", args))
    }

    p.cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "tool request", logAttrs...)

    // Execute
    result, err := p.ToolProvider.Handle(ctx, args)

    // Log response
    duration := time.Since(start)
    logAttrs = []any{
        slog.String("request_id", requestID),
        slog.String("tool", p.ToolProvider.Name()),
        slog.Duration("duration", duration),
    }

    if err != nil {
        logAttrs = append(logAttrs,
            slog.String("error", err.Error()),
            slog.Bool("success", false),
        )
        p.cfg.Logger.LogAttrs(ctx, slog.LevelError, "tool error", logAttrs...)
    } else {
        logAttrs = append(logAttrs, slog.Bool("success", true))
        if p.cfg.IncludeResponseBody {
            logAttrs = append(logAttrs, slog.Any("result", result))
        }
        p.cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "tool response", logAttrs...)
    }

    return result, err
}

// Name returns the wrapped provider's name.
func (p *loggingProvider) Name() string {
    return p.ToolProvider.Name()
}

// Enabled returns the wrapped provider's enabled status.
func (p *loggingProvider) Enabled() bool {
    return p.ToolProvider.Enabled()
}

// Tool returns the wrapped provider's tool definition.
func (p *loggingProvider) Tool() mcp.Tool {
    return p.ToolProvider.Tool()
}

// LoggingMiddlewareFactory creates a logging middleware from config.
func LoggingMiddlewareFactory(cfg map[string]any) (Middleware, error) {
    config := LoggingConfig{
        Logger: slog.Default(),
    }

    if level, ok := cfg["level"].(string); ok {
        var lvl slog.Level
        if err := lvl.UnmarshalText([]byte(level)); err == nil {
            config.Level = lvl
        }
    }

    if includeReq, ok := cfg["include_request_body"].(bool); ok {
        config.IncludeRequestBody = includeReq
    }

    if includeResp, ok := cfg["include_response_body"].(bool); ok {
        config.IncludeResponseBody = includeResp
    }

    return NewLoggingMiddleware(config), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestLoggingMiddleware -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/logging.go internal/middleware/logging_test.go
git commit -m "$(cat <<'EOF'
feat(middleware): implement LoggingMiddleware

- Logs request start with tool name and request_id
- Logs response with duration, success, and error
- Configurable request/response body inclusion
- Uses slog for structured logging

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Implement Metrics Middleware

**Files:**
- Create: `internal/middleware/metrics.go`
- Test: `internal/middleware/metrics_test.go`

**Step 1: Write failing test for MetricsMiddleware**

```go
// internal/middleware/metrics_test.go
package middleware

import (
    "context"
    "errors"
    "testing"
    "time"
)

func TestMetricsMiddleware(t *testing.T) {
    collector := NewInMemoryMetricsCollector()

    mw := NewMetricsMiddleware(MetricsConfig{
        Collector: collector,
    })

    original := &mockProvider{
        name:    "test_tool",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            time.Sleep(10 * time.Millisecond) // Simulate work
            return "result", nil
        },
    }

    wrapped := mw(original)

    _, err := wrapped.Handle(context.Background(), nil)
    if err != nil {
        t.Fatalf("Handle() error = %v", err)
    }

    // Check metrics
    metrics := collector.GetMetrics("test_tool")
    if metrics.TotalRequests != 1 {
        t.Errorf("TotalRequests = %d, want 1", metrics.TotalRequests)
    }
    if metrics.SuccessCount != 1 {
        t.Errorf("SuccessCount = %d, want 1", metrics.SuccessCount)
    }
    if metrics.ErrorCount != 0 {
        t.Errorf("ErrorCount = %d, want 0", metrics.ErrorCount)
    }
    if metrics.LastDuration < 10*time.Millisecond {
        t.Errorf("LastDuration = %v, want >= 10ms", metrics.LastDuration)
    }
}

func TestMetricsMiddleware_Error(t *testing.T) {
    collector := NewInMemoryMetricsCollector()

    mw := NewMetricsMiddleware(MetricsConfig{
        Collector: collector,
    })

    original := &mockProvider{
        name:    "failing_tool",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            return nil, errors.New("failed")
        },
    }

    wrapped := mw(original)
    wrapped.Handle(context.Background(), nil)

    metrics := collector.GetMetrics("failing_tool")
    if metrics.ErrorCount != 1 {
        t.Errorf("ErrorCount = %d, want 1", metrics.ErrorCount)
    }
    if metrics.SuccessCount != 0 {
        t.Errorf("SuccessCount = %d, want 0", metrics.SuccessCount)
    }
}

func TestMetricsMiddleware_ActiveRequests(t *testing.T) {
    collector := NewInMemoryMetricsCollector()

    mw := NewMetricsMiddleware(MetricsConfig{
        Collector: collector,
    })

    started := make(chan struct{})
    done := make(chan struct{})

    original := &mockProvider{
        name:    "slow_tool",
        enabled: true,
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            close(started)
            <-done
            return nil, nil
        },
    }

    wrapped := mw(original)

    go wrapped.Handle(context.Background(), nil)

    <-started // Wait for handler to start

    metrics := collector.GetMetrics("slow_tool")
    if metrics.ActiveRequests != 1 {
        t.Errorf("ActiveRequests = %d, want 1", metrics.ActiveRequests)
    }

    close(done) // Allow handler to complete
    time.Sleep(10 * time.Millisecond) // Let goroutine finish

    metrics = collector.GetMetrics("slow_tool")
    if metrics.ActiveRequests != 0 {
        t.Errorf("ActiveRequests after completion = %d, want 0", metrics.ActiveRequests)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestMetricsMiddleware -v`
Expected: FAIL - MetricsMiddleware doesn't exist

**Step 3: Implement MetricsMiddleware**

```go
// internal/middleware/metrics.go
package middleware

import (
    "context"
    "sync"
    "sync/atomic"
    "time"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// ToolMetrics holds metrics for a single tool.
type ToolMetrics struct {
    TotalRequests  int64
    SuccessCount   int64
    ErrorCount     int64
    ActiveRequests int64
    TotalDuration  time.Duration
    LastDuration   time.Duration
}

// MetricsCollector defines the interface for collecting metrics.
type MetricsCollector interface {
    RecordRequest(tool string)
    RecordSuccess(tool string, duration time.Duration)
    RecordError(tool string, duration time.Duration)
    RecordActive(tool string, delta int64)
    GetMetrics(tool string) ToolMetrics
}

// InMemoryMetricsCollector stores metrics in memory.
type InMemoryMetricsCollector struct {
    metrics map[string]*toolMetricsState
    mu      sync.RWMutex
}

type toolMetricsState struct {
    totalRequests  atomic.Int64
    successCount   atomic.Int64
    errorCount     atomic.Int64
    activeRequests atomic.Int64
    totalDuration  atomic.Int64 // nanoseconds
    lastDuration   atomic.Int64 // nanoseconds
}

// NewInMemoryMetricsCollector creates an in-memory collector.
func NewInMemoryMetricsCollector() *InMemoryMetricsCollector {
    return &InMemoryMetricsCollector{
        metrics: make(map[string]*toolMetricsState),
    }
}

func (c *InMemoryMetricsCollector) getOrCreate(tool string) *toolMetricsState {
    c.mu.RLock()
    state, ok := c.metrics[tool]
    c.mu.RUnlock()

    if ok {
        return state
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check after acquiring write lock
    if state, ok := c.metrics[tool]; ok {
        return state
    }

    state = &toolMetricsState{}
    c.metrics[tool] = state
    return state
}

func (c *InMemoryMetricsCollector) RecordRequest(tool string) {
    state := c.getOrCreate(tool)
    state.totalRequests.Add(1)
}

func (c *InMemoryMetricsCollector) RecordSuccess(tool string, duration time.Duration) {
    state := c.getOrCreate(tool)
    state.successCount.Add(1)
    state.totalDuration.Add(int64(duration))
    state.lastDuration.Store(int64(duration))
}

func (c *InMemoryMetricsCollector) RecordError(tool string, duration time.Duration) {
    state := c.getOrCreate(tool)
    state.errorCount.Add(1)
    state.totalDuration.Add(int64(duration))
    state.lastDuration.Store(int64(duration))
}

func (c *InMemoryMetricsCollector) RecordActive(tool string, delta int64) {
    state := c.getOrCreate(tool)
    state.activeRequests.Add(delta)
}

func (c *InMemoryMetricsCollector) GetMetrics(tool string) ToolMetrics {
    state := c.getOrCreate(tool)
    return ToolMetrics{
        TotalRequests:  state.totalRequests.Load(),
        SuccessCount:   state.successCount.Load(),
        ErrorCount:     state.errorCount.Load(),
        ActiveRequests: state.activeRequests.Load(),
        TotalDuration:  time.Duration(state.totalDuration.Load()),
        LastDuration:   time.Duration(state.lastDuration.Load()),
    }
}

// MetricsConfig configures the metrics middleware.
type MetricsConfig struct {
    Collector MetricsCollector
    Labels    map[string]string
}

// metricsProvider wraps a provider with metrics collection.
type metricsProvider struct {
    provider.ToolProvider
    collector MetricsCollector
}

// NewMetricsMiddleware creates a middleware that collects metrics.
func NewMetricsMiddleware(cfg MetricsConfig) Middleware {
    collector := cfg.Collector
    if collector == nil {
        collector = NewInMemoryMetricsCollector()
    }

    return func(next provider.ToolProvider) provider.ToolProvider {
        return &metricsProvider{
            ToolProvider: next,
            collector:    collector,
        }
    }
}

// Handle collects metrics around the wrapped provider call.
func (p *metricsProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    toolName := p.ToolProvider.Name()

    p.collector.RecordRequest(toolName)
    p.collector.RecordActive(toolName, 1)

    start := time.Now()

    result, err := p.ToolProvider.Handle(ctx, args)

    duration := time.Since(start)
    p.collector.RecordActive(toolName, -1)

    if err != nil {
        p.collector.RecordError(toolName, duration)
    } else {
        p.collector.RecordSuccess(toolName, duration)
    }

    return result, err
}

// Name returns the wrapped provider's name.
func (p *metricsProvider) Name() string {
    return p.ToolProvider.Name()
}

// Enabled returns the wrapped provider's enabled status.
func (p *metricsProvider) Enabled() bool {
    return p.ToolProvider.Enabled()
}

// Tool returns the wrapped provider's tool definition.
func (p *metricsProvider) Tool() mcp.Tool {
    return p.ToolProvider.Tool()
}

// MetricsMiddlewareFactory creates a metrics middleware from config.
func MetricsMiddlewareFactory(cfg map[string]any) (Middleware, error) {
    config := MetricsConfig{
        Collector: NewInMemoryMetricsCollector(),
    }

    if labels, ok := cfg["labels"].(map[string]any); ok {
        config.Labels = make(map[string]string)
        for k, v := range labels {
            if s, ok := v.(string); ok {
                config.Labels[k] = s
            }
        }
    }

    return NewMetricsMiddleware(config), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestMetricsMiddleware -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/metrics.go internal/middleware/metrics_test.go
git commit -m "$(cat <<'EOF'
feat(middleware): implement MetricsMiddleware

- MetricsCollector interface for pluggable backends
- InMemoryMetricsCollector for simple deployments
- Tracks requests, success/error counts, duration
- Active requests tracking for concurrency monitoring

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Add Configuration Support and Chain Builder

**Files:**
- Create: `internal/middleware/config.go`
- Test: `internal/middleware/config_test.go`

**Step 1: Write failing test for config-driven chain building**

```go
// internal/middleware/config_test.go
package middleware

import (
    "testing"
)

func TestMiddlewareConfig_Parse(t *testing.T) {
    yaml := `
middleware:
  chain:
    - logging
    - metrics
  logging:
    enabled: true
    level: info
  metrics:
    enabled: true
`
    cfg, err := ParseMiddlewareConfig([]byte(yaml))
    if err != nil {
        t.Fatalf("ParseMiddlewareConfig() error = %v", err)
    }

    if len(cfg.Chain) != 2 {
        t.Errorf("Chain length = %d, want 2", len(cfg.Chain))
    }

    if cfg.Chain[0] != "logging" || cfg.Chain[1] != "metrics" {
        t.Errorf("Chain = %v, want [logging, metrics]", cfg.Chain)
    }
}

func TestBuildChainFromConfig(t *testing.T) {
    registry := NewRegistry()
    registry.Register("logging", LoggingMiddlewareFactory)
    registry.Register("metrics", MetricsMiddlewareFactory)

    cfg := &MiddlewareConfig{
        Chain: []string{"logging", "metrics"},
        Configs: map[string]MiddlewareEntry{
            "logging": {Enabled: true, Config: map[string]any{"level": "info"}},
            "metrics": {Enabled: true, Config: nil},
        },
    }

    chain, err := BuildChainFromConfig(registry, cfg)
    if err != nil {
        t.Fatalf("BuildChainFromConfig() error = %v", err)
    }

    if chain.Len() != 2 {
        t.Errorf("Chain length = %d, want 2", chain.Len())
    }
}

func TestBuildChainFromConfig_DisabledMiddleware(t *testing.T) {
    registry := NewRegistry()
    registry.Register("logging", LoggingMiddlewareFactory)
    registry.Register("metrics", MetricsMiddlewareFactory)

    cfg := &MiddlewareConfig{
        Chain: []string{"logging", "metrics"},
        Configs: map[string]MiddlewareEntry{
            "logging": {Enabled: true},
            "metrics": {Enabled: false}, // Disabled
        },
    }

    chain, err := BuildChainFromConfig(registry, cfg)
    if err != nil {
        t.Fatalf("BuildChainFromConfig() error = %v", err)
    }

    if chain.Len() != 1 {
        t.Errorf("Chain length = %d, want 1 (metrics disabled)", chain.Len())
    }
}

func TestBuildChainFromConfig_UnknownMiddleware(t *testing.T) {
    registry := NewRegistry()
    // Don't register anything

    cfg := &MiddlewareConfig{
        Chain: []string{"unknown"},
        Configs: map[string]MiddlewareEntry{
            "unknown": {Enabled: true},
        },
    }

    _, err := BuildChainFromConfig(registry, cfg)
    if err == nil {
        t.Error("BuildChainFromConfig() should fail for unknown middleware")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -run TestMiddlewareConfig -v && go test ./internal/middleware/... -run TestBuildChain -v`
Expected: FAIL - Config types don't exist

**Step 3: Implement config types and chain builder**

```go
// internal/middleware/config.go
package middleware

import (
    "fmt"

    "gopkg.in/yaml.v3"
)

// MiddlewareConfig is the top-level middleware configuration.
type MiddlewareConfig struct {
    // Chain defines the order of middleware to apply
    Chain []string `yaml:"chain" koanf:"chain"`

    // Configs holds per-middleware configuration
    Configs map[string]MiddlewareEntry `yaml:",inline" koanf:",remain"`
}

// MiddlewareEntry configures a single middleware.
type MiddlewareEntry struct {
    Enabled bool           `yaml:"enabled" koanf:"enabled"`
    Config  map[string]any `yaml:"config,omitempty" koanf:"config"`
}

// ParseMiddlewareConfig parses YAML configuration.
func ParseMiddlewareConfig(data []byte) (*MiddlewareConfig, error) {
    // First parse the root structure
    var root struct {
        Middleware struct {
            Chain   []string       `yaml:"chain"`
            Configs map[string]any `yaml:",inline"`
        } `yaml:"middleware"`
    }

    if err := yaml.Unmarshal(data, &root); err != nil {
        return nil, fmt.Errorf("parse middleware config: %w", err)
    }

    cfg := &MiddlewareConfig{
        Chain:   root.Middleware.Chain,
        Configs: make(map[string]MiddlewareEntry),
    }

    // Parse each middleware config
    for name, raw := range root.Middleware.Configs {
        if name == "chain" {
            continue // Skip the chain key
        }

        entry := MiddlewareEntry{}

        switch v := raw.(type) {
        case map[string]any:
            if enabled, ok := v["enabled"].(bool); ok {
                entry.Enabled = enabled
            }
            delete(v, "enabled")
            if len(v) > 0 {
                entry.Config = v
            }
        case bool:
            entry.Enabled = v
        }

        cfg.Configs[name] = entry
    }

    return cfg, nil
}

// BuildChainFromConfig creates a middleware chain from configuration.
func BuildChainFromConfig(registry *Registry, cfg *MiddlewareConfig) (*Chain, error) {
    chain := NewChain()

    for _, name := range cfg.Chain {
        entry, ok := cfg.Configs[name]
        if !ok {
            // If not in configs, assume enabled with no config
            entry = MiddlewareEntry{Enabled: true}
        }

        if !entry.Enabled {
            continue
        }

        mw, err := registry.Create(name, entry.Config)
        if err != nil {
            return nil, fmt.Errorf("middleware %s: %w", name, err)
        }

        chain.Use(mw)
    }

    return chain, nil
}

// DefaultRegistry returns a registry with built-in middleware.
func DefaultRegistry() *Registry {
    registry := NewRegistry()
    registry.Register("logging", LoggingMiddlewareFactory)
    registry.Register("metrics", MetricsMiddlewareFactory)
    return registry
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/middleware/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/middleware/config.go internal/middleware/config_test.go
git commit -m "$(cat <<'EOF'
feat(middleware): add configuration support and chain builder

- MiddlewareConfig for YAML configuration
- ParseMiddlewareConfig for parsing
- BuildChainFromConfig for config-driven chain creation
- DefaultRegistry with built-in middleware

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Integrate with Server

**Files:**
- Create: `internal/server/middleware_adapter.go`
- Test: `internal/server/middleware_adapter_test.go`

**Step 1: Write integration test**

```go
// internal/server/middleware_adapter_test.go
package server

import (
    "context"
    "testing"

    "github.com/your-org/metatools-mcp/internal/middleware"
    "github.com/your-org/metatools-mcp/internal/provider"
)

func TestMiddlewareAdapter_ApplyToProviders(t *testing.T) {
    // Create provider registry
    providerRegistry := provider.NewRegistry()
    providerRegistry.MustRegister(&mockProvider{name: "tool1", enabled: true})
    providerRegistry.MustRegister(&mockProvider{name: "tool2", enabled: true})

    // Create middleware chain
    callCount := 0
    countingMiddleware := func(next provider.ToolProvider) provider.ToolProvider {
        return &wrappedProvider{
            ToolProvider: next,
            beforeFn: func() {
                callCount++
            },
        }
    }

    chain := middleware.NewChain(countingMiddleware)

    adapter := NewMiddlewareAdapter(chain)
    adapter.ApplyToProviders(providerRegistry)

    // Execute both tools
    p1, _ := providerRegistry.Get("tool1")
    p1.Handle(context.Background(), nil)

    p2, _ := providerRegistry.Get("tool2")
    p2.Handle(context.Background(), nil)

    if callCount != 2 {
        t.Errorf("Middleware called %d times, want 2", callCount)
    }
}

func TestMiddlewareAdapter_FromConfig(t *testing.T) {
    cfg := &middleware.MiddlewareConfig{
        Chain: []string{"logging"},
        Configs: map[string]middleware.MiddlewareEntry{
            "logging": {Enabled: true},
        },
    }

    adapter, err := NewMiddlewareAdapterFromConfig(cfg)
    if err != nil {
        t.Fatalf("NewMiddlewareAdapterFromConfig() error = %v", err)
    }

    if adapter.chain.Len() != 1 {
        t.Errorf("Chain length = %d, want 1", adapter.chain.Len())
    }
}

// mockProvider and wrappedProvider for testing
type mockProvider struct {
    name    string
    enabled bool
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Enabled() bool { return m.enabled }
func (m *mockProvider) Tool() mcp.Tool { return mcp.Tool{Name: m.name} }
func (m *mockProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    return nil, nil
}

type wrappedProvider struct {
    provider.ToolProvider
    beforeFn func()
}

func (w *wrappedProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    if w.beforeFn != nil {
        w.beforeFn()
    }
    return w.ToolProvider.Handle(ctx, args)
}
```

**Step 2: Implement adapter**

```go
// internal/server/middleware_adapter.go
package server

import (
    "github.com/your-org/metatools-mcp/internal/middleware"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// MiddlewareAdapter applies middleware to provider registries.
type MiddlewareAdapter struct {
    chain    *middleware.Chain
    registry *middleware.Registry
}

// NewMiddlewareAdapter creates a new adapter with an existing chain.
func NewMiddlewareAdapter(chain *middleware.Chain) *MiddlewareAdapter {
    return &MiddlewareAdapter{
        chain:    chain,
        registry: middleware.DefaultRegistry(),
    }
}

// NewMiddlewareAdapterFromConfig creates an adapter from configuration.
func NewMiddlewareAdapterFromConfig(cfg *middleware.MiddlewareConfig) (*MiddlewareAdapter, error) {
    registry := middleware.DefaultRegistry()

    chain, err := middleware.BuildChainFromConfig(registry, cfg)
    if err != nil {
        return nil, err
    }

    return &MiddlewareAdapter{
        chain:    chain,
        registry: registry,
    }, nil
}

// ApplyToProviders wraps all providers in a registry with middleware.
func (a *MiddlewareAdapter) ApplyToProviders(providerRegistry *provider.Registry) {
    providers := providerRegistry.List()

    for _, p := range providers {
        wrapped := a.chain.Apply(p)
        providerRegistry.Unregister(p.Name())
        providerRegistry.Register(wrapped)
    }
}

// Chain returns the middleware chain.
func (a *MiddlewareAdapter) Chain() *middleware.Chain {
    return a.chain
}

// Registry returns the middleware registry.
func (a *MiddlewareAdapter) Registry() *middleware.Registry {
    return a.registry
}
```

**Step 3: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/server/... -run TestMiddlewareAdapter -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/server/middleware_adapter.go internal/server/middleware_adapter_test.go
git commit -m "$(cat <<'EOF'
feat(server): add MiddlewareAdapter for server integration

- ApplyToProviders wraps all providers with middleware
- NewMiddlewareAdapterFromConfig for config-driven setup
- Access to chain and registry for customization

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Middleware type defined as function wrapper
- [ ] Chain for ordered middleware application
- [ ] Registry for middleware factories
- [ ] LoggingMiddleware with slog integration
- [ ] MetricsMiddleware with collector interface
- [ ] InMemoryMetricsCollector implementation
- [ ] Config parsing for YAML
- [ ] BuildChainFromConfig for config-driven chains
- [ ] MiddlewareAdapter for server integration
- [ ] All tests pass

## Definition of Done

1. All tests pass: `go test ./internal/middleware/...`
2. Middleware can wrap providers with logging
3. Metrics collection tracks requests, errors, duration
4. Configuration drives chain construction
5. Server can apply middleware to all providers

## Next PRD

PRD-008 will implement the tooladapter Library for protocol-agnostic tool definitions.
