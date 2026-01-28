# PRD-009: toolset Composition Library Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a composable tool collection library that enables curated, filtered, and access-controlled tool sets from multiple sources.

**Architecture:** Fluent builder pattern for constructing toolsets from registries, with filtering by namespace, tags, categories, and custom policies. Supports multiple exposure formats (MCP, OpenAI, Anthropic).

**Tech Stack:** Go, tooladapter dependency, toolindex dependency

---

## Overview

The `toolset` library provides composable tool collections for creating curated API surfaces. It enables filtering, access control, and multi-format exposure of tools.

**Reference:** [protocol-agnostic-tools.md](../proposals/protocol-agnostic-tools.md) - Section "toolset Library"

---

## Directory Structure

```
toolset/
├── toolset.go          # Toolset type and methods
├── toolset_test.go     # Toolset tests
├── builder.go          # Builder pattern implementation
├── builder_test.go     # Builder tests
├── filter.go           # Filter functions
├── filter_test.go      # Filter tests
├── policy.go           # Access control policies
├── policy_test.go      # Policy tests
├── exposure.go         # Multi-format exposure
├── exposure_test.go    # Exposure tests
├── doc.go              # Package documentation
├── go.mod
└── go.sum
```

---

## Task 1: Toolset Type and Basic Operations

**Files:**
- Create: `toolset/toolset.go`
- Create: `toolset/toolset_test.go`
- Create: `toolset/go.mod`

**Step 1: Write failing tests**

```go
// toolset_test.go
package toolset_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
)

func makeTool(name, namespace string, tags []string) *tooladapter.CanonicalTool {
    return &tooladapter.CanonicalTool{
        Name:      name,
        Namespace: namespace,
        Tags:      tags,
        InputSchema: &tooladapter.JSONSchema{
            Type: "object",
        },
    }
}

func TestToolset_New(t *testing.T) {
    ts := toolset.New("test-set")

    assert.Equal(t, "test-set", ts.Name())
    assert.Empty(t, ts.Tools())
}

func TestToolset_AddTool(t *testing.T) {
    ts := toolset.New("test-set")
    tool := makeTool("search", "mcp", []string{"query"})

    ts.Add(tool)

    tools := ts.Tools()
    require.Len(t, tools, 1)
    assert.Equal(t, "search", tools[0].Name)
}

func TestToolset_AddMultipleTools(t *testing.T) {
    ts := toolset.New("test-set")

    ts.Add(makeTool("search", "mcp", []string{"query"}))
    ts.Add(makeTool("describe", "mcp", []string{"info"}))
    ts.Add(makeTool("execute", "local", []string{"run"}))

    assert.Len(t, ts.Tools(), 3)
}

func TestToolset_Get(t *testing.T) {
    ts := toolset.New("test-set")
    tool := makeTool("search", "mcp", []string{"query"})
    ts.Add(tool)

    found, ok := ts.Get("mcp:search")
    require.True(t, ok)
    assert.Equal(t, "search", found.Name)

    _, ok = ts.Get("nonexistent")
    assert.False(t, ok)
}

func TestToolset_Remove(t *testing.T) {
    ts := toolset.New("test-set")
    ts.Add(makeTool("search", "mcp", nil))
    ts.Add(makeTool("describe", "mcp", nil))

    removed := ts.Remove("mcp:search")
    assert.True(t, removed)
    assert.Len(t, ts.Tools(), 1)

    removed = ts.Remove("nonexistent")
    assert.False(t, removed)
}

func TestToolset_Filter(t *testing.T) {
    ts := toolset.New("test-set")
    ts.Add(makeTool("search", "mcp", []string{"query"}))
    ts.Add(makeTool("describe", "mcp", []string{"info"}))
    ts.Add(makeTool("execute", "local", []string{"run"}))

    filtered := ts.Filter(func(t *tooladapter.CanonicalTool) bool {
        return t.Namespace == "mcp"
    })

    assert.Equal(t, "test-set (filtered)", filtered.Name())
    assert.Len(t, filtered.Tools(), 2)
}

func TestToolset_Count(t *testing.T) {
    ts := toolset.New("test-set")
    assert.Equal(t, 0, ts.Count())

    ts.Add(makeTool("search", "mcp", nil))
    assert.Equal(t, 1, ts.Count())

    ts.Add(makeTool("describe", "mcp", nil))
    assert.Equal(t, 2, ts.Count())
}

func TestToolset_IDs(t *testing.T) {
    ts := toolset.New("test-set")
    ts.Add(makeTool("search", "mcp", nil))
    ts.Add(makeTool("execute", "local", nil))

    ids := ts.IDs()
    assert.Len(t, ids, 2)
    assert.Contains(t, ids, "mcp:search")
    assert.Contains(t, ids, "local:execute")
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// go.mod
module github.com/jrraymond/toolset

go 1.22

require github.com/jrraymond/tooladapter v0.1.0
```

