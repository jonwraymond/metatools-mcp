# PRD-131: Migrate toolsearch

**Phase:** 3 - Discovery Layer
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-130

---

## Objective

Migrate the existing `toolsearch` repository into `tooldiscovery/search/` as the second package in the consolidated discovery layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolsearch`
**Target Location:** `github.com/ApertureStack/tooldiscovery/search`

**Package Contents:**
- BM25 search implementation using Bleve
- Fingerprint-based index caching
- Configurable field boosting
- ~1,500 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Search Package | `tooldiscovery/search/` | BM25 search implementation |
| Tests | `tooldiscovery/search/*_test.go` | All existing tests |
| Documentation | `tooldiscovery/search/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolsearch.git
cd toolsearch

# Analyze
ls -la
wc -l *.go
go test ./... -v
```

### Task 2: Copy Source Files

```bash
cd /tmp/migration/tooldiscovery

mkdir -p search
cp ../toolsearch/*.go search/

ls -la search/
```

### Task 3: Update Import Paths

```bash
cd /tmp/migration/tooldiscovery/search

# Update self-reference
OLD_IMPORT="github.com/ApertureStack/toolsearch"
NEW_IMPORT="github.com/ApertureStack/tooldiscovery/search"

for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
done

# Update toolindex to tooldiscovery/index
OLD_INDEX="github.com/ApertureStack/toolindex"
NEW_INDEX="github.com/ApertureStack/tooldiscovery/index"

for file in *.go; do
  sed -i '' "s|$OLD_INDEX|$NEW_INDEX|g" "$file"
done

# Update toolmodel to toolfoundation/model
OLD_MODEL="github.com/ApertureStack/toolmodel"
NEW_MODEL="github.com/ApertureStack/toolfoundation/model"

for file in *.go; do
  sed -i '' "s|$OLD_MODEL|$NEW_MODEL|g" "$file"
done

# Verify
grep -r "ApertureStack/toolsearch\|ApertureStack/toolindex\|ApertureStack/toolmodel" . || echo "✓ All imports updated"
```

### Task 4: Update Package Documentation

**File:** `tooldiscovery/search/doc.go`

```go
// Package search provides BM25-based full-text search for tool discovery.
//
// This package implements a high-quality search strategy using the BM25 algorithm
// via the Bleve search library. It integrates with tooldiscovery/index via the
// Searcher interface.
//
// # Features
//
//   - BM25 ranking algorithm for relevance scoring
//   - Fingerprint-based index caching for efficiency
//   - Configurable field boosting
//   - MaxDocs limiting for resource bounds
//   - Text length truncation
//
// # Usage
//
// Create a BM25 searcher and inject into index:
//
//	searcher := search.NewBM25Searcher(search.Config{
//	    NameBoost:        2.0,
//	    DescriptionBoost: 1.0,
//	    MaxDocs:          1000,
//	    MaxDocTextLen:    10000,
//	})
//
//	idx := index.NewInMemoryIndex(index.Options{
//	    Searcher: searcher,
//	})
//
// # BM25 Behavior
//
// The searcher provides three guarantees:
//
//   - Deterministic: Sorted by ID before indexing, consistent tie-breaking
//   - Efficient: Fingerprint-based caching, rebuild only when tools change
//   - Bounded: MaxDocs and MaxDocTextLen prevent resource exhaustion
//
// # Configuration
//
//	type Config struct {
//	    NameBoost        float64 // Boost for name field (default: 2.0)
//	    DescriptionBoost float64 // Boost for description (default: 1.0)
//	    TagBoost         float64 // Boost for tags (default: 1.5)
//	    MaxDocs          int     // Maximum indexed documents
//	    MaxDocTextLen    int     // Maximum text length per field
//	}
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolsearch as part of
// the ApertureStack consolidation.
package search
```

### Task 5: Verify Bleve Dependency

```bash
cd /tmp/migration/tooldiscovery

# Check for Bleve import
grep -r "blevesearch/bleve" search/*.go

# Add to go.mod if needed
go get github.com/blevesearch/bleve/v2
go mod tidy
```

### Task 6: Build and Test

```bash
cd /tmp/migration/tooldiscovery

go build ./...
go test -v -coverprofile=search_coverage.out ./search/...

# Check coverage
go tool cover -func=search_coverage.out | grep total
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/tooldiscovery

git add -A
git commit -m "feat(search): migrate toolsearch package

Migrate BM25 search implementation from standalone toolsearch repository.

Package contents:
- BM25Searcher implementing index.Searcher interface
- Fingerprint-based index caching
- Configurable field boosting
- MaxDocs and MaxDocTextLen bounds

Dependencies:
- github.com/blevesearch/bleve/v2
- github.com/ApertureStack/toolfoundation/model
- github.com/ApertureStack/tooldiscovery/index (interface only)

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolsearch → tooldiscovery/search

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Types

```go
package search

import (
    "context"
    "github.com/ApertureStack/toolfoundation/model"
)

// Config configures the BM25 searcher.
type Config struct {
    NameBoost        float64
    DescriptionBoost float64
    TagBoost         float64
    MaxDocs          int
    MaxDocTextLen    int
    CacheDir         string
}

// BM25Searcher implements index.Searcher using BM25 algorithm.
type BM25Searcher struct {
    config      Config
    fingerprint string
    index       bleve.Index
    mu          sync.RWMutex
}

// NewBM25Searcher creates a new BM25 searcher.
func NewBM25Searcher(config Config) *BM25Searcher

// Search implements index.Searcher.
func (s *BM25Searcher) Search(ctx context.Context, tools []model.Tool, query string) ([]model.Tool, error)

// Fingerprint returns the current document fingerprint.
func (s *BM25Searcher) Fingerprint() string
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated (search, index, model)
- [ ] Bleve dependency resolved
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Fingerprint caching works
- [ ] Package documentation updated
- [ ] Committed with proper message
- [ ] Pushed to main

---

## Acceptance Criteria

1. `tooldiscovery/search` package builds successfully
2. All tests pass
3. BM25 search returns relevant results
4. Fingerprint caching is efficient
5. Implements `index.Searcher` interface

---

## Rollback Plan

```bash
cd /tmp/migration/tooldiscovery
rm -rf search/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-132: Migrate toolsemantic
- PRD-133: Migrate tooldocs
