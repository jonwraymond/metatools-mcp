# PRD-012: Multi-Tenancy Core Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement multi-tenancy support with pluggable tenant resolution, tenant-aware middleware, and tenant-scoped registries.

**Architecture:** Tenant context injection via middleware, with pluggable resolvers (JWT, API Key, Header) and configurable isolation strategies (shared, namespace, process).

**Tech Stack:** Go, JWT validation (golang-jwt), Redis (tenant storage option)

---

## Overview

Multi-tenancy enables a single metatools-mcp deployment to serve multiple organizations with isolated tool access, rate limits, and audit trails.

**Reference:** [multi-tenancy.md](../proposals/multi-tenancy.md)

---

## Directory Structure

```
internal/tenancy/
├── tenant.go           # Tenant and TenantContext types
├── tenant_test.go
├── resolver.go         # TenantResolver interface
├── resolver_test.go
├── resolvers/
│   ├── jwt.go          # JWT-based resolver
│   ├── jwt_test.go
│   ├── apikey.go       # API key resolver
│   ├── apikey_test.go
│   ├── header.go       # Header-based resolver
│   ├── composite.go    # Composite resolver
│   └── composite_test.go
├── middleware.go       # Tenant middleware
├── middleware_test.go
├── store.go            # TenantStore interface
├── store_test.go
├── stores/
│   ├── memory.go       # In-memory store
│   ├── memory_test.go
│   ├── config.go       # Config-file store
│   └── postgres.go     # PostgreSQL store
├── config.go           # Configuration types
└── doc.go
```

---

## Task 1: Tenant and TenantContext Types

**Files:**
- Create: `internal/tenancy/tenant.go`
- Create: `internal/tenancy/tenant_test.go`

**Step 1: Write failing tests**

```go
// tenant_test.go
package tenancy_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/metatools-mcp/internal/tenancy"
)

func TestTenant_Validate(t *testing.T) {
    tests := []struct {
        name    string
        tenant  tenancy.Tenant
        wantErr bool
    }{
        {
            name: "valid tenant",
            tenant: tenancy.Tenant{
                ID:   "acme-corp",
                Name: "Acme Corporation",
                Tier: tenancy.TierPro,
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            tenant: tenancy.Tenant{
                Name: "Acme Corporation",
            },
            wantErr: true,
        },
        {
            name: "invalid tier",
            tenant: tenancy.Tenant{
                ID:   "test",
                Tier: "invalid",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.tenant.Validate()
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestTenantContext_FromContext(t *testing.T) {
    tenant := &tenancy.Tenant{
        ID:   "test-tenant",
        Name: "Test",
        Tier: tenancy.TierFree,
    }
    config := &tenancy.TenantConfig{
        AllowedTools: []string{"search", "describe"},
    }

    tc := &tenancy.TenantContext{
        Tenant: tenant,
        Config: config,
    }

    ctx := tenancy.WithTenantContext(context.Background(), tc)

    // Retrieve from context
    retrieved := tenancy.TenantFromContext(ctx)
    require.NotNil(t, retrieved)
    assert.Equal(t, "test-tenant", retrieved.Tenant.ID)
}

func TestTenantContext_FromContextMissing(t *testing.T) {
    ctx := context.Background()

    retrieved := tenancy.TenantFromContext(ctx)
    assert.Nil(t, retrieved)
}

func TestTenantID_FromContext(t *testing.T) {
    tc := &tenancy.TenantContext{
        Tenant: &tenancy.Tenant{ID: "my-tenant"},
    }
    ctx := tenancy.WithTenantContext(context.Background(), tc)

    id := tenancy.TenantIDFromContext(ctx)
    assert.Equal(t, "my-tenant", id)
}

func TestTenantConfig_IsToolAllowed(t *testing.T) {
    tests := []struct {
        name     string
        config   tenancy.TenantConfig
        toolID   string
        expected bool
    }{
        {
            name:     "empty config allows all",
            config:   tenancy.TenantConfig{},
            toolID:   "any-tool",
            expected: true,
        },
        {
            name: "allowed list - tool present",
            config: tenancy.TenantConfig{
                AllowedTools: []string{"search", "describe"},
            },
            toolID:   "search",
            expected: true,
        },
        {
            name: "allowed list - tool absent",
            config: tenancy.TenantConfig{
                AllowedTools: []string{"search", "describe"},
            },
            toolID:   "execute",
            expected: false,
        },
        {
            name: "denied list takes precedence",
            config: tenancy.TenantConfig{
                AllowedTools: []string{"execute"},
                DeniedTools:  []string{"execute"},
            },
            toolID:   "execute",
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.config.IsToolAllowed(tt.toolID))
        })
    }
}

func TestTenantTier_Constants(t *testing.T) {
    assert.Equal(t, tenancy.TenantTier("free"), tenancy.TierFree)
    assert.Equal(t, tenancy.TenantTier("pro"), tenancy.TierPro)
    assert.Equal(t, tenancy.TenantTier("enterprise"), tenancy.TierEnterprise)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tenancy/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// tenant.go
package tenancy

import (
    "context"
    "errors"
    "slices"
    "time"
)

type contextKey string

const tenantContextKey contextKey = "tenant"

// TenantTier defines service levels
type TenantTier string

const (
    TierFree       TenantTier = "free"
    TierPro        TenantTier = "pro"
    TierEnterprise TenantTier = "enterprise"
)

// ValidTiers contains all valid tier values
var ValidTiers = []TenantTier{TierFree, TierPro, TierEnterprise}

// Tenant represents a tenant in the system
type Tenant struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Tier      TenantTier        `json:"tier"`
    Metadata  map[string]any    `json:"metadata,omitempty"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
}

