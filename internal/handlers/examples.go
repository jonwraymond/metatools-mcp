package handlers

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// ExamplesHandler handles the list_tool_examples metatool
type ExamplesHandler struct {
	store Store
}

// NewExamplesHandler creates a new examples handler
func NewExamplesHandler(store Store) *ExamplesHandler {
	return &ExamplesHandler{store: store}
}

// Handle executes the list_tool_examples metatool
func (h *ExamplesHandler) Handle(ctx context.Context, input metatools.ListToolExamplesInput) (*metatools.ListToolExamplesOutput, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	maxExamples := input.GetMax()

	examples, err := h.store.ListExamples(ctx, input.ToolID, maxExamples)
	if err != nil {
		return nil, err
	}

	return &metatools.ListToolExamplesOutput{
		Examples: examples,
	}, nil
}