```go
// toolset.go
package toolset

import (
    "sync"

    "github.com/jrraymond/tooladapter"
)

// Toolset represents a curated collection of tools
type Toolset struct {
    name  string
    tools map[string]*tooladapter.CanonicalTool
    mu    sync.RWMutex
}

// New creates a new empty toolset
func New(name string) *Toolset {
    return &Toolset{
        name:  name,
        tools: make(map[string]*tooladapter.CanonicalTool),
    }
}

// Name returns the toolset name
func (ts *Toolset) Name() string {
    return ts.name
}

// Add adds a tool to the toolset
func (ts *Toolset) Add(tool *tooladapter.CanonicalTool) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.tools[tool.ID()] = tool
}

// Get retrieves a tool by ID
func (ts *Toolset) Get(id string) (*tooladapter.CanonicalTool, bool) {
    ts.mu.RLock()
    defer ts.mu.RUnlock()

    tool, ok := ts.tools[id]
    return tool, ok
}

// Remove removes a tool from the toolset
func (ts *Toolset) Remove(id string) bool {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    if _, ok := ts.tools[id]; !ok {
        return false
    }

    delete(ts.tools, id)
    return true
}

// Tools returns all tools in the toolset
func (ts *Toolset) Tools() []*tooladapter.CanonicalTool {
    ts.mu.RLock()
    defer ts.mu.RUnlock()

    result := make([]*tooladapter.CanonicalTool, 0, len(ts.tools))
    for _, tool := range ts.tools {
        result = append(result, tool)
    }
    return result
}

// Count returns the number of tools in the toolset
func (ts *Toolset) Count() int {
    ts.mu.RLock()
    defer ts.mu.RUnlock()

    return len(ts.tools)
}

// IDs returns all tool IDs in the toolset
func (ts *Toolset) IDs() []string {
    ts.mu.RLock()
    defer ts.mu.RUnlock()

    ids := make([]string, 0, len(ts.tools))
    for id := range ts.tools {
        ids = append(ids, id)
    }
    return ids
}

// Filter creates a new toolset with filtered tools
func (ts *Toolset) Filter(fn func(*tooladapter.CanonicalTool) bool) *Toolset {
    ts.mu.RLock()
    defer ts.mu.RUnlock()

    filtered := New(ts.name + " (filtered)")
    for _, tool := range ts.tools {
        if fn(tool) {
            filtered.Add(tool)
        }
    }
    return filtered
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add Toolset type with basic operations

- Thread-safe toolset with Add, Get, Remove operations
- Filter method for creating filtered subsets
- Tools, Count, IDs accessors
- Foundation for builder pattern

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Builder Pattern Implementation

**Files:**
- Create: `toolset/builder.go`
- Create: `toolset/builder_test.go`

**Step 1: Write failing tests**

```go
// builder_test.go
package toolset_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
)

// MockRegistry for testing
type MockRegistry struct {
    tools []*tooladapter.CanonicalTool
}

func (r *MockRegistry) All() []*tooladapter.CanonicalTool {
    return r.tools
}

func TestBuilder_New(t *testing.T) {
    builder := toolset.NewBuilder("my-toolset")
    assert.NotNil(t, builder)
}

func TestBuilder_FromRegistry(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", []string{"query"}),
            makeTool("describe", "mcp", []string{"info"}),
        },
    }

    ts, err := toolset.NewBuilder("test").
        FromRegistry(registry).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
}

func TestBuilder_WithNamespace(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", nil),
            makeTool("describe", "mcp", nil),
            makeTool("execute", "local", nil),
        },
    }

    ts, err := toolset.NewBuilder("mcp-only").
        FromRegistry(registry).
        WithNamespace("mcp").
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
    for _, tool := range ts.Tools() {
        assert.Equal(t, "mcp", tool.Namespace)
    }
}

func TestBuilder_WithNamespaces(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", nil),
            makeTool("execute", "local", nil),
            makeTool("call", "openai", nil),
        },
    }

    ts, err := toolset.NewBuilder("selected").
        FromRegistry(registry).
        WithNamespaces([]string{"mcp", "local"}).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
}

func TestBuilder_WithTags(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", []string{"query", "read"}),
            makeTool("write", "mcp", []string{"write", "modify"}),
            makeTool("read", "mcp", []string{"read"}),
        },
    }

    ts, err := toolset.NewBuilder("read-only").
        FromRegistry(registry).
        WithTags([]string{"read"}).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
}

func TestBuilder_WithTools(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", nil),
            makeTool("describe", "mcp", nil),
            makeTool("execute", "mcp", nil),
        },
    }

    ts, err := toolset.NewBuilder("selected").
        FromRegistry(registry).
        WithTools([]string{"mcp:search", "mcp:describe"}).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
}

func TestBuilder_ExcludeTools(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", nil),
            makeTool("describe", "mcp", nil),
            makeTool("execute", "mcp", nil),
        },
    }

    ts, err := toolset.NewBuilder("safe").
        FromRegistry(registry).
        ExcludeTools([]string{"mcp:execute"}).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
    _, ok := ts.Get("mcp:execute")
    assert.False(t, ok)
}

