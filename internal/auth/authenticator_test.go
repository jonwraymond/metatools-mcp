package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRequest_GetHeader(t *testing.T) {
	req := &AuthRequest{
		Headers: map[string][]string{
			"Authorization": {"Bearer token123"},
			"X-Multi":       {"value1", "value2"},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "existing header",
			key:      "Authorization",
			expected: "Bearer token123",
		},
		{
			name:     "multi-value returns first",
			key:      "X-Multi",
			expected: "value1",
		},
		{
			name:     "missing header",
			key:      "X-Missing",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, req.GetHeader(tt.key))
		})
	}
}

func TestAuthRequest_GetHeader_NilHeaders(t *testing.T) {
	req := &AuthRequest{}
	assert.Equal(t, "", req.GetHeader("Authorization"))
}

func TestAuthenticatorFunc(t *testing.T) {
	called := false
	expectedResult := &AuthResult{
		Authenticated: true,
		Identity:      &Identity{Principal: "test-user"},
	}

	fn := AuthenticatorFunc(func(_ context.Context, _ *AuthRequest) (*AuthResult, error) {
		called = true
		return expectedResult, nil
	})

	assert.Equal(t, "func", fn.Name())
	assert.True(t, fn.Supports(context.Background(), &AuthRequest{}))

	result, err := fn.Authenticate(context.Background(), &AuthRequest{})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, expectedResult, result)
}

func TestAuthResult_Success(t *testing.T) {
	identity := &Identity{Principal: "alice"}
	result := AuthSuccess(identity)

	assert.True(t, result.Authenticated)
	assert.Equal(t, identity, result.Identity)
	assert.Empty(t, result.Challenge)
	assert.NoError(t, result.Error)
}

func TestAuthResult_Failure(t *testing.T) {
	result := AuthFailure(ErrInvalidToken, "Bearer")

	assert.False(t, result.Authenticated)
	assert.Nil(t, result.Identity)
	assert.Equal(t, "Bearer", result.Challenge)
	assert.ErrorIs(t, result.Error, ErrInvalidToken)
}
