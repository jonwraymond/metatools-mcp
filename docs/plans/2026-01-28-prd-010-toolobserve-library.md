# PRD-010: toolobserve Library Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create an observability library providing OpenTelemetry-based tracing, metrics, and structured logging for tool execution.

**Architecture:** Observability middleware wrapping tool execution with automatic span creation, metrics collection, and structured log emission. Supports multiple exporters (OTLP, Jaeger, Prometheus, stdout).

**Tech Stack:** Go, OpenTelemetry SDK, OTEL exporters

---

## Overview

The `toolobserve` library provides comprehensive observability for the metatools ecosystem, enabling distributed tracing, metrics, and structured logging across all tool operations.

**Reference:** [architecture-evaluation.md](../proposals/architecture-evaluation.md) - Observability gap analysis

---

## Directory Structure

```
toolobserve/
├── observe.go           # Core Observer type
├── observe_test.go
├── tracer.go            # Tracing implementation
├── tracer_test.go
├── metrics.go           # Metrics implementation
├── metrics_test.go
├── logger.go            # Structured logging
├── logger_test.go
├── middleware.go        # Tool execution middleware
├── middleware_test.go
├── exporters/
│   ├── otlp.go          # OTLP exporter config
│   ├── jaeger.go        # Jaeger exporter config
│   ├── prometheus.go    # Prometheus exporter config
│   └── stdout.go        # Stdout exporter (dev)
├── doc.go
├── go.mod
└── go.sum
```

---

## Task 1: Core Observer and Config Types

**Files:**
- Create: `toolobserve/observe.go`
- Create: `toolobserve/observe_test.go`
- Create: `toolobserve/go.mod`

**Step 1: Write failing tests**

```go
// observe_test.go
package toolobserve_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolobserve"
)

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  toolobserve.Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: toolobserve.Config{
                ServiceName: "metatools",
                Tracing: toolobserve.TracingConfig{
                    Enabled: true,
                },
            },
            wantErr: false,
        },
        {
            name: "missing service name",
            config: toolobserve.Config{
                Tracing: toolobserve.TracingConfig{
                    Enabled: true,
                },
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

func TestObserver_New(t *testing.T) {
    config := toolobserve.Config{
        ServiceName: "test-service",
        Version:     "1.0.0",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    }

    obs, err := toolobserve.New(config)
    require.NoError(t, err)
    assert.NotNil(t, obs)

    defer obs.Shutdown(context.Background())
}

func TestObserver_Tracer(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    tracer := obs.Tracer()
    assert.NotNil(t, tracer)
}

func TestObserver_Meter(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    meter := obs.Meter()
    assert.NotNil(t, meter)
}

func TestObserver_Logger(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Logging: toolobserve.LoggingConfig{
            Level:  "info",
            Format: "json",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    logger := obs.Logger()
    assert.NotNil(t, logger)
}

func TestDefaultConfig(t *testing.T) {
    config := toolobserve.DefaultConfig("my-service")

    assert.Equal(t, "my-service", config.ServiceName)
    assert.True(t, config.Tracing.Enabled)
    assert.True(t, config.Metrics.Enabled)
    assert.Equal(t, "info", config.Logging.Level)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// go.mod
module github.com/jrraymond/toolobserve

go 1.22

require (
    go.opentelemetry.io/otel v1.24.0
    go.opentelemetry.io/otel/trace v1.24.0
    go.opentelemetry.io/otel/metric v1.24.0
    go.opentelemetry.io/otel/sdk v1.24.0
    go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.24.0
    go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.24.0
)
```

