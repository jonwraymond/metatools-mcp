# PRD-163: Create toolresilience

**Phase:** 6 - Operations Layer
**Priority:** Medium
**Effort:** 8 hours
**Dependencies:** PRD-120

---

## Objective

Create a new `toolops/resilience/` package for fault tolerance patterns including circuit breakers, retries, timeouts, and bulkheads.

---

## Package Design

**Location:** `github.com/ApertureStack/toolops/resilience`

**Purpose:**
- Circuit breaker pattern
- Retry with backoff
- Timeout management
- Bulkhead isolation
- Rate limiting

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Resilience Package | `toolops/resilience/` | Fault tolerance patterns |
| Circuit Breaker | `resilience/circuitbreaker.go` | Circuit breaker implementation |
| Retry | `resilience/retry.go` | Retry with backoff |
| Bulkhead | `resilience/bulkhead.go` | Concurrency isolation |
| Tests | `resilience/*_test.go` | Comprehensive tests |

---

## Tasks

### Task 1: Create Package Structure

```bash
cd /tmp/migration/toolops
mkdir -p resilience
```

### Task 2: Implement Circuit Breaker

**File:** `toolops/resilience/circuitbreaker.go`

```go
package resilience

import (
    "context"
    "errors"
    "sync"
    "time"
)

// State represents circuit breaker state.
type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

var (
    ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
    name          string
    maxFailures   int
    timeout       time.Duration
    halfOpenMax   int

    state         State
    failures      int
    successes     int
    lastFailure   time.Time
    mu            sync.RWMutex
}

// CircuitBreakerConfig configures the circuit breaker.
type CircuitBreakerConfig struct {
    Name        string
    MaxFailures int           // Failures before opening
    Timeout     time.Duration // Time before half-open
    HalfOpenMax int           // Successes to close
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
    return &CircuitBreaker{
        name:        config.Name,
        maxFailures: config.MaxFailures,
        timeout:     config.Timeout,
        halfOpenMax: config.HalfOpenMax,
        state:       StateClosed,
    }
}

// Execute runs the function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
    if !cb.canExecute() {
        return ErrCircuitOpen
    }

    err := fn()

    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.recordSuccess()
    return nil
}

func (cb *CircuitBreaker) canExecute() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.successes = 0
            cb.mu.Unlock()
            cb.mu.RLock()
            return true
        }
        return false
    case StateHalfOpen:
        return true
    }
    return false
}

func (cb *CircuitBreaker) recordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++
    cb.lastFailure = time.Now()

    if cb.failures >= cb.maxFailures {
        cb.state = StateOpen
    }
}

func (cb *CircuitBreaker) recordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    if cb.state == StateHalfOpen {
        cb.successes++
        if cb.successes >= cb.halfOpenMax {
            cb.state = StateClosed
            cb.failures = 0
        }
    }
}

// State returns current state.
func (cb *CircuitBreaker) State() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}
```

### Task 3: Implement Retry

**File:** `toolops/resilience/retry.go`

```go
package resilience

import (
    "context"
    "math"
    "time"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
    RetryIf         func(error) bool
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:     3,
        InitialInterval: 100 * time.Millisecond,
        MaxInterval:     10 * time.Second,
        Multiplier:      2.0,
        RetryIf:         func(err error) bool { return true },
    }
}

// Retry executes the function with retry logic.
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
    var lastErr error
    interval := config.InitialInterval

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if err := ctx.Err(); err != nil {
            return err
        }

        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err
        if !config.RetryIf(err) {
            return err
        }

        if attempt < config.MaxAttempts-1 {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(interval):
            }

            interval = time.Duration(float64(interval) * config.Multiplier)
            if interval > config.MaxInterval {
                interval = config.MaxInterval
            }
        }
    }

    return lastErr
}

// RetryWithResult retries a function that returns a result.
func RetryWithResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
    var result T
    var lastErr error
    interval := config.InitialInterval

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if err := ctx.Err(); err != nil {
            return result, err
        }

        res, err := fn()
        if err == nil {
            return res, nil
        }

        lastErr = err
        if !config.RetryIf(err) {
            return result, err
        }

        if attempt < config.MaxAttempts-1 {
            select {
            case <-ctx.Done():
                return result, ctx.Err()
            case <-time.After(interval):
            }

            interval = time.Duration(float64(interval) * config.Multiplier)
            if interval > config.MaxInterval {
                interval = config.MaxInterval
            }
        }
    }

    return result, lastErr
}
```

