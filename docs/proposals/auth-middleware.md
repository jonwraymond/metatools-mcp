# Pluggable Authentication & Authorization Middleware

**Status:** Draft
**Date:** 2026-01-30
**Related:** [Multi-Tenancy Proposal](./multi-tenancy.md), [Pluggable Architecture](./pluggable-architecture.md)

## Overview

This document defines a 100% pluggable authentication and authorization middleware for the metatools ecosystem. The design supports multiple auth providers (JWT, OAuth2, API Keys, mTLS), integrates with RBAC/ABAC systems (Casbin, OPA), and provides the foundation for multi-tenancy.

---

## Design Principles

1. **Interface-First** - All components defined as interfaces, implementations are swappable
2. **Zero Coupling** - Auth middleware knows nothing about specific providers
3. **Context Propagation** - Identity flows through context.Context
4. **Fail-Safe Defaults** - Deny by default, explicit allow required
5. **Composable** - Multiple authenticators/authorizers can chain together
6. **Observable** - Full audit trail of auth decisions
7. **Multi-Tenancy Ready** - Identity includes tenant context

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AUTH MIDDLEWARE ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   Request → [Transport Layer]                                                 │
│                    │                                                          │
│                    ▼                                                          │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    AUTHENTICATION MIDDLEWARE                         │   │
│   │                                                                       │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │              CompositeAuthenticator                           │   │   │
│   │   │   (tries each authenticator until one succeeds)               │   │   │
│   │   │                                                                │   │   │
│   │   │   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐            │   │   │
│   │   │   │   JWT   │→│ OAuth2  │→│ API Key │→│  mTLS   │            │   │   │
│   │   │   │  Auth   │ │  Auth   │ │  Auth   │ │  Auth   │            │   │   │
│   │   │   └─────────┘ └─────────┘ └─────────┘ └─────────┘            │   │   │
│   │   │                                                                │   │   │
│   │   │   Output: Identity (principal, tenant, claims, method)        │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                    │                                                          │
│                    ▼ Identity in context.Context                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    AUTHORIZATION MIDDLEWARE                          │   │
│   │                                                                       │   │
│   │   ┌──────────────────────────────────────────────────────────────┐   │   │
│   │   │              Authorizer (pluggable policy engine)             │   │   │
│   │   │                                                                │   │   │
│   │   │   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐            │   │   │
│   │   │   │ Casbin  │ │   OPA   │ │  RBAC   │ │ Custom  │            │   │   │
│   │   │   │Enforcer │ │ Rego    │ │ Simple  │ │ Policy  │            │   │   │
│   │   │   └─────────┘ └─────────┘ └─────────┘ └─────────┘            │   │   │
│   │   │                                                                │   │   │
│   │   │   Input: (subject, resource, action) → Output: allow/deny     │   │   │
│   │   └──────────────────────────────────────────────────────────────┘   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                    │                                                          │
│                    ▼ Authorized request                                      │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    TOOL PROVIDER (execution)                         │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Interfaces

### Identity Model

```go
// Identity represents an authenticated principal
type Identity struct {
    // Principal is the unique identifier (user ID, service account, etc.)
    Principal   string            `json:"principal"`

    // TenantID for multi-tenant systems (empty for single-tenant)
    TenantID    string            `json:"tenant_id,omitempty"`

    // Roles assigned to this identity
    Roles       []string          `json:"roles,omitempty"`

    // Permissions granted (for PBAC systems)
    Permissions []string          `json:"permissions,omitempty"`

    // Claims from the auth token (JWT claims, OAuth2 scopes, etc.)
    Claims      map[string]any    `json:"claims,omitempty"`

    // Method used for authentication
    Method      AuthMethod        `json:"method"`

    // Metadata for extensibility
    Metadata    map[string]string `json:"metadata,omitempty"`

    // ExpiresAt when the identity/session expires
    ExpiresAt   time.Time         `json:"expires_at,omitempty"`

    // IssuedAt when the identity was established
    IssuedAt    time.Time         `json:"issued_at"`
}

// AuthMethod identifies how the identity was authenticated
type AuthMethod string

const (
    AuthMethodJWT       AuthMethod = "jwt"
    AuthMethodOAuth2    AuthMethod = "oauth2"
    AuthMethodAPIKey    AuthMethod = "api_key"
    AuthMethodMTLS      AuthMethod = "mtls"
    AuthMethodBasic     AuthMethod = "basic"
    AuthMethodAnonymous AuthMethod = "anonymous"
    AuthMethodCustom    AuthMethod = "custom"
)

// AuthResult encapsulates authentication outcome
type AuthResult struct {
    // Authenticated indicates if authentication succeeded
    Authenticated bool

    // Identity is set when Authenticated is true
    Identity *Identity

    // Error describes why authentication failed (if applicable)
    Error error

    // Challenge is the WWW-Authenticate header value for 401 responses
    Challenge string
}
```

