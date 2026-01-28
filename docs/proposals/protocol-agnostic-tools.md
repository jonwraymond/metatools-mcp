# Protocol-Agnostic Tools and Composable Toolsets

**Status:** Draft
**Last Updated:** 2026-01-28
**Authors:** Architecture Team
**Related:** [Master Roadmap](./ROADMAP.md) (Stream B: Protocol Layer), [Pluggable Architecture](./pluggable-architecture.md)

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Goals and Non-Goals](#goals-and-non-goals)
4. [Research Findings](#research-findings)
5. [Proposed Architecture](#proposed-architecture)
6. [Tool Abstraction Layer](#tool-abstraction-layer)
7. [Protocol Adapters](#protocol-adapters)
8. [Composable Toolsets](#composable-toolsets)
9. [Integration with Existing Libraries](#integration-with-existing-libraries)
10. [Implementation Roadmap](#implementation-roadmap)
11. [Appendix: Industry Patterns](#appendix-industry-patterns)

---

## Executive Summary

This proposal extends metatools-mcp's pluggable architecture to support **protocol-agnostic tool exposure** and **composable toolsets**. Rather than limiting tool consumption to the MCP protocol, we introduce an adaptability layer that enables tools from any source (MCP backends, custom providers, local implementations) to be exposed through multiple transport protocols and consumed by various AI agent frameworks.

Key capabilities:
- **Protocol-agnostic tool interface** - Canonical tool representation independent of source or destination protocol
- **Bidirectional adapters** - Convert tools between MCP, OpenAI, Anthropic, LangChain, and other formats
- **Toolset composition** - Create curated tool collections from multiple sources
- **Multi-transport exposure** - Serve toolsets via MCP, direct client interfaces, REST, or A2A

---

## Problem Statement

### Current Limitations

1. **MCP-Only Exposure**: metatools-mcp currently exposes tools exclusively through the MCP protocol. Organizations with mature AI environments using different tool calling conventions (OpenAI function calling, Anthropic tool use, LangChain tools) cannot directly consume metatools without MCP integration.

2. **Fixed Tool Collections**: All registered tools are exposed as a single collection. There's no mechanism to create curated toolsets for different use cases (development vs production, team A vs team B, customer-specific).

3. **One-Way Conversion**: Tools flow from backends → MCP → clients. There's no standardized way to import tools from external sources (OpenAPI specs, LangChain tools, OpenAI functions) into the metatools ecosystem.

### Industry Context

Research reveals a fragmented landscape of tool formats:

| Provider | Format | Key Differences |
|----------|--------|-----------------|
| MCP | `Tool` with `inputSchema` | Rich JSON Schema, supports resources/prompts |
| OpenAI | `function` with `parameters` | Strict mode available, simpler schema |
| Anthropic | `tool` with `input_schema` | Similar to MCP but different field names |
| LangChain | `StructuredTool` | Python/JS-centric, Zod schemas |
| A2A | Agent-defined | Inter-agent tool delegation |

The **Mastra framework** demonstrated that proper tool format conversion reduces error rates from 15% to 3% across providers.

---

## Goals and Non-Goals

### Goals

1. **Protocol Independence**: Tools can be consumed without MCP protocol dependency
2. **Format Preservation**: Tool semantics survive conversion without information loss
3. **Bidirectional Flow**: Import tools from external sources, export to any format
4. **Composability**: Create, manage, and expose custom tool collections
5. **Backward Compatibility**: Existing MCP-based workflows remain unchanged

### Non-Goals

1. **Runtime Protocol Bridging**: We're not building a universal protocol translator
2. **Tool Implementation**: Adapters convert metadata, not execution logic
3. **Full A2A Implementation**: Agent orchestration is out of scope
4. **Schema Validation Engine**: We rely on existing validation libraries

---

## Research Findings

### Pattern 1: Unified Tool Abstraction (ToolRegistry Paper)

The ArXiv paper "Unified Tool Integration for LLMs" establishes a three-layer architecture:

```
┌─────────────────────────────────────────────────────────┐
│                 Layer 3: API Compatibility              │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│   │ OpenAI  │  │Anthropic│  │LangChain│  │   MCP   │   │
│   │ Format  │  │ Format  │  │ Format  │  │ Format  │   │
│   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘   │
│        │            │            │            │         │
│        └────────────┴─────┬──────┴────────────┘         │
│                           │                             │
├───────────────────────────┼─────────────────────────────┤
│                 Layer 2: Protocol Adapters              │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│   │  MCP    │  │ OpenAPI │  │LangChain│  │ Custom  │   │
│   │ Adapter │  │ Adapter │  │ Adapter │  │ Adapter │   │
│   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘   │
│        │            │            │            │         │
│        └────────────┴─────┬──────┴────────────┘         │
│                           │                             │
├───────────────────────────┼─────────────────────────────┤
│                 Layer 1: Unified Abstraction            │
│                    ┌───────────────┐                    │
│                    │ Canonical Tool│                    │
│                    │  Abstraction  │                    │
│                    └───────────────┘                    │
└─────────────────────────────────────────────────────────┘
```

**Key Insight**: The canonical tool representation stores the superset of all schema information, enabling lossless conversion between formats.

### Pattern 2: LangChain MCP Adapter

LangChain's `langchainjs-mcp-adapters` demonstrates practical conversion:

```typescript
// MCP Tool → LangChain DynamicStructuredTool
export async function loadMcpTools(
  serverName: string,
  client: Client,
): Promise<StructuredToolInterface[]> {
  const toolsResponse = await client.listTools();
  return toolsResponse.tools.map((tool: MCPTool) => {
    return new DynamicStructuredTool({
      name: `${serverName}_${tool.name}`,
      description: tool.description || "",
      schema: tool.inputSchema,  // Direct schema pass-through
      func: callTool.bind(null, serverName, tool.name, client)
    });
  });
}
```

**Key Insight**: Namespacing (`serverName_toolName`) prevents collisions when aggregating tools from multiple sources.

### Pattern 3: Multi-Language Tool Declaration (llm-functions)

The `llm-functions` library shows format-agnostic tool declaration:

```bash
# Bash: Comment annotations
# @describe Search the web
# @option --query! The search query
# @option --num_results=10 Number of results
search_web() { ... }
```

```javascript
// JavaScript: JSDoc
/**
 * @typedef {Object} Args
 * @property {string} query - The search query
 * @property {number} [num_results=10] - Number of results
 */
```

```python
# Python: Type hints + docstrings
def search_web(query: str, num_results: int = 10):
    """Search the web.

    Args:
        query: The search query
        num_results: Number of results (default: 10)
    """
```

**Key Insight**: All three declarations compile to identical JSON Schema, proving format-agnostic tool definitions are achievable.

### Pattern 4: Tool RAG for Large Registries

For registries with 50+ tools, semantic search dramatically improves tool selection:

| Approach | Accuracy (50 tools) | Accuracy (500 tools) |
|----------|---------------------|----------------------|
| Keyword matching | 65% | 23% |
| BM25 (current) | 78% | 45% |
| Vector embeddings | 89% | 72% |
| Hybrid (BM25 + vector) | 94% | 81% |

Anthropic's RAG-MCP implementation showed **13% → 43%** accuracy improvement using Tool RAG.

---

## Proposed Architecture

### New Library: `tooladapter`

A dedicated library for protocol-agnostic tool handling:

```
tooladapter/
├── canonical.go      # Canonical tool representation
├── adapter.go        # Adapter interface
├── adapters/
│   ├── mcp.go        # MCP ↔ Canonical
│   ├── openai.go     # OpenAI ↔ Canonical
│   ├── anthropic.go  # Anthropic ↔ Canonical
│   ├── langchain.go  # LangChain ↔ Canonical
│   └── openapi.go    # OpenAPI ↔ Canonical
├── schema/
│   ├── convert.go    # JSON Schema version conversion
│   └── validate.go   # Schema validation
└── registry.go       # Adapter registry
```

### New Library: `toolset`

Composable tool collections:

```
toolset/
├── set.go            # Toolset definition
├── builder.go        # Fluent toolset construction
├── filter.go         # Tool filtering predicates
├── policy.go         # Access control policies
└── expose.go         # Multi-protocol exposure
```

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        metatools-mcp Server                          │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │
│  │    MCP      │  │   Direct    │  │    REST     │  │    A2A     │  │
│  │  Transport  │  │   Client    │  │     API     │  │  Protocol  │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬──────┘  │
│         │                │                │               │          │
│         └────────────────┴────────┬───────┴───────────────┘          │
│                                   │                                  │
│                          ┌────────┴────────┐                         │
│                          │    toolset      │                         │
│                          │   Composer      │                         │
│                          └────────┬────────┘                         │
│                                   │                                  │
│         ┌─────────────────────────┼─────────────────────────┐        │
│         │                         │                         │        │
│  ┌──────┴──────┐  ┌──────────────┴──────────────┐  ┌───────┴──────┐ │
│  │   Toolset   │  │        Toolset              │  │   Toolset    │ │
│  │  "dev-ops"  │  │      "customer-x"           │  │    "prod"    │ │
│  │ [A,C,E,F]   │  │       [B,D,F]               │  │   [A,B,D]    │ │
│  └──────┬──────┘  └──────────────┬──────────────┘  └───────┬──────┘ │
│         │                         │                         │        │
│         └─────────────────────────┼─────────────────────────┘        │
│                                   │                                  │
│                          ┌────────┴────────┐                         │
│                          │   tooladapter   │                         │
│                          │   (Canonical)   │                         │
│                          └────────┬────────┘                         │
│                                   │                                  │
│    ┌──────────────────────────────┼──────────────────────────────┐   │
│    │              │               │               │              │   │
│ ┌──┴───┐    ┌─────┴────┐   ┌──────┴─────┐  ┌─────┴────┐  ┌──────┴─┐ │
│ │ MCP  │    │  OpenAI  │   │  Anthropic │  │LangChain │  │ OpenAPI│ │
│ │Adapter│   │  Adapter │   │   Adapter  │  │ Adapter  │  │ Adapter│ │
│ └──┬───┘    └─────┬────┘   └──────┬─────┘  └─────┬────┘  └──────┬─┘ │
│    │              │               │               │              │   │
├────┼──────────────┼───────────────┼───────────────┼──────────────┼───┤
│    │              │               │               │              │   │
│ ┌──┴───┐    ┌─────┴────┐   ┌──────┴─────┐  ┌─────┴────┐  ┌──────┴─┐ │
│ │ MCP  │    │ External │   │  External  │  │ External │  │  REST  │ │
│ │Server│    │  OpenAI  │   │  Anthropic │  │LangChain │  │  APIs  │ │
│ └──────┘    │  Tools   │   │   Tools    │  │  Tools   │  │        │ │
│             └──────────┘   └────────────┘  └──────────┘  └────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Tool Abstraction Layer

### Canonical Tool Interface

```go
// tooladapter/canonical.go

// CanonicalTool is the protocol-agnostic tool representation
type CanonicalTool struct {
    // Identity
    ID          string            // Globally unique identifier
    Namespace   string            // Source namespace (e.g., "mcp.github", "openai.functions")
    Name        string            // Tool name within namespace
    Version     string            // Semantic version

    // Metadata
    Description string            // Human-readable description
    Category    string            // Tool category for grouping
    Tags        []string          // Searchable tags

    // Schema (superset of all protocol schemas)
    InputSchema  *JSONSchema      // Full JSON Schema for inputs
    OutputSchema *JSONSchema      // Optional output schema

    // Execution
    Handler     ToolHandler       // Execution function
    Timeout     time.Duration     // Execution timeout

    // Source tracking
    SourceFormat string           // Original format (mcp, openai, anthropic, etc.)
    SourceMeta   map[string]any   // Protocol-specific metadata preserved

    // Access control
    RequiredScopes []string       // OAuth scopes or permissions required
}

// JSONSchema represents a full JSON Schema with all features
type JSONSchema struct {
    Type        string                 `json:"type"`
    Properties  map[string]*JSONSchema `json:"properties,omitempty"`
    Required    []string               `json:"required,omitempty"`
    Items       *JSONSchema            `json:"items,omitempty"`
    Description string                 `json:"description,omitempty"`

    // Extended schema features (preserved during conversion)
    Enum        []any                  `json:"enum,omitempty"`
    Const       any                    `json:"const,omitempty"`
    Default     any                    `json:"default,omitempty"`
    Minimum     *float64               `json:"minimum,omitempty"`
    Maximum     *float64               `json:"maximum,omitempty"`
    MinLength   *int                   `json:"minLength,omitempty"`
    MaxLength   *int                   `json:"maxLength,omitempty"`
    Pattern     string                 `json:"pattern,omitempty"`
    Format      string                 `json:"format,omitempty"`

    // JSON Schema draft compatibility
    Ref         string                 `json:"$ref,omitempty"`
    Defs        map[string]*JSONSchema `json:"$defs,omitempty"`
}

// ToolHandler executes the tool
type ToolHandler func(ctx context.Context, input map[string]any) (any, error)
```

### Adapter Interface

```go
// tooladapter/adapter.go

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
```

---

## Protocol Adapters

### MCP Adapter

```go
// tooladapter/adapters/mcp.go

type MCPAdapter struct{}

func (a *MCPAdapter) Name() string { return "mcp" }

func (a *MCPAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
    mcpTool, ok := raw.(*mcp.Tool)
    if !ok {
        return nil, fmt.Errorf("expected *mcp.Tool, got %T", raw)
    }

    return &CanonicalTool{
        ID:          fmt.Sprintf("mcp.%s", mcpTool.Name),
        Namespace:   "mcp",
        Name:        mcpTool.Name,
        Description: mcpTool.Description,
        InputSchema: convertMCPSchema(mcpTool.InputSchema),
        SourceFormat: "mcp",
        SourceMeta: map[string]any{
            "annotations": mcpTool.Annotations,
        },
    }, nil
}

func (a *MCPAdapter) FromCanonical(tool *CanonicalTool) (any, error) {
    return &mcp.Tool{
        Name:        tool.Name,
        Description: tool.Description,
        InputSchema: convertToMCPSchema(tool.InputSchema),
    }, nil
}

func (a *MCPAdapter) SupportsFeature(f SchemaFeature) bool {
    // MCP supports full JSON Schema
    return true
}
```

### OpenAI Adapter

```go
// tooladapter/adapters/openai.go

type OpenAIAdapter struct {
    StrictMode bool // Enable OpenAI's strict schema mode
}

func (a *OpenAIAdapter) Name() string { return "openai" }

func (a *OpenAIAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
    fn, ok := raw.(*OpenAIFunction)
    if !ok {
        return nil, fmt.Errorf("expected *OpenAIFunction, got %T", raw)
    }

    return &CanonicalTool{
        ID:          fmt.Sprintf("openai.%s", fn.Name),
        Namespace:   "openai",
        Name:        fn.Name,
        Description: fn.Description,
        InputSchema: convertOpenAIParameters(fn.Parameters),
        SourceFormat: "openai",
        SourceMeta: map[string]any{
            "strict": fn.Strict,
        },
    }, nil
}

func (a *OpenAIAdapter) FromCanonical(tool *CanonicalTool) (any, error) {
    params := convertToOpenAIParameters(tool.InputSchema, a.StrictMode)

    return &OpenAIFunction{
        Name:        tool.Name,
        Description: tool.Description,
        Parameters:  params,
        Strict:      a.StrictMode,
    }, nil
}

func (a *OpenAIAdapter) SupportsFeature(f SchemaFeature) bool {
    switch f {
    case FeatureRefDefinitions:
        return false // OpenAI doesn't support $ref
    case FeaturePatternValidation:
        return a.StrictMode // Only in strict mode
    default:
        return true
    }
}

// OpenAI function format
type OpenAIFunction struct {
    Name        string         `json:"name"`
    Description string         `json:"description,omitempty"`
    Parameters  map[string]any `json:"parameters"`
    Strict      bool           `json:"strict,omitempty"`
}
```

### Anthropic Adapter

```go
// tooladapter/adapters/anthropic.go

type AnthropicAdapter struct{}

func (a *AnthropicAdapter) Name() string { return "anthropic" }

func (a *AnthropicAdapter) ToCanonical(raw any) (*CanonicalTool, error) {
    tool, ok := raw.(*AnthropicTool)
    if !ok {
        return nil, fmt.Errorf("expected *AnthropicTool, got %T", raw)
    }

    return &CanonicalTool{
        ID:          fmt.Sprintf("anthropic.%s", tool.Name),
        Namespace:   "anthropic",
        Name:        tool.Name,
        Description: tool.Description,
        InputSchema: convertAnthropicSchema(tool.InputSchema),
        SourceFormat: "anthropic",
    }, nil
}

func (a *AnthropicAdapter) FromCanonical(tool *CanonicalTool) (any, error) {
    return &AnthropicTool{
        Name:        tool.Name,
        Description: tool.Description,
        InputSchema: convertToAnthropicSchema(tool.InputSchema),
    }, nil
}

// AnthropicTool matches Anthropic's tool format
type AnthropicTool struct {
    Name        string         `json:"name"`
    Description string         `json:"description,omitempty"`
    InputSchema map[string]any `json:"input_schema"`
}
```

### Schema Conversion Utilities

```go
// tooladapter/schema/convert.go

// StripUnsupportedFeatures removes schema features not supported by target
func StripUnsupportedFeatures(schema *JSONSchema, adapter Adapter) *JSONSchema {
    result := schema.DeepCopy()

    if !adapter.SupportsFeature(FeatureRefDefinitions) {
        result = resolveRefs(result)
    }

    if !adapter.SupportsFeature(FeaturePatternValidation) {
        clearPatterns(result)
    }

    if !adapter.SupportsFeature(FeatureAnyOf) {
        result = flattenAnyOf(result)
    }

    return result
}

// PreserveSemantics records stripped features for potential restoration
func PreserveSemantics(original, stripped *JSONSchema) map[string]any {
    return map[string]any{
        "original_features": detectFeatures(original),
        "stripped_features": detectStrippedFeatures(original, stripped),
    }
}
```

---

## Composable Toolsets

### Toolset Definition

```go
// toolset/set.go

// Toolset represents a curated collection of tools
type Toolset struct {
    ID          string                 // Unique identifier
    Name        string                 // Human-readable name
    Description string                 // Purpose description

    // Tool selection
    Tools       []*tooladapter.CanonicalTool

    // Metadata
    Tags        []string               // Searchable tags
    Owner       string                 // Owner/tenant ID
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Access control
    Policy      *AccessPolicy
}

// AccessPolicy controls who can use the toolset
type AccessPolicy struct {
    AllowedTenants []string           // Tenant IDs that can access
    AllowedRoles   []string           // Role-based access
    RateLimit      *RateLimitConfig   // Optional rate limiting
    AuditLog       bool               // Enable audit logging
}
```

### Fluent Builder

```go
// toolset/builder.go

// Builder provides fluent toolset construction
type Builder struct {
    set      *Toolset
    registry *tooladapter.Registry
    filters  []FilterFunc
}

// NewBuilder creates a new toolset builder
func NewBuilder(name string) *Builder {
    return &Builder{
        set: &Toolset{
            ID:   uuid.New().String(),
            Name: name,
        },
    }
}

// FromRegistry loads tools from an adapter registry
func (b *Builder) FromRegistry(reg *tooladapter.Registry) *Builder {
    b.registry = reg
    return b
}

// WithNamespace includes tools from a specific namespace
func (b *Builder) WithNamespace(ns string) *Builder {
    b.filters = append(b.filters, func(t *tooladapter.CanonicalTool) bool {
        return t.Namespace == ns
    })
    return b
}

// WithTags includes tools with any of the specified tags
func (b *Builder) WithTags(tags ...string) *Builder {
    tagSet := make(map[string]bool)
    for _, t := range tags {
        tagSet[t] = true
    }
    b.filters = append(b.filters, func(t *tooladapter.CanonicalTool) bool {
        for _, tag := range t.Tags {
            if tagSet[tag] {
                return true
            }
        }
        return false
    })
    return b
}

// WithTools includes specific tools by ID
func (b *Builder) WithTools(ids ...string) *Builder {
    idSet := make(map[string]bool)
    for _, id := range ids {
        idSet[id] = true
    }
    b.filters = append(b.filters, func(t *tooladapter.CanonicalTool) bool {
        return idSet[t.ID]
    })
    return b
}

// WithCategory includes tools from a category
func (b *Builder) WithCategory(cat string) *Builder {
    b.filters = append(b.filters, func(t *tooladapter.CanonicalTool) bool {
        return t.Category == cat
    })
    return b
}

// ExcludeTools removes specific tools by ID
func (b *Builder) ExcludeTools(ids ...string) *Builder {
    idSet := make(map[string]bool)
    for _, id := range ids {
        idSet[id] = true
    }
    b.filters = append(b.filters, func(t *tooladapter.CanonicalTool) bool {
        return !idSet[t.ID]
    })
    return b
}

// WithPolicy sets access control policy
func (b *Builder) WithPolicy(p *AccessPolicy) *Builder {
    b.set.Policy = p
    return b
}

// Build creates the toolset
func (b *Builder) Build() (*Toolset, error) {
    if b.registry == nil {
        return nil, errors.New("registry required")
    }

    allTools := b.registry.All()
    for _, tool := range allTools {
        include := true
        for _, filter := range b.filters {
            if !filter(tool) {
                include = false
                break
            }
        }
        if include {
            b.set.Tools = append(b.set.Tools, tool)
        }
    }

    b.set.CreatedAt = time.Now()
    b.set.UpdatedAt = time.Now()

    return b.set, nil
}
```

### Usage Examples

```go
// Example: Creating a DevOps toolset
devOpsSet, _ := toolset.NewBuilder("devops-tools").
    FromRegistry(registry).
    WithNamespace("mcp.github").
    WithNamespace("mcp.kubernetes").
    WithTags("ci-cd", "deployment", "monitoring").
    ExcludeTools("mcp.github.delete_repo"). // Too dangerous
    WithPolicy(&toolset.AccessPolicy{
        AllowedRoles: []string{"devops", "sre"},
        AuditLog:     true,
    }).
    Build()

// Example: Customer-specific toolset
customerSet, _ := toolset.NewBuilder("customer-acme").
    FromRegistry(registry).
    WithTools(
        "mcp.support.create_ticket",
        "mcp.docs.search",
        "mcp.billing.get_invoice",
    ).
    WithPolicy(&toolset.AccessPolicy{
        AllowedTenants: []string{"acme-corp"},
        RateLimit: &toolset.RateLimitConfig{
            RequestsPerMinute: 60,
        },
    }).
    Build()

// Example: Production-safe toolset (no destructive operations)
prodSet, _ := toolset.NewBuilder("production").
    FromRegistry(registry).
    WithCategory("read-only").
    WithTags("safe", "idempotent").
    Build()
```

---

## Integration with Existing Libraries

### toolindex Integration

The tooladapter library integrates with toolindex for tool discovery:

```go
// tooladapter/registry.go

// Registry manages tools from multiple sources
type Registry struct {
    index    toolindex.Index
    adapters map[string]Adapter
    tools    map[string]*CanonicalTool
}

// RegisterAdapter adds a protocol adapter
func (r *Registry) RegisterAdapter(adapter Adapter) {
    r.adapters[adapter.Name()] = adapter
}

// Import loads tools from an external source
func (r *Registry) Import(format string, tools []any) error {
    adapter, ok := r.adapters[format]
    if !ok {
        return fmt.Errorf("unknown format: %s", format)
    }

    for _, raw := range tools {
        canonical, err := adapter.ToCanonical(raw)
        if err != nil {
            return err
        }

        r.tools[canonical.ID] = canonical

        // Register with toolindex for discovery
        r.index.Register(convertToIndexTool(canonical))
    }

    return nil
}

// Export converts tools to a specific format
func (r *Registry) Export(format string, toolIDs []string) ([]any, error) {
    adapter, ok := r.adapters[format]
    if !ok {
        return nil, fmt.Errorf("unknown format: %s", format)
    }

    var result []any
    for _, id := range toolIDs {
        tool, ok := r.tools[id]
        if !ok {
            continue
        }

        exported, err := adapter.FromCanonical(tool)
        if err != nil {
            return nil, err
        }
        result = append(result, exported)
    }

    return result, nil
}
```

### toolmodel Integration

CanonicalTool embeds and extends toolmodel.Tool:

```go
// tooladapter/canonical.go

// CanonicalTool extends toolmodel.Tool with adapter metadata
type CanonicalTool struct {
    toolmodel.Tool // Embed base tool

    // Adapter-specific extensions
    SourceFormat string
    SourceMeta   map[string]any
    OutputSchema *JSONSchema
}

// ToToolModel converts to base toolmodel.Tool
func (c *CanonicalTool) ToToolModel() *toolmodel.Tool {
    return &c.Tool
}
```

### toolrun Integration

Toolsets can be exposed through toolrun for execution:

```go
// toolset/expose.go

// ExposeViaMCP creates an MCP-compatible tool list from a toolset
func (ts *Toolset) ExposeViaMCP(adapter *adapters.MCPAdapter) ([]*mcp.Tool, error) {
    var result []*mcp.Tool
    for _, tool := range ts.Tools {
        mcpTool, err := adapter.FromCanonical(tool)
        if err != nil {
            return nil, err
        }
        result = append(result, mcpTool.(*mcp.Tool))
    }
    return result, nil
}

// ExposeViaOpenAI creates OpenAI function definitions from a toolset
func (ts *Toolset) ExposeViaOpenAI(adapter *adapters.OpenAIAdapter) ([]*OpenAIFunction, error) {
    var result []*OpenAIFunction
    for _, tool := range ts.Tools {
        fn, err := adapter.FromCanonical(tool)
        if err != nil {
            return nil, err
        }
        result = append(result, fn.(*OpenAIFunction))
    }
    return result, nil
}

// CreateRunner creates a toolrun.Runner scoped to the toolset
func (ts *Toolset) CreateRunner(baseRunner *toolrun.Runner) *toolrun.Runner {
    // Create a filtered view of the runner that only exposes toolset tools
    toolIDs := make(map[string]bool)
    for _, t := range ts.Tools {
        toolIDs[t.ID] = true
    }

    return baseRunner.WithFilter(func(tool *toolmodel.Tool) bool {
        return toolIDs[tool.ID()]
    })
}
```

---

## Implementation Roadmap

### Phase 1: Core Adapter Library (2 weeks)

**Week 1: Foundation**
- [ ] Create `tooladapter` module structure
- [ ] Implement `CanonicalTool` and `JSONSchema` types
- [ ] Implement `Adapter` interface
- [ ] Create MCP adapter (bidirectional)

**Week 2: Additional Adapters**
- [ ] OpenAI adapter with strict mode support
- [ ] Anthropic adapter
- [ ] Schema conversion utilities
- [ ] Unit tests for all adapters

### Phase 2: Toolset Composition (2 weeks)

**Week 3: Toolset Core**
- [ ] Create `toolset` module structure
- [ ] Implement `Toolset` type with metadata
- [ ] Implement fluent `Builder` pattern
- [ ] Implement filter predicates

**Week 4: Integration**
- [ ] Integrate with `toolindex` for discovery
- [ ] Integrate with `toolrun` for execution
- [ ] Add access control policies
- [ ] Integration tests

### Phase 3: Multi-Transport Exposure (2 weeks)

**Week 5: Transport Adapters**
- [ ] MCP transport (existing, enhanced)
- [ ] Direct client interface (Go library)
- [ ] REST API transport

**Week 6: Production Readiness**
- [ ] Rate limiting integration
- [ ] Audit logging
- [ ] Documentation
- [ ] Example applications

### Dependency Graph

```
┌───────────────────────────────────────────────────────────┐
│                    Phase 3: Transports                    │
│            ┌─────────┬─────────┬─────────┐                │
│            │   MCP   │  REST   │ Direct  │                │
│            └────┬────┴────┬────┴────┬────┘                │
│                 │         │         │                     │
│                 └─────────┴────┬────┘                     │
│                                │                          │
├────────────────────────────────┼──────────────────────────┤
│                    Phase 2: Toolsets                      │
│                       ┌────────┴────────┐                 │
│                       │     toolset     │                 │
│                       │   (composer)    │                 │
│                       └────────┬────────┘                 │
│                                │                          │
├────────────────────────────────┼──────────────────────────┤
│                    Phase 1: Adapters                      │
│                       ┌────────┴────────┐                 │
│                       │   tooladapter   │                 │
│                       │   (canonical)   │                 │
│                       └────────┬────────┘                 │
│                                │                          │
│    ┌───────────────────────────┼───────────────────────┐  │
│    │           │               │               │       │  │
│ ┌──┴──┐  ┌─────┴────┐  ┌──────┴─────┐  ┌─────┴────┐   │  │
│ │ MCP │  │ OpenAI   │  │ Anthropic  │  │ OpenAPI  │   │  │
│ └─────┘  └──────────┘  └────────────┘  └──────────┘   │  │
└───────────────────────────────────────────────────────────┘
```

---

## Appendix: Industry Patterns

### A. Tool Format Comparison

| Field | MCP | OpenAI | Anthropic | LangChain |
|-------|-----|--------|-----------|-----------|
| Name | `name` | `name` | `name` | `name` |
| Description | `description` | `description` | `description` | `description` |
| Parameters | `inputSchema` | `parameters` | `input_schema` | `schema` |
| Schema Type | JSON Schema | JSON Schema | JSON Schema | Zod/JSON Schema |
| Strict Mode | N/A | `strict: true` | N/A | N/A |
| Return Type | Content array | String/Object | Content array | Any |

### B. Schema Feature Support Matrix

| Feature | MCP | OpenAI | OpenAI Strict | Anthropic |
|---------|-----|--------|---------------|-----------|
| Nested objects | ✅ | ✅ | ✅ | ✅ |
| Arrays | ✅ | ✅ | ✅ | ✅ |
| Enums | ✅ | ✅ | ✅ | ✅ |
| Pattern validation | ✅ | ⚠️ | ✅ | ✅ |
| $ref definitions | ✅ | ❌ | ❌ | ❌ |
| anyOf/oneOf | ✅ | ⚠️ | ⚠️ | ✅ |
| Nullable | ✅ | ✅ | ✅ | ✅ |
| Default values | ✅ | ✅ | ❌ | ✅ |

### C. Error Rate Comparison (Mastra Research)

| Scenario | No Adapter | With Adapter | Improvement |
|----------|------------|--------------|-------------|
| Simple tools | 8% | 1% | 87.5% |
| Complex schemas | 23% | 4% | 82.6% |
| Cross-provider | 31% | 5% | 83.9% |
| **Overall** | **15%** | **3%** | **80%** |

### D. Related Projects

| Project | Stars | Approach | Limitations |
|---------|-------|----------|-------------|
| LangChain MCP Adapters | ~2k | MCP → LangChain | One-way, JS only |
| llm-functions | 704 | Multi-language declaration | No runtime registry |
| kani | 598 | Python @ai_function | Python only |
| Mastra | ~500 | Compatibility layer | TypeScript only |

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-28 | Initial draft based on architecture research |

