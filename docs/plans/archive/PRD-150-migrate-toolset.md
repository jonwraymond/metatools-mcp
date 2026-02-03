# PRD-150: Migrate toolset

**Phase:** 5 - Composition Layer
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-121, PRD-130
**Status:** Done (2026-01-31)

---

## Objective

Migrate the existing `toolset` repository into `toolcompose/set/` as the first package in the consolidated composition layer.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/toolset`
**Target Location:** `github.com/jonwraymond/toolcompose/set`

**Package Contents:**
- Tool collection management
- Toolset composition and filtering
- Namespace-based organization
- Permission scoping
- ~1,500 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Set Package | `toolcompose/set/` | Toolset management |
| Tests | `toolcompose/set/*_test.go` | All existing tests |
| Documentation | `toolcompose/set/doc.go` | Package documentation |

---

## Tasks

### Task 1: Prepare Target Repository

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolcompose.git
cd toolcompose

mkdir -p set
```

### Task 2: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolset.git
cd toolset

ls -la
wc -l *.go
go test ./...
```

### Task 3: Copy Source Files

```bash
cd /tmp/migration

cp toolset/*.go toolcompose/set/
ls -la toolcompose/set/
```

### Task 4: Update Import Paths

```bash
cd /tmp/migration/toolcompose/set

# Update self-reference
sed -i '' 's|github.com/jonwraymond/toolset|github.com/jonwraymond/toolcompose/set|g' *.go

# Update dependencies
sed -i '' 's|github.com/jonwraymond/toolmodel|github.com/jonwraymond/toolfoundation/model|g' *.go
sed -i '' 's|github.com/jonwraymond/tooladapter|github.com/jonwraymond/toolfoundation/adapter|g' *.go
sed -i '' 's|github.com/jonwraymond/toolindex|github.com/jonwraymond/tooldiscovery/index|g' *.go

# Verify
grep -r "jonwraymond/toolset\|jonwraymond/toolmodel\|jonwraymond/toolindex" . || echo "✓ All imports updated"
```

### Task 5: Update Package Documentation

**File:** `toolcompose/set/doc.go`

```go
// Package set provides toolset composition and management.
//
// This package enables creating, managing, and filtering collections of tools.
// Toolsets provide a higher-level abstraction over individual tools, supporting
// namespace-based organization and permission scoping.
//
// # Overview
//
// A Toolset is a named collection of tools with shared configuration:
//
//   - Grouping related tools together
//   - Applying common settings (timeout, retries)
//   - Scoping permissions
//   - Filtering by namespace or tags
//
// # Usage
//
// Create a toolset:
//
//	ts := set.New(set.Config{
//	    Name:      "data-tools",
//	    Namespace: "data",
//	})
//
//	ts.Add(fetchTool)
//	ts.Add(transformTool)
//	ts.Add(storeTool)
//
// Filter tools:
//
//	filtered := ts.Filter(set.Filter{
//	    Tags:       []string{"input"},
//	    Namespace:  "data",
//	})
//
// # Composition
//
// Combine multiple toolsets:
//
//	combined := set.Merge(dataTools, apiTools, utilityTools)
//
// Apply restrictions:
//
//	restricted := ts.WithPermissions(set.Permissions{
//	    AllowedTools: []string{"fetch", "transform"},
//	    DeniedTools:  []string{"delete"},
//	})
//
// # Registry Integration
//
// Register toolsets with the index:
//
//	idx.RegisterSet(ctx, toolset)
//	tools, _ := idx.ListBySet(ctx, "data-tools")
//
// # Migration Note
//
// This package was migrated from github.com/jonwraymond/toolset as part of
// the ApertureStack consolidation.
package set
```

### Task 6: Build and Test

```bash
cd /tmp/migration/toolcompose

go mod tidy
go build ./...
go test -v -coverprofile=set_coverage.out ./set/...

go tool cover -func=set_coverage.out | grep total
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/toolcompose

git add -A
git commit -m "feat(set): migrate toolset package

Migrate toolset composition from standalone toolset repository.

Package contents:
- Toolset type for tool collections
- Namespace-based organization
- Tag-based filtering
- Permission scoping
- Toolset merging

Features:
- Add/remove tools from sets
- Filter by namespace, tags, capabilities
- Apply shared configuration
- Restrict permissions per set
- Merge multiple toolsets

Dependencies:
- github.com/jonwraymond/toolfoundation/model
- github.com/jonwraymond/tooldiscovery/index

This is part of the ApertureStack consolidation effort.

Migration: github.com/jonwraymond/toolset → toolcompose/set

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Interfaces

```go
package set

import (
    "context"
    "github.com/jonwraymond/toolfoundation/model"
)

// Toolset represents a collection of tools.
type Toolset struct {
    ID          string
    Name        string
    Description string
    Namespace   string
    Tools       []model.Tool
    Config      Config
    Permissions Permissions
}

// Config defines toolset configuration.
type Config struct {
    Timeout     time.Duration
    Retries     int
    CachePolicy string
    Metadata    map[string]any
}

// Permissions defines access restrictions.
type Permissions struct {
    AllowedTools []string
    DeniedTools  []string
    Roles        []string
}

// Filter defines filtering criteria.
type Filter struct {
    Namespace    string
    Tags         []string
    Capabilities []string
    Pattern      string // Glob pattern for IDs
}

// New creates a new toolset.
func New(cfg Config) *Toolset

// Add adds a tool to the set.
func (ts *Toolset) Add(tool model.Tool) error

// Remove removes a tool from the set.
func (ts *Toolset) Remove(toolID string) error

// Filter returns tools matching the filter.
func (ts *Toolset) Filter(f Filter) []model.Tool

// WithPermissions returns a restricted copy.
func (ts *Toolset) WithPermissions(p Permissions) *Toolset

// Merge combines multiple toolsets.
func Merge(sets ...*Toolset) *Toolset
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Toolset creation works
- [ ] Filtering works
- [ ] Permission scoping works
- [ ] Package documentation updated

---

## Acceptance Criteria

1. `toolcompose/set` package builds successfully
2. All tests pass
3. Toolsets can be created and populated
4. Filtering produces correct results
5. Merge combines tools correctly

## Completion Notes

- Toolset migrated into `toolcompose/set` with builder, filters, policy, and exposure helpers.
- Imports updated to `github.com/jonwraymond/...`.

---

## Rollback Plan

```bash
cd /tmp/migration/toolcompose
rm -rf set/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-151: Complete toolskill
- Gate G4: Composition layer complete
