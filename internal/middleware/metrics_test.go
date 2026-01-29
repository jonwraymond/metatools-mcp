package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type metricsMockProvider struct {
	name     string
	enabled  bool
	tool     mcp.Tool
	handleFn func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
}

func (m *metricsMockProvider) Name() string   { return m.name }
func (m *metricsMockProvider) Enabled() bool  { return m.enabled }
func (m *metricsMockProvider) Tool() mcp.Tool { return m.tool }
func (m *metricsMockProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	if m.handleFn != nil {
		return m.handleFn(ctx, req, args)
	}
	return nil, nil, nil
}

func TestMetricsMiddleware(t *testing.T) {
	collector := NewInMemoryMetricsCollector()
	mw := NewMetricsMiddleware(MetricsConfig{Collector: collector})

	original := &metricsMockProvider{
		name:    "test_tool",
		enabled: true,
		tool:    mcp.Tool{Name: "test_tool", InputSchema: map[string]any{"type": "object"}},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, "result", nil
		},
	}

	wrapped := mw(original)
	_, _, err := wrapped.Handle(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	metrics := collector.GetMetrics("test_tool")
	if metrics.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", metrics.TotalRequests)
	}
	if metrics.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", metrics.SuccessCount)
	}
	if metrics.ErrorCount != 0 {
		t.Errorf("ErrorCount = %d, want 0", metrics.ErrorCount)
	}
	if metrics.LastDuration < 10*time.Millisecond {
		t.Errorf("LastDuration = %v, want >= 10ms", metrics.LastDuration)
	}
}

func TestMetricsMiddleware_Error(t *testing.T) {
	collector := NewInMemoryMetricsCollector()
	mw := NewMetricsMiddleware(MetricsConfig{Collector: collector})

	original := &metricsMockProvider{
		name:    "failing_tool",
		enabled: true,
		tool:    mcp.Tool{Name: "failing_tool", InputSchema: map[string]any{"type": "object"}},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			return nil, nil, errors.New("failed")
		},
	}

	wrapped := mw(original)
	_, _, _ = wrapped.Handle(context.Background(), nil, nil)

	metrics := collector.GetMetrics("failing_tool")
	if metrics.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", metrics.ErrorCount)
	}
	if metrics.SuccessCount != 0 {
		t.Errorf("SuccessCount = %d, want 0", metrics.SuccessCount)
	}
}

func TestMetricsMiddleware_ActiveRequests(t *testing.T) {
	collector := NewInMemoryMetricsCollector()
	mw := NewMetricsMiddleware(MetricsConfig{Collector: collector})

	started := make(chan struct{})
	done := make(chan struct{})

	original := &metricsMockProvider{
		name:    "slow_tool",
		enabled: true,
		tool:    mcp.Tool{Name: "slow_tool", InputSchema: map[string]any{"type": "object"}},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			close(started)
			<-done
			return nil, nil, nil
		},
	}

	wrapped := mw(original)
	go func() {
		_, _, _ = wrapped.Handle(context.Background(), nil, nil)
	}()

	<-started
	metrics := collector.GetMetrics("slow_tool")
	if metrics.ActiveRequests != 1 {
		t.Fatalf("ActiveRequests = %d, want 1", metrics.ActiveRequests)
	}

	close(done)
	time.Sleep(10 * time.Millisecond)

	metrics = collector.GetMetrics("slow_tool")
	if metrics.ActiveRequests != 0 {
		t.Fatalf("ActiveRequests = %d, want 0", metrics.ActiveRequests)
	}
}
