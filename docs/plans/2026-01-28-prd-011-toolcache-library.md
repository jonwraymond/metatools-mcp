# PRD-011: toolcache Library Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a response caching library with pluggable backends (memory, Redis) supporting TTL, key generation, and cache invalidation for tool execution results.

**Architecture:** Cache middleware wrapping tool providers with configurable key generation, TTL per tool, and layered caching (L1 memory + L2 Redis).

**Tech Stack:** Go, go-redis (optional), hashicorp/golang-lru

---

## Overview

The `toolcache` library provides response caching for tool execution, reducing latency and backend load for repeated queries with identical inputs.

**Reference:** [pluggable-architecture.md](../proposals/pluggable-architecture.md) - Cache Layer section

---

## Directory Structure

```
toolcache/
├── cache.go           # Cache interface and types
├── cache_test.go
├── memory.go          # In-memory cache backend
├── memory_test.go
├── redis.go           # Redis cache backend
├── redis_test.go
├── layered.go         # L1 + L2 layered cache
├── layered_test.go
├── key.go             # Key generation strategies
├── key_test.go
├── middleware.go      # Caching middleware
├── middleware_test.go
├── config.go          # Configuration types
├── doc.go
├── go.mod
└── go.sum
```

---

## Task 1: Cache Interface and Types

**Files:**
- Create: `toolcache/cache.go`
- Create: `toolcache/cache_test.go`
- Create: `toolcache/go.mod`

**Step 1: Write failing tests**

```go
// cache_test.go
package toolcache_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolcache"
)

func TestCacheEntry_IsExpired(t *testing.T) {
    tests := []struct {
        name     string
        entry    toolcache.Entry
        expected bool
    }{
        {
            name: "not expired",
            entry: toolcache.Entry{
                ExpiresAt: time.Now().Add(time.Hour),
            },
            expected: false,
        },
        {
            name: "expired",
            entry: toolcache.Entry{
                ExpiresAt: time.Now().Add(-time.Hour),
            },
            expected: true,
        },
        {
            name: "zero expiration (never expires)",
            entry: toolcache.Entry{
                ExpiresAt: time.Time{},
            },
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.entry.IsExpired())
        })
    }
}

func TestCacheStats(t *testing.T) {
    stats := &toolcache.Stats{}

    stats.RecordHit()
    stats.RecordHit()
    stats.RecordMiss()

    assert.Equal(t, int64(2), stats.Hits)
    assert.Equal(t, int64(1), stats.Misses)
    assert.InDelta(t, 0.666, stats.HitRate(), 0.01)
}

func TestCacheConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  toolcache.Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: toolcache.Config{
                DefaultTTL: time.Minute,
                MaxSize:    1000,
            },
            wantErr: false,
        },
        {
            name: "negative TTL",
            config: toolcache.Config{
                DefaultTTL: -time.Minute,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolcache && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// go.mod
module github.com/jrraymond/toolcache

go 1.22

require (
    github.com/hashicorp/golang-lru/v2 v2.0.7
    github.com/redis/go-redis/v9 v9.5.1
)
```

