# Extraction Analysis: metatools-mcp → Standalone Libraries

**Date:** 2026-01-30
**Purpose:** Identify components in metatools-mcp that can be extracted into standalone tool* libraries.

---

## Executive Summary

After thorough analysis of metatools-mcp's internal packages, **3 major components** can be extracted into standalone libraries:

| Library | Lines | MCP Dependencies | Extraction Feasibility |
|---------|-------|------------------|----------------------|
| **toolauth** | ~2,500 | None | ✅ Fully extractable |
| **toolbackend** | ~300 | None (uses toolmodel) | ✅ Fully extractable |
| **tooltransport** | ~500 | Partial | ⚠️ Interface extractable, impls stay |

This would reduce metatools-mcp to a thin **composition layer** that wires together the libraries.

---

## Detailed Analysis

### 1. toolauth (✅ EXTRACT)

**Current Location:** `internal/auth/` (~2,500 lines)

**Why Extract:**
- Zero MCP dependencies - pure authentication/authorization
- Reusable by any tool server (not just MCP)
- Production-ready with 82.7% test coverage
- Comprehensive feature set already implemented

**Components:**

| File | Lines | Purpose | Dependencies |
|------|-------|---------|--------------|
| `identity.go` | 114 | Identity type, AuthMethod enum | None |
| `authenticator.go` | 99 | Authenticator interface, AuthRequest, AuthResult | None |
| `authorizer.go` | 123 | Authorizer interface, AuthzRequest, RBAC types | None |
| `jwt.go` | 253 | JWT authenticator, claims extraction | golang-jwt |
| `jwks.go` | 226 | JWKS key provider, caching, singleflight | x/sync |
| `apikey.go` | 218 | API key authenticator, store interface | None |
| `apikey_redis.go` | 106 | Redis API key store | go-redis (optional) |
| `oauth2_introspection.go` | 347 | OAuth2 token introspection | None |
| `rbac.go` | 176 | Simple RBAC authorizer | None |
| `composite.go` | 83 | Multi-authenticator chain | None |
| `context.go` | 72 | Context helpers | None |
| `middleware.go` | 174 | Auth middleware for provider wrapping | **provider pkg** |
| `factory.go` | 322 | Config-based factory | All above |
| `errors.go` | 39 | Error definitions | None |

**Proposed Structure:**

```go
// toolauth - Authentication & Authorization for tool servers
package toolauth

// Core interfaces (zero deps)
type Identity struct { ... }
type Authenticator interface { ... }
type Authorizer interface { ... }

// Implementations
type JWTAuthenticator struct { ... }      // toolauth/jwt
type APIKeyAuthenticator struct { ... }   // toolauth/apikey
type OAuth2Authenticator struct { ... }   // toolauth/oauth2
type RBACAuthorizer struct { ... }        // toolauth/rbac

// Key providers
type JWKSKeyProvider struct { ... }       // toolauth/jwks
type StaticKeyProvider struct { ... }

// Stores
type MemoryAPIKeyStore struct { ... }
type RedisAPIKeyStore struct { ... }      // toolauth/redis (build tag)

// Composition
type CompositeAuthenticator struct { ... }
```

**Extraction Notes:**
- `middleware.go` depends on `provider.ToolProvider` - needs generic interface or stays in metatools-mcp
- Factory pattern can use functional options

---

### 2. toolbackend (✅ EXTRACT)

**Current Location:** `internal/backend/` (~300 lines)

**Why Extract:**
- Already uses `toolmodel.Tool` NOT `mcp.Tool` - protocol agnostic!
- Clean Backend interface for tool execution
- Registry pattern is reusable
- Enables alternative server implementations

**Components:**

| File | Lines | Purpose | Dependencies |
|------|-------|---------|--------------|
| `backend.go` | 79 | Backend interface, errors | toolmodel |
| `registry.go` | 142 | Backend registry | None |
| `aggregator.go` | ~100 | Multi-backend aggregation | None |
| `local.go` | ~80 | Local tool handlers | None |

