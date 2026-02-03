package middleware

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/jonwraymond/toolops/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type stubProvider struct {
	called  bool
	lastCtx context.Context
}

func (s *stubProvider) Name() string   { return "stub" }
func (s *stubProvider) Enabled() bool  { return true }
func (s *stubProvider) Tool() mcp.Tool { return mcp.Tool{Name: "stub"} }
func (s *stubProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	s.called = true
	s.lastCtx = ctx
	return &mcp.CallToolResult{}, "ok", nil
}

func TestAuthMiddleware_AllowsAuthenticated(t *testing.T) {
	authn := auth.NewAuthenticatorFunc(
		"test",
		func(_ context.Context, _ *auth.AuthRequest) bool { return true },
		func(_ context.Context, _ *auth.AuthRequest) (*auth.AuthResult, error) {
			return auth.AuthSuccess(&auth.Identity{Principal: "user", Method: auth.AuthMethodJWT}), nil
		},
	)

	mw := AuthMiddleware(authn, auth.AllowAllAuthorizer{}, AuthConfig{})
	prov := &stubProvider{}
	wrapped := mw(prov)

	_, _, _ = wrapped.Handle(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "stub"}}, nil)

	if !prov.called {
		t.Fatal("expected provider to be called")
	}
	if got := auth.PrincipalFromContext(prov.lastCtx); got != "user" {
		t.Fatalf("principal = %q, want %q", got, "user")
	}
}

func TestAuthMiddleware_DeniesUnauthorized(t *testing.T) {
	authn := auth.NewAuthenticatorFunc(
		"test",
		func(_ context.Context, _ *auth.AuthRequest) bool { return true },
		func(_ context.Context, _ *auth.AuthRequest) (*auth.AuthResult, error) {
			return auth.AuthFailure(auth.ErrInvalidCredentials, "test"), nil
		},
	)

	mw := AuthMiddleware(authn, auth.AllowAllAuthorizer{}, AuthConfig{AllowAnonymous: false})
	prov := &stubProvider{}
	wrapped := mw(prov)

	result, _, _ := wrapped.Handle(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "stub"}}, nil)
	if result == nil || !result.IsError {
		t.Fatal("expected error result")
	}
	if prov.called {
		t.Fatal("provider should not be called when unauthorized")
	}
}

func TestAuthMiddleware_AllowsAnonymous(t *testing.T) {
	authn := auth.NewAuthenticatorFunc(
		"test",
		func(_ context.Context, _ *auth.AuthRequest) bool { return true },
		func(_ context.Context, _ *auth.AuthRequest) (*auth.AuthResult, error) {
			return auth.AuthFailure(auth.ErrInvalidCredentials, "test"), nil
		},
	)

	mw := AuthMiddleware(authn, auth.AllowAllAuthorizer{}, AuthConfig{AllowAnonymous: true})
	prov := &stubProvider{}
	wrapped := mw(prov)

	_, _, _ = wrapped.Handle(context.Background(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "stub"}}, nil)

	if !prov.called {
		t.Fatal("expected provider to be called for anonymous access")
	}
	if id := auth.IdentityFromContext(prov.lastCtx); id == nil || !id.IsAnonymous() {
		t.Fatal("expected anonymous identity in context")
	}
}

var _ provider.ToolProvider = (*stubProvider)(nil)
