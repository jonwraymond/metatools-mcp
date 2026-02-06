package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	describeToolFunc func(ctx context.Context, id string, level string) (ToolDoc, error)
	listExamplesFunc func(ctx context.Context, id string, maxExamples int) ([]metatools.ToolExample, error)
}

func (m *mockStore) DescribeTool(ctx context.Context, id string, level string) (ToolDoc, error) {
	if m.describeToolFunc != nil {
		return m.describeToolFunc(ctx, id, level)
	}
	return ToolDoc{}, nil
}

func (m *mockStore) ListExamples(ctx context.Context, id string, maxExamples int) ([]metatools.ToolExample, error) {
	if m.listExamplesFunc != nil {
		return m.listExamplesFunc(ctx, id, maxExamples)
	}
	return []metatools.ToolExample{}, nil
}

func TestDescribeTool_Summary(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, level string) (ToolDoc, error) {
			assert.Equal(t, "summary", level)
			return ToolDoc{
				Summary: "A test tool for testing",
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "summary",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "A test tool for testing", result.Summary)
}

func TestDescribeTool_Schema(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, level string) (ToolDoc, error) {
			assert.Equal(t, "schema", level)
			return ToolDoc{
				Summary: "A test tool",
				Tool: map[string]any{
					"name":        "tool",
					"description": "A test tool",
				},
				SchemaInfo: map[string]any{
					"inputSchema": map[string]any{"type": "object"},
				},
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "schema",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result.Tool)
	assert.NotNil(t, result.SchemaInfo)
}

func TestDescribeTool_TypedNilSchemaInfoOmitted(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			var schema map[string]any
			return ToolDoc{
				Summary:    "A test tool",
				SchemaInfo: schema, // typed-nil map: would JSON-marshal to null unless normalized
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "summary",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Nil(t, result.SchemaInfo)
}

func TestDescribeTool_Full(t *testing.T) {
	notes := "These are usage notes"
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, level string) (ToolDoc, error) {
			assert.Equal(t, "full", level)
			return ToolDoc{
				Summary: "A test tool",
				Tool: map[string]any{
					"name": "tool",
				},
				SchemaInfo: map[string]any{
					"inputSchema": map[string]any{"type": "object"},
				},
				Notes: &notes,
				Examples: []metatools.ToolExample{
					{Title: "Example 1", Description: "First example", Args: map[string]any{}},
				},
				ExternalRefs: []string{"https://docs.example.com/tool"},
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "full",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result.Tool)
	require.NotNil(t, result.Notes)
	assert.Equal(t, notes, *result.Notes)
	assert.Len(t, result.Examples, 1)
	assert.Len(t, result.ExternalRefs, 1)
}

func TestDescribeTool_SummaryOmitsTool(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{
				Summary: "A test tool",
				// Tool is intentionally nil for summary level
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "summary",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Nil(t, result.Tool)
}

func TestDescribeTool_SchemaIncludesTool(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{
				Summary: "A test tool",
				Tool:    map[string]any{"name": "tool"},
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "schema",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result.Tool)
}

func TestDescribeTool_FullIncludesExamples(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{
				Summary: "A test tool",
				Examples: []metatools.ToolExample{
					{Title: "Ex1", Description: "Desc1", Args: map[string]any{}},
					{Title: "Ex2", Description: "Desc2", Args: map[string]any{}},
				},
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "full",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Examples, 2)
}

func TestDescribeTool_ExamplesMaxCap(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{
				Summary: "A test tool",
				Examples: []metatools.ToolExample{
					{Title: "Ex1", Description: "Desc1", Args: map[string]any{}},
					{Title: "Ex2", Description: "Desc2", Args: map[string]any{}},
					{Title: "Ex3", Description: "Desc3", Args: map[string]any{}},
					{Title: "Ex4", Description: "Desc4", Args: map[string]any{}},
					{Title: "Ex5", Description: "Desc5", Args: map[string]any{}},
				},
			}, nil
		},
	}

	handler := NewDescribeHandler(store)
	maxExamples := 2
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "full",
		ExamplesMax: &maxExamples,
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Examples, 2)
}

func TestDescribeTool_NotFound(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{}, errors.New("tool not found")
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{
		ToolID:      "nonexistent.tool",
		DetailLevel: "summary",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}

func TestDescribeTool_NoTool(t *testing.T) {
	handler := NewDescribeHandler(&mockStore{})
	input := metatools.DescribeToolInput{
		ToolID:      "",
		DetailLevel: "summary",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}

func TestDescribeTool_InvalidDetailLevel(t *testing.T) {
	handler := NewDescribeHandler(&mockStore{})
	input := metatools.DescribeToolInput{
		ToolID:      "test.tool",
		DetailLevel: "invalid",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}
