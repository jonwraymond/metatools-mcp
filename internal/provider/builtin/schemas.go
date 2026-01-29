package builtin

import (
	"github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var errorCodes = []string{
	string(errors.CodeToolNotFound),
	string(errors.CodeNoBackends),
	string(errors.CodeBackendOverrideInvalid),
	string(errors.CodeBackendOverrideNoMatch),
	string(errors.CodeValidationInput),
	string(errors.CodeValidationOutput),
	string(errors.CodeExecutionFailed),
	string(errors.CodeStreamNotSupported),
	string(errors.CodeStreamFailed),
	string(errors.CodeChainStepFailed),
	string(errors.CodeCancelled),
	string(errors.CodeTimeout),
	string(errors.CodeInternal),
}

func searchToolsTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func listNamespacesTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func describeToolTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func listToolExamplesTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func runToolTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func runChainTool() mcp.Tool {
	return mcp.Tool{
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
	}
}

func executeCodeTool() mcp.Tool {
	return mcp.Tool{
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
	}
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
