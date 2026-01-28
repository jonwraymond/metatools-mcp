# PRD-005: Tool Provider Registry

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a Tool Provider Registry enabling plug-and-play tool registration, replacing the hard-coded `registerTools()` function with a dynamic, configuration-driven system.

**Architecture:** Define a ToolProvider interface that wraps tool definition and execution. Create a ProviderRegistry that manages provider lifecycle and discovery. Refactor existing handlers (search_tools, describe_tool, run_tool, etc.) into providers implementing the interface. Enable configuration-driven provider loading.

**Tech Stack:** Go interfaces, existing handler implementations, config integration from PRD-003

**Priority:** P0 - Stream A, Phase 3 (enables extensibility)

**Scope:** Provider interface + registry + refactored built-in providers

**Dependencies:** PRD-002 (CLI), PRD-003 (Config)

---

## Context

The current server has a ~200-line `registerTools()` function that hard-codes tool schemas. The pluggable architecture requires:
1. Dynamic tool registration
2. Configuration-driven enable/disable
3. Custom provider support
4. Runtime provider discovery

**Current State:**
```go
// server/server.go (current)
func (s *Server) registerTools() {
    // 200+ lines of hard-coded tool schemas
    s.AddTool(searchToolSchema)
    s.AddTool(describeToolSchema)
    // ...
}
```

**Target State:**
```go
// Provider interface
type ToolProvider interface {
    Name() string
    Tool() mcp.Tool
    Handle(ctx context.Context, args map[string]any) (any, error)
}

// Registry-driven registration
registry.Register("search_tools", &SearchToolsProvider{...})
registry.Register("describe_tool", &DescribeToolProvider{...})
```

---

## Tasks

### Task 1: Define ToolProvider Interface

**Files:**
- Create: `internal/provider/provider.go`
- Test: `internal/provider/provider_test.go`

**Step 1: Write failing test for ToolProvider interface**

```go
// internal/provider/provider_test.go
package provider

import (
    "context"
    "testing"

    "github.com/mark3labs/mcp-go/mcp"
)

// mockProvider implements ToolProvider for testing
type mockProvider struct {
    name     string
    enabled  bool
    tool     mcp.Tool
    handleFn func(ctx context.Context, args map[string]any) (any, error)
}

func (m *mockProvider) Name() string       { return m.name }
func (m *mockProvider) Enabled() bool      { return m.enabled }
func (m *mockProvider) Tool() mcp.Tool     { return m.tool }
func (m *mockProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    if m.handleFn != nil {
        return m.handleFn(ctx, args)
    }
    return nil, nil
}

func TestToolProvider_Interface(t *testing.T) {
    // Verify interface is implemented correctly
    var _ ToolProvider = (*mockProvider)(nil)
}

func TestToolProvider_Methods(t *testing.T) {
    provider := &mockProvider{
        name:    "test_tool",
        enabled: true,
        tool: mcp.Tool{
            Name:        "test_tool",
            Description: "A test tool",
        },
        handleFn: func(ctx context.Context, args map[string]any) (any, error) {
            return "result", nil
        },
    }

    if provider.Name() != "test_tool" {
        t.Errorf("Name() = %q, want %q", provider.Name(), "test_tool")
    }

    if !provider.Enabled() {
        t.Error("Enabled() = false, want true")
    }

    tool := provider.Tool()
    if tool.Name != "test_tool" {
        t.Errorf("Tool().Name = %q, want %q", tool.Name, "test_tool")
    }

    result, err := provider.Handle(context.Background(), nil)
    if err != nil {
        t.Errorf("Handle() error = %v", err)
    }
    if result != "result" {
        t.Errorf("Handle() = %v, want %v", result, "result")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -v`
Expected: FAIL - ToolProvider type doesn't exist

**Step 3: Implement ToolProvider interface**

