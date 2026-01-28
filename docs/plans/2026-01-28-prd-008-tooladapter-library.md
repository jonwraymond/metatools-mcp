# PRD-008: tooladapter Library Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a protocol-agnostic tool abstraction library that enables bidirectional conversion between MCP, OpenAI, Anthropic, LangChain, and OpenAPI tool formats.

**Architecture:** Introduce a canonical tool representation that stores the superset of all schema information, enabling lossless conversion between formats. Protocol adapters implement a common interface for bidirectional conversion.

**Tech Stack:** Go, toolmodel dependency, JSON Schema validation

---

## Overview

The `tooladapter` library provides protocol-agnostic tool handling, enabling tools from any source to be exposed through multiple transport protocols and consumed by various AI agent frameworks.

**Reference:** [protocol-agnostic-tools.md](../proposals/protocol-agnostic-tools.md)

---

## Directory Structure

```
tooladapter/
├── canonical.go        # CanonicalTool and JSONSchema types
├── canonical_test.go   # CanonicalTool tests
├── adapter.go          # Adapter interface
├── adapter_test.go     # Adapter interface tests
├── registry.go         # AdapterRegistry implementation
├── registry_test.go    # Registry tests
├── adapters/
│   ├── mcp.go          # MCP ↔ Canonical adapter
│   ├── mcp_test.go
│   ├── openai.go       # OpenAI ↔ Canonical adapter
│   ├── openai_test.go
│   ├── anthropic.go    # Anthropic ↔ Canonical adapter
│   ├── anthropic_test.go
│   ├── langchain.go    # LangChain ↔ Canonical adapter (optional)
│   └── openapi.go      # OpenAPI ↔ Canonical adapter (import only)
├── schema/
│   ├── convert.go      # JSON Schema version conversion
│   ├── convert_test.go
│   ├── validate.go     # Schema validation
│   └── validate_test.go
├── doc.go              # Package documentation
├── go.mod
└── go.sum
```

---

## Task 1: CanonicalTool and JSONSchema Types

**Files:**
- Create: `tooladapter/canonical.go`
- Create: `tooladapter/canonical_test.go`
- Create: `tooladapter/go.mod`

**Step 1: Write failing tests**

```go
// canonical_test.go
package tooladapter_test

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/tooladapter"
)

func TestCanonicalTool_ID(t *testing.T) {
    tool := &tooladapter.CanonicalTool{
        Namespace: "mcp",
        Name:      "search",
    }
    assert.Equal(t, "mcp:search", tool.ID())
}

func TestCanonicalTool_IDWithEmptyNamespace(t *testing.T) {
    tool := &tooladapter.CanonicalTool{
        Name: "search",
    }
    assert.Equal(t, "search", tool.ID())
}

func TestCanonicalTool_Validate(t *testing.T) {
    tests := []struct {
        name    string
        tool    *tooladapter.CanonicalTool
        wantErr bool
    }{
        {
            name: "valid tool",
            tool: &tooladapter.CanonicalTool{
                Name:        "search",
                Description: "Search for tools",
                InputSchema: &tooladapter.JSONSchema{
                    Type: "object",
                },
            },
            wantErr: false,
        },
        {
            name: "missing name",
            tool: &tooladapter.CanonicalTool{
                Description: "Search for tools",
            },
            wantErr: true,
        },
        {
            name: "missing input schema",
            tool: &tooladapter.CanonicalTool{
                Name: "search",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.tool.Validate()
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestJSONSchema_DeepCopy(t *testing.T) {
    original := &tooladapter.JSONSchema{
        Type: "object",
        Properties: map[string]*tooladapter.JSONSchema{
            "query": {
                Type:        "string",
                Description: "Search query",
            },
        },
        Required: []string{"query"},
    }

    copied := original.DeepCopy()

    // Modify original
    original.Properties["query"].Description = "Modified"
    original.Required = append(original.Required, "extra")

    // Verify copy is unchanged
    assert.Equal(t, "Search query", copied.Properties["query"].Description)
    assert.Equal(t, []string{"query"}, copied.Required)
}

func TestJSONSchema_ToMap(t *testing.T) {
    schema := &tooladapter.JSONSchema{
        Type: "object",
        Properties: map[string]*tooladapter.JSONSchema{
            "query": {
                Type: "string",
            },
        },
        Required: []string{"query"},
    }

    m := schema.ToMap()

    assert.Equal(t, "object", m["type"])
    props := m["properties"].(map[string]any)
    query := props["query"].(map[string]any)
    assert.Equal(t, "string", query["type"])
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./... -v`
Expected: FAIL with "package tooladapter is not in std"

**Step 3: Write minimal implementation**

```go
// go.mod
module github.com/jrraymond/tooladapter

go 1.22

require github.com/jrraymond/toolmodel v0.1.2
```

