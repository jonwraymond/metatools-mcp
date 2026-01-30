# PRD-017: Pluggable Authentication & Authorization Middleware

**Status:** Ready for Implementation
**Date:** 2026-01-30
**Priority:** P1 (High)
**Depends On:** PRD-016 (Interface Contracts)
**Proposal:** [auth-middleware.md](../proposals/auth-middleware.md)
**Architecture:** [persistence-boundary.md](../proposals/persistence-boundary.md)

---

## Overview

Implement a 100% pluggable authentication and authorization middleware following the persistence boundary architecture. Core interfaces live in `internal/auth`, implementations are pluggable via factory registration.

---

## Scope

### In Scope

- Core auth interfaces (`Authenticator`, `Authorizer`, `Identity`)
- JWT authenticator with JWKS support
- API Key authenticator with pluggable store interface
- Simple RBAC authorizer (no external dependencies)
- Casbin authorizer (optional, build-tagged)
- Auth middleware integration with existing middleware chain
- Context propagation helpers
- YAML configuration support
- CLI flags for basic auth configuration

### Out of Scope (Future PRDs)

- OAuth2 token introspection (PRD-018)
- mTLS authenticator (PRD-019)
- OPA authorizer (PRD-020)
- Persistence implementations (toolpersist library)
- Multi-tenancy integration (after auth foundation)

---

## Implementation Tasks

### Task 1: Core Identity Types

**Files:**
- `internal/auth/identity.go`
- `internal/auth/identity_test.go`

```go
// internal/auth/identity.go
package auth

import (
    "context"
    "time"
)

// Identity represents an authenticated principal.
type Identity struct {
    Principal   string            `json:"principal"`
    TenantID    string            `json:"tenant_id,omitempty"`
    Roles       []string          `json:"roles,omitempty"`
    Permissions []string          `json:"permissions,omitempty"`
    Claims      map[string]any    `json:"claims,omitempty"`
    Method      AuthMethod        `json:"method"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    ExpiresAt   time.Time         `json:"expires_at,omitempty"`
    IssuedAt    time.Time         `json:"issued_at"`
}

// AuthMethod identifies how the identity was authenticated.
type AuthMethod string

const (
    AuthMethodJWT       AuthMethod = "jwt"
    AuthMethodOAuth2    AuthMethod = "oauth2"
    AuthMethodAPIKey    AuthMethod = "api_key"
    AuthMethodMTLS      AuthMethod = "mtls"
    AuthMethodBasic     AuthMethod = "basic"
    AuthMethodAnonymous AuthMethod = "anonymous"
)

// HasRole checks if the identity has a specific role.
func (id *Identity) HasRole(role string) bool {
    for _, r := range id.Roles {
        if r == role {
            return true
        }
    }
    return false
}

// HasPermission checks if the identity has a specific permission.
func (id *Identity) HasPermission(perm string) bool {
    for _, p := range id.Permissions {
        if p == perm {
            return true
        }
    }
    return false
}

// HasAnyRole checks if the identity has any of the specified roles.
func (id *Identity) HasAnyRole(roles ...string) bool {
    for _, role := range roles {
        if id.HasRole(role) {
            return true
        }
    }
    return false
}

// IsExpired checks if the identity has expired.
func (id *Identity) IsExpired() bool {
    if id.ExpiresAt.IsZero() {
        return false
    }
    return time.Now().After(id.ExpiresAt)
}

// Context key for identity storage.
type identityKey struct{}

// WithIdentity adds an identity to the context.
func WithIdentity(ctx context.Context, id *Identity) context.Context {
    return context.WithValue(ctx, identityKey{}, id)
}

// IdentityFromContext retrieves the identity from context.
func IdentityFromContext(ctx context.Context) *Identity {
    if v := ctx.Value(identityKey{}); v != nil {
        return v.(*Identity)
    }
    return nil
}

// PrincipalFromContext retrieves the principal from context.
func PrincipalFromContext(ctx context.Context) string {
    if id := IdentityFromContext(ctx); id != nil {
        return id.Principal
    }
    return ""
}

// TenantIDFromContext retrieves the tenant ID from context.
func TenantIDFromContext(ctx context.Context) string {
    if id := IdentityFromContext(ctx); id != nil {
        return id.TenantID
    }
    return ""
}
```

**Tests:**

```go
// internal/auth/identity_test.go
package auth

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestIdentity_HasRole(t *testing.T) {
    id := &Identity{Roles: []string{"admin", "user"}}

    assert.True(t, id.HasRole("admin"))
    assert.True(t, id.HasRole("user"))
    assert.False(t, id.HasRole("superuser"))
}

func TestIdentity_HasPermission(t *testing.T) {
    id := &Identity{Permissions: []string{"read:tools", "write:tools"}}

    assert.True(t, id.HasPermission("read:tools"))
    assert.False(t, id.HasPermission("delete:tools"))
}

func TestIdentity_HasAnyRole(t *testing.T) {
    id := &Identity{Roles: []string{"user"}}

    assert.True(t, id.HasAnyRole("admin", "user"))
    assert.False(t, id.HasAnyRole("admin", "superuser"))
}

func TestIdentity_IsExpired(t *testing.T) {
    t.Run("not expired", func(t *testing.T) {
        id := &Identity{ExpiresAt: time.Now().Add(time.Hour)}
        assert.False(t, id.IsExpired())
    })

    t.Run("expired", func(t *testing.T) {
        id := &Identity{ExpiresAt: time.Now().Add(-time.Hour)}
        assert.True(t, id.IsExpired())
    })

    t.Run("no expiry", func(t *testing.T) {
        id := &Identity{}
        assert.False(t, id.IsExpired())
    })
}

func TestIdentityContext(t *testing.T) {
    ctx := context.Background()

    // No identity in context
    assert.Nil(t, IdentityFromContext(ctx))
    assert.Empty(t, PrincipalFromContext(ctx))
    assert.Empty(t, TenantIDFromContext(ctx))

    // Add identity
    id := &Identity{
        Principal: "alice",
        TenantID:  "acme-corp",
    }
    ctx = WithIdentity(ctx, id)

    // Retrieve identity
    retrieved := IdentityFromContext(ctx)
    require.NotNil(t, retrieved)
    assert.Equal(t, "alice", retrieved.Principal)
    assert.Equal(t, "acme-corp", retrieved.TenantID)

    // Convenience helpers
    assert.Equal(t, "alice", PrincipalFromContext(ctx))
    assert.Equal(t, "acme-corp", TenantIDFromContext(ctx))
}
```

**Verification:**

```bash
go test ./internal/auth/... -run TestIdentity -v
```

---

### Task 2: Authenticator Interface & Types

**Files:**
- `internal/auth/authenticator.go`
- `internal/auth/authenticator_test.go`

```go
// internal/auth/authenticator.go
package auth

import (
    "context"
    "crypto/x509"
    "errors"
)

// Common authentication errors.
var (
    ErrUnauthorized       = errors.New("unauthorized")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrInvalidToken       = errors.New("invalid token")
    ErrTokenExpired       = errors.New("token expired")
    ErrMissingCredentials = errors.New("missing credentials")
    ErrUnsupportedMethod  = errors.New("unsupported authentication method")
)

// AuthRequest contains request data for authentication.
type AuthRequest struct {
    Headers    map[string][]string
    Method     string
    Resource   string
    RemoteAddr string
    TLSInfo    *TLSInfo
    Raw        any
}

// TLSInfo contains client certificate information.
type TLSInfo struct {
    PeerCertificates []*x509.Certificate
    Verified         bool
    CommonName       string
    DNSNames         []string
}

// GetHeader retrieves a header value.
func (r *AuthRequest) GetHeader(name string) string {
    if values, ok := r.Headers[name]; ok && len(values) > 0 {
        return values[0]
    }
    return ""
}

