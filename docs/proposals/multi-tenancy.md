# Multi-Tenancy Extension for Pluggable Architecture

**Status:** Draft
**Date:** 2026-01-28
**Related:** [Pluggable Architecture Proposal](./pluggable-architecture.md)

## Overview

This document extends the pluggable architecture to support multi-tenancy in a flexible, pluggable way. The design allows different tenant isolation strategies without hardcoding any specific approach.

---

## Multi-Tenancy Architecture

### Design Principles

1. **Pluggable Tenant Resolution** - How tenants are identified (JWT, API key, header, etc.)
2. **Pluggable Isolation Strategy** - Level of isolation between tenants
3. **Tenant-Aware Middleware** - All middleware can access tenant context
4. **Tenant-Specific Configuration** - Override any config per tenant
5. **Backward Compatible** - Single-tenant mode works without changes

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       MULTI-TENANT MCP SERVER                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   Incoming Request                                                           │
│         │                                                                     │
│         ▼                                                                     │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    TENANT RESOLUTION MIDDLEWARE                      │   │
│   │                                                                       │   │
│   │   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │   │
│   │   │    JWT      │ │   API Key   │ │   Header    │ │   Custom    │   │   │
│   │   │  Resolver   │ │  Resolver   │ │  Resolver   │ │  Resolver   │   │   │
│   │   └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │   │
│   │                                                                       │   │
│   │   Output: Tenant Context injected into request context               │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                             │                                                │
│                             ▼                                                │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    TENANT-AWARE MIDDLEWARE CHAIN                     │   │
│   │                                                                       │   │
│   │   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐               │   │
│   │   │ Tenant   │→│ Tenant   │→│ Tenant   │→│ Tenant   │               │   │
│   │   │ Scoped   │ │ Rate     │ │ Audit    │ │ Tool     │               │   │
│   │   │ Config   │ │ Limits   │ │ Logging  │ │ Filter   │               │   │
│   │   └──────────┘ └──────────┘ └──────────┘ └──────────┘               │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                             │                                                │
│                             ▼                                                │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    TENANT-SCOPED REGISTRIES                          │   │
│   │                                                                       │   │
│   │   ┌──────────────────┐    ┌──────────────────┐                      │   │
│   │   │  Shared Registry │    │  Tenant Registry │                      │   │
│   │   │  (all tenants)   │    │  (tenant-only)   │                      │   │
│   │   └──────────────────┘    └──────────────────┘                      │   │
│   │                                                                       │   │
│   │   Tool visibility = Shared ∪ Tenant-specific - Denied               │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Interfaces

### Tenant Model

```go
// Tenant represents a tenant in the system
type Tenant struct {
    ID          string            // Unique tenant identifier
    Name        string            // Human-readable name
    Tier        TenantTier        // free, pro, enterprise
    Metadata    map[string]any    // Arbitrary tenant metadata
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// TenantTier defines service levels
type TenantTier string

const (
    TenantTierFree       TenantTier = "free"
    TenantTierPro        TenantTier = "pro"
    TenantTierEnterprise TenantTier = "enterprise"
)

// TenantContext holds runtime tenant information
type TenantContext struct {
    Tenant      *Tenant
    Permissions []string          // Resolved permissions
    Config      *TenantConfig     // Tenant-specific config overrides
    Quotas      *TenantQuotas     // Current quota state
}
```

### Tenant Resolver Interface

```go
// TenantResolver identifies the tenant from a request
type TenantResolver interface {
    // Resolve extracts tenant information from the request context
    // Returns nil tenant for anonymous/default tenant
    Resolve(ctx context.Context, req *Request) (*TenantContext, error)
}

// TenantResolverFunc is a convenience type
type TenantResolverFunc func(ctx context.Context, req *Request) (*TenantContext, error)
```

### Built-in Resolvers

