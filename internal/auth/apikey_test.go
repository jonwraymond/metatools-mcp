package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthenticator_ValidKey(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	keyHash := HashAPIKey("test-api-key-123", "sha256")
	err := store.Add(&APIKeyInfo{
		ID:        "key-1",
		KeyHash:   keyHash,
		Principal: "alice",
		TenantID:  "tenant-1",
		Roles:     []string{"user"},
	})
	require.NoError(t, err)

	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-API-Key": {"test-api-key-123"},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.NotNil(t, result.Identity)
	assert.Equal(t, "alice", result.Identity.Principal)
	assert.Equal(t, "tenant-1", result.Identity.TenantID)
	assert.Equal(t, []string{"user"}, result.Identity.Roles)
	assert.Equal(t, AuthMethodAPIKey, result.Identity.Method)
}

func TestAPIKeyAuthenticator_HashedKey(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	rawKey := "my-secret-api-key"
	keyHash := HashAPIKey(rawKey, "sha256")

	err := store.Add(&APIKeyInfo{
		ID:        "key-2",
		KeyHash:   keyHash,
		Principal: "bob",
	})
	require.NoError(t, err)

	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-API-Key": {rawKey},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "bob", result.Identity.Principal)
}

func TestAPIKeyAuthenticator_InvalidKey(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-API-Key": {"nonexistent-key"},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrInvalidCredentials)
}

func TestAPIKeyAuthenticator_ExpiredKey(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	expired := time.Now().Add(-time.Hour)
	keyHash := HashAPIKey("expired-key", "sha256")

	err := store.Add(&APIKeyInfo{
		ID:        "key-3",
		KeyHash:   keyHash,
		Principal: "charlie",
		ExpiresAt: &expired,
	})
	require.NoError(t, err)

	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-API-Key": {"expired-key"},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrTokenExpired)
}

func TestAPIKeyAuthenticator_MissingKey(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	req := &AuthRequest{
		Headers: map[string][]string{},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestAPIKeyAuthenticator_Supports(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	auth := NewAPIKeyAuthenticator(APIKeyConfig{}, store)

	tests := []struct {
		name     string
		req      *AuthRequest
		expected bool
	}{
		{
			name: "supports X-API-Key header",
			req: &AuthRequest{
				Headers: map[string][]string{
					"X-API-Key": {"some-key"},
				},
			},
			expected: true,
		},
		{
			name: "does not support missing header",
			req: &AuthRequest{
				Headers: map[string][]string{},
			},
			expected: false,
		},
		{
			name: "does not support bearer token",
			req: &AuthRequest{
				Headers: map[string][]string{
					"Authorization": {"Bearer token"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, auth.Supports(context.Background(), tt.req))
		})
	}
}

func TestMemoryAPIKeyStore_Add(t *testing.T) {
	store := NewMemoryAPIKeyStore()

	info := &APIKeyInfo{
		ID:        "key-1",
		KeyHash:   "hash123",
		Principal: "alice",
	}

	err := store.Add(info)
	require.NoError(t, err)

	// Duplicate hash should fail
	info2 := &APIKeyInfo{
		ID:        "key-2",
		KeyHash:   "hash123",
		Principal: "bob",
	}
	err = store.Add(info2)
	assert.Error(t, err)
}

func TestMemoryAPIKeyStore_Lookup(t *testing.T) {
	store := NewMemoryAPIKeyStore()

	info := &APIKeyInfo{
		ID:        "key-1",
		KeyHash:   "hash123",
		Principal: "alice",
	}
	err := store.Add(info)
	require.NoError(t, err)

	// Found
	found, err := store.Lookup(context.Background(), "hash123")
	require.NoError(t, err)
	assert.Equal(t, info, found)

	// Not found
	notFound, err := store.Lookup(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestMemoryAPIKeyStore_Remove(t *testing.T) {
	store := NewMemoryAPIKeyStore()

	info := &APIKeyInfo{
		ID:      "key-1",
		KeyHash: "hash123",
	}
	err := store.Add(info)
	require.NoError(t, err)

	store.Remove("hash123")

	found, err := store.Lookup(context.Background(), "hash123")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestHashAPIKey(t *testing.T) {
	hash1 := HashAPIKey("test-key", "sha256")
	hash2 := HashAPIKey("test-key", "sha256")
	hash3 := HashAPIKey("different-key", "sha256")

	// Same input produces same hash
	assert.Equal(t, hash1, hash2)

	// Different input produces different hash
	assert.NotEqual(t, hash1, hash3)

	// Hash is hex encoded
	assert.Len(t, hash1, 64) // SHA256 produces 32 bytes = 64 hex chars
}

func TestValidateAPIKey(t *testing.T) {
	rawKey := "my-secret-key"
	hash := HashAPIKey(rawKey, "sha256")

	// Valid key
	assert.True(t, ValidateAPIKey(rawKey, hash, "sha256"))

	// Invalid key
	assert.False(t, ValidateAPIKey("wrong-key", hash, "sha256"))
}

func TestAPIKeyAuthenticator_CustomHeaderName(t *testing.T) {
	store := NewMemoryAPIKeyStore()
	keyHash := HashAPIKey("custom-key", "sha256")
	err := store.Add(&APIKeyInfo{
		ID:        "key-1",
		KeyHash:   keyHash,
		Principal: "alice",
	})
	require.NoError(t, err)

	auth := NewAPIKeyAuthenticator(APIKeyConfig{
		HeaderName: "X-Custom-Key",
	}, store)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-Custom-Key": {"custom-key"},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
}
