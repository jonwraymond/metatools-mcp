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
	summaries, _, err := a.idx.SearchPage(query, limit, "")
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

// SearchPage delegates to toolindex and converts summaries to metatools summaries.
func (a *IndexAdapter) SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error) {
	_ = ctx
	summaries, nextCursor, err := a.idx.SearchPage(query, limit, cursor)
	if err != nil {
		return nil, "", err
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
	return out, nextCursor, nil
}

// ListNamespaces delegates to toolindex.
func (a *IndexAdapter) ListNamespaces(ctx context.Context) ([]string, error) {
	_ = ctx
	namespaces, _, err := a.idx.ListNamespacesPage(100, "")
	return namespaces, err
}

// ListNamespacesPage delegates to toolindex.
func (a *IndexAdapter) ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error) {
	_ = ctx
	return a.idx.ListNamespacesPage(limit, cursor)
}

// OnChange registers a listener for index mutations when supported.
// Returns a no-op unsubscribe when change notifications are unavailable.
func (a *IndexAdapter) OnChange(listener toolindex.ChangeListener) func() {
	if notifier, ok := a.idx.(toolindex.ChangeNotifier); ok {
		return notifier.OnChange(listener)
	}
	return func() {}
}
