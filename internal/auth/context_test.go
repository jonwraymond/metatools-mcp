package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithIdentity(t *testing.T) {
	ctx := context.Background()
	identity := &Identity{
		Principal: "alice",
		TenantID:  "tenant-1",
		Roles:     []string{"admin"},
	}

	newCtx := WithIdentity(ctx, identity)
	require.NotNil(t, newCtx)

	// Original context should not have identity
	assert.Nil(t, IdentityFromContext(ctx))

	// New context should have identity
	retrieved := IdentityFromContext(newCtx)
	assert.Equal(t, identity, retrieved)
}

func TestIdentityFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected *Identity
	}{
		{
			name:     "no identity",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name: "with identity",
			ctx: WithIdentity(context.Background(), &Identity{
				Principal: "bob",
			}),
			expected: &Identity{Principal: "bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IdentityFromContext(tt.ctx)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Principal, result.Principal)
			}
		})
	}
}

func TestPrincipalFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "no identity",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name: "with identity",
			ctx: WithIdentity(context.Background(), &Identity{
				Principal: "alice",
			}),
			expected: "alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PrincipalFromContext(tt.ctx))
		})
	}
}

func TestTenantIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "no identity",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name: "with identity",
			ctx: WithIdentity(context.Background(), &Identity{
				TenantID: "tenant-123",
			}),
			expected: "tenant-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, TenantIDFromContext(tt.ctx))
		})
	}
}

func TestWithHeaders(t *testing.T) {
	ctx := context.Background()
	headers := map[string][]string{
		"Authorization": {"Bearer token123"},
		"X-API-Key":     {"key456"},
	}

	newCtx := WithHeaders(ctx, headers)
	require.NotNil(t, newCtx)

	// Original context should not have headers
	assert.Nil(t, HeadersFromContext(ctx))

	// New context should have headers
	retrieved := HeadersFromContext(newCtx)
	assert.Equal(t, headers, retrieved)
}

func TestHeadersFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected map[string][]string
	}{
		{
			name:     "no headers",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name: "with headers",
			ctx: WithHeaders(context.Background(), map[string][]string{
				"Content-Type": {"application/json"},
			}),
			expected: map[string][]string{
				"Content-Type": {"application/json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HeadersFromContext(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHeader(t *testing.T) {
	headers := map[string][]string{
		"Authorization": {"Bearer token123"},
		"X-Multi":       {"value1", "value2"},
	}
	ctx := WithHeaders(context.Background(), headers)

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
			name:     "multi-value header returns first",
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
			assert.Equal(t, tt.expected, GetHeader(ctx, tt.key))
		})
	}
}

func TestGetHeader_NoHeaders(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", GetHeader(ctx, "Authorization"))
}
