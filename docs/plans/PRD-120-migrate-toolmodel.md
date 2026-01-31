# PRD-120: Migrate toolmodel

**Phase:** 2 - Foundation Layer
**Priority:** Critical
**Effort:** 4 hours
**Dependencies:** PRD-110

---

## Objective

Migrate the existing `toolmodel` repository into `toolfoundation/model/` as the first package in the consolidated foundation layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolmodel`
**Target Location:** `github.com/ApertureStack/toolfoundation/model`

**Package Contents:**
- Core tool schema types (`Tool`, `ToolInput`, `ToolOutput`)
- JSON Schema validation
- Serialization helpers
- ~2,500 lines of code
- 89.6% test coverage

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Model Package | `toolfoundation/model/` | Migrated tool schema types |
| Tests | `toolfoundation/model/*_test.go` | All existing tests |
| Documentation | `toolfoundation/model/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone Source Repository

```bash
# Create working directory
mkdir -p /tmp/migration
cd /tmp/migration

# Clone source with full history
git clone git@github.com:ApertureStack/toolmodel.git
cd toolmodel

# Verify contents
ls -la
go test ./...
```

### Task 2: Prepare Target Repository

```bash
# Clone target
cd /tmp/migration
git clone git@github.com:ApertureStack/toolfoundation.git
cd toolfoundation

# Create model directory
mkdir -p model

# Verify go.mod exists
cat go.mod
```

### Task 3: Copy Source Files

```bash
cd /tmp/migration

# Copy Go source files (exclude go.mod, go.sum, .git)
cp toolmodel/*.go toolfoundation/model/

# List copied files
ls -la toolfoundation/model/
```

### Task 4: Update Import Paths

**Script:** `scripts/update-imports.sh`

```bash
#!/bin/bash
set -euo pipefail

OLD_IMPORT="github.com/ApertureStack/toolmodel"
NEW_IMPORT="github.com/ApertureStack/toolfoundation/model"

cd toolfoundation/model

# Update all Go files
for file in *.go; do
  if [ -f "$file" ]; then
    sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
    echo "Updated: $file"
  fi
done

# Verify no old imports remain
grep -r "ApertureStack/toolmodel" . && echo "⚠ Old imports still exist!" || echo "✓ All imports updated"
```

### Task 5: Update Package Documentation

**File:** `toolfoundation/model/doc.go`

```go
// Package model provides the canonical tool schema types for the ApertureStack ecosystem.
//
// This package defines the core data structures used to represent tools across
// all layers of the stack, including:
//
//   - Tool: The primary tool definition with name, description, and schema
//   - ToolInput: JSON Schema-based input parameter definitions
//   - ToolOutput: Output type definitions and validation
//   - Metadata: Extensible tool metadata
//
// # Usage
//
// Create a new tool definition:
//
//	tool := model.Tool{
//	    ID:          "calculator",
//	    Name:        "Calculator",
//	    Description: "Performs arithmetic operations",
//	    InputSchema: model.InputSchema{
//	        Type: "object",
//	        Properties: map[string]model.Property{
//	            "operation": {Type: "string", Enum: []string{"add", "subtract"}},
//	            "a":         {Type: "number"},
//	            "b":         {Type: "number"},
//	        },
//	        Required: []string{"operation", "a", "b"},
//	    },
//	}
//
// Validate tool input:
//
//	if err := tool.ValidateInput(input); err != nil {
//	    return fmt.Errorf("invalid input: %w", err)
//	}
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolmodel as part of
// the ApertureStack consolidation. The API remains unchanged.
package model
```

### Task 6: Update go.mod Dependencies

```bash
cd toolfoundation

# Tidy dependencies
go mod tidy

# Verify build
go build ./...

# Run tests
go test -v ./model/...
```

### Task 7: Verify Test Coverage

```bash
cd toolfoundation

# Run tests with coverage
go test -coverprofile=coverage.out ./model/...

# Check coverage percentage (should be >= 89%)
go tool cover -func=coverage.out | grep total

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Task 8: Commit and Push

```bash
cd toolfoundation

git add -A
git commit -m "feat(model): migrate toolmodel package

Migrate the canonical tool schema types from standalone toolmodel repository.

Package contents:
- Tool, ToolInput, ToolOutput types
- JSON Schema validation
- Serialization helpers
- Full test coverage (89.6%)

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolmodel → toolfoundation/model

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

### Task 9: Update Dependent Repositories

Create a tracking issue for updating imports in dependent repos:

```bash
gh issue create -R ApertureStack/toolfoundation \
  --title "Update dependent repos to use toolfoundation/model" \
  --body "## Dependent Repositories

The following repositories need import updates:

- [ ] toolindex
- [ ] toolsearch
- [ ] toolrun
- [ ] toolcode
- [ ] metatools-mcp

## Old Import
\`\`\`go
import \"github.com/ApertureStack/toolmodel\"
\`\`\`

## New Import
\`\`\`go
import \"github.com/ApertureStack/toolfoundation/model\"
\`\`\`

## Timeline
Will be updated as part of each repo's migration PRD."
```

---

## File Mapping

| Source | Target |
|--------|--------|
| `toolmodel/tool.go` | `toolfoundation/model/tool.go` |
| `toolmodel/tool_test.go` | `toolfoundation/model/tool_test.go` |
| `toolmodel/schema.go` | `toolfoundation/model/schema.go` |
| `toolmodel/schema_test.go` | `toolfoundation/model/schema_test.go` |
| `toolmodel/validate.go` | `toolfoundation/model/validate.go` |
| `toolmodel/validate_test.go` | `toolfoundation/model/validate_test.go` |
| `toolmodel/doc.go` | `toolfoundation/model/doc.go` |

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Coverage >= 89%
- [ ] Package documentation updated
- [ ] Committed with proper message
- [ ] Pushed to main
- [ ] Dependent repos tracked

---

## Acceptance Criteria

1. `toolfoundation/model` package builds successfully
2. All tests pass with >= 89% coverage
3. No references to old import path
4. Package can be imported by other repos
5. API is unchanged from original toolmodel

**Verification:**
```go
// This should work in any dependent package
import "github.com/ApertureStack/toolfoundation/model"

func example() {
    tool := model.Tool{
        ID:   "test",
        Name: "Test Tool",
    }
    _ = tool
}
```

---

## Rollback Plan

```bash
cd toolfoundation

# Remove model package
rm -rf model/

# Reset to previous state
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-121: Migrate tooladapter
- PRD-122: Create toolversion
