package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonwraymond/metatools-mcp/internal/auth"
	"github.com/jonwraymond/metatools-mcp/internal/middleware"
)

const testSecret = "integration-test-secret-key-32ch"

func createJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

// mockToolProvider for integration tests
type mockToolProvider struct {
	name      string
	tool      mcp.Tool
	identity  *auth.Identity
	handleErr error
}

func (m *mockToolProvider) Name() string  { return m.name }
func (m *mockToolProvider) Enabled() bool { return true }
func (m *mockToolProvider) Tool() mcp.Tool {
	return m.tool
}

func (m *mockToolProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	// Capture identity from context
	m.identity = auth.IdentityFromContext(ctx)
	if m.handleErr != nil {
		return nil, nil, m.handleErr
	}
	return &mcp.CallToolResult{}, map[string]any{"status": "ok"}, nil
}

func TestIntegration_JWTAuth(t *testing.T) {
	// Setup JWT authenticator
	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
		Issuer:         "test-issuer",
		PrincipalClaim: "sub",
		RolesClaim:     "roles",
		TenantClaim:    "tenant_id",
	}, auth.NewStaticKeyProvider([]byte(testSecret)))

	// Create auth middleware
	authMW := auth.AuthMiddleware(jwtAuth, auth.AuthMiddlewareConfig{})

	// Create mock provider
	mp := &mockToolProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	// Apply middleware
	wrapped := authMW(mp)

	// Create valid JWT
	tokenStr := createJWT(t, jwt.MapClaims{
		"sub":       "alice",
		"iss":       "test-issuer",
		"roles":     []any{"admin", "user"},
		"tenant_id": "tenant-123",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})

	// Create context with headers
	ctx := auth.WithHeaders(context.Background(), map[string][]string{
		"Authorization": {"Bearer " + tokenStr},
	})

	// Call handler
	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify identity was set
	require.NotNil(t, mp.identity)
	assert.Equal(t, "alice", mp.identity.Principal)
	assert.Equal(t, "tenant-123", mp.identity.TenantID)
	assert.Equal(t, []string{"admin", "user"}, mp.identity.Roles)
	assert.Equal(t, auth.AuthMethodJWT, mp.identity.Method)
}

func TestIntegration_APIKeyAuth(t *testing.T) {
	// Setup API key store
	store := auth.NewMemoryAPIKeyStore()
	keyHash := auth.HashAPIKey("test-api-key-abc123", "sha256")
	err := store.Add(&auth.APIKeyInfo{
		ID:        "key-1",
		KeyHash:   keyHash,
		Principal: "service-account",
		TenantID:  "tenant-456",
		Roles:     []string{"service"},
	})
	require.NoError(t, err)

	// Setup API key authenticator
	apiKeyAuth := auth.NewAPIKeyAuthenticator(auth.APIKeyConfig{}, store)

	// Create auth middleware
	authMW := auth.AuthMiddleware(apiKeyAuth, auth.AuthMiddlewareConfig{})

	// Create mock provider
	mp := &mockToolProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	// Apply middleware
	wrapped := authMW(mp)

	// Create context with API key header
	ctx := auth.WithHeaders(context.Background(), map[string][]string{
		"X-API-Key": {"test-api-key-abc123"},
	})

	// Call handler
	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify identity was set
	require.NotNil(t, mp.identity)
	assert.Equal(t, "service-account", mp.identity.Principal)
	assert.Equal(t, "tenant-456", mp.identity.TenantID)
	assert.Equal(t, []string{"service"}, mp.identity.Roles)
	assert.Equal(t, auth.AuthMethodAPIKey, mp.identity.Method)
}

func TestIntegration_CompositeAuth(t *testing.T) {
	// Setup JWT authenticator
	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
		PrincipalClaim: "sub",
	}, auth.NewStaticKeyProvider([]byte(testSecret)))

	// Setup API key authenticator
	store := auth.NewMemoryAPIKeyStore()
	keyHash := auth.HashAPIKey("fallback-key", "sha256")
	_ = store.Add(&auth.APIKeyInfo{
		ID:        "key-1",
		KeyHash:   keyHash,
		Principal: "fallback-user",
	})
	apiKeyAuth := auth.NewAPIKeyAuthenticator(auth.APIKeyConfig{}, store)

	// Create composite authenticator (JWT first, API key fallback)
	composite := auth.NewCompositeAuthenticator(jwtAuth, apiKeyAuth)

	// Create auth middleware
	authMW := auth.AuthMiddleware(composite, auth.AuthMiddlewareConfig{})

	mp := &mockToolProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}
	wrapped := authMW(mp)

	t.Run("uses JWT when available", func(t *testing.T) {
		tokenStr := createJWT(t, jwt.MapClaims{
			"sub": "jwt-user",
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		ctx := auth.WithHeaders(context.Background(), map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		})

		result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
		assert.Equal(t, "jwt-user", mp.identity.Principal)
	})

	t.Run("falls back to API key", func(t *testing.T) {
		ctx := auth.WithHeaders(context.Background(), map[string][]string{
			"X-API-Key": {"fallback-key"},
		})

		result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
		assert.Equal(t, "fallback-user", mp.identity.Principal)
	})
}

