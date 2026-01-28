# PRD-013: toolsemantic Library Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a hybrid semantic search library combining BM25 lexical search, vector embeddings, and reranking for intelligent tool discovery.

**Architecture:** Three-stage pipeline: BM25 candidate retrieval, vector embedding similarity, and cross-encoder reranking. Pluggable embedding providers and vector stores.

**Tech Stack:** Go, toolsearch dependency, OpenAI/Anthropic embeddings API, optional pgvector/Qdrant

---

## Overview

The `toolsemantic` library extends toolsearch with semantic understanding, enabling natural language queries like "find tools for working with git repositories" to match tools even without exact keyword matches.

**Reference:** [ROADMAP.md](../proposals/ROADMAP.md) - toolsemantic specification

---

## Directory Structure

```
toolsemantic/
├── semantic.go          # HybridSearcher type
├── semantic_test.go
├── embedding.go         # Embedding interface
├── embedding_test.go
├── embeddings/
│   ├── openai.go        # OpenAI embeddings
│   ├── anthropic.go     # Anthropic embeddings
│   └── local.go         # Local embeddings (e.g., all-MiniLM)
├── vector.go            # VectorStore interface
├── vector_test.go
├── stores/
│   ├── memory.go        # In-memory vector store
│   ├── memory_test.go
│   └── pgvector.go      # PostgreSQL pgvector
├── rerank.go            # Reranker interface
├── rerank_test.go
├── rerankers/
│   ├── cross_encoder.go # Cross-encoder reranker
│   └── rrf.go           # Reciprocal Rank Fusion
├── config.go            # Configuration
├── doc.go
├── go.mod
└── go.sum
```

---

## Task 1: HybridSearcher and Configuration

**Files:**
- Create: `toolsemantic/semantic.go`
- Create: `toolsemantic/semantic_test.go`
- Create: `toolsemantic/go.mod`
- Create: `toolsemantic/config.go`

**Step 1: Write failing tests**

```go
// semantic_test.go
package toolsemantic_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolsemantic"
)

func TestHybridSearcher_New(t *testing.T) {
    config := toolsemantic.Config{
        BM25Weight:   0.3,
        VectorWeight: 0.5,
        RerankWeight: 0.2,
    }

    searcher, err := toolsemantic.NewHybridSearcher(config, nil, nil, nil)
    require.NoError(t, err)
    assert.NotNil(t, searcher)
}

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  toolsemantic.Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: toolsemantic.Config{
                BM25Weight:   0.3,
                VectorWeight: 0.5,
                RerankWeight: 0.2,
            },
            wantErr: false,
        },
        {
            name: "weights don't sum to 1",
            config: toolsemantic.Config{
                BM25Weight:   0.5,
                VectorWeight: 0.5,
                RerankWeight: 0.5,
            },
            wantErr: true,
        },
        {
            name: "negative weight",
            config: toolsemantic.Config{
                BM25Weight:   -0.1,
                VectorWeight: 0.6,
                RerankWeight: 0.5,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestDefaultConfig(t *testing.T) {
    config := toolsemantic.DefaultConfig()

    assert.InDelta(t, 1.0, config.BM25Weight+config.VectorWeight+config.RerankWeight, 0.01)
    assert.Greater(t, config.TopK, 0)
    assert.Greater(t, config.CandidateMultiplier, 0)
}

func TestHybridSearcher_Search_BM25Only(t *testing.T) {
    // BM25-only search (no embedder or vector store)
    config := toolsemantic.Config{
        BM25Weight:   1.0,
        VectorWeight: 0.0,
        RerankWeight: 0.0,
        TopK:         10,
    }

    bm25 := &MockBM25Searcher{
        results: []toolsemantic.SearchResult{
            {ID: "tool1", Score: 0.9},
            {ID: "tool2", Score: 0.8},
        },
    }

    searcher, _ := toolsemantic.NewHybridSearcher(config, bm25, nil, nil)

    results, err := searcher.Search(context.Background(), "test query", 10)
    require.NoError(t, err)
    assert.Len(t, results, 2)
    assert.Equal(t, "tool1", results[0].ID)
}

// MockBM25Searcher for testing
type MockBM25Searcher struct {
    results []toolsemantic.SearchResult
}

func (m *MockBM25Searcher) Search(query string, limit int) ([]toolsemantic.SearchResult, error) {
    return m.results, nil
}

func (m *MockBM25Searcher) Close() error { return nil }
```

**Step 2: Run tests to verify they fail**

