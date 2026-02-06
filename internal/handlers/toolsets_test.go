package handlers

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolfoundation/adapter"
	"github.com/stretchr/testify/require"
)

func TestToolsetsHandler_ListAndDescribe(t *testing.T) {
	tool := &adapter.CanonicalTool{
		Namespace:   "ns",
		Name:        "alpha",
		Description: "alpha tool",
		Summary:     "alpha summary",
		Tags:        []string{"a"},
	}
	ts := &toolset.Toolset{
		ID:          "toolset:alpha",
		Name:        "Alpha",
		Description: "Alpha toolset",
		Tools:       []*adapter.CanonicalTool{tool},
	}

	reg := toolset.NewRegistry([]*toolset.Toolset{ts})
	handler := NewToolsetsHandler(reg)

	list, err := handler.List(context.Background(), metatools.ListToolsetsInput{})
	require.NoError(t, err)
	require.Len(t, list.Toolsets, 1)
	require.Equal(t, "toolset:alpha", list.Toolsets[0].ID)
	require.Equal(t, 1, list.Toolsets[0].ToolCount)

	desc, err := handler.Describe(context.Background(), metatools.DescribeToolsetInput{ToolsetID: "toolset:alpha"})
	require.NoError(t, err)
	require.Equal(t, "Alpha", desc.Toolset.Name)
	require.Len(t, desc.Toolset.Tools, 1)
	require.Equal(t, "ns:alpha", desc.Toolset.Tools[0].ID)
}