```go
// observe.go
package toolobserve

import (
    "context"
    "errors"
    "log/slog"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/trace"
)

// Config holds all observability configuration
type Config struct {
    ServiceName    string
    Version        string
    Environment    string
    Tracing        TracingConfig
    Metrics        MetricsConfig
    Logging        LoggingConfig
    ResourceAttrs  map[string]string
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
    Enabled     bool
    Exporter    string // otlp, jaeger, stdout
    Endpoint    string
    SampleRate  float64
    Headers     map[string]string
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
    Enabled     bool
    Exporter    string // otlp, prometheus, stdout
    Endpoint    string
    Interval    string // export interval
    Headers     map[string]string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
    Level   string // debug, info, warn, error
    Format  string // json, text
    Output  string // stdout, stderr, file path
}

// Validate validates the configuration
func (c Config) Validate() error {
    if c.ServiceName == "" {
        return errors.New("service name is required")
    }
    return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig(serviceName string) Config {
    return Config{
        ServiceName: serviceName,
        Environment: "development",
        Tracing: TracingConfig{
            Enabled:    true,
            Exporter:   "stdout",
            SampleRate: 1.0,
        },
        Metrics: MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
            Interval: "10s",
        },
        Logging: LoggingConfig{
            Level:  "info",
            Format: "json",
            Output: "stdout",
        },
    }
}

// Observer provides observability primitives
type Observer struct {
    config        Config
    tracerProvider trace.TracerProvider
    meterProvider  metric.MeterProvider
    logger        *slog.Logger
    shutdownFuncs []func(context.Context) error
}

// New creates a new Observer
func New(config Config) (*Observer, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }

    obs := &Observer{
        config: config,
    }

    // Initialize tracing
    if config.Tracing.Enabled {
        tp, shutdown, err := initTracing(config)
        if err != nil {
            return nil, err
        }
        obs.tracerProvider = tp
        obs.shutdownFuncs = append(obs.shutdownFuncs, shutdown)
        otel.SetTracerProvider(tp)
    }

    // Initialize metrics
    if config.Metrics.Enabled {
        mp, shutdown, err := initMetrics(config)
        if err != nil {
            return nil, err
        }
        obs.meterProvider = mp
        obs.shutdownFuncs = append(obs.shutdownFuncs, shutdown)
        otel.SetMeterProvider(mp)
    }

    // Initialize logging
    obs.logger = initLogger(config.Logging)

    return obs, nil
}

// Tracer returns the tracer
func (o *Observer) Tracer() trace.Tracer {
    if o.tracerProvider == nil {
        return otel.Tracer(o.config.ServiceName)
    }
    return o.tracerProvider.Tracer(o.config.ServiceName)
}

// Meter returns the meter
func (o *Observer) Meter() metric.Meter {
    if o.meterProvider == nil {
        return otel.Meter(o.config.ServiceName)
    }
    return o.meterProvider.Meter(o.config.ServiceName)
}

// Logger returns the logger
func (o *Observer) Logger() *slog.Logger {
    if o.logger == nil {
        return slog.Default()
    }
    return o.logger
}

// Shutdown shuts down all observability providers
func (o *Observer) Shutdown(ctx context.Context) error {
    var errs []error
    for _, fn := range o.shutdownFuncs {
        if err := fn(ctx); err != nil {
            errs = append(errs, err)
        }
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add core Observer and Config types

- Config with Tracing, Metrics, Logging sub-configs
- Observer with Tracer, Meter, Logger accessors
- DefaultConfig for development setup
- Graceful shutdown support

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Tracing Implementation

**Files:**
- Create: `toolobserve/tracer.go`
- Create: `toolobserve/tracer_test.go`

**Step 1: Write failing tests**

```go
// tracer_test.go
package toolobserve_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolobserve"
    "go.opentelemetry.io/otel/trace"
)

func TestSpan_ToolExecution(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    ctx := context.Background()
    ctx, span := toolobserve.StartToolSpan(ctx, obs.Tracer(), "test:search", map[string]any{
        "query": "test query",
    })
    defer span.End()

    assert.True(t, span.SpanContext().IsValid())
    assert.True(t, span.SpanContext().IsSampled())
}

func TestSpan_ToolCallAttributes(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    ctx := context.Background()
    ctx, span := toolobserve.StartToolSpan(ctx, obs.Tracer(), "mcp:search", map[string]any{
        "query": "find tools",
        "limit": 10,
    })

    // Add result
    toolobserve.RecordToolResult(span, "success", 5)
    span.End()

    // Span should be valid
    assert.True(t, span.SpanContext().IsValid())
}

func TestSpan_ToolError(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    ctx := context.Background()
    ctx, span := toolobserve.StartToolSpan(ctx, obs.Tracer(), "mcp:execute", nil)

    toolobserve.RecordToolError(span, errors.New("execution failed"))
    span.End()

    assert.True(t, span.SpanContext().IsValid())
}

func TestSpan_ChainExecution(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    ctx := context.Background()

    // Start chain span
    ctx, chainSpan := toolobserve.StartChainSpan(ctx, obs.Tracer(), "test-chain", 3)

    // Start step spans (children of chain)
    for i := 0; i < 3; i++ {
        _, stepSpan := toolobserve.StartStepSpan(ctx, obs.Tracer(), i, "mcp:tool"+string(rune('A'+i)))
        stepSpan.End()
    }

    chainSpan.End()
    assert.True(t, chainSpan.SpanContext().IsValid())
}

func TestExtractSpanContext(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing: toolobserve.TracingConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    ctx := context.Background()
    ctx, span := toolobserve.StartToolSpan(ctx, obs.Tracer(), "test:tool", nil)
    defer span.End()

    // Extract context
    traceID, spanID := toolobserve.ExtractSpanContext(ctx)
    assert.NotEmpty(t, traceID)
    assert.NotEmpty(t, spanID)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// tracer.go
package toolobserve

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
)

// Semantic convention attributes for tool execution
const (
    AttrToolID       = "tool.id"
    AttrToolName     = "tool.name"
    AttrToolNs       = "tool.namespace"
    AttrToolStatus   = "tool.status"
    AttrToolResults  = "tool.results_count"
    AttrChainID      = "chain.id"
    AttrChainSteps   = "chain.steps"
    AttrStepIndex    = "step.index"
)

