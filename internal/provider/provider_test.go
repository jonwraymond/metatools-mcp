package provider

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockProvider implements ToolProvider for testing
//
//nolint:revive // test helper
type mockProvider struct {
	name     string
	enabled  bool
	tool     mcp.Tool
	handleFn func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
}

func (m *mockProvider) Name() string   { return m.name }
func (m *mockProvider) Enabled() bool  { return m.enabled }
func (m *mockProvider) Tool() mcp.Tool { return m.tool }
func (m *mockProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	if m.handleFn != nil {
		return m.handleFn(ctx, req, args)
	}
	return nil, nil, nil
}

func TestToolProvider_Interface(t *testing.T) {
	provider := &mockProvider{
		name:    "test_tool",
		enabled: true,
		tool: mcp.Tool{
			Name:        "test_tool",
			Description: "Test tool",
		},
		handleFn: func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
			return nil, "result", nil
		},
	}

	if provider.Name() != "test_tool" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "test_tool")
	}
	if !provider.Enabled() {
		t.Errorf("Enabled() = false, want true")
	}

	tool := provider.Tool()
	if tool.Name != "test_tool" {
		t.Errorf("Tool().Name = %q, want %q", tool.Name, "test_tool")
	}

	_, result, err := provider.Handle(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}
	if result != "result" {
		t.Errorf("Handle() = %v, want %v", result, "result")
	}
}
