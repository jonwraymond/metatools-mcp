package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func rsaPublicKeyToJWK(pub *rsa.PublicKey, kid string) map[string]any {
	return map[string]any{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func createJWKSServer(t *testing.T, keys []map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jwks := map[string]any{
			"keys": keys,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
}

// Tests

func TestJWKSKeyProvider_FetchAndParseKeys(t *testing.T) {
	// Setup: create a test RSA key and JWKS server
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "test-key-1")
	server := createJWKSServer(t, []map[string]any{jwk})
	defer server.Close()

	// Create provider
	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: server.URL,
	})

	// Fetch key
	ctx := context.Background()
	key, err := provider.GetKey(ctx, "test-key-1")

	// Verify
	require.NoError(t, err)
	require.NotNil(t, key)

	pubKey, ok := key.(*rsa.PublicKey)
	require.True(t, ok, "expected *rsa.PublicKey")
	assert.Equal(t, rsaKey.N, pubKey.N)
	assert.Equal(t, rsaKey.E, pubKey.E)
}

func TestJWKSKeyProvider_GetKeyByKID(t *testing.T) {
	// Setup: create multiple keys
	key1 := generateTestRSAKey(t)
	key2 := generateTestRSAKey(t)

	jwks := []map[string]any{
		rsaPublicKeyToJWK(&key1.PublicKey, "key-alpha"),
		rsaPublicKeyToJWK(&key2.PublicKey, "key-beta"),
	}
	server := createJWKSServer(t, jwks)
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: server.URL,
	})

	ctx := context.Background()

	// Fetch first key
	result1, err := provider.GetKey(ctx, "key-alpha")
	require.NoError(t, err)
	pubKey1 := result1.(*rsa.PublicKey)
	assert.Equal(t, key1.N, pubKey1.N)

	// Fetch second key
	result2, err := provider.GetKey(ctx, "key-beta")
	require.NoError(t, err)
	pubKey2 := result2.(*rsa.PublicKey)
	assert.Equal(t, key2.N, pubKey2.N)
}

func TestJWKSKeyProvider_KeyNotFound(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "existing-key")
	server := createJWKSServer(t, []map[string]any{jwk})
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: server.URL,
	})

	ctx := context.Background()
	_, err := provider.GetKey(ctx, "nonexistent-key")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestJWKSKeyProvider_ParseRSAPublicKey(t *testing.T) {
	testCases := []struct {
		name    string
		jwk     map[string]any
		wantErr bool
	}{
		{
			name: "valid RSA key",
			jwk: func() map[string]any {
				key := generateTestRSAKey(t)
				return rsaPublicKeyToJWK(&key.PublicKey, "valid-rsa")
			}(),
			wantErr: false,
		},
		{
			name: "missing n parameter",
			jwk: map[string]any{
				"kty": "RSA",
				"kid": "missing-n",
				"e":   "AQAB",
			},
			wantErr: true,
		},
		{
			name: "missing e parameter",
			jwk: map[string]any{
				"kty": "RSA",
				"kid": "missing-e",
				"n":   "test",
			},
			wantErr: true,
		},
		{
			name: "invalid base64 in n",
			jwk: map[string]any{
				"kty": "RSA",
				"kid": "invalid-n",
				"n":   "!!!invalid-base64!!!",
				"e":   "AQAB",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := createJWKSServer(t, []map[string]any{tc.jwk})
			defer server.Close()

			provider := NewJWKSKeyProvider(JWKSConfig{
				URL: server.URL,
			})

			ctx := context.Background()
			kid := tc.jwk["kid"].(string)
			_, err := provider.GetKey(ctx, kid)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJWKSKeyProvider_CachesKeys(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "cached-key")

	var fetchCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&fetchCount, 1)
		jwks := map[string]any{"keys": []map[string]any{jwk}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL:      server.URL,
		CacheTTL: time.Hour,
	})

	ctx := context.Background()

	// First call - should fetch
	_, err := provider.GetKey(ctx, "cached-key")
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetchCount))

	// Second call - should use cache
	_, err = provider.GetKey(ctx, "cached-key")
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetchCount))

	// Third call - still cached
	_, err = provider.GetKey(ctx, "cached-key")
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetchCount))
}

func TestJWKSKeyProvider_CacheExpiration(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "expiring-key")

	var fetchCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&fetchCount, 1)
		jwks := map[string]any{"keys": []map[string]any{jwk}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	// Use a very short TTL for testing
	provider := NewJWKSKeyProvider(JWKSConfig{
		URL:      server.URL,
		CacheTTL: 50 * time.Millisecond,
	})

	ctx := context.Background()

	// First call
	_, err := provider.GetKey(ctx, "expiring-key")
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetchCount))

	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// Should refetch
	_, err = provider.GetKey(ctx, "expiring-key")
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&fetchCount))
}