**Proposed Structure:**

```go
// toolbackend - Tool execution backends
package toolbackend

// Core interface
type Backend interface {
    Kind() string
    Name() string
    Enabled() bool
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
    Start(ctx context.Context) error
    Stop() error
}

// Extensions
type StreamingBackend interface { ... }
type ConfigurableBackend interface { ... }

// Registry
type Registry struct { ... }

// Implementations
type LocalBackend struct { ... }      // toolbackend/local
type AggregatorBackend struct { ... } // toolbackend/aggregator
```

**Extraction Notes:**
- MCP backend implementation stays in metatools-mcp (uses MCP SDK)
- HTTP backend could be in toolbackend/http
- Aggregator enables multi-source tool catalogs

---

### 3. tooltransport (⚠️ PARTIAL EXTRACT)

**Current Location:** `internal/transport/` (~500 lines)

**What to Extract:**
- Transport interface (generic)
- Server interface (generic)
- Transport Info type

**What Stays in metatools-mcp:**
- StdioTransport (uses mcp.Transport)
- SSETransport (uses mcp.Server)
- StreamableHTTPTransport (uses mcp.Server)

**Proposed Structure:**

```go
// tooltransport - Transport abstraction layer
package tooltransport

// Core interfaces (protocol-agnostic)
type Transport interface {
    Name() string
    Info() Info
    Serve(ctx context.Context, handler Handler) error
    Close() error
}

type Handler interface {
    Handle(ctx context.Context, req []byte) ([]byte, error)
}

type Info struct {
    Name string
    Addr string
    Path string
}

// Generic implementations could include:
// - HTTP transport (JSON-RPC over HTTP)
// - WebSocket transport
// - gRPC transport (future)
```

**Extraction Notes:**
- Interface is valuable for non-MCP servers
- MCP-specific transports stay in metatools-mcp as adapters

---

### 4. toolmiddleware (⚠️ CONSIDER)

**Current Location:** `internal/middleware/` (~500 lines)

**Challenge:** Middleware wraps `provider.ToolProvider` which has MCP types.

**Options:**

**Option A: Generic Middleware (Extract)**
```go
// toolmiddleware - Generic middleware chain
package toolmiddleware

type Handler[Req, Resp any] interface {
    Handle(ctx context.Context, req Req) (Resp, error)
}

type Middleware[Req, Resp any] func(Handler[Req, Resp]) Handler[Req, Resp]

type Chain[Req, Resp any] struct { ... }
```

**Option B: Keep in metatools-mcp**
- Middleware is tightly coupled to ToolProvider
- Specific middlewares (logging, metrics, auth) use provider context

**Recommendation:** Keep in metatools-mcp for now. The generic chain pattern is simple enough that duplicating it isn't costly, and the specific middlewares are MCP-aware.

---

### 5. Components That Stay in metatools-mcp

| Package | Reason |
|---------|--------|
| `internal/handlers/` | MCP-specific metatool implementations |
| `internal/server/` | MCP server wiring, adapters |
| `internal/adapters/` | Adapters between tool* libs and MCP |
| `internal/bootstrap/` | Index initialization |
| `internal/config/` | metatools-mcp specific config |
| `internal/provider/` | MCP ToolProvider interface |
| `internal/runtime/` | Docker/WASM implementations of toolruntime interfaces |

---

## Updated Library Categorization

With extractions, the consolidated structure becomes:

