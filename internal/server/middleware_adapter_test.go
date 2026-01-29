package server

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/middleware"
	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type countProvider struct {
	name  string
	count *int
}

func (c *countProvider) Name() string  { return c.name }
func (c *countProvider) Enabled() bool { return true }
func (c *countProvider) Tool() mcp.Tool {
	return mcp.Tool{Name: c.name, InputSchema: map[string]any{"type": "object"}}
}
func (c *countProvider) Handle(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	*c.count++
	return nil, "ok", nil
}

func TestMiddlewareAdapter_ApplyToProviders(t *testing.T) {
	registry := provider.NewRegistry()
	providerCount := 0
	middlewareCount := 0
	_ = registry.Register(&countProvider{name: "test", count: &providerCount})

	chain := middleware.NewChain(func(next provider.ToolProvider) provider.ToolProvider {
		return &countingMiddlewareProvider{next: next, count: &middlewareCount}
	})

	adapter := NewMiddlewareAdapter(chain)
	if err := adapter.ApplyToProviders(registry); err != nil {
		t.Fatalf("ApplyToProviders() error = %v", err)
	}

	p, ok := registry.Get("test")
	if !ok {
		t.Fatal("expected provider in registry")
	}
	_, _, _ = p.Handle(context.Background(), nil, nil)
	if providerCount != 1 {
		t.Fatalf("providerCount = %d, want 1", providerCount)
	}
	if middlewareCount != 1 {
		t.Fatalf("middlewareCount = %d, want 1", middlewareCount)
	}
}

func TestNewMiddlewareAdapterFromConfig(t *testing.T) {
	cfg := &middleware.Config{
		Chain: []string{},
	}
	adapter, err := NewMiddlewareAdapterFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewMiddlewareAdapterFromConfig() error = %v", err)
	}
	if adapter.Chain().Len() != 0 {
		t.Fatalf("Chain.Len() = %d, want 0", adapter.Chain().Len())
	}
}

type countingMiddlewareProvider struct {
	next  provider.ToolProvider
	count *int
}

func (c *countingMiddlewareProvider) Name() string   { return c.next.Name() }
func (c *countingMiddlewareProvider) Enabled() bool  { return c.next.Enabled() }
func (c *countingMiddlewareProvider) Tool() mcp.Tool { return c.next.Tool() }
func (c *countingMiddlewareProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	*c.count++
	return c.next.Handle(ctx, req, args)
}