// AuthResult encapsulates authentication outcome.
type AuthResult struct {
    Authenticated bool
    Identity      *Identity
    Error         error
    Challenge     string
}

// Authenticator validates credentials and returns an identity.
type Authenticator interface {
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)
    Name() string
    Supports(ctx context.Context, req *AuthRequest) bool
}

// AuthenticatorFunc is a function adapter for Authenticator.
type AuthenticatorFunc func(ctx context.Context, req *AuthRequest) (*AuthResult, error)

// Authenticate implements Authenticator.
func (f AuthenticatorFunc) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    return f(ctx, req)
}

// Name implements Authenticator.
func (f AuthenticatorFunc) Name() string { return "func" }

// Supports implements Authenticator.
func (f AuthenticatorFunc) Supports(ctx context.Context, req *AuthRequest) bool { return true }
```

**Tests:**

```go
// internal/auth/authenticator_test.go
package auth

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestAuthRequest_GetHeader(t *testing.T) {
    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer token123"},
            "X-API-Key":     {"key456"},
        },
    }

    assert.Equal(t, "Bearer token123", req.GetHeader("Authorization"))
    assert.Equal(t, "key456", req.GetHeader("X-API-Key"))
    assert.Empty(t, req.GetHeader("Missing"))
}

func TestAuthenticatorFunc(t *testing.T) {
    called := false
    auth := AuthenticatorFunc(func(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
        called = true
        return &AuthResult{Authenticated: true}, nil
    })

    result, err := auth.Authenticate(context.Background(), &AuthRequest{})
    assert.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.True(t, called)
    assert.Equal(t, "func", auth.Name())
    assert.True(t, auth.Supports(context.Background(), &AuthRequest{}))
}
```

---

### Task 3: Composite Authenticator

**Files:**
- `internal/auth/composite.go`
- `internal/auth/composite_test.go`

```go
// internal/auth/composite.go
package auth

import (
    "context"
    "errors"
)

// CompositeAuthenticator tries multiple authenticators in order.
type CompositeAuthenticator struct {
    Authenticators []Authenticator
    StopOnFirst    bool
}

// NewCompositeAuthenticator creates a composite authenticator.
func NewCompositeAuthenticator(authenticators ...Authenticator) *CompositeAuthenticator {
    return &CompositeAuthenticator{
        Authenticators: authenticators,
        StopOnFirst:    true,
    }
}

// Authenticate implements Authenticator.
func (c *CompositeAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    var lastResult *AuthResult
    var errs []error

    for _, auth := range c.Authenticators {
        if !auth.Supports(ctx, req) {
            continue
        }

        result, err := auth.Authenticate(ctx, req)
        if err != nil {
            errs = append(errs, err)
            continue
        }

        if result.Authenticated {
            if c.StopOnFirst {
                return result, nil
            }
            lastResult = result
        } else {
            lastResult = result
        }
    }

    if lastResult != nil && lastResult.Authenticated {
        return lastResult, nil
    }

    if lastResult != nil {
        return lastResult, nil
    }

    // No authenticator succeeded
    return &AuthResult{
        Authenticated: false,
        Challenge:     "Bearer",
        Error:         errors.Join(append([]error{ErrUnauthorized}, errs...)...),
    }, nil
}

// Name implements Authenticator.
func (c *CompositeAuthenticator) Name() string {
    return "composite"
}

// Supports implements Authenticator.
func (c *CompositeAuthenticator) Supports(ctx context.Context, req *AuthRequest) bool {
    for _, auth := range c.Authenticators {
        if auth.Supports(ctx, req) {
            return true
        }
    }
    return false
}
```

**Tests:**

```go
// internal/auth/composite_test.go
package auth

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCompositeAuthenticator_FirstSucceeds(t *testing.T) {
    auth1 := &mockAuthenticator{
        name:     "auth1",
        supports: true,
        result:   &AuthResult{Authenticated: true, Identity: &Identity{Principal: "user1"}},
    }
    auth2 := &mockAuthenticator{
        name:     "auth2",
        supports: true,
        result:   &AuthResult{Authenticated: true, Identity: &Identity{Principal: "user2"}},
    }

    composite := NewCompositeAuthenticator(auth1, auth2)
    result, err := composite.Authenticate(context.Background(), &AuthRequest{})

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "user1", result.Identity.Principal)
    assert.True(t, auth1.called)
    assert.False(t, auth2.called) // StopOnFirst
}

func TestCompositeAuthenticator_FallbackToSecond(t *testing.T) {
    auth1 := &mockAuthenticator{
        name:     "auth1",
        supports: true,
        result:   &AuthResult{Authenticated: false},
    }
    auth2 := &mockAuthenticator{
        name:     "auth2",
        supports: true,
        result:   &AuthResult{Authenticated: true, Identity: &Identity{Principal: "user2"}},
    }

    composite := NewCompositeAuthenticator(auth1, auth2)
    result, err := composite.Authenticate(context.Background(), &AuthRequest{})

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "user2", result.Identity.Principal)
}

func TestCompositeAuthenticator_AllFail(t *testing.T) {
    auth1 := &mockAuthenticator{
        name:     "auth1",
        supports: true,
        result:   &AuthResult{Authenticated: false},
    }
    auth2 := &mockAuthenticator{
        name:     "auth2",
        supports: true,
        result:   &AuthResult{Authenticated: false},
    }

    composite := NewCompositeAuthenticator(auth1, auth2)
    result, err := composite.Authenticate(context.Background(), &AuthRequest{})

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrUnauthorized)
}

func TestCompositeAuthenticator_SkipsUnsupported(t *testing.T) {
    auth1 := &mockAuthenticator{
        name:     "auth1",
        supports: false, // Does not support
        result:   &AuthResult{Authenticated: true},
    }
    auth2 := &mockAuthenticator{
        name:     "auth2",
        supports: true,
        result:   &AuthResult{Authenticated: true, Identity: &Identity{Principal: "user2"}},
    }

    composite := NewCompositeAuthenticator(auth1, auth2)
    result, err := composite.Authenticate(context.Background(), &AuthRequest{})

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "user2", result.Identity.Principal)
    assert.False(t, auth1.called)
    assert.True(t, auth2.called)
}

func TestCompositeAuthenticator_HandlesErrors(t *testing.T) {
    auth1 := &mockAuthenticator{
        name:     "auth1",
        supports: true,
        err:      errors.New("network error"),
    }
    auth2 := &mockAuthenticator{
        name:     "auth2",
        supports: true,
        result:   &AuthResult{Authenticated: true, Identity: &Identity{Principal: "user2"}},
    }

    composite := NewCompositeAuthenticator(auth1, auth2)
    result, err := composite.Authenticate(context.Background(), &AuthRequest{})

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "user2", result.Identity.Principal)
}

// mockAuthenticator for testing
type mockAuthenticator struct {
    name     string
    supports bool
    result   *AuthResult
    err      error
    called   bool
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    m.called = true
    return m.result, m.err
}

func (m *mockAuthenticator) Name() string { return m.name }

func (m *mockAuthenticator) Supports(ctx context.Context, req *AuthRequest) bool {
    return m.supports
}
```

---

### Task 4: JWT Authenticator

**Files:**
- `internal/auth/jwt.go`
- `internal/auth/jwt_test.go`

```go
// internal/auth/jwt.go
package auth

