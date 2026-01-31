//go:build !toolruntime

package main

import (
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
)

// maybeCreateExecutor returns nil unless the toolruntime build tag is enabled.
func maybeCreateExecutor(index.Index, tooldoc.Store, run.Runner) (code.Executor, error) {
	return nil, nil
}
