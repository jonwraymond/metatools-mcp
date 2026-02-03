package cmd

import (
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
)

// maybeCreateExecutor returns nil unless the toolruntime build tag is enabled.
func maybeCreateExecutor(_ config.ExecutionConfig, _ index.Index, _ tooldoc.Store, _ run.Runner) (code.Executor, error) {
	return nil, nil
}