import (
    "context"
    "crypto/rsa"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// JWTConfig configures JWT validation.
type JWTConfig struct {
    Issuer           string        `yaml:"issuer"`
    Audience         string        `yaml:"audience"`
    HeaderName       string        `yaml:"header_name"`
    TokenPrefix      string        `yaml:"token_prefix"`
    ClockSkew        time.Duration `yaml:"clock_skew"`
    RequiredClaims   []string      `yaml:"required_claims"`
    PrincipalClaim   string        `yaml:"principal_claim"`
    TenantClaim      string        `yaml:"tenant_claim"`
    RolesClaim       string        `yaml:"roles_claim"`
    PermissionsClaim string        `yaml:"permissions_claim"`
}

// DefaultJWTConfig returns sensible defaults.
func DefaultJWTConfig() JWTConfig {
    return JWTConfig{
        HeaderName:     "Authorization",
        TokenPrefix:    "Bearer ",
        ClockSkew:      5 * time.Minute,
        PrincipalClaim: "sub",
    }
}

// KeyProvider fetches signing keys for JWT validation.
type KeyProvider interface {
    GetKey(ctx context.Context, keyID string) (any, error)
    GetKeys(ctx context.Context) ([]any, error)
}

// StaticKeyProvider uses a single static key.
type StaticKeyProvider struct {
    Key any
}

// GetKey implements KeyProvider.
func (p *StaticKeyProvider) GetKey(ctx context.Context, keyID string) (any, error) {
    return p.Key, nil
}

// GetKeys implements KeyProvider.
func (p *StaticKeyProvider) GetKeys(ctx context.Context) ([]any, error) {
    return []any{p.Key}, nil
}

// JWTAuthenticator validates JWT tokens.
type JWTAuthenticator struct {
    Config      JWTConfig
    KeyProvider KeyProvider
}

// NewJWTAuthenticator creates a JWT authenticator.
func NewJWTAuthenticator(cfg JWTConfig, keyProvider KeyProvider) *JWTAuthenticator {
    if cfg.HeaderName == "" {
        cfg.HeaderName = "Authorization"
    }
    if cfg.TokenPrefix == "" {
        cfg.TokenPrefix = "Bearer "
    }
    if cfg.PrincipalClaim == "" {
        cfg.PrincipalClaim = "sub"
    }
    return &JWTAuthenticator{
        Config:      cfg,
        KeyProvider: keyProvider,
    }
}

// Authenticate implements Authenticator.
func (a *JWTAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    // Extract token from header
    header := req.GetHeader(a.Config.HeaderName)
    if header == "" {
        return &AuthResult{
            Authenticated: false,
            Challenge:     "Bearer",
            Error:         ErrMissingCredentials,
        }, nil
    }

    // Remove prefix
    if !strings.HasPrefix(header, a.Config.TokenPrefix) {
        return &AuthResult{
            Authenticated: false,
            Challenge:     "Bearer",
            Error:         ErrInvalidToken,
        }, nil
    }
    tokenString := strings.TrimPrefix(header, a.Config.TokenPrefix)

    // Parse and validate token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        // Get key ID from header
        kid, _ := token.Header["kid"].(string)
        return a.KeyProvider.GetKey(ctx, kid)
    }, jwt.WithLeeway(a.Config.ClockSkew))

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return &AuthResult{
                Authenticated: false,
                Challenge:     "Bearer error=\"invalid_token\", error_description=\"token expired\"",
                Error:         ErrTokenExpired,
            }, nil
        }
        return &AuthResult{
            Authenticated: false,
            Challenge:     "Bearer error=\"invalid_token\"",
            Error:         fmt.Errorf("%w: %v", ErrInvalidToken, err),
        }, nil
    }

    if !token.Valid {
        return &AuthResult{
            Authenticated: false,
            Challenge:     "Bearer error=\"invalid_token\"",
            Error:         ErrInvalidToken,
        }, nil
    }

    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return &AuthResult{
            Authenticated: false,
            Error:         errors.New("invalid claims format"),
        }, nil
    }

    // Validate issuer
    if a.Config.Issuer != "" {
        if iss, _ := claims["iss"].(string); iss != a.Config.Issuer {
            return &AuthResult{
                Authenticated: false,
                Error:         errors.New("invalid issuer"),
            }, nil
        }
    }

    // Validate audience
    if a.Config.Audience != "" {
        audClaim := claims["aud"]
        valid := false
        switch aud := audClaim.(type) {
        case string:
            valid = aud == a.Config.Audience
        case []any:
            for _, a := range aud {
                if s, ok := a.(string); ok && s == a.Config.Audience {
                    valid = true
                    break
                }
            }
        }
        if !valid {
            return &AuthResult{
                Authenticated: false,
                Error:         errors.New("invalid audience"),
            }, nil
        }
    }

    // Check required claims
    for _, claim := range a.Config.RequiredClaims {
        if _, ok := claims[claim]; !ok {
            return &AuthResult{
                Authenticated: false,
                Error:         fmt.Errorf("missing required claim: %s", claim),
            }, nil
        }
    }

    // Build identity
    identity := a.buildIdentity(claims)

    return &AuthResult{
        Authenticated: true,
        Identity:      identity,
    }, nil
}

func (a *JWTAuthenticator) buildIdentity(claims jwt.MapClaims) *Identity {
    id := &Identity{
        Method:   AuthMethodJWT,
        Claims:   make(map[string]any),
        IssuedAt: time.Now(),
    }

    // Principal
    if principal, ok := claims[a.Config.PrincipalClaim].(string); ok {
        id.Principal = principal
    }

    // Tenant
    if a.Config.TenantClaim != "" {
        if tenant, ok := claims[a.Config.TenantClaim].(string); ok {
            id.TenantID = tenant
        }
    }

    // Roles
    if a.Config.RolesClaim != "" {
        if roles, ok := claims[a.Config.RolesClaim].([]any); ok {
            for _, r := range roles {
                if s, ok := r.(string); ok {
                    id.Roles = append(id.Roles, s)
                }
            }
        }
    }

    // Permissions
    if a.Config.PermissionsClaim != "" {
        if perms, ok := claims[a.Config.PermissionsClaim].([]any); ok {
            for _, p := range perms {
                if s, ok := p.(string); ok {
                    id.Permissions = append(id.Permissions, s)
                }
            }
        }
    }

    // Expiration
    if exp, ok := claims["exp"].(float64); ok {
        id.ExpiresAt = time.Unix(int64(exp), 0)
    }

    // Copy all claims
    for k, v := range claims {
        id.Claims[k] = v
    }

    return id
}

// Name implements Authenticator.
func (a *JWTAuthenticator) Name() string {
    return "jwt"
}

// Supports implements Authenticator.
func (a *JWTAuthenticator) Supports(ctx context.Context, req *AuthRequest) bool {
    header := req.GetHeader(a.Config.HeaderName)
    return strings.HasPrefix(header, a.Config.TokenPrefix)
}

// NewStaticKeyProvider creates a key provider with a static key.
func NewStaticKeyProvider(key any) *StaticKeyProvider {
    return &StaticKeyProvider{Key: key}
}

// NewRSAKeyProvider creates a key provider with an RSA public key.
func NewRSAKeyProvider(key *rsa.PublicKey) *StaticKeyProvider {
    return &StaticKeyProvider{Key: key}
}

// NewHMACKeyProvider creates a key provider with an HMAC secret.
func NewHMACKeyProvider(secret []byte) *StaticKeyProvider {
    return &StaticKeyProvider{Key: secret}
}
```

**Tests:**

```go
// internal/auth/jwt_test.go
package auth

import (
    "context"
    "testing"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

var testSecret = []byte("test-secret-key-for-jwt-testing")

func createTestToken(claims jwt.MapClaims) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(testSecret)
    return tokenString
}

