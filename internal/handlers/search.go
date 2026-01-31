package handlers

import (
        "context"
        "errors"

        "github.com/jonwraymond/metatools-mcp/pkg/metatools"
        "github.com/jonwraymond/tooldiscovery/index"
        "github.com/modelcontextprotocol/go-sdk/jsonrpc"
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
        if err := ctx.Err(); err != nil {
                return nil, err
        }

        // Validate input
        if err := input.Validate(); err != nil {
                return nil, err
        }

        limit := input.GetLimit()

        // Search the index with cursor pagination
        var cursorStr string
        if input.Cursor != nil {
                cursorStr = *input.Cursor
        }

        tools, nextCursor, err := h.index.SearchPage(ctx, input.Query, limit, cursorStr)
        if err != nil {
                if errors.Is(err, index.ErrInvalidCursor) {
                        return nil, &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "invalid cursor"}
                }
                return nil, err
        }

        return &metatools.SearchToolsOutput{
                Tools:      tools,
                NextCursor: metatools.NullableCursor(nextCursor),
        }, nil
}
