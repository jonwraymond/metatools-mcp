//go:build !toolsearch

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateServer_BM25WithoutBuildTag_FailsFast(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "bm25")

	srv, err := createServer()
	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "toolsearch build tag")
}

