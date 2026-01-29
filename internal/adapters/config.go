package adapters

import (
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolcode"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolrun"
)

// NewConfig adapts the core tool libraries into a metatools server config.
func NewConfig(idx toolindex.Index, docs tooldocs.Store, runner toolrun.Runner, exec toolcode.Executor) config.Config {
	cfg := config.Config{
		Index:                           NewIndexAdapter(idx),
		Docs:                            NewDocsAdapter(docs),
		Runner:                          NewRunnerAdapter(runner),
		Providers:                       config.DefaultAppConfig().Providers,
		NotifyToolListChanged:           true,
		NotifyToolListChangedDebounceMs: int((150 * time.Millisecond) / time.Millisecond),
	}
	if exec != nil {
		cfg.Executor = NewExecutorAdapter(exec)
	}
	return cfg
}