### Authenticator Interface

```go
// Authenticator validates credentials and returns an identity
// This is the core pluggable interface for authentication
type Authenticator interface {
    // Authenticate validates the request and returns an identity
    // Returns AuthResult with Authenticated=false if credentials are invalid
    // Returns error only for system failures (not auth failures)
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)

    // Name returns the authenticator name for logging/metrics
    Name() string

    // Supports checks if this authenticator handles the given request
    // Used by CompositeAuthenticator to select appropriate authenticator
    Supports(ctx context.Context, req *AuthRequest) bool
}

// AuthRequest contains request data for authentication
type AuthRequest struct {
    // Headers from the incoming request
    Headers map[string][]string

    // Method is the request method (tools/call, tools/list, etc.)
    Method string

    // Resource is the target resource (tool name, etc.)
    Resource string

    // RemoteAddr is the client IP address
    RemoteAddr string

    // TLSInfo contains mTLS certificate info (if available)
    TLSInfo *TLSInfo

    // Raw allows access to transport-specific request data
    Raw any
}

// TLSInfo contains client certificate information for mTLS
type TLSInfo struct {
    // PeerCertificates from the TLS connection
    PeerCertificates []*x509.Certificate

    // Verified indicates if the certificate chain was verified
    Verified bool

    // CommonName from the client certificate
    CommonName string

    // DNSNames from the client certificate
    DNSNames []string
}
```

### Authorizer Interface

```go
// Authorizer makes access control decisions
// Implementations can use RBAC, ABAC, ReBAC, or custom logic
type Authorizer interface {
    // Authorize checks if the identity can perform action on resource
    // Returns nil if authorized, error describing denial otherwise
    Authorize(ctx context.Context, req *AuthzRequest) error

    // Name returns the authorizer name for logging/metrics
    Name() string
}

// AuthzRequest contains the authorization request parameters
type AuthzRequest struct {
    // Subject is the identity requesting access
    Subject *Identity

    // Resource is what's being accessed (tool name, backend, etc.)
    Resource string

    // Action is what's being done (call, list, describe, etc.)
    Action string

    // ResourceType categorizes the resource (tool, backend, config, etc.)
    ResourceType string

    // Context provides additional attributes for ABAC
    Context map[string]any
}

// AuthzResult provides detailed authorization outcome
type AuthzResult struct {
    Allowed bool
    Reason  string
    Policy  string // Which policy made the decision
}

// Common authorization errors
var (
    ErrUnauthorized     = errors.New("unauthorized")
    ErrForbidden        = errors.New("forbidden")
    ErrInsufficientRole = errors.New("insufficient role")
    ErrInvalidToken     = errors.New("invalid token")
    ErrTokenExpired     = errors.New("token expired")
    ErrMissingClaims    = errors.New("missing required claims")
)
```

---

## Built-in Authenticators

### JWT Authenticator