```go
// cache.go
package toolcache

import (
    "context"
    "errors"
    "sync/atomic"
    "time"
)

// Cache is the interface for cache backends
type Cache interface {
    // Get retrieves a value from cache
    Get(ctx context.Context, key string) ([]byte, bool, error)

    // Set stores a value in cache with TTL
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

    // Delete removes a value from cache
    Delete(ctx context.Context, key string) error

    // Clear removes entries matching pattern (empty pattern = all)
    Clear(ctx context.Context, pattern string) error

    // Stats returns cache statistics
    Stats() Stats

    // Close closes the cache
    Close() error
}

// Entry represents a cached value
type Entry struct {
    Value     []byte
    CreatedAt time.Time
    ExpiresAt time.Time
    ToolID    string
    InputHash string
}

// IsExpired checks if the entry is expired
func (e Entry) IsExpired() bool {
    if e.ExpiresAt.IsZero() {
        return false // Never expires
    }
    return time.Now().After(e.ExpiresAt)
}

// Stats holds cache statistics
type Stats struct {
    Hits       int64
    Misses     int64
    Evictions  int64
    Size       int64
    MaxSize    int64
}

// RecordHit increments hit counter
func (s *Stats) RecordHit() {
    atomic.AddInt64(&s.Hits, 1)
}

// RecordMiss increments miss counter
func (s *Stats) RecordMiss() {
    atomic.AddInt64(&s.Misses, 1)
}

// RecordEviction increments eviction counter
func (s *Stats) RecordEviction() {
    atomic.AddInt64(&s.Evictions, 1)
}

// HitRate returns the cache hit rate
func (s *Stats) HitRate() float64 {
    total := s.Hits + s.Misses
    if total == 0 {
        return 0
    }
    return float64(s.Hits) / float64(total)
}

// Config holds cache configuration
type Config struct {
    // General settings
    DefaultTTL time.Duration
    MaxSize    int

    // Per-tool TTL overrides
    ToolTTLs map[string]time.Duration

    // Memory cache specific
    Memory MemoryConfig

    // Redis cache specific
    Redis RedisConfig

    // Layered cache specific
    Layered LayeredConfig
}

// MemoryConfig holds in-memory cache settings
type MemoryConfig struct {
    MaxEntries int
    MaxBytes   int64
}

// RedisConfig holds Redis cache settings
type RedisConfig struct {
    Addr     string
    Password string
    DB       int
    Prefix   string
}

// LayeredConfig holds layered cache settings
type LayeredConfig struct {
    L1TTL time.Duration // Short TTL for L1 (memory)
    L2TTL time.Duration // Longer TTL for L2 (Redis)
}

// Validate validates the configuration
func (c Config) Validate() error {
    if c.DefaultTTL < 0 {
        return errors.New("default TTL cannot be negative")
    }
    return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
    return Config{
        DefaultTTL: 5 * time.Minute,
        MaxSize:    10000,
        Memory: MemoryConfig{
            MaxEntries: 10000,
        },
        Layered: LayeredConfig{
            L1TTL: time.Minute,
            L2TTL: 10 * time.Minute,
        },
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolcache && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolcache/
git commit -m "$(cat <<'EOF'
feat(toolcache): add Cache interface and types

- Cache interface with Get, Set, Delete, Clear, Stats
- Entry type with expiration checking
- Stats with atomic counters and hit rate
- Config with per-tool TTL overrides

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: In-Memory Cache Backend

**Files:**
- Create: `toolcache/memory.go`
- Create: `toolcache/memory_test.go`

**Step 1: Write failing tests**

```go
// memory_test.go
package toolcache_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolcache"
)

func TestMemoryCache_SetGet(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    ctx := context.Background()
    key := "test-key"
    value := []byte(`{"result": "test"}`)

    err := cache.Set(ctx, key, value, time.Minute)
    require.NoError(t, err)

    got, ok, err := cache.Get(ctx, key)
    require.NoError(t, err)
    assert.True(t, ok)
    assert.Equal(t, value, got)
}

func TestMemoryCache_GetMiss(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    _, ok, err := cache.Get(context.Background(), "nonexistent")
    require.NoError(t, err)
    assert.False(t, ok)
}

func TestMemoryCache_Expiration(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    ctx := context.Background()
    key := "expiring-key"
    value := []byte(`test`)

    // Set with very short TTL
    err := cache.Set(ctx, key, value, 10*time.Millisecond)
    require.NoError(t, err)

    // Should exist immediately
    _, ok, _ := cache.Get(ctx, key)
    assert.True(t, ok)

    // Wait for expiration
    time.Sleep(20 * time.Millisecond)

    // Should be expired
    _, ok, _ = cache.Get(ctx, key)
    assert.False(t, ok)
}

func TestMemoryCache_Delete(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    ctx := context.Background()
    key := "delete-key"

    cache.Set(ctx, key, []byte(`test`), time.Minute)
    cache.Delete(ctx, key)

    _, ok, _ := cache.Get(ctx, key)
    assert.False(t, ok)
}

func TestMemoryCache_Clear(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    ctx := context.Background()

    // Add multiple entries
    cache.Set(ctx, "key1", []byte(`1`), time.Minute)
    cache.Set(ctx, "key2", []byte(`2`), time.Minute)
    cache.Set(ctx, "key3", []byte(`3`), time.Minute)

    // Clear all
    err := cache.Clear(ctx, "")
    require.NoError(t, err)

    // All should be gone
    _, ok, _ := cache.Get(ctx, "key1")
    assert.False(t, ok)
}

