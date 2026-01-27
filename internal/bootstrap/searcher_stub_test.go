//go:build !toolsearch

package bootstrap

import (
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSearcherFromConfig_StubAlwaysReturnsNil(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
	}{
		{"lexical strategy", "lexical", false},
		{"bm25 strategy (not available)", "bm25", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.SearchConfig{
				Strategy: tc.strategy,
			}
			searcher, err := SearcherFromConfig(cfg)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, searcher)
				return
			}
			assert.NoError(t, err)
			assert.Nil(t, searcher, "lexical should delegate to toolindex default")
		})
	}
}