func TestJWTAuthenticator_ValidToken(t *testing.T) {
    auth := NewJWTAuthenticator(
        JWTConfig{
            Issuer:         "test-issuer",
            Audience:       "test-audience",
            PrincipalClaim: "sub",
            RolesClaim:     "roles",
        },
        NewHMACKeyProvider(testSecret),
    )

    token := createTestToken(jwt.MapClaims{
        "sub":   "alice",
        "iss":   "test-issuer",
        "aud":   "test-audience",
        "exp":   time.Now().Add(time.Hour).Unix(),
        "roles": []any{"admin", "user"},
    })

    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer " + token},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    require.NotNil(t, result.Identity)
    assert.Equal(t, "alice", result.Identity.Principal)
    assert.Equal(t, AuthMethodJWT, result.Identity.Method)
    assert.Contains(t, result.Identity.Roles, "admin")
    assert.Contains(t, result.Identity.Roles, "user")
}

func TestJWTAuthenticator_ExpiredToken(t *testing.T) {
    auth := NewJWTAuthenticator(
        DefaultJWTConfig(),
        NewHMACKeyProvider(testSecret),
    )

    token := createTestToken(jwt.MapClaims{
        "sub": "alice",
        "exp": time.Now().Add(-time.Hour).Unix(), // Expired
    })

    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer " + token},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrTokenExpired)
}

func TestJWTAuthenticator_InvalidSignature(t *testing.T) {
    auth := NewJWTAuthenticator(
        DefaultJWTConfig(),
        NewHMACKeyProvider([]byte("wrong-secret")),
    )

    token := createTestToken(jwt.MapClaims{
        "sub": "alice",
        "exp": time.Now().Add(time.Hour).Unix(),
    })

    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer " + token},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrInvalidToken)
}

