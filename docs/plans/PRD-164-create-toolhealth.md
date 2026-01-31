# PRD-164: Create toolhealth

**Phase:** 6 - Operations Layer
**Priority:** Medium
**Effort:** 6 hours
**Dependencies:** PRD-120

---

## Objective

Create a new `toolops/health/` package for health checking, readiness probes, and service status reporting.

---

## Package Design

**Location:** `github.com/ApertureStack/toolops/health`

**Purpose:**
- Health check endpoints
- Liveness and readiness probes
- Dependency health aggregation
- Status reporting

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Health Package | `toolops/health/` | Health check infrastructure |
| Checker | `health/checker.go` | Health check interface |
| Aggregator | `health/aggregator.go` | Multi-check aggregation |
| HTTP Handler | `health/handler.go` | HTTP endpoints |
| Tests | `health/*_test.go` | Comprehensive tests |

---

## Tasks

### Task 1: Create Package Structure

```bash
cd /tmp/migration/toolops
mkdir -p health
```

### Task 2: Define Health Check Interface

**File:** `toolops/health/checker.go`

```go
package health

import (
    "context"
    "time"
)

// Status represents health status.
type Status string

const (
    StatusHealthy   Status = "healthy"
    StatusUnhealthy Status = "unhealthy"
    StatusDegraded  Status = "degraded"
    StatusUnknown   Status = "unknown"
)

// CheckResult represents a health check result.
type CheckResult struct {
    Name      string
    Status    Status
    Message   string
    Timestamp time.Time
    Duration  time.Duration
    Details   map[string]any
}

// Checker performs health checks.
type Checker interface {
    // Name returns the check name.
    Name() string

    // Check performs the health check.
    Check(ctx context.Context) *CheckResult
}

// CheckFunc is a function that performs a health check.
type CheckFunc func(ctx context.Context) *CheckResult

// FuncChecker wraps a function as a Checker.
type FuncChecker struct {
    name string
    fn   CheckFunc
}

// NewFuncChecker creates a checker from a function.
func NewFuncChecker(name string, fn CheckFunc) *FuncChecker {
    return &FuncChecker{name: name, fn: fn}
}

func (c *FuncChecker) Name() string                         { return c.name }
func (c *FuncChecker) Check(ctx context.Context) *CheckResult { return c.fn(ctx) }
```

### Task 3: Implement Health Aggregator

**File:** `toolops/health/aggregator.go`

```go
package health

import (
    "context"
    "sync"
    "time"
)

// AggregatedResult contains results from multiple checks.
type AggregatedResult struct {
    Status    Status
    Checks    map[string]*CheckResult
    Timestamp time.Time
}

// Aggregator combines multiple health checkers.
type Aggregator struct {
    checkers []Checker
    timeout  time.Duration
    mu       sync.RWMutex
    cache    *AggregatedResult
    cacheTTL time.Duration
}

// AggregatorConfig configures the aggregator.
type AggregatorConfig struct {
    Timeout  time.Duration
    CacheTTL time.Duration
}

// NewAggregator creates a new health aggregator.
func NewAggregator(config AggregatorConfig) *Aggregator {
    return &Aggregator{
        checkers: make([]Checker, 0),
        timeout:  config.Timeout,
        cacheTTL: config.CacheTTL,
    }
}

// Register adds a checker to the aggregator.
func (a *Aggregator) Register(checker Checker) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.checkers = append(a.checkers, checker)
}

// Check runs all health checks.
func (a *Aggregator) Check(ctx context.Context) *AggregatedResult {
    // Check cache
    a.mu.RLock()
    if a.cache != nil && time.Since(a.cache.Timestamp) < a.cacheTTL {
        result := a.cache
        a.mu.RUnlock()
        return result
    }
    a.mu.RUnlock()

    // Run checks in parallel
    ctx, cancel := context.WithTimeout(ctx, a.timeout)
    defer cancel()

    results := make(map[string]*CheckResult)
    var wg sync.WaitGroup
    var mu sync.Mutex

    a.mu.RLock()
    checkers := make([]Checker, len(a.checkers))
    copy(checkers, a.checkers)
    a.mu.RUnlock()

    for _, checker := range checkers {
        wg.Add(1)
        go func(c Checker) {
            defer wg.Done()

            start := time.Now()
            result := c.Check(ctx)
            result.Duration = time.Since(start)
            result.Timestamp = start

            mu.Lock()
            results[c.Name()] = result
            mu.Unlock()
        }(checker)
    }

    wg.Wait()

    // Aggregate status
    overallStatus := StatusHealthy
    for _, result := range results {
        if result.Status == StatusUnhealthy {
            overallStatus = StatusUnhealthy
            break
        }
        if result.Status == StatusDegraded && overallStatus == StatusHealthy {
            overallStatus = StatusDegraded
        }
    }

    aggregated := &AggregatedResult{
        Status:    overallStatus,
        Checks:    results,
        Timestamp: time.Now(),
    }

    // Update cache
    a.mu.Lock()
    a.cache = aggregated
    a.mu.Unlock()

    return aggregated
}
```

### Task 4: Implement HTTP Handler

**File:** `toolops/health/handler.go`