```go
// JWTAuthenticator validates JWT tokens
type JWTAuthenticator struct {
    // Config holds JWT validation settings
    Config JWTConfig

    // KeyProvider fetches signing keys (supports JWKS, static, etc.)
    KeyProvider KeyProvider

    // ClaimsMapper extracts identity from JWT claims
    ClaimsMapper ClaimsMapper
}

// JWTConfig configures JWT validation
type JWTConfig struct {
    // Issuer expected in the token (validates iss claim)
    Issuer string `yaml:"issuer"`

    // Audience expected in the token (validates aud claim)
    Audience string `yaml:"audience"`

    // HeaderName where to find the token (default: Authorization)
    HeaderName string `yaml:"header_name"`

    // TokenPrefix expected before the token (default: Bearer)
    TokenPrefix string `yaml:"token_prefix"`

    // ClockSkew tolerance for expiration validation
    ClockSkew time.Duration `yaml:"clock_skew"`

    // RequiredClaims that must be present
    RequiredClaims []string `yaml:"required_claims"`

    // PrincipalClaim identifies the subject (default: sub)
    PrincipalClaim string `yaml:"principal_claim"`

    // TenantClaim identifies the tenant (optional)
    TenantClaim string `yaml:"tenant_claim"`

    // RolesClaim identifies roles (optional)
    RolesClaim string `yaml:"roles_claim"`

    // PermissionsClaim identifies permissions (optional)
    PermissionsClaim string `yaml:"permissions_claim"`
}

// KeyProvider fetches signing keys for JWT validation
type KeyProvider interface {
    // GetKey returns the key for the given key ID
    GetKey(ctx context.Context, keyID string) (any, error)

    // GetKeys returns all available keys
    GetKeys(ctx context.Context) ([]any, error)
}

// Built-in KeyProvider implementations
type StaticKeyProvider struct { ... }  // Single static key
type JWKSKeyProvider struct { ... }    // JWKS endpoint with caching
type FileKeyProvider struct { ... }    // Keys from filesystem

// ClaimsMapper extracts Identity from JWT claims
type ClaimsMapper interface {
    MapClaims(claims jwt.MapClaims) (*Identity, error)
}

// DefaultClaimsMapper uses JWTConfig to map claims
type DefaultClaimsMapper struct {
    Config JWTConfig
}
```

### API Key Authenticator

```go
// APIKeyAuthenticator validates API keys
type APIKeyAuthenticator struct {
    Config   APIKeyConfig
    Store    APIKeyStore
}

// APIKeyConfig configures API key validation
type APIKeyConfig struct {
    // HeaderName where to find the key (default: X-API-Key)
    HeaderName string `yaml:"header_name"`

    // QueryParam alternative location (optional)
    QueryParam string `yaml:"query_param"`

    // Prefix expected before the key (optional, e.g., "Bearer ")
    Prefix string `yaml:"prefix"`

    // HashAlgorithm for stored keys (none, sha256, bcrypt)
    HashAlgorithm string `yaml:"hash_algorithm"`
}

// APIKeyStore retrieves API key information
type APIKeyStore interface {
    // Lookup returns the identity for an API key
    // Returns nil, nil if key not found (not an error)
    Lookup(ctx context.Context, key string) (*APIKeyInfo, error)

    // Validate checks if a key is valid without full lookup
    Validate(ctx context.Context, key string) (bool, error)
}

// APIKeyInfo contains API key metadata
type APIKeyInfo struct {
    ID          string            // Key identifier
    Name        string            // Human-readable name
    Principal   string            // Owner identity
    TenantID    string            // Associated tenant
    Roles       []string          // Assigned roles
    Permissions []string          // Direct permissions
    Scopes      []string          // OAuth-style scopes
    Metadata    map[string]string // Custom metadata
    CreatedAt   time.Time
    ExpiresAt   *time.Time        // nil = never expires
    LastUsedAt  *time.Time
}

// Built-in APIKeyStore implementations
type MemoryAPIKeyStore struct { ... }     // For testing
type RedisAPIKeyStore struct { ... }      // For distributed deployments
type PostgresAPIKeyStore struct { ... }   // For persistent storage
type ConfigFileAPIKeyStore struct { ... } // For static configuration
```

### OAuth2 Authenticator

```go
// OAuth2Authenticator validates OAuth2 access tokens
type OAuth2Authenticator struct {
    Config       OAuth2Config
    TokenStore   TokenStore       // For opaque tokens
    Introspector TokenIntrospector // For introspection
}

// OAuth2Config configures OAuth2 validation
type OAuth2Config struct {
    // Issuer URL of the OAuth2 authorization server
    Issuer string `yaml:"issuer"`

    // IntrospectionURL for opaque token validation
    IntrospectionURL string `yaml:"introspection_url"`

    // ClientID for introspection authentication
    ClientID string `yaml:"client_id"`

    // ClientSecret for introspection authentication
    ClientSecret string `yaml:"client_secret"`

    // RequiredScopes that must be present
    RequiredScopes []string `yaml:"required_scopes"`

    // ScopesClaim where scopes are found in JWT tokens
    ScopesClaim string `yaml:"scopes_claim"`
}

// TokenIntrospector validates tokens via introspection endpoint
type TokenIntrospector interface {
    Introspect(ctx context.Context, token string) (*IntrospectionResult, error)
}

// IntrospectionResult from RFC 7662
type IntrospectionResult struct {
    Active    bool
    Scope     string
    ClientID  string
    Username  string
    TokenType string
    Exp       int64
    Iat       int64
    Nbf       int64
    Sub       string
    Aud       string
    Iss       string
}
```