func TestMemoryCache_Eviction(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 2,
    })
    defer cache.Close()

    ctx := context.Background()

    // Add more entries than max
    cache.Set(ctx, "key1", []byte(`1`), time.Minute)
    cache.Set(ctx, "key2", []byte(`2`), time.Minute)
    cache.Set(ctx, "key3", []byte(`3`), time.Minute)

    // At least one should have been evicted
    stats := cache.Stats()
    assert.GreaterOrEqual(t, stats.Evictions, int64(1))
}

func TestMemoryCache_Stats(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{
        MaxEntries: 100,
    })
    defer cache.Close()

    ctx := context.Background()

    cache.Set(ctx, "key", []byte(`test`), time.Minute)
    cache.Get(ctx, "key")        // hit
    cache.Get(ctx, "nonexistent") // miss

    stats := cache.Stats()
    assert.Equal(t, int64(1), stats.Hits)
    assert.Equal(t, int64(1), stats.Misses)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolcache && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// memory.go
package toolcache

import (
    "context"
    "sync"
    "time"

    lru "github.com/hashicorp/golang-lru/v2"
)

// memoryEntry stores value with metadata
type memoryEntry struct {
    value     []byte
    expiresAt time.Time
}

// MemoryCache is an in-memory LRU cache
type MemoryCache struct {
    cache *lru.Cache[string, *memoryEntry]
    stats Stats
    mu    sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config MemoryConfig) *MemoryCache {
    size := config.MaxEntries
    if size <= 0 {
        size = 10000
    }

    mc := &MemoryCache{}

    cache, _ := lru.NewWithEvict[string, *memoryEntry](size, func(key string, value *memoryEntry) {
        mc.stats.RecordEviction()
    })

    mc.cache = cache
    return mc
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.cache.Get(key)
    if !ok {
        c.stats.RecordMiss()
        return nil, false, nil
    }

    // Check expiration
    if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
        c.cache.Remove(key)
        c.stats.RecordMiss()
        return nil, false, nil
    }

    c.stats.RecordHit()
    return entry.value, true, nil
}

// Set stores a value in cache
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    entry := &memoryEntry{
        value: value,
    }

    if ttl > 0 {
        entry.expiresAt = time.Now().Add(ttl)
    }

    c.cache.Add(key, entry)
    return nil
}

// Delete removes a value from cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.cache.Remove(key)
    return nil
}

// Clear removes all entries (pattern ignored for memory cache)
func (c *MemoryCache) Clear(ctx context.Context, pattern string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.cache.Purge()
    return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() Stats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    stats := c.stats
    stats.Size = int64(c.cache.Len())
    return stats
}

