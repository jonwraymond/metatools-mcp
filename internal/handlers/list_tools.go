package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

// ListToolsHandler handles the list_tools metatool.
type ListToolsHandler struct {
	index     Index
	refresher Refresher
}

// NewListToolsHandler creates a new list tools handler.
func NewListToolsHandler(index Index) *ListToolsHandler {
	return &ListToolsHandler{index: index}
}

// NewListToolsHandlerWithRefresher creates a list tools handler that can refresh backends on demand.
func NewListToolsHandlerWithRefresher(index Index, refresher Refresher) *ListToolsHandler {
	return &ListToolsHandler{index: index, refresher: refresher}
}

// Handle executes the list_tools metatool.
func (h *ListToolsHandler) Handle(ctx context.Context, input metatools.ListToolsInput) (*metatools.ListToolsOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	if h.refresher != nil {
		if err := h.refresher.MaybeRefresh(ctx); err != nil {
			slog.Default().Warn("mcp backend refresh failed", "err", err)
		}
	}

	limit := input.GetLimit()
	cursor := ""
	if input.Cursor != nil {
		cursor = *input.Cursor
	}

	backendKind := ""
	if input.BackendKind != nil {
		backendKind = strings.TrimSpace(strings.ToLower(*input.BackendKind))
	}
	backendName := ""
	if input.BackendName != nil {
		backendName = strings.TrimSpace(*input.BackendName)
	}

	tools, nextCursor, err := h.listPage(ctx, limit, cursor, backendKind, backendName)
	if err != nil {
		if errors.Is(err, index.ErrInvalidCursor) {
			return nil, &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "invalid cursor"}
		}
		return nil, err
	}

	return &metatools.ListToolsOutput{
		Tools:      tools,
		NextCursor: metatools.NullableCursor(nextCursor),
	}, nil
}

func (h *ListToolsHandler) listPage(ctx context.Context, limit int, cursor string, backendKind string, backendName string) ([]metatools.ToolSummary, string, error) {
	if backendKind == "" && backendName == "" {
		return h.index.SearchPage(ctx, "", limit, cursor)
	}

	var out []metatools.ToolSummary
	nextCursor := cursor

	for len(out) < limit {
		page, next, err := h.index.SearchPage(ctx, "", limit, nextCursor)
		if err != nil {
			return nil, "", err
		}
		for _, summary := range page {
			backends, err := h.index.GetAllBackends(ctx, summary.ID)
			if err != nil {
				if errors.Is(err, index.ErrNotFound) {
					continue
				}
				return nil, "", err
			}
			if backendMatches(backends, backendKind, backendName) {
				out = append(out, summary)
				if len(out) >= limit {
					return out, next, nil
				}
			}
		}
		if next == "" {
			return out, "", nil
		}
		nextCursor = next
	}

	return out, nextCursor, nil
}

func backendMatches(backends []model.ToolBackend, kind string, name string) bool {
	for _, backend := range backends {
		if kind != "" && string(backend.Kind) != kind {
			continue
		}
		if name == "" {
			return true
		}
		switch backend.Kind {
		case model.BackendKindMCP:
			if backend.MCP != nil && backend.MCP.ServerName == name {
				return true
			}
		case model.BackendKindProvider:
			if backend.Provider != nil && backend.Provider.ProviderID == name {
				return true
			}
		case model.BackendKindLocal:
			if backend.Local != nil && backend.Local.Name == name {
				return true
			}
		}
	}
	return false
}
