package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuditEntry contains information about a tool invocation for audit logging.
type AuditEntry struct {
	// Timestamp is when the request was received.
	Timestamp time.Time

	// RequestID is a correlation ID for the request.
	RequestID string

	// Tool is the name of the invoked tool.
	Tool string

	// Principal is the identity making the request.
	Principal string

	// TenantID is the tenant context.
	TenantID string

	// Roles are the roles of the principal.
	Roles []string

	// Duration is how long the tool execution took.
	Duration time.Duration

	// Success indicates if the tool executed successfully.
	Success bool

	// ErrorMsg contains the error message if Success is false.
	ErrorMsg string

	// Args contains the input arguments (if IncludeArgs is enabled).
	Args map[string]any
}

// AuditLogger receives audit entries.
type AuditLogger interface {
	Log(ctx context.Context, entry AuditEntry)
}

// AuditLoggerFunc is an adapter to use functions as AuditLogger.
type AuditLoggerFunc func(ctx context.Context, entry AuditEntry)

// Log calls the function.
func (f AuditLoggerFunc) Log(ctx context.Context, entry AuditEntry) {
	f(ctx, entry)
}

// AuditIdentity contains identity information for audit logging.
type AuditIdentity struct {
	Principal string
	TenantID  string
	Roles     []string
}

// AuditConfig configures the audit logging middleware.
type AuditConfig struct {
	// Logger is the slog logger for audit entries.
	// Used when AuditLogger is nil.
	Logger *slog.Logger

	// AuditLogger is a custom audit entry sink.
	// If set, this takes precedence over Logger.
	AuditLogger AuditLogger

	// IncludeArgs includes tool arguments in audit entries.
	// WARNING: May log sensitive data.
	IncludeArgs bool

	// IncludeResult includes tool result in audit entries.
	IncludeResult bool

	// IncludeHeaders includes HTTP headers in audit entries.
	// Sensitive headers are automatically redacted.
	IncludeHeaders bool

	// IdentityExtractor extracts identity info from context.
	IdentityExtractor func(ctx context.Context) AuditIdentity

	// RequestIDExtractor extracts request ID from context.
	RequestIDExtractor func(ctx context.Context) string
}

// NewAuditLoggingMiddleware creates an audit logging middleware.
func NewAuditLoggingMiddleware(cfg AuditConfig) Middleware {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return func(next provider.ToolProvider) provider.ToolProvider {
		return &auditProvider{
			next:   next,
			config: cfg,
		}
	}
}

// AuditLoggingMiddlewareFactory creates an audit middleware from config.
func AuditLoggingMiddlewareFactory(cfg map[string]any) (Middleware, error) {
	config := AuditConfig{}

	if v, ok := cfg["include_args"].(bool); ok {
		config.IncludeArgs = v
	}
	if v, ok := cfg["include_result"].(bool); ok {
		config.IncludeResult = v
	}
	if v, ok := cfg["include_headers"].(bool); ok {
		config.IncludeHeaders = v
	}

	return NewAuditLoggingMiddleware(config), nil
}

type auditProvider struct {
	next   provider.ToolProvider
	config AuditConfig
}

func (a *auditProvider) Name() string  { return a.next.Name() }
func (a *auditProvider) Enabled() bool { return a.next.Enabled() }
func (a *auditProvider) Tool() mcp.Tool {
	return a.next.Tool()
}

func (a *auditProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	start := time.Now()

	// Execute the tool
	result, out, err := a.next.Handle(ctx, req, args)

	// Build audit entry
	entry := AuditEntry{
		Timestamp: start,
		Tool:      a.next.Name(),
		Duration:  time.Since(start),
	}

	// Extract identity if configured
	if a.config.IdentityExtractor != nil {
		identity := a.config.IdentityExtractor(ctx)
		entry.Principal = identity.Principal
		entry.TenantID = identity.TenantID
		entry.Roles = identity.Roles
	}

	// Extract request ID if configured
	if a.config.RequestIDExtractor != nil {
		entry.RequestID = a.config.RequestIDExtractor(ctx)
	}

	// Determine success
	if err != nil {
		entry.Success = false
		entry.ErrorMsg = err.Error()
	} else if result != nil && result.IsError {
		entry.Success = false
	} else {
		entry.Success = true
	}

	// Include args if configured
	if a.config.IncludeArgs && args != nil {
		entry.Args = args
	}

	// Log the entry
	a.logEntry(ctx, entry)

	return result, out, err
}

func (a *auditProvider) logEntry(ctx context.Context, entry AuditEntry) {
	// Use custom audit logger if provided
	if a.config.AuditLogger != nil {
		a.config.AuditLogger.Log(ctx, entry)
		return
	}

	// Fall back to slog
	attrs := []any{
		"tool", entry.Tool,
		"success", entry.Success,
		"duration_ms", entry.Duration.Milliseconds(),
	}

	if entry.RequestID != "" {
		attrs = append(attrs, "request_id", entry.RequestID)
	}
	if entry.Principal != "" {
		attrs = append(attrs, "principal", entry.Principal)
	}
	if entry.TenantID != "" {
		attrs = append(attrs, "tenant_id", entry.TenantID)
	}
	if len(entry.Roles) > 0 {
		attrs = append(attrs, "roles", entry.Roles)
	}
	if entry.ErrorMsg != "" {
		attrs = append(attrs, "error", entry.ErrorMsg)
	}
	if entry.Args != nil {
		attrs = append(attrs, "args", entry.Args)
	}

	if entry.Success {
		a.config.Logger.Info("tool_call", attrs...)
	} else {
		a.config.Logger.Warn("tool_call", attrs...)
	}
}
