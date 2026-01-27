package metatools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolSummary_JSON(t *testing.T) {
	summary := ToolSummary{
		ID:               "namespace.tool",
		Name:             "tool",
		Namespace:        "namespace",
		ShortDescription: "A test tool",
		Tags:             []string{"test", "example"},
	}

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	var decoded ToolSummary
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, summary.ID, decoded.ID)
	assert.Equal(t, summary.Name, decoded.Name)
	assert.Equal(t, summary.Namespace, decoded.Namespace)
	assert.Equal(t, summary.ShortDescription, decoded.ShortDescription)
	assert.Equal(t, summary.Tags, decoded.Tags)
}

func TestToolSummary_JSON_OmitsEmptyFields(t *testing.T) {
	summary := ToolSummary{
		ID:   "test.tool",
		Name: "tool",
	}

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	// Should not contain namespace or tags when empty
	assert.NotContains(t, string(data), "namespace")
	assert.NotContains(t, string(data), "tags")
}

func TestRunToolInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   RunToolInput
		wantErr bool
	}{
		{
			name:    "valid with tool_id only",
			input:   RunToolInput{ToolID: "test.tool"},
			wantErr: false,
		},
		{
			name:    "missing tool_id",
			input:   RunToolInput{},
			wantErr: true,
		},
		{
			name: "valid with all fields",
			input: RunToolInput{
				ToolID:           "test.tool",
				Args:             map[string]any{"key": "value"},
				Stream:           true,
				IncludeTool:      true,
				IncludeBackend:   true,
				IncludeMCPResult: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunToolOutput_WithError(t *testing.T) {
	output := RunToolOutput{
		Error: &ErrorObject{
			Code:    "tool_not_found",
			Message: "Tool not found",
			ToolID:  "test.tool",
		},
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded RunToolOutput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	require.NotNil(t, decoded.Error)
	assert.Equal(t, "tool_not_found", decoded.Error.Code)
	assert.Nil(t, decoded.Structured)
}

func TestRunToolOutput_WithStructured(t *testing.T) {
	output := RunToolOutput{
		Structured: map[string]any{"result": "success"},
		DurationMs: intPtr(150),
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded RunToolOutput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	require.NotNil(t, decoded.Structured)
	structuredMap, ok := decoded.Structured.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "success", structuredMap["result"])
	assert.Nil(t, decoded.Error)
	require.NotNil(t, decoded.DurationMs)
	assert.Equal(t, 150, *decoded.DurationMs)
}

func TestSearchToolsInput_Defaults(t *testing.T) {
	input := SearchToolsInput{Query: "test"}

	// Default limit should be applied
	limit := input.GetLimit()
	assert.Equal(t, 20, limit) // default
}

func TestSearchToolsInput_LimitCapped(t *testing.T) {
	input := SearchToolsInput{Query: "test", Limit: intPtr(200)}

	limit := input.GetLimit()
	assert.Equal(t, 100, limit) // max cap
}

func TestDescribeToolInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   DescribeToolInput
		wantErr bool
	}{
		{
			name:    "valid summary",
			input:   DescribeToolInput{ToolID: "test.tool", DetailLevel: "summary"},
			wantErr: false,
		},
		{
			name:    "valid schema",
			input:   DescribeToolInput{ToolID: "test.tool", DetailLevel: "schema"},
			wantErr: false,
		},
		{
			name:    "valid full",
			input:   DescribeToolInput{ToolID: "test.tool", DetailLevel: "full"},
			wantErr: false,
		},
		{
			name:    "missing tool_id",
			input:   DescribeToolInput{DetailLevel: "summary"},
			wantErr: true,
		},
		{
			name:    "invalid detail_level",
			input:   DescribeToolInput{ToolID: "test.tool", DetailLevel: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunChainInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   RunChainInput
		wantErr bool
	}{
		{
			name: "valid with one step",
			input: RunChainInput{
				Steps: []ChainStep{{ToolID: "test.tool"}},
			},
			wantErr: false,
		},
		{
			name:    "empty steps",
			input:   RunChainInput{Steps: []ChainStep{}},
			wantErr: true,
		},
		{
			name: "step missing tool_id",
			input: RunChainInput{
				Steps: []ChainStep{{Args: map[string]any{"key": "value"}}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