```go
// internal/provider/provider.go
package provider

import (
    "context"

    "github.com/mark3labs/mcp-go/mcp"
)

// ToolProvider defines the interface for MCP tool providers.
// A provider encapsulates a tool's definition and execution logic.
type ToolProvider interface {
    // Name returns the unique identifier for this provider.
    Name() string

    // Enabled returns whether this provider is currently enabled.
    Enabled() bool

    // Tool returns the MCP tool definition.
    Tool() mcp.Tool

    // Handle processes a tool invocation.
    Handle(ctx context.Context, args map[string]any) (any, error)
}

// ConfigurableProvider is a provider that can be configured at runtime.
type ConfigurableProvider interface {
    ToolProvider

    // Configure applies configuration to the provider.
    Configure(cfg map[string]any) error
}

// StreamingProvider is a provider that supports streaming responses.
type StreamingProvider interface {
    ToolProvider

    // HandleStream processes a streaming tool invocation.
    // Returns a channel that emits response parts.
    HandleStream(ctx context.Context, args map[string]any) (<-chan any, error)
}

// ProviderFactory creates provider instances.
type ProviderFactory func() ToolProvider

// ProviderInfo contains metadata about a provider.
type ProviderInfo struct {
    Name        string
    Description string
    Version     string
    Author      string
    Streaming   bool
}

// GetInfo returns provider metadata if available.
func GetInfo(p ToolProvider) ProviderInfo {
    info := ProviderInfo{
        Name: p.Name(),
    }

    tool := p.Tool()
    info.Description = tool.Description

    // Check if streaming
    _, info.Streaming = p.(StreamingProvider)

    return info
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/provider.go internal/provider/provider_test.go
git commit -m "$(cat <<'EOF'
feat(provider): define ToolProvider interface

- ToolProvider with Name, Enabled, Tool, Handle methods
- ConfigurableProvider for runtime configuration
- StreamingProvider for streaming responses
- ProviderFactory and ProviderInfo types

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Implement Provider Registry

**Files:**
- Create: `internal/provider/registry.go`
- Test: `internal/provider/registry_test.go`

**Step 1: Write failing test for registry**

```go
// internal/provider/registry_test.go
package provider

import (
    "testing"
)

func TestRegistry_Register(t *testing.T) {
    registry := NewRegistry()

    provider := &mockProvider{
        name:    "test_tool",
        enabled: true,
    }

    err := registry.Register(provider)
    if err != nil {
        t.Fatalf("Register() error = %v", err)
    }

    // Duplicate registration should fail
    err = registry.Register(provider)
    if err == nil {
        t.Error("Register() should fail on duplicate")
    }
}

func TestRegistry_Get(t *testing.T) {
    registry := NewRegistry()

    provider := &mockProvider{
        name:    "test_tool",
        enabled: true,
    }
    registry.Register(provider)

    got, ok := registry.Get("test_tool")
    if !ok {
        t.Fatal("Get() returned false")
    }
    if got.Name() != "test_tool" {
        t.Errorf("Get().Name() = %q, want %q", got.Name(), "test_tool")
    }

    _, ok = registry.Get("nonexistent")
    if ok {
        t.Error("Get() should return false for nonexistent provider")
    }
}

func TestRegistry_List(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockProvider{name: "tool_a", enabled: true})
    registry.Register(&mockProvider{name: "tool_b", enabled: true})
    registry.Register(&mockProvider{name: "tool_c", enabled: false})

    // List all
    all := registry.List()
    if len(all) != 3 {
        t.Errorf("List() returned %d providers, want 3", len(all))
    }

    // List enabled only
    enabled := registry.ListEnabled()
    if len(enabled) != 2 {
        t.Errorf("ListEnabled() returned %d providers, want 2", len(enabled))
    }
}

func TestRegistry_Tools(t *testing.T) {
    registry := NewRegistry()

    registry.Register(&mockProvider{
        name:    "tool_a",
        enabled: true,
        tool:    mcp.Tool{Name: "tool_a", Description: "Tool A"},
    })
    registry.Register(&mockProvider{
        name:    "tool_b",
        enabled: true,
        tool:    mcp.Tool{Name: "tool_b", Description: "Tool B"},
    })
    registry.Register(&mockProvider{
        name:    "tool_c",
        enabled: false,
        tool:    mcp.Tool{Name: "tool_c", Description: "Tool C"},
    })

    tools := registry.Tools()
    if len(tools) != 2 { // Only enabled providers
        t.Errorf("Tools() returned %d tools, want 2", len(tools))
    }
}