### mTLS Authenticator

```go
// MTLSAuthenticator validates client certificates
type MTLSAuthenticator struct {
    Config      MTLSConfig
    CertMapper  CertificateMapper
}

// MTLSConfig configures mTLS validation
type MTLSConfig struct {
    // RequireClientCert enforces client certificate requirement
    RequireClientCert bool `yaml:"require_client_cert"`

    // TrustAnchors are CA certificates to trust
    TrustAnchors string `yaml:"trust_anchors"` // Path to CA bundle

    // AllowedCNs restricts accepted Common Names (optional)
    AllowedCNs []string `yaml:"allowed_cns"`

    // AllowedDNSNames restricts accepted DNS SANs (optional)
    AllowedDNSNames []string `yaml:"allowed_dns_names"`

    // AllowedOUs restricts accepted Organizational Units (optional)
    AllowedOUs []string `yaml:"allowed_ous"`

    // CRLFile for certificate revocation checking (optional)
    CRLFile string `yaml:"crl_file"`

    // OCSPResponder URL for online revocation checking (optional)
    OCSPResponder string `yaml:"ocsp_responder"`
}

// CertificateMapper extracts identity from client certificate
type CertificateMapper interface {
    MapCertificate(cert *x509.Certificate) (*Identity, error)
}

// DefaultCertMapper uses CN as principal, OU as tenant
type DefaultCertMapper struct { ... }
```

### Composite Authenticator

```go
// CompositeAuthenticator tries multiple authenticators in order
type CompositeAuthenticator struct {
    Authenticators []Authenticator

    // RequireAll requires all authenticators to succeed (MFA)
    RequireAll bool

    // StopOnFirst stops after first successful auth
    StopOnFirst bool
}

func (c *CompositeAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    var lastResult *AuthResult
    var lastError error

    for _, auth := range c.Authenticators {
        // Skip authenticators that don't support this request
        if !auth.Supports(ctx, req) {
            continue
        }

        result, err := auth.Authenticate(ctx, req)
        if err != nil {
            lastError = err
            continue
        }

        if result.Authenticated {
            if c.StopOnFirst {
                return result, nil
            }
            // For RequireAll, continue checking others
            lastResult = result
        } else {
            if c.RequireAll {
                return result, nil // Any failure = overall failure
            }
            lastResult = result
        }
    }

    if lastResult != nil && lastResult.Authenticated {
        return lastResult, nil
    }

    if lastResult != nil {
        return lastResult, nil
    }

    return &AuthResult{
        Authenticated: false,
        Challenge:     "Bearer",
        Error:         errors.Join(ErrUnauthorized, lastError),
    }, nil
}
```

---

## Built-in Authorizers

### Casbin Authorizer

```go
// CasbinAuthorizer uses Casbin for RBAC/ABAC/ReBAC
type CasbinAuthorizer struct {
    Enforcer *casbin.Enforcer
    Config   CasbinConfig
}

// CasbinConfig configures the Casbin enforcer
type CasbinConfig struct {
    // ModelPath is the path to the Casbin model file
    ModelPath string `yaml:"model_path"`

    // PolicyPath is the path to the policy file (for file adapter)
    PolicyPath string `yaml:"policy_path"`

    // Adapter configures policy storage
    Adapter CasbinAdapterConfig `yaml:"adapter"`

    // AutoLoad enables automatic policy reloading
    AutoLoad bool `yaml:"auto_load"`

    // AutoLoadInterval for policy refresh
    AutoLoadInterval time.Duration `yaml:"auto_load_interval"`
}

// CasbinAdapterConfig configures policy storage
type CasbinAdapterConfig struct {
    Type   string         `yaml:"type"` // file, postgres, mysql, redis
    Config map[string]any `yaml:"config"`
}

func (a *CasbinAuthorizer) Authorize(ctx context.Context, req *AuthzRequest) error {
    // Map to Casbin's (sub, obj, act) model
    sub := req.Subject.Principal
    obj := req.Resource
    act := req.Action

    // For multi-tenant, use domain
    var ok bool
    var err error

    if req.Subject.TenantID != "" {
        // RBAC with domains
        ok, err = a.Enforcer.Enforce(sub, req.Subject.TenantID, obj, act)
    } else {
        ok, err = a.Enforcer.Enforce(sub, obj, act)
    }

    if err != nil {
        return fmt.Errorf("casbin enforce: %w", err)
    }

    if !ok {
        return &AuthzError{
            Subject:  sub,
            Resource: obj,
            Action:   act,
            Reason:   "policy denied",
        }
    }

    return nil
}

// Example Casbin model for metatools (RBAC with domains)
// model.conf:
// [request_definition]
// r = sub, dom, obj, act
//
// [policy_definition]
// p = sub, dom, obj, act
//
// [role_definition]
// g = _, _, _
//
// [policy_effect]
// e = some(where (p.eft == allow))
//
// [matchers]
// m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
```

