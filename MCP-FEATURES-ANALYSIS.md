# MCP Protocol Features Analysis

**Date:** 2026-01-30
**Protocol Revision:** 2025-11-25
**Go SDK Version:** v1.2.0

---

## Executive Summary

The Model Context Protocol (MCP) provides **6 core feature categories**. Currently, metatools-mcp implements only **Tools**. This analysis maps all MCP features to potential library implementations.

| Feature | Control | Current Status | Library Opportunity |
|---------|---------|----------------|---------------------|
| **Tools** | Model-controlled | ✅ Implemented | Already done |
| **Resources** | Application-controlled | ❌ Not implemented | **toolresource** |
| **Prompts** | User-controlled | ❌ Not implemented | **toolprompt** |
| **Sampling** | Server-initiated | ❌ Not implemented | **toolsampling** |
| **Roots** | Client-provided | ❌ Not implemented | Part of client |
| **Elicitation** | Server-initiated | ❌ Not implemented | **toolelicit** |

---

## MCP Feature Deep Dive

### 1. Tools (✅ IMPLEMENTED)

**Control:** Model-controlled - LLM decides when to invoke

**What it does:** Exposes executable functions that AI models can call to perform actions or retrieve information.

**JSON-RPC Methods:**
- `tools/list` - List available tools (paginated)
- `tools/call` - Invoke a tool

**Current Implementation:**
- `internal/provider/` - ToolProvider interface
- `internal/handlers/` - search_tools, describe_tool, run_tool
- Middleware chain for auth, logging, metrics

**Status:** Complete. This is the core of metatools-mcp.

---

### 2. Resources (❌ NOT IMPLEMENTED)

**Control:** Application-controlled - Client decides what context to include

**What it does:** Exposes read-only data (files, database content, API responses) that provides context to language models. Resources are identified by URIs and can be subscribed to for real-time updates.

**JSON-RPC Methods:**
- `resources/list` - List available resources (paginated)
- `resources/read` - Read resource contents
- `resources/templates/list` - List resource templates (parameterized URIs)
- `resources/subscribe` - Subscribe to resource changes
- `resources/unsubscribe` - Unsubscribe from resource
- `notifications/resources/list_changed` - Resource list changed
- `notifications/resources/updated` - Resource content changed

**Data Structures:**
```go
type Resource struct {
    URI         string       // Unique identifier (file://, https://, custom://)
    Name        string       // Display name
    Title       string       // Human-readable title
    Description string       // Optional description
    MimeType    string       // Content type
    Size        int64        // Optional size in bytes
    Icons       []Icon       // Display icons
    Annotations *Annotations // Audience, priority, lastModified
}

type ResourceTemplate struct {
    URITemplate string       // RFC 6570 URI template (e.g., "file:///{path}")
    Name        string
    Description string
    MimeType    string
}

type ResourceContents struct {
    URI      string
    MimeType string
    Text     string  // For text content
    Blob     string  // For binary (base64)
}
```

**Use Cases:**
- Expose file system contents
- Database schemas and query results
- API documentation
- Git repository contents
- Configuration files

**Library Opportunity: `toolresource`**

```go
// toolresource - Resource providers for MCP
package toolresource

// Core interface
type ResourceProvider interface {
    List(ctx context.Context, cursor string) ([]Resource, string, error)
    Read(ctx context.Context, uri string) (*ResourceContents, error)
    Subscribe(ctx context.Context, uri string) (<-chan ResourceUpdate, error)
    Unsubscribe(ctx context.Context, uri string) error
}

// Implementations
type FileSystemProvider struct { ... }  // file:// resources
type GitProvider struct { ... }         // git:// resources
type HTTPProvider struct { ... }        // Proxy external APIs
type DatabaseProvider struct { ... }    // Database schemas/data
type CompositeProvider struct { ... }   // Aggregate multiple providers
```

**Integration with metatools-mcp:**
- Add `internal/resource/` package
- Register ResourceProvider with MCP server
- Support resource templates with completion

---

### 3. Prompts (❌ NOT IMPLEMENTED)

**Control:** User-controlled - Users explicitly select prompts (slash commands, menus)

**What it does:** Provides pre-defined, versioned instruction templates that standardize how models perform common tasks. Can include dynamic arguments and embedded resources.

**JSON-RPC Methods:**
- `prompts/list` - List available prompts (paginated)
- `prompts/get` - Get prompt with arguments filled in
- `notifications/prompts/list_changed` - Prompt list changed

**Data Structures:**
```go
type Prompt struct {
    Name        string            // Unique identifier
    Title       string            // Human-readable name
    Description string            // What the prompt does
    Arguments   []PromptArgument  // Input parameters
    Icons       []Icon            // Display icons
}

type PromptArgument struct {
    Name        string
    Description string
    Required    bool
}

type PromptMessage struct {
    Role    Role     // "user" or "assistant"
    Content Content  // Text, Image, Audio, or EmbeddedResource
}

type GetPromptResult struct {
    Description string
    Messages    []PromptMessage
}
```

