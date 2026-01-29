package handlers

import (
        "context"
        "errors"

        "github.com/jonwraymond/metatools-mcp/pkg/metatools"
        "github.com/jonwraymond/toolindex"
        "github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

// NamespacesHandler handles the list_namespaces metatool
type NamespacesHandler struct {
        index Index
}

// NewNamespacesHandler creates a new namespaces handler
func NewNamespacesHandler(index Index) *NamespacesHandler {
        return &NamespacesHandler{index: index}
}

// Handle executes the list_namespaces metatool
func (h *NamespacesHandler) Handle(ctx context.Context, input metatools.ListNamespacesInput) (*metatools.ListNamespacesOutput, error) {
        if err := ctx.Err(); err != nil {
                return nil, err
        }

        if err := input.Validate(); err != nil {
                return nil, err
        }

        limit := input.GetLimit()
        var cursorStr string
        if input.Cursor != nil {
                cursorStr = *input.Cursor
        }

        namespaces, nextCursor, err := h.index.ListNamespacesPage(ctx, limit, cursorStr)
        if err != nil {
                if errors.Is(err, toolindex.ErrInvalidCursor) {
                        return nil, &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "invalid cursor"}
                }
                return nil, err
        }

        return &metatools.ListNamespacesOutput{
                Namespaces: namespaces,
                NextCursor: metatools.NullableCursor(nextCursor),
        }, nil
}