### OPA Authorizer

```go
// OPAAuthorizer uses Open Policy Agent for authorization
type OPAAuthorizer struct {
    Config OPAConfig
    Client *opa.Rego
}

// OPAConfig configures OPA
type OPAConfig struct {
    // PolicyPath is the path to Rego policy files
    PolicyPath string `yaml:"policy_path"`

    // BundleURL for fetching policy bundles
    BundleURL string `yaml:"bundle_url"`

    // Query is the Rego query to evaluate
    Query string `yaml:"query"`

    // DecisionPath in the policy output
    DecisionPath string `yaml:"decision_path"`
}

func (a *OPAAuthorizer) Authorize(ctx context.Context, req *AuthzRequest) error {
    input := map[string]any{
        "subject": map[string]any{
            "principal":   req.Subject.Principal,
            "tenant_id":   req.Subject.TenantID,
            "roles":       req.Subject.Roles,
            "permissions": req.Subject.Permissions,
        },
        "resource":      req.Resource,
        "action":        req.Action,
        "resource_type": req.ResourceType,
        "context":       req.Context,
    }

    results, err := a.Client.Eval(ctx, rego.EvalInput(input))
    if err != nil {
        return fmt.Errorf("opa eval: %w", err)
    }

    if len(results) == 0 || !results[0].Expressions[0].Value.(bool) {
        return &AuthzError{
            Subject:  req.Subject.Principal,
            Resource: req.Resource,
            Action:   req.Action,
            Reason:   "opa policy denied",
        }
    }

    return nil
}

// Example OPA policy for metatools:
// package metatools.authz
//
// default allow = false
//
// allow {
//     input.subject.roles[_] == "admin"
// }
//
// allow {
//     input.action == "tools/list"
// }
//
// allow {
//     input.action == "tools/call"
//     input.resource == input.subject.permissions[_]
// }
```

### Simple RBAC Authorizer

```go
// SimpleRBACAuthorizer provides basic RBAC without external dependencies
type SimpleRBACAuthorizer struct {
    Config SimpleRBACConfig
}

// SimpleRBACConfig defines roles and permissions
type SimpleRBACConfig struct {
    // Roles maps role names to permissions
    Roles map[string]RoleConfig `yaml:"roles"`

    // DefaultRole for unauthenticated or unassigned users
    DefaultRole string `yaml:"default_role"`
}

// RoleConfig defines a role's permissions
type RoleConfig struct {
    // Permissions granted to this role
    Permissions []string `yaml:"permissions"`

    // Inherits from other roles
    Inherits []string `yaml:"inherits"`

    // AllowedTools (tool name patterns)
    AllowedTools []string `yaml:"allowed_tools"`

    // DeniedTools (takes precedence)
    DeniedTools []string `yaml:"denied_tools"`

    // AllowedActions (call, list, describe, etc.)
    AllowedActions []string `yaml:"allowed_actions"`
}

func (a *SimpleRBACAuthorizer) Authorize(ctx context.Context, req *AuthzRequest) error {
    permissions := a.resolvePermissions(req.Subject)

    // Build permission key
    permKey := fmt.Sprintf("%s:%s", req.ResourceType, req.Action)
    toolKey := fmt.Sprintf("tool:%s", req.Resource)

    // Check explicit permission
    if slices.Contains(permissions, permKey) || slices.Contains(permissions, "*") {
        return nil
    }

    // Check tool-specific permission
    if slices.Contains(permissions, toolKey) {
        return nil
    }

    return &AuthzError{
        Subject:  req.Subject.Principal,
        Resource: req.Resource,
        Action:   req.Action,
        Reason:   "insufficient permissions",
    }
}
```

