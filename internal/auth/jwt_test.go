package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-that-is-long-enough"

func createTestJWT(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func createTestRSAJWT(t *testing.T, claims jwt.MapClaims, key *rsa.PrivateKey) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestJWTAuthenticator_ValidToken(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer:         "test-issuer",
		Audience:       "test-audience",
		PrincipalClaim: "sub",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"iss": "test-issuer",
		"aud": "test-audience",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.NotNil(t, result.Identity)
	assert.Equal(t, "alice", result.Identity.Principal)
	assert.Equal(t, AuthMethodJWT, result.Identity.Method)
}

func TestJWTAuthenticator_ExpiredToken(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer:         "test-issuer",
		PrincipalClaim: "sub",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"iss": "test-issuer",
		"exp": time.Now().Add(-time.Hour).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrTokenExpired)
}

func TestJWTAuthenticator_InvalidSignature(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer: "test-issuer",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"iss": "test-issuer",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, "wrong-secret")

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrInvalidToken)
}

func TestJWTAuthenticator_MissingToken(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{}, NewStaticKeyProvider([]byte(testSecret)))

	req := &AuthRequest{
		Headers: map[string][]string{},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestJWTAuthenticator_InvalidIssuer(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer: "expected-issuer",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"iss": "wrong-issuer",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrInvalidIssuer)
}

func TestJWTAuthenticator_TenantClaim(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer:         "test-issuer",
		PrincipalClaim: "sub",
		TenantClaim:    "tenant_id",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub":       "alice",
		"iss":       "test-issuer",
		"tenant_id": "tenant-123",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "tenant-123", result.Identity.TenantID)
}

func TestJWTAuthenticator_RolesClaim(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		Issuer:         "test-issuer",
		PrincipalClaim: "sub",
		RolesClaim:     "roles",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub":   "alice",
		"iss":   "test-issuer",
		"roles": []any{"admin", "user"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, []string{"admin", "user"}, result.Identity.Roles)
}

func TestJWTAuthenticator_Supports(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{}, NewStaticKeyProvider([]byte(testSecret)))

	tests := []struct {
		name     string
		req      *AuthRequest
		expected bool
	}{
		{
			name: "supports bearer token",
			req: &AuthRequest{
				Headers: map[string][]string{
					"Authorization": {"Bearer token"},
				},
			},
			expected: true,
		},
		{
			name: "does not support basic auth",
			req: &AuthRequest{
				Headers: map[string][]string{
					"Authorization": {"Basic dXNlcjpwYXNz"},
				},
			},
			expected: false,
		},
		{
			name:     "does not support missing auth",
			req:      &AuthRequest{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, auth.Supports(context.Background(), tt.req))
		})
	}
}

func TestStaticKeyProvider(t *testing.T) {
	key := []byte("test-key")
	provider := NewStaticKeyProvider(key)

	got, err := provider.GetKey(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, key, got)
}

func TestJWTAuthenticator_RSAKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	auth := NewJWTAuthenticator(JWTConfig{
		Issuer:         "test-issuer",
		PrincipalClaim: "sub",
	}, NewStaticKeyProvider(&privateKey.PublicKey))

	tokenStr := createTestRSAJWT(t, jwt.MapClaims{
		"sub": "bob",
		"iss": "test-issuer",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, privateKey)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "bob", result.Identity.Principal)
}

func TestJWTAuthenticator_CustomHeaderName(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		HeaderName:  "X-Auth-Token",
		TokenPrefix: "",
	}, NewStaticKeyProvider([]byte(testSecret)))

	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"X-Auth-Token": {tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
}

func TestJWTAuthenticator_ClockSkew(t *testing.T) {
	auth := NewJWTAuthenticator(JWTConfig{
		ClockSkew: 5 * time.Minute,
	}, NewStaticKeyProvider([]byte(testSecret)))

	// Token that expired 2 minutes ago (within 5 minute skew)
	tokenStr := createTestJWT(t, jwt.MapClaims{
		"sub": "alice",
		"exp": time.Now().Add(-2 * time.Minute).Unix(),
	}, testSecret)

	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	result, err := auth.Authenticate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
}