func TestJWTAuthenticator_MissingToken(t *testing.T) {
    auth := NewJWTAuthenticator(
        DefaultJWTConfig(),
        NewHMACKeyProvider(testSecret),
    )

    req := &AuthRequest{
        Headers: map[string][]string{},
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestJWTAuthenticator_InvalidIssuer(t *testing.T) {
    auth := NewJWTAuthenticator(
        JWTConfig{
            Issuer: "expected-issuer",
        },
        NewHMACKeyProvider(testSecret),
    )

    token := createTestToken(jwt.MapClaims{
        "sub": "alice",
        "iss": "wrong-issuer",
        "exp": time.Now().Add(time.Hour).Unix(),
    })

    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer " + token},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
}

func TestJWTAuthenticator_TenantClaim(t *testing.T) {
    auth := NewJWTAuthenticator(
        JWTConfig{
            PrincipalClaim: "sub",
            TenantClaim:    "org_id",
        },
        NewHMACKeyProvider(testSecret),
    )

    token := createTestToken(jwt.MapClaims{
        "sub":    "alice",
        "org_id": "acme-corp",
        "exp":    time.Now().Add(time.Hour).Unix(),
    })

    req := &AuthRequest{
        Headers: map[string][]string{
            "Authorization": {"Bearer " + token},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "acme-corp", result.Identity.TenantID)
}

func TestJWTAuthenticator_Supports(t *testing.T) {
    auth := NewJWTAuthenticator(DefaultJWTConfig(), NewHMACKeyProvider(testSecret))

    t.Run("supports bearer token", func(t *testing.T) {
        req := &AuthRequest{
            Headers: map[string][]string{
                "Authorization": {"Bearer token"},
            },
        }
        assert.True(t, auth.Supports(context.Background(), req))
    })

    t.Run("does not support api key", func(t *testing.T) {
        req := &AuthRequest{
            Headers: map[string][]string{
                "X-API-Key": {"key123"},
            },
        }
        assert.False(t, auth.Supports(context.Background(), req))
    })

    t.Run("does not support basic auth", func(t *testing.T) {
        req := &AuthRequest{
            Headers: map[string][]string{
                "Authorization": {"Basic dXNlcjpwYXNz"},
            },
        }
        assert.False(t, auth.Supports(context.Background(), req))
    })
}
```

---

### Task 5: API Key Authenticator

**Files:**
- `internal/auth/apikey.go`
- `internal/auth/apikey_test.go`

```go
// internal/auth/apikey.go
package auth

import (
    "context"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/hex"
    "strings"
    "time"
)

// APIKeyConfig configures API key validation.
type APIKeyConfig struct {
    HeaderName    string `yaml:"header_name"`
    QueryParam    string `yaml:"query_param"`
    Prefix        string `yaml:"prefix"`
    HashAlgorithm string `yaml:"hash_algorithm"` // none, sha256
}

// DefaultAPIKeyConfig returns sensible defaults.
func DefaultAPIKeyConfig() APIKeyConfig {
    return APIKeyConfig{
        HeaderName:    "X-API-Key",
        HashAlgorithm: "none",
    }
}

// APIKeyInfo contains API key metadata.
type APIKeyInfo struct {
    ID          string
    Name        string
    KeyHash     string // Hashed key for storage
    Principal   string
    TenantID    string
    Roles       []string
    Permissions []string
    Scopes      []string
    Metadata    map[string]string
    CreatedAt   time.Time
    ExpiresAt   *time.Time
    LastUsedAt  *time.Time
}

// IsExpired checks if the API key has expired.
func (k *APIKeyInfo) IsExpired() bool {
    if k.ExpiresAt == nil {
        return false
    }
    return time.Now().After(*k.ExpiresAt)
}

// APIKeyStore retrieves API key information.
// This is the pluggable interface - implementations live in toolpersist.
type APIKeyStore interface {
    Lookup(ctx context.Context, keyHash string) (*APIKeyInfo, error)
}

// APIKeyAuthenticator validates API keys.
type APIKeyAuthenticator struct {
    Config APIKeyConfig
    Store  APIKeyStore
}

// NewAPIKeyAuthenticator creates an API key authenticator.
func NewAPIKeyAuthenticator(cfg APIKeyConfig, store APIKeyStore) *APIKeyAuthenticator {
    if cfg.HeaderName == "" {
        cfg.HeaderName = "X-API-Key"
    }
    return &APIKeyAuthenticator{
        Config: cfg,
        Store:  store,
    }
}

// Authenticate implements Authenticator.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
    // Extract API key
    key := a.extractKey(req)
    if key == "" {
        return &AuthResult{
            Authenticated: false,
            Error:         ErrMissingCredentials,
        }, nil
    }

    // Hash the key for lookup
    keyHash := a.hashKey(key)

    // Lookup key info
    info, err := a.Store.Lookup(ctx, keyHash)
    if err != nil {
        return nil, err // System error
    }
    if info == nil {
        return &AuthResult{
            Authenticated: false,
            Error:         ErrInvalidCredentials,
        }, nil
    }

    // Check expiration
    if info.IsExpired() {
        return &AuthResult{
            Authenticated: false,
            Error:         ErrTokenExpired,
        }, nil
    }

    // Build identity
    identity := &Identity{
        Principal:   info.Principal,
        TenantID:    info.TenantID,
        Roles:       info.Roles,
        Permissions: info.Permissions,
        Method:      AuthMethodAPIKey,
        Metadata:    info.Metadata,
        IssuedAt:    info.CreatedAt,
    }
    if info.ExpiresAt != nil {
        identity.ExpiresAt = *info.ExpiresAt
    }

    return &AuthResult{
        Authenticated: true,
        Identity:      identity,
    }, nil
}

func (a *APIKeyAuthenticator) extractKey(req *AuthRequest) string {
    // Try header first
    key := req.GetHeader(a.Config.HeaderName)
    if key != "" {
        if a.Config.Prefix != "" {
            key = strings.TrimPrefix(key, a.Config.Prefix)
        }
        return key
    }
    return ""
}

func (a *APIKeyAuthenticator) hashKey(key string) string {
    switch a.Config.HashAlgorithm {
    case "sha256":
        hash := sha256.Sum256([]byte(key))
        return hex.EncodeToString(hash[:])
    default:
        return key
    }
}

// Name implements Authenticator.
func (a *APIKeyAuthenticator) Name() string {
    return "api_key"
}

// Supports implements Authenticator.
func (a *APIKeyAuthenticator) Supports(ctx context.Context, req *AuthRequest) bool {
    return req.GetHeader(a.Config.HeaderName) != ""
}

// HashAPIKey hashes an API key for storage.
func HashAPIKey(key string, algorithm string) string {
    switch algorithm {
    case "sha256":
        hash := sha256.Sum256([]byte(key))
        return hex.EncodeToString(hash[:])
    default:
        return key
    }
}

// ValidateAPIKey compares a key against a stored hash.
func ValidateAPIKey(key, storedHash, algorithm string) bool {
    keyHash := HashAPIKey(key, algorithm)
    return subtle.ConstantTimeCompare([]byte(keyHash), []byte(storedHash)) == 1
}

// MemoryAPIKeyStore is an in-memory implementation for testing.
type MemoryAPIKeyStore struct {
    keys map[string]*APIKeyInfo
}

// NewMemoryAPIKeyStore creates an in-memory API key store.
func NewMemoryAPIKeyStore() *MemoryAPIKeyStore {
    return &MemoryAPIKeyStore{
        keys: make(map[string]*APIKeyInfo),
    }
}

// Add stores an API key.
func (s *MemoryAPIKeyStore) Add(info *APIKeyInfo) {
    s.keys[info.KeyHash] = info
}

// Lookup implements APIKeyStore.
func (s *MemoryAPIKeyStore) Lookup(ctx context.Context, keyHash string) (*APIKeyInfo, error) {
    return s.keys[keyHash], nil
}
```

**Tests:**

```go
// internal/auth/apikey_test.go
package auth

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAPIKeyAuthenticator_ValidKey(t *testing.T) {
    store := NewMemoryAPIKeyStore()
    store.Add(&APIKeyInfo{
        ID:        "key1",
        KeyHash:   "test-api-key",
        Principal: "alice",
        TenantID:  "acme-corp",
        Roles:     []string{"user"},
    })

    auth := NewAPIKeyAuthenticator(DefaultAPIKeyConfig(), store)

    req := &AuthRequest{
        Headers: map[string][]string{
            "X-API-Key": {"test-api-key"},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    require.NotNil(t, result.Identity)
    assert.Equal(t, "alice", result.Identity.Principal)
    assert.Equal(t, "acme-corp", result.Identity.TenantID)
    assert.Equal(t, AuthMethodAPIKey, result.Identity.Method)
}

func TestAPIKeyAuthenticator_HashedKey(t *testing.T) {
    store := NewMemoryAPIKeyStore()
    keyHash := HashAPIKey("my-secret-key", "sha256")
    store.Add(&APIKeyInfo{
        ID:        "key1",
        KeyHash:   keyHash,
        Principal: "bob",
    })

    auth := NewAPIKeyAuthenticator(
        APIKeyConfig{
            HeaderName:    "X-API-Key",
            HashAlgorithm: "sha256",
        },
        store,
    )

    req := &AuthRequest{
        Headers: map[string][]string{
            "X-API-Key": {"my-secret-key"},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.True(t, result.Authenticated)
    assert.Equal(t, "bob", result.Identity.Principal)
}

func TestAPIKeyAuthenticator_InvalidKey(t *testing.T) {
    store := NewMemoryAPIKeyStore()
    auth := NewAPIKeyAuthenticator(DefaultAPIKeyConfig(), store)

    req := &AuthRequest{
        Headers: map[string][]string{
            "X-API-Key": {"invalid-key"},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrInvalidCredentials)
}

func TestAPIKeyAuthenticator_ExpiredKey(t *testing.T) {
    expiredTime := time.Now().Add(-time.Hour)
    store := NewMemoryAPIKeyStore()
    store.Add(&APIKeyInfo{
        ID:        "key1",
        KeyHash:   "expired-key",
        Principal: "alice",
        ExpiresAt: &expiredTime,
    })

    auth := NewAPIKeyAuthenticator(DefaultAPIKeyConfig(), store)

    req := &AuthRequest{
        Headers: map[string][]string{
            "X-API-Key": {"expired-key"},
        },
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrTokenExpired)
}

func TestAPIKeyAuthenticator_MissingKey(t *testing.T) {
    store := NewMemoryAPIKeyStore()
    auth := NewAPIKeyAuthenticator(DefaultAPIKeyConfig(), store)

    req := &AuthRequest{
        Headers: map[string][]string{},
    }

    result, err := auth.Authenticate(context.Background(), req)

    require.NoError(t, err)
    assert.False(t, result.Authenticated)
    assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestAPIKeyAuthenticator_Supports(t *testing.T) {
    auth := NewAPIKeyAuthenticator(DefaultAPIKeyConfig(), NewMemoryAPIKeyStore())

    t.Run("supports X-API-Key header", func(t *testing.T) {
        req := &AuthRequest{
            Headers: map[string][]string{
                "X-API-Key": {"key123"},
            },
        }
        assert.True(t, auth.Supports(context.Background(), req))
    })

    t.Run("does not support bearer token", func(t *testing.T) {
        req := &AuthRequest{
            Headers: map[string][]string{
                "Authorization": {"Bearer token"},
            },
        }
        assert.False(t, auth.Supports(context.Background(), req))
    })
}

func TestHashAPIKey(t *testing.T) {
    key := "my-secret-key"

    t.Run("sha256", func(t *testing.T) {
        hash := HashAPIKey(key, "sha256")
        assert.Len(t, hash, 64) // SHA256 hex = 64 chars
        assert.NotEqual(t, key, hash)
    })

    t.Run("none", func(t *testing.T) {
        hash := HashAPIKey(key, "none")
        assert.Equal(t, key, hash)
    })
}

func TestValidateAPIKey(t *testing.T) {
    key := "my-secret-key"
    hash := HashAPIKey(key, "sha256")

    assert.True(t, ValidateAPIKey(key, hash, "sha256"))
    assert.False(t, ValidateAPIKey("wrong-key", hash, "sha256"))
}
```

---

### Task 6: Authorizer Interface & Simple RBAC

**Files:**
- `internal/auth/authorizer.go`
- `internal/auth/rbac.go`
- `internal/auth/authorizer_test.go`
- `internal/auth/rbac_test.go`

```go
// internal/auth/authorizer.go
package auth

import (
    "context"
    "errors"
    "fmt"
)

// Common authorization errors.
var (
    ErrForbidden        = errors.New("forbidden")
    ErrInsufficientRole = errors.New("insufficient role")
)

// AuthzRequest contains the authorization request parameters.
type AuthzRequest struct {
    Subject      *Identity
    Resource     string
    Action       string
    ResourceType string
    Context      map[string]any
}

// AuthzError provides detailed authorization failure information.
type AuthzError struct {
    Subject  string
    Resource string
    Action   string
    Reason   string
}

func (e *AuthzError) Error() string {
    return fmt.Sprintf("authorization denied: %s cannot %s on %s: %s",
        e.Subject, e.Action, e.Resource, e.Reason)
}

func (e *AuthzError) Unwrap() error {
    return ErrForbidden
}

// Authorizer makes access control decisions.
type Authorizer interface {
    Authorize(ctx context.Context, req *AuthzRequest) error
    Name() string
}

// AuthorizerFunc is a function adapter for Authorizer.
type AuthorizerFunc func(ctx context.Context, req *AuthzRequest) error

// Authorize implements Authorizer.
func (f AuthorizerFunc) Authorize(ctx context.Context, req *AuthzRequest) error {
    return f(ctx, req)
}

// Name implements Authorizer.
func (f AuthorizerFunc) Name() string { return "func" }

// AllowAllAuthorizer permits all requests.
var AllowAllAuthorizer = AuthorizerFunc(func(ctx context.Context, req *AuthzRequest) error {
    return nil
})

// DenyAllAuthorizer denies all requests.
var DenyAllAuthorizer = AuthorizerFunc(func(ctx context.Context, req *AuthzRequest) error {
    return &AuthzError{
        Subject:  req.Subject.Principal,
        Resource: req.Resource,
        Action:   req.Action,
        Reason:   "all access denied",
    }
})
```

```go
// internal/auth/rbac.go
package auth

import (
    "context"
    "slices"
    "strings"
)

// RBACConfig defines roles and permissions.
type RBACConfig struct {
    Roles       map[string]RoleConfig `yaml:"roles"`
    DefaultRole string                `yaml:"default_role"`
}

// RoleConfig defines a role's permissions.
type RoleConfig struct {
    Permissions    []string `yaml:"permissions"`
    Inherits       []string `yaml:"inherits"`
    AllowedTools   []string `yaml:"allowed_tools"`
    DeniedTools    []string `yaml:"denied_tools"`
    AllowedActions []string `yaml:"allowed_actions"`
}

// SimpleRBACAuthorizer provides basic RBAC without external dependencies.
type SimpleRBACAuthorizer struct {
    Config RBACConfig
}

// NewSimpleRBACAuthorizer creates a simple RBAC authorizer.
func NewSimpleRBACAuthorizer(cfg RBACConfig) *SimpleRBACAuthorizer {
    return &SimpleRBACAuthorizer{Config: cfg}
}

// Authorize implements Authorizer.
func (a *SimpleRBACAuthorizer) Authorize(ctx context.Context, req *AuthzRequest) error {
    if req.Subject == nil {
        return &AuthzError{
            Subject:  "anonymous",
            Resource: req.Resource,
            Action:   req.Action,
            Reason:   "no identity",
        }
    }

    // Get effective permissions
    permissions := a.resolvePermissions(req.Subject)
    allowedTools := a.resolveAllowedTools(req.Subject)
    deniedTools := a.resolveDeniedTools(req.Subject)
    allowedActions := a.resolveAllowedActions(req.Subject)

    // Check for wildcard permission
    if slices.Contains(permissions, "*") {
        return nil
    }

    // Check denied tools first (takes precedence)
    for _, pattern := range deniedTools {
        if matchPattern(req.Resource, pattern) {
            return &AuthzError{
                Subject:  req.Subject.Principal,
                Resource: req.Resource,
                Action:   req.Action,
                Reason:   "tool explicitly denied",
            }
        }
    }

    // Check action allowed
    if len(allowedActions) > 0 && !slices.Contains(allowedActions, req.Action) && !slices.Contains(allowedActions, "*") {
        return &AuthzError{
            Subject:  req.Subject.Principal,
            Resource: req.Resource,
            Action:   req.Action,
            Reason:   "action not allowed",
        }
    }

    // Check tool allowed
    if len(allowedTools) > 0 {
        allowed := false
        for _, pattern := range allowedTools {
            if matchPattern(req.Resource, pattern) {
                allowed = true
                break
            }
        }
        if !allowed && !slices.Contains(allowedTools, "*") {
            return &AuthzError{
                Subject:  req.Subject.Principal,
                Resource: req.Resource,
                Action:   req.Action,
                Reason:   "tool not allowed",
            }
        }
    }

    // Check explicit permission
    permKey := req.ResourceType + ":" + req.Action
    toolKey := "tool:" + req.Resource

    if slices.Contains(permissions, permKey) || slices.Contains(permissions, toolKey) {
        return nil
    }

    // If we have allowed tools/actions and passed those checks, allow
    if len(allowedTools) > 0 || len(allowedActions) > 0 {
        return nil
    }

    return &AuthzError{
        Subject:  req.Subject.Principal,
        Resource: req.Resource,
        Action:   req.Action,
        Reason:   "insufficient permissions",
    }
}

func (a *SimpleRBACAuthorizer) resolvePermissions(id *Identity) []string {
    var perms []string
    seen := make(map[string]bool)

    var resolve func(roles []string)
    resolve = func(roles []string) {
        for _, role := range roles {
            if seen[role] {
                continue
            }
            seen[role] = true

            if cfg, ok := a.Config.Roles[role]; ok {
                perms = append(perms, cfg.Permissions...)
                resolve(cfg.Inherits)
            }
        }
    }

    // Start with identity's roles
    resolve(id.Roles)

    // Add default role if no roles
    if len(id.Roles) == 0 && a.Config.DefaultRole != "" {
        resolve([]string{a.Config.DefaultRole})
    }

    // Add identity's direct permissions
    perms = append(perms, id.Permissions...)

    return perms
}

func (a *SimpleRBACAuthorizer) resolveAllowedTools(id *Identity) []string {
    var tools []string
    seen := make(map[string]bool)

    var resolve func(roles []string)
    resolve = func(roles []string) {
        for _, role := range roles {
            if seen[role] {
                continue
            }
            seen[role] = true

            if cfg, ok := a.Config.Roles[role]; ok {
                tools = append(tools, cfg.AllowedTools...)
                resolve(cfg.Inherits)
            }
        }
    }

    resolve(id.Roles)
    if len(id.Roles) == 0 && a.Config.DefaultRole != "" {
        resolve([]string{a.Config.DefaultRole})
    }

    return tools
}

func (a *SimpleRBACAuthorizer) resolveDeniedTools(id *Identity) []string {
    var tools []string
    seen := make(map[string]bool)

    var resolve func(roles []string)
    resolve = func(roles []string) {
        for _, role := range roles {
            if seen[role] {
                continue
            }
            seen[role] = true

            if cfg, ok := a.Config.Roles[role]; ok {
                tools = append(tools, cfg.DeniedTools...)
                resolve(cfg.Inherits)
            }
        }
    }

    resolve(id.Roles)
    return tools
}

func (a *SimpleRBACAuthorizer) resolveAllowedActions(id *Identity) []string {
    var actions []string
    seen := make(map[string]bool)

    var resolve func(roles []string)
    resolve = func(roles []string) {
        for _, role := range roles {
            if seen[role] {
                continue
            }
            seen[role] = true

            if cfg, ok := a.Config.Roles[role]; ok {
                actions = append(actions, cfg.AllowedActions...)
                resolve(cfg.Inherits)
            }
        }
    }

    resolve(id.Roles)
    if len(id.Roles) == 0 && a.Config.DefaultRole != "" {
        resolve([]string{a.Config.DefaultRole})
    }

    return actions
}

// Name implements Authorizer.
func (a *SimpleRBACAuthorizer) Name() string {
    return "simple_rbac"
}

// matchPattern matches a resource against a glob-like pattern.
// Supports * as wildcard prefix/suffix (e.g., "search_*", "*_tool", "*")
func matchPattern(resource, pattern string) bool {
    if pattern == "*" {
        return true
    }
    if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
        return strings.Contains(resource, pattern[1:len(pattern)-1])
    }
    if strings.HasPrefix(pattern, "*") {
        return strings.HasSuffix(resource, pattern[1:])
    }
    if strings.HasSuffix(pattern, "*") {
        return strings.HasPrefix(resource, pattern[:len(pattern)-1])
    }
    return resource == pattern
}
```

**Tests:**

```go
// internal/auth/rbac_test.go
package auth

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSimpleRBACAuthorizer_AdminRole(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{
        Roles: map[string]RoleConfig{
            "admin": {Permissions: []string{"*"}},
        },
    })

    req := &AuthzRequest{
        Subject:      &Identity{Principal: "alice", Roles: []string{"admin"}},
        Resource:     "execute_code",
        Action:       "call",
        ResourceType: "tool",
    }

    err := authz.Authorize(context.Background(), req)
    assert.NoError(t, err)
}

func TestSimpleRBACAuthorizer_UserRole(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{
        Roles: map[string]RoleConfig{
            "user": {
                AllowedActions: []string{"call", "list"},
                AllowedTools:   []string{"search_*", "describe_*"},
            },
        },
    })

    t.Run("allowed tool", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "bob", Roles: []string{"user"}},
            Resource:     "search_tools",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        assert.NoError(t, err)
    })

    t.Run("denied tool", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "bob", Roles: []string{"user"}},
            Resource:     "execute_code",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        require.Error(t, err)
        assert.ErrorIs(t, err, ErrForbidden)
    })

    t.Run("denied action", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "bob", Roles: []string{"user"}},
            Resource:     "search_tools",
            Action:       "delete",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        require.Error(t, err)
    })
}

