package middleware

import (
	"context"
	"sync"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"
)

// IdentityExtractor extracts an identity principal from context.
// This breaks the import cycle between middleware and auth packages.
type IdentityExtractor func(ctx context.Context) string

// RateLimitRule defines rate limiting parameters.
type RateLimitRule struct {
	// Rate is the number of tokens per second.
	Rate float64

	// Burst is the maximum number of tokens that can be consumed at once.
	Burst int
}

// RateLimitConfig configures the rate limiting middleware.
type RateLimitConfig struct {
	// Rate is the global rate limit in requests per second.
	Rate float64

	// Burst is the maximum burst size.
	Burst int

	// PerTool allows configuring different limits per tool.
	PerTool map[string]RateLimitRule

	// PerIdentity enables per-user rate limiting.
	PerIdentity bool

	// IdentityRate is the rate limit applied per identity when PerIdentity is true.
	IdentityRate RateLimitRule

	// GlobalFallback is the rate limit for anonymous/unauthenticated requests.
	// If nil, global rate is used for anonymous requests.
	GlobalFallback *RateLimitRule

	// IdentityExtractor extracts identity principal from context.
	// If nil, a default extractor is used that looks for auth.Identity in context.
	IdentityExtractor IdentityExtractor
}

// NewRateLimitMiddleware creates a rate limiting middleware.
func NewRateLimitMiddleware(cfg RateLimitConfig) Middleware {
	// Apply defaults
	if cfg.Rate == 0 {
		cfg.Rate = 100.0
	}
	if cfg.Burst == 0 {
		cfg.Burst = 10
	}
	if cfg.IdentityRate.Rate == 0 {
		cfg.IdentityRate.Rate = cfg.Rate
	}
	if cfg.IdentityRate.Burst == 0 {
		cfg.IdentityRate.Burst = cfg.Burst
	}

	return func(next provider.ToolProvider) provider.ToolProvider {
		return &rateLimitProvider{
			next:              next,
			config:            cfg,
			globalLimiter:     rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst),
			toolLimiters:      make(map[string]*rate.Limiter),
			identityLimiters:  make(map[string]*rate.Limiter),
			identityExtractor: cfg.IdentityExtractor,
		}
	}
}

// RateLimitMiddlewareFactory creates a rate limiting middleware from config.
func RateLimitMiddlewareFactory(cfg map[string]any) (Middleware, error) {
	config := RateLimitConfig{}

	if v, ok := cfg["rate"].(float64); ok {
		config.Rate = v
	}
	if v, ok := cfg["burst"].(int); ok {
		config.Burst = v
	}
	if v, ok := cfg["burst"].(float64); ok {
		config.Burst = int(v)
	}
	if v, ok := cfg["per_identity"].(bool); ok {
		config.PerIdentity = v
	}

	// Parse per-tool limits
	if perTool, ok := cfg["per_tool"].(map[string]any); ok {
		config.PerTool = make(map[string]RateLimitRule)
		for toolName, toolCfg := range perTool {
			if tc, ok := toolCfg.(map[string]any); ok {
				rule := RateLimitRule{}
				if r, ok := tc["rate"].(float64); ok {
					rule.Rate = r
				}
				if b, ok := tc["burst"].(int); ok {
					rule.Burst = b
				}
				if b, ok := tc["burst"].(float64); ok {
					rule.Burst = int(b)
				}
				config.PerTool[toolName] = rule
			}
		}
	}

	// Parse global fallback
	if fallback, ok := cfg["global_fallback"].(map[string]any); ok {
		rule := RateLimitRule{}
		if r, ok := fallback["rate"].(float64); ok {
			rule.Rate = r
		}
		if b, ok := fallback["burst"].(int); ok {
			rule.Burst = b
		}
		if b, ok := fallback["burst"].(float64); ok {
			rule.Burst = int(b)
		}
		config.GlobalFallback = &rule
	}

	// Parse identity rate
	if identityRate, ok := cfg["identity_rate"].(map[string]any); ok {
		if r, ok := identityRate["rate"].(float64); ok {
			config.IdentityRate.Rate = r
		}
		if b, ok := identityRate["burst"].(int); ok {
			config.IdentityRate.Burst = b
		}
		if b, ok := identityRate["burst"].(float64); ok {
			config.IdentityRate.Burst = int(b)
		}
	}

	return NewRateLimitMiddleware(config), nil
}

type rateLimitProvider struct {
	next   provider.ToolProvider
	config RateLimitConfig

	mu                sync.RWMutex
	globalLimiter     *rate.Limiter
	toolLimiters      map[string]*rate.Limiter
	identityLimiters  map[string]*rate.Limiter
	fallbackLimiter   *rate.Limiter
	identityExtractor IdentityExtractor
}

func (r *rateLimitProvider) Name() string  { return r.next.Name() }
func (r *rateLimitProvider) Enabled() bool { return r.next.Enabled() }
func (r *rateLimitProvider) Tool() mcp.Tool {
	return r.next.Tool()
}

func (r *rateLimitProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	// Check rate limits
	if !r.allow(ctx) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "rate limit exceeded"},
			},
		}, nil, nil
	}

	return r.next.Handle(ctx, req, args)
}

func (r *rateLimitProvider) allow(ctx context.Context) bool {
	toolName := r.next.Name()

	// Check per-tool limit if configured
	if rule, ok := r.config.PerTool[toolName]; ok {
		limiter := r.getToolLimiter(toolName, rule)
		if !limiter.Allow() {
			return false
		}
	}

	// Check per-identity or fallback limit
	if r.config.PerIdentity {
		principal := r.extractPrincipal(ctx)
		if principal != "" {
			limiter := r.getIdentityLimiter(principal)
			return limiter.Allow()
		}
		// Anonymous request - use fallback if configured
		if r.config.GlobalFallback != nil {
			limiter := r.getFallbackLimiter()
			return limiter.Allow()
		}
	}

	// Use global limiter
	return r.globalLimiter.Allow()
}

func (r *rateLimitProvider) getToolLimiter(toolName string, rule RateLimitRule) *rate.Limiter {
	r.mu.RLock()
	limiter, ok := r.toolLimiters[toolName]
	r.mu.RUnlock()

	if ok {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok = r.toolLimiters[toolName]; ok {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(rule.Rate), rule.Burst)
	r.toolLimiters[toolName] = limiter
	return limiter
}

func (r *rateLimitProvider) getIdentityLimiter(principal string) *rate.Limiter {
	r.mu.RLock()
	limiter, ok := r.identityLimiters[principal]
	r.mu.RUnlock()

	if ok {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok = r.identityLimiters[principal]; ok {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(r.config.IdentityRate.Rate), r.config.IdentityRate.Burst)
	r.identityLimiters[principal] = limiter
	return limiter
}

func (r *rateLimitProvider) getFallbackLimiter() *rate.Limiter {
	r.mu.RLock()
	limiter := r.fallbackLimiter
	r.mu.RUnlock()

	if limiter != nil {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.fallbackLimiter != nil {
		return r.fallbackLimiter
	}

	r.fallbackLimiter = rate.NewLimiter(rate.Limit(r.config.GlobalFallback.Rate), r.config.GlobalFallback.Burst)
	return r.fallbackLimiter
}

func (r *rateLimitProvider) extractPrincipal(ctx context.Context) string {
	if r.identityExtractor != nil {
		return r.identityExtractor(ctx)
	}
	return ""
}