Run: `cd toolsemantic && go test ./... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// go.mod
module github.com/jrraymond/toolsemantic

go 1.22

require (
    github.com/jrraymond/toolsearch v0.1.9
)
```

```go
// config.go
package toolsemantic

import (
    "errors"
    "math"
)

// Config holds hybrid search configuration
type Config struct {
    // Weight distribution (must sum to 1.0)
    BM25Weight   float64 `yaml:"bm25_weight"`
    VectorWeight float64 `yaml:"vector_weight"`
    RerankWeight float64 `yaml:"rerank_weight"`

    // Search parameters
    TopK                int `yaml:"top_k"`
    CandidateMultiplier int `yaml:"candidate_multiplier"` // Candidates = TopK * Multiplier

    // Embedding config
    EmbeddingModel    string `yaml:"embedding_model"`
    EmbeddingProvider string `yaml:"embedding_provider"` // openai, anthropic, local

    // Vector store config
    VectorStore string `yaml:"vector_store"` // memory, pgvector, qdrant

    // Reranker config
    Reranker      string `yaml:"reranker"`       // none, cross_encoder, rrf
    RerankerModel string `yaml:"reranker_model"` // For cross-encoder
}

// Validate validates the configuration
func (c Config) Validate() error {
    // Check weights
    if c.BM25Weight < 0 || c.VectorWeight < 0 || c.RerankWeight < 0 {
        return errors.New("weights cannot be negative")
    }

    totalWeight := c.BM25Weight + c.VectorWeight + c.RerankWeight
    if math.Abs(totalWeight-1.0) > 0.01 {
        return errors.New("weights must sum to 1.0")
    }

    return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
    return Config{
        BM25Weight:          0.3,
        VectorWeight:        0.5,
        RerankWeight:        0.2,
        TopK:                10,
        CandidateMultiplier: 5,
        EmbeddingProvider:   "openai",
        EmbeddingModel:      "text-embedding-3-small",
        VectorStore:         "memory",
        Reranker:            "rrf",
    }
}
```

