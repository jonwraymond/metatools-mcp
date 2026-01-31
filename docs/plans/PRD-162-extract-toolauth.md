# PRD-162: Extract toolauth

**Phase:** 6 - Operations Layer
**Priority:** High
**Effort:** 8 hours
**Dependencies:** PRD-120

---

## Objective

Extract authentication code from `metatools-mcp/internal/auth/` into `toolops/auth/` for reusable authentication middleware.

---

## Source Analysis

**Current Location:** `metatools-mcp/internal/auth/` (embedded)
**Target Location:** `github.com/ApertureStack/toolops/auth`

**Code to Extract:**
- JWT authentication
- API key authentication
- RBAC authorization
- Auth middleware
- ~6,400 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Auth Package | `toolops/auth/` | Authentication/authorization |
| Tests | `toolops/auth/*_test.go` | Comprehensive tests |
| Documentation | `toolops/auth/doc.go` | Package documentation |

---

## Tasks

### Task 1: Analyze and Copy Source

```bash
cd /Users/jraymond/Documents/Projects/ApertureStack/metatools-mcp

wc -l internal/auth/*.go

cd /tmp/migration/toolops
mkdir -p auth
cp /Users/jraymond/Documents/Projects/ApertureStack/metatools-mcp/internal/auth/*.go auth/

# Update package name if needed
sed -i '' 's|package auth|package auth|g' auth/*.go
```

### Task 2: Update Imports

```bash
cd /tmp/migration/toolops/auth

# Update internal references
sed -i '' 's|metatools-mcp/internal/auth|github.com/ApertureStack/toolops/auth|g' *.go
sed -i '' 's|github.com/ApertureStack/toolmodel|github.com/ApertureStack/toolfoundation/model|g' *.go
```

### Task 3: Define Core Interfaces

**File:** `toolops/auth/auth.go`

```go
package auth

import (
    "context"
)

// Authenticator validates credentials and returns an identity.
type Authenticator interface {
    // Authenticate validates credentials in the request.
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)

    // Name returns the authenticator name (e.g., "jwt", "apikey").
    Name() string

    // Supports returns true if this authenticator can handle the request.
    Supports(ctx context.Context, req *AuthRequest) bool
}

// Authorizer checks if an identity can perform an action.
type Authorizer interface {
    // Authorize checks if the identity can perform the action.
    Authorize(ctx context.Context, identity *Identity, action *Action) (*AuthzResult, error)
}

// AuthRequest contains authentication request data.
type AuthRequest struct {
    Headers map[string]string
    Token   string
    APIKey  string
    Method  string
    Path    string
}

// AuthResult contains authentication result.
type AuthResult struct {
    Authenticated bool
    Identity      *Identity
    Error         error
}

// Identity represents an authenticated entity.
type Identity struct {
    Subject   string
    TenantID  string
    Roles     []string
    Claims    map[string]any
    ExpiresAt time.Time
}

// Action represents an authorization action.
type Action struct {
    Resource   string // e.g., "tool:calculator"
    Operation  string // e.g., "execute"
    Context    map[string]any
}

// AuthzResult contains authorization result.
type AuthzResult struct {
    Allowed bool
    Reason  string
}
```

### Task 4: Create Package Documentation

**File:** `toolops/auth/doc.go`

```go
// Package auth provides authentication and authorization for tool access.
//
// This package implements pluggable authentication strategies and role-based
// access control for tool execution.
//
// # Authenticators
//
// Built-in authenticators:
//
//   - JWTAuthenticator: Validates JWT tokens
//   - APIKeyAuthenticator: Validates API keys
//   - CompositeAuthenticator: Chains multiple authenticators
//
// # Usage
//
// Create and use authenticators:
//
//	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
//	    Secret: []byte("secret"),
//	    Issuer: "metatools",
//	})
//
//	apiKeyAuth := auth.NewAPIKeyAuthenticator(auth.APIKeyConfig{
//	    Store: keyStore,
//	})
//
//	composite := auth.NewCompositeAuthenticator(jwtAuth, apiKeyAuth)
//
// # Authorization
//
// Role-based access control:
//
//	rbac := auth.NewRBACAuthorizer(auth.RBACConfig{
//	    Roles: map[string][]string{
//	        "admin": {"*"},
//	        "user":  {"tool:*:execute"},
//	    },
//	})
//
//	result, _ := rbac.Authorize(ctx, identity, action)
//	if !result.Allowed {
//	    // Access denied
//	}
//
// # Middleware
//
// Protect tool providers with auth middleware:
//
//	protected := auth.Middleware(provider, auth.MiddlewareConfig{
//	    Authenticator: composite,
//	    Authorizer:    rbac,
//	})
//
// # Extraction Note
//
// This package was extracted from metatools-mcp/internal/auth as part of
// the ApertureStack consolidation to enable reuse across projects.
package auth
```

### Task 5: Build and Test

```bash
cd /tmp/migration/toolops

go mod tidy
go build ./...
go test -v ./auth/...
```

### Task 6: Commit and Push

```bash
cd /tmp/migration/toolops

git add -A
git commit -m "feat(auth): extract authentication package

Extract auth infrastructure from metatools-mcp for reuse.

Package contents:
- Authenticator interface
- JWTAuthenticator for JWT tokens
- APIKeyAuthenticator for API keys
- CompositeAuthenticator for chaining
- RBACAuthorizer for role-based access
- Auth middleware

Features:
- Pluggable authentication strategies
- Role-based access control
- Identity management
- Multi-tenant support
- Factory-based configuration

This extraction enables auth reuse across projects.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-163: Create toolresilience
- PRD-164: Create toolhealth
