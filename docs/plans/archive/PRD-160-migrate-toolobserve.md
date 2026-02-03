# PRD-160: Migrate toolobserve

**Phase:** 6 - Operations Layer
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-120
**Status:** Done (2026-01-31)

---

## Objective

Migrate the existing `toolobserve` repository into `toolops/observe/` as the first package in the consolidated operations layer.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/toolobserve`
**Target Location:** `github.com/jonwraymond/toolops/observe`

**Package Contents:**
- OpenTelemetry integration for tracing
- Metrics collection (Prometheus)
- Structured logging (slog)
- ~2,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Observe Package | `toolops/observe/` | Observability infrastructure |
| Tests | `toolops/observe/*_test.go` | All existing tests |
| Documentation | `toolops/observe/doc.go` | Package documentation |

---

## Tasks

### Task 1: Prepare Target Repository

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolops.git
cd toolops

mkdir -p observe
```

### Task 2: Clone and Migrate

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolobserve.git

cp toolobserve/*.go toolops/observe/

cd toolops/observe
sed -i '' 's|github.com/jonwraymond/toolobserve|github.com/jonwraymond/toolops/observe|g' *.go
sed -i '' 's|github.com/jonwraymond/toolmodel|github.com/jonwraymond/toolfoundation/model|g' *.go
```

### Task 3: Update Package Documentation

**File:** `toolops/observe/doc.go`

```go
// Package observe provides observability infrastructure for the ApertureStack ecosystem.
//
// This package implements distributed tracing, metrics collection, and structured
// logging using industry-standard tools (OpenTelemetry, Prometheus, slog).
//
// # Tracing
//
// Create spans for tool execution:
//
//	tracer := observe.NewTracer(observe.TracerConfig{
//	    ServiceName: "metatools-mcp",
//	    Endpoint:    "http://jaeger:4317",
//	})
//
//	ctx, span := tracer.Start(ctx, "tool.execute",
//	    observe.WithToolID(tool.ID),
//	    observe.WithToolName(tool.Name),
//	)
//	defer span.End()
//
// # Metrics
//
// Collect metrics for monitoring:
//
//	metrics := observe.NewMetrics(observe.MetricsConfig{
//	    Namespace: "metatools",
//	})
//
//	metrics.ToolExecutions.Inc()
//	metrics.ToolLatency.Observe(duration.Seconds())
//
// # Logging
//
// Structured logging with context:
//
//	logger := observe.NewLogger(observe.LogConfig{
//	    Level:  slog.LevelInfo,
//	    Format: "json",
//	})
//
//	logger.Info("tool executed",
//	    slog.String("tool_id", tool.ID),
//	    slog.Duration("duration", elapsed),
//	)
//
// # Middleware
//
// Wrap tool providers with observability:
//
//	observed := observe.Middleware(provider, observe.MiddlewareConfig{
//	    Tracer:  tracer,
//	    Metrics: metrics,
//	    Logger:  logger,
//	})
//
// # Migration Note
//
// This package was migrated from github.com/jonwraymond/toolobserve as part of
// the ApertureStack consolidation.
package observe
```

### Task 4: Build and Test

```bash
cd /tmp/migration/toolops

go mod tidy
go build ./...
go test -v ./observe/...
```

### Task 5: Commit and Push

```bash
cd /tmp/migration/toolops

git add -A
git commit -m "feat(observe): migrate toolobserve package

Migrate observability infrastructure from standalone toolobserve repository.

Package contents:
- OpenTelemetry tracing integration
- Prometheus metrics collection
- Structured logging with slog
- Observability middleware

Features:
- Distributed tracing across tool executions
- Standard metrics (requests, latency, errors)
- JSON/text logging formats
- Context propagation

This is part of the ApertureStack consolidation effort.

Migration: github.com/jonwraymond/toolobserve → toolops/observe

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Tracing works
- [ ] Metrics collection works
- [ ] Logging works

## Completion Notes

- `toolops/observe` includes observer, tracer, metrics, logger, and middleware helpers.
- Imports updated to `github.com/jonwraymond/...`.

---

## Next Steps

- PRD-161: Migrate toolcache
- PRD-162: Extract toolauth
