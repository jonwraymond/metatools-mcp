# Usage

## Build and run (stdio)

```bash
go run ./cmd/metatools
```

## Enable BM25 search (build tag + env)

```bash
go build -tags toolsearch ./cmd/metatools
METATOOLS_SEARCH_STRATEGY=bm25 ./metatools
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_SEARCH_STRATEGY` | `lexical` | `lexical` or `bm25` |
| `METATOOLS_SEARCH_BM25_NAME_BOOST` | `3` | BM25 name field boost |
| `METATOOLS_SEARCH_BM25_NAMESPACE_BOOST` | `2` | BM25 namespace field boost |
| `METATOOLS_SEARCH_BM25_TAGS_BOOST` | `2` | BM25 tags field boost |
| `METATOOLS_SEARCH_BM25_MAX_DOCS` | `0` | Max docs to index (0=unlimited) |
| `METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN` | `0` | Max doc text length (0=unlimited) |

## Optional toolruntime support

```bash
go run -tags toolruntime ./cmd/metatools
```

This enables `execute_code` backed by a `toolruntime` engine.
