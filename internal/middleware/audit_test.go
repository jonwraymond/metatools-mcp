package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuditLogger captures audit entries for testing
type mockAuditLogger struct {
	mu      sync.Mutex
	entries []AuditEntry
}

func (m *mockAuditLogger) Log(_ context.Context, entry AuditEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
}

func (m *mockAuditLogger) Entries() []AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]AuditEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

func TestAuditMiddleware_LogsToolCall(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "test-tool", entry.Tool)
	assert.True(t, entry.Success)
	assert.Empty(t, entry.ErrorMsg)
}

func TestAuditMiddleware_LogsIdentity(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
		IdentityExtractor: func(ctx context.Context) AuditIdentity {
			return AuditIdentity{
				Principal: "user@example.com",
				TenantID:  "tenant-123",
				Roles:     []string{"admin", "user"},
			}
		},
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "user@example.com", entry.Principal)
	assert.Equal(t, "tenant-123", entry.TenantID)
	assert.Equal(t, []string{"admin", "user"}, entry.Roles)
}

func TestAuditMiddleware_LogsDuration(t *testing.T) {
	mock := &mockProvider{
		name: "slow-tool",
		handleFunc: func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
			time.Sleep(50 * time.Millisecond)
			return &mcp.CallToolResult{}, nil, nil
		},
	}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.GreaterOrEqual(t, entry.Duration, 50*time.Millisecond)
	assert.Less(t, entry.Duration, 200*time.Millisecond)
}

func TestAuditMiddleware_LogsSuccess(t *testing.T) {
	mock := &mockProvider{name: "success-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.True(t, entry.Success)
	assert.Empty(t, entry.ErrorMsg)
}

func TestAuditMiddleware_LogsFailure(t *testing.T) {
	expectedErr := errors.New("tool execution failed")
	mock := &mockProvider{
		name: "failing-tool",
		handleFunc: func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
			return nil, nil, expectedErr
		},
	}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.Error(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.False(t, entry.Success)
	assert.Contains(t, entry.ErrorMsg, "tool execution failed")
}

func TestAuditMiddleware_LogsIsErrorResult(t *testing.T) {
	mock := &mockProvider{
		name: "error-result-tool",
		handleFunc: func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "validation failed"},
				},
			}, nil, nil
		},
	}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.False(t, entry.Success)
}

func TestAuditMiddleware_IncludeArgs(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
		IncludeArgs: true,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := map[string]any{
		"param1": "value1",
		"param2": 42,
	}

	_, _, err := wrapped.Handle(ctx, req, args)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	require.NotNil(t, entry.Args)
	assert.Equal(t, "value1", entry.Args["param1"])
	assert.Equal(t, 42, entry.Args["param2"])
}

func TestAuditMiddleware_ExcludesArgsByDefault(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
		// IncludeArgs: false (default)
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	args := map[string]any{"secret": "password123"}

	_, _, err := wrapped.Handle(ctx, req, args)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Nil(t, entry.Args)
}

func TestAuditMiddleware_CustomAuditLogger(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	var capturedEntry AuditEntry
	customLogger := AuditLoggerFunc(func(ctx context.Context, entry AuditEntry) {
		capturedEntry = entry
	})

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: customLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	assert.Equal(t, "test-tool", capturedEntry.Tool)
	assert.True(t, capturedEntry.Success)
}

func TestAuditMiddleware_RequestIDCorrelation(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
		RequestIDExtractor: func(ctx context.Context) string {
			return "req-12345"
		},
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "req-12345", entry.RequestID)
}

func TestAuditMiddleware_DefaultSlogLogger(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	// Capture slog output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	mw := NewAuditLoggingMiddleware(AuditConfig{
		Logger: logger,
		// No AuditLogger - should use slog
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	// Parse the JSON log output
	var logEntry map[string]any
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test-tool", logEntry["tool"])
	assert.Equal(t, true, logEntry["success"])
}

func TestAuditMiddlewareFactory_ValidConfig(t *testing.T) {
	cfg := map[string]any{
		"include_args":   true,
		"include_result": false,
	}

	mw, err := AuditLoggingMiddlewareFactory(cfg)
	require.NoError(t, err)
	require.NotNil(t, mw)
}

func TestAuditMiddleware_Timestamp(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}
	auditLogger := &mockAuditLogger{}

	beforeCall := time.Now()

	mw := NewAuditLoggingMiddleware(AuditConfig{
		AuditLogger: auditLogger,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)

	afterCall := time.Now()

	entries := auditLogger.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.True(t, entry.Timestamp.After(beforeCall) || entry.Timestamp.Equal(beforeCall))
	assert.True(t, entry.Timestamp.Before(afterCall) || entry.Timestamp.Equal(afterCall))
}

func TestAuditMiddleware_PreservesProviderMethods(t *testing.T) {
	mock := &mockProvider{name: "original-name"}

	mw := NewAuditLoggingMiddleware(AuditConfig{})

	wrapped := mw(mock)

	assert.Equal(t, "original-name", wrapped.Name())
	assert.True(t, wrapped.Enabled())
	assert.Equal(t, "original-name", wrapped.Tool().Name)
}