```go
// semantic.go
package toolsemantic

import (
    "context"
    "sort"
)

// SearchResult represents a search result
type SearchResult struct {
    ID          string
    Score       float64
    BM25Score   float64
    VectorScore float64
    RerankScore float64
}

// BM25Searcher interface for lexical search
type BM25Searcher interface {
    Search(query string, limit int) ([]SearchResult, error)
    Close() error
}

// Embedder interface for text embedding
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float64, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
    Dimensions() int
}

// VectorStore interface for vector storage
type VectorStore interface {
    Add(ctx context.Context, id string, vector []float64) error
    Search(ctx context.Context, vector []float64, limit int) ([]SearchResult, error)
    Delete(ctx context.Context, id string) error
    Close() error
}

// Reranker interface for result reranking
type Reranker interface {
    Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)
}

// HybridSearcher combines BM25, vector, and reranking
type HybridSearcher struct {
    config   Config
    bm25     BM25Searcher
    embedder Embedder
    vector   VectorStore
    reranker Reranker
}

// NewHybridSearcher creates a new hybrid searcher
func NewHybridSearcher(
    config Config,
    bm25 BM25Searcher,
    embedder Embedder,
    vector VectorStore,
) (*HybridSearcher, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }

    return &HybridSearcher{
        config:   config,
        bm25:     bm25,
        embedder: embedder,
        vector:   vector,
    }, nil
}

// WithReranker adds a reranker to the searcher
func (h *HybridSearcher) WithReranker(reranker Reranker) *HybridSearcher {
    h.reranker = reranker
    return h
}

// Search performs hybrid search
func (h *HybridSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
    // Calculate candidate count
    candidateLimit := limit * h.config.CandidateMultiplier
    if candidateLimit < limit {
        candidateLimit = limit
    }

    // Collect results from each source
    resultMap := make(map[string]*SearchResult)

    // Stage 1: BM25 search
    if h.bm25 != nil && h.config.BM25Weight > 0 {
        bm25Results, err := h.bm25.Search(query, candidateLimit)
        if err != nil {
            return nil, err
        }
        for _, r := range bm25Results {
            if existing, ok := resultMap[r.ID]; ok {
                existing.BM25Score = r.Score
            } else {
                resultMap[r.ID] = &SearchResult{
                    ID:        r.ID,
                    BM25Score: r.Score,
                }
            }
        }
    }

    // Stage 2: Vector search
    if h.embedder != nil && h.vector != nil && h.config.VectorWeight > 0 {
        queryVec, err := h.embedder.Embed(ctx, query)
        if err != nil {
            return nil, err
        }

        vectorResults, err := h.vector.Search(ctx, queryVec, candidateLimit)
        if err != nil {
            return nil, err
        }

        for _, r := range vectorResults {
            if existing, ok := resultMap[r.ID]; ok {
                existing.VectorScore = r.Score
            } else {
                resultMap[r.ID] = &SearchResult{
                    ID:          r.ID,
                    VectorScore: r.Score,
                }
            }
        }
    }

    // Convert to slice and calculate weighted scores
    results := make([]SearchResult, 0, len(resultMap))
    for _, r := range resultMap {
        r.Score = h.config.BM25Weight*r.BM25Score +
            h.config.VectorWeight*r.VectorScore +
            h.config.RerankWeight*r.RerankScore
        results = append(results, *r)
    }

    // Sort by score
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    // Stage 3: Rerank top candidates
    if h.reranker != nil && h.config.RerankWeight > 0 && len(results) > 0 {
        // Rerank top candidates
        reranked, err := h.reranker.Rerank(ctx, query, results)
        if err != nil {
            return nil, err
        }

        // Update scores with rerank scores
        for i := range reranked {
            reranked[i].Score = h.config.BM25Weight*reranked[i].BM25Score +
                h.config.VectorWeight*reranked[i].VectorScore +
                h.config.RerankWeight*reranked[i].RerankScore
        }

        results = reranked
        sort.Slice(results, func(i, j int) bool {
            return results[i].Score > results[j].Score
        })
    }

    // Limit results
    if len(results) > limit {
        results = results[:limit]
    }

    return results, nil
}

// Close closes all underlying resources
func (h *HybridSearcher) Close() error {
    if h.bm25 != nil {
        h.bm25.Close()
    }
    if h.vector != nil {
        h.vector.Close()
    }
    return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolsemantic && go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolsemantic/
git commit -m "$(cat <<'EOF'
feat(toolsemantic): add HybridSearcher and configuration

- HybridSearcher combining BM25, vector, reranking
- Config with weight distribution validation
- Pluggable BM25Searcher, Embedder, VectorStore, Reranker interfaces
- Three-stage search pipeline
- Weighted score combination

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Embedding Interface and OpenAI Provider

**Files:**
- Create: `toolsemantic/embedding.go`
- Create: `toolsemantic/embeddings/openai.go`
- Create: `toolsemantic/embeddings/openai_test.go`

**Step 1: Write failing tests**

```go
// embeddings/openai_test.go
package embeddings_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolsemantic/embeddings"
)

func TestOpenAIEmbedder_Dimensions(t *testing.T) {
    embedder := embeddings.NewOpenAIEmbedder(embeddings.OpenAIConfig{
        APIKey: "test-key",
        Model:  "text-embedding-3-small",
    })

    assert.Equal(t, 1536, embedder.Dimensions())
}

func TestOpenAIEmbedder_DimensionsLarge(t *testing.T) {
    embedder := embeddings.NewOpenAIEmbedder(embeddings.OpenAIConfig{
        APIKey: "test-key",
        Model:  "text-embedding-3-large",
    })

    assert.Equal(t, 3072, embedder.Dimensions())
}

