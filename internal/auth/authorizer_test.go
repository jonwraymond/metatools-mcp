package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthzError_Error(t *testing.T) {
	err := &AuthzError{
		Subject:  "alice",
		Resource: "tool:search_tools",
		Action:   "call",
		Reason:   "permission denied",
	}

	msg := err.Error()
	assert.Contains(t, msg, "alice")
	assert.Contains(t, msg, "tool:search_tools")
	assert.Contains(t, msg, "call")
	assert.Contains(t, msg, "permission denied")
}

func TestAuthzError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AuthzError{
		Reason: "test",
		Cause:  cause,
	}

	assert.ErrorIs(t, err, cause)
}

func TestAuthzError_Is(t *testing.T) {
	err := &AuthzError{Reason: "denied"}

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestAllowAllAuthorizer(t *testing.T) {
	authz := AllowAllAuthorizer{}

	req := &AuthzRequest{
		Subject:  &Identity{Principal: "anyone"},
		Resource: "anything",
		Action:   "do",
	}

	err := authz.Authorize(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "allow_all", authz.Name())
}

func TestDenyAllAuthorizer(t *testing.T) {
	authz := DenyAllAuthorizer{}

	req := &AuthzRequest{
		Subject:  &Identity{Principal: "anyone"},
		Resource: "anything",
		Action:   "do",
	}

	err := authz.Authorize(context.Background(), req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, "deny_all", authz.Name())
}

func TestAuthzRequest_ToolName(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		expected string
	}{
		{
			name:     "with tool prefix",
			resource: "tool:search_tools",
			expected: "search_tools",
		},
		{
			name:     "without prefix",
			resource: "search_tools",
			expected: "search_tools",
		},
		{
			name:     "empty",
			resource: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AuthzRequest{Resource: tt.resource}
			assert.Equal(t, tt.expected, req.ToolName())
		})
	}
}

func TestAuthorizerFunc(t *testing.T) {
	called := false

	fn := AuthorizerFunc(func(ctx context.Context, req *AuthzRequest) error {
		called = true
		return nil
	})

	assert.Equal(t, "func", fn.Name())

	err := fn.Authorize(context.Background(), &AuthzRequest{})
	require.NoError(t, err)
	assert.True(t, called)
}
