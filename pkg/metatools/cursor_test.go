package metatools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCursor(t *testing.T) {
	cursor := EncodeCursor(42)
	assert.NotEmpty(t, cursor)
	// Should be base64 encoded
	assert.NotContains(t, cursor, " ")
}

func TestDecodeCursor(t *testing.T) {
	// Encode then decode
	original := 42
	cursor := EncodeCursor(original)

	decoded, err := DecodeCursor(cursor)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDecodeCursor_Empty(t *testing.T) {
	decoded, err := DecodeCursor("")
	require.NoError(t, err)
	assert.Equal(t, 0, decoded) // empty cursor means start from beginning
}

func TestDecodeCursor_Invalid(t *testing.T) {
	_, err := DecodeCursor("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestDecodeCursor_InvalidContent(t *testing.T) {
	// Valid base64 but not a number
	_, err := DecodeCursor("aGVsbG8=") // "hello" in base64
	assert.Error(t, err)
}

func TestApplyCursor(t *testing.T) {
	items := []ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
	}

	// Start from offset 2
	cursor := EncodeCursor(2)
	result, nextCursor := ApplyCursor(items, cursor, 10)

	assert.Len(t, result, 3)
	assert.Equal(t, "c", result[0].ID)
	assert.Nil(t, nextCursor) // no more items
}

func TestApplyCursor_WithLimit(t *testing.T) {
	items := []ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
	}

	// Start from 0, limit 2
	result, nextCursor := ApplyCursor(items, "", 2)

	assert.Len(t, result, 2)
	assert.Equal(t, "a", result[0].ID)
	assert.Equal(t, "b", result[1].ID)
	require.NotNil(t, nextCursor)

	// Decode next cursor and verify it points to offset 2
	offset, err := DecodeCursor(*nextCursor)
	require.NoError(t, err)
	assert.Equal(t, 2, offset)
}

func TestApplyCursor_EmptyCursor(t *testing.T) {
	items := []ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}

	result, nextCursor := ApplyCursor(items, "", 10)

	assert.Len(t, result, 3)
	assert.Nil(t, nextCursor)
}

func TestApplyCursor_OffsetBeyondLength(t *testing.T) {
	items := []ToolSummary{
		{ID: "a"}, {ID: "b"},
	}

	cursor := EncodeCursor(100)
	result, nextCursor := ApplyCursor(items, cursor, 10)

	assert.Empty(t, result)
	assert.Nil(t, nextCursor)
}

func TestApplyCursor_ExactBoundary(t *testing.T) {
	items := []ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}

	// Request limit of 3, which is exactly the size
	result, nextCursor := ApplyCursor(items, "", 3)

	assert.Len(t, result, 3)
	assert.Nil(t, nextCursor) // no more items
}