func TestIntegration_AuthAndAuthzChain(t *testing.T) {
	// Setup JWT authenticator
	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
		PrincipalClaim: "sub",
		RolesClaim:     "roles",
	}, auth.NewStaticKeyProvider([]byte(testSecret)))

	// Setup RBAC authorizer
	rbacAuthz := auth.NewSimpleRBACAuthorizer(auth.RBACConfig{
		Roles: map[string]auth.RoleConfig{
			"admin": {
				Permissions: []string{"*"},
			},
			"user": {
				AllowedTools: []string{"search_*", "describe_*"},
			},
		},
	})

	// Create middleware chain
	authMW := auth.AuthMiddleware(jwtAuth, auth.AuthMiddlewareConfig{})
	authzMW := auth.AuthzMiddleware(rbacAuthz, auth.AuthzMiddlewareConfig{})

	// Apply middleware in order: auth -> authz -> handler
	chain := middleware.NewChain(authMW, authzMW)

	t.Run("admin can access any tool", func(t *testing.T) {
		mp := &mockToolProvider{
			name: "execute_code",
			tool: mcp.Tool{Name: "execute_code"},
		}
		wrapped := chain.Apply(mp)

		tokenStr := createJWT(t, jwt.MapClaims{
			"sub":   "admin-user",
			"roles": []any{"admin"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		ctx := auth.WithHeaders(context.Background(), map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		})

		result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})

	t.Run("user can access allowed tools", func(t *testing.T) {
		mp := &mockToolProvider{
			name: "search_tools",
			tool: mcp.Tool{Name: "search_tools"},
		}
		wrapped := chain.Apply(mp)

		tokenStr := createJWT(t, jwt.MapClaims{
			"sub":   "regular-user",
			"roles": []any{"user"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		ctx := auth.WithHeaders(context.Background(), map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		})

		result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})

	t.Run("user cannot access denied tools", func(t *testing.T) {
		mp := &mockToolProvider{
			name: "execute_code",
			tool: mcp.Tool{Name: "execute_code"},
		}
		wrapped := chain.Apply(mp)

		tokenStr := createJWT(t, jwt.MapClaims{
			"sub":   "regular-user",
			"roles": []any{"user"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		ctx := auth.WithHeaders(context.Background(), map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		})

		result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError, "expected authorization to fail")
	})
}

func TestIntegration_DeniedTool(t *testing.T) {
	// Setup with RBAC that explicitly denies execute_code
	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
		PrincipalClaim: "sub",
		RolesClaim:     "roles",
	}, auth.NewStaticKeyProvider([]byte(testSecret)))

	rbacAuthz := auth.NewSimpleRBACAuthorizer(auth.RBACConfig{
		Roles: map[string]auth.RoleConfig{
			"restricted": {
				AllowedTools: []string{"*"},
				DeniedTools:  []string{"execute_*"},
			},
		},
	})

	chain := middleware.NewChain(
		auth.AuthMiddleware(jwtAuth, auth.AuthMiddlewareConfig{}),
		auth.AuthzMiddleware(rbacAuthz, auth.AuthzMiddlewareConfig{}),
	)

	mp := &mockToolProvider{
		name: "execute_code",
		tool: mcp.Tool{Name: "execute_code"},
	}
	wrapped := chain.Apply(mp)

	tokenStr := createJWT(t, jwt.MapClaims{
		"sub":   "bob",
		"roles": []any{"restricted"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	ctx := auth.WithHeaders(context.Background(), map[string][]string{
		"Authorization": {"Bearer " + tokenStr},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError, "expected tool to be denied")
}

func TestIntegration_FactoryCreatedAuth(t *testing.T) {
	// Test that factory-created authenticators work correctly
	jwtAuth, err := auth.DefaultRegistry.CreateAuthenticator("jwt", map[string]any{
		"secret":          testSecret,
		"principal_claim": "sub",
		"roles_claim":     "roles",
	})
	require.NoError(t, err)

	rbacAuthz, err := auth.DefaultRegistry.CreateAuthorizer("simple_rbac", map[string]any{
		"roles": map[string]any{
			"admin": map[string]any{
				"permissions": []any{"*"},
			},
		},
	})
	require.NoError(t, err)

	chain := middleware.NewChain(
		auth.AuthMiddleware(jwtAuth, auth.AuthMiddlewareConfig{}),
		auth.AuthzMiddleware(rbacAuthz, auth.AuthzMiddlewareConfig{}),
	)

	mp := &mockToolProvider{
		name: "any_tool",
		tool: mcp.Tool{Name: "any_tool"},
	}
	wrapped := chain.Apply(mp)

	tokenStr := createJWT(t, jwt.MapClaims{
		"sub":   "factory-user",
		"roles": []any{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	ctx := auth.WithHeaders(context.Background(), map[string][]string{
		"Authorization": {"Bearer " + tokenStr},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "factory-user", mp.identity.Principal)
}
