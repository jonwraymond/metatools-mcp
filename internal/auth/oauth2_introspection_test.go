package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createIntrospectionServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func introspectionResponse(active bool, claims map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]any{"active": active}
		for k, v := range claims {
			response[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func TestOAuth2_ExtractBearerToken(t *testing.T) {
	testCases := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{
			name:      "valid bearer token",
			header:    "Bearer abc123token",
			wantToken: "abc123token",
			wantOK:    true,
		},
		{
			name:      "bearer lowercase",
			header:    "bearer abc123token",
			wantToken: "abc123token",
			wantOK:    true,
		},
		{
			name:   "missing bearer prefix",
			header: "abc123token",
			wantOK: false,
		},
		{
			name:   "empty header",
			header: "",
			wantOK: false,
		},
		{
			name:   "bearer only no token",
			header: "Bearer ",
			wantOK: false,
		},
		{
			name:   "basic auth",
			header: "Basic dXNlcjpwYXNz",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, ok := extractBearerToken(tc.header)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.wantToken, token)
			}
		})
	}
}

func TestOAuth2_ClientSecretBasic(t *testing.T) {
	var receivedAuth string
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		introspectionResponse(true, map[string]any{"sub": "user1"})(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "my-client",
		ClientSecret:          "my-secret",
		ClientAuthMethod:      "client_secret_basic",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer test-token"},
		},
	}

	_, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)

	// Verify Basic auth was sent
	assert.Contains(t, receivedAuth, "Basic ")
}

func TestOAuth2_ClientSecretPost(t *testing.T) {
	var receivedClientID, receivedClientSecret string
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		receivedClientID = r.Form.Get("client_id")
		receivedClientSecret = r.Form.Get("client_secret")
		introspectionResponse(true, map[string]any{"sub": "user1"})(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "my-client",
		ClientSecret:          "my-secret",
		ClientAuthMethod:      "client_secret_post",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer test-token"},
		},
	}

	_, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "my-client", receivedClientID)
	assert.Equal(t, "my-secret", receivedClientSecret)
}

func TestOAuth2_ActiveToken(t *testing.T) {
	server := createIntrospectionServer(t, introspectionResponse(true, map[string]any{
		"sub":      "user@example.com",
		"scope":    "read write admin",
		"tenant":   "tenant-123",
		"roles":    []string{"admin", "user"},
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
	}))
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		PrincipalClaim:        "sub",
		TenantClaim:           "tenant",
		RolesClaim:            "roles",
		ScopesClaim:           "scope",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer valid-token"},
		},
	}

	result, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	require.True(t, result.Authenticated)
	require.NotNil(t, result.Identity)

	identity := result.Identity
	assert.Equal(t, "user@example.com", identity.Principal)
	assert.Equal(t, "tenant-123", identity.TenantID)
	assert.Contains(t, identity.Roles, "admin")
	assert.Contains(t, identity.Roles, "user")
	assert.Equal(t, AuthMethodOAuth2, identity.Method)
}

func TestOAuth2_InactiveToken(t *testing.T) {
	server := createIntrospectionServer(t, introspectionResponse(false, nil))
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer inactive-token"},
		},
	}

	result, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrTokenInactive)
}

func TestOAuth2_NetworkError(t *testing.T) {
	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: "http://localhost:1", // Invalid port
		ClientID:              "client",
		ClientSecret:          "secret",
		Timeout:               100 * time.Millisecond,
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer test-token"},
		},
	}

	_, err := auth.Authenticate(ctx, req)
	require.Error(t, err)
}

func TestOAuth2_ServerError(t *testing.T) {
	server := createIntrospectionServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer test-token"},
		},
	}

	_, err := auth.Authenticate(ctx, req)
	require.Error(t, err)
}

