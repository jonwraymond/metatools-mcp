package middleware

import (
	"context"
	"log/slog"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LoggingConfig configures the logging middleware.
type LoggingConfig struct {
	Logger        *slog.Logger
	IncludeArgs   bool
	IncludeResult bool
}

// NewLoggingMiddleware creates a middleware that logs requests and responses.
func NewLoggingMiddleware(cfg LoggingConfig) Middleware {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(next provider.ToolProvider) provider.ToolProvider {
		return &loggingProvider{
			next:   next,
			logger: logger,
			cfg:    cfg,
		}
	}
}

// LoggingMiddlewareFactory creates a logging middleware from config.
func LoggingMiddlewareFactory(cfg map[string]any) (Middleware, error) {
	options := LoggingConfig{}
	if v, ok := cfg["include_args"].(bool); ok {
		options.IncludeArgs = v
	}
	if v, ok := cfg["include_result"].(bool); ok {
		options.IncludeResult = v
	}
	return NewLoggingMiddleware(options), nil
}

type loggingProvider struct {
	next   provider.ToolProvider
	logger *slog.Logger
	cfg    LoggingConfig
}

func (l *loggingProvider) Name() string  { return l.next.Name() }
func (l *loggingProvider) Enabled() bool { return l.next.Enabled() }
func (l *loggingProvider) Tool() mcp.Tool {
	return l.next.Tool()
}

func (l *loggingProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	attrs := []any{"tool", l.next.Name()}
	if l.cfg.IncludeArgs {
		attrs = append(attrs, "args", args)
	}
	l.logger.Info("tool call", attrs...)

	res, out, err := l.next.Handle(ctx, req, args)
	if err != nil {
		l.logger.Error("tool error", "tool", l.next.Name(), "error", err)
		return res, out, err
	}

	if l.cfg.IncludeResult {
		l.logger.Info("tool result", "tool", l.next.Name(), "result", out)
	} else {
		l.logger.Info("tool success", "tool", l.next.Name())
	}
	return res, out, nil
}