---

## Middleware Integration

### Authentication Middleware

```go
// AuthMiddleware wraps a ToolProvider with authentication
func AuthMiddleware(auth Authenticator, opts ...AuthMiddlewareOption) Middleware {
    cfg := defaultAuthMiddlewareConfig()
    for _, opt := range opts {
        opt(&cfg)
    }

    return func(next provider.ToolProvider) provider.ToolProvider {
        return &authMiddleware{
            auth:   auth,
            next:   next,
            config: cfg,
        }
    }
}

type authMiddleware struct {
    auth   Authenticator
    next   provider.ToolProvider
    config authMiddlewareConfig
}

type authMiddlewareConfig struct {
    // AllowAnonymous allows requests without credentials
    AllowAnonymous bool

    // AnonymousIdentity used when AllowAnonymous is true
    AnonymousIdentity *Identity

    // SkipPaths that don't require authentication
    SkipPaths []string

    // OnAuthFailure callback for custom error handling
    OnAuthFailure func(ctx context.Context, err error) error
}

func (m *authMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    // Build auth request from context
    req := buildAuthRequest(ctx)

    // Authenticate
    result, err := m.auth.Authenticate(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("authentication error: %w", err)
    }

    if !result.Authenticated {
        if m.config.AllowAnonymous {
            ctx = WithIdentity(ctx, m.config.AnonymousIdentity)
        } else {
            if m.config.OnAuthFailure != nil {
                return nil, m.config.OnAuthFailure(ctx, result.Error)
            }
            return nil, &AuthError{
                Code:      "UNAUTHENTICATED",
                Message:   "authentication required",
                Challenge: result.Challenge,
            }
        }
    } else {
        ctx = WithIdentity(ctx, result.Identity)
    }

    return m.next.Handle(ctx, input)
}

// Context helpers
func WithIdentity(ctx context.Context, id *Identity) context.Context {
    return context.WithValue(ctx, identityKey{}, id)
}

func IdentityFromContext(ctx context.Context) *Identity {
    if v := ctx.Value(identityKey{}); v != nil {
        return v.(*Identity)
    }
    return nil
}

func PrincipalFromContext(ctx context.Context) string {
    if id := IdentityFromContext(ctx); id != nil {
        return id.Principal
    }
    return ""
}

func TenantIDFromContext(ctx context.Context) string {
    if id := IdentityFromContext(ctx); id != nil {
        return id.TenantID
    }
    return ""
}
```

### Authorization Middleware

```go
// AuthzMiddleware wraps a ToolProvider with authorization
func AuthzMiddleware(authz Authorizer, opts ...AuthzMiddlewareOption) Middleware {
    cfg := defaultAuthzMiddlewareConfig()
    for _, opt := range opts {
        opt(&cfg)
    }

    return func(next provider.ToolProvider) provider.ToolProvider {
        return &authzMiddleware{
            authz:  authz,
            next:   next,
            config: cfg,
        }
    }
}

type authzMiddleware struct {
    authz  Authorizer
    next   provider.ToolProvider
    config authzMiddlewareConfig
}

type authzMiddlewareConfig struct {
    // ResourceResolver extracts resource from the request
    ResourceResolver func(ctx context.Context, input map[string]any) string

    // ActionResolver extracts action from the request
    ActionResolver func(ctx context.Context, input map[string]any) string

    // ContextBuilder adds context for ABAC decisions
    ContextBuilder func(ctx context.Context, input map[string]any) map[string]any

    // OnDenied callback for custom denial handling
    OnDenied func(ctx context.Context, err error) error
}

func (m *authzMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    identity := IdentityFromContext(ctx)
    if identity == nil {
        return nil, &AuthError{Code: "UNAUTHENTICATED", Message: "no identity in context"}
    }

    req := &AuthzRequest{
        Subject:      identity,
        Resource:     m.config.ResourceResolver(ctx, input),
        Action:       m.config.ActionResolver(ctx, input),
        ResourceType: "tool",
        Context:      m.config.ContextBuilder(ctx, input),
    }

    if err := m.authz.Authorize(ctx, req); err != nil {
        if m.config.OnDenied != nil {
            return nil, m.config.OnDenied(ctx, err)
        }
        return nil, &AuthError{
            Code:    "FORBIDDEN",
            Message: err.Error(),
        }
    }

    return m.next.Handle(ctx, input)
}

func (m *authzMiddleware) Name() string { return m.next.Name() }
func (m *authzMiddleware) Description() string { return m.next.Description() }
func (m *authzMiddleware) InputSchema() map[string]any { return m.next.InputSchema() }
```

