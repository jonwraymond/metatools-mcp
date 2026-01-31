# metatools-mcp Implementation Roadmap

**Last Updated:** 2026-01-30
**Status:** Active Development

---

## Current State

The core pluggable architecture is **complete**. All 5 foundation phases from the original implementation plan have been implemented and tested.

### Completed Work

| Phase | Component | Status | Coverage |
|-------|-----------|--------|----------|
| **Phase 1** | CLI + Config | ✅ Done | Koanf loader, app config, env support |
| **Phase 2** | Transport Layer | ✅ Done | Stdio, SSE, Streamable HTTP |
| **Phase 3** | Provider Registry | ✅ Done | Interface, registry, builtins |
| **Phase 4** | Backend Registry | ✅ Done | Interface, registry, aggregator, local |
| **Phase 5** | Middleware Chain | ✅ Done | Chain, registry, logging, metrics |
| **Auth MVP** | Authentication | ✅ Done | JWT, API Key, RBAC, composite |
| **Auth Enhancements** | Advanced Auth | ✅ Done | JWKS, Redis store, OAuth2 introspection |
| **Runtime** | Docker & WASM | ✅ Done | Docker client, WASM loader |

### Test Coverage

| Package | Coverage |
|---------|----------|
| internal/auth | 82.7% |
| internal/middleware | 85.1% |
| internal/provider | ~80% |
| internal/backend | ~75% |
| internal/transport | ~70% |

---

## Remaining Work

### Priority 1: Cross-Cutting Libraries (Weeks 1-4)

These libraries provide shared capabilities used by multiple components.

| Library | Purpose | Effort | Dependencies |
|---------|---------|--------|--------------|
| **toolobserve** | OpenTelemetry tracing + metrics | 2 weeks | None |
| **toolcache** | Pluggable caching (Memory/Redis) | 2 weeks | None |
| **toolhealth** | Health checks, readiness probes | 1 week | None |
| **toolresilience** | Circuit breaker, retry, bulkhead | 2 weeks | None |

#### toolobserve

OpenTelemetry integration for distributed tracing and metrics.

```go
// Core interfaces
type Observer interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
    RecordMetric(name string, value float64, tags ...Tag)
}

type Span interface {
    End()
    SetAttribute(key string, value any)
    RecordError(err error)
}
```

**Features:**
- Distributed tracing with span propagation
- Metrics recording (counters, histograms, gauges)
- Structured logging integration
- OTLP, Prometheus, stdout exporters
- Middleware for automatic instrumentation

#### toolcache

Response caching with pluggable backends.

```go
// Core interfaces
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

type KeyGenerator interface {
    Generate(toolID string, args map[string]any) string
}
```

**Features:**
- Memory cache with LRU eviction
- Redis cache for distributed deployments
- Per-tool TTL configuration
- Cache bypass for specific tools
- Statistics and hit rate tracking

#### toolhealth

Health and readiness checks for container orchestration.

```go
// Core interfaces
type HealthChecker interface {
    Check(ctx context.Context) HealthStatus
}

type HealthStatus struct {
    Status  Status // healthy, degraded, unhealthy
    Message string
    Details map[string]any
}
```

**Features:**
- Liveness probe endpoint (`/healthz`)
- Readiness probe endpoint (`/readyz`)
- Component health aggregation
- Graceful degradation reporting

#### toolresilience

Fault tolerance patterns for production reliability.

```go
// Core interfaces
type CircuitBreaker interface {
    Execute(ctx context.Context, fn func() error) error
    State() CircuitState
}

type RetryPolicy interface {
    Execute(ctx context.Context, fn func() error) error
}

type Bulkhead interface {
    Execute(ctx context.Context, fn func() error) error
}
```

**Features:**
- Circuit breaker with configurable thresholds
- Retry with exponential backoff
- Bulkhead for resource isolation
- Timeout enforcement
- Fallback handlers

---

### Priority 2: Protocol Layer (Weeks 5-8)

Enable protocol-agnostic tool handling and composition.

| Library | Purpose | Effort | Dependencies |
|---------|---------|--------|--------------|
| **tooladapter** | Protocol-agnostic tool abstraction | 2 weeks | toolmodel |
| **toolset** | Composable tool collections | 2 weeks | tooladapter, toolindex |
| **toolversion** | Semantic versioning, negotiation | 2 weeks | toolmodel |

#### tooladapter

Canonical tool representation with bidirectional protocol adapters.

```go
// Canonical tool (protocol-agnostic)
type CanonicalTool struct {
    ID          string
    Namespace   string
    Name        string
    Version     string
    Description string
    InputSchema *JSONSchema
    Handler     ToolHandler
    SourceFormat string
    SourceMeta   map[string]any
}

// Adapter interface
type Adapter interface {
    Name() string
    ToCanonical(raw any) (*CanonicalTool, error)
    FromCanonical(tool *CanonicalTool) (any, error)
    SupportsFeature(feature SchemaFeature) bool
}
```