```go
// canonical.go
package tooladapter

import (
    "errors"
    "time"
)

// CanonicalTool is the protocol-agnostic tool representation
type CanonicalTool struct {
    // Identity
    Namespace string // Source namespace (e.g., "mcp", "openai")
    Name      string // Tool name within namespace
    Version   string // Semantic version

    // Metadata
    Description string   // Human-readable description
    Category    string   // Tool category for grouping
    Tags        []string // Searchable tags

    // Schema (superset of all protocol schemas)
    InputSchema  *JSONSchema // Full JSON Schema for inputs
    OutputSchema *JSONSchema // Optional output schema

    // Execution
    Handler ToolHandler   // Execution function
    Timeout time.Duration // Execution timeout

    // Source tracking
    SourceFormat string         // Original format (mcp, openai, anthropic, etc.)
    SourceMeta   map[string]any // Protocol-specific metadata preserved

    // Access control
    RequiredScopes []string // OAuth scopes or permissions required
}

// ID returns the fully qualified tool identifier
func (c *CanonicalTool) ID() string {
    if c.Namespace == "" {
        return c.Name
    }
    return c.Namespace + ":" + c.Name
}

// Validate checks if the tool is valid
func (c *CanonicalTool) Validate() error {
    if c.Name == "" {
        return errors.New("tool name is required")
    }
    if c.InputSchema == nil {
        return errors.New("input schema is required")
    }
    return nil
}

// ToolHandler executes the tool
type ToolHandler func(ctx context.Context, input map[string]any) (any, error)

// JSONSchema represents a full JSON Schema with all features
type JSONSchema struct {
    Type        string                 `json:"type"`
    Properties  map[string]*JSONSchema `json:"properties,omitempty"`
    Required    []string               `json:"required,omitempty"`
    Items       *JSONSchema            `json:"items,omitempty"`
    Description string                 `json:"description,omitempty"`

    // Extended schema features
    Enum      []any    `json:"enum,omitempty"`
    Const     any      `json:"const,omitempty"`
    Default   any      `json:"default,omitempty"`
    Minimum   *float64 `json:"minimum,omitempty"`
    Maximum   *float64 `json:"maximum,omitempty"`
    MinLength *int     `json:"minLength,omitempty"`
    MaxLength *int     `json:"maxLength,omitempty"`
    Pattern   string   `json:"pattern,omitempty"`
    Format    string   `json:"format,omitempty"`

    // JSON Schema draft compatibility
    Ref  string                 `json:"$ref,omitempty"`
    Defs map[string]*JSONSchema `json:"$defs,omitempty"`

    // Additional properties
    AdditionalProperties *bool `json:"additionalProperties,omitempty"`
}

// DeepCopy creates a deep copy of the schema
func (s *JSONSchema) DeepCopy() *JSONSchema {
    if s == nil {
        return nil
    }

    copied := &JSONSchema{
        Type:        s.Type,
        Description: s.Description,
        Pattern:     s.Pattern,
        Format:      s.Format,
        Ref:         s.Ref,
    }

    // Copy required slice
    if s.Required != nil {
        copied.Required = make([]string, len(s.Required))
        copy(copied.Required, s.Required)
    }

    // Copy enum slice
    if s.Enum != nil {
        copied.Enum = make([]any, len(s.Enum))
        copy(copied.Enum, s.Enum)
    }

    // Copy pointer fields
    if s.Minimum != nil {
        v := *s.Minimum
        copied.Minimum = &v
    }
    if s.Maximum != nil {
        v := *s.Maximum
        copied.Maximum = &v
    }
    if s.MinLength != nil {
        v := *s.MinLength
        copied.MinLength = &v
    }
    if s.MaxLength != nil {
        v := *s.MaxLength
        copied.MaxLength = &v
    }
    if s.AdditionalProperties != nil {
        v := *s.AdditionalProperties
        copied.AdditionalProperties = &v
    }

    // Deep copy properties
    if s.Properties != nil {
        copied.Properties = make(map[string]*JSONSchema)
        for k, v := range s.Properties {
            copied.Properties[k] = v.DeepCopy()
        }
    }

    // Deep copy items
    copied.Items = s.Items.DeepCopy()

    // Deep copy defs
    if s.Defs != nil {
        copied.Defs = make(map[string]*JSONSchema)
        for k, v := range s.Defs {
            copied.Defs[k] = v.DeepCopy()
        }
    }

    // Copy Const and Default (shallow copy for primitives)
    copied.Const = s.Const
    copied.Default = s.Default

    return copied
}

// ToMap converts the schema to a map[string]any
func (s *JSONSchema) ToMap() map[string]any {
    if s == nil {
        return nil
    }

    m := make(map[string]any)

    if s.Type != "" {
        m["type"] = s.Type
    }
    if s.Description != "" {
        m["description"] = s.Description
    }
    if len(s.Required) > 0 {
        m["required"] = s.Required
    }
    if len(s.Enum) > 0 {
        m["enum"] = s.Enum
    }
    if s.Const != nil {
        m["const"] = s.Const
    }
    if s.Default != nil {
        m["default"] = s.Default
    }
    if s.Minimum != nil {
        m["minimum"] = *s.Minimum
    }
    if s.Maximum != nil {
        m["maximum"] = *s.Maximum
    }
    if s.MinLength != nil {
        m["minLength"] = *s.MinLength
    }
    if s.MaxLength != nil {
        m["maxLength"] = *s.MaxLength
    }
    if s.Pattern != "" {
        m["pattern"] = s.Pattern
    }
    if s.Format != "" {
        m["format"] = s.Format
    }
    if s.Ref != "" {
        m["$ref"] = s.Ref
    }
    if s.AdditionalProperties != nil {
        m["additionalProperties"] = *s.AdditionalProperties
    }

    // Convert properties
    if len(s.Properties) > 0 {
        props := make(map[string]any)
        for k, v := range s.Properties {
            props[k] = v.ToMap()
        }
        m["properties"] = props
    }

    // Convert items
    if s.Items != nil {
        m["items"] = s.Items.ToMap()
    }

    // Convert defs
    if len(s.Defs) > 0 {
        defs := make(map[string]any)
        for k, v := range s.Defs {
            defs[k] = v.ToMap()
        }
        m["$defs"] = defs
    }

    return m
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add CanonicalTool and JSONSchema types

- CanonicalTool provides protocol-agnostic tool representation
- JSONSchema supports full JSON Schema draft features
- DeepCopy for immutable schema operations
- ToMap for serialization to map[string]any

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Adapter Interface and SchemaFeature

**Files:**
- Create: `tooladapter/adapter.go`
- Create: `tooladapter/adapter_test.go`

**Step 1: Write failing tests**

```go
// adapter_test.go
package tooladapter_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/jrraymond/tooladapter"
)

func TestSchemaFeature_String(t *testing.T) {
    tests := []struct {
        feature  tooladapter.SchemaFeature
        expected string
    }{
        {tooladapter.FeatureNestedObjects, "nested_objects"},
        {tooladapter.FeatureArrays, "arrays"},
        {tooladapter.FeatureEnums, "enums"},
        {tooladapter.FeaturePatternValidation, "pattern_validation"},
        {tooladapter.FeatureRefDefinitions, "ref_definitions"},
        {tooladapter.FeatureNullable, "nullable"},
        {tooladapter.FeatureAnyOf, "any_of"},
        {tooladapter.FeatureOneOf, "one_of"},
    }

    for _, tt := range tests {
        t.Run(tt.expected, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.feature.String())
        })
    }
}