```go
package health

import (
    "context"
    "encoding/json"
    "net/http"
)

// Handler provides HTTP health check endpoints.
type Handler struct {
    aggregator *Aggregator
    liveness   Checker
}

// NewHandler creates a new health HTTP handler.
func NewHandler(aggregator *Aggregator, liveness Checker) *Handler {
    return &Handler{
        aggregator: aggregator,
        liveness:   liveness,
    }
}

// LivenessHandler handles /health/live endpoint.
func (h *Handler) LivenessHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        result := h.liveness.Check(ctx)

        w.Header().Set("Content-Type", "application/json")

        if result.Status != StatusHealthy {
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        json.NewEncoder(w).Encode(result)
    }
}

// ReadinessHandler handles /health/ready endpoint.
func (h *Handler) ReadinessHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        result := h.aggregator.Check(ctx)

        w.Header().Set("Content-Type", "application/json")

        switch result.Status {
        case StatusHealthy:
            w.WriteHeader(http.StatusOK)
        case StatusDegraded:
            w.WriteHeader(http.StatusOK) // Still ready, but degraded
        default:
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        json.NewEncoder(w).Encode(result)
    }
}

// HealthHandler handles /health endpoint (full status).
func (h *Handler) HealthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        result := h.aggregator.Check(ctx)

        w.Header().Set("Content-Type", "application/json")

        switch result.Status {
        case StatusHealthy:
            w.WriteHeader(http.StatusOK)
        case StatusDegraded:
            w.WriteHeader(http.StatusOK)
        default:
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        json.NewEncoder(w).Encode(result)
    }
}

// RegisterRoutes registers health endpoints on a mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/health", h.HealthHandler())
    mux.HandleFunc("/health/live", h.LivenessHandler())
    mux.HandleFunc("/health/ready", h.ReadinessHandler())
}
```

### Task 5: Create Common Checkers

**File:** `toolops/health/checkers.go`

```go
package health

import (
    "context"
    "database/sql"
    "net/http"
    "time"
)

// DatabaseChecker checks database connectivity.
func DatabaseChecker(name string, db *sql.DB) Checker {
    return NewFuncChecker(name, func(ctx context.Context) *CheckResult {
        if err := db.PingContext(ctx); err != nil {
            return &CheckResult{
                Name:    name,
                Status:  StatusUnhealthy,
                Message: err.Error(),
            }
        }
        return &CheckResult{
            Name:    name,
            Status:  StatusHealthy,
            Message: "database connection ok",
        }
    })
}

// HTTPChecker checks HTTP endpoint availability.
func HTTPChecker(name, url string, timeout time.Duration) Checker {
    client := &http.Client{Timeout: timeout}

    return NewFuncChecker(name, func(ctx context.Context) *CheckResult {
        req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
        resp, err := client.Do(req)

        if err != nil {
            return &CheckResult{
                Name:    name,
                Status:  StatusUnhealthy,
                Message: err.Error(),
            }
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return &CheckResult{
                Name:    name,
                Status:  StatusUnhealthy,
                Message: "service returned " + resp.Status,
            }
        }

        return &CheckResult{
            Name:    name,
            Status:  StatusHealthy,
            Message: "endpoint reachable",
        }
    })
}

// MemoryChecker checks memory usage.
func MemoryChecker(name string, maxBytes uint64) Checker {
    return NewFuncChecker(name, func(ctx context.Context) *CheckResult {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        if m.Alloc > maxBytes {
            return &CheckResult{
                Name:    name,
                Status:  StatusDegraded,
                Message: "memory usage high",
                Details: map[string]any{
                    "alloc_bytes": m.Alloc,
                    "max_bytes":   maxBytes,
                },
            }
        }

        return &CheckResult{
            Name:   name,
            Status: StatusHealthy,
            Details: map[string]any{
                "alloc_bytes": m.Alloc,
            },
        }
    })
}
```

### Task 6: Create Package Documentation

**File:** `toolops/health/doc.go`

```go
// Package health provides health checking infrastructure.
//
// This package implements Kubernetes-style health probes for
// monitoring service availability and dependencies.
//
// # Health Checks
//
// Create custom health checkers:
//
//	dbChecker := health.DatabaseChecker("postgres", db)
//	redisChecker := health.HTTPChecker("redis", "http://redis:6379/ping", 5*time.Second)
//
// # Aggregation
//
// Combine multiple checks:
//
//	aggregator := health.NewAggregator(health.AggregatorConfig{
//	    Timeout:  5 * time.Second,
//	    CacheTTL: 10 * time.Second,
//	})
//	aggregator.Register(dbChecker)
//	aggregator.Register(redisChecker)
//
//	result := aggregator.Check(ctx)
//
// # HTTP Endpoints
//
// Expose health endpoints:
//
//	handler := health.NewHandler(aggregator, liveness)
//	handler.RegisterRoutes(mux)
//
// Endpoints:
//   - /health - Full health status
//   - /health/live - Liveness probe (is the process alive?)
//   - /health/ready - Readiness probe (can it serve traffic?)
//
// # Kubernetes Integration
//
// Configure Kubernetes probes:
//
//	livenessProbe:
//	  httpGet:
//	    path: /health/live
//	    port: 8080
//	readinessProbe:
//	  httpGet:
//	    path: /health/ready
//	    port: 8080
package health
```

### Task 7: Build and Test

```bash
cd /tmp/migration/toolops

go mod tidy
go build ./...
go test -v ./health/...
```

### Task 8: Commit and Push

```bash
cd /tmp/migration/toolops

git add -A
git commit -m "feat(health): add health check infrastructure

Create new health package for service health monitoring.

Package contents:
- Checker interface for health checks
- Aggregator for combining multiple checks
- HTTP handler for health endpoints
- Common checkers (database, HTTP, memory)

Features:
- Kubernetes-style liveness/readiness probes
- Parallel check execution
- Result caching
- Status aggregation
- Built-in common checkers

Endpoints:
- /health - Full health status
- /health/live - Liveness probe
- /health/ready - Readiness probe

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [ ] Checker interface defined
- [ ] Aggregator implemented
- [ ] HTTP handlers work
- [ ] Common checkers available
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Next Steps

- Gate G4: Operations layer complete (all 5 packages)
- PRD-170: Create tooltransport
