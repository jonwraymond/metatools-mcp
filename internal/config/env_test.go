package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnv_Defaults(t *testing.T) {
	// Clear any existing env vars
	clearSearchEnvVars(t)

	cfg, err := LoadEnv()
	require.NoError(t, err)

	assert.Equal(t, "lexical", cfg.Search.Strategy)
	assert.Equal(t, 3, cfg.Search.BM25NameBoost)
	assert.Equal(t, 2, cfg.Search.BM25NamespaceBoost)
	assert.Equal(t, 2, cfg.Search.BM25TagsBoost)
	assert.Equal(t, 0, cfg.Search.BM25MaxDocs)
	assert.Equal(t, 0, cfg.Search.BM25MaxDocTextLen)
}

func TestLoadEnv_LexicalStrategy(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "lexical")

	cfg, err := LoadEnv()
	require.NoError(t, err)

	assert.Equal(t, "lexical", cfg.Search.Strategy)
}

func TestLoadEnv_BM25Strategy(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "bm25")

	cfg, err := LoadEnv()
	require.NoError(t, err)

	assert.Equal(t, "bm25", cfg.Search.Strategy)
}

func TestLoadEnv_BM25Strategy_CaseInsensitive(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "BM25")

	cfg, err := LoadEnv()
	require.NoError(t, err)

	assert.Equal(t, "bm25", cfg.Search.Strategy)
}

func TestLoadEnv_CustomBM25Values(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "bm25")
	t.Setenv("METATOOLS_SEARCH_BM25_NAME_BOOST", "5")
	t.Setenv("METATOOLS_SEARCH_BM25_NAMESPACE_BOOST", "4")
	t.Setenv("METATOOLS_SEARCH_BM25_TAGS_BOOST", "1")
	t.Setenv("METATOOLS_SEARCH_BM25_MAX_DOCS", "1000")
	t.Setenv("METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN", "500")

	cfg, err := LoadEnv()
	require.NoError(t, err)

	assert.Equal(t, "bm25", cfg.Search.Strategy)
	assert.Equal(t, 5, cfg.Search.BM25NameBoost)
	assert.Equal(t, 4, cfg.Search.BM25NamespaceBoost)
	assert.Equal(t, 1, cfg.Search.BM25TagsBoost)
	assert.Equal(t, 1000, cfg.Search.BM25MaxDocs)
	assert.Equal(t, 500, cfg.Search.BM25MaxDocTextLen)
}

func TestValidateEnv_ValidStrategies(t *testing.T) {
	tests := []struct {
		strategy string
	}{
		{"lexical"},
		{"bm25"},
	}

	for _, tc := range tests {
		t.Run(tc.strategy, func(t *testing.T) {
			cfg := EnvConfig{
				Search: SearchConfig{Strategy: tc.strategy},
			}
			err := cfg.ValidateEnv()
			assert.NoError(t, err)
		})
	}
}

func TestValidateEnv_UnknownStrategy(t *testing.T) {
	cfg := EnvConfig{
		Search: SearchConfig{Strategy: "unknown"},
	}

	err := cfg.ValidateEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown search strategy")
	assert.Contains(t, err.Error(), "unknown")
	assert.Contains(t, err.Error(), "lexical")
	assert.Contains(t, err.Error(), "bm25")
}

func TestValidateEnv_EmptyStrategy(t *testing.T) {
	cfg := EnvConfig{
		Search: SearchConfig{Strategy: ""},
	}

	err := cfg.ValidateEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown search strategy")
}

// clearSearchEnvVars unsets all METATOOLS_SEARCH_* env vars for test isolation
func clearSearchEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{
		"METATOOLS_SEARCH_STRATEGY",
		"METATOOLS_SEARCH_BM25_NAME_BOOST",
		"METATOOLS_SEARCH_BM25_NAMESPACE_BOOST",
		"METATOOLS_SEARCH_BM25_TAGS_BOOST",
		"METATOOLS_SEARCH_BM25_MAX_DOCS",
		"METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN",
	}
	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			t.Fatalf("unset %s: %v", v, err)
		}
	}
}
