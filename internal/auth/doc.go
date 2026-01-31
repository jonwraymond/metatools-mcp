// Package auth provides pluggable authentication and authorization middleware
// for metatools-mcp. It supports JWT tokens, API keys, and composable
// authentication strategies with simple RBAC-based authorization.
//
// # Architecture
//
// The auth package follows a layered architecture:
//
//   - Transport layer: [WithAuthHeaders] extracts HTTP headers into context
//   - Auth middleware: [AuthMiddleware] validates credentials via [Authenticator]
//   - Authz middleware: [AuthzMiddleware] checks permissions via [Authorizer]
//   - Identity: [Identity] represents authenticated principals
//
// # Authenticators
//
// Built-in authenticators:
//
//   - [JWTAuthenticator]: Validates JWT bearer tokens with configurable claims
//   - [APIKeyAuthenticator]: Validates API keys against a store
//   - [CompositeAuthenticator]: Chains multiple authenticators with fallback
//
// # Authorizers
//
// Built-in authorizers:
//
//   - [SimpleRBACAuthorizer]: Role-based access control with tool/action patterns
//   - [AllowAllAuthorizer]: Permits all requests (development/testing)
//   - [DenyAllAuthorizer]: Denies all requests (fail-closed default)
//
// # Factory Registration
//
// Use [DefaultRegistry] to create authenticators and authorizers from configuration:
//
//	auth, _ := auth.DefaultRegistry.CreateAuthenticator("jwt", map[string]any{
//	    "secret": "my-secret-key",
//	    "issuer": "https://auth.example.com",
//	})
//
// # Usage Example
//
//	// Setup authenticator
//	jwtAuth := auth.NewJWTAuthenticator(auth.JWTConfig{
//	    Issuer:         "https://auth.example.com",
//	    PrincipalClaim: "sub",
//	    RolesClaim:     "roles",
//	}, auth.NewStaticKeyProvider([]byte(secret)))
//
//	// Setup authorizer
//	rbacAuthz := auth.NewSimpleRBACAuthorizer(auth.RBACConfig{
//	    Roles: map[string]auth.RoleConfig{
//	        "admin": {Permissions: []string{"*"}},
//	        "user":  {AllowedTools: []string{"search_*", "describe_*"}},
//	    },
//	})
//
//	// Create middleware chain
//	chain := middleware.NewChain(
//	    auth.AuthMiddleware(jwtAuth, auth.AuthMiddlewareConfig{}),
//	    auth.AuthzMiddleware(rbacAuthz, auth.AuthzMiddlewareConfig{}),
//	)
//
//	// Apply to providers
//	wrapped := chain.Apply(myProvider)
package auth
