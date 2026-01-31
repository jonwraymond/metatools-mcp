//go:build !redis

package auth

import (
	"context"
	"fmt"
	"time"
)

// RedisAPIKeyStoreConfig configures the Redis API key store.
// This is a stub - build with -tags=redis for full implementation.
type RedisAPIKeyStoreConfig struct {
	// Client would be the Redis client.
	Client any

	// KeyPrefix is the prefix for all API key entries in Redis.
	KeyPrefix string

	// DefaultTTL is the default TTL for stored keys.
	DefaultTTL time.Duration
}

// RedisAPIKeyStore is a stub that requires the redis build tag.
type RedisAPIKeyStore struct{}

// NewRedisAPIKeyStore returns an error when built without the redis tag.
func NewRedisAPIKeyStore(_ RedisAPIKeyStoreConfig) (*RedisAPIKeyStore, error) {
	return nil, fmt.Errorf("redis API key store requires -tags=redis build flag")
}

// Lookup always returns an error in the stub.
func (s *RedisAPIKeyStore) Lookup(_ context.Context, _ string) (*APIKeyInfo, error) {
	return nil, fmt.Errorf("redis API key store not available")
}

// Add always returns an error in the stub.
func (s *RedisAPIKeyStore) Add(_ context.Context, _ *APIKeyInfo) error {
	return fmt.Errorf("redis API key store not available")
}

// Remove always returns an error in the stub.
func (s *RedisAPIKeyStore) Remove(_ context.Context, _ string) error {
	return fmt.Errorf("redis API key store not available")
}
