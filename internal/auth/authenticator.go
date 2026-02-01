//revive:disable:exported // Auth-prefixed types provide clarity across packages in the stack.
package auth

import "context"

// Authenticator validates credentials and produces an identity.
// Implementations should be safe for concurrent use.
type Authenticator interface {
	// Authenticate validates the request and returns an authentication result.
	// Returns an error only for unexpected failures (e.g., network errors).
	// Authentication failures should be returned via AuthResult.Error.
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error)

	// Name returns a unique identifier for this authenticator.
	Name() string

	// Supports returns true if this authenticator can handle the given request.
	// This allows composite authenticators to skip unsuitable authenticators.
	Supports(ctx context.Context, req *AuthRequest) bool
}

// AuthRequest contains the information needed to authenticate a request.
type AuthRequest struct {
	// Headers contains HTTP headers from the request.
	Headers map[string][]string

	// Method is the HTTP method of the request.
	Method string

	// Resource is the requested resource path.
	Resource string

	// RemoteAddr is the client's network address.
	RemoteAddr string
}

// GetHeader returns the first value for the given header key.
// Returns empty string if the header is not present.
func (r *AuthRequest) GetHeader(key string) string {
	if r.Headers == nil {
		return ""
	}
	values := r.Headers[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// AuthResult contains the outcome of an authentication attempt.
type AuthResult struct {
	// Authenticated is true if authentication succeeded.
	Authenticated bool

	// Identity is the authenticated identity. Nil if authentication failed.
	Identity *Identity

	// Error contains the authentication error if Authenticated is false.
	// May be nil for anonymous access scenarios.
	Error error

	// Challenge is the WWW-Authenticate challenge to send to the client
	// when authentication fails (e.g., "Bearer", "Basic realm=\"api\"").
	Challenge string
}

// AuthSuccess creates a successful authentication result.
func AuthSuccess(identity *Identity) *AuthResult {
	return &AuthResult{
		Authenticated: true,
		Identity:      identity,
	}
}

// AuthFailure creates a failed authentication result.
func AuthFailure(err error, challenge string) *AuthResult {
	return &AuthResult{
		Authenticated: false,
		Error:         err,
		Challenge:     challenge,
	}
}

// AuthenticatorFunc is an adapter to allow use of ordinary functions as Authenticators.
type AuthenticatorFunc func(ctx context.Context, req *AuthRequest) (*AuthResult, error)

// Authenticate calls the function.
func (f AuthenticatorFunc) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	return f(ctx, req)
}

// Name returns "func" for function-based authenticators.
func (f AuthenticatorFunc) Name() string {
	return "func"
}

// Supports always returns true for function-based authenticators.
func (f AuthenticatorFunc) Supports(_ context.Context, _ *AuthRequest) bool {
	return true
}