**Use Cases:**
- Code review prompts with code input
- Translation prompts with language selection
- Analysis prompts with document context
- Workflow templates (debugging, refactoring)

**Library Opportunity: `toolprompt`**

```go
// toolprompt - Prompt templates for MCP
package toolprompt

// Core interface
type PromptProvider interface {
    List(ctx context.Context, cursor string) ([]Prompt, string, error)
    Get(ctx context.Context, name string, args map[string]string) (*PromptResult, error)
}

// Template engine
type PromptTemplate struct {
    Name        string
    Description string
    Arguments   []ArgumentDef
    Template    string  // Go template or custom format
}

// Implementations
type FileBasedProvider struct { ... }    // Load from YAML/JSON files
type EmbeddedProvider struct { ... }     // Compiled-in prompts
type DynamicProvider struct { ... }      // Generated at runtime
type CompositeProvider struct { ... }    // Aggregate multiple
```

**Integration Points:**
- Could connect with **toolskill** for SKILL.md-style workflows
- Prompts can embed resources from **toolresource**

---

### 4. Sampling (❌ NOT IMPLEMENTED)

**Control:** Server-initiated - Server requests LLM completions from client

**What it does:** Allows MCP servers to request language model generations through the client. This enables servers to implement agentic behaviors without needing their own LLM API keys.

**JSON-RPC Methods:**
- `sampling/createMessage` - Request LLM completion

**Data Structures:**
```go
type CreateMessageParams struct {
    Messages         []SamplingMessage
    ModelPreferences *ModelPreferences
    SystemPrompt     string
    IncludeContext   string  // "none", "thisServer", "allServers"
    Temperature      float64
    MaxTokens        int64
    StopSequences    []string
    Metadata         map[string]any
    // Tool use support
    Tools            []Tool
    ToolChoice       *ToolChoice
}

type ModelPreferences struct {
    Hints               []ModelHint  // Suggested models
    CostPriority        float64      // 0-1, prefer cheaper
    SpeedPriority       float64      // 0-1, prefer faster
    IntelligencePriority float64     // 0-1, prefer smarter
}

type CreateMessageResult struct {
    Role       Role
    Content    Content
    Model      string
    StopReason string  // "endTurn", "stopSequence", "toolUse"
}
```

**Use Cases:**
- Agentic tool execution with LLM reasoning
- Content generation within tool pipelines
- Dynamic response formatting
- Multi-step reasoning chains

**Library Opportunity: `toolsampling`**

```go
// toolsampling - Server-initiated LLM sampling
package toolsampling

// Sampler interface for requesting completions
type Sampler interface {
    CreateMessage(ctx context.Context, params *CreateMessageParams) (*CreateMessageResult, error)
}

// ClientSampler wraps MCP client session
type ClientSampler struct {
    session *mcp.ServerSession
}

// Mock sampler for testing
type MockSampler struct { ... }

// Retry/fallback wrapper
type ResilientSampler struct {
    Primary   Sampler
    Fallback  Sampler
    MaxRetries int
}
```

**Security Considerations:**
- Human-in-the-loop approval required
- Client controls model selection
- Rate limiting recommended
- Server never sees full conversation

---

### 5. Roots (❌ NOT IMPLEMENTED - Client Feature)

**Control:** Client-provided - Client exposes filesystem boundaries

**What it does:** Clients expose specific directories that servers can operate within. This defines the workspace boundaries for file operations.

**JSON-RPC Methods:**
- `roots/list` - List available roots
- `notifications/roots/list_changed` - Roots changed

**Data Structures:**
```go
type Root struct {
    URI  string  // file:// URI
    Name string  // Display name
}
```

**Use Cases:**
- Project workspace boundaries
- Multi-repo development
- Secure file access control

**Library Opportunity:** This is primarily a **client feature**, but servers can query roots:

```go
// In metatools-mcp client interactions
type RootsAware interface {
    ListRoots(ctx context.Context) ([]Root, error)
    OnRootsChanged(callback func([]Root))
}
```

---

### 6. Elicitation (❌ NOT IMPLEMENTED)

**Control:** Server-initiated - Server requests user input

**What it does:** Allows servers to request additional information from users when needed. This formalizes how models ask for missing context during sessions.

**JSON-RPC Methods:**
- `elicitation/create` - Request user input

**Data Structures:**
```go
type ElicitParams struct {
    Message string
    Schema  *JSONSchema  // Expected response format
}

type ElicitResult struct {
    Action  string  // "accept", "decline", "cancel"
    Content any     // User's response
}
```

