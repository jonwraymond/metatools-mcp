package auth

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/metatools-mcp/internal/middleware"
	"github.com/jonwraymond/metatools-mcp/internal/provider"
)

// AuthMiddlewareConfig configures the authentication middleware.
type AuthMiddlewareConfig struct {
	// AllowAnonymous permits requests without authentication.
	// When true and authentication fails, an anonymous identity is used.
	AllowAnonymous bool

	// AnonymousIdentity is used when AllowAnonymous is true and auth fails.
	// If nil, a default anonymous identity is created.
	AnonymousIdentity *Identity
}

// AuthMiddleware creates authentication middleware that validates credentials
// and adds identity to the context.
func AuthMiddleware(auth Authenticator, cfg AuthMiddlewareConfig) middleware.Middleware {
	return func(next provider.ToolProvider) provider.ToolProvider {
		return &authMiddlewareProvider{
			next:   next,
			auth:   auth,
			config: cfg,
		}
	}
}

type authMiddlewareProvider struct {
	next   provider.ToolProvider
	auth   Authenticator
	config AuthMiddlewareConfig
}

func (p *authMiddlewareProvider) Name() string    { return p.next.Name() }
func (p *authMiddlewareProvider) Enabled() bool   { return p.next.Enabled() }
func (p *authMiddlewareProvider) Tool() mcp.Tool  { return p.next.Tool() }

func (p *authMiddlewareProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	// Build auth request from context headers
	headers := HeadersFromContext(ctx)
	authReq := &AuthRequest{
		Headers:  headers,
		Resource: p.next.Name(),
	}

	// Authenticate
	result, err := p.auth.Authenticate(ctx, authReq)
	if err != nil {
		// Internal error
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "authentication error: " + err.Error()},
			},
		}, nil, nil
	}

	var identity *Identity

	if result.Authenticated {
		identity = result.Identity
	} else if p.config.AllowAnonymous {
		// Use anonymous identity
		if p.config.AnonymousIdentity != nil {
			identity = p.config.AnonymousIdentity
		} else {
			identity = AnonymousIdentity()
		}
	} else {
		// Authentication failed
		errMsg := "unauthorized"
		if result.Error != nil {
			errMsg = result.Error.Error()
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: errMsg},
			},
		}, nil, nil
	}

	// Add identity to context
	ctx = WithIdentity(ctx, identity)

	return p.next.Handle(ctx, req, args)
}

// AuthzMiddlewareConfig configures the authorization middleware.
type AuthzMiddlewareConfig struct {
	// ResourceResolver extracts the resource from the request.
	// If nil, the tool name is used as the resource.
	ResourceResolver func(ctx context.Context, toolName string, input map[string]any) string

	// ActionResolver extracts the action from the request.
	// If nil, "call" is used as the default action.
	ActionResolver func(ctx context.Context, toolName string, input map[string]any) string

	// AllowAnonymous permits requests without an identity in context.
	AllowAnonymous bool
}

// AuthzMiddleware creates authorization middleware that checks permissions.
func AuthzMiddleware(authz Authorizer, cfg AuthzMiddlewareConfig) middleware.Middleware {
	return func(next provider.ToolProvider) provider.ToolProvider {
		return &authzMiddlewareProvider{
			next:   next,
			authz:  authz,
			config: cfg,
		}
	}
}

type authzMiddlewareProvider struct {
	next   provider.ToolProvider
	authz  Authorizer
	config AuthzMiddlewareConfig
}

func (p *authzMiddlewareProvider) Name() string    { return p.next.Name() }
func (p *authzMiddlewareProvider) Enabled() bool   { return p.next.Enabled() }
func (p *authzMiddlewareProvider) Tool() mcp.Tool  { return p.next.Tool() }

func (p *authzMiddlewareProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	identity := IdentityFromContext(ctx)

	// Check if identity is required
	if identity == nil && !p.config.AllowAnonymous {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "unauthorized: no identity in context"},
			},
		}, nil, nil
	}

	// Build authz request
	toolName := p.next.Name()
	resource := "tool:" + toolName
	action := "call"

	if p.config.ResourceResolver != nil {
		resource = p.config.ResourceResolver(ctx, toolName, args)
	}
	if p.config.ActionResolver != nil {
		action = p.config.ActionResolver(ctx, toolName, args)
	}

	authzReq := &AuthzRequest{
		Subject:      identity,
		Resource:     resource,
		Action:       action,
		ResourceType: "tool",
	}

	// Authorize
	if err := p.authz.Authorize(ctx, authzReq); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "forbidden: " + err.Error()},
			},
		}, nil, nil
	}

	return p.next.Handle(ctx, req, args)
}
