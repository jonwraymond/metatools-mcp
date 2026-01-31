//go:build !redis

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisAPIKeyStore_StubReturnsError(t *testing.T) {
	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{})
	require.Error(t, err)
	assert.Nil(t, store)
	assert.Contains(t, err.Error(), "requires -tags=redis")
}