// Close closes the cache
func (c *MemoryCache) Close() error {
    c.cache.Purge()
    return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolcache && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolcache/
git commit -m "$(cat <<'EOF'
feat(toolcache): add in-memory LRU cache backend

- LRU eviction with configurable size
- TTL-based expiration checking
- Thread-safe operations
- Statistics tracking (hits, misses, evictions)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Key Generation Strategies

**Files:**
- Create: `toolcache/key.go`
- Create: `toolcache/key_test.go`

**Step 1: Write failing tests**

```go
// key_test.go
package toolcache_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/jrraymond/toolcache"
)

func TestKeyGenerator_Default(t *testing.T) {
    gen := toolcache.DefaultKeyGenerator()

    key1 := gen.Generate("mcp:search", map[string]any{"query": "test"})
    key2 := gen.Generate("mcp:search", map[string]any{"query": "test"})
    key3 := gen.Generate("mcp:search", map[string]any{"query": "different"})

    // Same inputs should produce same key
    assert.Equal(t, key1, key2)

    // Different inputs should produce different key
    assert.NotEqual(t, key1, key3)
}

func TestKeyGenerator_WithPrefix(t *testing.T) {
    gen := toolcache.NewKeyGenerator(toolcache.KeyConfig{
        Prefix: "cache:",
    })

    key := gen.Generate("mcp:search", map[string]any{"query": "test"})
    assert.HasPrefix(t, key, "cache:")
}

func TestKeyGenerator_WithNamespace(t *testing.T) {
    gen := toolcache.NewKeyGenerator(toolcache.KeyConfig{
        Prefix:           "cache:",
        IncludeNamespace: true,
    })

    key := gen.Generate("mcp:search", map[string]any{"query": "test"})
    assert.Contains(t, key, "mcp")
}

func TestKeyGenerator_IgnoredArgs(t *testing.T) {
    gen := toolcache.NewKeyGenerator(toolcache.KeyConfig{
        IgnoredArgs: []string{"timestamp", "requestId"},
    })

    key1 := gen.Generate("mcp:search", map[string]any{
        "query":     "test",
        "timestamp": "2024-01-01",
    })
    key2 := gen.Generate("mcp:search", map[string]any{
        "query":     "test",
        "timestamp": "2024-12-31", // Different timestamp
    })

    // Should be same because timestamp is ignored
    assert.Equal(t, key1, key2)
}

func TestKeyGenerator_SortedArgs(t *testing.T) {
    gen := toolcache.DefaultKeyGenerator()

    key1 := gen.Generate("tool", map[string]any{"a": 1, "b": 2, "c": 3})
    key2 := gen.Generate("tool", map[string]any{"c": 3, "a": 1, "b": 2})

    // Order shouldn't matter
    assert.Equal(t, key1, key2)
}

func TestKeyGenerator_NestedArgs(t *testing.T) {
    gen := toolcache.DefaultKeyGenerator()

    key1 := gen.Generate("tool", map[string]any{
        "nested": map[string]any{
            "value": "test",
        },
    })
    key2 := gen.Generate("tool", map[string]any{
        "nested": map[string]any{
            "value": "different",
        },
    })

    assert.NotEqual(t, key1, key2)
}

func TestKeyGenerator_EmptyArgs(t *testing.T) {
    gen := toolcache.DefaultKeyGenerator()

    key := gen.Generate("mcp:search", nil)
    assert.NotEmpty(t, key)

    key2 := gen.Generate("mcp:search", map[string]any{})
    assert.Equal(t, key, key2)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolcache && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// key.go
package toolcache

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "sort"
    "strings"
)

// KeyGenerator generates cache keys
type KeyGenerator interface {
    Generate(toolID string, args map[string]any) string
}

// KeyConfig holds key generation configuration
type KeyConfig struct {
    Prefix           string
    IncludeNamespace bool
    IgnoredArgs      []string
    HashAlgorithm    string // sha256, md5
}

// DefaultKeyGenerator returns a default key generator
func DefaultKeyGenerator() KeyGenerator {
    return NewKeyGenerator(KeyConfig{})
}

// defaultKeyGenerator implements KeyGenerator
type defaultKeyGenerator struct {
    config KeyConfig
}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator(config KeyConfig) KeyGenerator {
    return &defaultKeyGenerator{config: config}
}

// Generate generates a cache key
func (g *defaultKeyGenerator) Generate(toolID string, args map[string]any) string {
    var sb strings.Builder

    // Add prefix
    if g.config.Prefix != "" {
        sb.WriteString(g.config.Prefix)
    }

    // Add namespace if configured
    if g.config.IncludeNamespace {
        if idx := strings.Index(toolID, ":"); idx > 0 {
            sb.WriteString(toolID[:idx])
            sb.WriteString(":")
        }
    }

    // Add tool ID
    sb.WriteString(toolID)
    sb.WriteString(":")

    // Hash the arguments
    argsHash := g.hashArgs(args)
    sb.WriteString(argsHash)

    return sb.String()
}

// hashArgs creates a deterministic hash of arguments
func (g *defaultKeyGenerator) hashArgs(args map[string]any) string {
    if args == nil || len(args) == 0 {
        return "empty"
    }

    // Filter ignored args
    filtered := make(map[string]any)
    for k, v := range args {
        if !g.isIgnored(k) {
            filtered[k] = v
        }
    }

    if len(filtered) == 0 {
        return "empty"
    }

    // Sort keys for deterministic output
    keys := make([]string, 0, len(filtered))
    for k := range filtered {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // Build sorted map for JSON encoding
    sorted := make([]any, 0, len(keys)*2)
    for _, k := range keys {
        sorted = append(sorted, k, filtered[k])
    }

    // JSON encode
    data, err := json.Marshal(sorted)
    if err != nil {
        // Fallback to simple concatenation
        var sb strings.Builder
        for _, k := range keys {
            sb.WriteString(k)
            sb.WriteString("=")
            sb.WriteString(stringify(filtered[k]))
            sb.WriteString(";")
        }
        data = []byte(sb.String())
    }

    // Hash
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:8]) // Use first 8 bytes
}

// isIgnored checks if an argument should be ignored
func (g *defaultKeyGenerator) isIgnored(key string) bool {
    for _, ignored := range g.config.IgnoredArgs {
        if key == ignored {
            return true
        }
    }
    return false
}

// stringify converts a value to string
func stringify(v any) string {
    switch val := v.(type) {
    case string:
        return val
    case nil:
        return "null"
    default:
        data, _ := json.Marshal(val)
        return string(data)
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolcache && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolcache/
git commit -m "$(cat <<'EOF'
feat(toolcache): add key generation strategies

- DefaultKeyGenerator with SHA256 hashing
- Configurable prefix and namespace inclusion
- IgnoredArgs for excluding non-deterministic fields
- Sorted arguments for deterministic keys

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Caching Middleware

**Files:**
- Create: `toolcache/middleware.go`
- Create: `toolcache/middleware_test.go`

**Step 1: Write failing tests**

```go
// middleware_test.go
package toolcache_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolcache"
)