func TestRegistry_Unregister(t *testing.T) {
    registry := NewRegistry()

    provider := &mockProvider{name: "test_tool", enabled: true}
    registry.Register(provider)

    registry.Unregister("test_tool")

    _, ok := registry.Get("test_tool")
    if ok {
        t.Error("Get() should return false after Unregister()")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -run TestRegistry -v`
Expected: FAIL - Registry doesn't exist

**Step 3: Implement Registry**

```go
// internal/provider/registry.go
package provider

import (
    "errors"
    "fmt"
    "sort"
    "sync"

    "github.com/mark3labs/mcp-go/mcp"
)

// ErrProviderExists is returned when registering a duplicate provider.
var ErrProviderExists = errors.New("provider already registered")

// ErrProviderNotFound is returned when a provider is not found.
var ErrProviderNotFound = errors.New("provider not found")

// Registry manages tool providers.
type Registry struct {
    providers map[string]ToolProvider
    mu        sync.RWMutex
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]ToolProvider),
    }
}

// Register adds a provider to the registry.
func (r *Registry) Register(p ToolProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := p.Name()
    if _, exists := r.providers[name]; exists {
        return fmt.Errorf("%w: %s", ErrProviderExists, name)
    }

    r.providers[name] = p
    return nil
}

// MustRegister adds a provider or panics.
func (r *Registry) MustRegister(p ToolProvider) {
    if err := r.Register(p); err != nil {
        panic(err)
    }
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(name string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.providers, name)
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) (ToolProvider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.providers[name]
    return p, ok
}

// List returns all registered providers.
func (r *Registry) List() []ToolProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    providers := make([]ToolProvider, 0, len(r.providers))
    for _, p := range r.providers {
        providers = append(providers, p)
    }

    // Sort by name for consistency
    sort.Slice(providers, func(i, j int) bool {
        return providers[i].Name() < providers[j].Name()
    })

    return providers
}

// ListEnabled returns only enabled providers.
func (r *Registry) ListEnabled() []ToolProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    providers := make([]ToolProvider, 0, len(r.providers))
    for _, p := range r.providers {
        if p.Enabled() {
            providers = append(providers, p)
        }
    }

    sort.Slice(providers, func(i, j int) bool {
        return providers[i].Name() < providers[j].Name()
    })

    return providers
}

// Tools returns MCP tool definitions for all enabled providers.
func (r *Registry) Tools() []mcp.Tool {
    enabled := r.ListEnabled()
    tools := make([]mcp.Tool, 0, len(enabled))
    for _, p := range enabled {
        tools = append(tools, p.Tool())
    }
    return tools
}

// Handle invokes the provider for the given tool name.
func (r *Registry) Handle(ctx context.Context, name string, args map[string]any) (any, error) {
    p, ok := r.Get(name)
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
    }

    if !p.Enabled() {
        return nil, fmt.Errorf("provider disabled: %s", name)
    }

    return p.Handle(ctx, args)
}

// Count returns the number of registered providers.
func (r *Registry) Count() int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.providers)
}

// Clear removes all providers.
func (r *Registry) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers = make(map[string]ToolProvider)
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -run TestRegistry -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/registry.go internal/provider/registry_test.go
git commit -m "$(cat <<'EOF'
feat(provider): implement Provider Registry

- Register/Unregister providers by name
- Get/List/ListEnabled for provider lookup
- Tools() returns MCP tool definitions
- Handle() invokes provider by name
- Thread-safe with RWMutex

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Create Base Provider Implementation

**Files:**
- Create: `internal/provider/base.go`
- Test: `internal/provider/base_test.go`

**Step 1: Write failing test for BaseProvider**