```go
// JWTTenantResolver extracts tenant from JWT claims
type JWTTenantResolver struct {
    ClaimKey    string                    // JWT claim containing tenant ID
    TenantStore TenantStore               // Lookup tenant details
    Validator   JWTValidator              // Validate JWT
}

func (r *JWTTenantResolver) Resolve(ctx context.Context, req *Request) (*TenantContext, error) {
    token := extractBearerToken(req)
    claims, err := r.Validator.Validate(token)
    if err != nil {
        return nil, err
    }

    tenantID, ok := claims[r.ClaimKey].(string)
    if !ok {
        return nil, ErrNoTenantClaim
    }

    tenant, err := r.TenantStore.Get(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    return &TenantContext{
        Tenant:      tenant,
        Permissions: extractPermissions(claims),
        Config:      r.TenantStore.GetConfig(ctx, tenantID),
    }, nil
}

// APIKeyTenantResolver extracts tenant from API key
type APIKeyTenantResolver struct {
    HeaderName  string       // Header containing API key
    KeyStore    APIKeyStore  // Lookup key -> tenant mapping
}

// HeaderTenantResolver extracts tenant from a header
type HeaderTenantResolver struct {
    HeaderName  string
    TenantStore TenantStore
}

// CompositeTenantResolver tries multiple resolvers in order
type CompositeTenantResolver struct {
    Resolvers []TenantResolver
}

func (r *CompositeTenantResolver) Resolve(ctx context.Context, req *Request) (*TenantContext, error) {
    for _, resolver := range r.Resolvers {
        tc, err := resolver.Resolve(ctx, req)
        if err == nil && tc != nil {
            return tc, nil
        }
    }
    return nil, ErrNoTenantResolved
}
```

---

## Tenant Configuration

### Tenant-Specific Config Overrides

```go
// TenantConfig holds per-tenant configuration overrides
type TenantConfig struct {
    // Tool access control
    AllowedTools     []string          // Whitelist (empty = all allowed)
    DeniedTools      []string          // Blacklist (takes precedence)
    AllowedBackends  []string          // Which backends tenant can use
    DeniedBackends   []string          // Blacklisted backends

    // Resource limits
    RateLimits       *RateLimitConfig  // Override rate limits
    Quotas           *QuotaConfig      // Usage quotas

    // Feature flags
    Features         map[string]bool   // Feature toggles

    // Execution
    MaxTimeout       time.Duration     // Max allowed timeout
    MaxChainSteps    int               // Max chain steps
    MaxToolCalls     int               // Max tool calls per request

    // Custom middleware config
    MiddlewareConfig map[string]any    // Per-middleware overrides
}

// QuotaConfig defines usage quotas
type QuotaConfig struct {
    DailyRequests    int64             // Requests per day
    DailyToolCalls   int64             // Tool calls per day
    MonthlyRequests  int64             // Requests per month
    MonthlyToolCalls int64             // Tool calls per month
}
```

### Configuration Hierarchy

```yaml
# metatools.yaml - Multi-tenant configuration

tenancy:
  enabled: true

  # Tenant resolution strategy
  resolver:
    type: composite
    resolvers:
      - type: jwt
        claim_key: tenant_id
        issuer: https://auth.example.com
      - type: api_key
        header: X-API-Key
      - type: header
        header: X-Tenant-ID

  # Default tenant (anonymous requests)
  default_tenant:
    id: default
    tier: free
    config:
      rate_limits:
        requests_per_minute: 10
      allowed_tools:
        - search_tools
        - describe_tool
      denied_tools:
        - execute_code

  # Tier-based defaults
  tiers:
    free:
      rate_limits:
        requests_per_minute: 60
        burst: 10
      quotas:
        daily_requests: 1000
        monthly_requests: 10000
      denied_tools:
        - execute_code
      max_chain_steps: 3

    pro:
      rate_limits:
        requests_per_minute: 300
        burst: 50
      quotas:
        daily_requests: 10000
        monthly_requests: 100000
      max_chain_steps: 10

    enterprise:
      rate_limits:
        requests_per_minute: 1000
        burst: 200
      quotas:
        daily_requests: -1  # unlimited
        monthly_requests: -1
      max_chain_steps: 50
      features:
        custom_backends: true
        audit_logging: true

# Per-tenant overrides (loaded from store or config)
tenants:
  acme-corp:
    tier: enterprise
    config:
      allowed_backends:
        - local
        - github
        - jira
      features:
        execute_code: true
      middleware_config:
        audit:
          destination: elasticsearch
          index: acme-audit

  startup-xyz:
    tier: pro
    config:
      rate_limits:
        requests_per_minute: 500  # Override pro default
```