func TestOAuth2_CachesPositiveResponse(t *testing.T) {
	var callCount int32
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		introspectionResponse(true, map[string]any{"sub": "user1"})(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		CacheTTL:              time.Hour,
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer cached-token"},
		},
	}

	// First call
	result1, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.True(t, result1.Authenticated)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Second call - should use cache
	result2, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.True(t, result2.Authenticated)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Third call - still cached
	result3, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.True(t, result3.Authenticated)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestOAuth2_DoesNotCacheNegative(t *testing.T) {
	var callCount int32
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		introspectionResponse(false, nil)(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		CacheTTL:              time.Hour,
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer invalid-token"},
		},
	}

	// First call
	result1, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.False(t, result1.Authenticated)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Second call - should NOT be cached (negative response)
	result2, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.False(t, result2.Authenticated)
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
}

func TestOAuth2_CacheExpiration(t *testing.T) {
	var callCount int32
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		introspectionResponse(true, map[string]any{"sub": "user1"})(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		CacheTTL:              50 * time.Millisecond,
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer expiring-token"},
		},
	}

	// First call
	_, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Wait for cache expiration
	time.Sleep(100 * time.Millisecond)

	// Second call - cache expired, should fetch again
	_, err = auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
}

func TestOAuth2_ScopeToPermissions(t *testing.T) {
	server := createIntrospectionServer(t, introspectionResponse(true, map[string]any{
		"sub":   "user1",
		"scope": "read:users write:users admin",
	}))
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		ScopesClaim:           "scope",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer scoped-token"},
		},
	}

	result, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	require.True(t, result.Authenticated)

	// Scopes should be in permissions
	assert.Contains(t, result.Identity.Permissions, "read:users")
	assert.Contains(t, result.Identity.Permissions, "write:users")
	assert.Contains(t, result.Identity.Permissions, "admin")
}

func TestOAuth2_Supports(t *testing.T) {
	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: "http://example.com/introspect",
	})

	testCases := []struct {
		name    string
		headers map[string][]string
		want    bool
	}{
		{
			name:    "bearer token",
			headers: map[string][]string{"Authorization": {"Bearer token"}},
			want:    true,
		},
		{
			name:    "basic auth",
			headers: map[string][]string{"Authorization": {"Basic dXNlcjpwYXNz"}},
			want:    false,
		},
		{
			name:    "no auth header",
			headers: map[string][]string{},
			want:    false,
		},
		{
			name:    "api key header",
			headers: map[string][]string{"X-API-Key": {"key123"}},
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &AuthRequest{Headers: tc.headers}
			got := auth.Supports(context.Background(), req)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOAuth2_Name(t *testing.T) {
	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: "http://example.com/introspect",
	})

	assert.Equal(t, "oauth2_introspection", auth.Name())
}

func TestOAuth2_ConcurrentAccess(t *testing.T) {
	server := createIntrospectionServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate slow response
		introspectionResponse(true, map[string]any{"sub": "user1"})(w, r)
	})
	defer server.Close()

	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: server.URL,
		ClientID:              "client",
		ClientSecret:          "secret",
		CacheTTL:              time.Hour,
	})

	ctx := context.Background()
	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)
	results := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			req := &AuthRequest{
				Headers: map[string][]string{
					"Authorization": {"Bearer concurrent-token"},
				},
			}
			result, err := auth.Authenticate(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			results <- result.Authenticated
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}

	for authenticated := range results {
		assert.True(t, authenticated)
	}
}

func TestOAuth2_MissingCredentials(t *testing.T) {
	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: "http://example.com/introspect",
	})

	ctx := context.Background()
	req := &AuthRequest{
		Headers: map[string][]string{}, // No Authorization header
	}

	result, err := auth.Authenticate(ctx, req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestOAuth2_DefaultValues(t *testing.T) {
	auth := NewOAuth2IntrospectionAuthenticator(OAuth2Config{
		IntrospectionEndpoint: "http://example.com/introspect",
		// No other config - should use defaults
	})

	// Just verify it doesn't panic and has sensible defaults
	assert.NotNil(t, auth)
}