```go
// internal/provider/base_test.go
package provider

import (
    "context"
    "testing"

    "github.com/mark3labs/mcp-go/mcp"
)

func TestBaseProvider(t *testing.T) {
    handler := func(ctx context.Context, args map[string]any) (any, error) {
        return args["input"], nil
    }

    provider := NewBaseProvider(BaseProviderConfig{
        Name:        "test_tool",
        Description: "A test tool",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "input": map[string]any{"type": "string"},
            },
        },
        Handler: handler,
    })

    // Test Name
    if provider.Name() != "test_tool" {
        t.Errorf("Name() = %q, want %q", provider.Name(), "test_tool")
    }

    // Test Enabled (default true)
    if !provider.Enabled() {
        t.Error("Enabled() = false, want true")
    }

    // Test Tool
    tool := provider.Tool()
    if tool.Name != "test_tool" {
        t.Errorf("Tool().Name = %q, want %q", tool.Name, "test_tool")
    }
    if tool.Description != "A test tool" {
        t.Errorf("Tool().Description = %q, want %q", tool.Description, "A test tool")
    }

    // Test Handle
    result, err := provider.Handle(context.Background(), map[string]any{"input": "hello"})
    if err != nil {
        t.Errorf("Handle() error = %v", err)
    }
    if result != "hello" {
        t.Errorf("Handle() = %v, want %v", result, "hello")
    }
}

func TestBaseProvider_Disabled(t *testing.T) {
    provider := NewBaseProvider(BaseProviderConfig{
        Name:    "disabled_tool",
        Enabled: false,
        Handler: func(ctx context.Context, args map[string]any) (any, error) {
            return nil, nil
        },
    })

    if provider.Enabled() {
        t.Error("Enabled() = true, want false")
    }
}

func TestBaseProvider_Configure(t *testing.T) {
    provider := NewBaseProvider(BaseProviderConfig{
        Name: "configurable_tool",
        Handler: func(ctx context.Context, args map[string]any) (any, error) {
            return nil, nil
        },
    })

    // Configure should enable/disable
    err := provider.Configure(map[string]any{"enabled": false})
    if err != nil {
        t.Errorf("Configure() error = %v", err)
    }

    if provider.Enabled() {
        t.Error("Enabled() = true after Configure(enabled: false)")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -run TestBaseProvider -v`
Expected: FAIL - BaseProvider doesn't exist

**Step 3: Implement BaseProvider**

```go
// internal/provider/base.go
package provider

import (
    "context"
    "sync"

    "github.com/mark3labs/mcp-go/mcp"
)

// HandlerFunc is the function signature for tool handlers.
type HandlerFunc func(ctx context.Context, args map[string]any) (any, error)

// BaseProviderConfig configures a base provider.
type BaseProviderConfig struct {
    Name        string
    Description string
    InputSchema map[string]any
    Enabled     bool
    Handler     HandlerFunc
}

// BaseProvider provides a simple implementation of ToolProvider.
type BaseProvider struct {
    name        string
    description string
    inputSchema map[string]any
    enabled     bool
    handler     HandlerFunc

    mu sync.RWMutex
}

// NewBaseProvider creates a new base provider.
func NewBaseProvider(cfg BaseProviderConfig) *BaseProvider {
    enabled := true
    if cfg.Enabled == false && cfg.Handler != nil {
        // Only set enabled=false if explicitly configured
        // This handles the zero-value case
    } else if !cfg.Enabled {
        enabled = cfg.Enabled
    }

    // Actually, simpler logic: if Handler is set, default to enabled
    enabled = cfg.Handler != nil
    if cfg.InputSchema != nil || cfg.Description != "" {
        enabled = true // Has tool definition, so enabled
    }

    return &BaseProvider{
        name:        cfg.Name,
        description: cfg.Description,
        inputSchema: cfg.InputSchema,
        enabled:     enabled,
        handler:     cfg.Handler,
    }
}

// Name returns the provider name.
func (p *BaseProvider) Name() string {
    return p.name
}

// Enabled returns whether the provider is enabled.
func (p *BaseProvider) Enabled() bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.enabled
}

// SetEnabled enables or disables the provider.
func (p *BaseProvider) SetEnabled(enabled bool) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.enabled = enabled
}

// Tool returns the MCP tool definition.
func (p *BaseProvider) Tool() mcp.Tool {
    return mcp.Tool{
        Name:        p.name,
        Description: p.description,
        InputSchema: mcp.ToolInputSchema{
            Type:       "object",
            Properties: p.inputSchema,
        },
    }
}

// Handle invokes the tool handler.
func (p *BaseProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    if p.handler == nil {
        return nil, ErrProviderNotFound
    }
    return p.handler(ctx, args)
}

// Configure applies configuration to the provider.
func (p *BaseProvider) Configure(cfg map[string]any) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    if enabled, ok := cfg["enabled"].(bool); ok {
        p.enabled = enabled
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/... -run TestBaseProvider -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/base.go internal/provider/base_test.go
git commit -m "$(cat <<'EOF'
feat(provider): implement BaseProvider for simple tools

- NewBaseProvider with config struct
- Name, Enabled, Tool, Handle methods
- Configure for runtime enable/disable
- Thread-safe enabled flag

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Refactor search_tools Provider

**Files:**
- Create: `internal/provider/builtin/search_tools.go`
- Test: `internal/provider/builtin/search_tools_test.go`

**Step 1: Write failing test for SearchToolsProvider**

```go
// internal/provider/builtin/search_tools_test.go
package builtin

