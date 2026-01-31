//go:build redis

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAPIKeyStoreConfig configures the Redis API key store.
type RedisAPIKeyStoreConfig struct {
	// Client is the Redis client to use.
	Client *redis.Client

	// KeyPrefix is the prefix for all API key entries in Redis.
	// Default: "apikey:"
	KeyPrefix string

	// DefaultTTL is the default TTL for stored keys.
	// If zero, keys don't expire in Redis (but may still have ExpiresAt).
	DefaultTTL time.Duration
}

// RedisAPIKeyStore stores API keys in Redis.
type RedisAPIKeyStore struct {
	client     *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

// NewRedisAPIKeyStore creates a new Redis-backed API key store.
func NewRedisAPIKeyStore(config RedisAPIKeyStoreConfig) (*RedisAPIKeyStore, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	keyPrefix := config.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "apikey:"
	}

	return &RedisAPIKeyStore{
		client:     config.Client,
		keyPrefix:  keyPrefix,
		defaultTTL: config.DefaultTTL,
	}, nil
}

// Lookup retrieves an API key by its hash.
func (s *RedisAPIKeyStore) Lookup(ctx context.Context, keyHash string) (*APIKeyInfo, error) {
	key := s.keyPrefix + keyHash

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var info APIKeyInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &info, nil
}

// Add stores an API key.
func (s *RedisAPIKeyStore) Add(ctx context.Context, info *APIKeyInfo) error {
	key := s.keyPrefix + info.KeyHash

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	ttl := s.defaultTTL
	if ttl == 0 {
		// No TTL, key persists until explicitly removed
		if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
			return fmt.Errorf("redis set: %w", err)
		}
	} else {
		if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
			return fmt.Errorf("redis set: %w", err)
		}
	}

	return nil
}

// Remove deletes an API key.
func (s *RedisAPIKeyStore) Remove(ctx context.Context, keyHash string) error {
	key := s.keyPrefix + keyHash

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
