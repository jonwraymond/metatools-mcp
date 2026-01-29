package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func decodeArgs(args map[string]any, out any) error {
	if args == nil {
		return nil
	}
	data, err := json.Marshal(args)
	if err != nil {
		return &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: fmt.Sprintf("invalid arguments: %v", err)}
	}
	if err := json.Unmarshal(data, out); err != nil {
		return &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: fmt.Sprintf("invalid arguments: %v", err)}
	}
	return nil
}

func progressNotifier(ctx context.Context, req *mcp.CallToolRequest) func(handlers.ProgressEvent) {
	if req == nil || req.Session == nil || req.Params == nil {
		return nil
	}
	token := req.Params.GetProgressToken()
	if token == nil {
		return nil
	}

	return func(ev handlers.ProgressEvent) {
		params := &mcp.ProgressNotificationParams{
			ProgressToken: token,
			Progress:      ev.Progress,
			Total:         ev.Total,
			Message:       ev.Message,
		}
		_ = req.Session.NotifyProgress(ctx, params)
	}
}
