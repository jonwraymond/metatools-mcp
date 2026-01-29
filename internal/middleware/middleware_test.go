package middleware

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type recordingProvider struct {
	name  string
	calls *[]string
}

func (r *recordingProvider) Name() string  { return r.name }
func (r *recordingProvider) Enabled() bool { return true }
func (r *recordingProvider) Tool() mcp.Tool {
	return mcp.Tool{Name: r.name, InputSchema: map[string]any{"type": "object"}}
}
func (r *recordingProvider) Handle(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	*r.calls = append(*r.calls, "provider")
	return nil, "ok", nil
}

func TestChain_ApplyOrder(t *testing.T) {
	calls := []string{}
	base := &recordingProvider{name: "test", calls: &calls}

	mw1 := func(next provider.ToolProvider) provider.ToolProvider {
		return &wrapProvider{
			name:   next.Name(),
			next:   next,
			before: "mw1:before",
			after:  "mw1:after",
			calls:  &calls,
		}
	}
	mw2 := func(next provider.ToolProvider) provider.ToolProvider {
		return &wrapProvider{
			name:   next.Name(),
			next:   next,
			before: "mw2:before",
			after:  "mw2:after",
			calls:  &calls,
		}
	}

	chain := NewChain(mw1, mw2)
	wrapped := chain.Apply(base)
	_, _, _ = wrapped.Handle(context.Background(), nil, nil)

	want := []string{"mw1:before", "mw2:before", "provider", "mw2:after", "mw1:after"}
	if len(calls) != len(want) {
		t.Fatalf("calls length = %d, want %d", len(calls), len(want))
	}
	for i, v := range want {
		if calls[i] != v {
			t.Fatalf("calls[%d] = %q, want %q", i, calls[i], v)
		}
	}
}

func TestChain_LenClear(t *testing.T) {
	chain := NewChain()
	if chain.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", chain.Len())
	}
	chain.Use(func(next provider.ToolProvider) provider.ToolProvider { return next })
	if chain.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", chain.Len())
	}
	chain.Clear()
	if chain.Len() != 0 {
		t.Fatalf("Len() = %d, want 0 after Clear()", chain.Len())
	}
}

func TestChain_ApplyToRegistry(t *testing.T) {
	calls := []string{}
	registry := provider.NewRegistry()
	_ = registry.Register(&recordingProvider{name: "test", calls: &calls})

	mw := func(next provider.ToolProvider) provider.ToolProvider {
		return &wrapProvider{
			name:   next.Name(),
			next:   next,
			before: "mw:before",
			after:  "mw:after",
			calls:  &calls,
		}
	}
	chain := NewChain(mw)
	if err := chain.ApplyToRegistry(registry); err != nil {
		t.Fatalf("ApplyToRegistry() error = %v", err)
	}

	p, ok := registry.Get("test")
	if !ok {
		t.Fatalf("expected provider in registry after ApplyToRegistry")
	}
	_, _, _ = p.Handle(context.Background(), nil, nil)

	if len(calls) < 2 {
		t.Fatalf("expected middleware to wrap provider")
	}
}

type wrapProvider struct {
	name   string
	next   provider.ToolProvider
	before string
	after  string
	calls  *[]string
}

func (w *wrapProvider) Name() string   { return w.name }
func (w *wrapProvider) Enabled() bool  { return w.next.Enabled() }
func (w *wrapProvider) Tool() mcp.Tool { return w.next.Tool() }
func (w *wrapProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	*w.calls = append(*w.calls, w.before)
	res, out, err := w.next.Handle(ctx, req, args)
	*w.calls = append(*w.calls, w.after)
	return res, out, err
}
