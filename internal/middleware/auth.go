package middleware

import (
	"context"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/jonwraymond/toolops/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuthConfig configures authentication and authorization middleware.
type AuthConfig struct {
	// AllowAnonymous permits requests without authentication.
	AllowAnonymous bool

	// AnonymousIdentity overrides the default anonymous identity.
	AnonymousIdentity *auth.Identity

	// Authenticators is the ordered list of authenticators to try.
	Authenticators []AuthEntry

	// Authorizer is the authorizer to use. If empty, allow-all is used.
	Authorizer *AuthEntry

	// ResourcePrefix prefixes tool names when constructing authz resources.
	// Default: "tool:"
	ResourcePrefix string

	// Action is the action used for authorization.
	// Default: "call"
	Action string
}

// AuthEntry defines an authenticator/authorizer configuration.
type AuthEntry struct {
	Name   string
	Config map[string]any
}

// AuthMiddlewareFactory builds auth middleware from config.
func AuthMiddlewareFactory(cfg map[string]any) (Middleware, error) {
	parsed := parseAuthConfig(cfg)

	authn, err := buildAuthenticator(parsed)
	if err != nil {
		return nil, err
	}

	authz, err := buildAuthorizer(parsed)
	if err != nil {
		return nil, err
	}

	return AuthMiddleware(authn, authz, parsed), nil
}

// AuthMiddleware creates middleware that authenticates and authorizes requests.
func AuthMiddleware(authn auth.Authenticator, authz auth.Authorizer, cfg AuthConfig) Middleware {
	return func(next provider.ToolProvider) provider.ToolProvider {
		return &authMiddlewareProvider{
			next:   next,
			authn:  authn,
			authz:  authz,
			config: cfg,
		}
	}
}

type authMiddlewareProvider struct {
	next   provider.ToolProvider
	authn  auth.Authenticator
	authz  auth.Authorizer
	config AuthConfig
}

func (p *authMiddlewareProvider) Name() string   { return p.next.Name() }
func (p *authMiddlewareProvider) Enabled() bool  { return p.next.Enabled() }
func (p *authMiddlewareProvider) Tool() mcp.Tool { return p.next.Tool() }

func (p *authMiddlewareProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	authReq := &auth.AuthRequest{
		Headers:  auth.HeadersFromContext(ctx),
		Resource: p.next.Name(),
		Metadata: map[string]any{
			"tool": p.next.Name(),
		},
	}

	result, err := p.authn.Authenticate(ctx, authReq)
	if err != nil {
		return errorResult("authentication error: "+err.Error()), nil, nil
	}

	var identity *auth.Identity
	if result != nil && result.Authenticated {
		identity = result.Identity
	} else if p.config.AllowAnonymous {
		if p.config.AnonymousIdentity != nil {
			identity = p.config.AnonymousIdentity
		} else {
			identity = auth.AnonymousIdentity()
		}
	} else {
		msg := "unauthorized"
		if result != nil && result.Error != nil {
			msg = result.Error.Error()
		}
		return errorResult(msg), nil, nil
	}

	ctx = auth.WithIdentity(ctx, identity)

	if p.authz != nil {
		resourcePrefix := p.config.ResourcePrefix
		if resourcePrefix == "" {
			resourcePrefix = "tool:"
		}
		action := p.config.Action
		if action == "" {
			action = "call"
		}
		authzReq := &auth.AuthzRequest{
			Subject:      identity,
			Resource:     resourcePrefix + p.next.Name(),
			Action:       action,
			ResourceType: "tool",
		}
		if err := p.authz.Authorize(ctx, authzReq); err != nil {
			return errorResult("forbidden: " + err.Error()), nil, nil
		}
	}

	return p.next.Handle(ctx, req, args)
}

func parseAuthConfig(cfg map[string]any) AuthConfig {
	parsed := AuthConfig{}
	if cfg == nil {
		return parsed
	}
	if v, ok := cfg["allow_anonymous"].(bool); ok {
		parsed.AllowAnonymous = v
	}
	if v, ok := cfg["resource_prefix"].(string); ok {
		parsed.ResourcePrefix = v
	}
	if v, ok := cfg["action"].(string); ok {
		parsed.Action = v
	}
	if list, ok := cfg["authenticators"].([]any); ok {
		for _, raw := range list {
			if entry, ok := parseAuthEntry(raw); ok {
				parsed.Authenticators = append(parsed.Authenticators, entry)
			}
		}
	}
	if raw, ok := cfg["authorizer"]; ok {
		if entry, ok := parseAuthEntry(raw); ok {
			parsed.Authorizer = &entry
		}
	}
	return parsed
}

func parseAuthEntry(raw any) (AuthEntry, bool) {
	entry := AuthEntry{}
	switch v := raw.(type) {
	case map[string]any:
		if name, ok := v["name"].(string); ok {
			entry.Name = name
		}
		if cfg, ok := v["config"].(map[string]any); ok {
			entry.Config = cfg
		} else {
			entry.Config = map[string]any{}
		}
	case string:
		entry.Name = v
		entry.Config = map[string]any{}
	default:
		return AuthEntry{}, false
	}
	if entry.Name == "" {
		return AuthEntry{}, false
	}
	if entry.Config == nil {
		entry.Config = map[string]any{}
	}
	return entry, true
}

func buildAuthenticator(cfg AuthConfig) (auth.Authenticator, error) {
	if len(cfg.Authenticators) == 0 {
		return auth.NewCompositeAuthenticator(), nil
	}
	auths := make([]auth.Authenticator, 0, len(cfg.Authenticators))
	for _, entry := range cfg.Authenticators {
		authn, err := auth.DefaultRegistry.CreateAuthenticator(entry.Name, entry.Config)
		if err != nil {
			return nil, fmt.Errorf("authenticator %q: %w", entry.Name, err)
		}
		auths = append(auths, authn)
	}
	return auth.NewCompositeAuthenticator(auths...), nil
}

func buildAuthorizer(cfg AuthConfig) (auth.Authorizer, error) {
	if cfg.Authorizer == nil {
		return auth.AllowAllAuthorizer{}, nil
	}
	authz, err := auth.DefaultRegistry.CreateAuthorizer(cfg.Authorizer.Name, cfg.Authorizer.Config)
	if err != nil {
		return nil, fmt.Errorf("authorizer %q: %w", cfg.Authorizer.Name, err)
	}
	return authz, nil
}

func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}
}