func TestJWKSKeyProvider_ConcurrentAccess(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "concurrent-key")

	var fetchCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&fetchCount, 1)
		// Simulate slow response
		time.Sleep(10 * time.Millisecond)
		jwks := map[string]any{"keys": []map[string]any{jwk}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL:      server.URL,
		CacheTTL: time.Hour,
	})

	ctx := context.Background()
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)
	keys := make(chan any, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			key, err := provider.GetKey(ctx, "concurrent-key")
			if err != nil {
				errors <- err
				return
			}
			keys <- key
		}()
	}

	wg.Wait()
	close(errors)
	close(keys)

	// Check no errors
	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}

	// All goroutines should get the same key
	var firstKey *rsa.PublicKey
	for key := range keys {
		pubKey := key.(*rsa.PublicKey)
		if firstKey == nil {
			firstKey = pubKey
		} else {
			assert.Equal(t, firstKey.N, pubKey.N)
		}
	}

	// Should have minimal fetches (ideally 1, but allow for race conditions)
	assert.LessOrEqual(t, atomic.LoadInt32(&fetchCount), int32(3))
}

func TestJWKSKeyProvider_HTTPError(t *testing.T) {
	testCases := []struct {
		name       string
		handler    http.HandlerFunc
		wantErrMsg string
	}{
		{
			name: "server returns 500",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErrMsg: "unexpected status",
		},
		{
			name: "server returns 404",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErrMsg: "unexpected status",
		},
		{
			name: "invalid JSON response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("not valid json"))
			},
			wantErrMsg: "decode",
		},
		{
			name: "missing keys field",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			},
			wantErrMsg: "key not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			provider := NewJWKSKeyProvider(JWKSConfig{
				URL: server.URL,
			})

			ctx := context.Background()
			_, err := provider.GetKey(ctx, "any-key")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrMsg)
		})
	}
}

func TestJWKSKeyProvider_NetworkError(t *testing.T) {
	// Use an invalid URL that will fail to connect
	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: "http://localhost:1", // Port 1 is reserved, connection will fail
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := provider.GetKey(ctx, "any-key")
	require.Error(t, err)
}

func TestJWKSKeyProvider_RefreshFailureUsesCached(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "fallback-key")

	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			// First call succeeds
			jwks := map[string]any{"keys": []map[string]any{jwk}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(jwks)
		} else {
			// Subsequent calls fail
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL:      server.URL,
		CacheTTL: 50 * time.Millisecond,
	})

	ctx := context.Background()

	// First call populates cache
	key1, err := provider.GetKey(ctx, "fallback-key")
	require.NoError(t, err)
	require.NotNil(t, key1)

	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// Second call fails to refresh but returns cached key
	key2, err := provider.GetKey(ctx, "fallback-key")
	require.NoError(t, err)
	require.NotNil(t, key2)

	// Should be the same key
	assert.Equal(t, key1.(*rsa.PublicKey).N, key2.(*rsa.PublicKey).N)
}

func TestJWKSKeyProvider_EmptyKIDMatchesFirstKey(t *testing.T) {
	// When no kid is specified in token, should use first available key
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "only-key")
	server := createJWKSServer(t, []map[string]any{jwk})
	defer server.Close()

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: server.URL,
	})

	ctx := context.Background()

	// Request with empty kid
	key, err := provider.GetKey(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, key)

	pubKey := key.(*rsa.PublicKey)
	assert.Equal(t, rsaKey.N, pubKey.N)
}

func TestJWKSKeyProvider_CustomHTTPClient(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "custom-client-key")
	server := createJWKSServer(t, []map[string]any{jwk})
	defer server.Close()

	// Create custom client with short timeout
	customClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	provider := NewJWKSKeyProvider(JWKSConfig{
		URL:        server.URL,
		HTTPClient: customClient,
	})

	ctx := context.Background()
	key, err := provider.GetKey(ctx, "custom-client-key")

	require.NoError(t, err)
	require.NotNil(t, key)
}

func TestJWKSKeyProvider_DefaultCacheTTL(t *testing.T) {
	rsaKey := generateTestRSAKey(t)
	jwk := rsaPublicKeyToJWK(&rsaKey.PublicKey, "default-ttl-key")
	server := createJWKSServer(t, []map[string]any{jwk})
	defer server.Close()

	// Don't specify CacheTTL - should use default
	provider := NewJWKSKeyProvider(JWKSConfig{
		URL: server.URL,
	})

	// Verify default is set (1 hour)
	assert.Equal(t, time.Hour, provider.config.CacheTTL)
}