---

## Tenant-Aware Middleware

### Tenant Context Middleware

```go
// TenantMiddleware resolves and injects tenant context
func TenantMiddleware(resolver TenantResolver) Middleware {
    return func(next ToolProvider) ToolProvider {
        return &tenantMiddleware{
            resolver: resolver,
            next:     next,
        }
    }
}

type tenantMiddleware struct {
    resolver TenantResolver
    next     ToolProvider
}

func (m *tenantMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    // Resolve tenant from context (set by transport layer)
    req := RequestFromContext(ctx)
    tc, err := m.resolver.Resolve(ctx, req)
    if err != nil {
        return nil, &TenantError{Op: "resolve", Err: err}
    }

    // Inject tenant context
    ctx = WithTenantContext(ctx, tc)

    return m.next.Handle(ctx, input)
}

// Context helpers
func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context
func TenantFromContext(ctx context.Context) *TenantContext
func TenantIDFromContext(ctx context.Context) string
```

### Tenant-Aware Rate Limiting

```go
// TenantRateLimitMiddleware applies per-tenant rate limits
type TenantRateLimitMiddleware struct {
    store    RateLimitStore
    defaults RateLimitConfig
    next     ToolProvider
}

func (m *TenantRateLimitMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    tc := TenantFromContext(ctx)

    // Get tenant-specific limits or tier defaults
    limits := m.resolveLimits(tc)

    // Check rate limit
    key := fmt.Sprintf("tenant:%s:tool:%s", tc.Tenant.ID, m.next.Name())
    allowed, err := m.store.Allow(ctx, key, limits)
    if err != nil {
        return nil, err
    }
    if !allowed {
        return nil, &RateLimitError{
            TenantID: tc.Tenant.ID,
            Limit:    limits.RequestsPerMinute,
        }
    }

    return m.next.Handle(ctx, input)
}
```

### Tenant Tool Filter

```go
// TenantToolFilterMiddleware filters tools based on tenant permissions
type TenantToolFilterMiddleware struct {
    next ToolProvider
}

func (m *TenantToolFilterMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    tc := TenantFromContext(ctx)
    toolName := m.next.Name()

    // Check denied tools first (takes precedence)
    if slices.Contains(tc.Config.DeniedTools, toolName) {
        return nil, &ToolDeniedError{
            TenantID: tc.Tenant.ID,
            Tool:     toolName,
            Reason:   "tool denied for tenant",
        }
    }

    // Check allowed tools (if whitelist is non-empty)
    if len(tc.Config.AllowedTools) > 0 {
        if !slices.Contains(tc.Config.AllowedTools, toolName) {
            return nil, &ToolDeniedError{
                TenantID: tc.Tenant.ID,
                Tool:     toolName,
                Reason:   "tool not in allowed list",
            }
        }
    }

    return m.next.Handle(ctx, input)
}
```

### Tenant Audit Middleware

```go
// TenantAuditMiddleware logs all actions with tenant context
type TenantAuditMiddleware struct {
    logger AuditLogger
    next   ToolProvider
}

func (m *TenantAuditMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    tc := TenantFromContext(ctx)
    start := time.Now()

    result, err := m.next.Handle(ctx, input)

    m.logger.Log(AuditEntry{
        Timestamp:  time.Now(),
        TenantID:   tc.Tenant.ID,
        TenantTier: string(tc.Tenant.Tier),
        Tool:       m.next.Name(),
        Input:      input,
        Success:    err == nil,
        Duration:   time.Since(start),
        RequestID:  RequestIDFromContext(ctx),
        UserID:     UserIDFromContext(ctx),
    })

    return result, err
}
```

---

## Tenant-Scoped Registries

### Multi-Tenant Tool Registry

