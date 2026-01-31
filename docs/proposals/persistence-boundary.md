# Persistence Boundary Architecture

**Status:** Draft
**Date:** 2026-01-30
**Related:** [Auth Middleware](./auth-middleware.md), [Multi-Tenancy](./multi-tenancy.md), [Pluggable Architecture](./pluggable-architecture.md)

## Overview

This document defines the boundary between **interface contracts** (in core libs/packages) and **persistence implementations** (in metatools-mcp internal packages). The principle: **interfaces define HOW to store, implementations define WHERE to store**.

Pluggability is achieved through **interfaces + dependency injection**, not library separation. Implementations live in `metatools-mcp/internal/` with build tags for optional backends.

---

## Design Principle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PERSISTENCE BOUNDARY                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   CORE LIBS / PACKAGES (interfaces)         metatools-mcp/internal/          │
│   ─────────────────────────────────         ───────────────────────────      │
│                                                                               │
│   toolindex                                 internal/persist/                 │
│     Index interface ──────────────────────→   index_memory.go                │
│     Searcher interface                        index_postgres.go (+build)     │
│                                               index_sqlite.go (+build)       │
│                                                                               │
│   tooldocs                                                                   │
│     Store interface ──────────────────────→   docs_memory.go                 │
│                                               docs_postgres.go (+build)      │
│                                               docs_file.go                   │
│                                                                               │
│   internal/auth                                                              │
│     APIKeyStore interface ────────────────→   apikey_memory.go               │
│     TokenStore interface                      apikey_redis.go (+build)       │
│     KeyProvider interface                     apikey_postgres.go (+build)    │
│                                                                               │
│   internal/tenant                                                            │
│     TenantStore interface ────────────────→   tenant_memory.go               │
│     QuotaStore interface                      tenant_postgres.go (+build)    │
│     RateLimitStore interface                  ratelimit_redis.go (+build)    │
│                                                                               │
│   internal/cache                                                             │
│     Cache interface ──────────────────────→   cache_memory.go                │
│                                               cache_redis.go (+build)        │
│                                                                               │
│   internal/vector (future)                                                   │
│     VectorIndex interface ────────────────→   vector_pgvector.go (+build)    │
│     Embedder interface                        vector_qdrant.go (+build)      │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Why Not a Separate Library?

A separate `toolpersist` library was considered but rejected:

| Separate Library | Internal Packages |
|------------------|-------------------|
| Extra repo to maintain | Single codebase |
| Complex versioning | Simple build tags |
| Public API surface | Internal, hidden |
| Overkill for single consumer | Right-sized |

**Pluggability comes from interfaces, not library separation.**

---

## Interface Contracts (Core Libraries)

### 1. toolindex - Registry Interface

```go
// toolindex/index.go (already exists, validated)

// Index is the core contract for tool registration and lookup.
// Implementations may be in-memory, file-backed, or database-backed.
type Index interface {
    // Register adds a tool to the index
    Register(ctx context.Context, tool Tool) error

    // Unregister removes a tool from the index
    Unregister(ctx context.Context, id string) error

    // Get retrieves a tool by ID
    Get(ctx context.Context, id string) (Tool, error)

    // List returns all tools matching the filter
    List(ctx context.Context, filter Filter) ([]Tool, error)

    // Search finds tools matching the query
    Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
}

// Searcher is the pluggable search strategy interface.
// Allows swapping BM25, semantic, or hybrid search without changing Index.
type Searcher interface {
    Search(ctx context.Context, query string, tools []Tool, opts SearchOptions) ([]SearchResult, error)
}
```

### 2. tooldocs - Documentation Store Interface

```go
// tooldocs/store.go (already exists, validated)

// Store is the contract for tool documentation storage.
// The interface supports tiered disclosure (summary → schema → full).
type Store interface {
    // Get retrieves documentation for a tool
    Get(ctx context.Context, toolID string, level DisclosureLevel) (*Documentation, error)

    // Set stores documentation for a tool
    Set(ctx context.Context, toolID string, doc *Documentation) error

    // Delete removes documentation for a tool
    Delete(ctx context.Context, toolID string) error

    // List returns all tool IDs with documentation
    List(ctx context.Context) ([]string, error)
}

// DisclosureLevel controls documentation detail
type DisclosureLevel int

const (
    LevelSummary DisclosureLevel = iota // Name + description
    LevelSchema                          // + input/output schema
    LevelFull                            // + examples, related tools
)
```