---

## Factory Registration

```go
// Register auth middleware factories
func init() {
    // Authentication middleware
    middleware.DefaultRegistry.Register("auth", func(cfg map[string]any) (middleware.Middleware, error) {
        authCfg, err := parseAuthConfig(cfg)
        if err != nil {
            return nil, err
        }

        auth, err := buildAuthenticator(authCfg)
        if err != nil {
            return nil, err
        }

        return AuthMiddleware(auth), nil
    })

    // Authorization middleware
    middleware.DefaultRegistry.Register("authz", func(cfg map[string]any) (middleware.Middleware, error) {
        authzCfg, err := parseAuthzConfig(cfg)
        if err != nil {
            return nil, err
        }

        authz, err := buildAuthorizer(authzCfg)
        if err != nil {
            return nil, err
        }

        return AuthzMiddleware(authz), nil
    })

    // Combined auth+authz middleware
    middleware.DefaultRegistry.Register("secure", func(cfg map[string]any) (middleware.Middleware, error) {
        auth, err := buildAuthenticator(cfg)
        if err != nil {
            return nil, err
        }

        authz, err := buildAuthorizer(cfg)
        if err != nil {
            return nil, err
        }

        return func(next provider.ToolProvider) provider.ToolProvider {
            return AuthMiddleware(auth)(AuthzMiddleware(authz)(next))
        }, nil
    })
}
```

---

## Configuration

### YAML Configuration

```yaml
# metatools.yaml - Authentication & Authorization

auth:
  # Authentication configuration
  authentication:
    # Type: jwt, oauth2, api_key, mtls, composite
    type: composite

    # Composite authenticator configuration
    composite:
      stop_on_first: true
      authenticators:
        # JWT authentication (primary)
        - type: jwt
          issuer: https://auth.example.com
          audience: metatools-api
          jwks_url: https://auth.example.com/.well-known/jwks.json
          principal_claim: sub
          tenant_claim: org_id
          roles_claim: roles
          clock_skew: 5m

        # API Key authentication (secondary)
        - type: api_key
          header_name: X-API-Key
          store:
            type: redis
            config:
              address: localhost:6379
              prefix: "apikey:"

        # mTLS (for service-to-service)
        - type: mtls
          require_client_cert: false  # Optional
          trust_anchors: /etc/ssl/ca-bundle.crt
          allowed_cns:
            - "*.internal.example.com"

  # Authorization configuration
  authorization:
    # Type: casbin, opa, simple_rbac, custom
    type: casbin

    casbin:
      model_path: /etc/metatools/rbac_model.conf
      adapter:
        type: postgres
        config:
          connection_string: ${DATABASE_URL}
      auto_load: true
      auto_load_interval: 30s

    # Alternative: Simple RBAC (no external dependencies)
    # simple_rbac:
    #   default_role: anonymous
    #   roles:
    #     admin:
    #       permissions: ["*"]
    #     user:
    #       allowed_actions: [call, list, describe]
    #       allowed_tools: ["search_*", "describe_*"]
    #       denied_tools: ["execute_code"]
    #     anonymous:
    #       allowed_actions: [list]

    # Alternative: OPA
    # opa:
    #   policy_path: /etc/metatools/policies
    #   query: data.metatools.authz.allow

  # Middleware options
  middleware:
    allow_anonymous: false
    skip_paths:
      - health
      - version

# Middleware chain includes auth
middleware:
  chain:
    - auth         # Authentication (resolves identity)
    - tenant       # Tenant resolution (from identity)
    - authz        # Authorization (checks permissions)
    - rate_limit   # Rate limiting (per-tenant)
    - logging
    - metrics
```

### Casbin Model Example