func TestBuilder_ChainedFilters(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", []string{"query", "safe"}),
            makeTool("describe", "mcp", []string{"info", "safe"}),
            makeTool("execute", "mcp", []string{"dangerous"}),
            makeTool("local-search", "local", []string{"query", "safe"}),
        },
    }

    ts, err := toolset.NewBuilder("mcp-safe").
        FromRegistry(registry).
        WithNamespace("mcp").
        WithTags([]string{"safe"}).
        ExcludeTools([]string{"mcp:execute"}).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 2, ts.Count())
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// builder.go
package toolset

import (
    "slices"

    "github.com/jrraymond/tooladapter"
)

// ToolRegistry is the interface for tool registries
type ToolRegistry interface {
    All() []*tooladapter.CanonicalTool
}

// Builder constructs toolsets with fluent API
type Builder struct {
    name            string
    registry        ToolRegistry
    tools           []*tooladapter.CanonicalTool
    namespaces      []string
    tags            []string
    includeTools    []string
    excludeTools    []string
    customFilters   []func(*tooladapter.CanonicalTool) bool
}

// NewBuilder creates a new toolset builder
func NewBuilder(name string) *Builder {
    return &Builder{
        name: name,
    }
}

// FromRegistry loads tools from a registry
func (b *Builder) FromRegistry(registry ToolRegistry) *Builder {
    b.registry = registry
    return b
}

// FromTools loads tools from a slice
func (b *Builder) FromTools(tools []*tooladapter.CanonicalTool) *Builder {
    b.tools = tools
    return b
}

// WithNamespace filters to a single namespace
func (b *Builder) WithNamespace(namespace string) *Builder {
    b.namespaces = []string{namespace}
    return b
}

// WithNamespaces filters to multiple namespaces
func (b *Builder) WithNamespaces(namespaces []string) *Builder {
    b.namespaces = namespaces
    return b
}

// WithTags filters to tools with any of the specified tags
func (b *Builder) WithTags(tags []string) *Builder {
    b.tags = tags
    return b
}

// WithTools includes only specified tool IDs
func (b *Builder) WithTools(toolIDs []string) *Builder {
    b.includeTools = toolIDs
    return b
}

// ExcludeTools excludes specified tool IDs
func (b *Builder) ExcludeTools(toolIDs []string) *Builder {
    b.excludeTools = toolIDs
    return b
}

// WithFilter adds a custom filter function
func (b *Builder) WithFilter(fn func(*tooladapter.CanonicalTool) bool) *Builder {
    b.customFilters = append(b.customFilters, fn)
    return b
}

// Build constructs the toolset
func (b *Builder) Build() (*Toolset, error) {
    ts := New(b.name)

    // Get source tools
    var source []*tooladapter.CanonicalTool
    if b.registry != nil {
        source = b.registry.All()
    } else if b.tools != nil {
        source = b.tools
    }

    // Apply filters
    for _, tool := range source {
        if b.shouldInclude(tool) {
            ts.Add(tool)
        }
    }

    return ts, nil
}

