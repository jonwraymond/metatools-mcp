package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIdentity_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		identity *Identity
		role     string
		expected bool
	}{
		{
			name:     "has role",
			identity: &Identity{Roles: []string{"admin", "user"}},
			role:     "admin",
			expected: true,
		},
		{
			name:     "does not have role",
			identity: &Identity{Roles: []string{"user"}},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty roles",
			identity: &Identity{Roles: nil},
			role:     "admin",
			expected: false,
		},
		{
			name:     "case sensitive",
			identity: &Identity{Roles: []string{"Admin"}},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.identity.HasRole(tt.role))
		})
	}
}

func TestIdentity_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		identity   *Identity
		permission string
		expected   bool
	}{
		{
			name:       "has permission",
			identity:   &Identity{Permissions: []string{"read", "write"}},
			permission: "read",
			expected:   true,
		},
		{
			name:       "does not have permission",
			identity:   &Identity{Permissions: []string{"read"}},
			permission: "write",
			expected:   false,
		},
		{
			name:       "empty permissions",
			identity:   &Identity{Permissions: nil},
			permission: "read",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.identity.HasPermission(tt.permission))
		})
	}
}

func TestIdentity_HasAnyRole(t *testing.T) {
	tests := []struct {
		name     string
		identity *Identity
		roles    []string
		expected bool
	}{
		{
			name:     "has one of many",
			identity: &Identity{Roles: []string{"user"}},
			roles:    []string{"admin", "user"},
			expected: true,
		},
		{
			name:     "has none",
			identity: &Identity{Roles: []string{"guest"}},
			roles:    []string{"admin", "user"},
			expected: false,
		},
		{
			name:     "empty check roles",
			identity: &Identity{Roles: []string{"admin"}},
			roles:    []string{},
			expected: false,
		},
		{
			name:     "has all",
			identity: &Identity{Roles: []string{"admin", "user"}},
			roles:    []string{"admin", "user"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.identity.HasAnyRole(tt.roles...))
		})
	}
}

func TestIdentity_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		identity *Identity
		expected bool
	}{
		{
			name:     "not expired - future",
			identity: &Identity{ExpiresAt: time.Now().Add(time.Hour)},
			expected: false,
		},
		{
			name:     "expired - past",
			identity: &Identity{ExpiresAt: time.Now().Add(-time.Hour)},
			expected: true,
		},
		{
			name:     "not expired - zero time",
			identity: &Identity{ExpiresAt: time.Time{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.identity.IsExpired())
		})
	}
}

func TestAuthMethod_String(t *testing.T) {
	tests := []struct {
		method   AuthMethod
		expected string
	}{
		{AuthMethodNone, "none"},
		{AuthMethodJWT, "jwt"},
		{AuthMethodAPIKey, "api_key"},
		{AuthMethodMTLS, "mtls"},
		{AuthMethodOAuth2, "oauth2"},
		{AuthMethod(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.String())
		})
	}
}

func TestIdentity_IsAnonymous(t *testing.T) {
	tests := []struct {
		name     string
		identity *Identity
		expected bool
	}{
		{
			name:     "anonymous with none method",
			identity: &Identity{Method: AuthMethodNone},
			expected: true,
		},
		{
			name:     "not anonymous with JWT",
			identity: &Identity{Method: AuthMethodJWT},
			expected: false,
		},
		{
			name:     "not anonymous with API key",
			identity: &Identity{Method: AuthMethodAPIKey},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.identity.IsAnonymous())
		})
	}
}
