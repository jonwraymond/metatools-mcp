package handlers

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// SearchHandler handles the search_tools metatool
type SearchHandler struct {
	index Index
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(index Index) *SearchHandler {
	return &SearchHandler{index: index}
}

// Handle executes the search_tools metatool
func (h *SearchHandler) Handle(ctx context.Context, input metatools.SearchToolsInput) (*metatools.SearchToolsOutput, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	limit := input.GetLimit()

	// Search the index
	tools, err := h.index.Search(ctx, input.Query, limit)
	if err != nil {
		return nil, err
	}

	// Apply cursor pagination
	var cursorStr string
	if input.Cursor != nil {
		cursorStr = *input.Cursor
	}

	paginatedTools, nextCursor := metatools.ApplyCursor(tools, cursorStr, limit)

	return &metatools.SearchToolsOutput{
		Tools:      paginatedTools,
		NextCursor: nextCursor,
	}, nil
}