```ini
# rbac_with_domains_model.conf
# For multi-tenant RBAC

[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && r.act == p.act
```

### Casbin Policy Example

```csv
# policy.csv
# Format: p, subject, domain(tenant), object(tool), action

# Admin role has full access in all tenants
p, role:admin, *, *, *

# Users can list and describe tools
p, role:user, *, *, list
p, role:user, *, *, describe

# Users can call specific tools
p, role:user, *, search_*, call
p, role:user, *, describe_*, call

# Enterprise tenants can execute code
p, role:user, tenant:enterprise-*, execute_code, call

# Role assignments
g, alice, role:admin, tenant:acme-corp
g, bob, role:user, tenant:acme-corp
g, service-account, role:admin, *
```

---

## CLI Integration

```bash
# Serve with authentication enabled
metatools serve --auth.type=jwt --auth.jwt.issuer=https://auth.example.com

# Generate API key
metatools auth create-key --name="my-api-key" --tenant="acme-corp" --roles=user

# Validate token
metatools auth validate-token --token="eyJ..."

# List API keys
metatools auth list-keys --tenant="acme-corp"

# Revoke API key
metatools auth revoke-key --id="key_abc123"

# Test authorization
metatools auth test --principal="alice" --tenant="acme-corp" --resource="execute_code" --action="call"
```

---

## Implementation Priority

### Phase 1: Core Interfaces (1 week)

1. Define `Identity`, `AuthResult`, `AuthRequest`, `AuthzRequest` types
2. Define `Authenticator`, `Authorizer` interfaces
3. Implement context helpers (`WithIdentity`, `IdentityFromContext`, etc.)
4. Implement `AuthMiddleware` and `AuthzMiddleware`

### Phase 2: JWT & API Key Authenticators (1 week)

1. Implement `JWTAuthenticator` with JWKS support
2. Implement `APIKeyAuthenticator` with pluggable stores
3. Implement `CompositeAuthenticator`
4. Add factory registration

### Phase 3: RBAC Authorizers (1 week)

1. Implement `SimpleRBACAuthorizer` (no dependencies)
2. Implement `CasbinAuthorizer` (optional dependency)
3. Add configuration parsing and validation

### Phase 4: Advanced Features (1 week)

1. Implement `OAuth2Authenticator` with token introspection
2. Implement `MTLSAuthenticator`
3. Implement `OPAAuthorizer` (optional dependency)
4. Add CLI commands for key management

---

## Testing Strategy

```go
// Mock authenticator for testing
type MockAuthenticator struct {
    AuthenticateFunc func(ctx context.Context, req *AuthRequest) (*AuthResult, error)
    SupportsFunc     func(ctx context.Context, req *AuthRequest) bool
}

// Test helper
func TestIdentity(principal, tenant string, roles ...string) *Identity {
    return &Identity{
        Principal: principal,
        TenantID:  tenant,
        Roles:     roles,
        Method:    AuthMethodJWT,
        IssuedAt:  time.Now(),
    }
}

// Example test
func TestAuthMiddleware_ValidJWT(t *testing.T) {
    auth := &MockAuthenticator{
        AuthenticateFunc: func(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
            return &AuthResult{
                Authenticated: true,
                Identity:      TestIdentity("alice", "acme-corp", "user"),
            }, nil
        },
        SupportsFunc: func(ctx context.Context, req *AuthRequest) bool {
            return true
        },
    }

    mw := AuthMiddleware(auth)
    wrapped := mw(mockProvider{})

    result, err := wrapped.Handle(context.Background(), nil)
    require.NoError(t, err)
    // ... assertions
}
```

---

## Summary

This design provides:

1. **100% Pluggable Authentication** - Any credential type via `Authenticator` interface
2. **100% Pluggable Authorization** - Any policy engine via `Authorizer` interface
3. **Multi-Provider Support** - JWT, OAuth2, API Keys, mTLS all swappable
4. **RBAC/ABAC/ReBAC** - Casbin, OPA, or simple built-in RBAC
5. **Multi-Tenancy Foundation** - Identity includes tenant context
6. **Zero Coupling** - Middleware knows nothing about specific implementations
7. **Configuration-Driven** - Full YAML configuration support
8. **Observable** - Audit logging and metrics built-in

All components follow the existing middleware chain architecture and can be replaced or extended without code changes.

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-30 | Initial auth middleware proposal based on research |