**Use Cases:**
- Request missing parameters
- Confirm destructive actions
- Gather preferences mid-workflow

**Library Opportunity: `toolelicit`**

```go
// toolelicit - User input elicitation
package toolelicit

type Elicitor interface {
    Elicit(ctx context.Context, message string, schema *JSONSchema) (*ElicitResult, error)
}

// Schema builder
type SchemaBuilder struct { ... }
func Text(message string) *ElicitParams
func Choice(message string, options []string) *ElicitParams
func Form(message string, fields []Field) *ElicitParams
```

---

## Complete Library Mapping

| MCP Feature | Library | Location | Dependencies |
|-------------|---------|----------|--------------|
| Tools | (existing) | metatools-mcp | toolmodel |
| Resources | **toolresource** | New library | toolmodel |
| Prompts | **toolprompt** | New library | toolmodel, toolresource |
| Sampling | **toolsampling** | New library | None |
| Roots | (client) | metatools-mcp | None |
| Elicitation | **toolelicit** | New library | None |

---

## Updated Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           metatools-mcp Server                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     TRANSPORT LAYER (✅ Done)                        │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────────┐                          │    │
│  │  │  Stdio  │  │   SSE   │  │ Streamable  │                          │    │
│  │  └─────────┘  └─────────┘  └─────────────┘                          │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────┴───────────────────────────────────┐    │
│  │                     MCP FEATURES LAYER                               │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │    │
│  │  │  Tools   │ │Resources │ │ Prompts  │ │ Sampling │ │ Elicit   │  │    │
│  │  │   ✅     │ │   ❌     │ │   ❌     │ │   ❌     │ │   ❌     │  │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────┴───────────────────────────────────┐    │
│  │              MIDDLEWARE CHAIN (✅ Done)                              │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐            │    │
│  │  │Logging │→│  Auth  │→│ Rate   │→│ Audit  │→│Metrics │            │    │
│  │  └────────┘ └────────┘ │ Limit  │ └────────┘ └────────┘            │    │
│  │                        └────────┘                                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │              PROVIDER/BACKEND REGISTRIES (✅ Done)                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘

LEGEND: ✅ Done  ❌ Not Started
```

---

## Implementation Priority

| Priority | Feature | Library | Effort | Value |
|----------|---------|---------|--------|-------|
| **P1** | Resources | toolresource | 2-3 weeks | High - enables context sharing |
| **P2** | Prompts | toolprompt | 1-2 weeks | Medium - user workflows |
| **P3** | Sampling | toolsampling | 2-3 weeks | High - agentic capabilities |
| **P4** | Elicitation | toolelicit | 1 week | Low - interactive workflows |

---

## Consolidated Library Update

Adding MCP features to the categorization:

```
toolops/                     # Operations
├── observe/                 # OpenTelemetry
├── cache/                   # Response caching
├── resilience/              # Circuit breaker
├── health/                  # Health checks
└── auth/                    # Authentication (extracted)

toolmcp/                     # NEW: MCP Protocol Features
├── resource/                # toolresource - Resource providers
├── prompt/                  # toolprompt - Prompt templates
├── sampling/                # toolsampling - LLM sampling
└── elicit/                  # toolelicit - User elicitation
```

Or integrate into existing structure:

```
tooldiscovery/
├── index/                   # Registry
├── search/                  # BM25 search
├── semantic/                # Hybrid search
├── docs/                    # Documentation
└── resource/                # ⭐ NEW: Resource discovery/access

toolcompose/
├── set/                     # Toolsets
├── skill/                   # Agent skills
└── prompt/                  # ⭐ NEW: Prompt templates
```

---

## Go SDK Support Status

The MCP Go SDK v1.2.0 has **full support** for all features:

| Feature | SDK Types | SDK Methods |
|---------|-----------|-------------|
| Tools | `Tool`, `CallToolResult` | `AddTool`, `CallTool` |
| Resources | `Resource`, `ResourceTemplate`, `ResourceContents` | `AddResource`, `ReadResource` |
| Prompts | `Prompt`, `PromptMessage`, `GetPromptResult` | `AddPrompt`, `GetPrompt` |
| Sampling | `CreateMessageParams`, `CreateMessageResult` | `CreateMessage` |
| Roots | `Root`, `ListRootsResult` | `ListRoots` |
| Elicitation | `ElicitParams`, `ElicitResult` | `Elicit` |

---

## Sources

- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP Features Guide - WorkOS](https://workos.com/blog/mcp-features-guide)
- [MCP Resources Specification](https://modelcontextprotocol.io/specification/2025-11-25/server/resources)
- [MCP Prompts Specification](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts)
- [MCP Sampling Specification](https://modelcontextprotocol.io/specification/2025-11-25/client/sampling)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