func TestConversionError_Error(t *testing.T) {
    err := &tooladapter.ConversionError{
        Adapter:   "openai",
        Direction: "to_canonical",
        Cause:     errors.New("invalid schema"),
    }

    assert.Contains(t, err.Error(), "openai")
    assert.Contains(t, err.Error(), "to_canonical")
    assert.Contains(t, err.Error(), "invalid schema")
}

func TestFeatureLossWarning(t *testing.T) {
    warning := tooladapter.FeatureLossWarning{
        Feature:     tooladapter.FeatureRefDefinitions,
        Adapter:     "openai",
        Description: "$ref definitions were flattened",
    }

    assert.Equal(t, tooladapter.FeatureRefDefinitions, warning.Feature)
    assert.Contains(t, warning.String(), "ref_definitions")
    assert.Contains(t, warning.String(), "openai")
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// adapter.go
package tooladapter

import (
    "context"
    "fmt"
)

// Adapter converts between canonical and protocol-specific formats
type Adapter interface {
    // Name returns the adapter identifier (e.g., "mcp", "openai")
    Name() string

    // ToCanonical converts protocol-specific tool to canonical form
    ToCanonical(raw any) (*CanonicalTool, error)

    // FromCanonical converts canonical tool to protocol-specific form
    FromCanonical(tool *CanonicalTool) (any, error)

    // SupportsFeature checks if adapter supports a schema feature
    SupportsFeature(feature SchemaFeature) bool
}

// SchemaFeature represents JSON Schema features that may not be universally supported
type SchemaFeature int

const (
    FeatureNestedObjects SchemaFeature = iota
    FeatureArrays
    FeatureEnums
    FeaturePatternValidation
    FeatureRefDefinitions
    FeatureNullable
    FeatureAnyOf
    FeatureOneOf
)

// String returns the string representation of the feature
func (f SchemaFeature) String() string {
    switch f {
    case FeatureNestedObjects:
        return "nested_objects"
    case FeatureArrays:
        return "arrays"
    case FeatureEnums:
        return "enums"
    case FeaturePatternValidation:
        return "pattern_validation"
    case FeatureRefDefinitions:
        return "ref_definitions"
    case FeatureNullable:
        return "nullable"
    case FeatureAnyOf:
        return "any_of"
    case FeatureOneOf:
        return "one_of"
    default:
        return "unknown"
    }
}

// AllFeatures returns all schema features
func AllFeatures() []SchemaFeature {
    return []SchemaFeature{
        FeatureNestedObjects,
        FeatureArrays,
        FeatureEnums,
        FeaturePatternValidation,
        FeatureRefDefinitions,
        FeatureNullable,
        FeatureAnyOf,
        FeatureOneOf,
    }
}

// ConversionError represents an error during tool conversion
type ConversionError struct {
    Adapter   string
    Direction string // "to_canonical" or "from_canonical"
    Cause     error
}

func (e *ConversionError) Error() string {
    return fmt.Sprintf("adapter %s %s: %v", e.Adapter, e.Direction, e.Cause)
}

func (e *ConversionError) Unwrap() error {
    return e.Cause
}

// FeatureLossWarning indicates a schema feature was lost during conversion
type FeatureLossWarning struct {
    Feature     SchemaFeature
    Adapter     string
    Description string
}

func (w FeatureLossWarning) String() string {
    return fmt.Sprintf("feature %s not supported by %s: %s",
        w.Feature.String(), w.Adapter, w.Description)
}

// ConversionResult contains the conversion result and any warnings
type ConversionResult struct {
    Tool     any
    Warnings []FeatureLossWarning
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add Adapter interface and SchemaFeature enum

- Adapter interface for bidirectional tool conversion
- SchemaFeature enum for feature support detection
- ConversionError for typed error handling
- FeatureLossWarning for tracking schema degradation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: AdapterRegistry Implementation

**Files:**
- Create: `tooladapter/registry.go`
- Create: `tooladapter/registry_test.go`

**Step 1: Write failing tests**

```go
// registry_test.go
package tooladapter_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/tooladapter"
)

// mockAdapter for testing
type mockAdapter struct {
    name string
}

func (m *mockAdapter) Name() string { return m.name }
func (m *mockAdapter) ToCanonical(raw any) (*tooladapter.CanonicalTool, error) {
    return nil, nil
}
func (m *mockAdapter) FromCanonical(tool *tooladapter.CanonicalTool) (any, error) {
    return nil, nil
}
func (m *mockAdapter) SupportsFeature(f tooladapter.SchemaFeature) bool {
    return true
}

func TestAdapterRegistry_Register(t *testing.T) {
    reg := tooladapter.NewAdapterRegistry()
    adapter := &mockAdapter{name: "test"}

    err := reg.Register(adapter)
    require.NoError(t, err)

    got, err := reg.Get("test")
    require.NoError(t, err)
    assert.Equal(t, adapter, got)
}

func TestAdapterRegistry_RegisterDuplicate(t *testing.T) {
    reg := tooladapter.NewAdapterRegistry()
    adapter1 := &mockAdapter{name: "test"}
    adapter2 := &mockAdapter{name: "test"}

    err := reg.Register(adapter1)
    require.NoError(t, err)

    err = reg.Register(adapter2)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "already registered")
}

func TestAdapterRegistry_GetNotFound(t *testing.T) {
    reg := tooladapter.NewAdapterRegistry()

    _, err := reg.Get("nonexistent")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}

func TestAdapterRegistry_List(t *testing.T) {
    reg := tooladapter.NewAdapterRegistry()
    reg.Register(&mockAdapter{name: "mcp"})
    reg.Register(&mockAdapter{name: "openai"})
    reg.Register(&mockAdapter{name: "anthropic"})

    names := reg.List()
    assert.Len(t, names, 3)
    assert.Contains(t, names, "mcp")
    assert.Contains(t, names, "openai")
    assert.Contains(t, names, "anthropic")
}

func TestAdapterRegistry_Unregister(t *testing.T) {
    reg := tooladapter.NewAdapterRegistry()
    reg.Register(&mockAdapter{name: "test"})

    err := reg.Unregister("test")
    require.NoError(t, err)

    _, err = reg.Get("test")
    require.Error(t, err)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// registry.go
package tooladapter

import (
    "fmt"
    "sync"
)

// AdapterRegistry manages protocol adapters
type AdapterRegistry struct {
    adapters map[string]Adapter
    mu       sync.RWMutex
}

// NewAdapterRegistry creates a new adapter registry
func NewAdapterRegistry() *AdapterRegistry {
    return &AdapterRegistry{
        adapters: make(map[string]Adapter),
    }
}

// Register adds an adapter to the registry
func (r *AdapterRegistry) Register(adapter Adapter) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := adapter.Name()
    if _, exists := r.adapters[name]; exists {
        return fmt.Errorf("adapter %q already registered", name)
    }

    r.adapters[name] = adapter
    return nil
}

// Get retrieves an adapter by name
func (r *AdapterRegistry) Get(name string) (Adapter, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    adapter, ok := r.adapters[name]
    if !ok {
        return nil, fmt.Errorf("adapter %q not found", name)
    }

    return adapter, nil
}

// List returns all registered adapter names
func (r *AdapterRegistry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    names := make([]string, 0, len(r.adapters))
    for name := range r.adapters {
        names = append(names, name)
    }
    return names
}

// Unregister removes an adapter from the registry
func (r *AdapterRegistry) Unregister(name string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.adapters[name]; !exists {
        return fmt.Errorf("adapter %q not found", name)
    }

    delete(r.adapters, name)
    return nil
}

// Convert converts a tool between formats using registered adapters
func (r *AdapterRegistry) Convert(tool any, fromFormat, toFormat string) (*ConversionResult, error) {
    fromAdapter, err := r.Get(fromFormat)
    if err != nil {
        return nil, fmt.Errorf("source adapter: %w", err)
    }

    toAdapter, err := r.Get(toFormat)
    if err != nil {
        return nil, fmt.Errorf("target adapter: %w", err)
    }

    // Convert to canonical
    canonical, err := fromAdapter.ToCanonical(tool)
    if err != nil {
        return nil, &ConversionError{
            Adapter:   fromFormat,
            Direction: "to_canonical",
            Cause:     err,
        }
    }

    // Track feature loss warnings
    var warnings []FeatureLossWarning
    for _, feature := range AllFeatures() {
        if !toAdapter.SupportsFeature(feature) {
            warnings = append(warnings, FeatureLossWarning{
                Feature:     feature,
                Adapter:     toFormat,
                Description: fmt.Sprintf("feature %s not supported", feature.String()),
            })
        }
    }

    // Convert from canonical to target format
    result, err := toAdapter.FromCanonical(canonical)
    if err != nil {
        return nil, &ConversionError{
            Adapter:   toFormat,
            Direction: "from_canonical",
            Cause:     err,
        }
    }

    return &ConversionResult{
        Tool:     result,
        Warnings: warnings,
    }, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add AdapterRegistry for managing adapters

- Thread-safe adapter registration and lookup
- Convert method for bidirectional tool conversion
- Feature loss tracking during conversion
- List and Unregister operations

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: MCP Adapter Implementation

**Files:**
- Create: `tooladapter/adapters/mcp.go`
- Create: `tooladapter/adapters/mcp_test.go`

**Step 1: Write failing tests**

```go
// adapters/mcp_test.go
package adapters_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/tooladapter"
    "github.com/jrraymond/tooladapter/adapters"
    "github.com/mark3labs/mcp-go/mcp"
)

func TestMCPAdapter_Name(t *testing.T) {
    adapter := adapters.NewMCPAdapter()
    assert.Equal(t, "mcp", adapter.Name())
}

func TestMCPAdapter_ToCanonical(t *testing.T) {
    adapter := adapters.NewMCPAdapter()

    mcpTool := &mcp.Tool{
        Name:        "search_files",
        Description: "Search for files",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "query": map[string]any{
                    "type":        "string",
                    "description": "Search query",
                },
            },
            Required: []string{"query"},
        },
    }

    canonical, err := adapter.ToCanonical(mcpTool)
    require.NoError(t, err)

    assert.Equal(t, "search_files", canonical.Name)
    assert.Equal(t, "mcp", canonical.Namespace)
    assert.Equal(t, "Search for files", canonical.Description)
    assert.Equal(t, "mcp", canonical.SourceFormat)
    assert.NotNil(t, canonical.InputSchema)
    assert.Equal(t, "object", canonical.InputSchema.Type)
}