**Adapters:**
- MCP adapter (bidirectional)
- OpenAI function calling adapter
- Anthropic tool use adapter
- LangChain adapter
- OpenAPI adapter

#### toolset

Composable tool collections with filtering and access control.

```go
// Toolset definition
type Toolset struct {
    ID          string
    Name        string
    Description string
    Tools       []*CanonicalTool
    Policy      *AccessPolicy
}

// Fluent builder
type Builder struct {
    // ...
}

func NewBuilder(name string) *Builder
func (b *Builder) WithNamespace(ns string) *Builder
func (b *Builder) WithTags(tags ...string) *Builder
func (b *Builder) WithTools(ids ...string) *Builder
func (b *Builder) ExcludeTools(ids ...string) *Builder
func (b *Builder) WithPolicy(p *AccessPolicy) *Builder
func (b *Builder) Build() (*Toolset, error)
```

**Features:**
- Filter by namespace, tags, category
- Whitelist/blacklist specific tools
- Access control policies
- Multi-protocol exposure

#### toolversion

Semantic versioning and protocol negotiation.

```go
// Version management
type Version struct {
    Major int
    Minor int
    Patch int
    Pre   string
}

type Negotiator interface {
    Negotiate(client, server []Version) (Version, error)
}

type Deprecation struct {
    Version     Version
    Replacement string
    Deadline    time.Time
}
```

**Features:**
- Semantic version parsing/comparison
- Protocol version negotiation
- Deprecation tracking
- Migration guidance

---

### Priority 3: Enterprise Features (Weeks 9-14)

Production-ready features for enterprise deployments.

| Library | Purpose | Effort | Dependencies |
|---------|---------|--------|--------------|
| **Multi-Tenancy** | Tenant isolation | 3 weeks | Auth, Middleware |
| **toolsemantic** | Hybrid search (BM25+vector) | 3 weeks | toolindex, toolsearch |
| **toolgateway** | Auth proxy, analytics | 3 weeks | All above |
| **toolresource** | MCP Resources support | 2 weeks | toolindex |

#### Multi-Tenancy (internal/tenant)

Pluggable tenant isolation with multiple strategies.

```go
// Tenant resolution
type TenantResolver interface {
    Resolve(ctx context.Context, req *Request) (*TenantContext, error)
}

// Tenant context
type TenantContext struct {
    Tenant      *Tenant
    Permissions []string
    Config      *TenantConfig
    Quotas      *TenantQuotas
}

// Tenant storage
type TenantStore interface {
    Get(ctx context.Context, id string) (*Tenant, error)
    GetConfig(ctx context.Context, id string) (*TenantConfig, error)
    IncrementUsage(ctx context.Context, id string, metric string, delta int64) error
}
```

**Features:**
- JWT, API Key, Header resolvers
- Tier-based configuration (free/pro/enterprise)
- Per-tenant rate limits and quotas
- Tool filtering by tenant
- Audit logging per tenant
- Isolation strategies (shared, namespace, process)

#### toolsemantic

Hybrid search combining BM25 and vector embeddings.

```go
// Embedder interface
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// Vector index
type VectorIndex interface {
    Add(ctx context.Context, id string, vector []float32) error
    Search(ctx context.Context, vector []float32, k int) ([]Match, error)
}

// Hybrid searcher
type HybridSearcher struct {
    BM25    *toolsearch.BM25Searcher
    Vector  VectorIndex
    Weights HybridWeights
}
```

**Features:**
- Pluggable embedding providers (OpenAI, Cohere, local)
- Vector index backends (in-memory, Qdrant, Pinecone)
- Hybrid scoring (BM25 + vector fusion)
- Optional reranking
- Accuracy: 94% (vs 78% BM25-only)

#### toolgateway

Centralized proxy for auth, rate limiting, and analytics.

```go
// Gateway configuration
type Gateway struct {
    Auth       AuthProvider
    RateLimit  RateLimiter
    Analytics  AnalyticsRecorder
    Upstream   []MCPServer
}

// Analytics
type AnalyticsRecorder interface {
    Record(ctx context.Context, event AnalyticsEvent)
    Query(ctx context.Context, filter Filter) ([]AnalyticsEvent, error)
}
```

**Features:**
- Centralized authentication
- Global rate limiting
- Request/response analytics
- Multi-upstream routing
- Load balancing

#### toolresource

MCP Resources support for file/data exposure.

```go
// Resource provider
type ResourceProvider interface {
    List(ctx context.Context) ([]Resource, error)
    Read(ctx context.Context, uri string) (io.Reader, error)
    Subscribe(ctx context.Context, uri string) (<-chan ResourceUpdate, error)
}

// Resource definition
type Resource struct {
    URI         string
    Name        string
    Description string
    MimeType    string
}
```

