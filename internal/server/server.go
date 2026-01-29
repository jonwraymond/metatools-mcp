package server

import (
	"context"
	"sync"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolindex"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	implementationName    = "metatools-mcp"
	implementationVersion = "0.1.0"
	defaultPageSize       = 50
)

var errorCodes = []string{
	"tool_not_found",
	"no_backends",
	"backend_override_invalid",
	"backend_override_no_match",
	"validation_input",
	"validation_output",
	"execution_failed",
	"stream_not_supported",
	"stream_failed",
	"chain_step_failed",
	"cancelled",
	"timeout",
	"internal",
}

// Capabilities represents the capabilities this server supports.
// This mirrors the MCP tools capability for simple checks in tests and callers.
type Capabilities struct {
	Tools bool
}

// Server is the metatools MCP server backed by the official MCP Go SDK.
type Server struct {
	config   config.Config
	mcp      *mcp.Server
	tools    []*mcp.Tool
	handlers *Handlers

	toolRegistrations []func()
	toolListMu        sync.Mutex
	toolListTimer     *time.Timer
	toolListUnsub     func()
}

// Handlers holds all the metatool handlers.
type Handlers struct {
	Search     *handlers.SearchHandler
	Namespaces *handlers.NamespacesHandler
	Describe   *handlers.DescribeHandler
	Examples   *handlers.ExamplesHandler
	Run        *handlers.RunHandler
	Chain      *handlers.ChainHandler
	Code       *handlers.CodeHandler
}

// New creates a new metatools server.
func New(cfg config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	h := &Handlers{
		Search:     handlers.NewSearchHandler(cfg.Index),
		Namespaces: handlers.NewNamespacesHandler(cfg.Index),
		Describe:   handlers.NewDescribeHandler(cfg.Docs),
		Examples:   handlers.NewExamplesHandler(cfg.Docs),
		Run:        handlers.NewRunHandler(cfg.Runner),
		Chain:      handlers.NewChainHandler(cfg.Runner),
	}
	if cfg.Executor != nil {
		h.Code = handlers.NewCodeHandler(cfg.Executor)
	}

	serverOptions := &mcp.ServerOptions{
		PageSize: defaultPageSize,
	}
	if !cfg.NotifyToolListChanged {
		serverOptions.Capabilities = &mcp.ServerCapabilities{
			Logging: &mcp.LoggingCapabilities{},
			Tools:   &mcp.ToolCapabilities{},
		}
	}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    implementationName,
		Version: implementationVersion,
	}, serverOptions)

	srv := &Server{
		config:   cfg,
		mcp:      mcpServer,
		handlers: h,
	}
	srv.registerTools()
	srv.registerToolListNotifications()
	return srv, nil
}

// MCPServer returns the underlying MCP SDK server.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}

// Run starts handling MCP requests over the provided transport.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.mcp.Run(ctx, transport)
}

// ListTools returns the registered MCP tools.
func (s *Server) ListTools() []*mcp.Tool {
	out := make([]*mcp.Tool, len(s.tools))
	copy(out, s.tools)
	return out
}

// Capabilities returns the server capabilities.
func (s *Server) Capabilities() Capabilities {
	return Capabilities{Tools: len(s.tools) > 0}
}

// Handlers returns the server's handlers for tool execution.
func (s *Server) Handlers() *Handlers {
	return s.handlers
}

