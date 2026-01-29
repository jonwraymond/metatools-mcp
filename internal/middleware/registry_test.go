package middleware

import (
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
)

func TestRegistry_RegisterGet(t *testing.T) {
	registry := NewRegistry()

	factory := func(_ map[string]any) (Middleware, error) {
		return func(next provider.ToolProvider) provider.ToolProvider { return next }, nil
	}

	if err := registry.Register("test", factory); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if !registry.Has("test") {
		t.Error("Has() returned false for registered middleware")
	}

	if _, ok := registry.Get("test"); !ok {
		t.Error("Get() returned false for registered middleware")
	}

	if err := registry.Register("test", factory); err == nil {
		t.Error("Register() should fail on duplicate")
	}
}

func TestRegistry_Create(t *testing.T) {
	registry := NewRegistry()

	factory := func(_ map[string]any) (Middleware, error) {
		return func(next provider.ToolProvider) provider.ToolProvider { return next }, nil
	}
	_ = registry.Register("test", factory)

	if _, err := registry.Create("missing", nil); err == nil {
		t.Error("Create() should fail for missing middleware")
	}

	if _, err := registry.Create("test", map[string]any{}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestRegistry_ListClear(t *testing.T) {
	registry := NewRegistry()

	factory := func(_ map[string]any) (Middleware, error) {
		return func(next provider.ToolProvider) provider.ToolProvider { return next }, nil
	}
	_ = registry.Register("a", factory)
	_ = registry.Register("b", factory)

	names := registry.List()
	if len(names) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(names))
	}

	registry.Clear()
	if len(registry.List()) != 0 {
		t.Fatal("Clear() should remove all middleware")
	}
}
