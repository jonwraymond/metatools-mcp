package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig configures the JWT authenticator.
type JWTConfig struct {
	// Issuer is the expected token issuer (iss claim).
	// If empty, issuer validation is skipped.
	Issuer string

	// Audience is the expected token audience (aud claim).
	// If empty, audience validation is skipped.
	Audience string

	// HeaderName is the HTTP header containing the token.
	// Default: "Authorization"
	HeaderName string

	// TokenPrefix is the prefix before the token value.
	// Default: "Bearer "
	TokenPrefix string

	// ClockSkew allows for clock differences between systems.
	// Default: 0
	ClockSkew time.Duration

	// PrincipalClaim is the claim to use as the identity principal.
	// Default: "sub"
	PrincipalClaim string

	// TenantClaim is the claim containing the tenant ID.
	// If empty, no tenant is extracted.
	TenantClaim string

	// RolesClaim is the claim containing user roles.
	// Expects a string array claim.
	RolesClaim string
}

// KeyProvider retrieves signing keys for JWT validation.
type KeyProvider interface {
	// GetKey returns the key for the given key ID.
	// For HMAC, return []byte. For RSA, return *rsa.PublicKey.
	GetKey(ctx context.Context, keyID string) (any, error)
}

// StaticKeyProvider returns a single static key.
type StaticKeyProvider struct {
	key any
}

// NewStaticKeyProvider creates a key provider with a single key.
func NewStaticKeyProvider(key any) *StaticKeyProvider {
	return &StaticKeyProvider{key: key}
}

// GetKey returns the static key.
func (p *StaticKeyProvider) GetKey(_ context.Context, _ string) (any, error) {
	return p.key, nil
}

// JWTAuthenticator validates JWT bearer tokens.
type JWTAuthenticator struct {
	config      JWTConfig
	keyProvider KeyProvider
}

// NewJWTAuthenticator creates a new JWT authenticator.
func NewJWTAuthenticator(config JWTConfig, keyProvider KeyProvider) *JWTAuthenticator {
	// Apply defaults
	if config.HeaderName == "" {
		config.HeaderName = "Authorization"
	}
	if config.TokenPrefix == "" && config.HeaderName == "Authorization" {
		config.TokenPrefix = "Bearer "
	}
	if config.PrincipalClaim == "" {
		config.PrincipalClaim = "sub"
	}

	return &JWTAuthenticator{
		config:      config,
		keyProvider: keyProvider,
	}
}

// Name returns "jwt".
func (a *JWTAuthenticator) Name() string {
	return "jwt"
}

// Supports returns true if the request contains a bearer token.
func (a *JWTAuthenticator) Supports(_ context.Context, req *AuthRequest) bool {
	header := req.GetHeader(a.config.HeaderName)
	if header == "" {
		return false
	}
	if a.config.TokenPrefix != "" {
		return strings.HasPrefix(header, a.config.TokenPrefix)
	}
	return true
}

// Authenticate validates the JWT and extracts the identity.
func (a *JWTAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	// Extract token from header
	header := req.GetHeader(a.config.HeaderName)
	if header == "" {
		return AuthFailure(ErrMissingCredentials, "Bearer"), nil
	}

	tokenStr := header
	if a.config.TokenPrefix != "" {
		if !strings.HasPrefix(header, a.config.TokenPrefix) {
			return AuthFailure(ErrInvalidToken, "Bearer"), nil
		}
		tokenStr = strings.TrimPrefix(header, a.config.TokenPrefix)
	}

	// Build parser options
	parserOpts := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
	}

	if a.config.Issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(a.config.Issuer))
	}
	if a.config.Audience != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(a.config.Audience))
	}
	if a.config.ClockSkew > 0 {
		parserOpts = append(parserOpts, jwt.WithLeeway(a.config.ClockSkew))
	}

	parser := jwt.NewParser(parserOpts...)

	// Parse and validate token
	token, err := parser.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		// Get key ID from header if present
		keyID := ""
		if kid, ok := t.Header["kid"].(string); ok {
			keyID = kid
		}

		key, err := a.keyProvider.GetKey(ctx, keyID)
		if err != nil {
			return nil, fmt.Errorf("get key: %w", err)
		}

		// Validate signing method matches key type
		switch t.Method.(type) {
		case *jwt.SigningMethodHMAC:
			if _, ok := key.([]byte); !ok {
				return nil, fmt.Errorf("expected []byte key for HMAC")
			}
		case *jwt.SigningMethodRSA:
			if _, ok := key.(*rsa.PublicKey); !ok {
				return nil, fmt.Errorf("expected *rsa.PublicKey for RSA")
			}
		}

		return key, nil
	})

	if err != nil {
		return a.handleParseError(err), nil
	}

	if !token.Valid {
		return AuthFailure(ErrInvalidToken, "Bearer"), nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return AuthFailure(ErrInvalidToken, "Bearer"), nil
	}

	// Build identity from claims
	identity := a.buildIdentity(claims)

	return AuthSuccess(identity), nil
}

func (a *JWTAuthenticator) handleParseError(err error) *AuthResult {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return AuthFailure(ErrTokenExpired, "Bearer")
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return AuthFailure(ErrTokenExpired, "Bearer")
	case errors.Is(err, jwt.ErrTokenMalformed):
		return AuthFailure(ErrInvalidToken, "Bearer")
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return AuthFailure(ErrInvalidToken, "Bearer")
	case errors.Is(err, jwt.ErrTokenInvalidIssuer):
		return AuthFailure(ErrInvalidIssuer, "Bearer")
	case errors.Is(err, jwt.ErrTokenInvalidAudience):
		return AuthFailure(ErrInvalidAudience, "Bearer")
	default:
		return AuthFailure(ErrInvalidToken, "Bearer")
	}
}

func (a *JWTAuthenticator) buildIdentity(claims jwt.MapClaims) *Identity {
	identity := &Identity{
		Method: AuthMethodJWT,
		Claims: map[string]any(claims),
	}

	// Extract principal
	if sub, ok := claims[a.config.PrincipalClaim].(string); ok {
		identity.Principal = sub
	}

	// Extract tenant
	if a.config.TenantClaim != "" {
		if tenant, ok := claims[a.config.TenantClaim].(string); ok {
			identity.TenantID = tenant
		}
	}

	// Extract roles
	if a.config.RolesClaim != "" {
		if roles, ok := claims[a.config.RolesClaim].([]any); ok {
			identity.Roles = make([]string, 0, len(roles))
			for _, r := range roles {
				if s, ok := r.(string); ok {
					identity.Roles = append(identity.Roles, s)
				}
			}
		}
	}

	// Extract expiration
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		identity.ExpiresAt = exp.Time
	}

	// Extract issued at
	if iat, err := claims.GetIssuedAt(); err == nil && iat != nil {
		identity.IssuedAt = iat.Time
	}

	return identity
}
