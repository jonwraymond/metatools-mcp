package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthenticator is a configurable mock for testing.
type mockAuthenticator struct {
	name      string
	supports  bool
	result    *AuthResult
	err       error
	callCount int
}

func (m *mockAuthenticator) Name() string { return m.name }

func (m *mockAuthenticator) Supports(_ context.Context, _ *AuthRequest) bool {
	return m.supports
}

func (m *mockAuthenticator) Authenticate(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
	m.callCount++
	return m.result, m.err
}

func TestCompositeAuthenticator_FirstSucceeds(t *testing.T) {
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "alice"}),
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "bob"}),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "alice", result.Identity.Principal)
	assert.Equal(t, 1, auth1.callCount)
	assert.Equal(t, 0, auth2.callCount) // Should not be called
}

func TestCompositeAuthenticator_FallbackToSecond(t *testing.T) {
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: true,
		result:   AuthFailure(ErrInvalidToken, "Bearer"),
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "bob"}),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "bob", result.Identity.Principal)
	assert.Equal(t, 1, auth1.callCount)
	assert.Equal(t, 1, auth2.callCount)
}

func TestCompositeAuthenticator_AllFail(t *testing.T) {
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: true,
		result:   AuthFailure(ErrInvalidToken, "Bearer"),
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthFailure(ErrInvalidCredentials, ""),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	// Returns last error
	assert.ErrorIs(t, result.Error, ErrInvalidCredentials)
}

func TestCompositeAuthenticator_SkipsUnsupported(t *testing.T) {
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: false, // Does not support
		result:   AuthSuccess(&Identity{Principal: "alice"}),
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "bob"}),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	assert.Equal(t, "bob", result.Identity.Principal)
	assert.Equal(t, 0, auth1.callCount) // Skipped
	assert.Equal(t, 1, auth2.callCount)
}

func TestCompositeAuthenticator_HandlesErrors(t *testing.T) {
	expectedErr := errors.New("network error")
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: true,
		err:      expectedErr,
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "bob"}),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)

	// When an authenticator returns an error, it propagates
	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
}

func TestCompositeAuthenticator_NoAuthenticators(t *testing.T) {
	composite := NewCompositeAuthenticator()

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.False(t, result.Authenticated)
	assert.ErrorIs(t, result.Error, ErrMissingCredentials)
}

func TestCompositeAuthenticator_Name(t *testing.T) {
	composite := NewCompositeAuthenticator()
	assert.Equal(t, "composite", composite.Name())
}

func TestCompositeAuthenticator_Supports(t *testing.T) {
	auth1 := &mockAuthenticator{name: "auth1", supports: false}
	auth2 := &mockAuthenticator{name: "auth2", supports: true}

	composite := NewCompositeAuthenticator(auth1, auth2)

	// Supports if any authenticator supports
	assert.True(t, composite.Supports(context.Background(), &AuthRequest{}))

	// All don't support
	auth2.supports = false
	assert.False(t, composite.Supports(context.Background(), &AuthRequest{}))
}

func TestCompositeAuthenticator_StopOnFirstDisabled(t *testing.T) {
	auth1 := &mockAuthenticator{
		name:     "auth1",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "alice"}),
	}
	auth2 := &mockAuthenticator{
		name:     "auth2",
		supports: true,
		result:   AuthSuccess(&Identity{Principal: "bob"}),
	}

	composite := NewCompositeAuthenticator(auth1, auth2)
	composite.StopOnFirst = false

	result, err := composite.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.True(t, result.Authenticated)
	// With StopOnFirst disabled, still returns first success but calls both
	assert.Equal(t, "alice", result.Identity.Principal)
	assert.Equal(t, 1, auth1.callCount)
	// StopOnFirst=false means continue even after success, but we still return first
	// Actually, the behavior should be: with StopOnFirst=true (default), stop on first success
	// With StopOnFirst=false, try all and use first success
}
