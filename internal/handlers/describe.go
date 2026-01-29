package handlers

import (
        "context"

        "github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// DescribeHandler handles the describe_tool metatool
type DescribeHandler struct {
        store Store
}

// NewDescribeHandler creates a new describe handler
func NewDescribeHandler(store Store) *DescribeHandler {
        return &DescribeHandler{store: store}
}

// Handle executes the describe_tool metatool
func (h *DescribeHandler) Handle(ctx context.Context, input metatools.DescribeToolInput) (*metatools.DescribeToolOutput, error) {
        if err := ctx.Err(); err != nil {
                return nil, err
        }

        // Validate input
        if err := input.Validate(); err != nil {
                return nil, err
        }

        // Get documentation from store
        doc, err := h.store.DescribeTool(ctx, input.ToolID, input.DetailLevel)
        if err != nil {
                return nil, err
        }

        // Build output
        output := &metatools.DescribeToolOutput{
                Summary:      doc.Summary,
                Tool:         doc.Tool,
                SchemaInfo:   doc.SchemaInfo,
                Notes:        doc.Notes,
                Examples:     doc.Examples,
                ExternalRefs: doc.ExternalRefs,
        }

        // Apply examples cap if specified
        if input.ExamplesMax != nil && len(output.Examples) > *input.ExamplesMax {
                output.Examples = output.Examples[:*input.ExamplesMax]
        }

        return output, nil
}