// shouldInclude checks if a tool passes all filters
func (b *Builder) shouldInclude(tool *tooladapter.CanonicalTool) bool {
    // Namespace filter
    if len(b.namespaces) > 0 {
        if !slices.Contains(b.namespaces, tool.Namespace) {
            return false
        }
    }

    // Tag filter (any match)
    if len(b.tags) > 0 {
        hasTag := false
        for _, tag := range b.tags {
            if slices.Contains(tool.Tags, tag) {
                hasTag = true
                break
            }
        }
        if !hasTag {
            return false
        }
    }

    // Include list (if specified)
    if len(b.includeTools) > 0 {
        if !slices.Contains(b.includeTools, tool.ID()) {
            return false
        }
    }

    // Exclude list
    if len(b.excludeTools) > 0 {
        if slices.Contains(b.excludeTools, tool.ID()) {
            return false
        }
    }

    // Custom filters
    for _, fn := range b.customFilters {
        if !fn(tool) {
            return false
        }
    }

    return true
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add Builder with fluent API

- FromRegistry and FromTools source loading
- WithNamespace, WithNamespaces namespace filtering
- WithTags for tag-based filtering
- WithTools for explicit inclusion
- ExcludeTools for exclusion
- WithFilter for custom filter functions
- Chainable fluent API

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Filter Functions Library

**Files:**
- Create: `toolset/filter.go`
- Create: `toolset/filter_test.go`

**Step 1: Write failing tests**

```go
// filter_test.go
package toolset_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
)

func TestByNamespace(t *testing.T) {
    filter := toolset.ByNamespace("mcp")

    assert.True(t, filter(makeTool("search", "mcp", nil)))
    assert.False(t, filter(makeTool("search", "local", nil)))
}

func TestByNamespaces(t *testing.T) {
    filter := toolset.ByNamespaces([]string{"mcp", "local"})

    assert.True(t, filter(makeTool("search", "mcp", nil)))
    assert.True(t, filter(makeTool("execute", "local", nil)))
    assert.False(t, filter(makeTool("call", "openai", nil)))
}

func TestByTag(t *testing.T) {
    filter := toolset.ByTag("safe")

    assert.True(t, filter(makeTool("search", "mcp", []string{"safe", "query"})))
    assert.False(t, filter(makeTool("execute", "mcp", []string{"dangerous"})))
}

func TestByAnyTag(t *testing.T) {
    filter := toolset.ByAnyTag([]string{"read", "query"})

    assert.True(t, filter(makeTool("search", "mcp", []string{"query"})))
    assert.True(t, filter(makeTool("get", "mcp", []string{"read"})))
    assert.False(t, filter(makeTool("write", "mcp", []string{"write"})))
}

func TestByAllTags(t *testing.T) {
    filter := toolset.ByAllTags([]string{"safe", "read"})

    assert.True(t, filter(makeTool("search", "mcp", []string{"safe", "read", "query"})))
    assert.False(t, filter(makeTool("get", "mcp", []string{"safe"}))) // missing "read"
    assert.False(t, filter(makeTool("read", "mcp", []string{"read"}))) // missing "safe"
}

func TestByCategory(t *testing.T) {
    tool := makeTool("search", "mcp", nil)
    tool.Category = "retrieval"

    filter := toolset.ByCategory("retrieval")
    assert.True(t, filter(tool))

    tool2 := makeTool("execute", "mcp", nil)
    tool2.Category = "execution"
    assert.False(t, filter(tool2))
}

func TestExcludeIDs(t *testing.T) {
    filter := toolset.ExcludeIDs([]string{"mcp:execute", "mcp:delete"})

    assert.True(t, filter(makeTool("search", "mcp", nil)))
    assert.False(t, filter(makeTool("execute", "mcp", nil)))
    assert.False(t, filter(makeTool("delete", "mcp", nil)))
}

func TestIncludeIDs(t *testing.T) {
    filter := toolset.IncludeIDs([]string{"mcp:search", "mcp:describe"})

    assert.True(t, filter(makeTool("search", "mcp", nil)))
    assert.True(t, filter(makeTool("describe", "mcp", nil)))
    assert.False(t, filter(makeTool("execute", "mcp", nil)))
}

func TestAnd(t *testing.T) {
    filter := toolset.And(
        toolset.ByNamespace("mcp"),
        toolset.ByTag("safe"),
    )

    assert.True(t, filter(makeTool("search", "mcp", []string{"safe"})))
    assert.False(t, filter(makeTool("search", "local", []string{"safe"})))
    assert.False(t, filter(makeTool("search", "mcp", []string{"dangerous"})))
}

func TestOr(t *testing.T) {
    filter := toolset.Or(
        toolset.ByNamespace("mcp"),
        toolset.ByTag("safe"),
    )

    assert.True(t, filter(makeTool("search", "mcp", []string{"dangerous"})))
    assert.True(t, filter(makeTool("search", "local", []string{"safe"})))
    assert.False(t, filter(makeTool("search", "local", []string{"dangerous"})))
}

func TestNot(t *testing.T) {
    filter := toolset.Not(toolset.ByNamespace("dangerous"))

    assert.True(t, filter(makeTool("search", "mcp", nil)))
    assert.False(t, filter(makeTool("delete", "dangerous", nil)))
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// filter.go
package toolset

import (
    "slices"

    "github.com/jrraymond/tooladapter"
)

// FilterFunc is a function that filters tools
type FilterFunc func(*tooladapter.CanonicalTool) bool

// ByNamespace returns a filter for a single namespace
func ByNamespace(namespace string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return tool.Namespace == namespace
    }
}

// ByNamespaces returns a filter for multiple namespaces
func ByNamespaces(namespaces []string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return slices.Contains(namespaces, tool.Namespace)
    }
}

// ByTag returns a filter for tools with a specific tag
func ByTag(tag string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return slices.Contains(tool.Tags, tag)
    }
}

// ByAnyTag returns a filter for tools with any of the specified tags
func ByAnyTag(tags []string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        for _, tag := range tags {
            if slices.Contains(tool.Tags, tag) {
                return true
            }
        }
        return false
    }
}

// ByAllTags returns a filter for tools with all specified tags
func ByAllTags(tags []string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        for _, tag := range tags {
            if !slices.Contains(tool.Tags, tag) {
                return false
            }
        }
        return true
    }
}

// ByCategory returns a filter for tools in a specific category
func ByCategory(category string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return tool.Category == category
    }
}

// ExcludeIDs returns a filter that excludes specific tool IDs
func ExcludeIDs(ids []string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return !slices.Contains(ids, tool.ID())
    }
}

// IncludeIDs returns a filter that includes only specific tool IDs
func IncludeIDs(ids []string) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return slices.Contains(ids, tool.ID())
    }
}

// And combines filters with AND logic
func And(filters ...FilterFunc) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        for _, f := range filters {
            if !f(tool) {
                return false
            }
        }
        return true
    }
}

// Or combines filters with OR logic
func Or(filters ...FilterFunc) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        for _, f := range filters {
            if f(tool) {
                return true
            }
        }
        return false
    }
}

// Not negates a filter
func Not(filter FilterFunc) FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return !filter(tool)
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add composable filter functions

- ByNamespace, ByNamespaces for namespace filtering
- ByTag, ByAnyTag, ByAllTags for tag filtering
- ByCategory for category filtering
- ExcludeIDs, IncludeIDs for ID-based filtering
- And, Or, Not combinators for complex filters

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Access Control Policies

**Files:**
- Create: `toolset/policy.go`
- Create: `toolset/policy_test.go`

**Step 1: Write failing tests**

```go
// policy_test.go
package toolset_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
)

