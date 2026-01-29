package metatools

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

// EncodeCursor encodes an offset into an opaque cursor string
func EncodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// DecodeCursor decodes a cursor string back to an offset
// Returns 0 for empty cursor (start from beginning)
func DecodeCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor: %w", err)
	}

	offset, err := strconv.Atoi(string(decoded))
	if err != nil {
		return 0, fmt.Errorf("invalid cursor content: %w", err)
	}

	return offset, nil
}

// ApplyCursor applies cursor pagination to a slice of ToolSummary
// Returns the paginated slice and an optional nextCursor
func ApplyCursor(items []ToolSummary, cursor string, limit int) ([]ToolSummary, *string) {
	offset, err := DecodeCursor(cursor)
	if err != nil {
		offset = 0
	}

	// Handle offset beyond length
	if offset >= len(items) {
		return []ToolSummary{}, nil
	}

	// Apply offset
	items = items[offset:]

	// Apply limit
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	// Generate next cursor if there are more items
	var nextCursor *string
	if hasMore {
		next := EncodeCursor(offset + limit)
		nextCursor = &next
	}

	return items, nextCursor
}

// NullableCursor returns a pointer to cursor when non-empty.
func NullableCursor(cursor string) *string {
	if cursor == "" {
		return nil
	}
	return &cursor
}