```go
// MultiTenantRegistry wraps a base registry with tenant scoping
type MultiTenantRegistry struct {
    shared        *provider.Registry  // Shared tools (all tenants)
    tenantTools   map[string]*provider.Registry  // Tenant-specific tools
    tenantStore   TenantStore
    mu            sync.RWMutex
}

// GetForTenant returns a tenant-scoped view of the registry
func (r *MultiTenantRegistry) GetForTenant(tenantID string) *TenantScopedRegistry {
    r.mu.RLock()
    defer r.mu.RUnlock()

    tenant, _ := r.tenantStore.Get(context.Background(), tenantID)
    tenantReg := r.tenantTools[tenantID]

    return &TenantScopedRegistry{
        tenant:   tenant,
        shared:   r.shared,
        specific: tenantReg,
    }
}

// TenantScopedRegistry provides a tenant's view of available tools
type TenantScopedRegistry struct {
    tenant   *Tenant
    shared   *provider.Registry
    specific *provider.Registry
}

// All returns all tools visible to the tenant
func (r *TenantScopedRegistry) All() []provider.ToolProvider {
    var result []provider.ToolProvider

    // Add shared tools (filtered by tenant config)
    for _, p := range r.shared.All() {
        if r.isToolAllowed(p.Name()) {
            result = append(result, p)
        }
    }

    // Add tenant-specific tools
    if r.specific != nil {
        result = append(result, r.specific.All()...)
    }

    return result
}

func (r *TenantScopedRegistry) isToolAllowed(name string) bool {
    if r.tenant == nil || r.tenant.Config == nil {
        return true
    }

    cfg := r.tenant.Config

    // Check denied list first
    if slices.Contains(cfg.DeniedTools, name) {
        return false
    }

    // Check allowed list (if specified)
    if len(cfg.AllowedTools) > 0 {
        return slices.Contains(cfg.AllowedTools, name)
    }

    return true
}
```

### Multi-Tenant Backend Registry

```go
// MultiTenantBackendRegistry manages tenant-scoped backends
type MultiTenantBackendRegistry struct {
    shared        *backend.Registry
    tenantBackends map[string]*backend.Registry
    mu            sync.RWMutex
}

// GetForTenant returns backends available to a tenant
func (r *MultiTenantBackendRegistry) GetForTenant(ctx context.Context, tenantID string) *TenantScopedBackendRegistry {
    tc := TenantFromContext(ctx)

    r.mu.RLock()
    defer r.mu.RUnlock()

    return &TenantScopedBackendRegistry{
        tenant:   tc,
        shared:   r.shared,
        specific: r.tenantBackends[tenantID],
    }
}
```

---

## Isolation Strategies

### Strategy 1: Shared Infrastructure (Default)

All tenants share the same server instance with logical isolation via middleware.

```yaml
tenancy:
  isolation: shared
  # All tenants use same tool/backend registries
  # Isolation via rate limits, tool filtering, audit logging
```

```
┌─────────────────────────────────────────────────────────────────┐
│                    SHARED INFRASTRUCTURE                          │
│                                                                   │
│   ┌─────────┐ ┌─────────┐ ┌─────────┐                           │
│   │Tenant A │ │Tenant B │ │Tenant C │                           │
│   └────┬────┘ └────┬────┘ └────┬────┘                           │
│        │           │           │                                 │
│        └───────────┴───────────┘                                 │
│                    │                                              │
│                    ▼                                              │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │              SHARED MCP SERVER INSTANCE                  │   │
│   │                                                           │   │
│   │   Tenant Middleware → Tool Filter → Rate Limit → Audit   │   │
│   │                                                           │   │
│   │   Shared Tool Registry + Tenant Configs                  │   │
│   │                                                           │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Strategy 2: Namespace Isolation

Tenants have isolated namespaces within shared infrastructure.

```yaml
tenancy:
  isolation: namespace
  # Each tenant gets own tool namespace prefix
  # Tenant A sees: tenant-a:*, shared:*
  # Tenant B sees: tenant-b:*, shared:*
