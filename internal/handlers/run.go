package handlers

import (
	"context"
	"fmt"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// RunHandler handles the run_tool metatool
type RunHandler struct {
	runner Runner
}

// NewRunHandler creates a new run handler
func NewRunHandler(runner Runner) *RunHandler {
	return &RunHandler{runner: runner}
}

// Handle executes the run_tool metatool
// Returns (result, isError, err) where:
// - result is the structured output
// - isError indicates if this is a tool error (isError: true in MCP)
// - err is a protocol-level error
func (h *RunHandler) Handle(ctx context.Context, input metatools.RunToolInput) (*metatools.RunToolOutput, bool, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, false, err
	}

	buildToolError := func(err error) (*metatools.RunToolOutput, bool, error) {
		errObj := merrors.MapToolError(err, input.ToolID, nil, -1)
		output := &metatools.RunToolOutput{
			Error: &metatools.ErrorObject{
				Code:      string(errObj.Code),
				Message:   errObj.Message,
				ToolID:    errObj.ToolID,
				Retryable: errObj.Retryable,
			},
		}
		if errObj.Op != nil {
			output.Error.Op = errObj.Op
		}
		if errObj.BackendKind != nil {
			output.Error.BackendKind = errObj.BackendKind
		}
		return output, true, nil
	}

	// Streaming is not supported by this handler interface.
	if input.Stream {
		return buildToolError(merrors.ErrStreamNotSupported)
	}

	// Per-call backend override is not supported by the current runner interface.
	if input.BackendOverride != nil {
		err := fmt.Errorf("%w: backend_override is not supported by this server", merrors.ErrBackendOverrideInvalid)
		return buildToolError(err)
	}

	// Execute the tool
	result, err := h.runner.Run(ctx, input.ToolID, input.Args)
	if err != nil {
		return buildToolError(err)
	}

	// Build successful output
	output := &metatools.RunToolOutput{
		Structured: result.Structured,
	}

	if result.DurationMs > 0 {
		output.DurationMs = &result.DurationMs
	}

	// Include optional fields based on input flags
	if input.IncludeTool && result.Tool != nil {
		output.Tool = result.Tool
	}
	if input.IncludeBackend && result.Backend != nil {
		output.Backend = result.Backend
	}
	if input.IncludeMCPResult && result.MCPResult != nil {
		output.MCPResult = result.MCPResult
	}

	return output, false, nil
}
