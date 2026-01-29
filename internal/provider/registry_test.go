package provider

import "testing"

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test_tool", enabled: true}

	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err = registry.Register(provider)
	if err == nil {
		t.Error("Register() should fail on duplicate")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	provider := &mockProvider{name: "test_tool", enabled: true}
	_ = registry.Register(provider)

	got, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("Get() returned false")
	}
	if got.Name() != "test_tool" {
		t.Errorf("Get().Name() = %q, want %q", got.Name(), "test_tool")
	}

	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for nonexistent provider")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	_ = registry.Register(&mockProvider{name: "tool_a", enabled: true})
	_ = registry.Register(&mockProvider{name: "tool_b", enabled: true})
	_ = registry.Register(&mockProvider{name: "tool_c", enabled: false})

	all := registry.List()
	if len(all) != 3 {
		t.Errorf("List() returned %d providers, want 3", len(all))
	}

	enabled := registry.ListEnabled()
	if len(enabled) != 2 {
		t.Errorf("ListEnabled() returned %d providers, want 2", len(enabled))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	_ = registry.Register(&mockProvider{name: "tool_a", enabled: true})

	if err := registry.Unregister("tool_a"); err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	if _, ok := registry.Get("tool_a"); ok {
		t.Error("Get() should return false after Unregister()")
	}

	if err := registry.Unregister("missing"); err == nil {
		t.Error("Unregister() should fail for missing provider")
	}
}
