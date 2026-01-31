# PRD-102: Schema Definitions

**Phase:** 0 - Planning & Documentation
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-100

---

## Objective

Define JSON Schema specifications for all core data types used across the consolidated ecosystem, ensuring type safety and validation consistency.

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Tool Schema | `schemas/tool.schema.json` | Canonical tool definition |
| Toolset Schema | `schemas/toolset.schema.json` | Tool collection schema |
| Execution Schema | `schemas/execution.schema.json` | Tool execution request/result |
| Discovery Schema | `schemas/discovery.schema.json` | Tool discovery structures |
| Config Schema | `schemas/config.schema.json` | Configuration file schema |

---

## Tasks

### Task 1: Create Tool Schema

**File:** `schemas/tool.schema.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aperturestack.dev/schemas/tool.schema.json",
  "title": "Tool",
  "description": "Canonical tool definition for ApertureStack",
  "type": "object",
  "required": ["id", "name", "description"],
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9_-]*$",
      "description": "Unique tool identifier"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100,
      "description": "Human-readable tool name"
    },
    "description": {
      "type": "string",
      "minLength": 1,
      "maxLength": 1000,
      "description": "Tool description for discovery"
    },
    "version": {
      "type": "string",
      "pattern": "^v?\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9.]+)?$",
      "description": "Semantic version"
    },
    "namespace": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9_-]*$",
      "description": "Tool namespace for grouping"
    },
    "inputSchema": {
      "$ref": "https://json-schema.org/draft/2020-12/schema",
      "description": "JSON Schema for tool input"
    },
    "outputSchema": {
      "$ref": "https://json-schema.org/draft/2020-12/schema",
      "description": "JSON Schema for tool output"
    },
    "metadata": {
      "type": "object",
      "additionalProperties": true,
      "description": "Arbitrary metadata"
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Searchable tags"
    },
    "capabilities": {
      "$ref": "#/$defs/Capabilities"
    }
  },
  "$defs": {
    "Capabilities": {
      "type": "object",
      "properties": {
        "streaming": {"type": "boolean", "default": false},
        "batch": {"type": "boolean", "default": false},
        "async": {"type": "boolean", "default": false},
        "cacheable": {"type": "boolean", "default": true},
        "idempotent": {"type": "boolean", "default": false}
      }
    }
  }
}
```

### Task 2: Create Toolset Schema

**File:** `schemas/toolset.schema.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aperturestack.dev/schemas/toolset.schema.json",
  "title": "Toolset",
  "description": "Collection of tools with shared configuration",
  "type": "object",
  "required": ["id", "name", "tools"],
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9_-]*$"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "description": {
      "type": "string",
      "maxLength": 1000
    },
    "version": {
      "type": "string",
      "pattern": "^v?\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9.]+)?$"
    },
    "tools": {
      "type": "array",
      "items": {
        "oneOf": [
          {"type": "string", "description": "Tool ID reference"},
          {"$ref": "tool.schema.json"}
        ]
      },
      "minItems": 1
    },
    "config": {
      "type": "object",
      "properties": {
        "timeout": {
          "type": "string",
          "pattern": "^\\d+[smh]$",
          "description": "Execution timeout (e.g., '30s', '5m')"
        },
        "retries": {
          "type": "integer",
          "minimum": 0,
          "maximum": 10
        },
        "cachePolicy": {
          "type": "string",
          "enum": ["none", "default", "aggressive"]
        }
      }
    },
    "permissions": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["read", "write", "execute", "admin"]
      }
    }
  }
}
```

### Task 3: Create Execution Schema