type MockProvider struct {
    name       string
    callCount  int
    handler    func(ctx context.Context, input map[string]any) (any, error)
}

func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) Handle(ctx context.Context, input map[string]any) (any, error) {
    m.callCount++
    return m.handler(ctx, input)
}

func TestCacheMiddleware_CachesResult(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{MaxEntries: 100})
    defer cache.Close()

    provider := &MockProvider{
        name: "mcp:search",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return map[string]any{"results": []string{"a", "b", "c"}}, nil
        },
    }

    middleware := toolcache.CacheMiddleware(cache, toolcache.CacheMiddlewareConfig{
        DefaultTTL: time.Minute,
    })

    wrapped := middleware(provider)

    ctx := context.Background()
    input := map[string]any{"query": "test"}

    // First call - should execute provider
    result1, err := wrapped.Handle(ctx, input)
    require.NoError(t, err)
    assert.Equal(t, 1, provider.callCount)

    // Second call - should use cache
    result2, err := wrapped.Handle(ctx, input)
    require.NoError(t, err)
    assert.Equal(t, 1, provider.callCount) // Still 1
    assert.Equal(t, result1, result2)
}

func TestCacheMiddleware_DifferentInputs(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{MaxEntries: 100})
    defer cache.Close()

    provider := &MockProvider{
        name: "mcp:search",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return map[string]any{"query": input["query"]}, nil
        },
    }

    wrapped := toolcache.CacheMiddleware(cache, toolcache.CacheMiddlewareConfig{
        DefaultTTL: time.Minute,
    })(provider)

    ctx := context.Background()

    wrapped.Handle(ctx, map[string]any{"query": "test1"})
    wrapped.Handle(ctx, map[string]any{"query": "test2"})

    assert.Equal(t, 2, provider.callCount)
}

func TestCacheMiddleware_SkipCache(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{MaxEntries: 100})
    defer cache.Close()

    provider := &MockProvider{
        name: "mcp:execute",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return map[string]any{"executed": true}, nil
        },
    }

    wrapped := toolcache.CacheMiddleware(cache, toolcache.CacheMiddlewareConfig{
        DefaultTTL:    time.Minute,
        SkippedTools: []string{"mcp:execute"},
    })(provider)

    ctx := context.Background()
    input := map[string]any{}

    wrapped.Handle(ctx, input)
    wrapped.Handle(ctx, input)

    // Should not cache - both calls execute
    assert.Equal(t, 2, provider.callCount)
}

