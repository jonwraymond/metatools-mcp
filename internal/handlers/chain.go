package handlers

import (
	"context"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// ChainHandler handles the run_chain metatool
type ChainHandler struct {
	runner Runner
}

// NewChainHandler creates a new chain handler
func NewChainHandler(runner Runner) *ChainHandler {
	return &ChainHandler{runner: runner}
}

// Handle executes the run_chain metatool
// Returns (result, isError, err) where:
// - result is the chain output with results
// - isError indicates if this is a tool error (isError: true in MCP)
// - err is a protocol-level error
func (h *ChainHandler) Handle(ctx context.Context, input metatools.RunChainInput) (*metatools.RunChainOutput, bool, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, false, err
	}

	includeBackends := input.GetIncludeBackends()
	includeTools := input.GetIncludeTools()

	// Convert input steps to handler steps
	steps := make([]ChainStep, len(input.Steps))
	for i, s := range input.Steps {
		steps[i] = ChainStep{
			ToolID:      s.ToolID,
			Args:        s.Args,
			UsePrevious: s.UsePrevious,
		}
	}

	// Execute the chain
	finalResult, stepResults, chainErr := h.runner.RunChain(ctx, steps)

	// Build step results output
	results := make([]metatools.ChainStepResult, len(stepResults))
	failingStepIndex := -1

	for i, sr := range stepResults {
		results[i] = metatools.ChainStepResult{
			ToolID:     sr.ToolID,
			Structured: sr.Structured,
		}

		// Include optional fields based on input flags
		if includeBackends && sr.Backend != nil {
			results[i].Backend = sr.Backend
		}
		if includeTools && sr.Tool != nil {
			results[i].Tool = sr.Tool
		}

		// Track failing step
		if sr.Error != nil {
			failingStepIndex = i
			errObj := merrors.MapToolError(sr.Error, sr.ToolID, nil, -1)
			results[i].Error = &metatools.ErrorObject{
				Code:      string(errObj.Code),
				Message:   errObj.Message,
				ToolID:    sr.ToolID,
				Retryable: errObj.Retryable,
			}
			if errObj.Op != nil {
				results[i].Error.Op = errObj.Op
			}
			if errObj.BackendKind != nil {
				results[i].Error.BackendKind = errObj.BackendKind
			}
		}
	}

	output := &metatools.RunChainOutput{
		Results: results,
	}

	// Handle chain error
	if chainErr != nil {
		// Map the chain error
		errObj := merrors.MapToolError(chainErr, "", nil, failingStepIndex)

		// Get the underlying cause code
		causeErrObj := merrors.MapToolError(chainErr, "", nil, -1)

		output.Error = &metatools.ErrorObject{
			Code:      string(errObj.Code),
			Message:   errObj.Message,
			StepIndex: errObj.StepIndex,
			Retryable: errObj.Retryable,
			Details: map[string]any{
				"cause_code": string(causeErrObj.Code),
			},
		}
		if causeErrObj.Op != nil {
			output.Error.Details["cause_op"] = *causeErrObj.Op
		}
		if causeErrObj.BackendKind != nil {
			output.Error.Details["cause_backend_kind"] = *causeErrObj.BackendKind
		}

		// Set final to last successful structured value if any
		for i := len(stepResults) - 1; i >= 0; i-- {
			if stepResults[i].Error == nil && stepResults[i].Structured != nil {
				output.Final = stepResults[i].Structured
				break
			}
		}

		return output, true, nil
	}

	// Success - set final from the chain result
	output.Final = finalResult.Structured

	return output, false, nil
}
