package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// Load reads configuration from defaults, optional file, and environment variables.
// Precedence: defaults < file < env.
func Load(path string) (AppConfig, error) {
	return loadConfig(path, nil)
}

// LoadWithOverrides reads configuration and applies explicit overrides (highest precedence).
func LoadWithOverrides(path string, overrides map[string]any) (AppConfig, error) {
	return loadConfig(path, overrides)
}

func loadConfig(path string, overrides map[string]any) (AppConfig, error) {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(DefaultAppConfig(), "koanf"), nil); err != nil {
		return AppConfig{}, fmt.Errorf("load defaults: %w", err)
	}

	if path != "" {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return AppConfig{}, fmt.Errorf("load file %q: %w", path, err)
		}
	}

	envProvider := env.Provider("METATOOLS_", ".", func(s string) string {
		trimmed := strings.TrimPrefix(s, "METATOOLS_")
		return strings.ToLower(strings.ReplaceAll(trimmed, "_", "."))
	})
	if err := k.Load(envProvider, nil); err != nil {
		return AppConfig{}, fmt.Errorf("load env: %w", err)
	}

	for key, value := range overrides {
		if err := k.Set(key, value); err != nil {
			return AppConfig{}, fmt.Errorf("apply override %q: %w", key, err)
		}
	}

	var cfg AppConfig
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		Tag: "koanf",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
			),
			WeaklyTypedInput: true,
			Result:           &cfg,
		},
	}); err != nil {
		return AppConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}