// Validate validates the tenant
func (t *Tenant) Validate() error {
    if t.ID == "" {
        return errors.New("tenant ID is required")
    }
    if t.Tier != "" && !slices.Contains(ValidTiers, t.Tier) {
        return errors.New("invalid tenant tier")
    }
    return nil
}

// TenantContext holds runtime tenant information
type TenantContext struct {
    Tenant      *Tenant
    Permissions []string
    Config      *TenantConfig
    Quotas      *TenantQuotas
}

// TenantConfig holds per-tenant configuration overrides
type TenantConfig struct {
    // Tool access control
    AllowedTools    []string `json:"allowed_tools,omitempty"`
    DeniedTools     []string `json:"denied_tools,omitempty"`
    AllowedBackends []string `json:"allowed_backends,omitempty"`
    DeniedBackends  []string `json:"denied_backends,omitempty"`

    // Resource limits
    RateLimits *RateLimitConfig `json:"rate_limits,omitempty"`
    Quotas     *QuotaConfig     `json:"quotas,omitempty"`

    // Feature flags
    Features map[string]bool `json:"features,omitempty"`

    // Execution limits
    MaxTimeout    time.Duration `json:"max_timeout,omitempty"`
    MaxChainSteps int           `json:"max_chain_steps,omitempty"`
    MaxToolCalls  int           `json:"max_tool_calls,omitempty"`

    // Custom middleware config
    MiddlewareConfig map[string]any `json:"middleware_config,omitempty"`
}

// IsToolAllowed checks if a tool is allowed for this tenant
func (c *TenantConfig) IsToolAllowed(toolID string) bool {
    // Denied list takes precedence
    if c.DeniedTools != nil && slices.Contains(c.DeniedTools, toolID) {
        return false
    }

    // If allowed list is specified, tool must be in it
    if len(c.AllowedTools) > 0 {
        return slices.Contains(c.AllowedTools, toolID)
    }

    // No restrictions
    return true
}

// IsBackendAllowed checks if a backend is allowed for this tenant
func (c *TenantConfig) IsBackendAllowed(backendName string) bool {
    if c.DeniedBackends != nil && slices.Contains(c.DeniedBackends, backendName) {
        return false
    }
    if len(c.AllowedBackends) > 0 {
        return slices.Contains(c.AllowedBackends, backendName)
    }
    return true
}