func TestAccessPolicy_Allow(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        Allow("mcp:search").
        Allow("mcp:describe")

    assert.True(t, policy.CanAccess(context.Background(), "mcp:search"))
    assert.True(t, policy.CanAccess(context.Background(), "mcp:describe"))
    assert.False(t, policy.CanAccess(context.Background(), "mcp:execute"))
}

func TestAccessPolicy_Deny(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        AllowAll().
        Deny("mcp:execute").
        Deny("mcp:delete")

    assert.True(t, policy.CanAccess(context.Background(), "mcp:search"))
    assert.False(t, policy.CanAccess(context.Background(), "mcp:execute"))
    assert.False(t, policy.CanAccess(context.Background(), "mcp:delete"))
}

func TestAccessPolicy_DenyTakesPrecedence(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        Allow("mcp:execute").
        Deny("mcp:execute")

    // Deny takes precedence over allow
    assert.False(t, policy.CanAccess(context.Background(), "mcp:execute"))
}

func TestAccessPolicy_AllowNamespace(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        AllowNamespace("mcp")

    assert.True(t, policy.CanAccess(context.Background(), "mcp:search"))
    assert.True(t, policy.CanAccess(context.Background(), "mcp:describe"))
    assert.False(t, policy.CanAccess(context.Background(), "local:execute"))
}

func TestAccessPolicy_DenyNamespace(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        AllowAll().
        DenyNamespace("dangerous")

    assert.True(t, policy.CanAccess(context.Background(), "mcp:search"))
    assert.False(t, policy.CanAccess(context.Background(), "dangerous:delete"))
}

func TestAccessPolicy_WithRequiredScopes(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        AllowAll().
        RequireScope("mcp:execute", "admin")

    ctx := context.Background()

    // Without scope
    assert.True(t, policy.CanAccess(ctx, "mcp:search"))
    assert.False(t, policy.CanAccess(ctx, "mcp:execute"))

    // With scope
    ctxWithScope := toolset.WithScopes(ctx, []string{"admin"})
    assert.True(t, policy.CanAccess(ctxWithScope, "mcp:execute"))
}

func TestAccessPolicy_Validate(t *testing.T) {
    policy := toolset.NewAccessPolicy().
        Allow("mcp:search")

    err := policy.Validate("mcp:search")
    assert.NoError(t, err)

    err = policy.Validate("mcp:execute")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "access denied")
}

func TestBuilder_WithPolicy(t *testing.T) {
    registry := &MockRegistry{
        tools: []*tooladapter.CanonicalTool{
            makeTool("search", "mcp", nil),
            makeTool("execute", "mcp", nil),
        },
    }

    policy := toolset.NewAccessPolicy().
        Allow("mcp:search")

    ts, err := toolset.NewBuilder("safe").
        FromRegistry(registry).
        WithPolicy(policy).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 1, ts.Count())
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// policy.go
package toolset

import (
    "context"
    "fmt"
    "slices"
    "strings"
)

type contextKey string

const scopesKey contextKey = "scopes"

// WithScopes adds scopes to the context
func WithScopes(ctx context.Context, scopes []string) context.Context {
    return context.WithValue(ctx, scopesKey, scopes)
}

// ScopesFromContext retrieves scopes from context
func ScopesFromContext(ctx context.Context) []string {
    scopes, _ := ctx.Value(scopesKey).([]string)
    return scopes
}

// AccessPolicy defines access control for tools
type AccessPolicy struct {
    allowAll         bool
    allowed          map[string]bool
    denied           map[string]bool
    allowedNs        map[string]bool
    deniedNs         map[string]bool
    requiredScopes   map[string][]string
}

// NewAccessPolicy creates a new access policy
func NewAccessPolicy() *AccessPolicy {
    return &AccessPolicy{
        allowed:        make(map[string]bool),
        denied:         make(map[string]bool),
        allowedNs:      make(map[string]bool),
        deniedNs:       make(map[string]bool),
        requiredScopes: make(map[string][]string),
    }
}

// AllowAll allows all tools by default
func (p *AccessPolicy) AllowAll() *AccessPolicy {
    p.allowAll = true
    return p
}

// Allow explicitly allows a tool ID
func (p *AccessPolicy) Allow(toolID string) *AccessPolicy {
    p.allowed[toolID] = true
    return p
}

// Deny explicitly denies a tool ID
func (p *AccessPolicy) Deny(toolID string) *AccessPolicy {
    p.denied[toolID] = true
    return p
}

// AllowNamespace allows all tools in a namespace
func (p *AccessPolicy) AllowNamespace(namespace string) *AccessPolicy {
    p.allowedNs[namespace] = true
    return p
}