```
ApertureStack/
├── toolfoundation/         # Foundation
│   ├── model/              # Canonical schema (toolmodel)
│   ├── adapter/            # Protocol adapters (tooladapter)
│   └── version/            # Versioning (toolversion)
│
├── tooldiscovery/          # Discovery
│   ├── index/              # Registry (toolindex)
│   ├── search/             # BM25 search (toolsearch)
│   ├── semantic/           # Hybrid search (toolsemantic)
│   └── docs/               # Documentation (tooldocs)
│
├── toolexec/               # Execution
│   ├── run/                # Execution pipeline (toolrun)
│   ├── runtime/            # Sandbox backends (toolruntime)
│   ├── code/               # Code orchestration (toolcode)
│   └── backend/            # Backend abstraction (NEW - from metatools-mcp)
│
├── toolcompose/            # Composition
│   ├── set/                # Toolsets (toolset)
│   └── skill/              # Agent skills (toolskill)
│
├── toolops/                # Operations
│   ├── observe/            # OpenTelemetry (toolobserve)
│   ├── cache/              # Response caching (toolcache)
│   ├── resilience/         # Circuit breaker (toolresilience)
│   ├── health/             # Health checks (toolhealth)
│   └── auth/               # Authentication (NEW - from metatools-mcp)
│
└── metatools-mcp/          # MCP Server (thin composition layer)
    ├── internal/transport/ # MCP transports (Stdio, SSE, HTTP)
    ├── internal/provider/  # MCP ToolProvider
    ├── internal/middleware/# MCP middleware
    ├── internal/handlers/  # Metatool handlers
    ├── internal/server/    # Server wiring
    └── cmd/metatools/      # CLI
```

---

## Extraction Priority

| Priority | Library | Effort | Value |
|----------|---------|--------|-------|
| **P1** | toolauth | 2-3 days | High - reusable auth for any server |
| **P2** | toolbackend | 1-2 days | Medium - clean execution abstraction |
| **P3** | tooltransport | 1 day | Low - interface only, impls stay |

---

## Benefits of Extraction

1. **Reusability**: toolauth can secure non-MCP servers (gRPC, REST, GraphQL)
2. **Testing**: Standalone libraries are easier to test in isolation
3. **Versioning**: Auth changes don't require metatools-mcp release
4. **Adoption**: Projects can use toolauth without full metatools stack
5. **Thin Server**: metatools-mcp becomes pure composition/wiring

---

## Migration Path

### Phase 1: Extract toolauth
1. Create `toolauth/` repo with core interfaces
2. Move identity, authenticator, authorizer, errors
3. Move implementations (jwt, apikey, oauth2, rbac, jwks)
4. Keep middleware.go in metatools-mcp (depends on ToolProvider)
5. Update metatools-mcp imports

### Phase 2: Extract toolbackend
1. Create `toolbackend/` in toolexec repo
2. Move Backend interface, Registry, Aggregator
3. Move LocalBackend
4. Keep MCP backend in metatools-mcp
5. Update metatools-mcp imports

### Phase 3: Extract tooltransport (optional)
1. Extract Transport/Handler interfaces
2. Keep MCP implementations in metatools-mcp

---

## Interface Contracts

### toolauth

```go
// Core contract - no external dependencies
type Authenticator interface {
    Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)
    Name() string
    Supports(ctx context.Context, req *AuthRequest) bool
}

type Authorizer interface {
    Authorize(ctx context.Context, req *AuthzRequest) error
    Name() string
}

type KeyProvider interface {
    GetKey(ctx context.Context, keyID string) (any, error)
}

type APIKeyStore interface {
    Lookup(ctx context.Context, keyHash string) (*APIKeyInfo, error)
}
```

### toolbackend

```go
// Core contract - depends only on toolmodel
type Backend interface {
    Kind() string
    Name() string
    Enabled() bool
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
    Start(ctx context.Context) error
    Stop() error
}
```

---

## Summary

| Extraction | Lines | MCP-Free | Recommendation |
|------------|-------|----------|----------------|
| **toolauth** | 2,500 | ✅ Yes | Extract to toolops/auth |
| **toolbackend** | 300 | ✅ Yes | Extract to toolexec/backend |
| **tooltransport** | 500 | ⚠️ Partial | Interface only to new lib |
| **toolmiddleware** | 500 | ❌ No | Keep in metatools-mcp |
| **handlers/server** | 1,000+ | ❌ No | Keep in metatools-mcp |

**Total extractable:** ~3,300 lines → Standalone libraries
**Remaining in metatools-mcp:** ~2,000 lines → Thin composition layer