// Note: Actual API tests would require integration testing
func TestOpenAIEmbedder_Config(t *testing.T) {
    config := embeddings.OpenAIConfig{
        APIKey:     "sk-test",
        Model:      "text-embedding-3-small",
        Dimensions: 512,
        BaseURL:    "https://custom.api.com",
    }

    embedder := embeddings.NewOpenAIEmbedder(config)
    assert.NotNil(t, embedder)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolsemantic && go test ./embeddings/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// embedding.go
package toolsemantic

import "context"

// EmbeddingResult holds embedding result
type EmbeddingResult struct {
    Vector     []float64
    TokenCount int
}

// EmbedderConfig is the common embedder configuration
type EmbedderConfig struct {
    Provider   string // openai, anthropic, local
    Model      string
    Dimensions int
    APIKey     string
    BaseURL    string
}
```

```go
// embeddings/openai.go
package embeddings

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

// OpenAIConfig holds OpenAI embeddings configuration
type OpenAIConfig struct {
    APIKey     string
    Model      string
    Dimensions int    // Optional: reduce dimensions
    BaseURL    string // Optional: custom base URL
}

// OpenAIEmbedder generates embeddings using OpenAI API
type OpenAIEmbedder struct {
    config OpenAIConfig
    client *http.Client
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(config OpenAIConfig) *OpenAIEmbedder {
    if config.BaseURL == "" {
        config.BaseURL = "https://api.openai.com/v1"
    }
    if config.Model == "" {
        config.Model = "text-embedding-3-small"
    }

    return &OpenAIEmbedder{
        config: config,
        client: &http.Client{},
    }
}

// Dimensions returns the embedding dimensions
func (e *OpenAIEmbedder) Dimensions() int {
    if e.config.Dimensions > 0 {
        return e.config.Dimensions
    }

    // Default dimensions per model
    switch e.config.Model {
    case "text-embedding-3-large":
        return 3072
    case "text-embedding-3-small":
        return 1536
    case "text-embedding-ada-002":
        return 1536
    default:
        return 1536
    }
}

// Embed generates embedding for a single text
func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
    vectors, err := e.EmbedBatch(ctx, []string{text})
    if err != nil {
        return nil, err
    }
    return vectors[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *OpenAIEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
    reqBody := map[string]any{
        "input": texts,
        "model": e.config.Model,
    }

    if e.config.Dimensions > 0 {
        reqBody["dimensions"] = e.config.Dimensions
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, "POST",
        e.config.BaseURL+"/embeddings", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+e.config.APIKey)

    resp, err := e.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("OpenAI API error: %s", string(respBody))
    }

    var result struct {
        Data []struct {
            Embedding []float64 `json:"embedding"`
            Index     int       `json:"index"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    vectors := make([][]float64, len(texts))
    for _, item := range result.Data {
        vectors[item.Index] = item.Embedding
    }

    return vectors, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolsemantic && go test ./embeddings/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolsemantic/
git commit -m "$(cat <<'EOF'
feat(toolsemantic): add Embedder interface and OpenAI provider

- Embedder interface with Embed and EmbedBatch
- OpenAIEmbedder with configurable model and dimensions
- Support for text-embedding-3-small and text-embedding-3-large
- Custom base URL for Azure OpenAI compatibility

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: In-Memory Vector Store

**Files:**
- Create: `toolsemantic/stores/memory.go`
- Create: `toolsemantic/stores/memory_test.go`

**Step 1: Write failing tests**

```go
// stores/memory_test.go
package stores_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jrraymond/toolsemantic/stores"
)

func TestMemoryVectorStore_AddAndSearch(t *testing.T) {
    store := stores.NewMemoryVectorStore(3) // 3 dimensions
    ctx := context.Background()

    // Add vectors
    store.Add(ctx, "tool1", []float64{1.0, 0.0, 0.0})
    store.Add(ctx, "tool2", []float64{0.0, 1.0, 0.0})
    store.Add(ctx, "tool3", []float64{0.0, 0.0, 1.0})

    // Search for vector similar to tool1
    query := []float64{0.9, 0.1, 0.0}
    results, err := store.Search(ctx, query, 2)
    require.NoError(t, err)

    // tool1 should be most similar
    assert.Equal(t, "tool1", results[0].ID)
    assert.Greater(t, results[0].Score, results[1].Score)
}

func TestMemoryVectorStore_Delete(t *testing.T) {
    store := stores.NewMemoryVectorStore(3)
    ctx := context.Background()

    store.Add(ctx, "tool1", []float64{1.0, 0.0, 0.0})
    store.Delete(ctx, "tool1")

    results, err := store.Search(ctx, []float64{1.0, 0.0, 0.0}, 10)
    require.NoError(t, err)
    assert.Len(t, results, 0)
}

func TestMemoryVectorStore_Update(t *testing.T) {
    store := stores.NewMemoryVectorStore(3)
    ctx := context.Background()

    store.Add(ctx, "tool1", []float64{1.0, 0.0, 0.0})
    store.Add(ctx, "tool1", []float64{0.0, 1.0, 0.0}) // Update

    results, err := store.Search(ctx, []float64{0.0, 1.0, 0.0}, 1)
    require.NoError(t, err)
    assert.Equal(t, "tool1", results[0].ID)
    assert.InDelta(t, 1.0, results[0].Score, 0.01)
}

func TestMemoryVectorStore_EmptySearch(t *testing.T) {
    store := stores.NewMemoryVectorStore(3)
    ctx := context.Background()

    results, err := store.Search(ctx, []float64{1.0, 0.0, 0.0}, 10)
    require.NoError(t, err)
    assert.Len(t, results, 0)
}

func TestCosineSimilarity(t *testing.T) {
    tests := []struct {
        name     string
        a, b     []float64
        expected float64
    }{
        {
            name:     "identical vectors",
            a:        []float64{1.0, 0.0, 0.0},
            b:        []float64{1.0, 0.0, 0.0},
            expected: 1.0,
        },
        {
            name:     "orthogonal vectors",
            a:        []float64{1.0, 0.0, 0.0},
            b:        []float64{0.0, 1.0, 0.0},
            expected: 0.0,
        },
        {
            name:     "opposite vectors",
            a:        []float64{1.0, 0.0, 0.0},
            b:        []float64{-1.0, 0.0, 0.0},
            expected: -1.0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := stores.CosineSimilarity(tt.a, tt.b)
            assert.InDelta(t, tt.expected, result, 0.01)
        })
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `cd toolsemantic && go test ./stores/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
// stores/memory.go
package stores

import (
    "context"
    "math"
    "sort"
    "sync"

    "github.com/jrraymond/toolsemantic"
)

// MemoryVectorStore is an in-memory vector store
type MemoryVectorStore struct {
    dimensions int
    vectors    map[string][]float64
    mu         sync.RWMutex
}

// NewMemoryVectorStore creates a new in-memory vector store
func NewMemoryVectorStore(dimensions int) *MemoryVectorStore {
    return &MemoryVectorStore{
        dimensions: dimensions,
        vectors:    make(map[string][]float64),
    }
}

// Add adds or updates a vector
func (s *MemoryVectorStore) Add(ctx context.Context, id string, vector []float64) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Normalize the vector
    normalized := normalize(vector)
    s.vectors[id] = normalized
    return nil
}

// Search finds the most similar vectors
func (s *MemoryVectorStore) Search(ctx context.Context, query []float64, limit int) ([]toolsemantic.SearchResult, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    if len(s.vectors) == 0 {
        return []toolsemantic.SearchResult{}, nil
    }

    // Normalize query
    queryNorm := normalize(query)

    // Calculate similarities
    type scoredResult struct {
        id    string
        score float64
    }

    results := make([]scoredResult, 0, len(s.vectors))
    for id, vec := range s.vectors {
        score := CosineSimilarity(queryNorm, vec)
        results = append(results, scoredResult{id: id, score: score})
    }

    // Sort by score descending
    sort.Slice(results, func(i, j int) bool {
        return results[i].score > results[j].score
    })

    // Limit results
    if len(results) > limit {
        results = results[:limit]
    }

    // Convert to SearchResult
    searchResults := make([]toolsemantic.SearchResult, len(results))
    for i, r := range results {
        searchResults[i] = toolsemantic.SearchResult{
            ID:          r.id,
            Score:       r.score,
            VectorScore: r.score,
        }
    }

    return searchResults, nil
}

// Delete removes a vector
func (s *MemoryVectorStore) Delete(ctx context.Context, id string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    delete(s.vectors, id)
    return nil
}

// Close closes the store
func (s *MemoryVectorStore) Close() error {
    return nil
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) {
        return 0
    }

    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    if normA == 0 || normB == 0 {
        return 0
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// normalize normalizes a vector to unit length
func normalize(v []float64) []float64 {
    var sum float64
    for _, val := range v {
        sum += val * val
    }

    if sum == 0 {
        return v
    }

    norm := math.Sqrt(sum)
    result := make([]float64, len(v))
    for i, val := range v {
        result[i] = val / norm
    }

    return result
}
```

**Step 4: Run tests to verify they pass**

Run: `cd toolsemantic && go test ./stores/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add toolsemantic/
git commit -m "$(cat <<'EOF'
feat(toolsemantic): add in-memory vector store

- MemoryVectorStore with Add, Search, Delete
- Cosine similarity calculation
- Vector normalization for accurate similarity
- Thread-safe operations

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking PRD-013 complete:

- [ ] All tests pass: `go test ./... -v`
- [ ] Code coverage > 80%
- [ ] No linting errors: `golangci-lint run`
- [ ] Documentation complete
- [ ] Integration verified:
  - [ ] HybridSearcher combines BM25 + vector
  - [ ] OpenAI embeddings work
  - [ ] Memory vector store works

---

## Definition of Done

1. **HybridSearcher** combining BM25, vector, reranking
2. **Config** with weight validation
3. **Embedder** interface with OpenAI implementation
4. **VectorStore** interface with memory implementation
5. **Reranker** interface (RRF implementation)
6. All tests passing with >80% coverage
7. Documentation complete
