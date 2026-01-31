# PRD-161: Migrate toolcache

**Phase:** 6 - Operations Layer
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-120

---

## Objective

Migrate the existing `toolcache` repository into `toolops/cache/` as the second package in the consolidated operations layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolcache`
**Target Location:** `github.com/ApertureStack/toolops/cache`

**Package Contents:**
- Response caching middleware
- Memory and Redis backends
- TTL-based expiration
- Key generation strategies
- ~1,500 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Cache Package | `toolops/cache/` | Response caching |
| Tests | `toolops/cache/*_test.go` | All existing tests |
| Documentation | `toolops/cache/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone and Migrate

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolcache.git

cp toolcache/*.go toolops/cache/

cd toolops/cache
sed -i '' 's|github.com/ApertureStack/toolcache|github.com/ApertureStack/toolops/cache|g' *.go
sed -i '' 's|github.com/ApertureStack/toolmodel|github.com/ApertureStack/toolfoundation/model|g' *.go
```

### Task 2: Update Package Documentation

**File:** `toolops/cache/doc.go`

```go
// Package cache provides response caching for tool executions.
//
// This package implements caching middleware that reduces latency and load
// by storing and reusing tool execution results.
//
// # Cache Backends
//
// Built-in cache implementations:
//
//   - MemoryCache: In-process LRU cache
//   - RedisCache: Distributed cache using Redis
//
// # Usage
//
// Create and use a cache:
//
//	cache := cache.NewMemoryCache(cache.MemoryConfig{
//	    MaxSize: 1000,
//	    TTL:     5 * time.Minute,
//	})
//
//	// Cache middleware
//	cached := cache.Middleware(provider, cache.MiddlewareConfig{
//	    Cache:       cache,
//	    KeyGen:      cache.DefaultKeyGenerator,
//	    SkipTools:   []string{"random"},
//	    ToolTTLs:    map[string]time.Duration{"search": 1*time.Minute},
//	})
//
// # Key Generation
//
// Keys are generated from tool ID and arguments:
//
//	keyGen := cache.NewKeyGenerator(cache.KeyConfig{
//	    Prefix:      "metatools:",
//	    IncludeNS:   true,
//	    IgnoredArgs: []string{"timestamp"},
//	})
//
// # Cache Statistics
//
// Monitor cache performance:
//
//	stats := cache.Stats()
//	fmt.Printf("Hits: %d, Misses: %d, Ratio: %.2f%%\n",
//	    stats.Hits, stats.Misses, stats.HitRatio*100)
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolcache as part of
// the ApertureStack consolidation.
package cache
```

### Task 3: Build and Test

```bash
cd /tmp/migration/toolops

go mod tidy
go build ./...
go test -v ./cache/...
```

### Task 4: Commit and Push

```bash
cd /tmp/migration/toolops

git add -A
git commit -m "feat(cache): migrate toolcache package

Migrate response caching from standalone toolcache repository.

Package contents:
- Cache interface for pluggable backends
- MemoryCache with LRU eviction
- RedisCache for distributed caching
- Caching middleware
- Key generation strategies

Features:
- TTL-based expiration
- Per-tool TTL overrides
- Skip list for uncacheable tools
- Cache statistics
- Thread-safe operations

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolcache → toolops/cache

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-162: Extract toolauth
- PRD-163: Create toolresilience
