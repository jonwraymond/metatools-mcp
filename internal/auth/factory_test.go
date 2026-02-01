package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAuthenticator(t *testing.T) {
	reg := NewRegistry()

	factory := func(_ map[string]any) (Authenticator, error) {
		return &JWTAuthenticator{}, nil
	}

	err := reg.RegisterAuthenticator("test", factory)
	require.NoError(t, err)

	// Duplicate registration should fail
	err = reg.RegisterAuthenticator("test", factory)
	assert.Error(t, err)
}

func TestRegistry_GetAuthenticator(t *testing.T) {
	reg := NewRegistry()

	factory := func(cfg map[string]any) (Authenticator, error) {
		secret, _ := cfg["secret"].(string)
		return NewJWTAuthenticator(JWTConfig{}, NewStaticKeyProvider([]byte(secret))), nil
	}

	err := reg.RegisterAuthenticator("jwt", factory)
	require.NoError(t, err)

	auth, err := reg.CreateAuthenticator("jwt", map[string]any{"secret": "test-secret"})
	require.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "jwt", auth.Name())
}

func TestRegistry_GetAuthenticator_NotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.CreateAuthenticator("nonexistent", nil)
	assert.Error(t, err)
}

func TestRegistry_RegisterAuthorizer(t *testing.T) {
	reg := NewRegistry()

	factory := func(_ map[string]any) (Authorizer, error) {
		return AllowAllAuthorizer{}, nil
	}

	err := reg.RegisterAuthorizer("test", factory)
	require.NoError(t, err)

	// Duplicate registration should fail
	err = reg.RegisterAuthorizer("test", factory)
	assert.Error(t, err)
}

func TestRegistry_GetAuthorizer(t *testing.T) {
	reg := NewRegistry()

	factory := func(cfg map[string]any) (Authorizer, error) {
		return NewSimpleRBACAuthorizer(RBACConfig{}), nil
	}

	err := reg.RegisterAuthorizer("simple_rbac", factory)
	require.NoError(t, err)

	authz, err := reg.CreateAuthorizer("simple_rbac", nil)
	require.NoError(t, err)
	assert.NotNil(t, authz)
	assert.Equal(t, "simple_rbac", authz.Name())
}

func TestDefaultRegistry_JWT(t *testing.T) {
	// Verify JWT factory is registered in default registry
	auth, err := DefaultRegistry.CreateAuthenticator("jwt", map[string]any{
		"secret": "test-secret-key-32-chars-minimum",
	})
	require.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "jwt", auth.Name())
}

func TestDefaultRegistry_APIKey(t *testing.T) {
	// Verify API key factory is registered in default registry
	auth, err := DefaultRegistry.CreateAuthenticator("api_key", nil)
	require.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, "api_key", auth.Name())
}

func TestDefaultRegistry_SimpleRBAC(t *testing.T) {
	// Verify simple RBAC factory is registered in default registry
	authz, err := DefaultRegistry.CreateAuthorizer("simple_rbac", map[string]any{
		"default_role": "anonymous",
	})
	require.NoError(t, err)
	assert.NotNil(t, authz)
	assert.Equal(t, "simple_rbac", authz.Name())
}

func TestDefaultRegistry_AllowAll(t *testing.T) {
	authz, err := DefaultRegistry.CreateAuthorizer("allow_all", nil)
	require.NoError(t, err)
	assert.NotNil(t, authz)
	assert.Equal(t, "allow_all", authz.Name())
}

func TestDefaultRegistry_DenyAll(t *testing.T) {
	authz, err := DefaultRegistry.CreateAuthorizer("deny_all", nil)
	require.NoError(t, err)
	assert.NotNil(t, authz)
	assert.Equal(t, "deny_all", authz.Name())
}

func TestRegistry_ListAuthenticators(t *testing.T) {
	reg := NewRegistry()
	_ = reg.RegisterAuthenticator("jwt", func(_ map[string]any) (Authenticator, error) {
		return &JWTAuthenticator{}, nil
	})
	_ = reg.RegisterAuthenticator("api_key", func(_ map[string]any) (Authenticator, error) {
		return &APIKeyAuthenticator{}, nil
	})

	names := reg.ListAuthenticators()
	assert.Contains(t, names, "jwt")
	assert.Contains(t, names, "api_key")
}

func TestRegistry_ListAuthorizers(t *testing.T) {
	reg := NewRegistry()
	_ = reg.RegisterAuthorizer("simple_rbac", func(_ map[string]any) (Authorizer, error) {
		return &SimpleRBACAuthorizer{}, nil
	})
	_ = reg.RegisterAuthorizer("allow_all", func(cfg map[string]any) (Authorizer, error) {
		return AllowAllAuthorizer{}, nil
	})

	names := reg.ListAuthorizers()
	assert.Contains(t, names, "simple_rbac")
	assert.Contains(t, names, "allow_all")
}
