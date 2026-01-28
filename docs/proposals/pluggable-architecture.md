# Pluggable Architecture Proposal

**Status:** Draft
**Date:** 2026-01-27
**Author:** Jon Raymond

## Executive Summary

This proposal outlines a pluggable, modular architecture for metatools-mcp that enables:
- Multiple transport protocols (stdio, HTTP/SSE, WebSocket)
- Plug-and-play tool providers
- Configurable search strategies
- Extensible backend registries
- Cross-cutting middleware

The design leverages Go's interface-based composition, build-tag gating, and configuration-driven initialization to create a flexible framework while maintaining the clean, canonical core.

---

## Table of Contents

1. [Motivation](#motivation)
2. [Current Architecture Analysis](#current-architecture-analysis)
3. [Proposed Architecture](#proposed-architecture)
4. [Extension Points](#extension-points)
5. [Multi-Backend Architecture](#multi-backend-architecture)
6. [Configuration Design](#configuration-design)
7. [Implementation Approach](#implementation-approach)
8. [Comparative Analysis](#comparative-analysis)
9. [References](#references)

---

## Motivation

### Goals

1. **Multi-transport support** - Run as stdio MCP server (current) or HTTP/SSE server (high availability)
2. **Plug-and-play extensibility** - Add new tools, backends, and search strategies without modifying core
3. **Configuration-driven** - YAML/JSON config files with environment variable overrides
4. **Framework potential** - Enable metatools-mcp as a reusable framework for building MCP servers

### Non-Goals

- Runtime plugin loading (Go's plugin package has platform limitations)
- Breaking changes to existing tool* library interfaces
- Over-engineering for hypothetical future requirements

---

## Current Architecture Analysis

### Strengths (85% Pluggable)

The existing architecture demonstrates excellent patterns:

```
┌─────────────────────────────────────────────────────────────────┐
│                      MCP Server (SDK)                           │
│           metatools-mcp/internal/server/server.go               │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────────┐
│                    Adapter Layer                                 │
│  - IndexAdapter (toolindex → handlers.Index)                    │
│  - DocsAdapter (tooldocs → handlers.Store)                      │
│  - RunnerAdapter (toolrun → handlers.Runner)                    │
│  - ExecutorAdapter (toolcode → handlers.Executor)               │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────────┐
│                   Handlers Layer                                 │
│  - SearchHandler, DescribeHandler, RunHandler, ChainHandler    │
│  - CodeHandler (optional), ExamplesHandler, NamespacesHandler  │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────────┐
│              Core Tool Libraries                                 │
│  toolindex, tooldocs, toolrun, toolcode, toolruntime, toolsearch│
└─────────────────────────────────────────────────────────────────┘
```

**What works well:**
- Clean interface contracts (`handlers/interfaces.go`)
- Adapter pattern prevents library leakage
- Build-tag gating for optional features (`toolsearch`, `toolruntime`)
- Configuration-driven bootstrap (`internal/config/env.go`)
- Stateless handlers with dependency injection

### Gap: Monolithic Tool Registration

The primary gap is in `server.registerTools()` (~200 lines):

```go
// Current: Hard-coded tool list
func (s *Server) registerTools() {
    s.addTool("search_tools", ...)    // inline schema
    s.addTool("describe_tool", ...)   // inline schema
    s.addTool("run_tool", ...)        // inline schema
    // ... 7 tools total
}
```

This requires code changes to add new tools.

---

## Proposed Architecture

### Five-Layer Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    METATOOLS-MCP FRAMEWORK                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                 TRANSPORT LAYER                              │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │ │
│  │  │  stdio  │  │  SSE    │  │  HTTP   │  │  gRPC   │        │ │
│  │  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │ │
│  │       └────────────┴────────────┴────────────┘              │ │
│  │                        ↓                                     │ │
│  │               transport.Transport interface                  │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                             ↓                                    │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                  MIDDLEWARE CHAIN                            │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │ │
│  │  │ Logging  │→│  Auth    │→│  Rate    │→│  Cache   │       │ │
│  │  │          │ │          │ │  Limit   │ │          │       │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                             ↓                                    │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │               TOOL PROVIDER REGISTRY                         │ │
│  │                                                               │ │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │ │
│  │  │ search_tools │ │ describe_    │ │ run_tool     │         │ │
│  │  │              │ │ tool         │ │              │         │ │
│  │  └──────────────┘ └──────────────┘ └──────────────┘         │ │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐         │ │
│  │  │ run_chain    │ │ execute_code │ │ [custom...]  │         │ │
│  │  └──────────────┘ └──────────────┘ └──────────────┘         │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                             ↓                                    │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              CORE SERVICES (Tool Libraries)                  │ │
│  │                                                               │ │
│  │  ┌────────────────┐  ┌────────────────┐                     │ │
│  │  │   toolindex    │  │   tooldocs     │                     │ │
│  │  │  (Registry)    │  │  (Disclosure)  │                     │ │
│  │  └────────────────┘  └────────────────┘                     │ │
│  │  ┌────────────────┐  ┌────────────────┐                     │ │
│  │  │   toolrun      │  │   toolcode     │                     │ │
│  │  │  (Execution)   │  │ (Orchestration)│                     │ │
│  │  └────────────────┘  └────────────────┘                     │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                             ↓                                    │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │               BACKEND REGISTRY                               │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │ │
│  │  │  local  │  │ openai  │  │  azure  │  │  mcp    │        │ │
│  │  │handlers │  │   api   │  │   api   │  │ servers │        │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Core Interfaces

```go
// Transport abstraction (enables stdio/SSE/HTTP)
type Transport interface {
    Serve(ctx context.Context, handler RequestHandler) error
}

// Tool provider (enables plug-and-play tools)
type ToolProvider interface {
    Name() string
    Tool() *mcp.Tool
    Handle(ctx context.Context, input []byte) (any, error)
}

// Middleware (enables cross-cutting concerns)
type Middleware func(ToolProvider) ToolProvider

// Backend registry (enables tool sources)
type BackendRegistry interface {
    Register(kind string, backend Backend)
    Get(kind string) (Backend, bool)
    List() []string
}
```

---

## Extension Points

### 1. Transport Layer

| Transport | Use Case | Implementation |
|-----------|----------|----------------|
| **stdio** | MCP client spawns process | Current default |
| **SSE** | Web-based MCP clients | HTTP + Server-Sent Events |
| **HTTP** | REST API access | Standard HTTP endpoints |
| **gRPC** | High-performance RPC | Future consideration |

```go
// Transport interface
type Transport interface {
    Serve(ctx context.Context, handler RequestHandler) error
    Close() error
}

// Factory pattern
func NewTransport(cfg TransportConfig) (Transport, error) {
    switch cfg.Type {
    case "stdio":
        return NewStdioTransport(), nil
    case "sse":
        return NewSSETransport(cfg.HTTP), nil
    default:
        return nil, fmt.Errorf("unknown transport: %s", cfg.Type)
    }
}
```

### 2. Search Strategy

Already implemented via build tags + interface:

```go
// toolindex.Searcher interface
type Searcher interface {
    Search(query string, limit int) ([]Summary, error)
}

// Implementations:
// - Default lexical (built into toolindex)
// - BM25 (toolsearch package, requires build tag)
// - Semantic/vector (future, could use embeddings)
```

### 3. Tool Provider Registry

New pattern to enable plug-and-play tools:

```go
// ToolProvider interface
type ToolProvider interface {
    Name() string
    Tool() *mcp.Tool  // MCP tool definition with schema
    Handle(ctx context.Context, input []byte) (any, error)
}

// Registry
type ToolRegistry struct {
    providers map[string]ToolProvider
}

func (r *ToolRegistry) Register(p ToolProvider) {
    r.providers[p.Name()] = p
}

func (r *ToolRegistry) All() []ToolProvider {
    // Returns all registered providers
}
```

**Migration path:** Convert existing handlers to ToolProvider implementations.

### 4. Backend Registry

For tool execution sources (local, API, MCP servers):

```go
// Backend interface (already in toolmodel conceptually)
type Backend interface {
    Kind() string
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
}

// Registry with configuration
type BackendRegistry struct {
    backends map[string]Backend
}

func (r *BackendRegistry) RegisterFromConfig(cfg BackendConfig) error {
    switch cfg.Kind {
    case "local":
        r.backends["local"] = NewLocalBackend(cfg.Local)
    case "openai":
        r.backends["openai"] = NewOpenAIBackend(cfg.OpenAI)
    // ...
    }
}
```

### 5. Middleware Chain

The middleware layer provides pluggable cross-cutting concerns using the decorator pattern.

#### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MIDDLEWARE CHAIN                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   Incoming Request                                                           │
│         │                                                                     │
│         ▼                                                                     │
│   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐            │
│   │  Logging  │ → │   Auth    │ → │   Rate    │ → │  Caching  │            │
│   │           │   │           │   │  Limiter  │   │           │            │
│   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘            │
│         │               │               │               │                    │
│         │   on error:   │   on error:   │   on error:   │                    │
│         │   log & pass  │   reject 401  │   reject 429  │                    │
│         │               │               │               │                    │
│         └───────────────┴───────────────┴───────────────┘                    │
│                                         │                                     │
│                                         ▼                                     │
│                               ┌───────────────┐                              │
│                               │  Tool Handler │                              │
│                               │   (actual)    │                              │
│                               └───────────────┘                              │
│                                         │                                     │
│                                         ▼                                     │
│   Response flows back through chain (for metrics, logging, caching)         │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### The Middleware Interface

```go
// Middleware wraps a ToolProvider with additional behavior
type Middleware func(ToolProvider) ToolProvider

// MiddlewareFunc is a convenience type for stateless middleware
type MiddlewareFunc func(ctx context.Context, input []byte, next NextFunc) (any, error)

type NextFunc func(ctx context.Context, input []byte) (any, error)

// MiddlewareRegistry manages available middleware
type MiddlewareRegistry struct {
    available map[string]MiddlewareFactory
    active    []Middleware
}

// MiddlewareFactory creates configured middleware instances
type MiddlewareFactory func(cfg MiddlewareConfig) (Middleware, error)
```

#### Built-in Middleware

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AVAILABLE MIDDLEWARE                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ LOGGING                                                                  │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - Request/response logging                                              │ │
│  │ - Configurable log levels (debug, info, warn, error)                   │ │
│  │ - Structured JSON output                                                │ │
│  │ - Request ID tracking                                                   │ │
│  │ - Duration metrics                                                      │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ AUTHENTICATION                                                           │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - Bearer token validation                                               │ │
│  │ - API key authentication                                                │ │
│  │ - OAuth2/OIDC integration                                               │ │
│  │ - mTLS client certificates                                              │ │
│  │ - Configurable bypass for specific tools                               │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ RATE LIMITING                                                            │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - Per-client rate limits                                                │ │
│  │ - Per-tool rate limits                                                  │ │
│  │ - Token bucket algorithm                                                │ │
│  │ - Sliding window counters                                               │ │
│  │ - Redis-backed for distributed deployments                             │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ CACHING                                                                  │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - Response caching for idempotent tools                                │ │
│  │ - Configurable TTL per tool                                            │ │
│  │ - Cache key customization                                               │ │
│  │ - In-memory or Redis backend                                           │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ METRICS                                                                  │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - Prometheus-compatible metrics                                         │ │
│  │ - Request counts, latencies, error rates                               │ │
│  │ - Per-tool and per-backend breakdowns                                  │ │
│  │ - Custom metric labels                                                  │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ TRACING                                                                  │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - OpenTelemetry integration                                             │ │
│  │ - Distributed trace propagation                                         │ │
│  │ - Span creation for each tool call                                     │ │
│  │ - Backend call tracing                                                  │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │ VALIDATION                                                               │ │
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ - JSON Schema validation on inputs                                      │ │
│  │ - Output validation                                                     │ │
│  │ - Custom validators per tool                                           │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Implementing Custom Middleware

```go
// Example: Custom audit logging middleware
type AuditMiddleware struct {
    logger AuditLogger
    next   ToolProvider
}

func NewAuditMiddleware(logger AuditLogger) Middleware {
    return func(next ToolProvider) ToolProvider {
        return &AuditMiddleware{
            logger: logger,
            next:   next,
        }
    }
}

func (m *AuditMiddleware) Name() string { return m.next.Name() }
func (m *AuditMiddleware) Tool() *mcp.Tool { return m.next.Tool() }

func (m *AuditMiddleware) Handle(ctx context.Context, input []byte) (any, error) {
    // Pre-execution: Log the request
    requestID := uuid.New().String()
    user := auth.UserFromContext(ctx)

    m.logger.LogRequest(AuditEntry{
        RequestID: requestID,
        User:      user,
        Tool:      m.next.Name(),
        Input:     input,
        Timestamp: time.Now(),
    })

    // Execute the actual tool
    result, err := m.next.Handle(ctx, input)

    // Post-execution: Log the result
    m.logger.LogResponse(AuditEntry{
        RequestID: requestID,
        User:      user,
        Tool:      m.next.Name(),
        Success:   err == nil,
        Duration:  time.Since(start),
        Timestamp: time.Now(),
    })

    return result, err
}

// Register custom middleware
func init() {
    middleware.Register("audit", func(cfg MiddlewareConfig) (Middleware, error) {
        logger := newAuditLogger(cfg)
        return NewAuditMiddleware(logger), nil
    })
}
```

#### Configuration-Driven Middleware

```yaml
# metatools.yaml
middleware:
  # Order matters - first in config = first in chain
  chain:
    - logging
    - auth
    - rate_limit
    - metrics
    - audit  # Custom middleware

  # Per-middleware configuration
  logging:
    enabled: true
    level: info
    format: json
    include_request_body: false
    include_response_body: false

  auth:
    enabled: true
    type: bearer
    token_validation:
      issuer: https://auth.example.com
      audience: metatools-api
    bypass_tools:
      - search_tools  # Allow anonymous search
      - list_namespaces

  rate_limit:
    enabled: true
    default:
      requests_per_minute: 100
      burst: 20
    per_tool:
      execute_code:
        requests_per_minute: 10
        burst: 2
    storage: memory  # or redis

  metrics:
    enabled: true
    endpoint: /metrics
    labels:
      environment: production
      service: metatools

  cache:
    enabled: true
    backend: memory  # or redis
    default_ttl: 5m
    per_tool:
      describe_tool:
        ttl: 1h
      search_tools:
        ttl: 30s

  audit:
    enabled: true
    destination: file
    path: /var/log/metatools/audit.log
```

#### Middleware Chain Construction

```go
// Chain construction from config
func BuildMiddlewareChain(cfg MiddlewareConfig) ([]Middleware, error) {
    var chain []Middleware

    for _, name := range cfg.Chain {
        mwCfg, ok := cfg.Middlewares[name]
        if !ok || !mwCfg.Enabled {
            continue
        }

        // Look up factory in registry
        factory, ok := middleware.Get(name)
        if !ok {
            return nil, fmt.Errorf("unknown middleware: %s", name)
        }

        // Create middleware instance
        mw, err := factory(mwCfg)
        if err != nil {
            return nil, fmt.Errorf("middleware %s: %w", name, err)
        }

        chain = append(chain, mw)
    }

    return chain, nil
}

// Apply chain to all providers
func ApplyToRegistry(registry *ToolRegistry, chain []Middleware) {
    for name, provider := range registry.providers {
        wrapped := provider
        for i := len(chain) - 1; i >= 0; i-- {
            wrapped = chain[i](wrapped)
        }
        registry.providers[name] = wrapped
    }
}
```

#### Request Flow Through Middleware

```
┌────────────────────────────────────────────────────────────────────────────┐
│                    REQUEST FLOW EXAMPLE                                     │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Client Request: run_tool("github/create_issue", {...})                    │
│                                                                              │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 1. LOGGING MIDDLEWARE                                                 │  │
│   │    - Generate request ID: "req-abc123"                               │  │
│   │    - Log: "Incoming request for github/create_issue"                 │  │
│   │    - Start timer                                                      │  │
│   │    → Pass to next                                                     │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 2. AUTH MIDDLEWARE                                                    │  │
│   │    - Extract token from Authorization header                         │  │
│   │    - Validate JWT signature                                          │  │
│   │    - Check claims (issuer, audience, expiry)                        │  │
│   │    - Inject user into context                                        │  │
│   │    → Pass to next (or reject with 401)                               │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 3. RATE LIMIT MIDDLEWARE                                              │  │
│   │    - Identify client (by user, IP, or API key)                       │  │
│   │    - Check token bucket for "github/create_issue"                    │  │
│   │    - Consume token                                                    │  │
│   │    → Pass to next (or reject with 429)                               │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 4. METRICS MIDDLEWARE                                                 │  │
│   │    - Increment request counter                                        │  │
│   │    - Start latency timer                                              │  │
│   │    → Pass to next                                                     │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 5. ACTUAL TOOL HANDLER                                                │  │
│   │    - Route to GitHub backend                                         │  │
│   │    - Execute create_issue                                            │  │
│   │    - Return result                                                    │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│   Response bubbles back through chain:                                      │
│   - Metrics: Record latency, success/failure                                │
│   - Rate limit: (no action on response)                                     │
│   - Auth: (no action on response)                                           │
│   - Logging: Log response, duration, status                                 │
│                                                                              │
│   Final Response → Client                                                   │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

#### Registering Custom Middleware

```go
// Register at init time (compile-time pluggability)
func init() {
    middleware.Register("my-custom", NewMyCustomMiddleware)
}

// Or register at runtime (config-driven)
func setupMiddleware(registry *MiddlewareRegistry) {
    // Built-in middleware (always available)
    registry.Register("logging", NewLoggingMiddleware)
    registry.Register("auth", NewAuthMiddleware)
    registry.Register("rate_limit", NewRateLimitMiddleware)
    registry.Register("metrics", NewMetricsMiddleware)
    registry.Register("cache", NewCacheMiddleware)

    // Custom middleware (loaded from config or plugins)
    registry.Register("audit", NewAuditMiddleware)
    registry.Register("pii-filter", NewPIIFilterMiddleware)
}
```

#### Middleware Composition Patterns

```go
// Conditional middleware (only apply to certain tools)
func OnlyForTools(tools []string, mw Middleware) Middleware {
    return func(next ToolProvider) ToolProvider {
        if slices.Contains(tools, next.Name()) {
            return mw(next)
        }
        return next
    }
}

// Except middleware (skip for certain tools)
func ExceptForTools(tools []string, mw Middleware) Middleware {
    return func(next ToolProvider) ToolProvider {
        if slices.Contains(tools, next.Name()) {
            return next
        }
        return mw(next)
    }
}

// Usage
chain := []Middleware{
    LoggingMiddleware,
    ExceptForTools([]string{"search_tools"}, AuthMiddleware),
    OnlyForTools([]string{"execute_code"}, StrictRateLimitMiddleware),
}
```

---

## Multi-Backend Architecture

The backend layer is the foundation for tool discovery and execution. It enables metatools-mcp to aggregate tools from multiple sources while presenting a unified interface to MCP clients.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           MCP CLIENT                                         │
│                    (Claude, Cursor, Custom)                                  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │ MCP Protocol
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         METATOOLS-MCP SERVER                                 │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                      UNIFIED TOOL INTERFACE                              │ │
│  │                                                                           │ │
│  │   search_tools    describe_tool    run_tool    run_chain    execute_code │ │
│  │                                                                           │ │
│  └─────────────────────────────────┬───────────────────────────────────────┘ │
│                                    │                                          │
│  ┌─────────────────────────────────▼───────────────────────────────────────┐ │
│  │                       BACKEND REGISTRY                                   │ │
│  │                                                                           │ │
│  │   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │ │
│  │   │   LOCAL     │ │   OPENAI    │ │    MCP      │ │   CUSTOM    │       │ │
│  │   │  Backend    │ │  Backend    │ │  Backend    │ │  Backend    │       │ │
│  │   └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘       │ │
│  │          │               │               │               │               │ │
│  └──────────┼───────────────┼───────────────┼───────────────┼───────────────┘ │
└─────────────┼───────────────┼───────────────┼───────────────┼───────────────┘
              │               │               │               │
              ▼               ▼               ▼               ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
        │  Local   │   │  OpenAI  │   │  Other   │   │  Custom  │
        │  Files   │   │   API    │   │   MCP    │   │   API    │
        │          │   │          │   │ Servers  │   │          │
        └──────────┘   └──────────┘   └──────────┘   └──────────┘
```

### The Backend Interface

Each backend implements a common interface, allowing uniform handling regardless of the tool source:

```go
// Backend defines a source of tools
type Backend interface {
    // Identity
    Kind() string                                    // e.g., "local", "openai", "mcp"
    Name() string                                    // Instance name for disambiguation

    // Configuration
    Configure(raw []byte) error                      // Parse backend-specific config

    // Discovery
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)

    // Execution
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)

    // Lifecycle
    Start(ctx context.Context) error
    Stop() error
}
```

### Backend Types

#### 1. Local Backend
Tools defined as files on disk (YAML, JSON, or Go handlers).

```
┌─────────────────────────────────────────────────────┐
│                   LOCAL BACKEND                      │
├─────────────────────────────────────────────────────┤
│                                                       │
│   ~/.config/metatools/tools/                         │
│   ├── calculator.yaml      → Tool definition         │
│   ├── file-ops.yaml        → Tool definition         │
│   └── custom/                                        │
│       └── my-tool.yaml     → Tool definition         │
│                                                       │
│   /usr/share/metatools/tools/                        │
│   └── system-tools.yaml    → System-wide tools       │
│                                                       │
├─────────────────────────────────────────────────────┤
│  Config:                                             │
│    paths:                                            │
│      - ~/.config/metatools/tools                    │
│      - /usr/share/metatools/tools                   │
│    watch: true  # Hot reload on changes             │
└─────────────────────────────────────────────────────┘
```

#### 2. API Backends (OpenAI, Azure, Anthropic)
Tools exposed via LLM provider APIs.

```
┌─────────────────────────────────────────────────────┐
│                   OPENAI BACKEND                     │
├─────────────────────────────────────────────────────┤
│                                                       │
│   ┌─────────────┐        ┌─────────────────────┐    │
│   │ metatools   │ ──────▶│  OpenAI API         │    │
│   │   run_tool  │        │  /chat/completions  │    │
│   └─────────────┘        │  (function calling) │    │
│                          └─────────────────────┘    │
│                                                       │
├─────────────────────────────────────────────────────┤
│  Config:                                             │
│    api_key: ${OPENAI_API_KEY}                       │
│    organization: ${OPENAI_ORG}                      │
│    models:                                           │
│      - gpt-4                                        │
│      - gpt-4-turbo                                  │
│    timeout: 30s                                     │
└─────────────────────────────────────────────────────┘
```

#### 3. MCP Backend
Connect to other MCP servers as tool sources.

```
┌─────────────────────────────────────────────────────┐
│                    MCP BACKEND                       │
├─────────────────────────────────────────────────────┤
│                                                       │
│   metatools-mcp                                      │
│        │                                             │
│        │ stdio                                       │
│        ▼                                             │
│   ┌─────────────┐                                   │
│   │  GitHub     │  ← npx @modelcontextprotocol/     │
│   │  MCP Server │       server-github               │
│   └─────────────┘                                   │
│        │                                             │
│        ▼                                             │
│   GitHub API tools:                                  │
│   - create_issue                                    │
│   - list_pull_requests                              │
│   - search_code                                     │
│                                                       │
├─────────────────────────────────────────────────────┤
│  Config:                                             │
│    kind: mcp                                        │
│    command: npx                                     │
│    args:                                            │
│      - "-y"                                         │
│      - "@modelcontextprotocol/server-github"        │
│    env:                                             │
│      GITHUB_TOKEN: ${GITHUB_TOKEN}                  │
└─────────────────────────────────────────────────────┘
```

#### 4. HTTP Backend
Tools exposed via REST APIs.

```
┌─────────────────────────────────────────────────────┐
│                   HTTP BACKEND                       │
├─────────────────────────────────────────────────────┤
│                                                       │
│   metatools-mcp                                      │
│        │                                             │
│        │ HTTPS                                       │
│        ▼                                             │
│   ┌─────────────────────────┐                       │
│   │  Internal Tool Server   │                       │
│   │  tools.company.internal │                       │
│   └─────────────────────────┘                       │
│        │                                             │
│        ▼                                             │
│   Endpoints:                                         │
│   - POST /tools/list                                │
│   - POST /tools/{name}/execute                      │
│                                                       │
├─────────────────────────────────────────────────────┤
│  Config:                                             │
│    base_url: https://tools.company.internal         │
│    auth:                                            │
│      type: oauth2                                   │
│      client_id: ${OAUTH_CLIENT_ID}                  │
│      client_secret: ${OAUTH_CLIENT_SECRET}          │
│    headers:                                         │
│      X-Custom-Header: value                         │
│    timeout: 30s                                     │
│    retry:                                           │
│      max_attempts: 3                                │
│      backoff: exponential                           │
└─────────────────────────────────────────────────────┘
```

#### 5. Custom Backend
For specialized integrations that don't fit standard patterns.

```
┌─────────────────────────────────────────────────────┐
│                  CUSTOM BACKEND                      │
├─────────────────────────────────────────────────────┤
│                                                       │
│   Implement the Backend interface in Go:             │
│                                                       │
│   type MyCustomBackend struct {                     │
│       // Your fields                                │
│   }                                                  │
│                                                       │
│   func (b *MyCustomBackend) Kind() string {         │
│       return "my-custom"                            │
│   }                                                  │
│                                                       │
│   func (b *MyCustomBackend) Configure(raw []byte)   │
│       error {                                       │
│       // Parse your custom config                   │
│       return yaml.Unmarshal(raw, &b.config)         │
│   }                                                  │
│                                                       │
│   func (b *MyCustomBackend) Execute(ctx, tool,      │
│       args) (any, error) {                          │
│       // Your execution logic                       │
│   }                                                  │
│                                                       │
├─────────────────────────────────────────────────────┤
│  Config (passed as raw bytes):                      │
│    kind: custom                                     │
│    handler: my-custom                               │
│    config:                                          │
│      whatever_you_need: value                       │
│      nested:                                        │
│        custom: structure                            │
└─────────────────────────────────────────────────────┘
```

### Configuration Examples

#### Comprehensive Multi-Backend Setup

```yaml
# metatools.yaml
backends:
  # Local file-based tools
  local:
    enabled: true
    paths:
      - ~/.config/metatools/tools
      - /usr/share/metatools/tools
    watch: true  # Hot reload

  # OpenAI function calling
  openai:
    enabled: true
    api_key: ${OPENAI_API_KEY}
    organization: ${OPENAI_ORG}
    models:
      - gpt-4
      - gpt-4-turbo
    timeout: 30s

  # Azure OpenAI
  azure-openai:
    enabled: true
    kind: azure
    config:
      endpoint: https://my-resource.openai.azure.com
      api_key: ${AZURE_OPENAI_KEY}
      api_version: "2024-02-15-preview"
      deployment: gpt-4

  # GitHub MCP server
  github:
    enabled: true
    kind: mcp
    config:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-github"]
      env:
        GITHUB_TOKEN: ${GITHUB_TOKEN}

  # Filesystem MCP server
  filesystem:
    enabled: true
    kind: mcp
    config:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"]

  # Internal company tool server
  internal-tools:
    enabled: true
    kind: http
    config:
      base_url: https://tools.internal.company.com/api/v1
      auth:
        type: oauth2
        token_url: https://auth.company.com/oauth/token
        client_id: ${INTERNAL_CLIENT_ID}
        client_secret: ${INTERNAL_CLIENT_SECRET}
        scopes: ["tools:read", "tools:execute"]
      timeout: 60s

  # LangChain tools
  langchain:
    enabled: false
    kind: custom
    config:
      toolkit: serpapi
      api_key: ${SERPAPI_KEY}

  # Database tools
  database:
    enabled: true
    kind: custom
    config:
      driver: postgres
      connection_string: ${DATABASE_URL}
      read_only: true
      allowed_schemas: ["public", "analytics"]
```

### Tool Aggregation Flow

When a client searches for or executes tools, metatools-mcp aggregates across all backends:

```
┌────────────────────────────────────────────────────────────────────────────┐
│                         TOOL SEARCH FLOW                                    │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Client: search_tools("file operations")                                   │
│                           │                                                  │
│                           ▼                                                  │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    BACKEND REGISTRY                                  │   │
│   │                                                                       │   │
│   │   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐            │   │
│   │   │  local  │   │ github  │   │  azure  │   │  http   │            │   │
│   │   └────┬────┘   └────┬────┘   └────┬────┘   └────┬────┘            │   │
│   │        │             │             │             │                  │   │
│   │        ▼             ▼             ▼             ▼                  │   │
│   │   ListTools()   ListTools()   ListTools()   ListTools()            │   │
│   │        │             │             │             │                  │   │
│   │        └─────────────┴──────┬──────┴─────────────┘                  │   │
│   │                             │                                        │   │
│   │                             ▼                                        │   │
│   │                    ┌─────────────────┐                              │   │
│   │                    │   AGGREGATOR    │                              │   │
│   │                    │  - Merge tools  │                              │   │
│   │                    │  - Deduplicate  │                              │   │
│   │                    │  - Apply search │                              │   │
│   │                    └────────┬────────┘                              │   │
│   │                             │                                        │   │
│   └─────────────────────────────┼───────────────────────────────────────┘   │
│                                 ▼                                            │
│   Results:                                                                   │
│   [                                                                          │
│     { id: "local/file-read",      backend: "local",    score: 0.95 },       │
│     { id: "github/get-file",      backend: "github",   score: 0.87 },       │
│     { id: "filesystem/read_file", backend: "mcp",      score: 0.82 },       │
│   ]                                                                          │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

### Tool Execution Flow

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        TOOL EXECUTION FLOW                                  │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Client: run_tool("github/create_issue", { title: "Bug", body: "..." })   │
│                           │                                                  │
│                           ▼                                                  │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                      TOOL ROUTER                                     │   │
│   │                                                                       │   │
│   │   1. Parse tool ID: "github/create_issue"                           │   │
│   │      └─ backend: "github"                                           │   │
│   │      └─ tool: "create_issue"                                        │   │
│   │                                                                       │   │
│   │   2. Lookup backend in registry                                     │   │
│   │      └─ Found: GitHubMCPBackend                                     │   │
│   │                                                                       │   │
│   │   3. Delegate execution                                              │   │
│   │      └─ backend.Execute(ctx, "create_issue", args)                  │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                           │                                                  │
│                           ▼                                                  │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                   GITHUB MCP BACKEND                                 │   │
│   │                                                                       │   │
│   │   1. Forward to MCP subprocess                                      │   │
│   │      └─ npx @modelcontextprotocol/server-github                     │   │
│   │                                                                       │   │
│   │   2. MCP call: tools/call { name: "create_issue", arguments: ... }  │   │
│   │                                                                       │   │
│   │   3. Receive result                                                  │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                           │                                                  │
│                           ▼                                                  │
│   Result:                                                                    │
│   {                                                                          │
│     result: { issue_number: 123, url: "https://github.com/..." },          │
│     backend: { kind: "mcp", name: "github" },                              │
│     duration_ms: 1250                                                       │
│   }                                                                          │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

### Registration Patterns

Backends can be registered via YAML configuration or programmatically:

```go
// YAML-driven registration (config file)
func (r *BackendRegistry) LoadFromConfig(cfg BackendsConfig) error {
    for name, backendCfg := range cfg.Backends {
        if !backendCfg.Enabled {
            continue
        }

        // Create backend based on kind
        backend, err := r.createBackend(backendCfg.Kind)
        if err != nil {
            return fmt.Errorf("backend %s: %w", name, err)
        }

        // Configure with raw config (backend parses itself)
        if err := backend.Configure(backendCfg.RawConfig); err != nil {
            return fmt.Errorf("backend %s config: %w", name, err)
        }

        r.backends[name] = backend
    }
    return nil
}

// Programmatic registration (code-driven)
func main() {
    registry := NewBackendRegistry()

    // Register a custom backend programmatically
    myBackend := &MyCustomBackend{
        // ... configuration
    }
    registry.Register("my-backend", myBackend)

    // Or use the builder pattern
    registry.
        WithLocal("~/.config/metatools/tools").
        WithOpenAI(os.Getenv("OPENAI_API_KEY")).
        WithMCP("github", "npx", "-y", "@modelcontextprotocol/server-github")
}
```

### Hybrid Configuration

For maximum flexibility, backends support both YAML and code:

```go
// Backend interface with hybrid config support
type Backend interface {
    Kind() string
    Name() string

    // Option 1: YAML-driven config
    Configure(raw []byte) error

    // Option 2: Programmatic config (for complex backends)
    ConfigureWith(opts ...BackendOption) error

    // Operations
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
}

// Usage: Some config in YAML, some in code
func setupBackends(registry *BackendRegistry, cfg Config) error {
    // Load standard backends from YAML
    if err := registry.LoadFromConfig(cfg.Backends); err != nil {
        return err
    }

    // Add a complex custom backend programmatically
    customBackend := NewDatabaseBackend(
        WithConnectionPool(pool),
        WithQueryValidator(validator),
        WithAuditLogger(auditLog),
    )
    registry.Register("database", customBackend)

    return nil
}
```

### Error Handling Across Backends

```go
// Backend errors include source information
type BackendError struct {
    Backend   string    // Which backend failed
    Operation string    // What operation failed
    Tool      string    // Which tool (if applicable)
    Cause     error     // Underlying error
    Retryable bool      // Can this be retried?
}

// Aggregated errors for multi-backend operations
type AggregatedError struct {
    Errors []BackendError
}

func (e *AggregatedError) Error() string {
    // Format: "3 backends failed: local: connection refused, openai: rate limited, ..."
}

// Partial success handling
type AggregatedResult struct {
    Tools   []toolmodel.Tool  // Successfully retrieved tools
    Errors  []BackendError    // Backends that failed
    Partial bool              // True if some backends failed
}
```

### Summary: What This Enables

| Capability | Description |
|------------|-------------|
| **Tool Aggregation** | Single search across all backends |
| **Unified Execution** | Consistent interface regardless of backend |
| **Hot Plugging** | Add/remove backends via config without code changes |
| **Custom Backends** | Write Go code for specialized integrations |
| **MCP Composition** | Chain multiple MCP servers together |
| **Hybrid Config** | YAML for standard backends, code for complex ones |
| **Graceful Degradation** | Continue working if some backends fail |

---

## Configuration Design

### Recommended: Koanf + Cobra

| Component | Library | Rationale |
|-----------|---------|-----------|
| CLI framework | Cobra | Subcommands (`stdio`, `serve`, `version`) |
| Config loading | Koanf | Lighter than Viper, modular providers |

### Configuration Structure

```yaml
# metatools.yaml
server:
  name: "metatools-mcp"
  version: "0.2.0"

transport:
  type: sse           # stdio | sse | http
  http:
    host: "0.0.0.0"
    port: 8080
    tls:
      enabled: true
      cert: /etc/ssl/cert.pem
      key: /etc/ssl/key.pem
    timeouts:
      read: 30s
      write: 30s
      idle: 120s

search:
  strategy: bm25      # lexical | bm25 | semantic
  bm25:
    name_boost: 3
    namespace_boost: 2
    tags_boost: 2
    max_docs: 0
    max_doctext_len: 0

execution:
  timeout: 10s
  max_tool_calls: 64
  max_chain_steps: 8

# Tool providers (each gets own config section)
providers:
  search_tools:
    enabled: true
  describe_tool:
    enabled: true
    default_level: summary
  run_tool:
    enabled: true
  run_chain:
    enabled: true
  execute_code:
    enabled: true      # Requires toolruntime build tag
    sandbox: dev

# Backend sources for tools
backends:
  local:
    enabled: true
    paths:
      - ~/.config/metatools/tools
      - /usr/share/metatools/tools
  openai:
    enabled: false
    api_key: ${OPENAI_API_KEY}
  azure:
    enabled: false
    tenant_id: ${AZURE_TENANT_ID}

# Middleware chain
middleware:
  logging:
    enabled: true
    level: info
  auth:
    enabled: false
    type: bearer
  rate_limit:
    enabled: false
    requests_per_minute: 100
```

### Configuration Precedence

```
CLI flags > Environment variables > Config file > Defaults
```

### Plugin Configuration Pattern

Plugins receive raw config and parse themselves:

```go
type Plugin interface {
    Name() string
    Configure(raw []byte) error  // Plugin parses its own config
    Start(ctx context.Context) error
    Stop() error
}
```

---

## Implementation Approach

### Phase 1: CLI Framework (Cobra + Koanf)

1. Add Cobra for subcommands (`metatools stdio`, `metatools serve`)
2. Add Koanf for config file loading
3. Maintain backward compatibility (env vars still work)

### Phase 2: Transport Abstraction

1. Define `Transport` interface
2. Extract current stdio logic into `StdioTransport`
3. Add `SSETransport` for HTTP/SSE mode
4. Wire transport selection to config/CLI

### Phase 3: Tool Provider Registry

1. Define `ToolProvider` interface
2. Convert existing handlers to providers
3. Replace `registerTools()` with registry iteration
4. Enable external provider registration

### Phase 4: Backend Registry

1. Define `BackendRegistry` interface
2. Implement config-driven backend loading
3. Support local, API, and MCP server backends

### Phase 5: Middleware Chain

1. Define `Middleware` type
2. Implement logging, auth, rate limiting
3. Apply middleware via config

---

## Comparative Analysis

### Your Tool Libraries vs. Industry Patterns

| Your Pattern | Industry Standard | Alignment |
|--------------|-------------------|-----------|
| **toolindex.Searcher interface** | Register-on-init pattern | Excellent |
| **Build-tag gating** | HashiCorp conditional compilation | Same approach |
| **Adapter layer** | Clean Architecture boundaries | Textbook |
| **Progressive disclosure** | Apple API design principles | Ahead of most |

### Comparison with Other Go MCP Servers

| Project | Transport | Plugin System | Your Advantage |
|---------|-----------|---------------|----------------|
| Official go-sdk | stdio, SSE | None | Your tool* libraries |
| mark3labs/mcp-go | stdio, SSE | Basic | Your progressive disclosure |
| viant/mcp | stdio | None | Your modular architecture |
| **metatools-mcp** | stdio (SSE planned) | Build-tag + interfaces | Full stack orchestration |

**Unique value:** No other Go MCP server has layered tool libraries with progressive disclosure, BM25 search, and code execution.

---

## References

### Go Plugin Patterns
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - RPC-based plugin system
- [Register-on-Init Pattern](https://entropy.cat/modular-programming-in-go-part-1-the-register-on-init-pattern/)
- [Interface Extension Pattern](https://www.dolthub.com/blog/2022-09-12-golang-interface-extension/)
- [Clean Architecture with Plugins](https://cekrem.github.io/posts/clean-architecture-and-plugins-in-go/)

### MCP Implementations
- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
- [viant/mcp](https://github.com/viant/mcp)

### Configuration Libraries
- [Koanf](https://github.com/knadh/koanf) - Lighter Viper alternative
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Dependency Injection Patterns](https://www.glukhov.org/post/2025/12/dependency-injection-in-go/)

### API Design
- [Progressive Disclosure](https://en.wikipedia.org/wiki/Progressive_disclosure)
- [Apple WWDC22: API Design](https://developer.apple.com/videos/play/wwdc2022/10059/)

---

## Open Questions

1. **ToolProvider interface** - Does the proposed interface feel right for plug-and-play tools?
2. **Middleware chain** - Useful now, or defer until needed?
3. **Backend configuration** - YAML-driven or code-only for now?
4. **Semantic search** - Priority for vector/embedding-based search strategy?

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-27 | Initial draft |
| 2026-01-27 | Added Multi-Backend Architecture section with diagrams |
| 2026-01-27 | Expanded Middleware Chain section with pluggable design |
