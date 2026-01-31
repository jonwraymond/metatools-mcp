package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// APIKeyConfig configures the API key authenticator.
type APIKeyConfig struct {
	// HeaderName is the HTTP header containing the API key.
	// Default: "X-API-Key"
	HeaderName string

	// HashAlgorithm is the algorithm used to hash keys.
	// Default: "sha256"
	HashAlgorithm string
}

// APIKeyStore provides API key lookup functionality.
type APIKeyStore interface {
	// Lookup retrieves API key info by key hash.
	// Returns nil if the key is not found.
	Lookup(ctx context.Context, keyHash string) (*APIKeyInfo, error)
}

// APIKeyInfo contains metadata about an API key.
type APIKeyInfo struct {
	// ID is the unique identifier for this key.
	ID string

	// KeyHash is the SHA256 hash of the API key.
	KeyHash string

	// Principal is the identity associated with this key.
	Principal string

	// TenantID is the tenant this key belongs to.
	TenantID string

	// Roles are the roles granted to this key.
	Roles []string

	// Permissions are explicit permissions for this key.
	Permissions []string

	// ExpiresAt is when this key expires. Nil means no expiration.
	ExpiresAt *time.Time

	// Metadata contains additional key attributes.
	Metadata map[string]string
}

// IsExpired checks if the API key has expired.
func (k *APIKeyInfo) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// MemoryAPIKeyStore is an in-memory API key store for development and testing.
type MemoryAPIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]*APIKeyInfo
}

// NewMemoryAPIKeyStore creates a new in-memory API key store.
func NewMemoryAPIKeyStore() *MemoryAPIKeyStore {
	return &MemoryAPIKeyStore{
		keys: make(map[string]*APIKeyInfo),
	}
}

// Add stores an API key.
func (s *MemoryAPIKeyStore) Add(info *APIKeyInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.keys[info.KeyHash]; exists {
		return errors.New("key already exists")
	}

	s.keys[info.KeyHash] = info
	return nil
}

// Lookup retrieves an API key by hash.
func (s *MemoryAPIKeyStore) Lookup(_ context.Context, keyHash string) (*APIKeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.keys[keyHash]
	if !ok {
		return nil, nil
	}
	return info, nil
}

// Remove deletes an API key by hash.
func (s *MemoryAPIKeyStore) Remove(keyHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, keyHash)
}

// List returns all stored keys.
func (s *MemoryAPIKeyStore) List() []*APIKeyInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*APIKeyInfo, 0, len(s.keys))
	for _, info := range s.keys {
		result = append(result, info)
	}
	return result
}

// APIKeyAuthenticator validates API keys.
type APIKeyAuthenticator struct {
	config APIKeyConfig
	store  APIKeyStore
}

// NewAPIKeyAuthenticator creates a new API key authenticator.
func NewAPIKeyAuthenticator(config APIKeyConfig, store APIKeyStore) *APIKeyAuthenticator {
	// Apply defaults
	if config.HeaderName == "" {
		config.HeaderName = "X-API-Key"
	}
	if config.HashAlgorithm == "" {
		config.HashAlgorithm = "sha256"
	}

	return &APIKeyAuthenticator{
		config: config,
		store:  store,
	}
}

// Name returns "api_key".
func (a *APIKeyAuthenticator) Name() string {
	return "api_key"
}

// Supports returns true if the request contains an API key header.
func (a *APIKeyAuthenticator) Supports(_ context.Context, req *AuthRequest) bool {
	return req.GetHeader(a.config.HeaderName) != ""
}

// Authenticate validates the API key and returns an identity.
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResult, error) {
	rawKey := req.GetHeader(a.config.HeaderName)
	if rawKey == "" {
		return AuthFailure(ErrMissingCredentials, ""), nil
	}

	// Hash the provided key
	keyHash := HashAPIKey(rawKey, a.config.HashAlgorithm)

	// Lookup the key
	info, err := a.store.Lookup(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	if info == nil {
		return AuthFailure(ErrInvalidCredentials, ""), nil
	}

	// Check expiration
	if info.IsExpired() {
		return AuthFailure(ErrTokenExpired, ""), nil
	}

	// Build identity
	identity := &Identity{
		Principal:   info.Principal,
		TenantID:    info.TenantID,
		Roles:       info.Roles,
		Permissions: info.Permissions,
		Method:      AuthMethodAPIKey,
		IssuedAt:    time.Now(),
		Claims: map[string]any{
			"key_id": info.ID,
		},
	}

	if info.ExpiresAt != nil {
		identity.ExpiresAt = *info.ExpiresAt
	}

	return AuthSuccess(identity), nil
}

// HashAPIKey computes the hash of an API key.
func HashAPIKey(key, algorithm string) string {
	switch algorithm {
	case "sha256":
		h := sha256.Sum256([]byte(key))
		return hex.EncodeToString(h[:])
	default:
		// Default to SHA256
		h := sha256.Sum256([]byte(key))
		return hex.EncodeToString(h[:])
	}
}

// ValidateAPIKey checks if a raw key matches a stored hash.
// Uses constant-time comparison to prevent timing attacks.
func ValidateAPIKey(rawKey, storedHash, algorithm string) bool {
	computedHash := HashAPIKey(rawKey, algorithm)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) == 1
}
