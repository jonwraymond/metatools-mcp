package adapters

import (
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
)

// NewConfig adapts the core tool libraries into a metatools server config.
func NewConfig(idx index.Index, docs tooldoc.Store, runner run.Runner, exec code.Executor) config.Config {
	cfg := config.Config{
		Index:                           NewIndexAdapter(idx),
		Docs:                            NewDocsAdapter(docs),
		Runner:                          NewRunnerAdapter(runner),
		Providers:                       config.DefaultAppConfig().Providers,
		Middleware:                      config.DefaultAppConfig().Middleware,
		NotifyToolListChanged:           true,
		NotifyToolListChangedDebounceMs: int((150 * time.Millisecond) / time.Millisecond),
	}
	if exec != nil {
		cfg.Executor = NewExecutorAdapter(exec)
	}
	return cfg
}