import (
    "context"
    "testing"

    "github.com/your-org/metatools-mcp/internal/provider"
)

func TestSearchToolsProvider_Interface(t *testing.T) {
    // Verify it implements ToolProvider
    var _ provider.ToolProvider = (*SearchToolsProvider)(nil)
}

func TestSearchToolsProvider_Name(t *testing.T) {
    p := NewSearchToolsProvider(nil) // nil deps for now
    if p.Name() != "search_tools" {
        t.Errorf("Name() = %q, want %q", p.Name(), "search_tools")
    }
}

func TestSearchToolsProvider_Tool(t *testing.T) {
    p := NewSearchToolsProvider(nil)
    tool := p.Tool()

    if tool.Name != "search_tools" {
        t.Errorf("Tool().Name = %q, want %q", tool.Name, "search_tools")
    }

    // Check schema has required properties
    schema := tool.InputSchema
    if schema.Type != "object" {
        t.Errorf("InputSchema.Type = %q, want %q", schema.Type, "object")
    }
}

func TestSearchToolsProvider_Handle(t *testing.T) {
    // Create mock searcher
    mockSearcher := &mockSearcher{
        results: []SearchResult{
            {ID: "tool1", Score: 0.9},
            {ID: "tool2", Score: 0.8},
        },
    }

    p := NewSearchToolsProvider(SearchToolsDeps{
        Searcher: mockSearcher,
    })

    result, err := p.Handle(context.Background(), map[string]any{
        "query": "test query",
        "limit": 10,
    })

    if err != nil {
        t.Fatalf("Handle() error = %v", err)
    }

    // Check result structure
    results, ok := result.([]SearchResult)
    if !ok {
        t.Fatalf("Handle() result type = %T, want []SearchResult", result)
    }

    if len(results) != 2 {
        t.Errorf("Handle() returned %d results, want 2", len(results))
    }
}

// Mock searcher for testing
type mockSearcher struct {
    results []SearchResult
}

func (m *mockSearcher) Search(query string, limit int) ([]SearchResult, error) {
    return m.results, nil
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/builtin/... -v`
Expected: FAIL - SearchToolsProvider doesn't exist

**Step 3: Implement SearchToolsProvider**

```go
// internal/provider/builtin/search_tools.go
package builtin

import (
    "context"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// SearchResult represents a search result.
type SearchResult struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Namespace   string  `json:"namespace"`
    Description string  `json:"description"`
    Score       float64 `json:"score"`
}

// Searcher interface for tool search.
type Searcher interface {
    Search(query string, limit int) ([]SearchResult, error)
}

// SearchToolsDeps holds dependencies for SearchToolsProvider.
type SearchToolsDeps struct {
    Searcher Searcher
}

// SearchToolsProvider implements the search_tools MCP tool.
type SearchToolsProvider struct {
    deps    SearchToolsDeps
    enabled bool
}

// NewSearchToolsProvider creates a new search_tools provider.
func NewSearchToolsProvider(deps SearchToolsDeps) *SearchToolsProvider {
    return &SearchToolsProvider{
        deps:    deps,
        enabled: true,
    }
}

// Name returns "search_tools".
func (p *SearchToolsProvider) Name() string {
    return "search_tools"
}

// Enabled returns whether the provider is enabled.
func (p *SearchToolsProvider) Enabled() bool {
    return p.enabled
}

// SetEnabled enables or disables the provider.
func (p *SearchToolsProvider) SetEnabled(enabled bool) {
    p.enabled = enabled
}

// Tool returns the MCP tool definition.
func (p *SearchToolsProvider) Tool() mcp.Tool {
    return mcp.Tool{
        Name:        "search_tools",
        Description: "Search for tools by query. Returns ranked results with tool IDs and descriptions.",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "query": map[string]any{
                    "type":        "string",
                    "description": "Search query to find relevant tools",
                },
                "limit": map[string]any{
                    "type":        "integer",
                    "description": "Maximum number of results to return",
                    "default":     10,
                    "minimum":     1,
                    "maximum":     100,
                },
            },
            Required: []string{"query"},
        },
    }
}

