package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithAuthHeaders_Middleware(t *testing.T) {
	// Create a test handler that checks context headers
	var capturedHeaders map[string][]string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = HeadersFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth headers middleware
	wrapped := WithAuthHeaders(inner)

	// Create test request with headers
	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Key", "test-api-key")
	req.Header.Set("X-Tenant-ID", "tenant-123")

	recorder := httptest.NewRecorder()
	wrapped.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotNil(t, capturedHeaders)
	assert.Equal(t, []string{"Bearer test-token"}, capturedHeaders["Authorization"])
	assert.Equal(t, []string{"test-api-key"}, capturedHeaders["X-Api-Key"])
	assert.Equal(t, []string{"tenant-123"}, capturedHeaders["X-Tenant-Id"])
}

func TestWithAuthHeaders_PreservesExistingContext(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should have headers in context
		headers := HeadersFromContext(r.Context())
		assert.NotNil(t, headers)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := WithAuthHeaders(inner)

	req := httptest.NewRequest("GET", "/mcp", nil)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	wrapped.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