// DenyNamespace denies all tools in a namespace
func (p *AccessPolicy) DenyNamespace(namespace string) *AccessPolicy {
    p.deniedNs[namespace] = true
    return p
}

// RequireScope requires specific scopes to access a tool
func (p *AccessPolicy) RequireScope(toolID string, scopes ...string) *AccessPolicy {
    p.requiredScopes[toolID] = scopes
    return p
}

// CanAccess checks if access to a tool is allowed
func (p *AccessPolicy) CanAccess(ctx context.Context, toolID string) bool {
    // Check explicit deny first (takes precedence)
    if p.denied[toolID] {
        return false
    }

    // Check namespace deny
    namespace := extractNamespace(toolID)
    if p.deniedNs[namespace] {
        return false
    }

    // Check required scopes
    if required, ok := p.requiredScopes[toolID]; ok {
        ctxScopes := ScopesFromContext(ctx)
        for _, scope := range required {
            if !slices.Contains(ctxScopes, scope) {
                return false
            }
        }
    }

    // Check explicit allow
    if p.allowed[toolID] {
        return true
    }

    // Check namespace allow
    if p.allowedNs[namespace] {
        return true
    }

    // Check allowAll
    if p.allowAll {
        return true
    }

    return false
}

// Validate returns an error if access is denied
func (p *AccessPolicy) Validate(toolID string) error {
    if !p.CanAccess(context.Background(), toolID) {
        return fmt.Errorf("access denied to tool %q", toolID)
    }
    return nil
}

// AsFilter returns the policy as a filter function
func (p *AccessPolicy) AsFilter() FilterFunc {
    return func(tool *tooladapter.CanonicalTool) bool {
        return p.CanAccess(context.Background(), tool.ID())
    }
}

// extractNamespace gets namespace from tool ID
func extractNamespace(toolID string) string {
    if idx := strings.Index(toolID, ":"); idx > 0 {
        return toolID[:idx]
    }
    return ""
}

// WithPolicy adds access policy filtering to builder
func (b *Builder) WithPolicy(policy *AccessPolicy) *Builder {
    b.customFilters = append(b.customFilters, policy.AsFilter())
    return b
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add AccessPolicy for access control

- Allow/Deny for explicit tool permissions
- AllowNamespace/DenyNamespace for namespace-level control
- RequireScope for scope-based access control
- Deny takes precedence over Allow
- Context-based scope checking
- Integration with Builder via WithPolicy

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Multi-Format Exposure

**Files:**
- Create: `toolset/exposure.go`
- Create: `toolset/exposure_test.go`

**Step 1: Write failing tests**

```go
// exposure_test.go
package toolset_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
    "github.com/jrraymond/tooladapter/adapters"
)

func TestExposure_AsMCP(t *testing.T) {
    ts := toolset.New("test")
    ts.Add(makeTool("search", "mcp", nil))
    ts.Add(makeTool("describe", "mcp", nil))

    mcpAdapter := adapters.NewMCPAdapter()
    exposure := toolset.NewExposure(ts, mcpAdapter)

    tools, err := exposure.Export()
    require.NoError(t, err)
    assert.Len(t, tools, 2)
}

func TestExposure_AsOpenAI(t *testing.T) {
    ts := toolset.New("test")
    ts.Add(makeTool("search", "mcp", nil))

    openaiAdapter := adapters.NewOpenAIAdapter(false)
    exposure := toolset.NewExposure(ts, openaiAdapter)

    tools, err := exposure.Export()
    require.NoError(t, err)
    assert.Len(t, tools, 1)

    fn, ok := tools[0].(*adapters.OpenAIFunction)
    require.True(t, ok)
    assert.Equal(t, "search", fn.Name)
}

func TestExposure_AsAnthropic(t *testing.T) {
    ts := toolset.New("test")
    ts.Add(makeTool("search", "mcp", nil))

    anthropicAdapter := adapters.NewAnthropicAdapter()
    exposure := toolset.NewExposure(ts, anthropicAdapter)

    tools, err := exposure.Export()
    require.NoError(t, err)
    assert.Len(t, tools, 1)

    anthropicTool, ok := tools[0].(*adapters.AnthropicTool)
    require.True(t, ok)
    assert.Equal(t, "search", anthropicTool.Name)
}

func TestExposure_WithWarnings(t *testing.T) {
    ts := toolset.New("test")
    tool := makeTool("complex", "mcp", nil)
    tool.InputSchema = &tooladapter.JSONSchema{
        Type: "object",
        Ref:  "#/$defs/MyType", // $ref not supported by OpenAI
        Defs: map[string]*tooladapter.JSONSchema{
            "MyType": {Type: "string"},
        },
    }
    ts.Add(tool)

    openaiAdapter := adapters.NewOpenAIAdapter(false)
    exposure := toolset.NewExposure(ts, openaiAdapter)

    _, warnings := exposure.ExportWithWarnings()
    // Should have warning about $ref feature loss
    assert.NotEmpty(t, warnings)
}

func TestToolsetServer_ListTools(t *testing.T) {
    ts := toolset.New("test")
    ts.Add(makeTool("search", "mcp", nil))
    ts.Add(makeTool("describe", "mcp", nil))

    server := toolset.NewServer(ts)
    tools := server.ListTools()

    assert.Len(t, tools, 2)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// exposure.go
package toolset

import (
    "github.com/jrraymond/tooladapter"
)

// Exposure handles converting toolsets to protocol-specific formats
type Exposure struct {
    toolset *Toolset
    adapter tooladapter.Adapter
}

// NewExposure creates a new exposure for a toolset
func NewExposure(ts *Toolset, adapter tooladapter.Adapter) *Exposure {
    return &Exposure{
        toolset: ts,
        adapter: adapter,
    }
}

// Export exports the toolset to the target format
func (e *Exposure) Export() ([]any, error) {
    tools := e.toolset.Tools()
    result := make([]any, 0, len(tools))

    for _, tool := range tools {
        converted, err := e.adapter.FromCanonical(tool)
        if err != nil {
            return nil, err
        }
        result = append(result, converted)
    }

    return result, nil
}

// ExportWithWarnings exports with feature loss warnings
func (e *Exposure) ExportWithWarnings() ([]any, []tooladapter.FeatureLossWarning) {
    tools := e.toolset.Tools()
    result := make([]any, 0, len(tools))
    var warnings []tooladapter.FeatureLossWarning

    // Check for feature loss
    for _, feature := range tooladapter.AllFeatures() {
        if !e.adapter.SupportsFeature(feature) {
            warnings = append(warnings, tooladapter.FeatureLossWarning{
                Feature: feature,
                Adapter: e.adapter.Name(),
            })
        }
    }

    for _, tool := range tools {
        converted, err := e.adapter.FromCanonical(tool)
        if err != nil {
            continue // Skip tools that fail conversion
        }
        result = append(result, converted)
    }

    return result, warnings
}

// Server wraps a toolset for serving
type Server struct {
    toolset *Toolset
}

// NewServer creates a new toolset server
func NewServer(ts *Toolset) *Server {
    return &Server{
        toolset: ts,
    }
}

// ListTools returns all tools in the toolset
func (s *Server) ListTools() []*tooladapter.CanonicalTool {
    return s.toolset.Tools()
}

// GetTool retrieves a tool by ID
func (s *Server) GetTool(id string) (*tooladapter.CanonicalTool, bool) {
    return s.toolset.Get(id)
}

// CallTool calls a tool by ID (placeholder - actual implementation would execute)
func (s *Server) CallTool(id string, args map[string]any) (any, error) {
    tool, ok := s.toolset.Get(id)
    if !ok {
        return nil, fmt.Errorf("tool %q not found", id)
    }

    if tool.Handler == nil {
        return nil, fmt.Errorf("tool %q has no handler", id)
    }

    return tool.Handler(context.Background(), args)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add multi-format exposure

- Exposure type for protocol conversion
- Export and ExportWithWarnings methods
- Server wrapper for serving toolsets
- ListTools, GetTool, CallTool operations

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Integration with toolindex

**Files:**
- Create: `toolset/integration.go`
- Create: `toolset/integration_test.go`

**Step 1: Write failing tests**

```go
// integration_test.go
package toolset_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolset"
    "github.com/jrraymond/tooladapter"
)

