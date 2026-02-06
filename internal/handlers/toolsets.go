package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolfoundation/adapter"
)

// ToolsetsHandler handles toolset metatools.
type ToolsetsHandler struct {
	registry ToolsetRegistry
}

// NewToolsetsHandler creates a new toolsets handler.
func NewToolsetsHandler(registry ToolsetRegistry) *ToolsetsHandler {
	return &ToolsetsHandler{registry: registry}
}

// List handles list_toolsets.
func (h *ToolsetsHandler) List(ctx context.Context, input metatools.ListToolsetsInput) (*metatools.ListToolsetsOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if h.registry == nil {
		return nil, errors.New("toolset registry not configured")
	}

	toolsets := h.registry.List()
	out := make([]metatools.ToolsetSummary, 0, len(toolsets))
	for _, ts := range toolsets {
		if ts == nil {
			continue
		}
		out = append(out, metatools.ToolsetSummary{
			ID:          ts.ID,
			Name:        ts.Name,
			Description: ts.Description,
			ToolCount:   len(ts.Tools),
		})
	}
	return &metatools.ListToolsetsOutput{Toolsets: out}, nil
}

// Describe handles describe_toolset.
func (h *ToolsetsHandler) Describe(ctx context.Context, input metatools.DescribeToolsetInput) (*metatools.DescribeToolsetOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if h.registry == nil {
		return nil, errors.New("toolset registry not configured")
	}

	ts, ok := h.registry.Get(strings.TrimSpace(input.ToolsetID))
	if !ok || ts == nil {
		return nil, errors.New("toolset not found")
	}

	tools := make([]metatools.ToolSummary, 0, len(ts.Tools))
	for _, tool := range ts.Tools {
		if tool == nil {
			continue
		}
		tools = append(tools, toolSummaryFromCanonical(tool))
	}

	return &metatools.DescribeToolsetOutput{
		Toolset: metatools.ToolsetDetail{
			ID:          ts.ID,
			Name:        ts.Name,
			Description: ts.Description,
			Tools:       tools,
		},
	}, nil
}

func toolSummaryFromCanonical(tool *adapter.CanonicalTool) metatools.ToolSummary {
	if tool == nil {
		return metatools.ToolSummary{}
	}
	short := strings.TrimSpace(tool.Summary)
	if short == "" {
		short = strings.TrimSpace(tool.Description)
	}
	return metatools.ToolSummary{
		ID:               tool.ID(),
		Name:             tool.Name,
		Namespace:        tool.Namespace,
		ShortDescription: short,
		Tags:             tool.Tags,
	}
}
