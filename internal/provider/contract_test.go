// Package provider defines contracts for tool providers.
package provider

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type streamingProvider struct {
	tool mcp.Tool
}

func (s *streamingProvider) Name() string   { return "streaming" }
func (s *streamingProvider) Enabled() bool  { return true }
func (s *streamingProvider) Tool() mcp.Tool { return s.tool }
func (s *streamingProvider) Handle(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	return nil, nil, nil
}
func (s *streamingProvider) HandleStream(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (<-chan any, error) {
	ch := make(chan any)
	close(ch)
	return ch, nil
}

type configurableProvider struct {
	streamingProvider
}

func (c *configurableProvider) Configure(_ map[string]any) error { return nil }

func TestProviderContracts(t *testing.T) {
	var _ StreamingProvider = (*streamingProvider)(nil)
	var _ ConfigurableProvider = (*configurableProvider)(nil)

	p := &streamingProvider{}
	ch, err := p.HandleStream(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("HandleStream error: %v", err)
	}
	if ch == nil {
		t.Fatalf("HandleStream should return non-nil channel when err is nil")
	}
}