func TestMCPAdapter_FromCanonical(t *testing.T) {
    adapter := adapters.NewMCPAdapter()

    canonical := &tooladapter.CanonicalTool{
        Namespace:   "test",
        Name:        "my_tool",
        Description: "A test tool",
        InputSchema: &tooladapter.JSONSchema{
            Type: "object",
            Properties: map[string]*tooladapter.JSONSchema{
                "input": {
                    Type:        "string",
                    Description: "Input value",
                },
            },
            Required: []string{"input"},
        },
    }

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    mcpTool, ok := result.(*mcp.Tool)
    require.True(t, ok)

    assert.Equal(t, "my_tool", mcpTool.Name)
    assert.Equal(t, "A test tool", mcpTool.Description)
    assert.Equal(t, "object", mcpTool.InputSchema.Type)
}

func TestMCPAdapter_SupportsAllFeatures(t *testing.T) {
    adapter := adapters.NewMCPAdapter()

    for _, feature := range tooladapter.AllFeatures() {
        assert.True(t, adapter.SupportsFeature(feature),
            "MCP should support %s", feature.String())
    }
}

func TestMCPAdapter_RoundTrip(t *testing.T) {
    adapter := adapters.NewMCPAdapter()

    original := &mcp.Tool{
        Name:        "roundtrip_tool",
        Description: "Test round trip conversion",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "name": map[string]any{
                    "type": "string",
                },
                "count": map[string]any{
                    "type":    "integer",
                    "minimum": 0,
                },
            },
            Required: []string{"name"},
        },
    }

    canonical, err := adapter.ToCanonical(original)
    require.NoError(t, err)

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    restored, ok := result.(*mcp.Tool)
    require.True(t, ok)

    assert.Equal(t, original.Name, restored.Name)
    assert.Equal(t, original.Description, restored.Description)
    assert.Equal(t, original.InputSchema.Type, restored.InputSchema.Type)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// adapters/mcp.go
