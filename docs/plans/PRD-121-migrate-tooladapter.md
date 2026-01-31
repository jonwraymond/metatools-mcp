# PRD-121: Migrate tooladapter

**Phase:** 2 - Foundation Layer  
**Priority:** Critical  
**Effort:** 4 hours  
**Dependencies:** PRD-120  
**Status:** Done (2026-01-31)

---

## Objective

Migrate the existing `tooladapter` repository into `toolfoundation/adapter/` as the second package in the consolidated foundation layer.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/tooladapter`
**Target Location:** `github.com/jonwraymond/toolfoundation/adapter`

**Package Contents:**
- Schema adapters for converting between tool formats
- MCP ↔ OpenAI ↔ Anthropic format conversion
- LangChain tool format support
- ~1,500 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Adapter Package | `toolfoundation/adapter/` | Schema format adapters |
| Tests | `toolfoundation/adapter/*_test.go` | All existing tests |
| Documentation | `toolfoundation/adapter/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone Source Repository

```bash
cd /tmp/migration

# Clone source with full history
git clone git@github.com:jonwraymond/tooladapter.git
cd tooladapter

# Verify contents
ls -la
go test ./...
```

### Task 2: Copy Source Files

```bash
cd /tmp/migration/toolfoundation

# Create adapter directory
mkdir -p adapter

# Copy Go source files
cp ../tooladapter/*.go adapter/

# List copied files
ls -la adapter/
```

### Task 3: Update Import Paths

```bash
cd /tmp/migration/toolfoundation/adapter

OLD_IMPORT="github.com/jonwraymond/tooladapter"
NEW_IMPORT="github.com/jonwraymond/toolfoundation/adapter"

# Update all Go files
for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
  echo "Updated: $file"
done

# Also update toolmodel import to toolfoundation/model
OLD_MODEL="github.com/jonwraymond/toolmodel"
NEW_MODEL="github.com/jonwraymond/toolfoundation/model"

for file in *.go; do
  sed -i '' "s|$OLD_MODEL|$NEW_MODEL|g" "$file"
done

# Verify
grep -r "jonwraymond/tooladapter\|jonwraymond/toolmodel" . || echo "✓ All imports updated"
```

### Task 4: Update Package Documentation

**File:** `toolfoundation/adapter/doc.go`

```go
// Package adapter provides schema format conversion between different AI tool specifications.
//
// This package enables interoperability between various AI tool formats:
//
//   - MCP (Model Context Protocol) - Anthropic's standard
//   - OpenAI Function Calling format
//   - Anthropic Tool Use format
//   - LangChain Tool format
//
// # Adapters
//
// Each adapter implements bidirectional conversion:
//
//	// Convert MCP tool to OpenAI format
//	openaiTool := adapter.ToOpenAI(mcpTool)
//
//	// Convert OpenAI tool to MCP format
//	mcpTool := adapter.FromOpenAI(openaiTool)
//
// # Supported Formats
//
//	| Format     | To MCP | From MCP |
//	|------------|--------|----------|
//	| OpenAI     | ✓      | ✓        |
//	| Anthropic  | ✓      | ✓        |
//	| LangChain  | ✓      | ✓        |
//
// # Schema Compatibility
//
// Not all schema features are supported across formats. The adapter handles:
//
//   - Parameter type mapping
//   - Required field conversion
//   - Description propagation
//   - Enum value handling
//
// Features not supported in target format are gracefully degraded.
//
// # Migration Note
//
// This package was migrated from github.com/jonwraymond/tooladapter as part of
// the ApertureStack consolidation.
package adapter
```

### Task 5: Verify Internal Dependencies

The adapter package depends on the model package. Verify the import works:

```bash
cd /tmp/migration/toolfoundation

# Check imports in adapter files
grep -h "import" adapter/*.go | sort -u

# Should include:
# "github.com/jonwraymond/toolfoundation/model"
```

### Task 6: Update go.mod and Build

```bash
cd /tmp/migration/toolfoundation

# Tidy dependencies
go mod tidy

# Verify build
go build ./...

# Run all tests
go test -v ./...
```

### Task 7: Verify Test Coverage

```bash
cd /tmp/migration/toolfoundation

# Run adapter tests with coverage
go test -coverprofile=adapter_coverage.out ./adapter/...

# Check coverage
go tool cover -func=adapter_coverage.out | grep total
```

### Task 8: Commit and Push

```bash
cd /tmp/migration/toolfoundation

git add -A
git commit -m "feat(adapter): migrate tooladapter package

Migrate schema format adapters from standalone tooladapter repository.

Package contents:
- MCP ↔ OpenAI format conversion
- MCP ↔ Anthropic format conversion
- MCP ↔ LangChain format conversion
- Bidirectional adapters with graceful degradation

Internal dependency:
- Uses toolfoundation/model for canonical types

This is part of the ApertureStack consolidation effort.

Migration: github.com/jonwraymond/tooladapter → toolfoundation/adapter

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Adapter Interface Design

The migrated adapter package should follow this interface pattern:

```go
package adapter

import "github.com/jonwraymond/toolfoundation/model"

// Adapter converts between MCP tool format and external formats.
type Adapter interface {
    // ToExternal converts an MCP tool to the external format.
    ToExternal(tool model.Tool) (any, error)

    // FromExternal converts an external format tool to MCP.
    FromExternal(external any) (model.Tool, error)

    // Name returns the adapter name (e.g., "openai", "anthropic").
    Name() string
}

// OpenAIAdapter converts between MCP and OpenAI function calling format.
type OpenAIAdapter struct{}

func (a *OpenAIAdapter) ToExternal(tool model.Tool) (any, error) { ... }
func (a *OpenAIAdapter) FromExternal(external any) (model.Tool, error) { ... }
func (a *OpenAIAdapter) Name() string { return "openai" }

// AnthropicAdapter converts between MCP and Anthropic tool use format.
type AnthropicAdapter struct{}

// LangChainAdapter converts between MCP and LangChain tool format.
type LangChainAdapter struct{}
```

---

## File Mapping

| Source | Target |
|--------|--------|
| `tooladapter/adapter.go` | `toolfoundation/adapter/adapter.go` |
| `tooladapter/adapter_test.go` | `toolfoundation/adapter/adapter_test.go` |
| `tooladapter/openai.go` | `toolfoundation/adapter/openai.go` |
| `tooladapter/openai_test.go` | `toolfoundation/adapter/openai_test.go` |
| `tooladapter/anthropic.go` | `toolfoundation/adapter/anthropic.go` |
| `tooladapter/langchain.go` | `toolfoundation/adapter/langchain.go` |
| `tooladapter/doc.go` | `toolfoundation/adapter/doc.go` |

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated (both adapter and model)
- [ ] Internal dependency on model/ works
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Package documentation updated
- [ ] Committed with proper message
- [ ] Pushed to main

---

## Acceptance Criteria

1. `toolfoundation/adapter` package builds successfully
2. All tests pass
3. Can import both `model` and `adapter` from same repo
4. Adapter interface is well-defined
5. No references to old import paths

**Verification:**
```go
import (
    "github.com/jonwraymond/toolfoundation/model"
    "github.com/jonwraymond/toolfoundation/adapter"
)

func example() {
    tool := model.Tool{ID: "test", Name: "Test"}
    openaiAdapter := &adapter.OpenAIAdapter{}
    external, _ := openaiAdapter.ToExternal(tool)
    _ = external
}
```

---

## Completion Evidence

- `toolfoundation/adapter/` contains migrated sources and tests.
- `toolfoundation/adapter/doc.go` documents the adapter contract.
- `go test ./adapter/...` passes in `toolfoundation`.

---

## Rollback Plan

```bash
cd /tmp/migration/toolfoundation

# Remove adapter package
rm -rf adapter/

# Reset to previous state
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-122: Create toolversion
- Gate G2: Foundation layer complete