```

```
┌─────────────────────────────────────────────────────────────────┐
│                    NAMESPACE ISOLATION                            │
│                                                                   │
│   Tenant A View:              Tenant B View:                     │
│   ┌─────────────────┐        ┌─────────────────┐                │
│   │ shared:*        │        │ shared:*        │                │
│   │ tenant-a:*      │        │ tenant-b:*      │                │
│   └─────────────────┘        └─────────────────┘                │
│                                                                   │
│   Backend Routing:                                               │
│   shared:* → Shared backends                                     │
│   tenant-a:* → Tenant A's backends                              │
│   tenant-b:* → Tenant B's backends                              │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Strategy 3: Process Isolation

Each tenant gets a dedicated server process.

```yaml
tenancy:
  isolation: process
  # Tenant requests routed to dedicated processes
  # Maximum isolation, higher resource usage
```

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROCESS ISOLATION                              │
│                                                                   │
│   ┌─────────────┐        ┌─────────────┐        ┌─────────────┐ │
│   │  Tenant A   │        │  Tenant B   │        │  Tenant C   │ │
│   │   Process   │        │   Process   │        │   Process   │ │
│   │             │        │             │        │             │ │
│   │ Own config  │        │ Own config  │        │ Own config  │ │
│   │ Own tools   │        │ Own tools   │        │ Own tools   │ │
│   │ Own backends│        │ Own backends│        │ Own backends│ │
│   └──────┬──────┘        └──────┬──────┘        └──────┬──────┘ │
│          │                      │                      │         │
│          └──────────────────────┴──────────────────────┘         │
│                              │                                    │
│                              ▼                                    │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    TENANT ROUTER                         │   │
│   │          (Load balancer / API Gateway)                   │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tenant Storage

### TenantStore Interface

```go
// TenantStore manages tenant data
type TenantStore interface {
    // CRUD
    Get(ctx context.Context, id string) (*Tenant, error)
    Create(ctx context.Context, tenant *Tenant) error
    Update(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, opts ListOptions) ([]*Tenant, error)

    // Config
    GetConfig(ctx context.Context, id string) (*TenantConfig, error)
    UpdateConfig(ctx context.Context, id string, cfg *TenantConfig) error

    // Quotas
    GetQuotaUsage(ctx context.Context, id string) (*QuotaUsage, error)
    IncrementUsage(ctx context.Context, id string, metric string, delta int64) error
}

// Built-in implementations
type MemoryTenantStore struct { ... }      // For testing/development
type RedisTenantStore struct { ... }       // For distributed deployments
type PostgresTenantStore struct { ... }    // For persistent storage
type ConfigFileTenantStore struct { ... }  // For static configuration
```

---

## End-to-End Example

### Enterprise SaaS Configuration

```yaml
# metatools-saas.yaml

server:
  name: "metatools-saas"
  version: "1.0.0"

transport:
  type: sse
  http:
    port: 8080

tenancy:
  enabled: true

  resolver:
    type: composite
    resolvers:
      - type: jwt
        issuer: https://auth.saas.example.com
        claim_key: org_id
      - type: api_key
        header: X-API-Key

  store:
    type: postgres
    postgres:
      connection_string: ${DATABASE_URL}

  isolation: shared

  tiers:
    free:
      rate_limits:
        requests_per_minute: 30
      quotas:
        daily_requests: 500
      denied_tools:
        - execute_code
        - run_chain

    startup:
      rate_limits:
        requests_per_minute: 120
      quotas:
        daily_requests: 5000
      max_chain_steps: 5

    enterprise:
      rate_limits:
        requests_per_minute: 1000
      quotas:
        daily_requests: -1
      features:
        custom_backends: true
        dedicated_support: true

middleware:
  chain:
    - tenant          # Resolve tenant first
    - tenant_config   # Load tenant-specific config
    - tenant_filter   # Filter tools by tenant permissions
    - tenant_rate_limit
    - tenant_quota
    - tenant_audit
    - logging
    - metrics

backends:
  shared:
    github:
      enabled: true
      kind: mcp
      config:
        command: npx
        args: ["-y", "@modelcontextprotocol/server-github"]

    filesystem:
      enabled: true
      kind: mcp
      config:
        command: npx
        args: ["-y", "@modelcontextprotocol/server-filesystem"]

  # Enterprise tenants can have custom backends
  # Loaded from tenant config in database
```