**File:** `schemas/execution.schema.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aperturestack.dev/schemas/execution.schema.json",
  "title": "Execution",
  "description": "Tool execution request and result schemas",
  "$defs": {
    "ExecutionRequest": {
      "type": "object",
      "required": ["toolId", "input"],
      "properties": {
        "toolId": {
          "type": "string",
          "description": "Tool to execute"
        },
        "input": {
          "type": "object",
          "description": "Tool input arguments"
        },
        "context": {
          "$ref": "#/$defs/ExecutionContext"
        },
        "options": {
          "$ref": "#/$defs/ExecutionOptions"
        }
      }
    },
    "ExecutionContext": {
      "type": "object",
      "properties": {
        "requestId": {"type": "string", "format": "uuid"},
        "traceId": {"type": "string"},
        "spanId": {"type": "string"},
        "userId": {"type": "string"},
        "tenantId": {"type": "string"},
        "metadata": {"type": "object"}
      }
    },
    "ExecutionOptions": {
      "type": "object",
      "properties": {
        "timeout": {"type": "string", "pattern": "^\\d+[smh]$"},
        "backend": {"type": "string", "enum": ["local", "docker", "wasm", "remote"]},
        "sandbox": {"type": "string", "enum": ["none", "basic", "strict"]},
        "stream": {"type": "boolean", "default": false},
        "cache": {"type": "boolean", "default": true}
      }
    },
    "ExecutionResult": {
      "type": "object",
      "required": ["status"],
      "properties": {
        "status": {
          "type": "string",
          "enum": ["success", "error", "timeout", "cancelled"]
        },
        "output": {
          "description": "Tool output on success"
        },
        "error": {
          "$ref": "#/$defs/ExecutionError"
        },
        "metrics": {
          "$ref": "#/$defs/ExecutionMetrics"
        }
      }
    },
    "ExecutionError": {
      "type": "object",
      "required": ["code", "message"],
      "properties": {
        "code": {"type": "string"},
        "message": {"type": "string"},
        "details": {"type": "object"},
        "retryable": {"type": "boolean", "default": false}
      }
    },
    "ExecutionMetrics": {
      "type": "object",
      "properties": {
        "startTime": {"type": "string", "format": "date-time"},
        "endTime": {"type": "string", "format": "date-time"},
        "durationMs": {"type": "integer", "minimum": 0},
        "backend": {"type": "string"},
        "cached": {"type": "boolean"}
      }
    }
  }
}
```

### Task 4: Create Discovery Schema

**File:** `schemas/discovery.schema.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aperturestack.dev/schemas/discovery.schema.json",
  "title": "Discovery",
  "description": "Tool discovery and search schemas",
  "$defs": {
    "SearchQuery": {
      "type": "object",
      "properties": {
        "query": {
          "type": "string",
          "description": "Search query text"
        },
        "filters": {
          "$ref": "#/$defs/SearchFilters"
        },
        "options": {
          "$ref": "#/$defs/SearchOptions"
        }
      }
    },
    "SearchFilters": {
      "type": "object",
      "properties": {
        "namespace": {"type": "string"},
        "tags": {"type": "array", "items": {"type": "string"}},
        "capabilities": {"type": "array", "items": {"type": "string"}},
        "minVersion": {"type": "string"},
        "maxVersion": {"type": "string"}
      }
    },
    "SearchOptions": {
      "type": "object",
      "properties": {
        "limit": {"type": "integer", "minimum": 1, "maximum": 100, "default": 10},
        "offset": {"type": "integer", "minimum": 0, "default": 0},
        "sortBy": {"type": "string", "enum": ["relevance", "name", "updated"]},
        "sortOrder": {"type": "string", "enum": ["asc", "desc"]},
        "includeDeprecated": {"type": "boolean", "default": false},
        "searchMode": {"type": "string", "enum": ["bm25", "semantic", "hybrid"]}
      }
    },
    "SearchResult": {
      "type": "object",
      "required": ["tools", "total"],
      "properties": {
        "tools": {
          "type": "array",
          "items": {"$ref": "#/$defs/SearchHit"}
        },
        "total": {"type": "integer", "minimum": 0},
        "hasMore": {"type": "boolean"},
        "searchTime": {"type": "integer", "description": "Search time in milliseconds"}
      }
    },
    "SearchHit": {
      "type": "object",
      "required": ["tool", "score"],
      "properties": {
        "tool": {"$ref": "tool.schema.json"},
        "score": {"type": "number", "minimum": 0, "maximum": 1},
        "highlights": {
          "type": "object",
          "additionalProperties": {"type": "array", "items": {"type": "string"}}
        }
      }
    }
  }
}
```

### Task 5: Create Config Schema

