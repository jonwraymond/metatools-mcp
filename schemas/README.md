# ApertureStack JSON Schemas

This directory contains JSON Schema definitions for core data types in the ApertureStack ecosystem.

## Schemas

| Schema | Description | Go Package |
|---|---|---|
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
