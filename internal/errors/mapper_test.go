//revive:disable:var-naming // Package name matches internal/errors path for in-package tests.
package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapToolError_ToolNotFound(t *testing.T) {
	result := MapToolError(ErrToolNotFound, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeToolNotFound, result.Code)
	assert.Contains(t, result.Message, "not found")
}

func TestMapToolError_NoBackends(t *testing.T) {
	result := MapToolError(ErrNoBackends, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeNoBackends, result.Code)
	assert.Contains(t, result.Message, "backend")
}

func TestMapToolError_ValidationInput(t *testing.T) {
	result := MapToolError(ErrValidationInput, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeValidationInput, result.Code)
}

func TestMapToolError_ValidationOutput(t *testing.T) {
	result := MapToolError(ErrValidationOutput, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeValidationOutput, result.Code)
}

func TestMapToolError_StreamNotSupported(t *testing.T) {
	result := MapToolError(ErrStreamNotSupported, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeStreamNotSupported, result.Code)
}

func TestMapToolError_ExecutionFailed(t *testing.T) {
	result := MapToolError(ErrExecution, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeExecutionFailed, result.Code)
}

func TestMapToolError_Cancelled(t *testing.T) {
	result := MapToolError(context.Canceled, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeCancelled, result.Code)
}

func TestMapToolError_Timeout(t *testing.T) {
	result := MapToolError(context.DeadlineExceeded, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeTimeout, result.Code)
}

func TestMapToolError_ChainStepFailed(t *testing.T) {
	// Chain step errors should have step_index set
	result := MapToolError(ErrExecution, "test.tool", nil, 2)

	require.NotNil(t, result)
	assert.Equal(t, CodeChainStepFailed, result.Code)
	assert.NotNil(t, result.StepIndex)
	assert.Equal(t, 2, *result.StepIndex)
}

func TestMapToolError_Internal(t *testing.T) {
	unknownErr := errors.New("some unknown error")
	result := MapToolError(unknownErr, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, CodeInternal, result.Code)
}

func TestMapToolError_SetsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		stepIndex int
		retryable bool
	}{
		{"execution_failed is retryable", ErrExecution, -1, true},
		{"internal is retryable", errors.New("unknown"), -1, true},
		{"cancelled is not retryable", context.Canceled, -1, false},
		{"timeout is not retryable", context.DeadlineExceeded, -1, false},
		{"tool_not_found is not retryable", ErrToolNotFound, -1, false},
		{"no_backends is not retryable", ErrNoBackends, -1, false},
		{"validation_input is not retryable", ErrValidationInput, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToolError(tt.err, "test.tool", nil, tt.stepIndex)
			assert.Equal(t, tt.retryable, result.Retryable)
		})
	}
}

func TestMapToolError_SetsBackendKind(t *testing.T) {
	backend := &BackendInfo{Kind: "mcp"}
	result := MapToolError(ErrExecution, "test.tool", backend, -1)

	require.NotNil(t, result)
	assert.NotNil(t, result.BackendKind)
	assert.Equal(t, "mcp", *result.BackendKind)
}

func TestMapToolError_SetsToolID(t *testing.T) {
	result := MapToolError(ErrExecution, "namespace.mytool", nil, -1)

	require.NotNil(t, result)
	assert.Equal(t, "namespace.mytool", result.ToolID)
}

func TestMapToolError_SetsOp(t *testing.T) {
	// ToolError with Op field
	toolErr := &ToolError{
		Err: ErrExecution,
		Op:  "execute",
	}
	result := MapToolError(toolErr, "test.tool", nil, -1)

	require.NotNil(t, result)
	assert.NotNil(t, result.Op)
	assert.Equal(t, "execute", *result.Op)
}