### 3. toolcache - Cache Interface (NEW)

```go
// toolcache/cache.go - Interface contract

// Cache is the contract for response caching.
// Implementations: MemoryCache, RedisCache, etc.
type Cache interface {
    // Get retrieves a cached value
    Get(ctx context.Context, key string) ([]byte, bool, error)

    // Set stores a value with TTL
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

    // Delete removes a cached value
    Delete(ctx context.Context, key string) error

    // Clear removes all cached values
    Clear(ctx context.Context) error

    // Stats returns cache statistics
    Stats(ctx context.Context) (*CacheStats, error)
}

// CacheStats provides cache metrics
type CacheStats struct {
    Hits       int64
    Misses     int64
    Size       int64
    MaxSize    int64
    Evictions  int64
}
```

### 4. Auth Stores - Authentication Persistence Interfaces

```go
// internal/auth/stores.go - Interface contracts

// APIKeyStore retrieves and validates API keys.
// Implementations: Memory (testing), Redis (distributed), Postgres (persistent)
type APIKeyStore interface {
    // Lookup returns the identity for an API key
    Lookup(ctx context.Context, key string) (*APIKeyInfo, error)

    // Create stores a new API key
    Create(ctx context.Context, info *APIKeyInfo) error

    // Revoke invalidates an API key
    Revoke(ctx context.Context, keyID string) error

    // List returns all API keys for a principal/tenant
    List(ctx context.Context, filter APIKeyFilter) ([]*APIKeyInfo, error)

    // UpdateLastUsed records key usage
    UpdateLastUsed(ctx context.Context, keyID string, at time.Time) error
}

// TokenStore manages opaque tokens (for OAuth2 introspection caching).
type TokenStore interface {
    // Get retrieves cached token info
    Get(ctx context.Context, token string) (*TokenInfo, bool, error)

    // Set caches token info
    Set(ctx context.Context, token string, info *TokenInfo, ttl time.Duration) error

    // Invalidate removes a token from cache
    Invalidate(ctx context.Context, token string) error
}

// KeyProvider fetches signing keys for JWT validation.
type KeyProvider interface {
    // GetKey returns the key for the given key ID
    GetKey(ctx context.Context, keyID string) (any, error)

    // GetKeys returns all available keys
    GetKeys(ctx context.Context) ([]any, error)
}

// SessionStore manages user sessions (optional).
type SessionStore interface {
    // Create starts a new session
    Create(ctx context.Context, identity *Identity) (*Session, error)

    // Get retrieves a session
    Get(ctx context.Context, sessionID string) (*Session, error)

    // Refresh extends session expiry
    Refresh(ctx context.Context, sessionID string) error

    // Destroy ends a session
    Destroy(ctx context.Context, sessionID string) error
}
```

### 5. Tenant Stores - Multi-Tenancy Persistence Interfaces

