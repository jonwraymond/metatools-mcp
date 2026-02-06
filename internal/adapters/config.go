package adapters

import (
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
)

// NewConfig adapts the core tool libraries into a metatools server config.
func NewConfig(idx index.Index, docs tooldoc.Store, runner run.Runner, exec code.Executor) config.Config {
	defaults := config.DefaultAppConfig().SkillDefaults
	cfg := config.Config{
		Index:    NewIndexAdapter(idx),
		Docs:     NewDocsAdapter(docs),
		Runner:   NewRunnerAdapter(runner),
		Toolsets: toolset.NewRegistry(nil),
		Skills:   skills.NewRegistry(nil),
		SkillDefaults: handlers.SkillDefaults{
			MaxSteps:     defaults.MaxSteps,
			MaxToolCalls: defaults.MaxToolCalls,
			Timeout:      defaults.Timeout,
		},
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
