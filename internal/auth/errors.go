package auth

import "errors"

// Standard authentication errors.
var (
	// ErrUnauthorized indicates the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidToken indicates the provided token is malformed or invalid.
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired indicates the provided token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrMissingCredentials indicates no credentials were provided.
	ErrMissingCredentials = errors.New("missing credentials")

	// ErrInvalidCredentials indicates the provided credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrForbidden indicates the authenticated user lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidIssuer indicates the token issuer is not trusted.
	ErrInvalidIssuer = errors.New("invalid issuer")

	// ErrInvalidAudience indicates the token audience is not valid.
	ErrInvalidAudience = errors.New("invalid audience")

	// ErrKeyNotFound indicates the requested signing key was not found.
	ErrKeyNotFound = errors.New("key not found")

	// ErrIntrospectionFailed indicates the OAuth2 token introspection request failed.
	ErrIntrospectionFailed = errors.New("introspection failed")

	// ErrTokenInactive indicates the OAuth2 token is not active.
	ErrTokenInactive = errors.New("token inactive")
)