```go
// internal/tenant/stores.go - Interface contracts

// TenantStore manages tenant configuration and metadata.
type TenantStore interface {
    // Get retrieves tenant configuration
    Get(ctx context.Context, tenantID string) (*Tenant, error)

    // Create registers a new tenant
    Create(ctx context.Context, tenant *Tenant) error

    // Update modifies tenant configuration
    Update(ctx context.Context, tenant *Tenant) error

    // Delete removes a tenant
    Delete(ctx context.Context, tenantID string) error

    // List returns all tenants matching filter
    List(ctx context.Context, filter TenantFilter) ([]*Tenant, error)
}

// QuotaStore tracks resource usage against quotas.
type QuotaStore interface {
    // GetUsage returns current usage for a tenant
    GetUsage(ctx context.Context, tenantID string) (*QuotaUsage, error)

    // Increment adds to usage counter
    Increment(ctx context.Context, tenantID string, resource string, amount int64) error

    // Reset clears usage (e.g., at billing period)
    Reset(ctx context.Context, tenantID string) error

    // CheckQuota returns error if quota would be exceeded
    CheckQuota(ctx context.Context, tenantID string, resource string, amount int64) error
}

// RateLimitStore tracks rate limit state.
type RateLimitStore interface {
    // Allow checks if request is within rate limit
    Allow(ctx context.Context, key string, limit RateLimit) (bool, error)

    // GetState returns current rate limit state
    GetState(ctx context.Context, key string) (*RateLimitState, error)
}

// AuditLogger records audit events.
type AuditLogger interface {
    // Log records an audit event
    Log(ctx context.Context, event *AuditEvent) error

    // Query retrieves audit events
    Query(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error)
}
```

### 6. toolsemantic - Vector Search Interfaces (NEW)

```go
// toolsemantic/interfaces.go - Interface contracts

// VectorIndex stores and searches vector embeddings.
// Implementations: PgVector, Qdrant, Chroma, Pinecone, etc.
type VectorIndex interface {
    // Index stores a vector embedding
    Index(ctx context.Context, id string, vector []float32, metadata map[string]any) error

    // Search finds similar vectors
    Search(ctx context.Context, vector []float32, opts VectorSearchOptions) ([]VectorResult, error)

    // Delete removes a vector
    Delete(ctx context.Context, id string) error

    // BatchIndex stores multiple vectors
    BatchIndex(ctx context.Context, items []VectorItem) error
}

// Embedder generates vector embeddings from text.
// Implementations: OpenAI, Cohere, local models, etc.
type Embedder interface {
    // Embed generates embedding for text
    Embed(ctx context.Context, text string) ([]float32, error)

    // EmbedBatch generates embeddings for multiple texts
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Dimensions returns the embedding dimension
    Dimensions() int
}

// Reranker reorders search results by relevance.
// Implementations: Cohere, cross-encoder, etc.
type Reranker interface {
    // Rerank reorders results by relevance to query
    Rerank(ctx context.Context, query string, results []RerankCandidate) ([]RerankResult, error)
}
```

---

## Directory Structure

All implementations live in `metatools-mcp/internal/`:

```
metatools-mcp/
├── internal/
│   ├── auth/
│   │   ├── authenticator.go       # Authenticator interface
│   │   ├── authorizer.go          # Authorizer interface
│   │   ├── identity.go            # Identity types
│   │   ├── jwt.go                 # JWT implementation
│   │   ├── apikey.go              # APIKeyStore interface + memory impl
│   │   ├── apikey_redis.go        # +build redis
│   │   ├── apikey_postgres.go     # +build postgres
│   │   └── rbac.go                # Simple RBAC authorizer
│   │
│   ├── tenant/
│   │   ├── store.go               # TenantStore interface
│   │   ├── store_memory.go        # Memory implementation
│   │   ├── store_postgres.go      # +build postgres
│   │   ├── quota.go               # QuotaStore interface
│   │   ├── quota_memory.go        # Memory implementation
│   │   ├── quota_redis.go         # +build redis
│   │   ├── ratelimit.go           # RateLimitStore interface
│   │   └── ratelimit_redis.go     # +build redis
│   │
│   ├── cache/
│   │   ├── cache.go               # Cache interface
│   │   ├── memory.go              # LRU memory cache
│   │   └── redis.go               # +build redis
│   │
│   ├── persist/
│   │   ├── index_memory.go        # toolindex.Index memory impl
│   │   ├── index_postgres.go      # +build postgres
│   │   ├── index_sqlite.go        # +build sqlite
│   │   ├── docs_memory.go         # tooldocs.Store memory impl
│   │   ├── docs_postgres.go       # +build postgres
│   │   └── docs_file.go           # File-based implementation
│   │
│   └── vector/                    # Future: semantic search
│       ├── index.go               # VectorIndex interface
│       ├── embedder.go            # Embedder interface
│       ├── pgvector.go            # +build pgvector
│       └── qdrant.go              # +build qdrant
```