func TestSimpleRBACAuthorizer_DeniedToolsTakePrecedence(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{
        Roles: map[string]RoleConfig{
            "user": {
                AllowedTools: []string{"*"},
                DeniedTools:  []string{"execute_code", "dangerous_*"},
            },
        },
    })

    t.Run("explicitly denied", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "bob", Roles: []string{"user"}},
            Resource:     "execute_code",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        require.Error(t, err)
        var authzErr *AuthzError
        require.ErrorAs(t, err, &authzErr)
        assert.Contains(t, authzErr.Reason, "denied")
    })

    t.Run("pattern denied", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "bob", Roles: []string{"user"}},
            Resource:     "dangerous_operation",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        require.Error(t, err)
    })
}

func TestSimpleRBACAuthorizer_RoleInheritance(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{
        Roles: map[string]RoleConfig{
            "admin": {
                Permissions: []string{"*"},
            },
            "power_user": {
                Inherits:     []string{"user"},
                AllowedTools: []string{"execute_code"},
            },
            "user": {
                AllowedActions: []string{"call", "list"},
                AllowedTools:   []string{"search_*"},
            },
        },
    })

    t.Run("inherits from user", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "carol", Roles: []string{"power_user"}},
            Resource:     "search_tools",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        assert.NoError(t, err)
    })

    t.Run("has own permissions", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "carol", Roles: []string{"power_user"}},
            Resource:     "execute_code",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        assert.NoError(t, err)
    })
}