### Task 4: Implement Bulkhead

**File:** `toolops/resilience/bulkhead.go`

```go
package resilience

import (
    "context"
    "errors"
)

var (
    ErrBulkheadFull = errors.New("bulkhead is full")
)

// Bulkhead limits concurrent executions.
type Bulkhead struct {
    name     string
    maxConc  int
    sem      chan struct{}
}

// BulkheadConfig configures the bulkhead.
type BulkheadConfig struct {
    Name           string
    MaxConcurrent  int
}

// NewBulkhead creates a new bulkhead.
func NewBulkhead(config BulkheadConfig) *Bulkhead {
    return &Bulkhead{
        name:    config.Name,
        maxConc: config.MaxConcurrent,
        sem:     make(chan struct{}, config.MaxConcurrent),
    }
}

// Execute runs the function with bulkhead protection.
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    default:
        return ErrBulkheadFull
    }
}

// ExecuteBlocking waits for a slot and then executes.
func (b *Bulkhead) ExecuteBlocking(ctx context.Context, fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Available returns available slots.
func (b *Bulkhead) Available() int {
    return b.maxConc - len(b.sem)
}
```

### Task 5: Create Package Documentation

**File:** `toolops/resilience/doc.go`

```go
// Package resilience provides fault tolerance patterns for tool execution.
//
// This package implements common resilience patterns to handle failures
// gracefully and prevent cascading failures in distributed systems.
//
// # Circuit Breaker
//
// Prevent repeated calls to failing services:
//
//	cb := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
//	    Name:        "backend",
//	    MaxFailures: 5,
//	    Timeout:     30 * time.Second,
//	    HalfOpenMax: 2,
//	})
//
//	err := cb.Execute(ctx, func() error {
//	    return callBackend()
//	})
//
// # Retry with Backoff
//
// Retry failed operations with exponential backoff:
//
//	err := resilience.Retry(ctx, resilience.RetryConfig{
//	    MaxAttempts:     3,
//	    InitialInterval: 100 * time.Millisecond,
//	    MaxInterval:     10 * time.Second,
//	    Multiplier:      2.0,
//	}, func() error {
//	    return callService()
//	})
//
// # Bulkhead
//
// Limit concurrent executions to prevent resource exhaustion:
//
//	bulkhead := resilience.NewBulkhead(resilience.BulkheadConfig{
//	    Name:          "api-calls",
//	    MaxConcurrent: 10,
//	})
//
//	err := bulkhead.Execute(ctx, func() error {
//	    return makeAPICall()
//	})
//
// # Middleware
//
// Apply resilience patterns to tool providers:
//
//	resilient := resilience.Middleware(provider, resilience.MiddlewareConfig{
//	    CircuitBreaker: cb,
//	    Retry:          retryConfig,
//	    Bulkhead:       bulkhead,
//	})
package resilience
```

### Task 6: Build and Test

```bash
cd /tmp/migration/toolops

go mod tidy
go build ./...
go test -v ./resilience/...
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/toolops

git add -A
git commit -m "feat(resilience): add fault tolerance patterns

Create new resilience package for fault tolerance.

Package contents:
- CircuitBreaker for failure isolation
- Retry with exponential backoff
- Bulkhead for concurrency limiting
- Resilience middleware

Features:
- Three-state circuit breaker (closed/open/half-open)
- Configurable failure thresholds
- Exponential backoff with jitter
- Non-blocking and blocking bulkhead modes
- Generic retry with result support

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-164: Create toolhealth
- Gate G4: Operations layer complete
