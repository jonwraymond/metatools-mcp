package errors

import (
	"context"
	"errors"

	"github.com/jonwraymond/toolexec/run"
)

// ErrorCode represents the standard error codes for metatools
type ErrorCode string

const (
	CodeToolNotFound           ErrorCode = "tool_not_found"
	CodeNoBackends             ErrorCode = "no_backends"
	CodeBackendOverrideInvalid ErrorCode = "backend_override_invalid"
	CodeBackendOverrideNoMatch ErrorCode = "backend_override_no_match"
	CodeValidationInput        ErrorCode = "validation_input"
	CodeValidationOutput       ErrorCode = "validation_output"
	CodeExecutionFailed        ErrorCode = "execution_failed"
	CodeStreamNotSupported     ErrorCode = "stream_not_supported"
	CodeStreamFailed           ErrorCode = "stream_failed"
	CodeChainStepFailed        ErrorCode = "chain_step_failed"
	CodeCancelled              ErrorCode = "cancelled"
	CodeTimeout                ErrorCode = "timeout"
	CodeInternal               ErrorCode = "internal"
)

// Sentinel errors for mapping
var (
	ErrToolNotFound           = errors.New("tool not found")
	ErrNoBackends             = errors.New("no backends available")
	ErrBackendOverrideInvalid = errors.New("backend override invalid")
	ErrBackendOverrideNoMatch = errors.New("backend override no match")
	ErrValidationInput        = errors.New("input validation failed")
	ErrValidationOutput       = errors.New("output validation failed")
	ErrStreamNotSupported     = errors.New("streaming not supported")
	ErrExecution              = errors.New("execution failed")
)

// BackendInfo represents backend information for error context
type BackendInfo struct {
	Kind string
}

// ToolError represents an error with operation context
type ToolError struct {
	Err error
	Op  string
}

func (e *ToolError) Error() string {
	return e.Err.Error()
}

func (e *ToolError) Unwrap() error {
	return e.Err
}

// ErrorObject is the structured error returned in metatool responses
type ErrorObject struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	ToolID      string                 `json:"tool_id,omitempty"`
	Op          *string                `json:"op,omitempty"`
	BackendKind *string                `json:"backend_kind,omitempty"`
	StepIndex   *int                   `json:"step_index,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// MapToolError maps an error to a structured ErrorObject
// stepIndex should be -1 if not in a chain context
func MapToolError(err error, toolID string, backend *BackendInfo, stepIndex int) *ErrorObject {
	result := &ErrorObject{
		ToolID:  toolID,
		Message: err.Error(),
	}

	// Extract run.ToolError context when present.
	var trErr *run.ToolError
	if errors.As(err, &trErr) {
		op := trErr.Op
		result.Op = &op
		if trErr.Backend != nil {
			kind := string(trErr.Backend.Kind)
			result.BackendKind = &kind
		}
		err = trErr.Unwrap()
	}

	// Check for ToolError to extract Op
	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		op := toolErr.Op
		result.Op = &op
		err = toolErr.Unwrap()
	}

	// Set backend kind if provided
	if backend != nil {
		result.BackendKind = &backend.Kind
	}

	// If in chain context with step index, it's a chain step failure
	if stepIndex >= 0 {
		result.Code = CodeChainStepFailed
		result.StepIndex = &stepIndex
		result.Retryable = isRetryable(CodeChainStepFailed)
		return result
	}

	// Map error to code
	result.Code = mapErrorToCode(err)
	result.Retryable = isRetryable(result.Code)

	return result
}

func mapErrorToCode(err error) ErrorCode {
	switch {
	case errors.Is(err, ErrToolNotFound) || errors.Is(err, run.ErrToolNotFound):
		return CodeToolNotFound
	case errors.Is(err, ErrNoBackends) || errors.Is(err, run.ErrNoBackends):
		return CodeNoBackends
	case errors.Is(err, ErrBackendOverrideInvalid):
		return CodeBackendOverrideInvalid
	case errors.Is(err, ErrBackendOverrideNoMatch):
		return CodeBackendOverrideNoMatch
	case errors.Is(err, ErrValidationInput) || errors.Is(err, run.ErrValidation):
		return CodeValidationInput
	case errors.Is(err, ErrValidationOutput) || errors.Is(err, run.ErrOutputValidation):
		return CodeValidationOutput
	case errors.Is(err, ErrStreamNotSupported) || errors.Is(err, run.ErrStreamNotSupported):
		return CodeStreamNotSupported
	case errors.Is(err, ErrExecution) || errors.Is(err, run.ErrExecution):
		return CodeExecutionFailed
	case errors.Is(err, context.Canceled):
		return CodeCancelled
	case errors.Is(err, context.DeadlineExceeded):
		return CodeTimeout
	default:
		return CodeInternal
	}
}

func isRetryable(code ErrorCode) bool {
	switch code {
	case CodeExecutionFailed, CodeStreamFailed, CodeInternal:
		return true
	default:
		return false
	}
}