## Build Tags

Optional backends use Go build tags:

```go
//go:build redis

package cache

import "github.com/redis/go-redis/v9"

type RedisCache struct {
    client *redis.Client
}
```

**Build commands:**

```bash
# Default build (memory only)
go build ./cmd/metatools

# With Redis support
go build -tags redis ./cmd/metatools

# With PostgreSQL support
go build -tags postgres ./cmd/metatools

# Full build (all backends)
go build -tags "redis,postgres,sqlite" ./cmd/metatools
```

**Benefits:**
- Memory implementations always available (zero dependencies)
- Optional backends don't bloat binary size
- Clear opt-in for heavy dependencies (pgx, redis client)
- Single codebase, no library sprawl

---

## Dependency Injection Pattern

### Factory Registration

```go
// internal/persist/registry.go

package persist

// StoreType identifies a persistence backend
type StoreType string

const (
    StoreMemory   StoreType = "memory"
    StoreRedis    StoreType = "redis"
    StorePostgres StoreType = "postgres"
    StoreSQLite   StoreType = "sqlite"
)

// Registry holds factory functions for persistence implementations
type Registry struct {
    mu       sync.RWMutex
    index    map[StoreType]IndexFactory
    docs     map[StoreType]DocStoreFactory
    cache    map[StoreType]CacheFactory
}

// IndexFactory creates Index implementations
type IndexFactory func(cfg map[string]any) (toolindex.Index, error)

// DocStoreFactory creates DocStore implementations
type DocStoreFactory func(cfg map[string]any) (tooldocs.Store, error)

// CacheFactory creates Cache implementations
type CacheFactory func(cfg map[string]any) (Cache, error)

// DefaultRegistry is the global factory registry
var DefaultRegistry = NewRegistry()

func init() {
    // Memory implementations always registered
    DefaultRegistry.RegisterIndex(StoreMemory, NewMemoryIndex)
    DefaultRegistry.RegisterDocStore(StoreMemory, NewMemoryDocStore)
    DefaultRegistry.RegisterCache(StoreMemory, NewMemoryCache)
}
```

Build-tagged files register their implementations:

```go
//go:build postgres

// internal/persist/index_postgres.go
package persist

func init() {
    DefaultRegistry.RegisterIndex(StorePostgres, NewPostgresIndex)
    DefaultRegistry.RegisterDocStore(StorePostgres, NewPostgresDocStore)
}
```

### Configuration-Driven Initialization

```yaml
# metatools.yaml

persistence:
  # Tool index storage
  index:
    type: postgres  # memory, redis, postgres, sqlite
    config:
      connection_string: ${DATABASE_URL}
      table_prefix: "metatools_"

  # Documentation storage
  docs:
    type: postgres
    config:
      connection_string: ${DATABASE_URL}

  # Response cache
  cache:
    type: redis
    config:
      address: localhost:6379
      prefix: "cache:"

  # API key storage
  api_keys:
    type: postgres
    config:
      connection_string: ${DATABASE_URL}

  # Tenant storage
  tenants:
    type: postgres
    config:
      connection_string: ${DATABASE_URL}

  # Vector index (for semantic search)
  vectors:
    type: pgvector
    config:
      connection_string: ${DATABASE_URL}
      dimensions: 1536
```

### Runtime Initialization

