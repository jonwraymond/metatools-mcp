package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test context key for identity
type testContextKey string

const testPrincipalKey testContextKey = "principal"

func withTestPrincipal(ctx context.Context, principal string) context.Context {
	return context.WithValue(ctx, testPrincipalKey, principal)
}

func testIdentityExtractor(ctx context.Context) string {
	if v, ok := ctx.Value(testPrincipalKey).(string); ok {
		return v
	}
	return ""
}

// mockProvider implements provider.ToolProvider for testing
type mockProvider struct {
	name       string
	callCount  int32
	handleFunc func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
}

func (m *mockProvider) Name() string    { return m.name }
func (m *mockProvider) Enabled() bool   { return true }
func (m *mockProvider) Tool() mcp.Tool  { return mcp.Tool{Name: m.name} }
func (m *mockProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	atomic.AddInt32(&m.callCount, 1)
	if m.handleFunc != nil {
		return m.handleFunc(ctx, req, args)
	}
	return &mcp.CallToolResult{}, nil, nil
}

func TestRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  10.0, // 10 requests per second
		Burst: 5,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Should allow up to burst requests
	for i := 0; i < 5; i++ {
		result, _, err := wrapped.Handle(ctx, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	assert.Equal(t, int32(5), atomic.LoadInt32(&mock.callCount))
}

func TestRateLimitMiddleware_BlocksOverLimit(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  1.0, // 1 request per second
		Burst: 2,   // Allow only 2 burst
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// First two should succeed (burst)
	for i := 0; i < 2; i++ {
		result, _, err := wrapped.Handle(ctx, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	// Third should be rate limited
	result, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err) // No error, but result indicates rate limit
	assert.True(t, result.IsError)

	// Verify the error message
	require.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "rate limit exceeded")
}

func TestRateLimitMiddleware_BurstAllowed(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  1.0,  // 1 per second
		Burst: 10,   // But allow 10 burst
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// All 10 burst requests should succeed immediately
	for i := 0; i < 10; i++ {
		result, _, err := wrapped.Handle(ctx, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError, "request %d should succeed", i)
	}

	assert.Equal(t, int32(10), atomic.LoadInt32(&mock.callCount))
}

func TestRateLimitMiddleware_PerToolLimit(t *testing.T) {
	mock1 := &mockProvider{name: "slow-tool"}
	mock2 := &mockProvider{name: "fast-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  100.0, // High global rate
		Burst: 100,
		PerTool: map[string]RateLimitRule{
			"slow-tool": {Rate: 1.0, Burst: 1},
			"fast-tool": {Rate: 100.0, Burst: 100},
		},
	})

	wrapped1 := mw(mock1)
	wrapped2 := mw(mock2)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// slow-tool: first request OK
	result, _, err := wrapped1.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// slow-tool: second request blocked
	result, _, err = wrapped1.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	// fast-tool: should still work
	for i := 0; i < 10; i++ {
		result, _, err := wrapped2.Handle(ctx, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}
}

func TestRateLimitMiddleware_PerIdentityLimit(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:        100.0, // High global rate
		Burst:       100,
		PerIdentity: true,
		IdentityRate: RateLimitRule{
			Rate:  1.0,
			Burst: 2,
		},
		IdentityExtractor: testIdentityExtractor,
	})

	wrapped := mw(mock)
	req := &mcp.CallToolRequest{}

	// User1 context
	ctx1 := withTestPrincipal(context.Background(), "user1")

	// User2 context
	ctx2 := withTestPrincipal(context.Background(), "user2")

	// User1: first two requests OK
	for i := 0; i < 2; i++ {
		result, _, err := wrapped.Handle(ctx1, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	// User1: third blocked
	result, _, err := wrapped.Handle(ctx1, req, nil)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	// User2: should still have their own limit
	for i := 0; i < 2; i++ {
		result, _, err := wrapped.Handle(ctx2, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}
}

func TestRateLimitMiddleware_GlobalFallback(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:        100.0, // High rate for authenticated
		Burst:       100,
		PerIdentity: true,
		IdentityRate: RateLimitRule{
			Rate:  100.0,
			Burst: 100,
		},
		GlobalFallback: &RateLimitRule{
			Rate:  1.0, // Low rate for anonymous
			Burst: 2,
		},
		IdentityExtractor: testIdentityExtractor, // Returns "" for anonymous
	})

	wrapped := mw(mock)
	req := &mcp.CallToolRequest{}

	// Anonymous context (no identity)
	ctx := context.Background()

	// First two OK
	for i := 0; i < 2; i++ {
		result, _, err := wrapped.Handle(ctx, req, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	// Third blocked by global fallback
	result, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestRateLimitMiddleware_ThreadSafe(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  1000.0,
		Burst: 1000,
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	const goroutines = 50
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines*requestsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				_, _, err := wrapped.Handle(ctx, req, nil)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRateLimitMiddlewareFactory_ValidConfig(t *testing.T) {
	cfg := map[string]any{
		"rate":  10.0,
		"burst": 20,
		"per_tool": map[string]any{
			"slow-tool": map[string]any{
				"rate":  1.0,
				"burst": 2,
			},
		},
		"per_identity": true,
	}

	mw, err := RateLimitMiddlewareFactory(cfg)
	require.NoError(t, err)
	require.NotNil(t, mw)

	// Test that the middleware works
	mock := &mockProvider{name: "test-tool"}
	wrapped := mw(mock)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := wrapped.Handle(ctx, req, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestRateLimitMiddlewareFactory_DefaultValues(t *testing.T) {
	// Empty config should use defaults
	mw, err := RateLimitMiddlewareFactory(map[string]any{})
	require.NoError(t, err)
	require.NotNil(t, mw)
}

func TestRateLimitMiddleware_PreservesProviderMethods(t *testing.T) {
	mock := &mockProvider{name: "original-name"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  10.0,
		Burst: 10,
	})

	wrapped := mw(mock)

	assert.Equal(t, "original-name", wrapped.Name())
	assert.True(t, wrapped.Enabled())
	assert.Equal(t, "original-name", wrapped.Tool().Name)
}

func TestRateLimitMiddleware_RefillsOverTime(t *testing.T) {
	mock := &mockProvider{name: "test-tool"}

	mw := NewRateLimitMiddleware(RateLimitConfig{
		Rate:  10.0, // 10 per second = refill every 100ms
		Burst: 1,    // Only 1 burst
	})

	wrapped := mw(mock)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Use the burst
	result, _, _ := wrapped.Handle(ctx, req, nil)
	assert.False(t, result.IsError)

	// Immediate second request should fail
	result, _, _ = wrapped.Handle(ctx, req, nil)
	assert.True(t, result.IsError)

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should work again
	result, _, _ = wrapped.Handle(ctx, req, nil)
	assert.False(t, result.IsError)
}
