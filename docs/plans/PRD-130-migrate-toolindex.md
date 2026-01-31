# PRD-130: Migrate toolindex

**Phase:** 3 - Discovery Layer
**Priority:** Critical
**Effort:** 4 hours
**Dependencies:** PRD-120

---

## Objective

Migrate the existing `toolindex` repository into `tooldiscovery/index/` as the first package in the consolidated discovery layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolindex`
**Target Location:** `github.com/ApertureStack/tooldiscovery/index`

**Package Contents:**
- Tool registry with CRUD operations
- In-memory and file-based index implementations
- Searcher interface for pluggable search backends
- Progressive disclosure support
- ~3,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Index Package | `tooldiscovery/index/` | Tool registry implementation |
| Tests | `tooldiscovery/index/*_test.go` | All existing tests |
| Documentation | `tooldiscovery/index/doc.go` | Package documentation |

---

## Tasks

### Task 1: Prepare Target Repository

```bash
# Clone/create target
cd /tmp/migration
git clone git@github.com:ApertureStack/tooldiscovery.git
cd tooldiscovery

# Create index directory
mkdir -p index
```

### Task 2: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolindex.git
cd toolindex

# Analyze contents
ls -la
wc -l *.go
go test ./... -v
```

### Task 3: Copy Source Files

```bash
cd /tmp/migration

# Copy Go source files
cp toolindex/*.go tooldiscovery/index/

# Verify
ls -la tooldiscovery/index/
```

### Task 4: Update Import Paths

```bash
cd /tmp/migration/tooldiscovery/index

# Update toolindex import
OLD_IMPORT="github.com/ApertureStack/toolindex"
NEW_IMPORT="github.com/ApertureStack/tooldiscovery/index"

for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
done

# Update toolmodel to toolfoundation/model
OLD_MODEL="github.com/ApertureStack/toolmodel"
NEW_MODEL="github.com/ApertureStack/toolfoundation/model"

for file in *.go; do
  sed -i '' "s|$OLD_MODEL|$NEW_MODEL|g" "$file"
done

# Verify
grep -r "ApertureStack/toolindex\|ApertureStack/toolmodel" . || echo "✓ All imports updated"
```

### Task 5: Update go.mod

```bash
cd /tmp/migration/tooldiscovery

# Add dependency on toolfoundation
cat >> go.mod << EOF
require github.com/ApertureStack/toolfoundation v0.1.0
EOF

go mod tidy
```

### Task 6: Update Package Documentation

**File:** `tooldiscovery/index/doc.go`

```go
// Package index provides the core tool registry for the ApertureStack ecosystem.
//
// This package implements tool registration, storage, retrieval, and search
// capabilities. It supports multiple index backends and pluggable search strategies.
//
// # Index Types
//
// The package provides two built-in index implementations:
//
//   - InMemoryIndex: Fast, ephemeral storage for development and testing
//   - FileIndex: Persistent JSON-based storage for single-node deployments
//
// # Usage
//
// Create and populate an index:
//
//	idx := index.NewInMemoryIndex(index.Options{})
//
//	tool := model.Tool{
//	    ID:          "calculator",
//	    Name:        "Calculator",
//	    Description: "Performs arithmetic",
//	}
//	err := idx.Add(ctx, tool)
//
// Search for tools:
//
//	results, err := idx.Search(ctx, "arithmetic")
//
// # Pluggable Search
//
// The index accepts a custom Searcher for advanced search capabilities:
//
//	searcher := search.NewBM25Searcher(config)
//	idx := index.NewInMemoryIndex(index.Options{
//	    Searcher: searcher,
//	})
//
// # Progressive Disclosure
//
// Tools support three disclosure levels:
//
//   - Summary: ID, name, description only
//   - Schema: Includes input/output schemas
//   - Full: Complete tool definition with examples
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolindex as part of
// the ApertureStack consolidation.
package index
```

### Task 7: Build and Test

```bash
cd /tmp/migration/tooldiscovery

go mod tidy
go build ./...
go test -v -coverprofile=coverage.out ./index/...

# Check coverage
go tool cover -func=coverage.out | grep total
```

### Task 8: Commit and Push

```bash
cd /tmp/migration/tooldiscovery

git add -A
git commit -m "feat(index): migrate toolindex package

Migrate the tool registry from standalone toolindex repository.

Package contents:
- Index interface with CRUD operations
- InMemoryIndex for ephemeral storage
- FileIndex for persistent storage
- Pluggable Searcher interface
- Progressive disclosure support

Dependencies:
- github.com/ApertureStack/toolfoundation/model

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolindex → tooldiscovery/index

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Interfaces

The migrated package should expose these key interfaces:

```go
package index

import (
    "context"
    "github.com/ApertureStack/toolfoundation/model"
)

// Index provides tool storage and retrieval.
type Index interface {
    // Add registers a tool in the index.
    Add(ctx context.Context, tool model.Tool) error

    // Get retrieves a tool by ID.
    Get(ctx context.Context, id string) (model.Tool, error)

    // Remove deletes a tool from the index.
    Remove(ctx context.Context, id string) error

    // List returns all tools.
    List(ctx context.Context) ([]model.Tool, error)

    // Search finds tools matching the query.
    Search(ctx context.Context, query string) ([]model.Tool, error)

    // Count returns the number of tools.
    Count(ctx context.Context) (int, error)
}

// Searcher defines the search strategy interface.
type Searcher interface {
    // Search searches for tools matching the query.
    Search(ctx context.Context, tools []model.Tool, query string) ([]model.Tool, error)
}

// Options configures index behavior.
type Options struct {
    Searcher Searcher
    MaxTools int
}
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] Dependency on toolfoundation works
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Package documentation updated
- [ ] Committed with proper message
- [ ] Pushed to main

---

## Acceptance Criteria

1. `tooldiscovery/index` package builds successfully
2. All tests pass
3. Can import from `toolfoundation/model`
4. Index and Searcher interfaces preserved
5. Progressive disclosure works

---

## Rollback Plan

```bash
cd /tmp/migration/tooldiscovery
rm -rf index/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-131: Migrate toolsearch
- PRD-132: Migrate toolsemantic