```go
// cmd/metatools/persistence.go

package main

import (
    "github.com/jonwraymond/metatools-mcp/internal/persist"
    "github.com/jonwraymond/metatools-mcp/internal/cache"
    "github.com/jonwraymond/metatools-mcp/internal/auth"
)

func initPersistence(cfg *config.Config) (*PersistenceLayer, error) {
    // Create index (uses registry populated by build tags)
    indexFactory, ok := persist.DefaultRegistry.GetIndex(persist.StoreType(cfg.Persistence.Index.Type))
    if !ok {
        return nil, fmt.Errorf("unknown index type: %s (did you build with correct tags?)", cfg.Persistence.Index.Type)
    }
    idx, err := indexFactory(cfg.Persistence.Index.Config)
    if err != nil {
        return nil, fmt.Errorf("create index: %w", err)
    }

    // Create doc store
    docsFactory, ok := persist.DefaultRegistry.GetDocStore(persist.StoreType(cfg.Persistence.Docs.Type))
    if !ok {
        return nil, fmt.Errorf("unknown docs type: %s", cfg.Persistence.Docs.Type)
    }
    docs, err := docsFactory(cfg.Persistence.Docs.Config)
    if err != nil {
        return nil, fmt.Errorf("create docs: %w", err)
    }

    // Create cache
    cacheImpl, err := cache.New(cfg.Cache.Type, cfg.Cache.Config)
    if err != nil {
        return nil, fmt.Errorf("create cache: %w", err)
    }

    return &PersistenceLayer{
        Index: idx,
        Docs:  docs,
        Cache: cacheImpl,
    }, nil
}
```

---

## Migration Path

### Phase 1: Define Interfaces (Current)

1. Validate existing interfaces in toolindex, tooldocs
2. Add Cache interface to internal/cache
3. Add auth store interfaces to internal/auth
4. Add tenant store interfaces to internal/tenant

### Phase 2: Memory Implementations

1. Create internal/persist/ with memory implementations
2. Create internal/cache/ with LRU memory cache
3. Add factory registry pattern
4. Wire into existing handlers

### Phase 3: Add Optional Backends (Build Tags)

1. Redis implementations (cache, rate limit, api keys) - `+build redis`
2. PostgreSQL implementations (index, docs, tenants) - `+build postgres`
3. SQLite implementations (embedded deployments) - `+build sqlite`

### Phase 4: Configuration & CLI

1. Add persistence configuration section to metatools.yaml
2. Add CLI flags for common options (`--cache-type=redis`)
3. Initialize via factory registry at startup
4. Document build tag requirements

---

## Benefits

1. **Single codebase** - No extra libraries to maintain
2. **Testability** - Memory implementations always available
3. **Flexibility** - Swap backends via config without code changes
4. **Optional dependencies** - Build tags exclude heavy deps (pgx, redis)
5. **Clear contracts** - Interfaces define exactly what's needed
6. **Right-sized binary** - Only compiled backends are included
7. **Internal implementations** - Not part of public API surface

---

## Interface Location Summary

| Interface | Location | Purpose |
|-----------|----------|---------|
| `Index` | toolindex | Tool registration/lookup |
| `Searcher` | toolindex | Pluggable search strategy |
| `Store` (docs) | tooldocs | Documentation storage |
| `Cache` | internal/cache | Response caching |
| `APIKeyStore` | internal/auth | API key persistence |
| `TokenStore` | internal/auth | OAuth token caching |
| `KeyProvider` | internal/auth | JWT key fetching |
| `TenantStore` | internal/tenant | Tenant configuration |
| `QuotaStore` | internal/tenant | Quota tracking |
| `RateLimitStore` | internal/tenant | Rate limit state |
| `AuditLogger` | internal/audit | Audit logging |
| `VectorIndex` | internal/vector | Vector embeddings (future) |
| `Embedder` | internal/vector | Text→vector (future) |

## Implementation Location Summary

| Implementation | Location | Build Tag |
|----------------|----------|-----------|
| MemoryIndex | internal/persist/index_memory.go | (none) |
| PostgresIndex | internal/persist/index_postgres.go | `postgres` |
| SQLiteIndex | internal/persist/index_sqlite.go | `sqlite` |
| MemoryCache | internal/cache/memory.go | (none) |
| RedisCache | internal/cache/redis.go | `redis` |
| MemoryAPIKeyStore | internal/auth/apikey.go | (none) |
| RedisAPIKeyStore | internal/auth/apikey_redis.go | `redis` |
| PostgresAPIKeyStore | internal/auth/apikey_postgres.go | `postgres` |

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-30 | Initial persistence boundary architecture |
| 2026-01-30 | Simplified: removed toolpersist library, use internal packages with build tags |
