package handlers

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
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
func (h *NamespacesHandler) Handle(ctx context.Context) (*metatools.ListNamespacesOutput, error) {
	namespaces, err := h.index.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	return &metatools.ListNamespacesOutput{
		Namespaces: namespaces,
	}, nil
}