// RateLimitConfig defines rate limits
type RateLimitConfig struct {
    RequestsPerMinute int `json:"requests_per_minute"`
    Burst             int `json:"burst"`
}

// QuotaConfig defines usage quotas
type QuotaConfig struct {
    DailyRequests    int64 `json:"daily_requests"`
    DailyToolCalls   int64 `json:"daily_tool_calls"`
    MonthlyRequests  int64 `json:"monthly_requests"`
    MonthlyToolCalls int64 `json:"monthly_tool_calls"`
}

// TenantQuotas holds current quota state
type TenantQuotas struct {
    DailyRequestsUsed    int64
    DailyToolCallsUsed   int64
    MonthlyRequestsUsed  int64
    MonthlyToolCallsUsed int64
    ResetAt              time.Time
}

// WithTenantContext adds tenant context to the context
func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
    return context.WithValue(ctx, tenantContextKey, tc)
}

// TenantFromContext retrieves tenant context from context
func TenantFromContext(ctx context.Context) *TenantContext {
    tc, _ := ctx.Value(tenantContextKey).(*TenantContext)
    return tc
}

// TenantIDFromContext retrieves tenant ID from context
func TenantIDFromContext(ctx context.Context) string {
    tc := TenantFromContext(ctx)
    if tc == nil || tc.Tenant == nil {
        return ""
    }
    return tc.Tenant.ID
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tenancy/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tenancy/
git commit -m "$(cat <<'EOF'
feat(tenancy): add Tenant and TenantContext types

- Tenant with ID, Name, Tier, Metadata
- TenantContext for runtime tenant info
- TenantConfig with tool/backend access control
- RateLimitConfig and QuotaConfig
- Context helpers for tenant propagation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: TenantResolver Interface and JWT Resolver

**Files:**
- Create: `internal/tenancy/resolver.go`
- Create: `internal/tenancy/resolvers/jwt.go`
- Create: `internal/tenancy/resolvers/jwt_test.go`

**Step 1: Write failing tests**

```go
// resolvers/jwt_test.go
package resolvers_test

import (
    "context"
    "testing"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/metatools-mcp/internal/tenancy"
    "github.com/jrraymond/metatools-mcp/internal/tenancy/resolvers"
)

func TestJWTResolver_Resolve(t *testing.T) {
    secret := []byte("test-secret")

    store := &MockTenantStore{
        tenants: map[string]*tenancy.Tenant{
            "acme-corp": {
                ID:   "acme-corp",
                Name: "Acme Corporation",
                Tier: tenancy.TierPro,
            },
        },
        configs: map[string]*tenancy.TenantConfig{
            "acme-corp": {
                AllowedTools: []string{"search", "describe"},
            },
        },
    }

    resolver := resolvers.NewJWTResolver(resolvers.JWTConfig{
        ClaimKey: "tenant_id",
        Secret:   secret,
        Store:    store,
    })

    // Create valid JWT
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "tenant_id": "acme-corp",
        "exp":       time.Now().Add(time.Hour).Unix(),
    })
    tokenString, _ := token.SignedString(secret)

    req := &tenancy.Request{
        Headers: map[string]string{
            "Authorization": "Bearer " + tokenString,
        },
    }

    tc, err := resolver.Resolve(context.Background(), req)
    require.NoError(t, err)
    require.NotNil(t, tc)
    assert.Equal(t, "acme-corp", tc.Tenant.ID)
}

func TestJWTResolver_InvalidToken(t *testing.T) {
    resolver := resolvers.NewJWTResolver(resolvers.JWTConfig{
        ClaimKey: "tenant_id",
        Secret:   []byte("secret"),
    })

    req := &tenancy.Request{
        Headers: map[string]string{
            "Authorization": "Bearer invalid-token",
        },
    }

    _, err := resolver.Resolve(context.Background(), req)
    require.Error(t, err)
}

func TestJWTResolver_MissingHeader(t *testing.T) {
    resolver := resolvers.NewJWTResolver(resolvers.JWTConfig{
        ClaimKey: "tenant_id",
        Secret:   []byte("secret"),
    })

    req := &tenancy.Request{
        Headers: map[string]string{},
    }

    _, err := resolver.Resolve(context.Background(), req)
    require.Error(t, err)
}

func TestJWTResolver_MissingClaim(t *testing.T) {
    secret := []byte("secret")
    resolver := resolvers.NewJWTResolver(resolvers.JWTConfig{
        ClaimKey: "tenant_id",
        Secret:   secret,
    })

    // Token without tenant_id claim
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": "user-123",
        "exp":     time.Now().Add(time.Hour).Unix(),
    })
    tokenString, _ := token.SignedString(secret)

    req := &tenancy.Request{
        Headers: map[string]string{
            "Authorization": "Bearer " + tokenString,
        },
    }

    _, err := resolver.Resolve(context.Background(), req)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "tenant claim")
}

// MockTenantStore for testing
type MockTenantStore struct {
    tenants map[string]*tenancy.Tenant
    configs map[string]*tenancy.TenantConfig
}

func (s *MockTenantStore) Get(ctx context.Context, id string) (*tenancy.Tenant, error) {
    t, ok := s.tenants[id]
    if !ok {
        return nil, tenancy.ErrTenantNotFound
    }
    return t, nil
}

func (s *MockTenantStore) GetConfig(ctx context.Context, id string) (*tenancy.TenantConfig, error) {
    c, ok := s.configs[id]
    if !ok {
        return &tenancy.TenantConfig{}, nil
    }
    return c, nil
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tenancy/resolvers/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// resolver.go
package tenancy

import (
    "context"
    "errors"
)

// Errors
var (
    ErrNoTenantResolved = errors.New("no tenant could be resolved")
    ErrTenantNotFound   = errors.New("tenant not found")
    ErrNoTenantClaim    = errors.New("no tenant claim in token")
    ErrInvalidToken     = errors.New("invalid or expired token")
)

// Request represents an incoming request for tenant resolution
type Request struct {
    Headers map[string]string
    Query   map[string]string
    Body    []byte
}

// TenantResolver identifies the tenant from a request
type TenantResolver interface {
    // Resolve extracts tenant information from the request context
    // Returns nil tenant for anonymous/default tenant
    Resolve(ctx context.Context, req *Request) (*TenantContext, error)
}

// TenantResolverFunc is a convenience type
type TenantResolverFunc func(ctx context.Context, req *Request) (*TenantContext, error)

func (f TenantResolverFunc) Resolve(ctx context.Context, req *Request) (*TenantContext, error) {
    return f(ctx, req)
}
```

```go
// resolvers/jwt.go
package resolvers

import (
    "context"
    "errors"
    "strings"

    "github.com/golang-jwt/jwt/v5"
    "github.com/jrraymond/metatools-mcp/internal/tenancy"
)

// JWTConfig holds JWT resolver configuration
type JWTConfig struct {
    ClaimKey string
    Secret   []byte
    Issuer   string
    Audience string
    Store    TenantStore
}

// TenantStore interface for JWT resolver
type TenantStore interface {
    Get(ctx context.Context, id string) (*tenancy.Tenant, error)
    GetConfig(ctx context.Context, id string) (*tenancy.TenantConfig, error)
}

// JWTResolver extracts tenant from JWT claims
type JWTResolver struct {
    config JWTConfig
}

// NewJWTResolver creates a new JWT resolver
func NewJWTResolver(config JWTConfig) *JWTResolver {
    return &JWTResolver{config: config}
}

// Resolve extracts tenant from JWT token
func (r *JWTResolver) Resolve(ctx context.Context, req *tenancy.Request) (*tenancy.TenantContext, error) {
    // Extract token from Authorization header
    authHeader := req.Headers["Authorization"]
    if authHeader == "" {
        return nil, errors.New("missing Authorization header")
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    if tokenString == authHeader {
        return nil, errors.New("invalid Authorization header format")
    }

    // Parse and validate token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return r.config.Secret, nil
    })

    if err != nil {
        return nil, tenancy.ErrInvalidToken
    }

    if !token.Valid {
        return nil, tenancy.ErrInvalidToken
    }

    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }

    // Get tenant ID from claims
    tenantID, ok := claims[r.config.ClaimKey].(string)
    if !ok || tenantID == "" {
        return nil, tenancy.ErrNoTenantClaim
    }

    // Lookup tenant from store
    var tenant *tenancy.Tenant
    var config *tenancy.TenantConfig

    if r.config.Store != nil {
        tenant, err = r.config.Store.Get(ctx, tenantID)
        if err != nil {
            return nil, err
        }
        config, _ = r.config.Store.GetConfig(ctx, tenantID)
    } else {
        // Minimal tenant without store
        tenant = &tenancy.Tenant{
            ID: tenantID,
        }
        config = &tenancy.TenantConfig{}
    }

    // Extract permissions from claims
    var permissions []string
    if perms, ok := claims["permissions"].([]any); ok {
        for _, p := range perms {
            if s, ok := p.(string); ok {
                permissions = append(permissions, s)
            }
        }
    }

    return &tenancy.TenantContext{
        Tenant:      tenant,
        Permissions: permissions,
        Config:      config,
    }, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tenancy/resolvers/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tenancy/
git commit -m "$(cat <<'EOF'
feat(tenancy): add TenantResolver interface and JWT resolver

- TenantResolver interface for pluggable resolution
- JWTResolver extracts tenant from JWT claims
- Token validation with configurable secret
- Permission extraction from claims
- Integration with TenantStore

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Tenant Middleware

**Files:**
- Create: `internal/tenancy/middleware.go`
- Create: `internal/tenancy/middleware_test.go`

**Step 1: Write failing tests**

```go
// middleware_test.go
package tenancy_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/metatools-mcp/internal/tenancy"
)

