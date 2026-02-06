//go:build !toolsemantic

package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
)

func semanticSearcherFromConfig(_ config.SearchConfig) (index.Searcher, error) {
	return nil, fmt.Errorf("semantic strategy requested but toolsemantic build tag is not enabled")
}