package adapters

import (
    "fmt"

    "github.com/jrraymond/tooladapter"
    "github.com/mark3labs/mcp-go/mcp"
)

// MCPAdapter converts between MCP and canonical tool formats
type MCPAdapter struct{}

// NewMCPAdapter creates a new MCP adapter
func NewMCPAdapter() *MCPAdapter {
    return &MCPAdapter{}
}

// Name returns the adapter identifier
func (a *MCPAdapter) Name() string {
    return "mcp"
}

// ToCanonical converts an MCP tool to canonical form
func (a *MCPAdapter) ToCanonical(raw any) (*tooladapter.CanonicalTool, error) {
    mcpTool, ok := raw.(*mcp.Tool)
    if !ok {
        return nil, fmt.Errorf("expected *mcp.Tool, got %T", raw)
    }

    inputSchema := convertMCPSchemaToJSONSchema(mcpTool.InputSchema)

    return &tooladapter.CanonicalTool{
        Namespace:    "mcp",
        Name:         mcpTool.Name,
        Description:  mcpTool.Description,
        InputSchema:  inputSchema,
        SourceFormat: "mcp",
        SourceMeta: map[string]any{
            "annotations": mcpTool.Annotations,
        },
    }, nil
}

// FromCanonical converts a canonical tool to MCP format
func (a *MCPAdapter) FromCanonical(tool *tooladapter.CanonicalTool) (any, error) {
    if tool == nil {
        return nil, fmt.Errorf("tool cannot be nil")
    }

    inputSchema := convertJSONSchemaToMCPSchema(tool.InputSchema)

    return &mcp.Tool{
        Name:        tool.Name,
        Description: tool.Description,
        InputSchema: inputSchema,
    }, nil
}

// SupportsFeature returns true for all features (MCP supports full JSON Schema)
func (a *MCPAdapter) SupportsFeature(feature tooladapter.SchemaFeature) bool {
    return true
}

// convertMCPSchemaToJSONSchema converts MCP schema to JSONSchema
func convertMCPSchemaToJSONSchema(s mcp.ToolInputSchema) *tooladapter.JSONSchema {
    schema := &tooladapter.JSONSchema{
        Type:     s.Type,
        Required: s.Required,
    }

    if s.Properties != nil {
        schema.Properties = make(map[string]*tooladapter.JSONSchema)
        for k, v := range s.Properties {
            schema.Properties[k] = convertMapToJSONSchema(v)
        }
    }

    return schema
}

// convertMapToJSONSchema converts a map[string]any to JSONSchema
func convertMapToJSONSchema(m any) *tooladapter.JSONSchema {
    if m == nil {
        return nil
    }

    mMap, ok := m.(map[string]any)
    if !ok {
        return nil
    }

    schema := &tooladapter.JSONSchema{}

    if t, ok := mMap["type"].(string); ok {
        schema.Type = t
    }
    if d, ok := mMap["description"].(string); ok {
        schema.Description = d
    }
    if p, ok := mMap["pattern"].(string); ok {
        schema.Pattern = p
    }
    if f, ok := mMap["format"].(string); ok {
        schema.Format = f
    }
    if min, ok := mMap["minimum"].(float64); ok {
        schema.Minimum = &min
    }
    if max, ok := mMap["maximum"].(float64); ok {
        schema.Maximum = &max
    }
    if e, ok := mMap["enum"].([]any); ok {
        schema.Enum = e
    }
    if d, ok := mMap["default"]; ok {
        schema.Default = d
    }

    if props, ok := mMap["properties"].(map[string]any); ok {
        schema.Properties = make(map[string]*tooladapter.JSONSchema)
        for k, v := range props {
            schema.Properties[k] = convertMapToJSONSchema(v)
        }
    }

    if items, ok := mMap["items"]; ok {
        schema.Items = convertMapToJSONSchema(items)
    }

    if req, ok := mMap["required"].([]any); ok {
        schema.Required = make([]string, len(req))
        for i, r := range req {
            schema.Required[i] = r.(string)
        }
    }

    return schema
}

// convertJSONSchemaToMCPSchema converts JSONSchema to MCP schema
func convertJSONSchemaToMCPSchema(s *tooladapter.JSONSchema) mcp.ToolInputSchema {
    if s == nil {
        return mcp.ToolInputSchema{
            Type: "object",
        }
    }

    schema := mcp.ToolInputSchema{
        Type:     s.Type,
        Required: s.Required,
    }

    if s.Properties != nil {
        schema.Properties = make(map[string]any)
        for k, v := range s.Properties {
            schema.Properties[k] = convertJSONSchemaToMap(v)
        }
    }

    return schema
}

