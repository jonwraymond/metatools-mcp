//go:build toolruntime

package main

import (
	"time"

	"github.com/jonwraymond/toolcode"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolrun"
	"github.com/jonwraymond/toolruntime"
	"github.com/jonwraymond/toolruntime/backend/unsafe"
	"github.com/jonwraymond/toolruntime/toolcodeengine"
)

// maybeCreateExecutor wires toolruntime into toolcode when the build tag is set.
func maybeCreateExecutor(idx toolindex.Index, docs tooldocs.Store, runner toolrun.Runner) (toolcode.Executor, error) {
	backend := unsafe.New(unsafe.Config{
		Mode:         unsafe.ModeSubprocess,
		RequireOptIn: false, // Unsafe dev mode is explicit via build tag.
	})

	rt := toolruntime.NewDefaultRuntime(toolruntime.RuntimeConfig{
		Backends: map[toolruntime.SecurityProfile]toolruntime.Backend{
			toolruntime.ProfileDev: backend,
		},
		DefaultProfile: toolruntime.ProfileDev,
	})

	engine := toolcodeengine.New(toolcodeengine.Config{
		Runtime: rt,
		Profile: toolruntime.ProfileDev,
	})

	return toolcode.NewDefaultExecutor(toolcode.Config{
		Index:          idx,
		Docs:           docs,
		Run:            runner,
		Engine:         engine,
		DefaultTimeout: 10 * time.Second,
		MaxToolCalls:   64,
		MaxChainSteps:  8,
	})
}
