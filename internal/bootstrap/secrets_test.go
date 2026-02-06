package bootstrap

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolops/secret"
)

type stubProvider struct{}

func (p *stubProvider) Name() string { return "stub" }
func (p *stubProvider) Resolve(_ context.Context, ref string) (string, error) {
	if ref == "empty" {
		return "", nil
	}
	return "resolved-" + ref, nil
}
func (p *stubProvider) Close() error { return nil }

func registerStub(reg *secret.Registry) error {
	return reg.Register("stub", func(map[string]any) (secret.Provider, error) {
		return &stubProvider{}, nil
	})
}

func TestSecretsBootstrap_NoProviders_DoesNotFail(t *testing.T) {
	resolver, closeFn, err := NewSecretResolver(config.SecretsConfig{Strict: true})
	if err != nil {
		t.Fatalf("NewSecretResolver returned error: %v", err)
	}
	if resolver == nil {
		t.Fatalf("expected resolver")
	}
	if closeFn == nil {
		t.Fatalf("expected closeFn")
	}
	if err := closeFn(); err != nil {
		t.Fatalf("closeFn returned error: %v", err)
	}
}

func TestSecretsBootstrap_ProviderEnabledButNotRegistered_Fails(t *testing.T) {
	_, _, err := NewSecretResolver(config.SecretsConfig{
		Strict: true,
		Providers: map[string]config.SecretProviderConfig{
			"bws": {Enabled: true, Config: map[string]any{}},
		},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSecretsResolve_BackendHeadersAndURL(t *testing.T) {
	ctx := context.Background()

	resolver, closeFn, err := NewSecretResolver(config.SecretsConfig{
		Strict: true,
		Providers: map[string]config.SecretProviderConfig{
			"stub": {Enabled: true, Config: map[string]any{}},
		},
	}, registerStub)
	if err != nil {
		t.Fatalf("NewSecretResolver returned error: %v", err)
	}
	defer func() { _ = closeFn() }()

	backends := []config.MCPBackendConfig{
		{
			Name: "example",
			URL:  "https://example.com/mcp?token=secretref:stub:abc",
			Headers: map[string]string{
				"Authorization": "Bearer secretref:stub:def",
			},
		},
	}

	resolved, err := ResolveMCPBackendConfigs(ctx, resolver, backends)
	if err != nil {
		t.Fatalf("ResolveMCPBackendConfigs returned error: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(resolved))
	}
	if resolved[0].URL != "https://example.com/mcp?token=resolved-abc" {
		t.Fatalf("unexpected url: %q", resolved[0].URL)
	}
	if got := resolved[0].Headers["Authorization"]; got != "Bearer resolved-def" {
		t.Fatalf("unexpected auth header: %q", got)
	}
}

func TestSecretsResolve_StrictMode_EmptySecretFails(t *testing.T) {
	ctx := context.Background()

	resolver, closeFn, err := NewSecretResolver(config.SecretsConfig{
		Strict: true,
		Providers: map[string]config.SecretProviderConfig{
			"stub": {Enabled: true, Config: map[string]any{}},
		},
	}, registerStub)
	if err != nil {
		t.Fatalf("NewSecretResolver returned error: %v", err)
	}
	defer func() { _ = closeFn() }()

	backends := []config.MCPBackendConfig{
		{Name: "example", URL: "secretref:stub:empty"},
	}

	_, err = ResolveMCPBackendConfigs(ctx, resolver, backends)
	if err == nil {
		t.Fatalf("expected error")
	}
}

