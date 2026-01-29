package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type loggingMockProvider struct {
	name     string
	enabled  bool
	tool     mcp.Tool
	handleFn func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
}

func (m *loggingMockProvider) Name() string   { return m.name }
func (m *loggingMockProvider) Enabled() bool  { return m.enabled }
func (m *loggingMockProvider) Tool() mcp.Tool { return m.tool }
func (m *loggingMockProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	if m.handleFn != nil {
		return m.handleFn(ctx, req, args)
	}
	return nil, nil, nil
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mw := NewLoggingMiddleware(LoggingConfig{Logger: logger})

	original := &loggingMockProvider{
		name:    "test_tool",
		enabled: true,
		tool:    mcp.Tool{Name: "test_tool", InputSchema: map[string]any{"type": "object"}},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			return nil, "result", nil
		},
	}

	var _ provider.ToolProvider = original

	wrapped := mw(original)
	_, _, err := wrapped.Handle(context.Background(), nil, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test_tool") {
		t.Errorf("Log output missing tool name: %s", output)
	}
}

func TestLoggingMiddleware_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mw := NewLoggingMiddleware(LoggingConfig{Logger: logger})

	original := &loggingMockProvider{
		name:    "failing_tool",
		enabled: true,
		tool:    mcp.Tool{Name: "failing_tool", InputSchema: map[string]any{"type": "object"}},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			return nil, nil, errors.New("tool failed")
		},
	}

	wrapped := mw(original)
	_, _, err := wrapped.Handle(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("Handle() should return error")
	}

	output := buf.String()
	if !strings.Contains(output, "error") || !strings.Contains(output, "tool failed") {
		t.Errorf("Log output missing error: %s", output)
	}
}
