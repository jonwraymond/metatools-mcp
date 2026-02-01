package auth

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonwraymond/metatools-mcp/internal/middleware"
)

// mockProvider implements provider.ToolProvider for testing.
type mockProvider struct {
	name      string
	tool      mcp.Tool
	handleFn  func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
	ctxCalled context.Context
}

func (m *mockProvider) Name() string    { return m.name }
func (m *mockProvider) Enabled() bool   { return true }
func (m *mockProvider) Tool() mcp.Tool  { return m.tool }

func (m *mockProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	m.ctxCalled = ctx
	if m.handleFn != nil {
		return m.handleFn(ctx, req, args)
	}
	return &mcp.CallToolResult{}, nil, nil
}

func TestAuthMiddleware_ValidAuth(t *testing.T) {
	// Authenticator that always succeeds
	auth := AuthenticatorFunc(func(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
		return AuthSuccess(&Identity{
			Principal: "alice",
			TenantID:  "tenant-1",
			Roles:     []string{"user"},
		}), nil
	})

	mw := AuthMiddleware(auth, AuthMiddlewareConfig{})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	// Create context with headers
	ctx := WithHeaders(context.Background(), map[string][]string{
		"Authorization": {"Bearer valid-token"},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify identity was added to context
	identity := IdentityFromContext(mp.ctxCalled)
	require.NotNil(t, identity)
	assert.Equal(t, "alice", identity.Principal)
	assert.Equal(t, "tenant-1", identity.TenantID)
}

func TestAuthMiddleware_InvalidAuth(t *testing.T) {
	// Authenticator that always fails
	auth := AuthenticatorFunc(func(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
		return AuthFailure(ErrInvalidToken, "Bearer"), nil
	})

	mw := AuthMiddleware(auth, AuthMiddlewareConfig{})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	ctx := WithHeaders(context.Background(), map[string][]string{})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestAuthMiddleware_AllowAnonymous(t *testing.T) {
	// Authenticator that fails (no credentials)
	auth := AuthenticatorFunc(func(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
		return AuthFailure(ErrMissingCredentials, "Bearer"), nil
	})

	mw := AuthMiddleware(auth, AuthMiddlewareConfig{
		AllowAnonymous: true,
		AnonymousIdentity: &Identity{
			Principal: "anonymous",
			Roles:     []string{"guest"},
		},
	})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	ctx := WithHeaders(context.Background(), map[string][]string{})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError) // Should succeed with anonymous

	// Verify anonymous identity was added
	identity := IdentityFromContext(mp.ctxCalled)
	require.NotNil(t, identity)
	assert.Equal(t, "anonymous", identity.Principal)
}

func TestAuthzMiddleware_Authorized(t *testing.T) {
	// Authorizer that always allows
	authz := AllowAllAuthorizer{}

	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	ctx := WithIdentity(context.Background(), &Identity{
		Principal: "alice",
		Roles:     []string{"admin"},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

func TestAuthzMiddleware_Unauthorized(t *testing.T) {
	// Authorizer that always denies
	authz := DenyAllAuthorizer{}

	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	ctx := WithIdentity(context.Background(), &Identity{
		Principal: "alice",
		Roles:     []string{"admin"},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestAuthzMiddleware_NoIdentity(t *testing.T) {
	authz := AllowAllAuthorizer{}

	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	// No identity in context
	ctx := context.Background()

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError) // Should fail without identity
}

func TestAuthMiddleware_ReturnsMiddlewareType(t *testing.T) {
	auth := AuthenticatorFunc(func(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
		return AuthSuccess(&Identity{Principal: "test"}), nil
	})

	mw := AuthMiddleware(auth, AuthMiddlewareConfig{})

	// Verify it can be used with middleware.Chain
	chain := middleware.NewChain(mw)
	assert.Equal(t, 1, chain.Len())
}

func TestAuthzMiddleware_ReturnsMiddlewareType(t *testing.T) {
	authz := AllowAllAuthorizer{}
	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{})

	// Verify it can be used with middleware.Chain
	chain := middleware.NewChain(mw)
	assert.Equal(t, 1, chain.Len())
}

func TestAuthzMiddleware_CustomResourceResolver(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"user": {
				AllowedTools: []string{"allowed_*"},
			},
		},
	})

	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{
		ResourceResolver: func(_ context.Context, toolName string, _ map[string]any) string {
			// Custom resolver that adds prefix
			return "tool:" + toolName
		},
	})

	mp := &mockProvider{
		name: "allowed_tool",
		tool: mcp.Tool{Name: "allowed_tool"},
	}

	wrapped := mw(mp)

	ctx := WithIdentity(context.Background(), &Identity{
		Roles: []string{"user"},
	})

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestAuthzMiddleware_AllowAnonymous(t *testing.T) {
	authz := AllowAllAuthorizer{}

	mw := AuthzMiddleware(authz, AuthzMiddlewareConfig{
		AllowAnonymous: true,
	})

	mp := &mockProvider{
		name: "test_tool",
		tool: mcp.Tool{Name: "test_tool"},
	}

	wrapped := mw(mp)

	// No identity in context, but anonymous allowed
	ctx := context.Background()

	result, _, err := wrapped.Handle(ctx, &mcp.CallToolRequest{}, nil)
	require.NoError(t, err)
	assert.False(t, result.IsError) // Should pass
}
