package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// EnvConfig holds configuration parsed from environment variables
type EnvConfig struct {
	Search SearchConfig `envPrefix:"METATOOLS_SEARCH_"`
}

// SearchConfig holds search-related configuration
type SearchConfig struct {
	Strategy           string `env:"STRATEGY" envDefault:"lexical"`
	BM25NameBoost      int    `env:"BM25_NAME_BOOST" envDefault:"3"`
	BM25NamespaceBoost int    `env:"BM25_NAMESPACE_BOOST" envDefault:"2"`
	BM25TagsBoost      int    `env:"BM25_TAGS_BOOST" envDefault:"2"`
	BM25MaxDocs        int    `env:"BM25_MAX_DOCS" envDefault:"0"`
	BM25MaxDocTextLen  int    `env:"BM25_MAX_DOCTEXT_LEN" envDefault:"0"`
}

// validStrategies defines the allowed search strategies
var validStrategies = map[string]bool{
	"lexical": true,
	"bm25":    true,
}

// LoadEnv parses environment variables into EnvConfig
func LoadEnv() (EnvConfig, error) {
	var cfg EnvConfig
	if err := env.Parse(&cfg); err != nil {
		return EnvConfig{}, fmt.Errorf("parsing env config: %w", err)
	}
	// Normalize strategy to avoid case/whitespace footguns.
	cfg.Search.Strategy = strings.ToLower(strings.TrimSpace(cfg.Search.Strategy))
	return cfg, nil
}

// ValidateEnv checks that the configuration values are valid
func (c *EnvConfig) ValidateEnv() error {
	if !validStrategies[c.Search.Strategy] {
		return fmt.Errorf("unknown search strategy %q: valid strategies are lexical, bm25", c.Search.Strategy)
	}
	return nil
}