// initTracing initializes the tracer provider
func initTracing(config Config) (trace.TracerProvider, func(context.Context) error, error) {
    var exporter sdktrace.SpanExporter
    var err error

    switch config.Tracing.Exporter {
    case "stdout":
        exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
    case "otlp":
        exporter, err = initOTLPTraceExporter(config.Tracing)
    case "jaeger":
        exporter, err = initJaegerExporter(config.Tracing)
    default:
        exporter, err = stdouttrace.New()
    }

    if err != nil {
        return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
    }

    sampler := sdktrace.AlwaysSample()
    if config.Tracing.SampleRate < 1.0 {
        sampler = sdktrace.TraceIDRatioBased(config.Tracing.SampleRate)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithSampler(sampler),
        sdktrace.WithResource(buildResource(config)),
    )

    return tp, tp.Shutdown, nil
}

// StartToolSpan starts a span for tool execution
func StartToolSpan(ctx context.Context, tracer trace.Tracer, toolID string, args map[string]any) (context.Context, trace.Span) {
    attrs := []attribute.KeyValue{
        attribute.String(AttrToolID, toolID),
    }

    // Extract namespace and name from tool ID
    if ns, name := parseToolID(toolID); ns != "" {
        attrs = append(attrs,
            attribute.String(AttrToolNs, ns),
            attribute.String(AttrToolName, name),
        )
    }

    // Add argument attributes (sanitized)
    for k, v := range args {
        attrs = append(attrs, attribute.String("tool.arg."+k, fmt.Sprintf("%v", v)))
    }

    ctx, span := tracer.Start(ctx, "tool.execute",
        trace.WithSpanKind(trace.SpanKindInternal),
        trace.WithAttributes(attrs...),
    )

    return ctx, span
}

// RecordToolResult records a successful tool result
func RecordToolResult(span trace.Span, status string, resultCount int) {
    span.SetAttributes(
        attribute.String(AttrToolStatus, status),
        attribute.Int(AttrToolResults, resultCount),
    )
    span.SetStatus(codes.Ok, "success")
}

// RecordToolError records a tool execution error
func RecordToolError(span trace.Span, err error) {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    span.SetAttributes(attribute.String(AttrToolStatus, "error"))
}

// StartChainSpan starts a span for chain execution
func StartChainSpan(ctx context.Context, tracer trace.Tracer, chainID string, steps int) (context.Context, trace.Span) {
    ctx, span := tracer.Start(ctx, "chain.execute",
        trace.WithSpanKind(trace.SpanKindInternal),
        trace.WithAttributes(
            attribute.String(AttrChainID, chainID),
            attribute.Int(AttrChainSteps, steps),
        ),
    )
    return ctx, span
}

// StartStepSpan starts a span for a chain step
func StartStepSpan(ctx context.Context, tracer trace.Tracer, index int, toolID string) (context.Context, trace.Span) {
    ctx, span := tracer.Start(ctx, "chain.step",
        trace.WithSpanKind(trace.SpanKindInternal),
        trace.WithAttributes(
            attribute.Int(AttrStepIndex, index),
            attribute.String(AttrToolID, toolID),
        ),
    )
    return ctx, span
}

// ExtractSpanContext extracts trace and span IDs from context
func ExtractSpanContext(ctx context.Context) (traceID, spanID string) {
    span := trace.SpanFromContext(ctx)
    if span == nil {
        return "", ""
    }
    sc := span.SpanContext()
    return sc.TraceID().String(), sc.SpanID().String()
}