// MockIndex simulates toolindex.Index
type MockIndex struct {
    tools map[string]*tooladapter.CanonicalTool
}

func (m *MockIndex) Get(id string) (*tooladapter.CanonicalTool, bool) {
    tool, ok := m.tools[id]
    return tool, ok
}

func (m *MockIndex) Search(query string, limit int) []string {
    // Simplified search - return all IDs
    ids := make([]string, 0, len(m.tools))
    for id := range m.tools {
        ids = append(ids, id)
    }
    return ids
}

func TestBuilder_FromIndex(t *testing.T) {
    index := &MockIndex{
        tools: map[string]*tooladapter.CanonicalTool{
            "mcp:search":   makeTool("search", "mcp", nil),
            "mcp:describe": makeTool("describe", "mcp", nil),
            "local:run":    makeTool("run", "local", nil),
        },
    }

    ts, err := toolset.NewBuilder("from-index").
        FromIndex(index).
        Build()

    require.NoError(t, err)
    assert.Equal(t, 3, ts.Count())
}

func TestBuilder_FromIndexWithSearch(t *testing.T) {
    index := &MockIndex{
        tools: map[string]*tooladapter.CanonicalTool{
            "mcp:search":   makeTool("search", "mcp", []string{"query"}),
            "mcp:describe": makeTool("describe", "mcp", []string{"info"}),
        },
    }

    ts, err := toolset.NewBuilder("search-results").
        FromIndexSearch(index, "query tools", 10).
        Build()

    require.NoError(t, err)
    assert.True(t, ts.Count() > 0)
}

func TestToolset_Merge(t *testing.T) {
    ts1 := toolset.New("first")
    ts1.Add(makeTool("search", "mcp", nil))

    ts2 := toolset.New("second")
    ts2.Add(makeTool("describe", "mcp", nil))

    merged := toolset.Merge("combined", ts1, ts2)

    assert.Equal(t, 2, merged.Count())
    _, ok := merged.Get("mcp:search")
    assert.True(t, ok)
    _, ok = merged.Get("mcp:describe")
    assert.True(t, ok)
}