func TestSimpleRBACAuthorizer_DefaultRole(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{
        DefaultRole: "anonymous",
        Roles: map[string]RoleConfig{
            "anonymous": {
                AllowedActions: []string{"list"},
            },
        },
    })

    t.Run("uses default role", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "unknown"}, // No roles
            Resource:     "any_tool",
            Action:       "list",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        assert.NoError(t, err)
    })

    t.Run("denied without role", func(t *testing.T) {
        req := &AuthzRequest{
            Subject:      &Identity{Principal: "unknown"},
            Resource:     "any_tool",
            Action:       "call",
            ResourceType: "tool",
        }
        err := authz.Authorize(context.Background(), req)
        require.Error(t, err)
    })
}

func TestSimpleRBACAuthorizer_NoIdentity(t *testing.T) {
    authz := NewSimpleRBACAuthorizer(RBACConfig{})

    req := &AuthzRequest{
        Subject:  nil,
        Resource: "any_tool",
        Action:   "call",
    }

    err := authz.Authorize(context.Background(), req)
    require.Error(t, err)
}

func TestMatchPattern(t *testing.T) {
    tests := []struct {
        resource string
        pattern  string
        expected bool
    }{
        {"search_tools", "search_*", true},
        {"search_tools", "*_tools", true},
        {"search_tools", "*", true},
        {"search_tools", "search_tools", true},
        {"search_tools", "other_*", false},
        {"search_tools", "*_other", false},
        {"search_tools", "exact_match", false},
        {"long_search_tools", "*search*", true},
    }

    for _, tc := range tests {
        t.Run(tc.resource+"/"+tc.pattern, func(t *testing.T) {
            assert.Equal(t, tc.expected, matchPattern(tc.resource, tc.pattern))
        })
    }
}
```

---

### Task 7: Auth Middleware Integration

**Files:**
- `internal/auth/middleware.go`
- `internal/auth/middleware_test.go`

```go
// internal/auth/middleware.go
package auth

import (
    "context"
    "fmt"

    "github.com/jonwraymond/metatools-mcp/internal/mcp"
    "github.com/jonwraymond/metatools-mcp/internal/middleware"
    "github.com/jonwraymond/metatools-mcp/internal/provider"
)

// AuthMiddlewareConfig configures the authentication middleware.
type AuthMiddlewareConfig struct {
    AllowAnonymous    bool
    AnonymousIdentity *Identity
    SkipMethods       []string
    OnAuthFailure     func(ctx context.Context, err error) error
}

// DefaultAuthMiddlewareConfig returns sensible defaults.
func DefaultAuthMiddlewareConfig() AuthMiddlewareConfig {
    return AuthMiddlewareConfig{
        AllowAnonymous: false,
        AnonymousIdentity: &Identity{
            Principal: "anonymous",
            Method:    AuthMethodAnonymous,
        },
    }
}

// AuthMiddleware creates authentication middleware.
func AuthMiddleware(auth Authenticator, cfg AuthMiddlewareConfig) middleware.Middleware {
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
    config AuthMiddlewareConfig
}

