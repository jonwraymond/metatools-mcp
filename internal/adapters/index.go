package adapters

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolindex"
)

// IndexAdapter bridges toolindex.Index to the handlers.Index interface.
type IndexAdapter struct {
	idx toolindex.Index
}

// NewIndexAdapter creates a new index adapter.
func NewIndexAdapter(idx toolindex.Index) *IndexAdapter {
	return &IndexAdapter{idx: idx}
}

// Search delegates to toolindex and converts summaries to metatools summaries.
func (a *IndexAdapter) Search(ctx context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
	_ = ctx
	summaries, err := a.idx.Search(query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]metatools.ToolSummary, len(summaries))
	for i, s := range summaries {
		out[i] = metatools.ToolSummary{
			ID:               s.ID,
			Name:             s.Name,
			Namespace:        s.Namespace,
			ShortDescription: s.ShortDescription,
			Tags:             s.Tags,
		}
	}
	return out, nil
}

// ListNamespaces delegates to toolindex.
func (a *IndexAdapter) ListNamespaces(ctx context.Context) ([]string, error) {
	_ = ctx
	return a.idx.ListNamespaces()
}
