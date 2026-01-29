package loader

import (
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoadFromConfig_LocalEnabled(t *testing.T) {
	cfg := config.DefaultAppConfig().Backends
	cfg.Local.Enabled = true

	registry, err := LoadFromConfig(cfg)
	require.NoError(t, err)

	_, ok := registry.Get("local")
	require.True(t, ok)
}

func TestLoadFromConfig_LocalDisabled(t *testing.T) {
	cfg := config.DefaultAppConfig().Backends
	cfg.Local.Enabled = false

	registry, err := LoadFromConfig(cfg)
	require.NoError(t, err)

	_, ok := registry.Get("local")
	require.False(t, ok)
}