func (m *authMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    // Build auth request from context
    req := buildAuthRequestFromContext(ctx)

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

func (m *authMiddleware) Name() string        { return m.next.Name() }
func (m *authMiddleware) Description() string { return m.next.Description() }
func (m *authMiddleware) InputSchema() map[string]any { return m.next.InputSchema() }

// AuthzMiddlewareConfig configures the authorization middleware.
type AuthzMiddlewareConfig struct {
    ResourceResolver func(ctx context.Context, input map[string]any) string
    ActionResolver   func(ctx context.Context, input map[string]any) string
    ContextBuilder   func(ctx context.Context, input map[string]any) map[string]any
    OnDenied         func(ctx context.Context, err error) error
}

// DefaultAuthzMiddlewareConfig returns sensible defaults.
func DefaultAuthzMiddlewareConfig() AuthzMiddlewareConfig {
    return AuthzMiddlewareConfig{
        ResourceResolver: func(ctx context.Context, input map[string]any) string {
            if name, ok := input["name"].(string); ok {
                return name
            }
            return ""
        },
        ActionResolver: func(ctx context.Context, input map[string]any) string {
            return "call"
        },
        ContextBuilder: func(ctx context.Context, input map[string]any) map[string]any {
            return nil
        },
    }
}

// AuthzMiddleware creates authorization middleware.
func AuthzMiddleware(authz Authorizer, cfg AuthzMiddlewareConfig) middleware.Middleware {
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
    config AuthzMiddlewareConfig
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

func (m *authzMiddleware) Name() string        { return m.next.Name() }
func (m *authzMiddleware) Description() string { return m.next.Description() }
func (m *authzMiddleware) InputSchema() map[string]any { return m.next.InputSchema() }

// AuthError represents an authentication/authorization error.
type AuthError struct {
    Code      string
    Message   string
    Challenge string
}

func (e *AuthError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// buildAuthRequestFromContext creates an AuthRequest from context.
// In real usage, this would extract headers from the transport layer.
func buildAuthRequestFromContext(ctx context.Context) *AuthRequest {
    req := &AuthRequest{
        Headers: make(map[string][]string),
    }

    // Extract headers from context if available
    if headers, ok := ctx.Value(headersKey{}).(map[string][]string); ok {
        req.Headers = headers
    }

    return req
}

// Context key for headers.
type headersKey struct{}

// WithHeaders adds request headers to context.
func WithHeaders(ctx context.Context, headers map[string][]string) context.Context {
    return context.WithValue(ctx, headersKey{}, headers)
}

// HeadersFromContext retrieves headers from context.
func HeadersFromContext(ctx context.Context) map[string][]string {
    if headers, ok := ctx.Value(headersKey{}).(map[string][]string); ok {
        return headers
    }
    return nil
}
```

---

### Task 8: Factory Registration

**Files:**
- `internal/auth/factory.go`
- `internal/auth/factory_test.go`

```go
// internal/auth/factory.go
package auth

import (
    "fmt"
    "sync"

    "github.com/jonwraymond/metatools-mcp/internal/middleware"
)

// AuthenticatorFactory creates an Authenticator from configuration.
type AuthenticatorFactory func(cfg map[string]any) (Authenticator, error)

// AuthorizerFactory creates an Authorizer from configuration.
type AuthorizerFactory func(cfg map[string]any) (Authorizer, error)

// Registry holds auth component factories.
type Registry struct {
    mu             sync.RWMutex
    authenticators map[string]AuthenticatorFactory
    authorizers    map[string]AuthorizerFactory
}

// NewRegistry creates a new auth registry.
func NewRegistry() *Registry {
    return &Registry{
        authenticators: make(map[string]AuthenticatorFactory),
        authorizers:    make(map[string]AuthorizerFactory),
    }
}

// RegisterAuthenticator registers an authenticator factory.
func (r *Registry) RegisterAuthenticator(name string, factory AuthenticatorFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.authenticators[name] = factory
}

// RegisterAuthorizer registers an authorizer factory.
func (r *Registry) RegisterAuthorizer(name string, factory AuthorizerFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.authorizers[name] = factory
}

// GetAuthenticator retrieves an authenticator factory.
func (r *Registry) GetAuthenticator(name string) (AuthenticatorFactory, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    f, ok := r.authenticators[name]
    return f, ok
}

// GetAuthorizer retrieves an authorizer factory.
func (r *Registry) GetAuthorizer(name string) (AuthorizerFactory, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    f, ok := r.authorizers[name]
    return f, ok
}

// DefaultRegistry is the global auth registry.
var DefaultRegistry = NewRegistry()

func init() {
    // Register built-in authenticators
    DefaultRegistry.RegisterAuthenticator("jwt", func(cfg map[string]any) (Authenticator, error) {
        jwtCfg := DefaultJWTConfig()
        if err := mapToStruct(cfg, &jwtCfg); err != nil {
            return nil, fmt.Errorf("jwt config: %w", err)
        }

        // Get key provider based on config
        keyProvider, err := buildKeyProvider(cfg)
        if err != nil {
            return nil, fmt.Errorf("jwt key provider: %w", err)
        }

        return NewJWTAuthenticator(jwtCfg, keyProvider), nil
    })

    DefaultRegistry.RegisterAuthenticator("api_key", func(cfg map[string]any) (Authenticator, error) {
        apiKeyCfg := DefaultAPIKeyConfig()
        if err := mapToStruct(cfg, &apiKeyCfg); err != nil {
            return nil, fmt.Errorf("api_key config: %w", err)
        }

        // Get key store based on config
        store, err := buildAPIKeyStore(cfg)
        if err != nil {
            return nil, fmt.Errorf("api_key store: %w", err)
        }

        return NewAPIKeyAuthenticator(apiKeyCfg, store), nil
    })

    // Register built-in authorizers
    DefaultRegistry.RegisterAuthorizer("simple_rbac", func(cfg map[string]any) (Authorizer, error) {
        rbacCfg := RBACConfig{}
        if err := mapToStruct(cfg, &rbacCfg); err != nil {
            return nil, fmt.Errorf("rbac config: %w", err)
        }
        return NewSimpleRBACAuthorizer(rbacCfg), nil
    })

    DefaultRegistry.RegisterAuthorizer("allow_all", func(cfg map[string]any) (Authorizer, error) {
        return AllowAllAuthorizer, nil
    })

    DefaultRegistry.RegisterAuthorizer("deny_all", func(cfg map[string]any) (Authorizer, error) {
        return DenyAllAuthorizer, nil
    })

    // Register middleware factories with global middleware registry
    middleware.DefaultRegistry.Register("auth", func(cfg map[string]any) (middleware.Middleware, error) {
        authType, _ := cfg["type"].(string)
        if authType == "" {
            authType = "jwt"
        }

        factory, ok := DefaultRegistry.GetAuthenticator(authType)
        if !ok {
            return nil, fmt.Errorf("unknown authenticator type: %s", authType)
        }

        auth, err := factory(cfg)
        if err != nil {
            return nil, err
        }

        mwCfg := DefaultAuthMiddlewareConfig()
        if allowAnon, ok := cfg["allow_anonymous"].(bool); ok {
            mwCfg.AllowAnonymous = allowAnon
        }

        return AuthMiddleware(auth, mwCfg), nil
    })

    middleware.DefaultRegistry.Register("authz", func(cfg map[string]any) (middleware.Middleware, error) {
        authzType, _ := cfg["type"].(string)
        if authzType == "" {
            authzType = "simple_rbac"
        }

        factory, ok := DefaultRegistry.GetAuthorizer(authzType)
        if !ok {
            return nil, fmt.Errorf("unknown authorizer type: %s", authzType)
        }

        authz, err := factory(cfg)
        if err != nil {
            return nil, err
        }

        return AuthzMiddleware(authz, DefaultAuthzMiddlewareConfig()), nil
    })
}

// buildKeyProvider creates a KeyProvider from configuration.
func buildKeyProvider(cfg map[string]any) (KeyProvider, error) {
    // Check for static secret
    if secret, ok := cfg["secret"].(string); ok {
        return NewHMACKeyProvider([]byte(secret)), nil
    }

    // Check for JWKS URL (future: implement JWKSKeyProvider)
    if jwksURL, ok := cfg["jwks_url"].(string); ok {
        _ = jwksURL // TODO: Implement JWKSKeyProvider
        return nil, fmt.Errorf("JWKS not yet implemented")
    }

    return nil, fmt.Errorf("no key provider configured")
}

// buildAPIKeyStore creates an APIKeyStore from configuration.
func buildAPIKeyStore(cfg map[string]any) (APIKeyStore, error) {
    storeCfg, _ := cfg["store"].(map[string]any)
    storeType, _ := storeCfg["type"].(string)

    switch storeType {
    case "memory", "":
        return NewMemoryAPIKeyStore(), nil
    // Future: case "redis", "postgres"
    default:
        return nil, fmt.Errorf("unknown api_key store type: %s", storeType)
    }
}

// mapToStruct converts a map to a struct using reflection.
// This is a simplified version - real implementation would use mapstructure.
func mapToStruct(m map[string]any, target any) error {
    // TODO: Use github.com/mitchellh/mapstructure
    return nil
}
```

---

## Verification

```bash
# Run all auth tests
go test ./internal/auth/... -v

# Run with coverage
go test ./internal/auth/... -cover

# Run specific test
go test ./internal/auth/... -run TestJWTAuthenticator -v

# Lint
golangci-lint run ./internal/auth/...
```

---

## Definition of Done

- [ ] All tests pass with >80% coverage
- [ ] No golangci-lint errors
- [ ] Interfaces documented with godoc comments
- [ ] Factory registration working
- [ ] Middleware chain integration verified
- [ ] YAML configuration parsing works
- [ ] Context propagation verified

---

## Dependencies

```bash
go get github.com/golang-jwt/jwt/v5
```

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-30 | Initial auth middleware PRD |
