# PRD-133: Migrate tooldocs

**Phase:** 3 - Discovery Layer  
**Priority:** Medium  
**Effort:** 4 hours  
**Dependencies:** PRD-120  
**Status:** Done (2026-01-31)

---

## Objective

Migrate the existing `tooldocs` repository into `tooldiscovery/tooldoc/` as the fourth package in the consolidated discovery layer.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/tooldocs`
**Target Location:** `github.com/jonwraymond/tooldiscovery/tooldoc`

**Package Contents:**
- Tool documentation storage and retrieval
- Progressive disclosure (Summary/Schema/Full)
- Example management
- Markdown rendering support
- ~1,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Docs Package | `tooldiscovery/tooldoc/` | Documentation management |
| Tests | `tooldiscovery/tooldoc/*_test.go` | All existing tests |
| Documentation | `tooldiscovery/tooldoc/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/tooldocs.git
cd tooldocs

ls -la
wc -l *.go
go test ./...
```

### Task 2: Copy Source Files

```bash
cd /tmp/migration/tooldiscovery

# Note: using 'docs' as package name (not 'tooldocs')
mkdir -p docs
cp ../tooldocs/*.go docs/

ls -la docs/
```

### Task 3: Update Import Paths

```bash
cd /tmp/migration/tooldiscovery/tooldoc

# Update self-reference
OLD_IMPORT="github.com/jonwraymond/tooldocs"
NEW_IMPORT="github.com/jonwraymond/tooldiscovery/tooldoc"

for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
done

# Update toolmodel to toolfoundation/model
OLD_MODEL="github.com/jonwraymond/toolmodel"
NEW_MODEL="github.com/jonwraymond/toolfoundation/model"

for file in *.go; do
  sed -i '' "s|$OLD_MODEL|$NEW_MODEL|g" "$file"
done

# Verify
grep -r "jonwraymond/tooldocs\|jonwraymond/toolmodel" . || echo "✓ All imports updated"
```

### Task 4: Update Package Documentation

**File:** `tooldiscovery/tooldoc/doc.go`

```go
// Package docs provides documentation storage and retrieval for tools.
//
// This package manages tool documentation with support for progressive disclosure,
// allowing clients to request varying levels of detail based on their needs.
//
// # Disclosure Levels
//
// The package supports three disclosure levels:
//
//   - Summary: Minimal information (ID, name, description)
//   - Schema: Includes input/output JSON schemas
//   - Full: Complete documentation including examples and usage
//
// # Usage
//
// Create a documentation store:
//
//	store := docs.NewStore(docs.Config{
//	    BaseDir: "./tool-docs",
//	})
//
// Get documentation at different levels:
//
//	summary, _ := store.Get(ctx, "calculator", docs.LevelSummary)
//	schema, _ := store.Get(ctx, "calculator", docs.LevelSchema)
//	full, _ := store.Get(ctx, "calculator", docs.LevelFull)
//
// # Storage
//
// Documentation is stored as structured files:
//
//	tool-docs/
//	├── calculator/
//	│   ├── README.md      # Full documentation
//	│   ├── schema.json    # Input/output schemas
//	│   └── examples/
//	│       ├── basic.json
//	│       └── advanced.json
//
// # Examples
//
// The package includes example management:
//
//	examples, _ := store.GetExamples(ctx, "calculator")
//	for _, ex := range examples {
//	    fmt.Printf("Example: %s\n", ex.Name)
//	    fmt.Printf("Input: %s\n", ex.Input)
//	    fmt.Printf("Output: %s\n", ex.Output)
//	}
//
// # Migration Note
//
// This package was migrated from github.com/jonwraymond/tooldocs as part of
// the ApertureStack consolidation.
package docs
```

### Task 5: Define Core Types

Ensure these types exist in the migrated code:

**File:** `tooldiscovery/tooldoc/types.go`

```go
package docs

import (
    "context"
    "github.com/jonwraymond/toolfoundation/model"
)

// Level represents the disclosure level for documentation.
type Level int

const (
    // LevelSummary provides minimal information.
    LevelSummary Level = iota
    // LevelSchema includes input/output schemas.
    LevelSchema
    // LevelFull provides complete documentation.
    LevelFull
)

// Documentation holds tool documentation at various levels.
type Documentation struct {
    Tool        model.Tool
    Level       Level
    Markdown    string
    Examples    []Example
    LastUpdated string
}

// Example represents a tool usage example.
type Example struct {
    Name        string
    Description string
    Input       map[string]any
    Output      any
    Tags        []string
}

// Store provides documentation storage and retrieval.
type Store interface {
    // Get retrieves documentation at the specified level.
    Get(ctx context.Context, toolID string, level Level) (*Documentation, error)

    // Set stores documentation for a tool.
    Set(ctx context.Context, doc *Documentation) error

    // GetExamples retrieves examples for a tool.
    GetExamples(ctx context.Context, toolID string) ([]Example, error)

    // AddExample adds an example to a tool.
    AddExample(ctx context.Context, toolID string, example Example) error

    // List returns all documented tool IDs.
    List(ctx context.Context) ([]string, error)
}

// Config configures the documentation store.
type Config struct {
    BaseDir     string
    CacheSize   int
    EnableCache bool
}
```

### Task 6: Build and Test

```bash
cd /tmp/migration/tooldiscovery

go mod tidy
go build ./...
go test -v -coverprofile=docs_coverage.out ./docs/...

go tool cover -func=docs_coverage.out | grep total
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/tooldiscovery

git add -A
git commit -m "feat(docs): migrate tooldocs package

Migrate tool documentation management from standalone tooldocs repository.

Package contents:
- Store interface for documentation CRUD
- Progressive disclosure levels (Summary/Schema/Full)
- Example management
- File-based storage implementation
- Optional caching

Features:
- Three-tier disclosure: summary, schema, full
- Markdown documentation support
- JSON example storage
- Last-updated tracking

Dependencies:
- github.com/jonwraymond/toolfoundation/model

This is part of the ApertureStack consolidation effort.

Migration: github.com/jonwraymond/tooldocs → tooldiscovery/tooldoc

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## File Mapping

| Source | Target |
|--------|--------|
| `tooldocs/store.go` | `tooldiscovery/tooldoc/store.go` |
| `tooldocs/store_test.go` | `tooldiscovery/tooldoc/store_test.go` |
| `tooldocs/types.go` | `tooldiscovery/tooldoc/types.go` |
| `tooldocs/file.go` | `tooldiscovery/tooldoc/file.go` |
| `tooldocs/cache.go` | `tooldiscovery/tooldoc/cache.go` |
| `tooldocs/doc.go` | `tooldiscovery/tooldoc/doc.go` |

---

## Verification Checklist

- [x] All source files copied
- [x] Import paths updated
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] Progressive disclosure works
- [x] Example management works
- [x] Package documentation updated
- [x] Committed with proper message
- [x] Pushed to main

---

## Acceptance Criteria

1. `tooldiscovery/tooldoc` package builds successfully
2. All tests pass
3. Three disclosure levels work correctly
4. Examples can be stored and retrieved
5. File-based storage persists data

---

## Completion Evidence

- `tooldiscovery/tooldoc/` contains migrated sources and tests.
- `tooldiscovery/tooldoc/doc.go` documents the package.
- `go test ./tooldoc/...` passes in `tooldiscovery`.

---

## Rollback Plan

```bash
cd /tmp/migration/tooldiscovery
rm -rf tooldoc/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- Gate G3: Discovery layer complete (all 4 packages)
- PRD-140: Migrate toolrun