**File:** `schemas/config.schema.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aperturestack.dev/schemas/config.schema.json",
  "title": "Config",
  "description": "metatools-mcp configuration schema",
  "type": "object",
  "properties": {
    "server": {
      "type": "object",
      "properties": {
        "transport": {
          "type": "string",
          "enum": ["stdio", "sse", "websocket", "grpc"],
          "default": "stdio"
        },
        "host": {"type": "string", "default": "localhost"},
        "port": {"type": "integer", "minimum": 1, "maximum": 65535, "default": 8080},
        "tls": {
          "type": "object",
          "properties": {
            "enabled": {"type": "boolean", "default": false},
            "certFile": {"type": "string"},
            "keyFile": {"type": "string"}
          }
        }
      }
    },
    "tools": {
      "type": "object",
      "properties": {
        "directories": {
          "type": "array",
          "items": {"type": "string"},
          "default": ["./tools"]
        },
        "providers": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["type", "url"],
            "properties": {
              "type": {"type": "string", "enum": ["http", "grpc", "mcp"]},
              "url": {"type": "string", "format": "uri"},
              "timeout": {"type": "string", "pattern": "^\\d+[smh]$"}
            }
          }
        }
      }
    },
    "execution": {
      "type": "object",
      "properties": {
        "defaultBackend": {
          "type": "string",
          "enum": ["local", "docker", "wasm", "kubernetes"],
          "default": "local"
        },
        "timeout": {"type": "string", "default": "30s"},
        "maxConcurrent": {"type": "integer", "minimum": 1, "default": 10}
      }
    },
    "cache": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean", "default": true},
        "backend": {"type": "string", "enum": ["memory", "redis", "file"]},
        "ttl": {"type": "string", "default": "5m"},
        "maxSize": {"type": "string", "default": "100MB"}
      }
    },
    "observability": {
      "type": "object",
      "properties": {
        "tracing": {
          "type": "object",
          "properties": {
            "enabled": {"type": "boolean", "default": false},
            "exporter": {"type": "string", "enum": ["otlp", "jaeger", "zipkin"]},
            "endpoint": {"type": "string", "format": "uri"}
          }
        },
        "metrics": {
          "type": "object",
          "properties": {
            "enabled": {"type": "boolean", "default": true},
            "port": {"type": "integer", "default": 9090}
          }
        },
        "logging": {
          "type": "object",
          "properties": {
            "level": {"type": "string", "enum": ["debug", "info", "warn", "error"]},
            "format": {"type": "string", "enum": ["json", "text"]}
          }
        }
      }
    },
    "auth": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean", "default": false},
        "provider": {"type": "string", "enum": ["jwt", "apikey", "oauth2"]},
        "config": {"type": "object"}
      }
    }
  }
}
```

### Task 6: Create Schema Index

**File:** `schemas/README.md`

```markdown
# ApertureStack JSON Schemas

This directory contains JSON Schema definitions for all core data types in the ApertureStack ecosystem.

## Schemas

| Schema | Description | Go Package |
|--------|-------------|------------|
| [tool.schema.json](./tool.schema.json) | Canonical tool definition | `toolfoundation/model` |
| [toolset.schema.json](./toolset.schema.json) | Tool collection | `toolcompose/set` |
| [execution.schema.json](./execution.schema.json) | Execution request/result | `toolexec/run` |
| [discovery.schema.json](./discovery.schema.json) | Search query/result | `tooldiscovery/search` |
| [config.schema.json](./config.schema.json) | Configuration file | `metatools-mcp` |

## Usage

### Validation in Go

```go
import "github.com/xeipuuv/gojsonschema"

schemaLoader := gojsonschema.NewReferenceLoader("file:///path/to/tool.schema.json")
documentLoader := gojsonschema.NewGoLoader(myTool)

result, err := gojsonschema.Validate(schemaLoader, documentLoader)
```

### Validation in TypeScript

```typescript
import Ajv from "ajv";
import toolSchema from "./tool.schema.json";

const ajv = new Ajv();
const validate = ajv.compile(toolSchema);
const valid = validate(myTool);
```

## Schema URIs

All schemas are published at:
- `https://aperturestack.dev/schemas/{name}.schema.json`

## Versioning

Schemas follow semantic versioning. Breaking changes increment the major version and are published as new files (e.g., `tool.v2.schema.json`).
```

---

## Verification Checklist

- [ ] All 5 schema files created
- [ ] JSON syntax is valid
- [ ] Schema `$id` URIs are consistent
- [ ] Cross-references use correct paths
- [ ] README documents all schemas
- [ ] Go package mappings are accurate

**Validation Commands:**
```bash
# Validate JSON syntax
for f in schemas/*.json; do
  python3 -m json.tool "$f" > /dev/null && echo "✓ $f" || echo "✗ $f"
done

# Validate schema structure (requires ajv-cli)
npx ajv validate -s schemas/tool.schema.json -d /dev/null --strict=false
```

---

## Acceptance Criteria

1. All schemas are valid JSON Schema draft 2020-12
2. Schemas accurately represent Go types
3. Cross-references resolve correctly
4. Documentation is complete

---

## Rollback Plan

```bash
rm -rf schemas/
```

---

## Next Steps

- PRD-110: Repository Creation
- PRD-120: Migrate toolmodel (implement schema validation)