// MockResolver for testing
type MockResolver struct {
    tc  *tenancy.TenantContext
    err error
}

func (r *MockResolver) Resolve(ctx context.Context, req *tenancy.Request) (*tenancy.TenantContext, error) {
    return r.tc, r.err
}

// MockToolProvider for testing
type MockToolProvider struct {
    name      string
    capturedCtx context.Context
    result    any
    err       error
}

func (p *MockToolProvider) Name() string { return p.name }
func (p *MockToolProvider) Handle(ctx context.Context, input map[string]any) (any, error) {
    p.capturedCtx = ctx
    return p.result, p.err
}

func TestTenantMiddleware_InjectsContext(t *testing.T) {
    tc := &tenancy.TenantContext{
        Tenant: &tenancy.Tenant{ID: "test-tenant"},
        Config: &tenancy.TenantConfig{},
    }

    resolver := &MockResolver{tc: tc}
    provider := &MockToolProvider{name: "test:tool", result: "ok"}

    middleware := tenancy.TenantMiddleware(resolver, nil)
    wrapped := middleware(provider)

    _, err := wrapped.Handle(context.Background(), nil)
    require.NoError(t, err)

    // Verify tenant context was injected
    injected := tenancy.TenantFromContext(provider.capturedCtx)
    require.NotNil(t, injected)
    assert.Equal(t, "test-tenant", injected.Tenant.ID)
}