// convertJSONSchemaToMap converts JSONSchema to map[string]any
func convertJSONSchemaToMap(s *tooladapter.JSONSchema) map[string]any {
    if s == nil {
        return nil
    }

    m := make(map[string]any)

    if s.Type != "" {
        m["type"] = s.Type
    }
    if s.Description != "" {
        m["description"] = s.Description
    }
    if s.Pattern != "" {
        m["pattern"] = s.Pattern
    }
    if s.Format != "" {
        m["format"] = s.Format
    }
    if s.Minimum != nil {
        m["minimum"] = *s.Minimum
    }
    if s.Maximum != nil {
        m["maximum"] = *s.Maximum
    }
    if len(s.Enum) > 0 {
        m["enum"] = s.Enum
    }
    if s.Default != nil {
        m["default"] = s.Default
    }
    if len(s.Required) > 0 {
        m["required"] = s.Required
    }

    if s.Properties != nil {
        props := make(map[string]any)
        for k, v := range s.Properties {
            props[k] = convertJSONSchemaToMap(v)
        }
        m["properties"] = props
    }

    if s.Items != nil {
        m["items"] = convertJSONSchemaToMap(s.Items)
    }

    return m
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add MCP adapter implementation

- Bidirectional MCP ↔ Canonical conversion
- Full JSON Schema support
- Round-trip conversion preserves all data
- Helper functions for schema conversion

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: OpenAI Adapter Implementation

**Files:**
- Create: `tooladapter/adapters/openai.go`
- Create: `tooladapter/adapters/openai_test.go`

**Step 1: Write failing tests**

```go
// adapters/openai_test.go
package adapters_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/tooladapter"
    "github.com/jrraymond/tooladapter/adapters"
)

func TestOpenAIAdapter_Name(t *testing.T) {
    adapter := adapters.NewOpenAIAdapter(false)
    assert.Equal(t, "openai", adapter.Name())
}

func TestOpenAIAdapter_ToCanonical(t *testing.T) {
    adapter := adapters.NewOpenAIAdapter(false)

    fn := &adapters.OpenAIFunction{
        Name:        "get_weather",
        Description: "Get current weather",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "location": map[string]any{
                    "type":        "string",
                    "description": "City name",
                },
            },
            "required": []any{"location"},
        },
    }

    canonical, err := adapter.ToCanonical(fn)
    require.NoError(t, err)

    assert.Equal(t, "get_weather", canonical.Name)
    assert.Equal(t, "openai", canonical.Namespace)
    assert.Equal(t, "Get current weather", canonical.Description)
    assert.Equal(t, "openai", canonical.SourceFormat)
}

func TestOpenAIAdapter_FromCanonical(t *testing.T) {
    adapter := adapters.NewOpenAIAdapter(false)

    canonical := &tooladapter.CanonicalTool{
        Name:        "my_function",
        Description: "A test function",
        InputSchema: &tooladapter.JSONSchema{
            Type: "object",
            Properties: map[string]*tooladapter.JSONSchema{
                "input": {Type: "string"},
            },
            Required: []string{"input"},
        },
    }

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    fn, ok := result.(*adapters.OpenAIFunction)
    require.True(t, ok)

    assert.Equal(t, "my_function", fn.Name)
    assert.Equal(t, "A test function", fn.Description)
}

func TestOpenAIAdapter_StrictMode(t *testing.T) {
    adapter := adapters.NewOpenAIAdapter(true) // strict mode

    canonical := &tooladapter.CanonicalTool{
        Name:        "strict_tool",
        Description: "A strict tool",
        InputSchema: &tooladapter.JSONSchema{
            Type: "object",
        },
    }

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    fn, ok := result.(*adapters.OpenAIFunction)
    require.True(t, ok)

    assert.True(t, fn.Strict)
}

func TestOpenAIAdapter_FeatureSupport(t *testing.T) {
    adapter := adapters.NewOpenAIAdapter(false)

    // OpenAI doesn't support $ref definitions
    assert.False(t, adapter.SupportsFeature(tooladapter.FeatureRefDefinitions))

    // OpenAI supports basic features
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureNestedObjects))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureArrays))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureEnums))
}

func TestOpenAIAdapter_StrictModePatternSupport(t *testing.T) {
    nonStrict := adapters.NewOpenAIAdapter(false)
    strict := adapters.NewOpenAIAdapter(true)

    // Pattern validation only in strict mode
    assert.False(t, nonStrict.SupportsFeature(tooladapter.FeaturePatternValidation))
    assert.True(t, strict.SupportsFeature(tooladapter.FeaturePatternValidation))
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// adapters/openai.go
package adapters

import (
    "fmt"

    "github.com/jrraymond/tooladapter"
)

// OpenAIFunction represents an OpenAI function definition
type OpenAIFunction struct {
    Name        string         `json:"name"`
    Description string         `json:"description,omitempty"`
    Parameters  map[string]any `json:"parameters"`
    Strict      bool           `json:"strict,omitempty"`
}

// OpenAIAdapter converts between OpenAI and canonical tool formats
type OpenAIAdapter struct {
    strictMode bool
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(strictMode bool) *OpenAIAdapter {
    return &OpenAIAdapter{
        strictMode: strictMode,
    }
}

// Name returns the adapter identifier
func (a *OpenAIAdapter) Name() string {
    return "openai"
}

// ToCanonical converts an OpenAI function to canonical form
func (a *OpenAIAdapter) ToCanonical(raw any) (*tooladapter.CanonicalTool, error) {
    fn, ok := raw.(*OpenAIFunction)
    if !ok {
        return nil, fmt.Errorf("expected *OpenAIFunction, got %T", raw)
    }

    inputSchema := convertOpenAIParametersToJSONSchema(fn.Parameters)

    return &tooladapter.CanonicalTool{
        Namespace:    "openai",
        Name:         fn.Name,
        Description:  fn.Description,
        InputSchema:  inputSchema,
        SourceFormat: "openai",
        SourceMeta: map[string]any{
            "strict": fn.Strict,
        },
    }, nil
}

// FromCanonical converts a canonical tool to OpenAI format
func (a *OpenAIAdapter) FromCanonical(tool *tooladapter.CanonicalTool) (any, error) {
    if tool == nil {
        return nil, fmt.Errorf("tool cannot be nil")
    }

    params := convertJSONSchemaToOpenAIParameters(tool.InputSchema, a.strictMode)

    return &OpenAIFunction{
        Name:        tool.Name,
        Description: tool.Description,
        Parameters:  params,
        Strict:      a.strictMode,
    }, nil
}

// SupportsFeature checks if the adapter supports a schema feature
func (a *OpenAIAdapter) SupportsFeature(feature tooladapter.SchemaFeature) bool {
    switch feature {
    case tooladapter.FeatureRefDefinitions:
        return false // OpenAI doesn't support $ref
    case tooladapter.FeaturePatternValidation:
        return a.strictMode // Only in strict mode
    case tooladapter.FeatureAnyOf, tooladapter.FeatureOneOf:
        return false // Limited support
    default:
        return true
    }
}

// convertOpenAIParametersToJSONSchema converts OpenAI parameters to JSONSchema
func convertOpenAIParametersToJSONSchema(params map[string]any) *tooladapter.JSONSchema {
    if params == nil {
        return &tooladapter.JSONSchema{Type: "object"}
    }

    schema := &tooladapter.JSONSchema{}

    if t, ok := params["type"].(string); ok {
        schema.Type = t
    }
    if d, ok := params["description"].(string); ok {
        schema.Description = d
    }

    if props, ok := params["properties"].(map[string]any); ok {
        schema.Properties = make(map[string]*tooladapter.JSONSchema)
        for k, v := range props {
            if vMap, ok := v.(map[string]any); ok {
                schema.Properties[k] = convertOpenAIParametersToJSONSchema(vMap)
            }
        }
    }

    if req, ok := params["required"].([]any); ok {
        schema.Required = make([]string, len(req))
        for i, r := range req {
            if s, ok := r.(string); ok {
                schema.Required[i] = s
            }
        }
    }

    if e, ok := params["enum"].([]any); ok {
        schema.Enum = e
    }

    return schema
}

// convertJSONSchemaToOpenAIParameters converts JSONSchema to OpenAI parameters
func convertJSONSchemaToOpenAIParameters(s *tooladapter.JSONSchema, strict bool) map[string]any {
    if s == nil {
        return map[string]any{"type": "object"}
    }

    params := make(map[string]any)

    if s.Type != "" {
        params["type"] = s.Type
    }
    if s.Description != "" {
        params["description"] = s.Description
    }
    if len(s.Required) > 0 {
        params["required"] = s.Required
    }
    if len(s.Enum) > 0 {
        params["enum"] = s.Enum
    }

    if s.Properties != nil {
        props := make(map[string]any)
        for k, v := range s.Properties {
            props[k] = convertJSONSchemaToOpenAIParameters(v, strict)
        }
        params["properties"] = props
    }

    // OpenAI strict mode requirements
    if strict {
        if s.Pattern != "" {
            params["pattern"] = s.Pattern
        }
        // additionalProperties must be false in strict mode
        params["additionalProperties"] = false
    }

    return params
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add OpenAI adapter implementation

- OpenAI function format conversion
- Strict mode support with additionalProperties: false
- Feature support detection for $ref, patterns, anyOf/oneOf
- Bidirectional conversion with schema preservation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Anthropic Adapter Implementation

**Files:**
- Create: `tooladapter/adapters/anthropic.go`
- Create: `tooladapter/adapters/anthropic_test.go`

**Step 1: Write failing tests**

```go
// adapters/anthropic_test.go
package adapters_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/tooladapter"
    "github.com/jrraymond/tooladapter/adapters"
)