func TestCacheMiddleware_PerToolTTL(t *testing.T) {
    cache := toolcache.NewMemoryCache(toolcache.MemoryConfig{MaxEntries: 100})
    defer cache.Close()

    provider := &MockProvider{
        name: "mcp:search",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return "result", nil
        },
    }

    wrapped := toolcache.CacheMiddleware(cache, toolcache.CacheMiddlewareConfig{
        DefaultTTL: time.Hour,
        ToolTTLs: map[string]time.Duration{
            "mcp:search": 10 * time.Millisecond,
        },
    })(provider)

    ctx := context.Background()
    input := map[string]any{"query": "test"}

    wrapped.Handle(ctx, input)
    time.Sleep(20 * time.Millisecond)
    wrapped.Handle(ctx, input)

    // Should have expired and called twice
    assert.Equal(t, 2, provider.callCount)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolcache && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// middleware.go
package toolcache

import (
    "context"
    "encoding/json"
    "time"
)

// ToolProvider is the interface for tool providers
type ToolProvider interface {
    Name() string
    Handle(ctx context.Context, input map[string]any) (any, error)
}

// Middleware wraps a ToolProvider
type Middleware func(ToolProvider) ToolProvider

// CacheMiddlewareConfig holds middleware configuration
type CacheMiddlewareConfig struct {
    DefaultTTL   time.Duration
    ToolTTLs     map[string]time.Duration
    SkippedTools []string
    KeyGenerator KeyGenerator
}

// cachedProvider wraps a provider with caching
type cachedProvider struct {
    cache     Cache
    config    CacheMiddlewareConfig
    keyGen    KeyGenerator
    next      ToolProvider
}

// CacheMiddleware creates caching middleware
func CacheMiddleware(cache Cache, config CacheMiddlewareConfig) Middleware {
    keyGen := config.KeyGenerator
    if keyGen == nil {
        keyGen = DefaultKeyGenerator()
    }

    return func(next ToolProvider) ToolProvider {
        return &cachedProvider{
            cache:  cache,
            config: config,
            keyGen: keyGen,
            next:   next,
        }
    }
}

func (p *cachedProvider) Name() string {
    return p.next.Name()
}

func (p *cachedProvider) Handle(ctx context.Context, input map[string]any) (any, error) {
    toolID := p.next.Name()

    // Check if tool should skip cache
    if p.shouldSkip(toolID) {
        return p.next.Handle(ctx, input)
    }

    // Generate cache key
    key := p.keyGen.Generate(toolID, input)

    // Try to get from cache
    data, ok, err := p.cache.Get(ctx, key)
    if err == nil && ok {
        var result any
        if json.Unmarshal(data, &result) == nil {
            return result, nil
        }
    }

    // Execute provider
    result, err := p.next.Handle(ctx, input)
    if err != nil {
        return nil, err
    }

    // Cache result
    ttl := p.getTTL(toolID)
    if ttl > 0 {
        if data, jsonErr := json.Marshal(result); jsonErr == nil {
            p.cache.Set(ctx, key, data, ttl)
        }
    }

    return result, nil
}

// shouldSkip checks if tool should skip caching
func (p *cachedProvider) shouldSkip(toolID string) bool {
    for _, skipped := range p.config.SkippedTools {
        if toolID == skipped {
            return true
        }
    }
    return false
}

// getTTL gets the TTL for a tool
func (p *cachedProvider) getTTL(toolID string) time.Duration {
    if ttl, ok := p.config.ToolTTLs[toolID]; ok {
        return ttl
    }
    return p.config.DefaultTTL
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolcache && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolcache/
git commit -m "$(cat <<'EOF'
feat(toolcache): add caching middleware

- CacheMiddleware wraps ToolProvider with caching
- Per-tool TTL configuration
- Skipped tools list for non-cacheable operations
- Configurable key generator

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-011 complete:

- [ ] All tests pass: `go test ./... -v`
- [ ] Code coverage > 80%: `go test ./... -cover`
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete
- [ ] Integration verified:
  - [ ] Memory cache works correctly
  - [ ] Key generation is deterministic
  - [ ] Middleware caches appropriately
  - [ ] TTL expiration works

---

## Definition of Done

1. **Cache** interface with Get, Set, Delete, Clear, Stats
2. **MemoryCache** with LRU eviction and TTL
3. **KeyGenerator** with deterministic hashing
4. **CacheMiddleware** for automatic caching
5. **Configuration** with per-tool TTLs
6. All tests passing with >80% coverage
7. Documentation complete