func TestTenantToolFilterMiddleware_AllowedTool(t *testing.T) {
    tc := &tenancy.TenantContext{
        Tenant: &tenancy.Tenant{ID: "test"},
        Config: &tenancy.TenantConfig{
            AllowedTools: []string{"test:tool"},
        },
    }

    ctx := tenancy.WithTenantContext(context.Background(), tc)
    provider := &MockToolProvider{name: "test:tool", result: "ok"}

    middleware := tenancy.TenantToolFilterMiddleware()
    wrapped := middleware(provider)

    result, err := wrapped.Handle(ctx, nil)
    require.NoError(t, err)
    assert.Equal(t, "ok", result)
}

func TestTenantToolFilterMiddleware_DeniedTool(t *testing.T) {
    tc := &tenancy.TenantContext{
        Tenant: &tenancy.Tenant{ID: "test"},
        Config: &tenancy.TenantConfig{
            DeniedTools: []string{"test:dangerous"},
        },
    }

    ctx := tenancy.WithTenantContext(context.Background(), tc)
    provider := &MockToolProvider{name: "test:dangerous", result: "ok"}

    middleware := tenancy.TenantToolFilterMiddleware()
    wrapped := middleware(provider)

    _, err := wrapped.Handle(ctx, nil)
    require.Error(t, err)

    var toolErr *tenancy.ToolDeniedError
    require.ErrorAs(t, err, &toolErr)
    assert.Equal(t, "test:dangerous", toolErr.Tool)
}

func TestTenantRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
    tc := &tenancy.TenantContext{
        Tenant: &tenancy.Tenant{ID: "test"},
        Config: &tenancy.TenantConfig{
            RateLimits: &tenancy.RateLimitConfig{
                RequestsPerMinute: 100,
                Burst:             10,
            },
        },
    }

    store := tenancy.NewMemoryRateLimitStore()
    ctx := tenancy.WithTenantContext(context.Background(), tc)
    provider := &MockToolProvider{name: "test:tool", result: "ok"}

    middleware := tenancy.TenantRateLimitMiddleware(store, nil)
    wrapped := middleware(provider)

    // Should allow within limit
    result, err := wrapped.Handle(ctx, nil)
    require.NoError(t, err)
    assert.Equal(t, "ok", result)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tenancy/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// middleware.go
package tenancy

import (
    "context"
    "fmt"
)

// ToolProvider interface for middleware
type ToolProvider interface {
    Name() string
    Handle(ctx context.Context, input map[string]any) (any, error)
}

// Middleware wraps a ToolProvider
type Middleware func(ToolProvider) ToolProvider

// ToolDeniedError indicates a tool was denied for a tenant
type ToolDeniedError struct {
    TenantID string
    Tool     string
    Reason   string
}

func (e *ToolDeniedError) Error() string {
    return fmt.Sprintf("tool %q denied for tenant %q: %s", e.Tool, e.TenantID, e.Reason)
}

// RateLimitError indicates rate limit exceeded
type RateLimitError struct {
    TenantID string
    Limit    int
}

func (e *RateLimitError) Error() string {
    return fmt.Sprintf("rate limit exceeded for tenant %q: limit %d", e.TenantID, e.Limit)
}

// tenantMiddleware injects tenant context
type tenantMiddleware struct {
    resolver TenantResolver
    defaults *TenantContext
    next     ToolProvider
}

// TenantMiddleware creates middleware that resolves and injects tenant context
func TenantMiddleware(resolver TenantResolver, defaults *TenantContext) Middleware {
    return func(next ToolProvider) ToolProvider {
        return &tenantMiddleware{
            resolver: resolver,
            defaults: defaults,
            next:     next,
        }
    }
}

func (m *tenantMiddleware) Name() string {
    return m.next.Name()
}

func (m *tenantMiddleware) Handle(ctx context.Context, input map[string]any) (any, error) {
    // Build request from context
    req := RequestFromContext(ctx)

    // Resolve tenant
    tc, err := m.resolver.Resolve(ctx, req)
    if err != nil {
        if m.defaults != nil {
            tc = m.defaults
        } else {
            return nil, fmt.Errorf("tenant resolution failed: %w", err)
        }
    }

    // Inject tenant context
    ctx = WithTenantContext(ctx, tc)

    return m.next.Handle(ctx, input)
}

// toolFilterMiddleware filters tools based on tenant permissions
type toolFilterMiddleware struct {
    next ToolProvider
}

// TenantToolFilterMiddleware creates middleware that filters tools by tenant
func TenantToolFilterMiddleware() Middleware {
    return func(next ToolProvider) ToolProvider {
        return &toolFilterMiddleware{next: next}
    }
}

