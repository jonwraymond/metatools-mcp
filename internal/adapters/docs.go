package adapters

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/tooldocs"
)

// DocsAdapter bridges tooldocs.Store to the handlers.Store interface.
type DocsAdapter struct {
	store tooldocs.Store
}

// NewDocsAdapter creates a new docs adapter.
func NewDocsAdapter(store tooldocs.Store) *DocsAdapter {
	return &DocsAdapter{store: store}
}

// DescribeTool delegates to tooldocs and converts the result.
func (a *DocsAdapter) DescribeTool(ctx context.Context, id string, level string) (handlers.ToolDoc, error) {
	_ = ctx
	doc, err := a.store.DescribeTool(id, tooldocs.DetailLevel(level))
	if err != nil {
		return handlers.ToolDoc{}, err
	}

	var notes *string
	if doc.Notes != "" {
		n := doc.Notes
		notes = &n
	}

	return handlers.ToolDoc{
		Tool:         doc.Tool,
		Summary:      doc.Summary,
		SchemaInfo:   doc.SchemaInfo,
		Notes:        notes,
		Examples:     convertExamples(doc.Examples),
		ExternalRefs: doc.ExternalRefs,
	}, nil
}

// ListExamples delegates to tooldocs and converts the result.
func (a *DocsAdapter) ListExamples(ctx context.Context, id string, max int) ([]metatools.ToolExample, error) {
	_ = ctx
	examples, err := a.store.ListExamples(id, max)
	if err != nil {
		return nil, err
	}
	return convertExamples(examples), nil
}

func convertExamples(examples []tooldocs.ToolExample) []metatools.ToolExample {
	out := make([]metatools.ToolExample, len(examples))
	for i, ex := range examples {
		out[i] = metatools.ToolExample{
			ID:          ex.ID,
			Title:       ex.Title,
			Description: ex.Description,
			Args:        ex.Args,
			ResultHint:  ex.ResultHint,
		}
	}
	return out
}