// Handle processes a search_tools request.
func (p *SearchToolsProvider) Handle(ctx context.Context, args map[string]any) (any, error) {
    // Extract arguments
    query, _ := args["query"].(string)
    limit := 10
    if l, ok := args["limit"].(float64); ok {
        limit = int(l)
    } else if l, ok := args["limit"].(int); ok {
        limit = l
    }

    // Perform search
    if p.deps.Searcher == nil {
        return []SearchResult{}, nil
    }

    results, err := p.deps.Searcher.Search(query, limit)
    if err != nil {
        return nil, err
    }

    return results, nil
}

// Configure applies configuration.
func (p *SearchToolsProvider) Configure(cfg map[string]any) error {
    if enabled, ok := cfg["enabled"].(bool); ok {
        p.enabled = enabled
    }
    return nil
}

// Ensure SearchToolsProvider implements ToolProvider and ConfigurableProvider
var _ provider.ToolProvider = (*SearchToolsProvider)(nil)
var _ provider.ConfigurableProvider = (*SearchToolsProvider)(nil)
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/builtin/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/provider/builtin/search_tools.go internal/provider/builtin/search_tools_test.go
git commit -m "$(cat <<'EOF'
feat(provider): add SearchToolsProvider

- Implements ToolProvider interface
- MCP tool definition with query and limit params
- Configurable enabled state
- Dependency injection for Searcher

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Create Default Registry with Built-in Providers

**Files:**
- Create: `internal/provider/builtin/registry.go`
- Test: `internal/provider/builtin/registry_test.go`

**Step 1: Write failing test**

```go
// internal/provider/builtin/registry_test.go
package builtin

import (
    "testing"

    "github.com/your-org/metatools-mcp/internal/provider"
)

func TestDefaultRegistry(t *testing.T) {
    deps := Dependencies{
        // Minimal deps for testing
    }

    registry := NewDefaultRegistry(deps)

    // Should have built-in providers
    if registry.Count() == 0 {
        t.Error("DefaultRegistry has no providers")
    }

    // search_tools should exist
    _, ok := registry.Get("search_tools")
    if !ok {
        t.Error("search_tools provider not found")
    }
}

func TestDefaultRegistry_Tools(t *testing.T) {
    deps := Dependencies{}
    registry := NewDefaultRegistry(deps)

    tools := registry.Tools()
    if len(tools) == 0 {
        t.Error("Tools() returned empty slice")
    }

    // Check tool names
    hasSearchTools := false
    for _, tool := range tools {
        if tool.Name == "search_tools" {
            hasSearchTools = true
        }
    }

    if !hasSearchTools {
        t.Error("Tools() missing search_tools")
    }
}
```

**Step 2: Implement default registry**

```go
// internal/provider/builtin/registry.go
package builtin

import (
    "github.com/your-org/metatools-mcp/internal/provider"
)

// Dependencies holds all dependencies for built-in providers.
type Dependencies struct {
    Searcher    Searcher
    // Future: Index, Docs, Runner, etc.
}

// NewDefaultRegistry creates a registry with all built-in providers.
func NewDefaultRegistry(deps Dependencies) *provider.Registry {
    registry := provider.NewRegistry()

    // Register built-in providers
    registry.MustRegister(NewSearchToolsProvider(SearchToolsDeps{
        Searcher: deps.Searcher,
    }))

    // Future providers:
    // registry.MustRegister(NewDescribeToolProvider(...))
    // registry.MustRegister(NewRunToolProvider(...))
    // registry.MustRegister(NewRunChainProvider(...))
    // registry.MustRegister(NewListNamespacesProvider(...))

    return registry
}

// BuiltinProviders returns a list of built-in provider names.
func BuiltinProviders() []string {
    return []string{
        "search_tools",
        "describe_tool",
        "run_tool",
        "run_chain",
        "list_namespaces",
        "list_tool_examples",
        "execute_code",
    }
}
```

