package main

import (
	"github.com/jonwraymond/toolcode"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolrun"
)

// maybeCreateExecutor returns nil unless the toolruntime build tag is enabled.
func maybeCreateExecutor(toolindex.Index, tooldocs.Store, toolrun.Runner) (toolcode.Executor, error) {
	return nil, nil
}