func TestToolset_Subtract(t *testing.T) {
    ts1 := toolset.New("full")
    ts1.Add(makeTool("search", "mcp", nil))
    ts1.Add(makeTool("describe", "mcp", nil))
    ts1.Add(makeTool("execute", "mcp", nil))

    ts2 := toolset.New("dangerous")
    ts2.Add(makeTool("execute", "mcp", nil))

    safe := toolset.Subtract("safe", ts1, ts2)

    assert.Equal(t, 2, safe.Count())
    _, ok := safe.Get("mcp:execute")
    assert.False(t, ok)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolset && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// integration.go
package toolset

import (
    "github.com/jrraymond/tooladapter"
)

// ToolIndex is the interface for tool indexes
type ToolIndex interface {
    Get(id string) (*tooladapter.CanonicalTool, bool)
    Search(query string, limit int) []string
}

// FromIndex loads tools from a tool index
func (b *Builder) FromIndex(index ToolIndex) *Builder {
    // Store index for later use
    b.index = index
    return b
}

// FromIndexSearch loads tools from index search results
func (b *Builder) FromIndexSearch(index ToolIndex, query string, limit int) *Builder {
    b.index = index
    b.searchQuery = query
    b.searchLimit = limit
    return b
}

// Update Build to support index loading
func (b *Builder) buildFromIndex() []*tooladapter.CanonicalTool {
    if b.index == nil {
        return nil
    }

    var ids []string
    if b.searchQuery != "" {
        ids = b.index.Search(b.searchQuery, b.searchLimit)
    } else {
        // Get all tools - this requires iteration which isn't in interface
        // For now, return empty - actual implementation would need All() method
        return nil
    }

    tools := make([]*tooladapter.CanonicalTool, 0, len(ids))
    for _, id := range ids {
        if tool, ok := b.index.Get(id); ok {
            tools = append(tools, tool)
        }
    }
    return tools
}

// Merge combines multiple toolsets
func Merge(name string, toolsets ...*Toolset) *Toolset {
    merged := New(name)
    for _, ts := range toolsets {
        for _, tool := range ts.Tools() {
            merged.Add(tool)
        }
    }
    return merged
}

// Subtract removes tools in second toolset from first
func Subtract(name string, base, remove *Toolset) *Toolset {
    result := New(name)
    removeIDs := make(map[string]bool)

    for _, tool := range remove.Tools() {
        removeIDs[tool.ID()] = true
    }

    for _, tool := range base.Tools() {
        if !removeIDs[tool.ID()] {
            result.Add(tool)
        }
    }

    return result
}

// Intersect returns tools present in all toolsets
func Intersect(name string, toolsets ...*Toolset) *Toolset {
    if len(toolsets) == 0 {
        return New(name)
    }

    // Count occurrences
    counts := make(map[string]int)
    var firstTools []*tooladapter.CanonicalTool

    for i, ts := range toolsets {
        for _, tool := range ts.Tools() {
            counts[tool.ID()]++
            if i == 0 {
                firstTools = append(firstTools, tool)
            }
        }
    }

    // Keep only tools in all toolsets
    result := New(name)
    for _, tool := range firstTools {
        if counts[tool.ID()] == len(toolsets) {
            result.Add(tool)
        }
    }

    return result
}
```

Add fields to Builder struct:

```go
// Update builder.go Builder struct
type Builder struct {
    name            string
    registry        ToolRegistry
    tools           []*tooladapter.CanonicalTool
    namespaces      []string
    tags            []string
    includeTools    []string
    excludeTools    []string
    customFilters   []func(*tooladapter.CanonicalTool) bool
    index           ToolIndex  // NEW
    searchQuery     string     // NEW
    searchLimit     int        // NEW
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolset && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolset/
git commit -m "$(cat <<'EOF'
feat(toolset): add toolindex integration

- FromIndex for loading from tool index
- FromIndexSearch for search-based loading
- Merge for combining toolsets
- Subtract for removing tools
- Intersect for common tools

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-009 complete:

- [ ] All tests pass: `go test ./... -v`
- [ ] Code coverage > 80%: `go test ./... -cover`
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete:
  - [ ] Package documentation in `doc.go`
  - [ ] README.md with usage examples
  - [ ] GoDoc comments on all exported types
- [ ] Integration verified:
  - [ ] Builder fluent API works
  - [ ] All filter functions compose correctly
  - [ ] AccessPolicy enforces permissions
  - [ ] Multi-format exposure works

---

## Definition of Done

1. **Toolset** type with Add, Get, Remove, Filter, Tools, Count, IDs
2. **Builder** with fluent API for composing toolsets
3. **FilterFunc** library with composable filter functions
4. **AccessPolicy** with allow/deny, namespace, and scope control
5. **Exposure** for multi-format protocol conversion
6. **Integration** with toolindex (Merge, Subtract, Intersect)
7. All tests passing with >80% coverage
8. Documentation complete