func TestAnthropicAdapter_Name(t *testing.T) {
    adapter := adapters.NewAnthropicAdapter()
    assert.Equal(t, "anthropic", adapter.Name())
}

func TestAnthropicAdapter_ToCanonical(t *testing.T) {
    adapter := adapters.NewAnthropicAdapter()

    tool := &adapters.AnthropicTool{
        Name:        "search",
        Description: "Search for information",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "query": map[string]any{
                    "type": "string",
                },
            },
            "required": []any{"query"},
        },
    }

    canonical, err := adapter.ToCanonical(tool)
    require.NoError(t, err)

    assert.Equal(t, "search", canonical.Name)
    assert.Equal(t, "anthropic", canonical.Namespace)
    assert.Equal(t, "Search for information", canonical.Description)
    assert.Equal(t, "anthropic", canonical.SourceFormat)
}

func TestAnthropicAdapter_FromCanonical(t *testing.T) {
    adapter := adapters.NewAnthropicAdapter()

    canonical := &tooladapter.CanonicalTool{
        Name:        "my_tool",
        Description: "A test tool",
        InputSchema: &tooladapter.JSONSchema{
            Type: "object",
            Properties: map[string]*tooladapter.JSONSchema{
                "input": {Type: "string"},
            },
        },
    }

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    tool, ok := result.(*adapters.AnthropicTool)
    require.True(t, ok)

    assert.Equal(t, "my_tool", tool.Name)
    assert.Equal(t, "A test tool", tool.Description)

    // Verify input_schema field (Anthropic uses input_schema, not inputSchema)
    assert.NotNil(t, tool.InputSchema)
}

func TestAnthropicAdapter_FeatureSupport(t *testing.T) {
    adapter := adapters.NewAnthropicAdapter()

    // Anthropic supports most features except $ref
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureNestedObjects))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureArrays))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureEnums))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeaturePatternValidation))
    assert.True(t, adapter.SupportsFeature(tooladapter.FeatureNullable))

    // Anthropic doesn't support $ref
    assert.False(t, adapter.SupportsFeature(tooladapter.FeatureRefDefinitions))
}

func TestAnthropicAdapter_RoundTrip(t *testing.T) {
    adapter := adapters.NewAnthropicAdapter()

    original := &adapters.AnthropicTool{
        Name:        "roundtrip",
        Description: "Test round trip",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "value": map[string]any{
                    "type":        "integer",
                    "description": "A number",
                },
            },
            "required": []any{"value"},
        },
    }

    canonical, err := adapter.ToCanonical(original)
    require.NoError(t, err)

    result, err := adapter.FromCanonical(canonical)
    require.NoError(t, err)

    restored, ok := result.(*adapters.AnthropicTool)
    require.True(t, ok)

    assert.Equal(t, original.Name, restored.Name)
    assert.Equal(t, original.Description, restored.Description)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// adapters/anthropic.go
package adapters

import (
    "fmt"

    "github.com/jrraymond/tooladapter"
)