func (s *Server) registerTools() {
	registerTool(s, &mcp.Tool{
		Name:        "search_tools",
		Description: "Search for tools by query",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":  map[string]any{"type": "string"},
				"limit":  map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
				"cursor": map[string]any{"type": "string"},
			},
			"required":             []string{"query"},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tools": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":               map[string]any{"type": "string"},
							"name":             map[string]any{"type": "string"},
							"namespace":        map[string]any{"type": "string"},
							"shortDescription": map[string]any{"type": "string"},
							"tags": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
						},
						"required":             []string{"id", "name"},
						"additionalProperties": false,
					},
				},
				"nextCursor": map[string]any{"type": "string"},
			},
			"required":             []string{"tools"},
			"additionalProperties": false,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input metatools.SearchToolsInput) (*mcp.CallToolResult, metatools.SearchToolsOutput, error) {
		out, err := s.handlers.Search.Handle(ctx, input)
		if err != nil {
			return nil, metatools.SearchToolsOutput{}, err
		}
		if out == nil {
			out = &metatools.SearchToolsOutput{}
		}
		return nil, *out, nil
	})

	registerTool(s, &mcp.Tool{
		Name:        "list_namespaces",
		Description: "List all tool namespaces",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit":  map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
				"cursor": map[string]any{"type": "string"},
			},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"namespaces": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
				"nextCursor": map[string]any{"type": "string"},
			},
			"required":             []string{"namespaces"},
			"additionalProperties": false,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input metatools.ListNamespacesInput) (*mcp.CallToolResult, metatools.ListNamespacesOutput, error) {
		out, err := s.handlers.Namespaces.Handle(ctx, input)
		if err != nil {
			return nil, metatools.ListNamespacesOutput{}, err
		}
		if out == nil {
			out = &metatools.ListNamespacesOutput{}
		}
		return nil, *out, nil
	})

	registerTool(s, &mcp.Tool{
		Name:        "describe_tool",
		Description: "Get detailed documentation for a tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tool_id":      map[string]any{"type": "string"},
				"detail_level": map[string]any{"type": "string", "enum": []string{"summary", "schema", "full"}},
				"examples_max": map[string]any{"type": "integer", "minimum": 0, "maximum": 5},
			},
			"required":             []string{"tool_id", "detail_level"},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tool":       map[string]any{"type": "object"},
				"summary":    map[string]any{"type": "string"},
				"schemaInfo": map[string]any{"type": "object"},
				"notes":      map[string]any{"type": "string"},
				"examples":   map[string]any{"type": "array"},
				"externalRefs": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
			},
			"required":             []string{"summary"},
			"additionalProperties": false,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input metatools.DescribeToolInput) (*mcp.CallToolResult, metatools.DescribeToolOutput, error) {
		out, err := s.handlers.Describe.Handle(ctx, input)
		if err != nil {
			return nil, metatools.DescribeToolOutput{}, err
		}
		if out == nil {
			out = &metatools.DescribeToolOutput{}
		}
		return nil, *out, nil
	})

	registerTool(s, &mcp.Tool{
		Name:        "list_tool_examples",
		Description: "Get usage examples for a tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tool_id": map[string]any{"type": "string"},
				"max":     map[string]any{"type": "integer", "minimum": 1, "maximum": 5},
			},
			"required":             []string{"tool_id"},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"examples": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":          map[string]any{"type": "string"},
							"title":       map[string]any{"type": "string"},
							"description": map[string]any{"type": "string"},
							"args":        map[string]any{"type": "object"},
							"resultHint":  map[string]any{"type": "string"},
						},
						"required":             []string{"title", "description", "args"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"examples"},
			"additionalProperties": false,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input metatools.ListToolExamplesInput) (*mcp.CallToolResult, metatools.ListToolExamplesOutput, error) {
		out, err := s.handlers.Examples.Handle(ctx, input)
		if err != nil {
			return nil, metatools.ListToolExamplesOutput{}, err
		}
		if out == nil {
			out = &metatools.ListToolExamplesOutput{}
		}
		return nil, *out, nil
	})

	registerTool(s, &mcp.Tool{
		Name:        "run_tool",
		Description: "Execute a tool by ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tool_id":            map[string]any{"type": "string"},
				"args":               map[string]any{"type": "object"},
				"stream":             map[string]any{"type": "boolean", "default": false},
				"include_tool":       map[string]any{"type": "boolean", "default": false},
				"include_backend":    map[string]any{"type": "boolean", "default": false},
				"include_mcp_result": map[string]any{"type": "boolean", "default": false},
				"backend_override": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"kind":       map[string]any{"type": "string", "enum": []string{"local", "provider", "mcp"}},
						"serverName": map[string]any{"type": "string"},
						"providerId": map[string]any{"type": "string"},
						"toolId":     map[string]any{"type": "string"},
						"name":       map[string]any{"type": "string"},
					},
					"required":             []string{"kind"},
					"additionalProperties": false,
				},
			},
			"required":             []string{"tool_id"},
			"additionalProperties": false,
		},
		OutputSchema: runToolOutputSchema(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input metatools.RunToolInput) (*mcp.CallToolResult, metatools.RunToolOutput, error) {
		progress := progressNotifier(ctx, req)
		out, isError, err := s.handlers.Run.HandleWithProgress(ctx, input, progress)
		if err != nil {
			return nil, metatools.RunToolOutput{}, err
		}
		if out == nil {
			out = &metatools.RunToolOutput{}
		}
		return &mcp.CallToolResult{IsError: isError}, *out, nil
	})

	registerTool(s, &mcp.Tool{
		Name:        "run_chain",
		Description: "Execute multiple tools in sequence",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"steps": map[string]any{
					"type":     "array",
					"minItems": 1,
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"tool_id":      map[string]any{"type": "string"},
							"args":         map[string]any{"type": "object"},
							"use_previous": map[string]any{"type": "boolean"},
						},
						"required":             []string{"tool_id"},
						"additionalProperties": false,
					},
				},
				"include_backends": map[string]any{"type": "boolean", "default": true},
				"include_tools":    map[string]any{"type": "boolean", "default": false},
			},
			"required":             []string{"steps"},
			"additionalProperties": false,
		},
		OutputSchema: runChainOutputSchema(),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input metatools.RunChainInput) (*mcp.CallToolResult, metatools.RunChainOutput, error) {
		progress := progressNotifier(ctx, req)
		out, isError, err := s.handlers.Chain.HandleWithProgress(ctx, input, progress)
		if err != nil {
			return nil, metatools.RunChainOutput{}, err
		}
		if out == nil {
			out = &metatools.RunChainOutput{}
		}
		return &mcp.CallToolResult{IsError: isError}, *out, nil
	})

	if s.handlers.Code != nil {
		registerTool(s, &mcp.Tool{
			Name:        "execute_code",
			Description: "Execute code-based orchestration",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"language":       map[string]any{"type": "string"},
					"code":           map[string]any{"type": "string"},
					"timeout_ms":     map[string]any{"type": "integer", "minimum": 1, "maximum": 60000},
					"max_tool_calls": map[string]any{"type": "integer", "minimum": 1, "maximum": 1000},
				},
				"required":             []string{"language", "code"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"value":      map[string]any{},
					"stdout":     map[string]any{"type": "string"},
					"stderr":     map[string]any{"type": "string"},
					"durationMs": map[string]any{"type": "integer"},
				},
				"required":             []string{"value"},
				"additionalProperties": false,
			},
		}, func(ctx context.Context, req *mcp.CallToolRequest, input metatools.ExecuteCodeInput) (*mcp.CallToolResult, metatools.ExecuteCodeOutput, error) {
			progress := progressNotifier(ctx, req)
			if progress != nil {
				progress(handlers.ProgressEvent{Progress: 0, Total: 1, Message: "started"})
			}
			out, err := s.handlers.Code.Handle(ctx, input)
			if progress != nil {
				msg := "completed"
				if err != nil {
					msg = "error"
				}
				progress(handlers.ProgressEvent{Progress: 1, Total: 1, Message: msg})
			}
			if err != nil {
				return nil, metatools.ExecuteCodeOutput{}, err
			}
			if out == nil {
				out = &metatools.ExecuteCodeOutput{}
			}
			return nil, *out, nil
		})
	}
}