### Request Flow

```
┌────────────────────────────────────────────────────────────────────────────┐
│                    MULTI-TENANT REQUEST FLOW                                │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   1. INCOMING REQUEST                                                        │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  POST /mcp HTTP/1.1                                                 │   │
│   │  Authorization: Bearer eyJhbGc...                                  │   │
│   │  X-API-Key: key_abc123 (fallback)                                  │   │
│   │                                                                     │   │
│   │  { "method": "tools/call", "params": { "name": "github/..." } }   │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│   2. TENANT RESOLUTION                                                       │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  JWT Resolver:                                                      │   │
│   │  - Validate token                                                   │   │
│   │  - Extract org_id: "acme-corp"                                     │   │
│   │  - Load tenant from store                                          │   │
│   │  - Inject TenantContext into ctx                                   │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│   TenantContext:                     │                                       │
│   - ID: "acme-corp"                  │                                       │
│   - Tier: "enterprise"               │                                       │
│   - Config: { allowed_backends: [...], features: {...} }                    │
│                                      │                                       │
│                                      ▼                                       │
│   3. TENANT TOOL FILTER                                                      │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  Check: Is "github/create_issue" allowed for acme-corp?            │   │
│   │  - Not in denied_tools ✓                                           │   │
│   │  - github backend is in allowed_backends ✓                         │   │
│   │  → ALLOWED                                                          │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│   4. TENANT RATE LIMIT                                                       │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  Key: "tenant:acme-corp:tool:github/create_issue"                  │   │
│   │  Limit: 1000 req/min (enterprise tier)                             │   │
│   │  Current: 42 → ALLOWED                                             │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│   5. EXECUTION (with tenant context)                                         │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  Route to: github backend                                          │   │
│   │  Execute: create_issue                                             │   │
│   │  Tenant-scoped audit log written                                   │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│   6. RESPONSE                                                                │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │  { "result": { "issue_number": 456 }, "id": "req-123" }           │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Priority

### Phase 1: Core Multi-Tenancy (2 weeks)

1. Define `Tenant`, `TenantContext`, `TenantConfig` types
2. Implement `TenantResolver` interface + JWT/API Key resolvers
3. Implement `TenantMiddleware` for context injection
4. Add `TenantFromContext()` helper

### Phase 2: Tenant-Aware Middleware (1 week)

1. Implement `TenantRateLimitMiddleware`
2. Implement `TenantToolFilterMiddleware`
3. Implement `TenantAuditMiddleware`

### Phase 3: Tenant Storage (1 week)

1. Define `TenantStore` interface
2. Implement `MemoryTenantStore` for development
3. Implement `PostgresTenantStore` for production
4. Add quota tracking

### Phase 4: Advanced Features (2 weeks)

1. Multi-tenant tool registry
2. Multi-tenant backend registry
3. Namespace isolation strategy
4. Process isolation strategy (optional)

---

## Changes to Component Libraries

### toolrun Changes

```go
// Add tenant context propagation
type RunOptions struct {
    // Existing...
    TenantID string  // NEW: Tenant context for execution
}
```

### toolindex Changes

```go
// Add tenant-scoped search
type Index interface {
    // Existing...
    SearchForTenant(tenantID string, query string, limit int) ([]Summary, error)  // NEW
}
```

---

## Summary

Multi-tenancy integrates cleanly into the pluggable architecture via:

1. **Pluggable TenantResolver** - Any identification strategy
2. **Tenant-Aware Middleware** - Transparent isolation via middleware chain
3. **Tenant-Scoped Registries** - Logical isolation of tools/backends
4. **Configuration Hierarchy** - Defaults → Tier → Tenant overrides
5. **Multiple Isolation Strategies** - Shared, namespace, or process isolation

All components remain pluggable and can be replaced or extended.

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-28 | Initial multi-tenancy extension proposal |
