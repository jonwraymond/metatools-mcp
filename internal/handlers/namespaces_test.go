package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNamespaces_Empty(t *testing.T) {
	idx := &mockIndex{
		listNamespacesFunc: func(_ context.Context) ([]string, error) {
			return []string{}, nil
		},
	}

	handler := NewNamespacesHandler(idx)
	result, err := handler.Handle(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result.Namespaces)
}

func TestListNamespaces_ReturnsSorted(t *testing.T) {
	idx := &mockIndex{
		listNamespacesFunc: func(_ context.Context) ([]string, error) {
			// Index should return sorted, but we verify
			return []string{"alpha", "beta", "gamma"}, nil
		},
	}

	handler := NewNamespacesHandler(idx)
	result, err := handler.Handle(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, result.Namespaces)
}

func TestListNamespaces_IndexError(t *testing.T) {
	idx := &mockIndex{
		listNamespacesFunc: func(_ context.Context) ([]string, error) {
			return nil, errors.New("index error")
		},
	}

	handler := NewNamespacesHandler(idx)
	_, err := handler.Handle(context.Background())

	assert.Error(t, err)
}
