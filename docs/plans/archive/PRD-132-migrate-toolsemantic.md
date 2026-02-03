# PRD-132: Migrate toolsemantic

**Phase:** 3 - Discovery Layer  
**Priority:** High  
**Effort:** 6 hours  
**Dependencies:** PRD-131  
**Status:** Done (2026-01-31)

---

## Objective

Migrate and complete the partial `toolsemantic` implementation into `tooldiscovery/semantic/` for vector-based semantic search capabilities.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/toolsemantic` (partial implementation)
**Target Location:** `github.com/jonwraymond/tooldiscovery/semantic`

**Current State:**
- Partial implementation with interfaces defined
- Embedder interface for text-to-vector conversion
- Vector index interface for similarity search
- Needs completion: actual implementations

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Semantic Package | `tooldiscovery/semantic/` | Vector search implementation |
| Embedder Interface | `semantic/embedder.go` | Text embedding abstraction |
| Vector Index | `semantic/index.go` | Vector similarity search |
| Hybrid Searcher | `semantic/hybrid.go` | Combined BM25 + vector search |
| Tests | `semantic/*_test.go` | Comprehensive tests |

---

## Tasks

### Task 1: Create Package Structure

```bash
cd /tmp/migration/tooldiscovery

mkdir -p semantic
```

### Task 2: Copy Existing Code

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolsemantic.git
cd toolsemantic

# Copy existing files
cp *.go ../tooldiscovery/semantic/

# Update imports
cd ../tooldiscovery/semantic
sed -i '' 's|github.com/jonwraymond/toolsemantic|github.com/jonwraymond/tooldiscovery/semantic|g' *.go
sed -i '' 's|github.com/jonwraymond/toolmodel|github.com/jonwraymond/toolfoundation/model|g' *.go
```

### Task 3: Define Core Interfaces

**File:** `tooldiscovery/semantic/embedder.go`

```go
package semantic

import "context"

// Embedder converts text to vector embeddings.
type Embedder interface {
    // Embed converts text to a vector embedding.
    Embed(ctx context.Context, text string) ([]float32, error)

    // EmbedBatch converts multiple texts to embeddings.
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Dimension returns the embedding dimension.
    Dimension() int

    // Name returns the embedder name (e.g., "openai", "cohere").
    Name() string
}

// EmbedderConfig is the base configuration for embedders.
type EmbedderConfig struct {
    Model     string
    Dimension int
    BatchSize int
}
```

### Task 4: Define Vector Index Interface

**File:** `tooldiscovery/semantic/index.go`

```go
package semantic

import (
    "context"
)

// VectorIndex stores and searches vector embeddings.
type VectorIndex interface {
    // Add adds a vector with its ID.
    Add(ctx context.Context, id string, vector []float32) error

    // AddBatch adds multiple vectors.
    AddBatch(ctx context.Context, ids []string, vectors [][]float32) error

    // Search finds the k most similar vectors to the query.
    Search(ctx context.Context, query []float32, k int) ([]SearchResult, error)

    // Remove deletes a vector by ID.
    Remove(ctx context.Context, id string) error

    // Count returns the number of vectors.
    Count(ctx context.Context) (int, error)
}

// SearchResult represents a similarity search result.
type SearchResult struct {
    ID       string
    Score    float32
    Distance float32
}

// VectorIndexConfig configures the vector index.
type VectorIndexConfig struct {
    Dimension    int
    Metric       string // "cosine", "euclidean", "dot"
    MaxElements  int
    EfSearch     int // HNSW search parameter
    EfConstruct  int // HNSW construction parameter
}
```

### Task 5: Implement In-Memory Vector Index

**File:** `tooldiscovery/semantic/memory_index.go`

```go
package semantic

import (
    "context"
    "fmt"
    "math"
    "sort"
    "sync"
)

// MemoryVectorIndex is an in-memory vector index using brute-force search.
// Suitable for small to medium datasets (<10k vectors).
type MemoryVectorIndex struct {
    vectors   map[string][]float32
    dimension int
    metric    string
    mu        sync.RWMutex
}

// NewMemoryVectorIndex creates a new in-memory vector index.
func NewMemoryVectorIndex(config VectorIndexConfig) *MemoryVectorIndex {
    return &MemoryVectorIndex{
        vectors:   make(map[string][]float32),
        dimension: config.Dimension,
        metric:    config.Metric,
    }
}

func (m *MemoryVectorIndex) Add(ctx context.Context, id string, vector []float32) error {
    if len(vector) != m.dimension {
        return fmt.Errorf("expected dimension %d, got %d", m.dimension, len(vector))
    }
    m.mu.Lock()
    defer m.mu.Unlock()
    m.vectors[id] = vector
    return nil
}

func (m *MemoryVectorIndex) AddBatch(ctx context.Context, ids []string, vectors [][]float32) error {
    if len(ids) != len(vectors) {
        return fmt.Errorf("ids and vectors length mismatch")
    }
    for i, id := range ids {
        if err := m.Add(ctx, id, vectors[i]); err != nil {
            return err
        }
    }
    return nil
}

func (m *MemoryVectorIndex) Search(ctx context.Context, query []float32, k int) ([]SearchResult, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    results := make([]SearchResult, 0, len(m.vectors))
    for id, vector := range m.vectors {
        score := m.similarity(query, vector)
        results = append(results, SearchResult{
            ID:    id,
            Score: score,
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    if k < len(results) {
        results = results[:k]
    }
    return results, nil
}

func (m *MemoryVectorIndex) Remove(ctx context.Context, id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.vectors, id)
    return nil
}

func (m *MemoryVectorIndex) Count(ctx context.Context) (int, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return len(m.vectors), nil
}

func (m *MemoryVectorIndex) similarity(a, b []float32) float32 {
    switch m.metric {
    case "cosine":
        return cosineSimilarity(a, b)
    case "dot":
        return dotProduct(a, b)
    default:
        return cosineSimilarity(a, b)
    }
}

func cosineSimilarity(a, b []float32) float32 {
    var dot, normA, normB float32
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func dotProduct(a, b []float32) float32 {
    var dot float32
    for i := range a {
        dot += a[i] * b[i]
    }
    return dot
}
```

### Task 6: Implement Hybrid Searcher

**File:** `tooldiscovery/semantic/hybrid.go`

```go
package semantic

import (
    "context"
    "sort"

    "github.com/jonwraymond/toolfoundation/model"
    "github.com/jonwraymond/tooldiscovery/index"
)

// HybridSearcher combines BM25 and semantic search.
type HybridSearcher struct {
    bm25Searcher   index.Searcher
    embedder       Embedder
    vectorIndex    VectorIndex
    bm25Weight     float32
    semanticWeight float32
}

// HybridConfig configures the hybrid searcher.
type HybridConfig struct {
    BM25Searcher   index.Searcher
    Embedder       Embedder
    VectorIndex    VectorIndex
    BM25Weight     float32 // default: 0.5
    SemanticWeight float32 // default: 0.5
}

// NewHybridSearcher creates a hybrid BM25 + semantic searcher.
func NewHybridSearcher(config HybridConfig) *HybridSearcher {
    bm25Weight := config.BM25Weight
    semanticWeight := config.SemanticWeight
    if bm25Weight == 0 && semanticWeight == 0 {
        bm25Weight = 0.5
        semanticWeight = 0.5
    }

    return &HybridSearcher{
        bm25Searcher:   config.BM25Searcher,
        embedder:       config.Embedder,
        vectorIndex:    config.VectorIndex,
        bm25Weight:     bm25Weight,
        semanticWeight: semanticWeight,
    }
}

// Search performs hybrid search combining BM25 and semantic results.
func (h *HybridSearcher) Search(ctx context.Context, tools []model.Tool, query string) ([]model.Tool, error) {
    // Index tools if needed
    if err := h.indexTools(ctx, tools); err != nil {
        return nil, err
    }

    // Get BM25 results
    bm25Results, err := h.bm25Searcher.Search(ctx, tools, query)
    if err != nil {
        return nil, err
    }

    // Get semantic results
    queryVec, err := h.embedder.Embed(ctx, query)
    if err != nil {
        return nil, err
    }

    semanticResults, err := h.vectorIndex.Search(ctx, queryVec, len(tools))
    if err != nil {
        return nil, err
    }

    // Combine results using reciprocal rank fusion
    return h.fuseResults(tools, bm25Results, semanticResults), nil
}

func (h *HybridSearcher) indexTools(ctx context.Context, tools []model.Tool) error {
    count, _ := h.vectorIndex.Count(ctx)
    if count >= len(tools) {
        return nil // Already indexed
    }

    texts := make([]string, len(tools))
    ids := make([]string, len(tools))
    for i, tool := range tools {
        texts[i] = tool.Name + " " + tool.Description
        ids[i] = tool.ID
    }

    vectors, err := h.embedder.EmbedBatch(ctx, texts)
    if err != nil {
        return err
    }

    return h.vectorIndex.AddBatch(ctx, ids, vectors)
}

func (h *HybridSearcher) fuseResults(tools []model.Tool, bm25Results []model.Tool, semanticResults []SearchResult) []model.Tool {
    // Reciprocal Rank Fusion
    const k = 60.0 // RRF constant

    scores := make(map[string]float32)

    // BM25 scores
    for i, tool := range bm25Results {
        rank := float32(i + 1)
        scores[tool.ID] += h.bm25Weight * (1.0 / (k + rank))
    }

    // Semantic scores
    for i, result := range semanticResults {
        rank := float32(i + 1)
        scores[result.ID] += h.semanticWeight * (1.0 / (k + rank))
    }

    // Sort by combined score
    type scored struct {
        tool  model.Tool
        score float32
    }
    toolMap := make(map[string]model.Tool)
    for _, t := range tools {
        toolMap[t.ID] = t
    }

    results := make([]scored, 0, len(scores))
    for id, score := range scores {
        if tool, ok := toolMap[id]; ok {
            results = append(results, scored{tool, score})
        }
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].score > results[j].score
    })

    output := make([]model.Tool, len(results))
    for i, r := range results {
        output[i] = r.tool
    }
    return output
}
```

### Task 7: Create Package Documentation

**File:** `tooldiscovery/semantic/doc.go`

```go
// Package semantic provides vector-based semantic search for tool discovery.
//
// This package enables semantic similarity search using vector embeddings,
// complementing the BM25 keyword search in tooldiscovery/search.
//
// # Components
//
//   - Embedder: Converts text to vector embeddings
//   - VectorIndex: Stores and searches vector embeddings
//   - HybridSearcher: Combines BM25 and semantic search
//
// # Usage
//
// Create a hybrid searcher:
//
//	embedder := openai.NewEmbedder(apiKey)
//	vectorIndex := semantic.NewMemoryVectorIndex(semantic.VectorIndexConfig{
//	    Dimension: 1536,
//	    Metric:    "cosine",
//	})
//	bm25 := search.NewBM25Searcher(search.Config{})
//
//	hybrid := semantic.NewHybridSearcher(semantic.HybridConfig{
//	    BM25Searcher:   bm25,
//	    Embedder:       embedder,
//	    VectorIndex:    vectorIndex,
//	    BM25Weight:     0.5,
//	    SemanticWeight: 0.5,
//	})
//
// # Fusion Strategy
//
// The hybrid searcher uses Reciprocal Rank Fusion (RRF) to combine results:
//
//	score(d) = Σ(weight_i / (k + rank_i(d)))
//
// This produces robust rankings without score normalization.
//
// # Embedder Implementations
//
// The package defines the Embedder interface. Implementations are provided
// separately to avoid API key dependencies:
//
//   - OpenAI: text-embedding-3-small/large
//   - Cohere: embed-english-v3
//   - Local: sentence-transformers via gRPC
//
// # Migration Note
//
// This package consolidates and completes the partial toolsemantic
// implementation as part of the ApertureStack consolidation.
package semantic
```

### Task 8: Create Tests

**File:** `tooldiscovery/semantic/semantic_test.go`

```go
package semantic

import (
    "context"
    "testing"
)

func TestMemoryVectorIndex(t *testing.T) {
    ctx := context.Background()
    idx := NewMemoryVectorIndex(VectorIndexConfig{
        Dimension: 3,
        Metric:    "cosine",
    })

    // Add vectors
    err := idx.Add(ctx, "a", []float32{1, 0, 0})
    if err != nil {
        t.Fatal(err)
    }
    err = idx.Add(ctx, "b", []float32{0, 1, 0})
    if err != nil {
        t.Fatal(err)
    }
    err = idx.Add(ctx, "c", []float32{0.9, 0.1, 0})
    if err != nil {
        t.Fatal(err)
    }

    // Search for similar to a
    results, err := idx.Search(ctx, []float32{1, 0, 0}, 2)
    if err != nil {
        t.Fatal(err)
    }

    if len(results) != 2 {
        t.Errorf("expected 2 results, got %d", len(results))
    }
    if results[0].ID != "a" {
        t.Errorf("expected first result to be 'a', got '%s'", results[0].ID)
    }
    if results[1].ID != "c" {
        t.Errorf("expected second result to be 'c', got '%s'", results[1].ID)
    }
}

func TestCosineSimilarity(t *testing.T) {
    tests := []struct {
        a, b []float32
        want float32
    }{
        {[]float32{1, 0}, []float32{1, 0}, 1.0},
        {[]float32{1, 0}, []float32{0, 1}, 0.0},
        {[]float32{1, 0}, []float32{-1, 0}, -1.0},
    }

    for _, tt := range tests {
        got := cosineSimilarity(tt.a, tt.b)
        if got != tt.want {
            t.Errorf("cosineSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
        }
    }
}
```

### Task 9: Build and Test

```bash
cd /tmp/migration/tooldiscovery

go mod tidy
go build ./...
go test -v -coverprofile=semantic_coverage.out ./semantic/...
go tool cover -func=semantic_coverage.out | grep total
```

### Task 10: Commit and Push

```bash
cd /tmp/migration/tooldiscovery

git add -A
git commit -m "feat(semantic): add semantic search package

Add vector-based semantic search capabilities for tool discovery.

Package contents:
- Embedder interface for text-to-vector conversion
- VectorIndex interface for similarity search
- MemoryVectorIndex for in-memory brute-force search
- HybridSearcher combining BM25 + semantic with RRF fusion

Features:
- Cosine and dot product similarity metrics
- Reciprocal Rank Fusion for result combination
- Configurable BM25/semantic weight balance
- Pluggable embedder and vector index backends

This consolidates and completes the partial toolsemantic implementation.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [x] Core interfaces defined (Embedder, VectorIndex)
- [x] MemoryVectorIndex implemented
- [x] HybridSearcher implemented
- [x] Cosine similarity works correctly
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] Package documentation complete

---

## Acceptance Criteria

1. `tooldiscovery/semantic` builds successfully
2. MemoryVectorIndex passes tests
3. HybridSearcher combines BM25 and semantic results
4. RRF fusion produces meaningful rankings
5. Implements `index.Searcher` interface

---

## Completion Evidence

- `tooldiscovery/semantic/` contains interfaces and in-memory implementations.
- `tooldiscovery/semantic/doc.go` documents the package.
- `go test ./semantic/...` passes in `tooldiscovery`.

---

## Rollback Plan

```bash
cd /tmp/migration/tooldiscovery
rm -rf semantic/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-133: Migrate tooldocs
- Gate G3: Discovery + Execution layers complete