func registerTool[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	mcp.AddTool(s.mcp, tool, handler)
	s.tools = append(s.tools, tool)
	s.toolRegistrations = append(s.toolRegistrations, func() {
		mcp.AddTool(s.mcp, tool, handler)
	})
}

func progressNotifier(ctx context.Context, req *mcp.CallToolRequest) func(handlers.ProgressEvent) {
	if req == nil || req.Session == nil || req.Params == nil {
		return nil
	}
	token := req.Params.GetProgressToken()
	if token == nil {
		return nil
	}

	return func(ev handlers.ProgressEvent) {
		params := &mcp.ProgressNotificationParams{
			ProgressToken: token,
			Progress:      ev.Progress,
			Total:         ev.Total,
			Message:       ev.Message,
		}
		_ = req.Session.NotifyProgress(ctx, params)
	}
}

func (s *Server) registerToolListNotifications() {
	if !s.config.NotifyToolListChanged {
		return
	}
	changeNotifier, ok := s.config.Index.(toolindex.ChangeNotifier)
	if !ok {
		return
	}
	debounce := time.Duration(s.config.NotifyToolListChangedDebounceMs) * time.Millisecond
	if debounce <= 0 {
		debounce = 150 * time.Millisecond
	}

	s.toolListUnsub = changeNotifier.OnChange(func(_ toolindex.ChangeEvent) {
		s.toolListMu.Lock()
		defer s.toolListMu.Unlock()
		if s.toolListTimer == nil {
			s.toolListTimer = time.AfterFunc(debounce, s.reregisterTools)
			return
		}
		s.toolListTimer.Reset(debounce)
	})
}