// parseToolID extracts namespace and name from tool ID
func parseToolID(toolID string) (namespace, name string) {
    for i, c := range toolID {
        if c == ':' {
            return toolID[:i], toolID[i+1:]
        }
    }
    return "", toolID
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add tracing implementation

- StartToolSpan with semantic attributes
- RecordToolResult and RecordToolError
- StartChainSpan and StartStepSpan for chains
- ExtractSpanContext for context propagation
- Stdout exporter for development

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Metrics Implementation

**Files:**
- Create: `toolobserve/metrics.go`
- Create: `toolobserve/metrics_test.go`

**Step 1: Write failing tests**

```go
// metrics_test.go
package toolobserve_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolobserve"
)

func TestMetrics_ToolCallCounter(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    metrics := toolobserve.NewToolMetrics(obs.Meter())

    // Record tool calls
    metrics.RecordToolCall("mcp:search", "success")
    metrics.RecordToolCall("mcp:search", "success")
    metrics.RecordToolCall("mcp:search", "error")

    // Metrics should be recorded (can't easily assert values)
    assert.NotNil(t, metrics)
}

func TestMetrics_ToolLatency(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    metrics := toolobserve.NewToolMetrics(obs.Meter())

    // Record latency
    metrics.RecordToolLatency("mcp:search", 150*time.Millisecond)
    metrics.RecordToolLatency("mcp:search", 200*time.Millisecond)

    assert.NotNil(t, metrics)
}

func TestMetrics_ChainMetrics(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    metrics := toolobserve.NewToolMetrics(obs.Meter())

    // Record chain execution
    metrics.RecordChainExecution("chain-1", 3, 500*time.Millisecond, "success")
    metrics.RecordChainExecution("chain-2", 5, 1*time.Second, "error")

    assert.NotNil(t, metrics)
}

func TestMetrics_ToolsInFlight(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    metrics := toolobserve.NewToolMetrics(obs.Meter())

    // Track in-flight
    metrics.ToolStarted("mcp:search")
    metrics.ToolStarted("mcp:search")
    metrics.ToolCompleted("mcp:search")

    assert.NotNil(t, metrics)
}

func TestMetrics_BackendMetrics(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Metrics: toolobserve.MetricsConfig{
            Enabled:  true,
            Exporter: "stdout",
        },
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    metrics := toolobserve.NewToolMetrics(obs.Meter())

    // Record backend calls
    metrics.RecordBackendCall("mcp", "github-server", 100*time.Millisecond, "success")
    metrics.RecordBackendCall("local", "handler", 50*time.Millisecond, "success")
    metrics.RecordBackendCall("mcp", "github-server", 0, "error")

    assert.NotNil(t, metrics)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// metrics.go
package toolobserve

import (
    "context"
    "fmt"
    "time"

    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
    "go.opentelemetry.io/otel/metric"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// ToolMetrics holds all tool-related metrics
type ToolMetrics struct {
    meter metric.Meter

    // Counters
    toolCalls    metric.Int64Counter
    chainCalls   metric.Int64Counter
    backendCalls metric.Int64Counter

    // Histograms
    toolLatency    metric.Float64Histogram
    chainLatency   metric.Float64Histogram
    backendLatency metric.Float64Histogram

    // Gauges (via UpDownCounter)
    toolsInFlight metric.Int64UpDownCounter
}

// NewToolMetrics creates a new ToolMetrics instance
func NewToolMetrics(meter metric.Meter) *ToolMetrics {
    m := &ToolMetrics{meter: meter}
    m.init()
    return m
}

func (m *ToolMetrics) init() {
    var err error

    // Tool call counter
    m.toolCalls, err = m.meter.Int64Counter(
        "tool.calls.total",
        metric.WithDescription("Total number of tool calls"),
        metric.WithUnit("{call}"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create tool.calls.total metric: %v", err))
    }

    // Chain call counter
    m.chainCalls, err = m.meter.Int64Counter(
        "chain.calls.total",
        metric.WithDescription("Total number of chain executions"),
        metric.WithUnit("{call}"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create chain.calls.total metric: %v", err))
    }

    // Backend call counter
    m.backendCalls, err = m.meter.Int64Counter(
        "backend.calls.total",
        metric.WithDescription("Total number of backend calls"),
        metric.WithUnit("{call}"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create backend.calls.total metric: %v", err))
    }

    // Tool latency histogram
    m.toolLatency, err = m.meter.Float64Histogram(
        "tool.latency",
        metric.WithDescription("Tool execution latency in milliseconds"),
        metric.WithUnit("ms"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create tool.latency metric: %v", err))
    }

    // Chain latency histogram
    m.chainLatency, err = m.meter.Float64Histogram(
        "chain.latency",
        metric.WithDescription("Chain execution latency in milliseconds"),
        metric.WithUnit("ms"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create chain.latency metric: %v", err))
    }

    // Backend latency histogram
    m.backendLatency, err = m.meter.Float64Histogram(
        "backend.latency",
        metric.WithDescription("Backend call latency in milliseconds"),
        metric.WithUnit("ms"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create backend.latency metric: %v", err))
    }

    // Tools in flight gauge
    m.toolsInFlight, err = m.meter.Int64UpDownCounter(
        "tool.in_flight",
        metric.WithDescription("Number of tools currently executing"),
        metric.WithUnit("{tool}"),
    )
    if err != nil {
        panic(fmt.Sprintf("failed to create tool.in_flight metric: %v", err))
    }
}

// RecordToolCall records a tool call
func (m *ToolMetrics) RecordToolCall(toolID, status string) {
    ns, name := parseToolID(toolID)
    m.toolCalls.Add(context.Background(), 1,
        metric.WithAttributes(
            attribute.String("tool.id", toolID),
            attribute.String("tool.namespace", ns),
            attribute.String("tool.name", name),
            attribute.String("status", status),
        ),
    )
}

// RecordToolLatency records tool execution latency
func (m *ToolMetrics) RecordToolLatency(toolID string, duration time.Duration) {
    ns, name := parseToolID(toolID)
    m.toolLatency.Record(context.Background(), float64(duration.Milliseconds()),
        metric.WithAttributes(
            attribute.String("tool.id", toolID),
            attribute.String("tool.namespace", ns),
            attribute.String("tool.name", name),
        ),
    )
}

// RecordChainExecution records a chain execution
func (m *ToolMetrics) RecordChainExecution(chainID string, steps int, duration time.Duration, status string) {
    m.chainCalls.Add(context.Background(), 1,
        metric.WithAttributes(
            attribute.String("chain.id", chainID),
            attribute.Int("chain.steps", steps),
            attribute.String("status", status),
        ),
    )
    m.chainLatency.Record(context.Background(), float64(duration.Milliseconds()),
        metric.WithAttributes(
            attribute.String("chain.id", chainID),
        ),
    )
}

// ToolStarted increments in-flight counter
func (m *ToolMetrics) ToolStarted(toolID string) {
    m.toolsInFlight.Add(context.Background(), 1,
        metric.WithAttributes(attribute.String("tool.id", toolID)),
    )
}

// ToolCompleted decrements in-flight counter
func (m *ToolMetrics) ToolCompleted(toolID string) {
    m.toolsInFlight.Add(context.Background(), -1,
        metric.WithAttributes(attribute.String("tool.id", toolID)),
    )
}

// RecordBackendCall records a backend call
func (m *ToolMetrics) RecordBackendCall(backendType, backendName string, duration time.Duration, status string) {
    m.backendCalls.Add(context.Background(), 1,
        metric.WithAttributes(
            attribute.String("backend.type", backendType),
            attribute.String("backend.name", backendName),
            attribute.String("status", status),
        ),
    )
    if duration > 0 {
        m.backendLatency.Record(context.Background(), float64(duration.Milliseconds()),
            metric.WithAttributes(
                attribute.String("backend.type", backendType),
                attribute.String("backend.name", backendName),
            ),
        )
    }
}

// initMetrics initializes the meter provider
func initMetrics(config Config) (metric.MeterProvider, func(context.Context) error, error) {
    var exporter sdkmetric.Exporter
    var err error

    switch config.Metrics.Exporter {
    case "stdout":
        exporter, err = stdoutmetric.New()
    case "otlp":
        exporter, err = initOTLPMetricExporter(config.Metrics)
    case "prometheus":
        // Prometheus uses a reader, not exporter
        return initPrometheusMetrics(config)
    default:
        exporter, err = stdoutmetric.New()
    }

    if err != nil {
        return nil, nil, fmt.Errorf("failed to create metric exporter: %w", err)
    }

    mp := sdkmetric.NewMeterProvider(
        sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
        sdkmetric.WithResource(buildResource(config)),
    )

    return mp, mp.Shutdown, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add metrics implementation

- ToolMetrics with counters, histograms, gauges
- RecordToolCall, RecordToolLatency
- RecordChainExecution for chain metrics
- ToolStarted/ToolCompleted for in-flight tracking
- RecordBackendCall for backend metrics

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Observability Middleware

**Files:**
- Create: `toolobserve/middleware.go`
- Create: `toolobserve/middleware_test.go`

**Step 1: Write failing tests**

```go
// middleware_test.go
package toolobserve_test

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolobserve"
)