**Step 3: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/provider/builtin/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/provider/builtin/registry.go internal/provider/builtin/registry_test.go
git commit -m "$(cat <<'EOF'
feat(provider): add DefaultRegistry with built-in providers

- NewDefaultRegistry creates registry with all built-ins
- Dependency injection for provider dependencies
- BuiltinProviders() lists all built-in provider names

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Integrate Registry with Server

**Files:**
- Modify: `internal/server/server.go` (or create adapter)
- Test: Integration test

**Step 1: Create server integration**

```go
// internal/server/provider_adapter.go
package server

import (
    "context"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/your-org/metatools-mcp/internal/provider"
)

// ProviderAdapter adapts the provider registry to the MCP server.
type ProviderAdapter struct {
    registry *provider.Registry
}

// NewProviderAdapter creates a new adapter.
func NewProviderAdapter(registry *provider.Registry) *ProviderAdapter {
    return &ProviderAdapter{registry: registry}
}

// RegisterTools registers all enabled providers as MCP tools.
func (a *ProviderAdapter) RegisterTools(server *Server) error {
    for _, p := range a.registry.ListEnabled() {
        tool := p.Tool()

        // Add tool to server
        server.AddTool(tool, a.createHandler(p))
    }
    return nil
}

// createHandler creates an MCP handler for a provider.
func (a *ProviderAdapter) createHandler(p provider.ToolProvider) mcp.ToolHandler {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args := request.Params.Arguments

        result, err := p.Handle(ctx, args)
        if err != nil {
            return nil, err
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{
                mcp.TextContent{
                    Type: "text",
                    Text: formatResult(result),
                },
            },
        }, nil
    }
}

// formatResult converts a result to string for MCP response.
func formatResult(result any) string {
    // JSON encode the result
    data, err := json.Marshal(result)
    if err != nil {
        return fmt.Sprintf("%v", result)
    }
    return string(data)
}
```

**Step 2: Write integration test**

```go
// internal/server/provider_adapter_test.go
package server

import (
    "testing"

    "github.com/your-org/metatools-mcp/internal/provider"
    "github.com/your-org/metatools-mcp/internal/provider/builtin"
)

func TestProviderAdapter_RegisterTools(t *testing.T) {
    // Create registry with test providers
    registry := provider.NewRegistry()
    registry.MustRegister(builtin.NewSearchToolsProvider(builtin.SearchToolsDeps{}))

    adapter := NewProviderAdapter(registry)

    // Create mock server
    server := &mockServer{tools: make(map[string]mcp.Tool)}

    err := adapter.RegisterTools(server)
    if err != nil {
        t.Fatalf("RegisterTools() error = %v", err)
    }

    if len(server.tools) == 0 {
        t.Error("No tools registered")
    }

    if _, ok := server.tools["search_tools"]; !ok {
        t.Error("search_tools not registered")
    }
}
```

**Step 3: Commit**

```bash
git add internal/server/provider_adapter.go internal/server/provider_adapter_test.go
git commit -m "$(cat <<'EOF'
feat(server): add ProviderAdapter for registry integration

- RegisterTools adds providers as MCP tools
- Creates handler wrappers for provider.Handle
- Formats results as MCP response content

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] ToolProvider interface defined
- [ ] ConfigurableProvider interface
- [ ] StreamingProvider interface
- [ ] Registry with Register/Get/List
- [ ] BaseProvider implementation
- [ ] SearchToolsProvider implemented
- [ ] DefaultRegistry with built-ins
- [ ] ProviderAdapter for server integration
- [ ] All tests pass

## Definition of Done

1. All tests pass: `go test ./internal/provider/...`
2. Built-in providers implement ToolProvider interface
3. Registry manages provider lifecycle
4. Server uses registry instead of hard-coded tools
5. Providers can be enabled/disabled via config

## Next PRD

PRD-006 will implement the Backend Registry for multi-source tool aggregation.