// AnthropicTool represents an Anthropic tool definition
type AnthropicTool struct {
    Name        string         `json:"name"`
    Description string         `json:"description,omitempty"`
    InputSchema map[string]any `json:"input_schema"` // Note: input_schema, not inputSchema
}

// AnthropicAdapter converts between Anthropic and canonical tool formats
type AnthropicAdapter struct{}

// NewAnthropicAdapter creates a new Anthropic adapter
func NewAnthropicAdapter() *AnthropicAdapter {
    return &AnthropicAdapter{}
}

// Name returns the adapter identifier
func (a *AnthropicAdapter) Name() string {
    return "anthropic"
}

// ToCanonical converts an Anthropic tool to canonical form
func (a *AnthropicAdapter) ToCanonical(raw any) (*tooladapter.CanonicalTool, error) {
    tool, ok := raw.(*AnthropicTool)
    if !ok {
        return nil, fmt.Errorf("expected *AnthropicTool, got %T", raw)
    }

    inputSchema := convertAnthropicSchemaToJSONSchema(tool.InputSchema)

    return &tooladapter.CanonicalTool{
        Namespace:    "anthropic",
        Name:         tool.Name,
        Description:  tool.Description,
        InputSchema:  inputSchema,
        SourceFormat: "anthropic",
    }, nil
}

// FromCanonical converts a canonical tool to Anthropic format
func (a *AnthropicAdapter) FromCanonical(tool *tooladapter.CanonicalTool) (any, error) {
    if tool == nil {
        return nil, fmt.Errorf("tool cannot be nil")
    }

    inputSchema := convertJSONSchemaToAnthropicSchema(tool.InputSchema)

    return &AnthropicTool{
        Name:        tool.Name,
        Description: tool.Description,
        InputSchema: inputSchema,
    }, nil
}

// SupportsFeature checks if the adapter supports a schema feature
func (a *AnthropicAdapter) SupportsFeature(feature tooladapter.SchemaFeature) bool {
    switch feature {
    case tooladapter.FeatureRefDefinitions:
        return false // Anthropic doesn't support $ref
    default:
        return true
    }
}

// convertAnthropicSchemaToJSONSchema converts Anthropic schema to JSONSchema
func convertAnthropicSchemaToJSONSchema(schema map[string]any) *tooladapter.JSONSchema {
    if schema == nil {
        return &tooladapter.JSONSchema{Type: "object"}
    }

    result := &tooladapter.JSONSchema{}

    if t, ok := schema["type"].(string); ok {
        result.Type = t
    }
    if d, ok := schema["description"].(string); ok {
        result.Description = d
    }
    if p, ok := schema["pattern"].(string); ok {
        result.Pattern = p
    }
    if f, ok := schema["format"].(string); ok {
        result.Format = f
    }
    if e, ok := schema["enum"].([]any); ok {
        result.Enum = e
    }
    if d, ok := schema["default"]; ok {
        result.Default = d
    }

    if props, ok := schema["properties"].(map[string]any); ok {
        result.Properties = make(map[string]*tooladapter.JSONSchema)
        for k, v := range props {
            if vMap, ok := v.(map[string]any); ok {
                result.Properties[k] = convertAnthropicSchemaToJSONSchema(vMap)
            }
        }
    }

    if items, ok := schema["items"].(map[string]any); ok {
        result.Items = convertAnthropicSchemaToJSONSchema(items)
    }

    if req, ok := schema["required"].([]any); ok {
        result.Required = make([]string, len(req))
        for i, r := range req {
            if s, ok := r.(string); ok {
                result.Required[i] = s
            }
        }
    }

    return result
}

// convertJSONSchemaToAnthropicSchema converts JSONSchema to Anthropic schema
func convertJSONSchemaToAnthropicSchema(s *tooladapter.JSONSchema) map[string]any {
    if s == nil {
        return map[string]any{"type": "object"}
    }

    schema := make(map[string]any)

    if s.Type != "" {
        schema["type"] = s.Type
    }
    if s.Description != "" {
        schema["description"] = s.Description
    }
    if s.Pattern != "" {
        schema["pattern"] = s.Pattern
    }
    if s.Format != "" {
        schema["format"] = s.Format
    }
    if len(s.Enum) > 0 {
        schema["enum"] = s.Enum
    }
    if s.Default != nil {
        schema["default"] = s.Default
    }
    if len(s.Required) > 0 {
        schema["required"] = s.Required
    }

    if s.Properties != nil {
        props := make(map[string]any)
        for k, v := range s.Properties {
            props[k] = convertJSONSchemaToAnthropicSchema(v)
        }
        schema["properties"] = props
    }

    if s.Items != nil {
        schema["items"] = convertJSONSchemaToAnthropicSchema(s.Items)
    }

    return schema
}
```

**Step 4: Run tests to verify they pass**

Run: `cd tooladapter && go test ./adapters/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tooladapter/
git commit -m "$(cat <<'EOF'
feat(tooladapter): add Anthropic adapter implementation

- Anthropic tool format with input_schema field naming
- Bidirectional conversion with schema preservation
- Feature support detection ($ref not supported)
- Round-trip conversion tests

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-008 complete:

- [ ] All tests pass: `go test ./... -v`
- [ ] Code coverage > 80%: `go test ./... -cover`
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete:
  - [ ] Package documentation in `doc.go`
  - [ ] README.md with usage examples
  - [ ] GoDoc comments on all exported types
- [ ] Integration verified:
  - [ ] MCP adapter round-trip works
  - [ ] OpenAI adapter handles strict mode
  - [ ] Anthropic adapter handles input_schema naming
  - [ ] AdapterRegistry manages all adapters

---

## Definition of Done

1. **CanonicalTool** type with all fields from proposal
2. **JSONSchema** type with DeepCopy and ToMap methods
3. **Adapter** interface with ToCanonical, FromCanonical, SupportsFeature
4. **AdapterRegistry** with Register, Get, List, Convert
5. **MCPAdapter** with full JSON Schema support
6. **OpenAIAdapter** with strict mode support
7. **AnthropicAdapter** with input_schema field naming
8. All tests passing with >80% coverage
9. Documentation complete
