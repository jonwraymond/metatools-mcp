package handlers

import (
	"context"
	"errors"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// CodeHandler handles the execute_code metatool
type CodeHandler struct {
	executor Executor
}

// NewCodeHandler creates a new code handler
func NewCodeHandler(executor Executor) *CodeHandler {
	return &CodeHandler{executor: executor}
}

// Handle executes the execute_code metatool
func (h *CodeHandler) Handle(ctx context.Context, input metatools.ExecuteCodeInput) (*metatools.ExecuteCodeOutput, error) {
	// Validate input
	if input.Language == "" {
		return nil, errors.New("language is required")
	}
	if input.Code == "" {
		return nil, errors.New("code is required")
	}

	// Build execution params
	params := ExecuteParams{
		Language: input.Language,
		Code:     input.Code,
	}

	if input.TimeoutMs != nil {
		params.TimeoutMs = *input.TimeoutMs
	}
	if input.MaxToolCalls != nil {
		params.MaxToolCalls = *input.MaxToolCalls
	}

	// Execute the code
	result, err := h.executor.ExecuteCode(ctx, params)
	if err != nil {
		return nil, err
	}

	return &metatools.ExecuteCodeOutput{
		Value:      result.Value,
		Stdout:     result.Stdout,
		Stderr:     result.Stderr,
		DurationMs: result.DurationMs,
	}, nil
}