**Features:**
- File system resources
- Database resources
- API resources
- Subscription for real-time updates

---

### Priority 4: Agent Skills (Weeks 15-18)

Higher-level capability composition for AI agents.

| Library | Purpose | Effort | Dependencies |
|---------|---------|--------|--------------|
| **toolskill** | SKILL.md-compatible agent skills | 4 weeks | toolset, toolrun |

#### toolskill

Agent skills framework compatible with Claude Code SKILL.md format.

```go
// Skill definition
type Skill struct {
    Name        string
    Description string
    Triggers    []Trigger
    Steps       []Step
    Outputs     []Output
}

// Skill runtime
type SkillRuntime interface {
    Load(ctx context.Context, path string) (*Skill, error)
    Execute(ctx context.Context, skill *Skill, input map[string]any) (*Result, error)
}

// Step execution
type Step struct {
    Name      string
    Tool      string
    Arguments map[string]any
    OnSuccess string
    OnFailure string
}
```

**Features:**
- SKILL.md parsing and validation
- Multi-step workflows
- Conditional branching
- Error handling and retry
- Output composition
- Skill composition (skills calling skills)

---

## Timeline Summary

```
Week 1-2:   toolobserve (OpenTelemetry)
Week 3-4:   toolcache + toolhealth
Week 5-6:   tooladapter (Protocol adapters)
Week 7-8:   toolset + toolversion
Week 9-11:  Multi-Tenancy
Week 12-14: toolsemantic + toolgateway
Week 15-18: toolskill (Agent Skills)
```

**Milestones:**
- **Week 4:** Cross-cutting complete → Production observability ready
- **Week 8:** Protocol layer complete → Multi-protocol support
- **Week 14:** Enterprise complete → SaaS-ready deployment
- **Week 18:** Full stack complete → Agent skills platform

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           metatools-mcp Server                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     TRANSPORT LAYER (✅ Done)                        │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────────┐                          │    │
│  │  │  Stdio  │  │   SSE   │  │ Streamable  │                          │    │
│  │  │         │  │         │  │    HTTP     │                          │    │
│  │  └─────────┘  └─────────┘  └─────────────┘                          │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────┼───────────────────────────────────┐    │
│  │              MIDDLEWARE CHAIN (✅ Done + Extensions)                 │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ │    │
│  │  │Logging │→│Tracing │→│  Auth  │→│ Rate   │→│ Cache  │→│ Audit  │ │    │
│  │  │  ✅    │ │  ❌    │ │   ✅   │ │ Limit✅│ │  ❌    │ │  ✅    │ │    │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘ └────────┘ │    │
│  └─────────────────────────────────┬───────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────┼───────────────────────────────────┐    │
│  │              PROVIDER REGISTRY (✅ Done)                             │    │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                 │    │
│  │  │ search_tools │ │ describe_    │ │   run_tool   │                 │    │
│  │  │              │ │    tool      │ │              │                 │    │
│  │  └──────────────┘ └──────────────┘ └──────────────┘                 │    │
│  └─────────────────────────────────┬───────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────┼───────────────────────────────────┐    │
│  │              BACKEND REGISTRY (✅ Done)                              │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                 │    │
│  │  │  Local  │  │   MCP   │  │  HTTP   │  │ Custom  │                 │    │
│  │  │   ✅    │  │   ⚠️    │  │   ❌    │  │   ✅    │                 │    │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘                 │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │              CORE LIBRARIES (✅ External - Production)               │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │    │
│  │  │toolmodel │ │toolindex │ │ tooldocs │ │ toolrun  │ │ toolcode │  │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │    │
│  │  ┌──────────┐ ┌──────────┐                                          │    │
│  │  │toolsearch│ │toolruntime│                                         │    │
│  │  └──────────┘ └──────────┘                                          │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘

LEGEND: ✅ Done  ⚠️ Partial  ❌ Not Started
```

---

## References

| Document | Purpose |
|----------|---------|
| `docs/proposals/ROADMAP.md` | Master architecture roadmap |
| `docs/proposals/pluggable-architecture.md` | Core architecture design |
| `docs/proposals/implementation-phases.md` | Phase breakdown |
| `docs/proposals/auth-middleware.md` | Auth system design |
| `docs/proposals/multi-tenancy.md` | Multi-tenant design |
| `docs/proposals/protocol-agnostic-tools.md` | Protocol adapter design |
| `docs/proposals/architecture-evaluation.md` | Gap analysis |
| `docs/plans/` | PRD documents |

---

## Contributing

When implementing new features:

1. **Create PRD** in `docs/plans/` with test cases
2. **Follow TDD** - write tests first, then implementation
3. **Maintain coverage** - target >80% for new code
4. **Update this roadmap** when work completes
5. **Run verification**:
   ```bash
   GOWORK=off go test ./... -cover
   GOWORK=off golangci-lint run ./...
   ```