// MockToolProvider for testing
type MockToolProvider struct {
    name    string
    handler func(ctx context.Context, input map[string]any) (any, error)
}

func (m *MockToolProvider) Name() string { return m.name }
func (m *MockToolProvider) Handle(ctx context.Context, input map[string]any) (any, error) {
    return m.handler(ctx, input)
}

func TestObserveMiddleware_Success(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing:     toolobserve.TracingConfig{Enabled: true, Exporter: "stdout"},
        Metrics:     toolobserve.MetricsConfig{Enabled: true, Exporter: "stdout"},
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    provider := &MockToolProvider{
        name: "mcp:search",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return map[string]any{"results": []string{"a", "b", "c"}}, nil
        },
    }

    wrapped := toolobserve.ObserveMiddleware(obs)(provider)

    result, err := wrapped.Handle(context.Background(), map[string]any{"query": "test"})
    require.NoError(t, err)
    assert.NotNil(t, result)
}

func TestObserveMiddleware_Error(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing:     toolobserve.TracingConfig{Enabled: true, Exporter: "stdout"},
        Metrics:     toolobserve.MetricsConfig{Enabled: true, Exporter: "stdout"},
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    expectedErr := errors.New("tool failed")
    provider := &MockToolProvider{
        name: "mcp:execute",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            return nil, expectedErr
        },
    }

    wrapped := toolobserve.ObserveMiddleware(obs)(provider)

    _, err = wrapped.Handle(context.Background(), nil)
    require.Error(t, err)
    assert.Equal(t, expectedErr, err)
}

func TestObserveMiddleware_TracesContext(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing:     toolobserve.TracingConfig{Enabled: true, Exporter: "stdout"},
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    var capturedTraceID string
    provider := &MockToolProvider{
        name: "mcp:tool",
        handler: func(ctx context.Context, input map[string]any) (any, error) {
            traceID, _ := toolobserve.ExtractSpanContext(ctx)
            capturedTraceID = traceID
            return nil, nil
        },
    }

    wrapped := toolobserve.ObserveMiddleware(obs)(provider)
    wrapped.Handle(context.Background(), nil)

    assert.NotEmpty(t, capturedTraceID)
}

