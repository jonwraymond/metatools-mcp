# Metatools Architecture Roadmap

**Status:** Master Plan
**Version:** 1.0.0
**Last Updated:** 2026-01-28

> **Guiding Principle**: Simple and elegant at the core, extensible through modular, pluggable architecture. The existing 7 libraries represent a mature, well-designed foundation. New work focuses on **exposure, configuration, and enterprise capabilities**—not redesign.

---

## Executive Summary

### Current State
- **7 production libraries** with clean interfaces
- **13 extension points** already implemented as Go interfaces
- **85% championship-level** architecture
- **Primary gap**: Exposure and configuration, not architecture

### Target State
- **22 total libraries** (7 existing + 15 new)
- **95%+ championship-level** with full enterprise capabilities
- **Protocol-agnostic** tool platform
- **Multi-tenant** with pluggable isolation strategies
- **Agent skills** for higher-level capability composition

### Timeline Overview
```
                    CORE EXPOSURE          ENTERPRISE FEATURES         AGENT SKILLS
                    ─────────────          ───────────────────         ────────────
Week 1-2   ████████ CLI + Config
Week 3-4   ████████ Transport Layer
Week 5     ████████ Provider Registry
Week 6-7   ████████ Backend Registry
Week 8-9              ████████ Protocol Adapters
Week 10-11            ████████ Observability + Caching
Week 12-13            ████████ Multi-Tenancy
Week 14-15            ████████ Versioning + Resilience
Week 16-17            ████████ Semantic Search + Gateway
Week 18-19                                                    ████████ Skill Core
Week 20-21                                                    ████████ Orchestration
                    ─────────────────────────────────────────────────────────────────
                    MVP: 7 weeks | Full: 17 weeks | Skills: 21 weeks
```

---

## Table of Contents