func (m *toolFilterMiddleware) Name() string {
    return m.next.Name()
}

func (m *toolFilterMiddleware) Handle(ctx context.Context, input map[string]any) (any, error) {
    tc := TenantFromContext(ctx)
    if tc == nil || tc.Config == nil {
        return m.next.Handle(ctx, input)
    }

    toolName := m.next.Name()
    if !tc.Config.IsToolAllowed(toolName) {
        return nil, &ToolDeniedError{
            TenantID: tc.Tenant.ID,
            Tool:     toolName,
            Reason:   "tool not allowed for tenant",
        }
    }

    return m.next.Handle(ctx, input)
}

// RateLimitStore interface for rate limit storage
type RateLimitStore interface {
    Allow(ctx context.Context, key string, limit, burst int) (bool, error)
}

// rateLimitMiddleware applies per-tenant rate limits
type rateLimitMiddleware struct {
    store    RateLimitStore
    defaults *RateLimitConfig
    next     ToolProvider
}

// TenantRateLimitMiddleware creates middleware that enforces rate limits
func TenantRateLimitMiddleware(store RateLimitStore, defaults *RateLimitConfig) Middleware {
    return func(next ToolProvider) ToolProvider {
        return &rateLimitMiddleware{
            store:    store,
            defaults: defaults,
            next:     next,
        }
    }
}

func (m *rateLimitMiddleware) Name() string {
    return m.next.Name()
}

func (m *rateLimitMiddleware) Handle(ctx context.Context, input map[string]any) (any, error) {
    tc := TenantFromContext(ctx)
    if tc == nil {
        return m.next.Handle(ctx, input)
    }

    // Get rate limit config
    limits := tc.Config.RateLimits
    if limits == nil {
        limits = m.defaults
    }
    if limits == nil {
        return m.next.Handle(ctx, input)
    }

    // Check rate limit
    key := fmt.Sprintf("tenant:%s:tool:%s", tc.Tenant.ID, m.next.Name())
    allowed, err := m.store.Allow(ctx, key, limits.RequestsPerMinute, limits.Burst)
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

// MemoryRateLimitStore is an in-memory rate limit store
type MemoryRateLimitStore struct {
    // Simplified - real implementation would use token bucket
    counts map[string]int
}

// NewMemoryRateLimitStore creates a new memory rate limit store
func NewMemoryRateLimitStore() *MemoryRateLimitStore {
    return &MemoryRateLimitStore{
        counts: make(map[string]int),
    }
}

func (s *MemoryRateLimitStore) Allow(ctx context.Context, key string, limit, burst int) (bool, error) {
    s.counts[key]++
    return s.counts[key] <= limit, nil
}

// RequestFromContext builds a Request from context
func RequestFromContext(ctx context.Context) *Request {
    // Extract headers from context if available
    // This would be set by the transport layer
    return &Request{
        Headers: make(map[string]string),
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tenancy/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tenancy/
git commit -m "$(cat <<'EOF'
feat(tenancy): add tenant middleware

- TenantMiddleware for context injection
- TenantToolFilterMiddleware for access control
- TenantRateLimitMiddleware for rate limiting
- ToolDeniedError and RateLimitError types
- MemoryRateLimitStore for development

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-012 complete:

- [ ] All tests pass: `go test ./internal/tenancy/... -v`
- [ ] Code coverage > 80%
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete
- [ ] Integration verified:
  - [ ] Tenant types validate correctly
  - [ ] JWT resolver extracts tenant
  - [ ] Middleware injects context
  - [ ] Tool filtering works

---

## Definition of Done

1. **Tenant** and **TenantContext** types
2. **TenantConfig** with tool/backend access control
3. **TenantResolver** interface with JWT resolver
4. **TenantMiddleware** for context injection
5. **TenantToolFilterMiddleware** for access control
6. **TenantRateLimitMiddleware** for rate limiting
7. All tests passing with >80% coverage
8. Documentation complete