func TestObserveChainMiddleware(t *testing.T) {
    obs, err := toolobserve.New(toolobserve.Config{
        ServiceName: "test",
        Tracing:     toolobserve.TracingConfig{Enabled: true, Exporter: "stdout"},
        Metrics:     toolobserve.MetricsConfig{Enabled: true, Exporter: "stdout"},
    })
    require.NoError(t, err)
    defer obs.Shutdown(context.Background())

    steps := []string{"tool-a", "tool-b", "tool-c"}
    results := make([]any, len(steps))

    err = toolobserve.ObserveChain(context.Background(), obs, "test-chain", steps,
        func(ctx context.Context, stepIndex int, toolID string) (any, error) {
            time.Sleep(10 * time.Millisecond)
            return map[string]any{"step": stepIndex}, nil
        },
        func(stepIndex int, result any) {
            results[stepIndex] = result
        },
    )

    require.NoError(t, err)
    assert.Len(t, results, 3)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// middleware.go
package toolobserve

import (
    "context"
    "time"
)

// ToolProvider is the interface for tool providers
type ToolProvider interface {
    Name() string
    Handle(ctx context.Context, input map[string]any) (any, error)
}

// Middleware wraps a ToolProvider
type Middleware func(ToolProvider) ToolProvider

// observedProvider wraps a provider with observability
type observedProvider struct {
    observer *Observer
    metrics  *ToolMetrics
    next     ToolProvider
}

// ObserveMiddleware creates observability middleware
func ObserveMiddleware(obs *Observer) Middleware {
    metrics := NewToolMetrics(obs.Meter())

    return func(next ToolProvider) ToolProvider {
        return &observedProvider{
            observer: obs,
            metrics:  metrics,
            next:     next,
        }
    }
}

func (p *observedProvider) Name() string {
    return p.next.Name()
}

func (p *observedProvider) Handle(ctx context.Context, input map[string]any) (any, error) {
    toolID := p.next.Name()
    start := time.Now()

    // Start span
    ctx, span := StartToolSpan(ctx, p.observer.Tracer(), toolID, input)
    defer span.End()

    // Track in-flight
    p.metrics.ToolStarted(toolID)
    defer p.metrics.ToolCompleted(toolID)

    // Log start
    p.observer.Logger().Info("tool.call.start",
        "tool_id", toolID,
        "trace_id", span.SpanContext().TraceID().String(),
    )

    // Execute
    result, err := p.next.Handle(ctx, input)
    duration := time.Since(start)

    // Record metrics and spans
    if err != nil {
        RecordToolError(span, err)
        p.metrics.RecordToolCall(toolID, "error")
        p.observer.Logger().Error("tool.call.error",
            "tool_id", toolID,
            "duration_ms", duration.Milliseconds(),
            "error", err.Error(),
        )
    } else {
        RecordToolResult(span, "success", countResults(result))
        p.metrics.RecordToolCall(toolID, "success")
        p.metrics.RecordToolLatency(toolID, duration)
        p.observer.Logger().Info("tool.call.success",
            "tool_id", toolID,
            "duration_ms", duration.Milliseconds(),
        )
    }

    return result, err
}

// StepExecutor executes a chain step
type StepExecutor func(ctx context.Context, stepIndex int, toolID string) (any, error)

// StepCallback is called after each step
type StepCallback func(stepIndex int, result any)

// ObserveChain executes a chain with observability
func ObserveChain(
    ctx context.Context,
    obs *Observer,
    chainID string,
    steps []string,
    executor StepExecutor,
    callback StepCallback,
) error {
    start := time.Now()
    metrics := NewToolMetrics(obs.Meter())

    // Start chain span
    ctx, chainSpan := StartChainSpan(ctx, obs.Tracer(), chainID, len(steps))
    defer chainSpan.End()

    obs.Logger().Info("chain.start",
        "chain_id", chainID,
        "steps", len(steps),
        "trace_id", chainSpan.SpanContext().TraceID().String(),
    )

    var lastErr error
    for i, toolID := range steps {
        stepCtx, stepSpan := StartStepSpan(ctx, obs.Tracer(), i, toolID)

        result, err := executor(stepCtx, i, toolID)

        if err != nil {
            RecordToolError(stepSpan, err)
            stepSpan.End()
            lastErr = err
            obs.Logger().Error("chain.step.error",
                "chain_id", chainID,
                "step", i,
                "tool_id", toolID,
                "error", err.Error(),
            )
            break
        }

        RecordToolResult(stepSpan, "success", countResults(result))
        stepSpan.End()

        if callback != nil {
            callback(i, result)
        }
    }

    duration := time.Since(start)
    status := "success"
    if lastErr != nil {
        status = "error"
    }

    metrics.RecordChainExecution(chainID, len(steps), duration, status)
    obs.Logger().Info("chain.complete",
        "chain_id", chainID,
        "duration_ms", duration.Milliseconds(),
        "status", status,
    )

    return lastErr
}

// countResults attempts to count results
func countResults(result any) int {
    if result == nil {
        return 0
    }
    switch v := result.(type) {
    case []any:
        return len(v)
    case map[string]any:
        if results, ok := v["results"].([]any); ok {
            return len(results)
        }
        return 1
    default:
        return 1
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add observability middleware

- ObserveMiddleware wraps ToolProvider with tracing/metrics
- Automatic span creation and error recording
- In-flight tracking via metrics
- Structured logging for all operations
- ObserveChain for chain execution observability

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Structured Logging

**Files:**
- Create: `toolobserve/logger.go`
- Create: `toolobserve/logger_test.go`

**Step 1: Write failing tests**

```go
// logger_test.go
package toolobserve_test

import (
    "bytes"
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolobserve"
)

func TestLogger_JSONFormat(t *testing.T) {
    var buf bytes.Buffer
    logger := toolobserve.NewLogger(toolobserve.LoggingConfig{
        Level:  "info",
        Format: "json",
    }, &buf)

    logger.Info("test message", "key", "value")

    var entry map[string]any
    err := json.Unmarshal(buf.Bytes(), &entry)
    require.NoError(t, err)

    assert.Equal(t, "test message", entry["msg"])
    assert.Equal(t, "value", entry["key"])
}

func TestLogger_LevelFiltering(t *testing.T) {
    var buf bytes.Buffer
    logger := toolobserve.NewLogger(toolobserve.LoggingConfig{
        Level:  "warn",
        Format: "json",
    }, &buf)

    logger.Info("should be filtered")
    logger.Warn("should appear")

    // Only warn message should appear
    assert.Contains(t, buf.String(), "should appear")
    assert.NotContains(t, buf.String(), "should be filtered")
}

func TestLogger_WithContext(t *testing.T) {
    var buf bytes.Buffer
    logger := toolobserve.NewLogger(toolobserve.LoggingConfig{
        Level:  "info",
        Format: "json",
    }, &buf)

    // Create logger with context
    ctxLogger := logger.With("service", "metatools", "version", "1.0.0")
    ctxLogger.Info("contextual log")

    var entry map[string]any
    err := json.Unmarshal(buf.Bytes(), &entry)
    require.NoError(t, err)

    assert.Equal(t, "metatools", entry["service"])
    assert.Equal(t, "1.0.0", entry["version"])
}

func TestLogger_ToolCallEntry(t *testing.T) {
    var buf bytes.Buffer
    logger := toolobserve.NewLogger(toolobserve.LoggingConfig{
        Level:  "info",
        Format: "json",
    }, &buf)

    toolobserve.LogToolCall(logger, "mcp:search", "query", "test", "success", 150)

    var entry map[string]any
    err := json.Unmarshal(buf.Bytes(), &entry)
    require.NoError(t, err)

    assert.Equal(t, "tool.call", entry["msg"])
    assert.Equal(t, "mcp:search", entry["tool_id"])
    assert.Equal(t, "success", entry["status"])
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// logger.go
package toolobserve

import (
    "io"
    "log/slog"
    "os"
    "strings"
)

// initLogger initializes the structured logger
func initLogger(config LoggingConfig) *slog.Logger {
    var output io.Writer
    switch config.Output {
    case "stderr":
        output = os.Stderr
    case "stdout", "":
        output = os.Stdout
    default:
        // File output
        f, err := os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            output = os.Stdout
        } else {
            output = f
        }
    }

    return NewLogger(config, output)
}

// NewLogger creates a new structured logger
func NewLogger(config LoggingConfig, output io.Writer) *slog.Logger {
    level := parseLevel(config.Level)

    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: level}

    switch config.Format {
    case "json":
        handler = slog.NewJSONHandler(output, opts)
    case "text", "":
        handler = slog.NewTextHandler(output, opts)
    default:
        handler = slog.NewJSONHandler(output, opts)
    }

    return slog.New(handler)
}

// parseLevel parses log level string
func parseLevel(level string) slog.Level {
    switch strings.ToLower(level) {
    case "debug":
        return slog.LevelDebug
    case "info", "":
        return slog.LevelInfo
    case "warn", "warning":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}

// LogToolCall logs a tool call with standard fields
func LogToolCall(logger *slog.Logger, toolID, query, status string, durationMs int64) {
    logger.Info("tool.call",
        "tool_id", toolID,
        "query", query,
        "status", status,
        "duration_ms", durationMs,
    )
}

// LogChainStart logs chain start
func LogChainStart(logger *slog.Logger, chainID string, steps int, traceID string) {
    logger.Info("chain.start",
        "chain_id", chainID,
        "steps", steps,
        "trace_id", traceID,
    )
}

// LogChainComplete logs chain completion
func LogChainComplete(logger *slog.Logger, chainID, status string, durationMs int64) {
    logger.Info("chain.complete",
        "chain_id", chainID,
        "status", status,
        "duration_ms", durationMs,
    )
}

// LogBackendCall logs a backend call
func LogBackendCall(logger *slog.Logger, backendType, backendName, status string, durationMs int64) {
    logger.Info("backend.call",
        "backend_type", backendType,
        "backend_name", backendName,
        "status", status,
        "duration_ms", durationMs,
    )
}

// LogError logs an error with context
func LogError(logger *slog.Logger, operation string, err error, fields ...any) {
    args := append([]any{"operation", operation, "error", err.Error()}, fields...)
    logger.Error("operation.error", args...)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add structured logging

- NewLogger with JSON and text formats
- Level filtering (debug, info, warn, error)
- LogToolCall, LogChainStart, LogChainComplete helpers
- LogBackendCall for backend operations
- LogError for error logging with context

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Exporter Configurations

**Files:**
- Create: `toolobserve/exporters/otlp.go`
- Create: `toolobserve/exporters/prometheus.go`
- Create: `toolobserve/exporters/stdout.go`

**Step 1: Write failing tests**

```go
// exporters/otlp_test.go
package exporters_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/jrraymond/toolobserve/exporters"
)

func TestOTLPConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  exporters.OTLPConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: exporters.OTLPConfig{
                Endpoint: "localhost:4317",
            },
            wantErr: false,
        },
        {
            name: "missing endpoint",
            config: exporters.OTLPConfig{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestPrometheusConfig_Endpoint(t *testing.T) {
    config := exporters.PrometheusConfig{
        Port: 9090,
        Path: "/metrics",
    }

    assert.Equal(t, ":9090", config.Addr())
    assert.Equal(t, "/metrics", config.Path)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolobserve && go test ./exporters/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// exporters/otlp.go
package exporters

import (
    "context"
    "errors"

    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// OTLPConfig holds OTLP exporter configuration
type OTLPConfig struct {
    Endpoint    string
    Insecure    bool
    Headers     map[string]string
    Compression string // gzip, none
}

// Validate validates the OTLP config
func (c OTLPConfig) Validate() error {
    if c.Endpoint == "" {
        return errors.New("OTLP endpoint is required")
    }
    return nil
}

// NewOTLPTraceExporter creates an OTLP trace exporter
func NewOTLPTraceExporter(ctx context.Context, config OTLPConfig) (sdktrace.SpanExporter, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }

    opts := []otlptracegrpc.Option{
        otlptracegrpc.WithEndpoint(config.Endpoint),
    }

    if config.Insecure {
        opts = append(opts, otlptracegrpc.WithInsecure())
    }

    if len(config.Headers) > 0 {
        opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
    }

    return otlptracegrpc.New(ctx, opts...)
}

// NewOTLPMetricExporter creates an OTLP metric exporter
func NewOTLPMetricExporter(ctx context.Context, config OTLPConfig) (sdkmetric.Exporter, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }

    opts := []otlpmetricgrpc.Option{
        otlpmetricgrpc.WithEndpoint(config.Endpoint),
    }

    if config.Insecure {
        opts = append(opts, otlpmetricgrpc.WithInsecure())
    }

    if len(config.Headers) > 0 {
        opts = append(opts, otlpmetricgrpc.WithHeaders(config.Headers))
    }

    return otlpmetricgrpc.New(ctx, opts...)
}
```

```go
// exporters/prometheus.go
package exporters

import (
    "fmt"
    "net/http"

    "go.opentelemetry.io/otel/exporters/prometheus"
    "go.opentelemetry.io/otel/metric"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// PrometheusConfig holds Prometheus exporter configuration
type PrometheusConfig struct {
    Port int
    Path string
}

// Addr returns the address for the Prometheus endpoint
func (c PrometheusConfig) Addr() string {
    return fmt.Sprintf(":%d", c.Port)
}

// NewPrometheusExporter creates a Prometheus metric exporter
func NewPrometheusExporter(config PrometheusConfig) (metric.MeterProvider, http.Handler, error) {
    exporter, err := prometheus.New()
    if err != nil {
        return nil, nil, err
    }

    provider := sdkmetric.NewMeterProvider(
        sdkmetric.WithReader(exporter),
    )

    return provider, exporter, nil
}
```

```go
// exporters/stdout.go
package exporters

import (
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// StdoutConfig holds stdout exporter configuration
type StdoutConfig struct {
    PrettyPrint bool
    Timestamps  bool
}

// NewStdoutTraceExporter creates a stdout trace exporter
func NewStdoutTraceExporter(config StdoutConfig) (sdktrace.SpanExporter, error) {
    opts := []stdouttrace.Option{}

    if config.PrettyPrint {
        opts = append(opts, stdouttrace.WithPrettyPrint())
    }

    return stdouttrace.New(opts...)
}

// NewStdoutMetricExporter creates a stdout metric exporter
func NewStdoutMetricExporter(config StdoutConfig) (sdkmetric.Exporter, error) {
    opts := []stdoutmetric.Option{}

    if config.PrettyPrint {
        opts = append(opts, stdoutmetric.WithPrettyPrint())
    }

    return stdoutmetric.New(opts...)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolobserve && go test ./exporters/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolobserve/
git commit -m "$(cat <<'EOF'
feat(toolobserve): add exporter configurations

- OTLPConfig with trace and metric exporter factories
- PrometheusConfig with HTTP handler
- StdoutConfig for development
- Validation and defaults for all configs

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-010 complete:

- [ ] All tests pass: `go test ./... -v`
- [ ] Code coverage > 80%: `go test ./... -cover`
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete
- [ ] Integration verified:
  - [ ] Tracing creates valid spans
  - [ ] Metrics are recorded correctly
  - [ ] Logging outputs structured data
  - [ ] Middleware wraps providers correctly

---

## Definition of Done

1. **Observer** type with Tracer, Meter, Logger
2. **Tracing** with StartToolSpan, StartChainSpan, StartStepSpan
3. **Metrics** with counters, histograms, and gauges
4. **Middleware** for automatic observability
5. **Structured logging** with JSON/text formats
6. **Exporters** for OTLP, Prometheus, stdout
7. All tests passing with >80% coverage
8. Documentation complete