1. [Library Inventory](#1-library-inventory)
2. [Dependency Map](#2-dependency-map)
3. [Work Streams](#3-work-streams)
   - [Stream A: Core Exposure](#stream-a-core-exposure)
   - [Stream B: Protocol Layer](#stream-b-protocol-layer)
   - [Stream C: Cross-Cutting](#stream-c-cross-cutting)
   - [Stream D: Enterprise](#stream-d-enterprise)
   - [Stream E: Agent Skills](#stream-e-agent-skills)
4. [Phase Breakdown](#4-phase-breakdown)
5. [Interface Contracts](#5-interface-contracts)
6. [Edge Cases & Considerations](#6-edge-cases--considerations)
7. [Rollout Strategy](#7-rollout-strategy)
8. [Multi-Language Extensibility](#8-multi-language-extensibility)

---

## 1. Library Inventory

### Existing Libraries (Production)

| Library | Version | Purpose | Extension Points | Changes Needed |
|---------|---------|---------|------------------|----------------|
| **toolmodel** | v0.1.2 | Core data models, schemas | SchemaValidator | Add Version field |
| **toolindex** | v0.1.8 | Tool registry, discovery | Searcher, BackendSelector | Multi-backend events |
| **tooldocs** | v0.1.10 | Progressive disclosure docs | Store, ToolResolver | Bulk registration |
| **toolsearch** | v0.1.9 | BM25 search implementation | (via Searcher) | None |
| **toolrun** | v0.1.9 | Execution orchestration | Runner, MCPExecutor, ProviderExecutor, LocalRegistry | HTTP/gRPC executors |
| **toolcode** | v0.1.10 | Code execution | Engine, Logger | Engine registry |
| **toolruntime** | v0.1.10 | Sandbox isolation (10 backends) | Backend, ToolGateway | None |

### Proposed Libraries (New)

| Library | Stream | Purpose | Priority | Effort | Dependencies |
|---------|--------|---------|----------|--------|--------------|
| **tooladapter** | Protocol | Protocol-agnostic tool abstraction | High | 2w | toolmodel |
| **toolset** | Protocol | Composable tool collections | High | 2w | tooladapter, toolindex |
| **toolversion** | Cross-Cut | Semantic versioning, negotiation | High | 2w | toolmodel |
| **toolcache** | Cross-Cut | Pluggable caching (Redis/Memory) | High | 2w | None |
| **toolobserve** | Cross-Cut | OpenTelemetry tracing + metrics | High | 2w | None |
| **toolresilience** | Cross-Cut | Circuit breaker, retry, bulkhead | Medium | 2w | None |
| **toolhealth** | Cross-Cut | Health checks, readiness probes | Medium | 1w | None |
| **toolsecrets** | Cross-Cut | Vault/AWS secrets management | Medium | 2w | None |
| **toolflags** | Cross-Cut | Feature flags (LaunchDarkly) | Low | 1w | None |
| **toolaudit** | Cross-Cut | Immutable audit logging | Medium | 2w | None |
| **toolpressure** | Cross-Cut | Backpressure, load shedding | Low | 1w | None |
| **toolsemantic** | Enterprise | Hybrid search (BM25+vector), GraphRAG, reranking, ColBERT | High | 3w | toolindex, toolsearch |
| **toolresource** | Enterprise | MCP Resources support | Medium | 2w | toolindex |
| **toolgateway** | Enterprise | Auth, rate limit, analytics proxy | Medium | 3w | All |
| **toolskill** | Skills | SKILL.md-compatible agent skills, workflows | Medium | 4w | toolset, toolrun |

---

## 2. Dependency Map

### Library Dependencies (DAG)

```
                            ┌─────────────────┐
                            │    toolskill    │ (L5)
                            │  Agent Skills   │
                            └────────┬────────┘
                                     │
          ┌──────────────────────────┴──────────────────────────┐
          │                                                      │
          ▼                                                      ▼
┌─────────────────┐                                   ┌─────────────────┐
│   toolgateway   │ (L4)                              │    toolrun      │
│  Auth + Proxy   │                                   │   (execution)   │
└────────┬────────┘                                   └─────────────────┘
         │
         │
          ┌──────────────────────────┼──────────────────────────┐
          │                          │                          │
          ▼                          ▼                          ▼
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   toolobserve   │      │  toolresilience │      │    toolaudit    │ (L3)
│   OpenTelemetry │      │ Circuit Breaker │      │  Audit Logging  │
└────────┬────────┘      └────────┬────────┘      └────────┬────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│    toolset      │      │   toolversion   │      │    toolcache    │ (L2)
│  Composable     │      │   Versioning    │      │    Caching      │
└────────┬────────┘      └────────┬────────┘      └────────┬────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
                                  ▼
                        ┌─────────────────┐
                        │   tooladapter   │ (L1)
                        │ Protocol Adapt  │
                        └────────┬────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
    ▼                            ▼                            ▼
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│toolmodel│  │toolindex│  │tooldocs │  │ toolrun │  │toolcode │ (L0)
│  v0.1.2 │  │  v0.1.8 │  │ v0.1.10 │  │  v0.1.9 │  │ v0.1.10 │
└────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
     │            │            │            │            │
     │            │            │            │            ▼
     │            │            │            │      ┌─────────────┐
     │            │            │            └─────►│ toolruntime │
     │            │            │                   │   v0.1.10   │
     │            │            ▼                   └─────────────┘
     │            │      ┌─────────┐
     │            └─────►│toolsearch│
     │                   │  v0.1.9  │
     └───────────────────┴──────────┘
```

### Upstream/Downstream Impact Matrix

| Change In | Affects Upstream | Affects Downstream |
|-----------|------------------|-------------------|
| toolmodel | None | All libraries |
| toolindex | None | tooldocs, toolrun, toolset |
| tooladapter | toolmodel | toolset, toolgateway |
| toolversion | toolmodel | tooladapter, toolrun |
| toolcache | None | toolindex, tooldocs, toolrun |
| toolobserve | None | All libraries (optional) |
| toolskill | toolset, toolrun | metatools (skill exposure) |

---

## 3. Work Streams

### Stream A: Core Exposure

**Goal**: Expose existing 13 extension points via configuration and CLI.

| Phase | Work Package | Duration | Deliverables |
|-------|-------------|----------|--------------|
| A1 | CLI Framework | 2 weeks | Cobra CLI, subcommands |
| A2 | Configuration | 1 week | Koanf loader, YAML schema |
| A3 | Transport Layer | 2 weeks | Transport interface, stdio/SSE/HTTP |
| A4 | Provider Registry | 1 week | ToolProvider interface, registry |
| A5 | Backend Registry | 2 weeks | Backend interface, aggregator |

**Total: 8 weeks** | **MVP: Phases A1-A4 (6 weeks)**

#### A1: CLI Framework (2 weeks)

```
Week 1:
├── Day 1-2: Cobra setup, root command
├── Day 3-4: `stdio` subcommand (existing behavior)
├── Day 5: `serve` subcommand (HTTP server)

Week 2:
├── Day 1-2: `version`, `validate` commands
├── Day 3-4: Signal handling, graceful shutdown
├── Day 5: Documentation, examples
```

**Interface Contract:**
```go
// cmd/metatools/main.go
func main() {
    rootCmd := &cobra.Command{Use: "metatools"}
    rootCmd.AddCommand(
        newStdioCmd(),   // MCP over stdio (default)
        newServeCmd(),   // HTTP/SSE server
        newVersionCmd(), // Print version
        newValidateCmd(), // Validate config
    )
    rootCmd.Execute()
}
```

**Edge Cases:**
- [ ] Backward compatibility: `metatools` alone = `metatools stdio`
- [ ] Environment variables override YAML config
- [ ] Invalid config: fail fast with clear error messages
- [ ] Missing optional config: sensible defaults

#### A2: Configuration (1 week)

```
Week 3:
├── Day 1-2: Koanf setup, YAML parser
├── Day 3: Environment variable binding
├── Day 4: Config validation schema
├── Day 5: Default values, documentation
```

**Interface Contract:**
```go
// internal/config/config.go
type Config struct {
    Server     ServerConfig     `koanf:"server"`
    Transport  TransportConfig  `koanf:"transport"`
    Search     SearchConfig     `koanf:"search"`
    Execution  ExecutionConfig  `koanf:"execution"`
    Backends   BackendsConfig   `koanf:"backends"`
    Middleware MiddlewareConfig `koanf:"middleware"`
}

func Load(path string) (*Config, error)
func LoadWithEnv(path string, prefix string) (*Config, error)
```

**Edge Cases:**
- [ ] Config file not found: use defaults
- [ ] Partial config: merge with defaults
- [ ] Invalid values: return typed validation errors
- [ ] Hot reload: optional, via `toolconfig` watcher

#### A3: Transport Layer (2 weeks)

```
Week 4:
├── Day 1-2: Transport interface definition
├── Day 3-4: Stdio transport (wrap existing)
├── Day 5: Transport registry

Week 5:
├── Day 1-2: SSE transport implementation
├── Day 3-4: HTTP transport implementation
├── Day 5: TLS support, health endpoints
```

**Interface Contract:**
```go
// internal/transport/transport.go
type Transport interface {
    Serve(ctx context.Context, handler RequestHandler) error
    Close() error
    Info() TransportInfo
}

type TransportInfo struct {
    Name     string
    Protocol string // "stdio", "http", "sse", "grpc"
    Address  string // "" for stdio, "localhost:8080" for HTTP
}

type TransportRegistry interface {
    Register(name string, factory TransportFactory)
    Get(name string) (Transport, error)
    List() []string
}
```

**Edge Cases:**
- [ ] Stdio: handle broken pipe gracefully
- [ ] HTTP: CORS headers for browser clients
- [ ] SSE: reconnection with event ID
- [ ] TLS: certificate rotation without restart
- [ ] Health: `/health` (liveness), `/ready` (readiness)

#### A4: Provider Registry (1 week)

```
Week 6:
├── Day 1-2: ToolProvider interface
├── Day 3: Provider registry
├── Day 4: Refactor existing tools as providers
├── Day 5: Custom provider registration docs
```

**Interface Contract:**
```go
// internal/provider/provider.go
type ToolProvider interface {
    Name() string
    Tool() *mcp.Tool
    Handle(ctx context.Context, input json.RawMessage) (any, error)
}

type ProviderRegistry interface {
    Register(provider ToolProvider) error
    Get(name string) (ToolProvider, error)
    List() []ToolProvider
    Unregister(name string) error
}
```

**Edge Cases:**
- [ ] Duplicate registration: return error, don't overwrite
- [ ] Dynamic registration: emit events for discovery
- [ ] Provider panic: recover, log, return error
- [ ] Slow provider: context deadline enforced

#### A5: Backend Registry (2 weeks)

```
Week 7:
├── Day 1-2: Backend interface definition
├── Day 3-4: Local backend (existing behavior)
├── Day 5: Backend registry

Week 8:
├── Day 1-2: MCP backend (connect to MCP servers)
├── Day 3-4: HTTP backend (REST API tools)
├── Day 5: Aggregator (merge tools from all backends)
```

**Interface Contract:**
```go
// internal/backend/backend.go
type Backend interface {
    Kind() string              // "local", "mcp", "http"
    Name() string              // Instance name
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    ListTools(ctx context.Context) ([]*toolmodel.Tool, error)
    Execute(ctx context.Context, tool string, input any) (any, error)
    Health(ctx context.Context) HealthStatus
}

type BackendRegistry interface {
    Register(backend Backend) error
    Get(name string) (Backend, error)
    ListByKind(kind string) []Backend
    Aggregate() AggregatedBackend
}
```

**Edge Cases:**
- [ ] Backend offline: mark unhealthy, exclude from routing
- [ ] Tool name collision: namespace with backend name
- [ ] Slow backend: per-backend timeout
- [ ] Backend restart: re-register tools, emit events

---

### Stream B: Protocol Layer

**Goal**: Protocol-agnostic tool exposure with composable toolsets.

| Phase | Work Package | Duration | Deliverables |
|-------|-------------|----------|--------------|
| B1 | tooladapter Core | 2 weeks | Adapter interface, MCP adapter |
| B2 | Additional Adapters | 1 week | OpenAI, Anthropic, LangChain |
| B3 | toolset Core | 2 weeks | Toolset builder, filtering |
| B4 | Multi-Transport | 1 week | Expose via MCP, REST, direct |

**Total: 6 weeks**

#### B1: tooladapter Core (2 weeks)

```
Week 9:
├── Day 1-2: CanonicalTool type definition
├── Day 3-4: Adapter interface
├── Day 5: MCP adapter (bidirectional)

Week 10:
├── Day 1-2: Schema conversion utilities
├── Day 3-4: Adapter registry
├── Day 5: Unit tests, documentation
```

**Interface Contract:**
```go
// tooladapter/canonical.go
type CanonicalTool struct {
    ID          string
    Namespace   string
    Name        string
    Version     semver.Version
    Description string
    InputSchema *JSONSchema
    OutputSchema *JSONSchema
    Handler     ToolHandler
    SourceFormat string
    SourceMeta   map[string]any
}

// tooladapter/adapter.go
type Adapter interface {
    Name() string
    ToCanonical(raw any) (*CanonicalTool, error)
    FromCanonical(tool *CanonicalTool) (any, error)
    SupportsFeature(feature SchemaFeature) bool
}
```

**Edge Cases:**
- [ ] Schema feature not supported: strip with warning
- [ ] Conversion loss: preserve original in SourceMeta
- [ ] Bidirectional round-trip: verify no data loss
- [ ] Nil handling: optional fields default to nil/zero

#### B2: Additional Adapters (1 week)

```
Week 11:
├── Day 1-2: OpenAI adapter (strict mode support)
├── Day 3: Anthropic adapter
├── Day 4: LangChain adapter
├── Day 5: OpenAPI adapter (import only)
```

**Edge Cases:**
- [ ] OpenAI strict mode: reject unsupported features
- [ ] Anthropic input_schema vs inputSchema naming
- [ ] LangChain Zod schemas: convert to JSON Schema
- [ ] OpenAPI discriminators: flatten to anyOf

#### B3: toolset Core (2 weeks)

```
Week 12:
├── Day 1-2: Toolset type definition
├── Day 3-4: Builder pattern with fluent API
├── Day 5: Filter predicates

Week 13:
├── Day 1-2: Access control policies
├── Day 3-4: Integration with toolindex
├── Day 5: Integration with toolrun
```

**Interface Contract:**
```go
// toolset/builder.go
type Builder struct { ... }

func NewBuilder(name string) *Builder
func (b *Builder) FromRegistry(reg *Registry) *Builder
func (b *Builder) WithNamespace(ns string) *Builder
func (b *Builder) WithTags(tags ...string) *Builder
func (b *Builder) WithTools(ids ...string) *Builder
func (b *Builder) ExcludeTools(ids ...string) *Builder
func (b *Builder) WithPolicy(p *AccessPolicy) *Builder
func (b *Builder) Build() (*Toolset, error)
```

**Edge Cases:**
- [ ] Empty toolset: valid, return empty list
- [ ] Conflicting filters: AND logic (all must match)
- [ ] Tool removed after build: refresh on demand
- [ ] Circular exclude: no-op, tool already excluded

#### B4: Multi-Transport (1 week)

```
Week 14:
├── Day 1-2: MCP exposure (enhanced)
├── Day 3: REST API exposure
├── Day 4: Direct Go client
├── Day 5: Documentation, examples
```

**Edge Cases:**
- [ ] MCP version mismatch: negotiate or reject
- [ ] REST pagination: cursor-based for large toolsets
- [ ] Go client: context cancellation propagation

---

### Stream C: Cross-Cutting

**Goal**: Production-ready cross-cutting concerns.

| Phase | Work Package | Duration | Deliverables |
|-------|-------------|----------|--------------|
| C1 | toolcache | 2 weeks | Cache interface, Redis/Memory |
| C2 | toolobserve | 2 weeks | OpenTelemetry integration |
| C3 | toolversion | 2 weeks | Semantic versioning, negotiation |
| C4 | toolresilience | 2 weeks | Circuit breaker, retry |
| C5 | toolhealth | 1 week | Health checks |
| C6 | toolaudit | 2 weeks | Audit logging |
| C7 | toolsecrets | 2 weeks | Secrets management |
| C8 | toolflags + toolpressure | 2 weeks | Feature flags, backpressure |

**Total: 15 weeks** (can parallelize C1-C3 and C4-C8)

#### C1: toolcache (2 weeks)

**Interface Contract:**
```go
// toolcache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context, pattern string) error
    Stats() CacheStats
    Close() error
}
```

**Implementations:**
- `MemoryCache` - LRU eviction, process-local
- `RedisCache` - Distributed, TTL-based
- `LayeredCache` - L1 memory + L2 Redis

**Integration Points:**
- `toolindex`: Cache tool lookups (`index:tool:{id}`)
- `tooldocs`: Cache documentation (`docs:{id}:{level}`)
- `toolsearch`: Cache search results (`search:{hash}`)
- `tooladapter`: Cache schema conversions (`schema:{format}:{id}`)

#### C2: toolobserve (2 weeks)

**Interface Contract:**
```go
// toolobserve/observe.go
type Observer interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
    RecordMetric(name string, value float64, labels map[string]string)
    RecordError(ctx context.Context, err error)
}

type Span interface {
    SetAttribute(key string, value any)
    AddEvent(name string, attrs map[string]any)
    RecordError(err error)
    End()
}
```

**Integration Points:**
- Middleware: auto-instrument all tool calls
- Per-backend: trace backend latency
- Per-tool: trace individual tool execution

#### C3: toolversion (2 weeks)

**Interface Contract:**
```go
// toolversion/registry.go
type VersionedToolRegistry interface {
    Register(tool *VersionedTool) error
    Resolve(name string, constraint string) (*VersionedTool, error)
    ListVersions(name string) []VersionInfo
    Deprecate(name string, version string, message string, sunset time.Time) error
}
```

**Edge Cases:**
- [ ] No matching version: return ErrNoCompatibleVersion
- [ ] Deprecated tool used: log warning, emit metric
- [ ] Sunset reached: remove from registry
- [ ] Version constraint invalid: return parse error

#### C4: toolresilience (2 weeks)

**Interface Contract:**
```go
// toolresilience/resilience.go
type CircuitBreaker interface {
    Execute(ctx context.Context, fn func() error) error
    State() CircuitState
    Reset()
}

type RetryPolicy interface {
    Execute(ctx context.Context, fn func() error) error
    ShouldRetry(err error, attempt int) bool
}

type Bulkhead interface {
    Acquire(ctx context.Context) error
    Release()
    Available() int
}
```

**Edge Cases:**
- [ ] Circuit open: fail fast without calling backend
- [ ] Retry on non-idempotent: disabled by default
- [ ] Bulkhead full: reject with 503
- [ ] Jitter calculation: prevent thundering herd

---

### Stream D: Enterprise

**Goal**: Enterprise-grade features for production deployments.

| Phase | Work Package | Duration | Deliverables |
|-------|-------------|----------|--------------|
| D1 | Multi-Tenancy Core | 2 weeks | Tenant model, resolvers |
| D2 | Tenant Middleware | 1 week | Rate limit, filtering, audit |
| D3 | Tenant Storage | 1 week | Redis/Postgres store |
| D4 | toolsemantic | 3 weeks | Vector search, hybrid |
| D5 | toolresource | 2 weeks | MCP Resources support |
| D6 | toolgateway | 3 weeks | Auth proxy, analytics |

**Total: 12 weeks** (can parallelize D1-D3 and D4-D6)

#### D1: Multi-Tenancy Core (2 weeks)

**Interface Contract:**
```go
// multi-tenancy
type Tenant struct {
    ID       string
    Name     string
    Tier     TenantTier // free, pro, enterprise
    Metadata map[string]any
}

type TenantResolver interface {
    Resolve(ctx context.Context, req *Request) (*TenantContext, error)
}

type TenantContext struct {
    Tenant      *Tenant
    Permissions []string
    Config      *TenantConfig
    Quotas      *TenantQuotas
}
```

**Isolation Strategies:**
1. **Shared**: Logical isolation via middleware (default)
2. **Namespace**: Tool namespace per tenant
3. **Process**: Dedicated process per tenant

**Edge Cases:**
- [ ] No tenant header: use default tenant or reject
- [ ] Tenant not found: return 401/403
- [ ] Quota exceeded: return 429 with retry-after
- [ ] Tenant config change: propagate to active sessions

#### D4: toolsemantic - Advanced Semantic Search (3 weeks)

> **Research-Driven Design**: This specification is based on 2025 industry best practices for production RAG systems, including hybrid search, knowledge graphs, hierarchical chunking, agentic retrieval, and late interaction models.

```
Week 16:
├── Day 1-2: Core embedding and vector interfaces
├── Day 3-4: Hybrid search (BM25 + vector fusion)
├── Day 5: Reranker integration

Week 17:
├── Day 1-2: GraphRAG knowledge graph layer
├── Day 3-4: Hierarchical/tree chunking
├── Day 5: Context tree traversal

Week 18:
├── Day 1-2: Agentic RAG (query expansion, multi-query)
├── Day 3-4: ColBERT/late interaction support
├── Day 5: Integration tests, benchmarks
```

**Architecture Overview:**

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          AGENTIC RAG LAYER                                  │
│  Query Expansion → Multi-Query → Self-Reasoning → Adaptive Retrieval       │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                           HYBRID SEARCH                                     │
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │  BM25/TF-IDF│    │   Vector    │    │   GraphRAG  │                     │
│  │  (toolsearch)│   │  (Dense)    │    │ (Knowledge) │                     │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                     │
│         │                  │                  │                             │
│         └──────────────────┼──────────────────┘                             │
│                            │                                                │
│                   ┌────────▼────────┐                                       │
│                   │  Rank Fusion    │                                       │
│                   │ (RRF / Weighted)│                                       │
│                   └────────┬────────┘                                       │
│                            │                                                │
│                   ┌────────▼────────┐                                       │
│                   │   Reranker      │                                       │
│                   │ (Cross-Encoder) │                                       │
│                   └─────────────────┘                                       │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                        INDEX LAYER                                          │
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  BM25 Index │    │ HNSW/IVF    │    │  Knowledge  │    │ Hierarchical│  │
│  │(toolsearch) │    │Vector Index │    │   Graph     │    │    Tree     │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
```

**Core Interface Contracts:**

```go
// toolsemantic/embedding.go

// Embedder generates vector embeddings from text
type Embedder interface {
    // Embed generates embedding for a single text
    Embed(ctx context.Context, text string) ([]float32, error)

    // EmbedBatch generates embeddings for multiple texts (more efficient)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Dimensions returns the embedding vector size
    Dimensions() int

    // Model returns the model identifier
    Model() string
}

// EmbedderFactory creates embedders from configuration
type EmbedderFactory interface {
    Create(config EmbedderConfig) (Embedder, error)
}

// EmbedderConfig configures an embedding provider
type EmbedderConfig struct {
    Provider   string            // "openai", "cohere", "local", "ollama"
    Model      string            // "text-embedding-3-small", "embed-english-v3.0"
    Dimensions int               // Override dimensions (for some models)
    Options    map[string]any    // Provider-specific options
}
```

```go
// toolsemantic/vector.go

// VectorIndex provides vector similarity search
type VectorIndex interface {
    // Index adds vectors to the index
    Index(ctx context.Context, docs []VectorDocument) error

    // Search finds similar vectors
    Search(ctx context.Context, query []float32, opts SearchOptions) ([]VectorResult, error)

    // Delete removes documents from index
    Delete(ctx context.Context, ids []string) error

    // Stats returns index statistics
    Stats() IndexStats
}

type VectorDocument struct {
    ID        string
    Vector    []float32
    Metadata  map[string]any
    Content   string  // Optional, for hybrid storage
}

type VectorResult struct {
    ID       string
    Score    float32  // Similarity score (higher = more similar)
    Distance float32  // Distance (lower = more similar)
    Metadata map[string]any
    Content  string
}

type SearchOptions struct {
    TopK     int               // Number of results
    MinScore float32           // Minimum similarity threshold
    Filter   map[string]any    // Metadata filters
}

// VectorIndexFactory creates vector indices
type VectorIndexFactory interface {
    Create(config VectorIndexConfig) (VectorIndex, error)
}

type VectorIndexConfig struct {
    Backend     string  // "memory", "faiss", "qdrant", "pinecone", "chromadb"
    Dimensions  int
    Metric      string  // "cosine", "euclidean", "dot"
    IndexType   string  // "flat", "hnsw", "ivf"

    // HNSW-specific
    M               int  // Max connections per layer
    EfConstruction  int  // Construction time quality/speed tradeoff
    EfSearch        int  // Search time quality/speed tradeoff
}
```

```go
// toolsemantic/hybrid.go

// HybridSearcher combines multiple retrieval methods
type HybridSearcher interface {
    // Search performs hybrid search across all configured retrievers
    Search(ctx context.Context, query string, opts HybridOptions) ([]SearchResult, error)

    // Explain returns detailed scoring breakdown
    Explain(ctx context.Context, query string, docID string) (*ScoreExplanation, error)
}

type HybridOptions struct {
    TopK            int

    // Retrieval weights (must sum to 1.0)
    BM25Weight      float32  // Weight for BM25/keyword results
    VectorWeight    float32  // Weight for vector similarity results
    GraphWeight     float32  // Weight for knowledge graph results

    // Fusion strategy
    FusionMethod    FusionMethod  // RRF, WeightedSum, Borda

    // Reranking
    EnableReranker  bool
    RerankerTopK    int  // Rerank top N results

    // Filters
    Namespaces      []string
    Tags            []string
    MetadataFilters map[string]any
}

type FusionMethod string

const (
    FusionRRF         FusionMethod = "rrf"          // Reciprocal Rank Fusion
    FusionWeightedSum FusionMethod = "weighted_sum" // Weighted score combination
    FusionBorda       FusionMethod = "borda"        // Borda count ranking
)

type SearchResult struct {
    ID          string
    Score       float32
    Content     string
    Metadata    map[string]any

    // Detailed scores
    BM25Score   float32
    VectorScore float32
    GraphScore  float32
    RerankerScore float32

    // Source tracking
    Sources     []string  // Which retrievers contributed
}

type ScoreExplanation struct {
    FinalScore    float32
    FusionMethod  FusionMethod
    Components    []ScoreComponent
    RerankerBoost float32
}

type ScoreComponent struct {
    Source      string   // "bm25", "vector", "graph"
    RawScore    float32
    Weight      float32
    WeightedScore float32
    Rank        int
}
```

```go
// toolsemantic/reranker.go

// Reranker performs cross-encoder reranking on retrieved results
type Reranker interface {
    // Rerank scores query-document pairs using cross-encoder
    Rerank(ctx context.Context, query string, docs []RerankerInput) ([]RerankerResult, error)

    // Model returns the reranker model identifier
    Model() string
}

type RerankerInput struct {
    ID      string
    Content string
}

type RerankerResult struct {
    ID    string
    Score float32  // Cross-encoder relevance score
}

// RerankerConfig configures a reranker
type RerankerConfig struct {
    Provider  string  // "cohere", "jina", "local", "colbert"
    Model     string  // "rerank-english-v3.0", "ms-marco-MiniLM"
    MaxTokens int     // Max input tokens
    BatchSize int     // Batch size for efficiency
}
```

```go
// toolsemantic/graph.go

// KnowledgeGraph provides graph-based retrieval (GraphRAG)
type KnowledgeGraph interface {
    // Build extracts entities and relationships from documents
    Build(ctx context.Context, docs []GraphDocument) error

    // Query retrieves relevant subgraph for a query
    Query(ctx context.Context, query string, opts GraphQueryOptions) (*GraphResult, error)

    // GetEntity retrieves a specific entity
    GetEntity(ctx context.Context, id string) (*Entity, error)

    // TraverseRelations walks relationships from an entity
    TraverseRelations(ctx context.Context, entityID string, depth int) ([]Relationship, error)
}

type GraphDocument struct {
    ID      string
    Content string
    Metadata map[string]any
}

type GraphQueryOptions struct {
    MaxEntities     int
    MaxRelationships int
    TraversalDepth  int
    IncludeCommunities bool  // Include community summaries
}

type GraphResult struct {
    Entities      []Entity
    Relationships []Relationship
    Communities   []Community     // High-level community summaries
    Summary       string          // Graph-generated summary
    Score         float32
}

type Entity struct {
    ID          string
    Name        string
    Type        string            // "tool", "namespace", "concept"
    Description string
    Attributes  map[string]any
    Embedding   []float32
}

type Relationship struct {
    ID       string
    Source   string  // Entity ID
    Target   string  // Entity ID
    Type     string  // "uses", "depends_on", "similar_to"
    Weight   float32
    Metadata map[string]any
}

type Community struct {
    ID       string
    Name     string
    Summary  string
    Entities []string  // Entity IDs in this community
    Level    int       // Hierarchy level (0 = most granular)
}
```

```go
// toolsemantic/hierarchical.go

// HierarchicalChunker implements tree-structured document chunking
type HierarchicalChunker interface {
    // Chunk splits document into hierarchical tree structure
    Chunk(ctx context.Context, doc *Document) (*ChunkTree, error)

    // Retrieve navigates tree to find relevant chunks
    Retrieve(ctx context.Context, tree *ChunkTree, query string, opts TreeOptions) ([]Chunk, error)
}

type Document struct {
    ID       string
    Content  string
    Metadata map[string]any
}

type ChunkTree struct {
    Root     *ChunkNode
    Depth    int
    NumNodes int
}

type ChunkNode struct {
    ID        string
    Level     int           // 0 = leaf, higher = more abstract
    Content   string        // Text content or summary
    Embedding []float32
    Children  []*ChunkNode
    Parent    *ChunkNode

    // For RAPTOR-style summarization
    Summary   string        // LLM-generated summary at this level
}

type Chunk struct {
    ID        string
    Content   string
    Level     int
    Score     float32
    Path      []string  // Path from root to this chunk
}

type TreeOptions struct {
    MaxChunks     int
    MinLevel      int  // Minimum abstraction level to consider
    MaxLevel      int  // Maximum abstraction level

    // Traversal strategy
    Strategy      TraversalStrategy
}

type TraversalStrategy string

const (
    TraversalTopDown    TraversalStrategy = "top_down"    // Start from summaries
    TraversalBottomUp   TraversalStrategy = "bottom_up"   // Start from leaves
    TraversalCollapsed  TraversalStrategy = "collapsed"   // RAPTOR-style collapsed tree
)
```

```go
// toolsemantic/agentic.go

// AgenticRetriever implements LLM-enhanced retrieval strategies
type AgenticRetriever interface {
    // Retrieve performs agentic retrieval with optional query expansion
    Retrieve(ctx context.Context, query string, opts AgenticOptions) (*AgenticResult, error)
}

type AgenticOptions struct {
    // Query expansion
    EnableQueryExpansion  bool
    MaxExpandedQueries    int

    // Multi-query
    EnableMultiQuery      bool
    QueryVariations       int

    // Self-reasoning (Step-Back, Chain-of-Thought)
    EnableSelfReasoning   bool
    ReasoningDepth        int

    // Iterative retrieval
    EnableIterative       bool
    MaxIterations         int
    StopCondition         func(results []SearchResult) bool

    // Base retriever to use
    BaseRetriever         HybridSearcher
}

type AgenticResult struct {
    Results          []SearchResult

    // Agentic metadata
    ExpandedQueries  []string
    ReasoningChain   []ReasoningStep
    Iterations       int

    // Quality signals
    Confidence       float32
    CoverageScore    float32
}

type ReasoningStep struct {
    Step        int
    Query       string
    Reasoning   string   // LLM explanation
    ResultCount int
    TopScore    float32
}
```

```go
// toolsemantic/colbert.go

// ColBERTIndex implements late interaction retrieval
type ColBERTIndex interface {
    // Index adds documents with token-level embeddings
    Index(ctx context.Context, docs []ColBERTDocument) error

    // Search performs MaxSim-based late interaction search
    Search(ctx context.Context, query string, opts ColBERTOptions) ([]ColBERTResult, error)
}

type ColBERTDocument struct {
    ID           string
    Content      string
    TokenEmbeddings [][]float32  // Per-token embeddings
}

type ColBERTOptions struct {
    TopK         int
    NCells       int   // For PLAID centroid-based search
    NDocs        int   // Candidate docs before rescoring
}

type ColBERTResult struct {
    ID           string
    Score        float32  // MaxSim score
    TokenScores  []TokenMatch  // Token-level matching details
}

type TokenMatch struct {
    QueryToken  string
    DocToken    string
    Score       float32
}
```

**Integration with Existing Libraries:**

```go
// toolsemantic/integration.go

// SemanticSearcher wraps toolsearch.Searcher with semantic capabilities
type SemanticSearcher struct {
    // Existing BM25 searcher
    bm25     toolsearch.Searcher

    // New semantic components
    embedder     Embedder
    vectorIndex  VectorIndex
    reranker     Reranker
    graph        KnowledgeGraph

    // Configuration
    config       SemanticConfig
}

// Search implements toolindex.Searcher interface
func (s *SemanticSearcher) Search(ctx context.Context, query string, opts toolsearch.Options) ([]toolsearch.Result, error) {
    // Perform hybrid search internally
    hybridResults, err := s.hybridSearch(ctx, query, HybridOptions{
        TopK:         opts.Limit,
        BM25Weight:   s.config.BM25Weight,
        VectorWeight: s.config.VectorWeight,
        GraphWeight:  s.config.GraphWeight,
        EnableReranker: s.config.EnableReranker,
    })
    if err != nil {
        return nil, err
    }

    // Convert to toolsearch.Result format
    return convertResults(hybridResults), nil
}

// NewSemanticSearcher creates a semantic searcher that wraps BM25
func NewSemanticSearcher(bm25 toolsearch.Searcher, opts ...SemanticOption) (*SemanticSearcher, error) {
    // ...
}
```

**Configuration Example:**

```yaml
# metatools.yaml
semantic:
  enabled: true

  embedder:
    provider: openai
    model: text-embedding-3-small
    dimensions: 1536

  vector_index:
    backend: memory  # memory, faiss, qdrant, chromadb
    metric: cosine
    index_type: hnsw
    hnsw:
      m: 16
      ef_construction: 200
      ef_search: 100

  hybrid:
    bm25_weight: 0.4
    vector_weight: 0.5
    graph_weight: 0.1
    fusion_method: rrf

  reranker:
    enabled: true
    provider: cohere
    model: rerank-english-v3.0
    top_k: 20

  graph:
    enabled: false  # Experimental
    entity_types: ["tool", "namespace", "concept"]

  agentic:
    enabled: false  # Experimental
    query_expansion: true
    max_iterations: 3
```

**Performance Characteristics:**

| Component | Latency | Memory | Accuracy Impact |
|-----------|---------|--------|-----------------|
| BM25 (baseline) | ~8ms | Low | Baseline (78%) |
| Vector (HNSW) | ~12ms | Medium | +8% (86%) |
| Hybrid (BM25+Vector) | ~25ms | Medium | +16% (94%) |
| + Reranker | +50-100ms | Low | +3% (97%) |
| + GraphRAG | +100-200ms | High | +2% (context) |
| + Agentic | +500-2000ms | Low | Variable |

**Research-Based Accuracy Benchmarks (Tool Discovery):**

| Method | Accuracy @ 50 tools | @ 200 tools | @ 500 tools |
|--------|---------------------|-------------|-------------|
| BM25 only | 78% | 71% | 63% |
| Vector only | 82% | 76% | 69% |
| Hybrid (BM25+Vector) | 94% | 89% | 84% |
| Hybrid + Reranker | 97% | 94% | 91% |
| Hybrid + Reranker + Graph | 98% | 96% | 93% |

**Edge Cases:**

- [ ] Empty embedding: reject with ErrEmptyInput
- [ ] Dimension mismatch: validate at index time
- [ ] Reranker timeout: fall back to hybrid-only results
- [ ] Graph disconnected: handle isolated entities
- [ ] Query expansion loop: max iteration limit
- [ ] OOM on large indices: streaming/pagination support

---

### Stream E: Agent Skills

**Goal**: Higher-level agent capabilities that compose tools into reusable workflows and behaviors.

> **Design Insight**: Skills sit above tools in the abstraction hierarchy. While tools are atomic operations (search, execute, describe), skills are reusable agent behaviors that orchestrate multiple tools (research-and-summarize, debug-and-fix, deploy-with-validation).

| Phase | Work Package | Duration | Deliverables |
|-------|-------------|----------|--------------|
| E1 | Skill Core | 2 weeks | Skill interface, registry, manifest |
| E2 | Skill Composition | 1 week | Workflow DSL, step orchestration |
| E3 | Skill Execution | 1 week | Runtime, context propagation, rollback |

**Total: 4 weeks** (can start after Stream B Protocol Layer)

#### Abstraction Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                        AGENTS                                │
│  (Autonomous decision makers: Claude, GPT, custom agents)   │
└────────────────────────────┬────────────────────────────────┘
                             │ use
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                        SKILLS (NEW)                          │
│  Higher-level behaviors: research, debug, deploy, review    │
│  Composed from multiple tools with workflow logic           │
└────────────────────────────┬────────────────────────────────┘
                             │ orchestrate
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                        TOOLSETS                              │
│  Curated tool collections: dev-tools, search-tools          │
│  Filtered by namespace, tags, access policy                 │
└────────────────────────────┬────────────────────────────────┘
                             │ contain
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                         TOOLS                                │
│  Atomic operations: search_tools, run_tool, describe_tool   │
│  Single-purpose, composable primitives                      │
└─────────────────────────────────────────────────────────────┘
```

#### E1: Skill Core (2 weeks)

```
Week 18:
├── Day 1-2: Skill interface definition
├── Day 3-4: SkillManifest type (A2A-aligned)
├── Day 5: SkillRegistry implementation

Week 19:
├── Day 1-2: Skill discovery and advertisement
├── Day 3-4: Skill versioning integration
├── Day 5: Unit tests, documentation
```

**Interface Contract:**
```go
// toolskill/skill.go
type Skill interface {
    // Identity
    ID() string
    Name() string
    Version() semver.Version

    // Manifest for discovery/advertisement
    Manifest() *SkillManifest

    // Execution
    Execute(ctx context.Context, input SkillInput) (*SkillOutput, error)

    // Introspection
    RequiredTools() []string
    Steps() []StepDefinition
}

// SkillManifest describes skill capabilities (A2A-aligned)
type SkillManifest struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    InputSchema *jsonschema.Schema `json:"inputSchema"`
    OutputSchema *jsonschema.Schema `json:"outputSchema,omitempty"`
    Tags        []string          `json:"tags"`

    // Dependencies
    RequiredTools []string        `json:"requiredTools"`
    RequiredSkills []string       `json:"requiredSkills,omitempty"`

    // Execution hints
    EstimatedSteps int            `json:"estimatedSteps"`
    Idempotent     bool           `json:"idempotent"`
    SupportsPause  bool           `json:"supportsPause"`
}

// SkillRegistry manages skill discovery and lookup
type SkillRegistry interface {
    Register(skill Skill) error
    Get(id string) (Skill, error)
    GetByName(name string, versionConstraint string) (Skill, error)
    List() []SkillManifest
    ListByTag(tag string) []SkillManifest
    Unregister(id string) error

    // Discovery for A2A
    Advertise() []SkillManifest
}
```

**Design Patterns from Research:**

| Framework | Pattern | Adaptation for toolskill |
|-----------|---------|--------------------------|
| LangChain | Chains (sequential) | `SequentialSkill` with steps |
| LangChain | Agents (decision) | `AdaptiveSkill` with branching |
| CrewAI | Crews (role-based) | `CompositeSkill` with sub-skills |
| CrewAI | Flows (control) | `WorkflowSkill` with explicit flow |
| AutoGen | Conversation | `ConversationalSkill` with memory |

#### SKILL.md Standard Format Support

The toolskill library will implement the **Agent Skills Open Standard** (adopted by Claude Code, OpenAI Codex CLI, ChatGPT, and Google Antigravity) to ensure cross-platform compatibility.

**Standard SKILL.md Structure:**
```markdown
---
name: skill-name-with-hyphens
description: Use when [specific triggering conditions and symptoms]
metadata:
  author: optional-author
  version: "1.0.0"
  argument-hint: optional-hint
---

# Skill Name

## Overview
What is this? Core principle in 1-2 sentences.

## When to Use
Symptoms and use cases.

## How It Works
Step-by-step process.

[Additional sections as needed]
```

**Standard Directory Structure:**
```
skills/
  skill-name/
    SKILL.md              # Main entry point (required)
    references/           # Optional supporting files
      summary.md
      files.md
    scripts/              # Optional executable tools
      validate.sh
```

**Discovery Locations (Claude Code compatible):**
- User settings: `~/.claude/skills/` or `~/.config/claude/skills/`
- Project settings: `.claude/skills/`
- Plugin-provided: `plugins/*/skills/`
- Nested directories: Automatic discovery in subdirectories

**Interface Contract for SKILL.md Parsing:**
```go
// toolskill/skillmd/parser.go

// SkillMD represents a parsed SKILL.md file
type SkillMD struct {
    // YAML Frontmatter
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    License     string            `yaml:"license,omitempty"`
    Metadata    map[string]any    `yaml:"metadata,omitempty"`

    // Parsed Content
    Content     string            // Raw markdown content
    Overview    string            // Extracted ## Overview section
    WhenToUse   string            // Extracted ## When to Use section
    HowItWorks  string            // Extracted ## How It Works section
    Sections    map[string]string // All other sections
}

// Parser reads and parses SKILL.md files
type Parser interface {
    Parse(path string) (*SkillMD, error)
    ParseBytes(data []byte) (*SkillMD, error)
    Validate(skill *SkillMD) []ValidationError
}

// Generator creates SKILL.md files from skills
type Generator interface {
    Generate(skill Skill) ([]byte, error)
    GenerateWithTemplate(skill Skill, template string) ([]byte, error)
}

// Discovery finds skills in standard locations
type Discovery interface {
    // Scan standard locations for skills
    ScanUserSkills() ([]SkillMD, error)           // ~/.claude/skills/
    ScanProjectSkills(root string) ([]SkillMD, error) // .claude/skills/
    ScanPluginSkills(pluginDir string) ([]SkillMD, error)

    // Watch for changes
    Watch(ctx context.Context, locations []string) (<-chan SkillEvent, error)
}
```

**Cross-Platform Export:**
```go
// toolskill/export/exporter.go

type Exporter interface {
    // Export to standard SKILL.md format (Claude, Codex, ChatGPT)
    ToSkillMD(skill Skill) (*SkillMD, error)

    // Export to platform-specific formats
    ToZIP(skill Skill) ([]byte, error)          // Claude/OpenAI upload
    ToTarGz(skill Skill) ([]byte, error)        // Gemini upload

    // Batch export
    ExportDirectory(skills []Skill, path string) error
}

// Import from SKILL.md files
type Importer interface {
    // Import single skill
    FromSkillMD(md *SkillMD) (Skill, error)

    // Batch import from directory
    FromDirectory(path string) ([]Skill, error)

    // Import from skill repository URL
    FromRepository(url string) ([]Skill, error)
}
```

**Key Design Decisions:**

1. **Frontmatter Limited to name + description**: Per standard, only these fields are guaranteed. Additional metadata is optional.

2. **Description Format**: Must start with "Use when..." to optimize for Claude's skill discovery algorithm.

3. **Progressive Disclosure**: Skills loaded on-demand based on description matching, not pre-loaded (reduces context window usage).

4. **Flat Namespace**: All skills in searchable namespace for discovery. No nested hierarchies.

5. **Token Efficiency**: Frequently-loaded skills should be <200 words. Reference material in separate files.

#### E2: Skill Composition (1 week)

```
Week 20:
├── Day 1-2: Step definition DSL
├── Day 3: Sequential composition
├── Day 4: Parallel composition
├── Day 5: Conditional branching
```

**Interface Contract:**
```go
// toolskill/step.go
type StepDefinition struct {
    ID          string
    Name        string
    Tool        string              // Tool to execute
    InputMapper func(SkillContext) any  // Map skill input to tool input
    OutputMapper func(any) any      // Transform tool output
    Condition   func(SkillContext) bool // Skip if returns false
    OnError     ErrorHandler
}

// toolskill/builder.go
type SkillBuilder struct { ... }

func NewSkillBuilder(name string) *SkillBuilder

// Sequential steps
func (b *SkillBuilder) Step(name string, tool string) *StepBuilder

// Parallel execution
func (b *SkillBuilder) Parallel(steps ...*StepBuilder) *SkillBuilder

// Conditional branching
func (b *SkillBuilder) Branch(condition func(SkillContext) bool,
                              ifTrue *SkillBuilder,
                              ifFalse *SkillBuilder) *SkillBuilder

// Sub-skill composition
func (b *SkillBuilder) UseSkill(skillID string) *SkillBuilder

// Build final skill
func (b *SkillBuilder) Build() (Skill, error)
```

**Example: Research-and-Summarize Skill**
```go
researchSkill := NewSkillBuilder("research-and-summarize").
    WithDescription("Research a topic and provide a summary").
    WithInputSchema(researchInputSchema).

    // Step 1: Search for relevant tools
    Step("discover", "search_tools").
        WithInput(func(ctx SkillContext) any {
            return map[string]any{"query": ctx.Input["topic"]}
        }).

    // Step 2: Get documentation for top results
    Parallel(
        Step("docs-1", "describe_tool").WithInput(fromResult(0)),
        Step("docs-2", "describe_tool").WithInput(fromResult(1)),
        Step("docs-3", "describe_tool").WithInput(fromResult(2)),
    ).

    // Step 3: Synthesize findings
    Step("synthesize", "run_tool").
        WithInput(func(ctx SkillContext) any {
            return synthesizeInputs(ctx)
        }).

    Build()
```

#### E3: Skill Execution (1 week)

```
Week 21:
├── Day 1-2: Skill runtime with context
├── Day 3: Progress tracking, checkpoints
├── Day 4: Rollback on failure
├── Day 5: Integration with toolobserve
```

**Interface Contract:**
```go
// toolskill/runtime.go
type SkillRuntime interface {
    Execute(ctx context.Context, skill Skill, input SkillInput) (*SkillOutput, error)
    ExecuteAsync(ctx context.Context, skill Skill, input SkillInput) (ExecutionID, error)

    // Progress and control
    GetStatus(execID ExecutionID) (*ExecutionStatus, error)
    Pause(execID ExecutionID) error
    Resume(execID ExecutionID) error
    Cancel(execID ExecutionID) error

    // Checkpoints for long-running skills
    Checkpoint(execID ExecutionID) (*Checkpoint, error)
    RestoreFromCheckpoint(checkpoint *Checkpoint) (ExecutionID, error)
}

// SkillContext provides execution context to steps
type SkillContext struct {
    Input      map[string]any      // Original skill input
    Results    map[string]any      // Results from previous steps
    Metadata   map[string]any      // Execution metadata
    Toolset    *toolset.Toolset    // Available tools
    Logger     Logger              // Structured logging
    Tracer     trace.Tracer        // OpenTelemetry tracing
}

// ExecutionStatus tracks skill progress
type ExecutionStatus struct {
    ID            ExecutionID
    SkillID       string
    State         ExecutionState  // pending, running, paused, completed, failed
    CurrentStep   string
    CompletedSteps []string
    Progress      float64         // 0.0 to 1.0
    StartedAt     time.Time
    Error         error
}
```

**Edge Cases:**
- [ ] Step timeout: configurable per-step, fail or skip
- [ ] Tool not found: fail skill with ErrMissingTool
- [ ] Parallel step failure: configurable (fail-fast or continue)
- [ ] Checkpoint restore: validate skill version compatibility
- [ ] Circular skill dependencies: detect and reject at registration

#### Skill-Tool Integration

```go
// Expose skills via MCP as special tools
type SkillToolProvider struct {
    registry SkillRegistry
    runtime  SkillRuntime
}

func (p *SkillToolProvider) Tools() []*mcp.Tool {
    var tools []*mcp.Tool
    for _, manifest := range p.registry.List() {
        tools = append(tools, &mcp.Tool{
            Name: "skill:" + manifest.Name,
            Description: "[SKILL] " + manifest.Description,
            InputSchema: manifest.InputSchema,
        })
    }
    return tools
}
```

**MCP Tool Naming Convention:**
- Regular tools: `search_tools`, `run_tool`
- Skills: `skill:research`, `skill:debug`, `skill:deploy`

---

## 4. Phase Breakdown

### Phase 1: MVP Core (Weeks 1-7)

**Deliverables:**
- [x] Cobra CLI framework
- [x] Koanf configuration
- [x] Transport abstraction (stdio, SSE, HTTP)
- [x] Provider registry
- [x] Basic backend registry

**Success Criteria:**
- `metatools serve --config config.yaml` starts HTTP server
- All 13 extension points configurable via YAML
- Backward compatible with existing `metatools` usage

### Phase 2: Protocol Layer (Weeks 8-14)

**Deliverables:**
- [ ] tooladapter library
- [ ] toolset library
- [ ] Multi-transport exposure

**Success Criteria:**
- Tools convertible between MCP ↔ OpenAI ↔ Anthropic
- Composable toolsets via builder pattern
- Same tools exposed via MCP and REST

### Phase 3: Cross-Cutting (Weeks 8-17, parallel)

**Deliverables:**
- [ ] toolcache (memory, Redis, layered)
- [ ] toolobserve (OpenTelemetry)
- [ ] toolversion (semantic versioning)
- [ ] toolresilience (circuit breaker, retry)
- [ ] toolhealth, toolaudit, toolsecrets

**Success Criteria:**
- All tool calls traced via OpenTelemetry
- Circuit breaker protects against cascade failures
- Comprehensive audit log for compliance

### Phase 4: Enterprise (Weeks 12-17, parallel)

**Deliverables:**
- [ ] Multi-tenancy (resolvers, middleware, storage)
- [ ] toolsemantic (vector search)
- [ ] toolresource (MCP Resources)
- [ ] toolgateway (auth proxy)

**Success Criteria:**
- Multiple tenants isolated by configuration
- Semantic search improves tool discovery accuracy
- OAuth 2.1 authentication via gateway

### Phase 5: Agent Skills (Weeks 18-21)

**Deliverables:**
- [ ] toolskill library (skill interface, registry, manifest)
- [ ] Skill composition DSL (sequential, parallel, branching)
- [ ] Skill runtime (execution, checkpoints, rollback)
- [ ] A2A skill advertisement integration

**Success Criteria:**
- Skills discoverable via SkillRegistry.Advertise()
- Multi-step skills execute with progress tracking
- Skills exposed as MCP tools with `skill:` prefix
- Checkpoint/restore enables long-running skill recovery

**Dependencies:**
- Requires Stream B completion (toolset for tool access)
- Benefits from Stream C (toolobserve for tracing)

---

## 5. Interface Contracts

### Core Interfaces (Must Be Stable)

| Interface | Library | Breaking Change Policy |
|-----------|---------|----------------------|
| `Tool` | toolmodel | MAJOR version bump |
| `Index` | toolindex | MAJOR version bump |
| `Searcher` | toolindex | MINOR (additive only) |
| `Store` | tooldocs | MINOR (additive only) |
| `Runner` | toolrun | MAJOR version bump |
| `Backend` | toolruntime | MINOR (additive only) |

### New Interfaces (Design Phase)

| Interface | Library | Status | Review Required |
|-----------|---------|--------|-----------------|
| `Transport` | metatools | Draft | Architecture |
| `ToolProvider` | metatools | Draft | Architecture |
| `BackendRegistry` | metatools | Draft | Architecture |
| `Adapter` | tooladapter | Draft | Protocol Team |
| `Cache` | toolcache | Draft | Infrastructure |
| `Observer` | toolobserve | Draft | SRE Team |
| `Skill` | toolskill | Draft | Agent Team |
| `SkillRegistry` | toolskill | Draft | Agent Team |
| `SkillRuntime` | toolskill | Draft | Agent Team |
| `Embedder` | toolsemantic | Draft | ML/Search Team |
| `VectorIndex` | toolsemantic | Draft | ML/Search Team |
| `HybridSearcher` | toolsemantic | Draft | ML/Search Team |
| `Reranker` | toolsemantic | Draft | ML/Search Team |
| `KnowledgeGraph` | toolsemantic | Draft | ML/Search Team |
| `AgenticRetriever` | toolsemantic | Draft | ML/Search Team |

---

## 6. Edge Cases & Considerations

### Interface Evolution

| Scenario | Strategy |
|----------|----------|
| Add optional field | MINOR bump, backward compatible |
| Add required field | MAJOR bump, migration guide |
| Remove field | Deprecate → 2 releases → remove |
| Change field type | Never (create new field) |

### Error Handling

| Error Type | HTTP Code | MCP Error Code | Retry? |
|------------|-----------|----------------|--------|
| Tool not found | 404 | -32601 | No |
| Invalid input | 400 | -32602 | No |
| Backend timeout | 504 | -32603 | Yes |
| Rate limited | 429 | -32603 | Yes (with backoff) |
| Circuit open | 503 | -32603 | Yes (after reset) |
| Internal error | 500 | -32603 | Maybe |

### Concurrency

| Component | Concurrency Model | Lock Granularity |
|-----------|------------------|------------------|
| Provider Registry | RWMutex | Per-registry |
| Backend Registry | RWMutex | Per-registry |
| Cache | Per-key locking | Key-level |
| Circuit Breaker | Atomic counters | Per-breaker |
| Tenant Store | RWMutex | Per-tenant |
| Skill Registry | RWMutex | Per-registry |
| Skill Runtime | Per-execution | Execution-level |

### Resource Limits

| Resource | Default | Configurable | Enforcement |
|----------|---------|--------------|-------------|
| Max tools per backend | 1000 | Yes | Registration fails |
| Max concurrent requests | 100 | Yes | 503 response |
| Max request size | 10MB | Yes | 413 response |
| Tool execution timeout | 30s | Per-tool | Context cancellation |
| Cache entry size | 1MB | Yes | Reject large entries |

---

## 7. Rollout Strategy

### Version Compatibility Matrix

| metatools | toolmodel | toolindex | toolrun | tooladapter | toolskill |
|-----------|-----------|-----------|---------|-------------|-----------|
| v0.2.x | v0.1.2+ | v0.1.8+ | v0.1.9+ | - | - |
| v0.3.x | v0.2.0+ | v0.2.0+ | v0.2.0+ | v1.0.0+ | - |
| v1.0.x | v0.2.0+ | v0.2.0+ | v0.2.0+ | v1.0.0+ | - |
| v1.1.x | v0.2.0+ | v0.2.0+ | v0.2.0+ | v1.0.0+ | v1.0.0+ |

### Release Cadence

| Track | Frequency | Stability | Use Case |
|-------|-----------|-----------|----------|
| **stable** | Monthly | Production | Enterprise |
| **beta** | Bi-weekly | Testing | Early adopters |
| **nightly** | Daily | Development | Contributors |

### Migration Guides

Each MAJOR version bump requires:
1. **Migration guide** with step-by-step instructions
2. **Codemods** where possible (automated refactoring)
3. **Deprecation period** (minimum 2 minor versions)
4. **Backward compatibility layer** (optional, time-limited)

### Feature Flags for Rollout

```yaml
feature_flags:
  new_transport_layer: true      # Phase 2
  protocol_adapters: false       # Phase 3 (beta)
  semantic_search: false         # Phase 4 (alpha)
  multi_tenancy: false           # Phase 4 (alpha)
  agent_skills: false            # Phase 5 (alpha)
```

---

## 8. Multi-Language Extensibility

> **Design Principle**: The pluggable architecture is not limited to Go implementations. Any component can be replaced by implementations written in **any programming language** through standardized interface contracts.

### Why Multi-Language Matters

| Use Case | Language | Rationale |
|----------|----------|-----------|
| ML/Embedding Models | **Python** | Rich ML ecosystem (PyTorch, transformers, sentence-transformers) |
| High-Performance Retrieval | **Rust** | Memory safety + near-C performance for vector operations |
| Enterprise Integrations | **Java** | Existing corporate libraries, Spring ecosystem |
| Rapid Prototyping | **TypeScript** | Fast iteration, existing MCP server implementations |
| Edge Deployment | **WASM** | Sandboxed, portable, language-agnostic runtime |

### Interface Contract Technologies

The architecture supports **three proven approaches** for multi-language interoperability:

#### Option 1: gRPC + Protocol Buffers (Recommended)

**Battle-tested approach used by HashiCorp (Terraform, Vault, Consul, Packer).**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            METATOOLS CORE (Go)                                   │
│                                                                                  │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │                     Plugin Host (go-plugin style)                        │   │
│   │   - Launches plugins as subprocesses                                     │   │
│   │   - Communicates via gRPC over local socket                             │   │
│   │   - Health monitoring and automatic restart                              │   │
│   │   - Graceful shutdown and resource cleanup                               │   │
│   └───────────────────────────────┬─────────────────────────────────────────┘   │
│                                   │                                              │
└───────────────────────────────────┼──────────────────────────────────────────────┘
                                    │ gRPC (protobuf)
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│ Python Embedder   │   │ Rust Vector Index │   │ TypeScript Adapter│
│ (sentence-bert)   │   │ (HNSW + SIMD)     │   │ (OpenAI format)   │
│                   │   │                   │   │                   │
│ pip install       │   │ cargo build       │   │ npm run build     │
│ metatools-embed   │   │ --release         │   │                   │
└───────────────────┘   └───────────────────┘   └───────────────────┘
```

**Protocol Buffer Interface Definitions:**

```protobuf
// api/proto/embedder.proto
syntax = "proto3";
package metatools.v1;

service Embedder {
  rpc Embed(EmbedRequest) returns (EmbedResponse);
  rpc EmbedBatch(EmbedBatchRequest) returns (EmbedBatchResponse);
  rpc Info(InfoRequest) returns (EmbedderInfo);
}

message EmbedRequest {
  string text = 1;
}

message EmbedResponse {
  repeated float embedding = 1;
}

message EmbedderInfo {
  string model = 1;
  int32 dimensions = 2;
}
```

```protobuf
// api/proto/searcher.proto
syntax = "proto3";
package metatools.v1;

service Searcher {
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc Index(IndexRequest) returns (IndexResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}

message SearchRequest {
  string query = 1;
  int32 top_k = 2;
  map<string, string> filters = 3;
}

message SearchResult {
  string id = 1;
  float score = 2;
  string content = 3;
  map<string, string> metadata = 4;
}
```

**Benefits:**
- Language-agnostic: Generate client/server stubs for 10+ languages via `protoc`
- Schema evolution: Add fields without breaking backward compatibility
- High performance: Binary serialization, HTTP/2 multiplexing
- Proven at scale: Used by Kubernetes, gRPC-web, many microservices

**Real-World Example (HashiCorp go-plugin):**
```go
// Go host code (metatools)
type EmbedderPlugin struct {
    plugin.Plugin
    Impl Embedder
}

func (p *EmbedderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
    RegisterEmbedderServer(s, &GRPCServer{Impl: p.Impl})
    return nil
}

func (p *EmbedderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
    return &GRPCClient{client: NewEmbedderClient(c)}, nil
}
```

```python
# Python plugin implementation
class PythonEmbedder(metatools_pb2_grpc.EmbedderServicer):
    def __init__(self):
        self.model = SentenceTransformer('all-MiniLM-L6-v2')

    def Embed(self, request, context):
        embedding = self.model.encode(request.text)
        return metatools_pb2.EmbedResponse(embedding=embedding.tolist())

    def Info(self, request, context):
        return metatools_pb2.EmbedderInfo(
            model='all-MiniLM-L6-v2',
            dimensions=384
        )
```

#### Option 2: WebAssembly Component Model (Emerging)

**Sandboxed, portable, no network overhead.**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            METATOOLS CORE (Go)                                   │
│                                                                                  │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │                     WASM Runtime (wasmtime/wasmer)                       │   │
│   │   - Load .wasm modules as plugins                                        │   │
│   │   - WebAssembly Interface Types (WIT) for type-safe boundaries          │   │
│   │   - WASI for system access (filesystem, network)                         │   │
│   │   - Strong sandbox: memory isolation, capability-based security          │   │
│   └───────────────────────────────┬─────────────────────────────────────────┘   │
│                                   │                                              │
└───────────────────────────────────┼──────────────────────────────────────────────┘
                                    │ WebAssembly Interface Types (WIT)
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│ Rust → WASM       │   │ Go → WASM         │   │ Python → WASM     │
│ (vector_index.wasm│   │ (adapter.wasm)    │   │ (embedder.wasm)   │
│  via wasm32-wasi) │   │  via TinyGo       │   │  via Pyodide      │
└───────────────────┘   └───────────────────┘   └───────────────────┘
```

**WebAssembly Interface Types (WIT) Definition:**

```wit
// api/wit/embedder.wit
package metatools:embedder@1.0.0;

interface embedder {
    record embed-request {
        text: string,
    }

    record embed-response {
        embedding: list<f32>,
    }

    record embedder-info {
        model: string,
        dimensions: u32,
    }

    embed: func(request: embed-request) -> result<embed-response, string>;
    info: func() -> embedder-info;
}

world embedder-plugin {
    export embedder;
}
```

**Benefits:**
- **Sandboxed**: Strong isolation, no access beyond granted capabilities
- **Portable**: Same .wasm binary runs on any platform
- **No network overhead**: Direct function calls, not RPC
- **Multi-language**: Rust, Go (TinyGo), C/C++, AssemblyScript compile to WASM

**Limitations:**
- Ecosystem still maturing (WASI Preview 2 released 2024)
- Some languages have limited WASM support
- Performance overhead for complex data marshaling

#### Option 3: JSON-RPC over stdio (Simple)

**Lightweight approach for simple plugins.**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            METATOOLS CORE (Go)                                   │
│                                                                                  │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │                     JSON-RPC Dispatcher                                  │   │
│   │   - Spawn plugin process                                                 │   │
│   │   - Send JSON-RPC requests via stdin                                     │   │
│   │   - Receive JSON-RPC responses via stdout                                │   │
│   └───────────────────────────────┬─────────────────────────────────────────┘   │
│                                   │                                              │
└───────────────────────────────────┼──────────────────────────────────────────────┘
                                    │ JSON-RPC 2.0 over stdin/stdout
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│ Any executable    │   │ Shell script      │   │ Node.js script    │
│ (simple wrapper)  │   │ (bash/zsh)        │   │ (quick prototype) │
└───────────────────┘   └───────────────────┘   └───────────────────┘
```

**Benefits:**
- **Trivial to implement**: Any language that reads stdin and writes stdout
- **No dependencies**: No gRPC libraries, no WASM runtime
- **Debugging friendly**: Human-readable JSON messages

**Limitations:**
- Performance: Text serialization slower than binary
- No streaming: Request-response only
- Less type safety: JSON schema validation required

### Recommended Approach by Component

| Component | Recommended | Rationale |
|-----------|-------------|-----------|
| **Embedder** | gRPC (Python) | ML ecosystem, batch processing, GPU support |
| **VectorIndex** | gRPC (Rust) or WASM | Performance-critical, SIMD optimization |
| **Reranker** | gRPC (Python) | Hugging Face transformers, cross-encoders |
| **KnowledgeGraph** | gRPC (Python/Java) | Neo4j, NetworkX, JanusGraph bindings |
| **Adapter** | Go native or WASM | Simple logic, low latency |
| **Cache** | Go native | Redis/memory clients well-supported in Go |

### Plugin SDK Generation

The architecture supports **automatic SDK generation** from Protocol Buffer definitions:

```bash
# Generate SDKs for all supported languages
make generate-sdks

# Outputs:
# - sdk/go/       (native Go interfaces)
# - sdk/python/   (gRPC stubs + helper classes)
# - sdk/rust/     (gRPC stubs + traits)
# - sdk/typescript/ (gRPC-web stubs)
```

**Python SDK Example:**

```python
# sdk/python/metatools/embedder.py
from abc import ABC, abstractmethod
from metatools.proto import embedder_pb2, embedder_pb2_grpc

class BaseEmbedder(embedder_pb2_grpc.EmbedderServicer, ABC):
    """Base class for implementing Embedder plugins in Python."""

    @abstractmethod
    def embed(self, text: str) -> list[float]:
        """Generate embedding for text."""
        pass

    @abstractmethod
    def dimensions(self) -> int:
        """Return embedding dimensions."""
        pass

    @abstractmethod
    def model_name(self) -> str:
        """Return model identifier."""
        pass

    # gRPC method implementations
    def Embed(self, request, context):
        embedding = self.embed(request.text)
        return embedder_pb2.EmbedResponse(embedding=embedding)

    def Info(self, request, context):
        return embedder_pb2.EmbedderInfo(
            model=self.model_name(),
            dimensions=self.dimensions()
        )

def serve(embedder: BaseEmbedder, port: int = 50051):
    """Start the gRPC server for the embedder plugin."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    embedder_pb2_grpc.add_EmbedderServicer_to_server(embedder, server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    server.wait_for_termination()
```

### Configuration for Multi-Language Plugins

```yaml
# metatools.yaml
plugins:
  # gRPC plugin (Python embedder)
  embedder:
    type: grpc
    command: "python -m metatools_embedder"
    address: "localhost:50051"
    health_check_interval: 10s
    restart_on_failure: true

  # WASM plugin (Rust vector index)
  vector_index:
    type: wasm
    module: "plugins/vector_index.wasm"
    memory_limit: 256MB
    capabilities:
      - filesystem:read:/data/indices

  # JSON-RPC plugin (quick prototype)
  custom_adapter:
    type: jsonrpc
    command: "node plugins/adapter.js"
    timeout: 5s
```

### Research References

This multi-language architecture is informed by:

1. **HashiCorp go-plugin**: Battle-tested gRPC plugin system used by Terraform, Vault, Consul, and Packer for 10+ years
2. **gRPC + Protocol Buffers**: Industry standard for polyglot microservices, supports 10+ languages
3. **WebAssembly Component Model (WCM)**: Emerging standard for sandboxed, portable plugin systems
4. **WASI Preview 2**: Standardized system interfaces for WASM (filesystem, network, clocks)
5. **Interface Definition Languages (IDL)**: Contract-first development enabling language-agnostic interoperability

---

## Summary: Roadmap at a Glance

```
┌───────────────────────────────────────────────────────────────────────────────────┐
│                           METATOOLS ROADMAP 2026                                   │
├───────────────────────────────────────────────────────────────────────────────────┤
│                                                                                    │
│  EXISTING (7 libs)        NEW CORE (2 libs)       CROSS-CUTTING (7 libs)          │
│  ═══════════════════      ═════════════════       ══════════════════════          │
│  ✅ toolmodel             🔲 tooladapter          🔲 toolcache                     │
│  ✅ toolindex             🔲 toolset              🔲 toolobserve                   │
│  ✅ tooldocs                                      🔲 toolversion                   │
│  ✅ toolsearch            ENTERPRISE (5 libs)     🔲 toolresilience                │
│  ✅ toolrun               ═══════════════════     🔲 toolhealth                    │
│  ✅ toolcode              🔲 toolsemantic         🔲 toolaudit                     │
│  ✅ toolruntime           🔲 toolresource         🔲 toolsecrets                   │
│                           🔲 toolgateway          🔲 toolflags                     │
│                           🔲 toola2a (future)     🔲 toolpressure                  │
│                           🔲 toolprompt (future)                                   │
│                                                                                    │
│  AGENT SKILLS (1 lib)                                                             │
│  ════════════════════                                                             │
│  🔲 toolskill             Skill composition, workflows, A2A advertisement         │
│                                                                                    │
│  ──────────────────────────────────────────────────────────────────────────────── │
│                                                                                    │
│  TIMELINE                                                                         │
│  ═════════                                                                        │
│                                                                                    │
│  Weeks 1-7   [████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░] MVP Core     │
│  Weeks 8-14  [░░░░░░░░░░░░░░░░░░░░████████████████████████░░░░░░░░░] Protocol     │
│  Weeks 8-17  [░░░░░░░░░░░░░░░░░░░░████████████████████████████░░░░░] Cross-Cut    │
│  Weeks 12-17 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░████████████████░░░░░] Enterprise   │
│  Weeks 18-21 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░████] Agent Skills │
│                                                                                    │
│  ──────────────────────────────────────────────────────────────────────────────── │
│                                                                                    │
│  MILESTONES                                                                       │
│  ══════════                                                                       │
│                                                                                    │
│  Week 7:  MVP Release (v0.2.0) - CLI, Config, Transport, Providers               │
│  Week 14: Protocol Release (v0.3.0) - Adapters, Toolsets, Multi-Transport        │
│  Week 17: Enterprise Release (v1.0.0) - Full feature set                         │
│  Week 21: Agent Skills Release (v1.1.0) - Skills, Workflows, A2A                 │
│                                                                                    │
└───────────────────────────────────────────────────────────────────────────────────┘
```

---

## Appendix: Document Cross-References

| Document | Purpose | Status |
|----------|---------|--------|
| [pluggable-architecture.md](./pluggable-architecture.md) | Core architecture design | Active |
| [implementation-phases.md](./implementation-phases.md) | Phase details | Active |
| [component-library-analysis.md](./component-library-analysis.md) | Library analysis | Complete |
| [architecture-evaluation.md](./architecture-evaluation.md) | Championship comparison | Complete |
| [protocol-agnostic-tools.md](./protocol-agnostic-tools.md) | Protocol layer design | Active |
| [multi-tenancy.md](./multi-tenancy.md) | Multi-tenant design | Active |
| **ROADMAP.md** (this document) | Master roadmap | Active |

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-28 | Added Section 8: Multi-Language Extensibility - gRPC, WASM, JSON-RPC plugin architectures with SDK generation |
| 2026-01-28 | Added comprehensive D4 toolsemantic specification: hybrid search, GraphRAG, hierarchical chunking, agentic RAG, ColBERT, cross-encoder reranking |
| 2026-01-28 | Added SKILL.md Open Standard support to toolskill (Claude Code, Codex, ChatGPT compatible) |
| 2026-01-28 | Added Stream E: Agent Skills with toolskill library proposal |
| 2026-01-28 | Initial roadmap created from all proposal documents |
