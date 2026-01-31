// Package auth provides pluggable authentication and authorization middleware
// for metatools-mcp. It supports JWT tokens, API keys, and composable
// authentication strategies with simple RBAC-based authorization.
package auth

import (
	"slices"
	"time"
)

// AuthMethod identifies the authentication mechanism used.
type AuthMethod int

const (
	// AuthMethodNone indicates no authentication (anonymous access).
	AuthMethodNone AuthMethod = iota
	// AuthMethodJWT indicates JWT bearer token authentication.
	AuthMethodJWT
	// AuthMethodAPIKey indicates API key authentication.
	AuthMethodAPIKey
	// AuthMethodMTLS indicates mutual TLS client certificate authentication.
	AuthMethodMTLS
	// AuthMethodOAuth2 indicates OAuth2 token authentication.
	AuthMethodOAuth2
)

// String returns the string representation of the auth method.
func (m AuthMethod) String() string {
	switch m {
	case AuthMethodNone:
		return "none"
	case AuthMethodJWT:
		return "jwt"
	case AuthMethodAPIKey:
		return "api_key"
	case AuthMethodMTLS:
		return "mtls"
	case AuthMethodOAuth2:
		return "oauth2"
	default:
		return "unknown"
	}
}

// Identity represents an authenticated principal with associated attributes.
// It is placed in context after successful authentication and can be used
// by authorization middleware and handlers to make access control decisions.
type Identity struct {
	// Principal is the unique identifier for the authenticated entity
	// (e.g., user ID, service account name, API key ID).
	Principal string

	// TenantID identifies the tenant context for multi-tenant deployments.
	// Empty if not applicable.
	TenantID string

	// Roles are the assigned roles for RBAC-based authorization.
	Roles []string

	// Permissions are explicit permissions granted to this identity.
	Permissions []string

	// Claims contains additional authentication claims from the token
	// or authentication source (e.g., JWT claims).
	Claims map[string]any

	// Method indicates how this identity was authenticated.
	Method AuthMethod

	// ExpiresAt is when this identity's authentication expires.
	// Zero value means no expiration.
	ExpiresAt time.Time

	// IssuedAt is when the identity was authenticated.
	IssuedAt time.Time
}

// HasRole checks if the identity has the specified role.
func (id *Identity) HasRole(role string) bool {
	return slices.Contains(id.Roles, role)
}

// HasAnyRole checks if the identity has any of the specified roles.
func (id *Identity) HasAnyRole(roles ...string) bool {
	return slices.ContainsFunc(roles, id.HasRole)
}

// HasPermission checks if the identity has the specified permission.
func (id *Identity) HasPermission(permission string) bool {
	return slices.Contains(id.Permissions, permission)
}

// IsExpired checks if the identity has expired.
// Returns false if ExpiresAt is zero (no expiration).
func (id *Identity) IsExpired() bool {
	if id.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(id.ExpiresAt)
}

// IsAnonymous returns true if this identity represents an unauthenticated user.
func (id *Identity) IsAnonymous() bool {
	return id.Method == AuthMethodNone
}

// AnonymousIdentity returns a new anonymous identity.
func AnonymousIdentity() *Identity {
	return &Identity{
		Principal: "anonymous",
		Method:    AuthMethodNone,
		IssuedAt:  time.Now(),
	}
}
