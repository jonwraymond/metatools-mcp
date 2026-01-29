package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMetrics captures basic per-tool execution metrics.
type ToolMetrics struct {
	TotalRequests  int
	SuccessCount   int
	ErrorCount     int
	ActiveRequests int
	LastDuration   time.Duration
}

// MetricsCollector records tool execution metrics.
type MetricsCollector interface {
	Start(tool string)
	Finish(tool string, err error, duration time.Duration)
}

// MetricsConfig configures the metrics middleware.
type MetricsConfig struct {
	Collector MetricsCollector
}

// NewMetricsMiddleware creates a middleware that collects metrics.
func NewMetricsMiddleware(cfg MetricsConfig) Middleware {
	collector := cfg.Collector
	if collector == nil {
		collector = NewInMemoryMetricsCollector()
	}

	return func(next provider.ToolProvider) provider.ToolProvider {
		return &metricsProvider{
			next:      next,
			collector: collector,
		}
	}
}

// MetricsMiddlewareFactory creates a metrics middleware from config.
func MetricsMiddlewareFactory(_ map[string]any) (Middleware, error) {
	return NewMetricsMiddleware(MetricsConfig{}), nil
}

type metricsProvider struct {
	next      provider.ToolProvider
	collector MetricsCollector
}

func (m *metricsProvider) Name() string   { return m.next.Name() }
func (m *metricsProvider) Enabled() bool  { return m.next.Enabled() }
func (m *metricsProvider) Tool() mcp.Tool { return m.next.Tool() }

func (m *metricsProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	m.collector.Start(m.next.Name())
	res, out, err := m.next.Handle(ctx, req, args)
	m.collector.Finish(m.next.Name(), err, time.Since(start))
	return res, out, err
}

// InMemoryMetricsCollector stores metrics in memory for testing and local use.
type InMemoryMetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*ToolMetrics
}

// NewInMemoryMetricsCollector creates a new in-memory collector.
func NewInMemoryMetricsCollector() *InMemoryMetricsCollector {
	return &InMemoryMetricsCollector{metrics: make(map[string]*ToolMetrics)}
}

// Start records the start of a request.
func (c *InMemoryMetricsCollector) Start(tool string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := c.getOrCreate(tool)
	m.TotalRequests++
	m.ActiveRequests++
}

// Finish records the completion of a request.
func (c *InMemoryMetricsCollector) Finish(tool string, err error, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := c.getOrCreate(tool)
	if m.ActiveRequests > 0 {
		m.ActiveRequests--
	}
	m.LastDuration = duration
	if err != nil {
		m.ErrorCount++
	} else {
		m.SuccessCount++
	}
}

// GetMetrics returns metrics for a tool.
func (c *InMemoryMetricsCollector) GetMetrics(tool string) ToolMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.metrics[tool]; ok {
		return *m
	}
	return ToolMetrics{}
}

func (c *InMemoryMetricsCollector) getOrCreate(tool string) *ToolMetrics {
	m, ok := c.metrics[tool]
	if !ok {
		m = &ToolMetrics{}
		c.metrics[tool] = m
	}
	return m
}