func (s *Server) reregisterTools() {
	if len(s.toolRegistrations) == 0 {
		return
	}
	// Re-register a single tool to trigger one list_changed notification.
	s.toolRegistrations[0]()
}

// Close releases resources associated with the server.
func (s *Server) Close() error {
	s.toolListMu.Lock()
	if s.toolListTimer != nil {
		s.toolListTimer.Stop()
		s.toolListTimer = nil
	}
	s.toolListMu.Unlock()

	if s.toolListUnsub != nil {
		s.toolListUnsub()
		s.toolListUnsub = nil
	}
	return nil
}

func errorSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"code": map[string]any{
				"type": "string",
				"enum": errorCodes,
			},
			"message":      map[string]any{"type": "string"},
			"tool_id":      map[string]any{"type": "string"},
			"op":           map[string]any{"type": "string"},
			"backend_kind": map[string]any{"type": "string", "enum": []string{"mcp", "provider", "local"}},
			"step_index":   map[string]any{"type": "integer", "minimum": 0},
			"retryable":    map[string]any{"type": "boolean"},
			"details":      map[string]any{"type": "object"},
		},
		"required":             []string{"code", "message"},
		"additionalProperties": false,
	}
}

func runToolOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"structured": map[string]any{},
			"error":      errorSchema(),
			"tool":       map[string]any{"type": "object"},
			"backend":    map[string]any{"type": "object"},
			"mcpResult":  map[string]any{"type": "object"},
			"durationMs": map[string]any{"type": "integer"},
		},
		"additionalProperties": false,
		"anyOf": []map[string]any{
			{"required": []string{"structured"}},
			{"required": []string{"error"}},
		},
	}
}

func runChainOutputSchema() map[string]any {
	stepSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tool_id":    map[string]any{"type": "string"},
			"structured": map[string]any{},
			"backend":    map[string]any{"type": "object"},
			"tool":       map[string]any{"type": "object"},
			"error":      errorSchema(),
		},
		"required":             []string{"tool_id"},
		"additionalProperties": false,
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"results": map[string]any{
				"type":  "array",
				"items": stepSchema,
			},
			"final": map[string]any{},
			"error": errorSchema(),
		},
		"required":             []string{"results"},
		"additionalProperties": false,
	}
}
