//go:build redis

package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require a running Redis instance.
// Run with: go test -tags=redis ./internal/auth/... -v -run Redis

func getTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use DB 15 for testing
	})

	// Check connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Clean up test DB
	client.FlushDB(ctx)

	t.Cleanup(func() {
		client.FlushDB(ctx)
		client.Close()
	})

	return client
}

func TestRedisAPIKeyStore_Lookup_Found(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	// Pre-populate key
	info := &APIKeyInfo{
		ID:        "key-123",
		KeyHash:   "abc123hash",
		Principal: "user@example.com",
		TenantID:  "tenant-1",
		Roles:     []string{"reader", "writer"},
	}

	// Store directly in Redis
	data, _ := json.Marshal(info)
	client.Set(ctx, "test:apikey:abc123hash", data, 0)

	// Lookup
	result, err := store.Lookup(ctx, "abc123hash")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "key-123", result.ID)
	assert.Equal(t, "abc123hash", result.KeyHash)
	assert.Equal(t, "user@example.com", result.Principal)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, []string{"reader", "writer"}, result.Roles)
}

func TestRedisAPIKeyStore_Lookup_NotFound(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	result, err := store.Lookup(ctx, "nonexistent-hash")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisAPIKeyStore_Add(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	info := &APIKeyInfo{
		ID:          "key-456",
		KeyHash:     "def456hash",
		Principal:   "admin@example.com",
		TenantID:    "tenant-2",
		Roles:       []string{"admin"},
		Permissions: []string{"read:*", "write:*"},
		Metadata:    map[string]string{"env": "test"},
	}

	err = store.Add(ctx, info)
	require.NoError(t, err)

	// Verify in Redis
	data, err := client.Get(ctx, "test:apikey:def456hash").Bytes()
	require.NoError(t, err)

	var stored APIKeyInfo
	require.NoError(t, json.Unmarshal(data, &stored))

	assert.Equal(t, "key-456", stored.ID)
	assert.Equal(t, "admin@example.com", stored.Principal)
	assert.Equal(t, []string{"admin"}, stored.Roles)
	assert.Equal(t, map[string]string{"env": "test"}, stored.Metadata)
}

func TestRedisAPIKeyStore_Serialization(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	expires := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	info := &APIKeyInfo{
		ID:          "key-789",
		KeyHash:     "ghi789hash",
		Principal:   "service@example.com",
		TenantID:    "tenant-3",
		Roles:       []string{"service"},
		Permissions: []string{"execute:*"},
		ExpiresAt:   &expires,
		Metadata:    map[string]string{"type": "service", "version": "1.0"},
	}

	// Add and then retrieve
	err = store.Add(ctx, info)
	require.NoError(t, err)

	result, err := store.Lookup(ctx, "ghi789hash")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all fields survived serialization
	assert.Equal(t, info.ID, result.ID)
	assert.Equal(t, info.KeyHash, result.KeyHash)
	assert.Equal(t, info.Principal, result.Principal)
	assert.Equal(t, info.TenantID, result.TenantID)
	assert.Equal(t, info.Roles, result.Roles)
	assert.Equal(t, info.Permissions, result.Permissions)
	assert.Equal(t, info.Metadata, result.Metadata)

	// Time comparison with tolerance
	require.NotNil(t, result.ExpiresAt)
	assert.WithinDuration(t, expires, *result.ExpiresAt, time.Second)
}

func TestRedisAPIKeyStore_KeyExpiration(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:     client,
		KeyPrefix:  "test:apikey:",
		DefaultTTL: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	info := &APIKeyInfo{
		ID:        "expiring-key",
		KeyHash:   "expiring-hash",
		Principal: "temp@example.com",
	}

	err = store.Add(ctx, info)
	require.NoError(t, err)

	// Should be found immediately
	result, err := store.Lookup(ctx, "expiring-hash")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Should be gone
	result, err = store.Lookup(ctx, "expiring-hash")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisAPIKeyStore_ConnectionFailure(t *testing.T) {
	// Use invalid address
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:1", // Invalid port
	})
	defer client.Close()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = store.Lookup(ctx, "any-hash")
	require.Error(t, err)
}

func TestRedisAPIKeyStore_ConcurrentAccess(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	// Pre-populate
	info := &APIKeyInfo{
		ID:        "concurrent-key",
		KeyHash:   "concurrent-hash",
		Principal: "concurrent@example.com",
	}
	err = store.Add(ctx, info)
	require.NoError(t, err)

	// Concurrent reads
	const goroutines = 20
	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			result, err := store.Lookup(ctx, "concurrent-hash")
			if err != nil {
				errors <- err
				done <- true
				return
			}
			if result == nil || result.ID != "concurrent-key" {
				errors <- assert.AnError
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestRedisAPIKeyStore_Remove(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client:    client,
		KeyPrefix: "test:apikey:",
	})
	require.NoError(t, err)

	info := &APIKeyInfo{
		ID:        "removable-key",
		KeyHash:   "removable-hash",
		Principal: "remove@example.com",
	}

	err = store.Add(ctx, info)
	require.NoError(t, err)

	// Verify exists
	result, err := store.Lookup(ctx, "removable-hash")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Remove
	err = store.Remove(ctx, "removable-hash")
	require.NoError(t, err)

	// Verify gone
	result, err = store.Lookup(ctx, "removable-hash")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisAPIKeyStore_DefaultKeyPrefix(t *testing.T) {
	client := getTestRedisClient(t)
	ctx := context.Background()

	// Don't specify prefix
	store, err := NewRedisAPIKeyStore(RedisAPIKeyStoreConfig{
		Client: client,
	})
	require.NoError(t, err)

	info := &APIKeyInfo{
		ID:        "default-prefix-key",
		KeyHash:   "default-prefix-hash",
		Principal: "default@example.com",
	}

	err = store.Add(ctx, info)
	require.NoError(t, err)

	// Check that default prefix was used
	_, err = client.Get(ctx, "apikey:default-prefix-hash").Result()
	require.NoError(t, err)
}
