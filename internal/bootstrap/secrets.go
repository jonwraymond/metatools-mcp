package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolops/secret"
)

// SecretRegistrar registers secret providers into a registry.
type SecretRegistrar func(reg *secret.Registry) error

// NewSecretResolver builds a secret resolver from configuration and a set of registrars.
//
// Providers are created only when enabled. When strict mode is enabled, providers returning an empty
// string will cause resolution to fail.
func NewSecretResolver(cfg config.SecretsConfig, registrars ...SecretRegistrar) (*secret.Resolver, func() error, error) {
	reg := secret.NewRegistry()
	for _, registrar := range registrars {
		if registrar == nil {
			continue
		}
		if err := registrar(reg); err != nil {
			return nil, nil, err
		}
	}

	enabled := make([]string, 0, len(cfg.Providers))
	for name, pcfg := range cfg.Providers {
		if !pcfg.Enabled {
			continue
		}
		enabled = append(enabled, name)
	}
	sort.Strings(enabled)

	providers := make([]secret.Provider, 0, len(enabled))
	for _, name := range enabled {
		pcfg := cfg.Providers[name]
		p, err := reg.Create(name, pcfg.Config)
		if err != nil {
			for _, created := range providers {
				_ = created.Close()
			}
			return nil, nil, err
		}
		providers = append(providers, p)
	}

	resolver := secret.NewResolver(cfg.Strict, providers...)
	closeFn := func() error {
		var errs []error
		for _, p := range providers {
			if p == nil {
				continue
			}
			if err := p.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	}
	return resolver, closeFn, nil
}

// ResolveMCPBackendConfigs resolves env and secret refs in MCP backend URLs and headers.
func ResolveMCPBackendConfigs(ctx context.Context, r *secret.Resolver, backends []config.MCPBackendConfig) ([]config.MCPBackendConfig, error) {
	if len(backends) == 0 {
		return nil, nil
	}
	out := make([]config.MCPBackendConfig, len(backends))
	for i := range backends {
		out[i] = backends[i]

		url, err := r.ResolveValue(ctx, out[i].URL)
		if err != nil {
			return nil, fmt.Errorf("mcp backend %q url: %w", out[i].Name, err)
		}
		out[i].URL = url

		headers, err := r.ResolveMap(ctx, out[i].Headers)
		if err != nil {
			return nil, fmt.Errorf("mcp backend %q headers: %w", out[i].Name, err)
		}
		out[i].Headers = headers
	}
	return out, nil
}

