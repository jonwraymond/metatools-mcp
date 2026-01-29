package config

import (
	"errors"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
)

// Config holds the server configuration with injected dependencies
type Config struct {
	Index    handlers.Index
	Docs     handlers.Store
	Runner   handlers.Runner
	Executor handlers.Executor // optional

	NotifyToolListChanged           bool
	NotifyToolListChangedDebounceMs int
}

// Validate checks that required dependencies are provided
func (c *Config) Validate() error {
	if c.Index == nil {
		return errors.New("index is required")
	}
	if c.Docs == nil {
		return errors.New("docs is required")
	}
	if c.Runner == nil {
		return errors.New("runner is required")
	}
	// Executor is optional
	return nil
}
