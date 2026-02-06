package config

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

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
		// #nosec G304 -- config path is explicitly user-controlled (CLI/env) and is intended to be read.
		b, err := os.ReadFile(path)
		if err != nil {
			return AppConfig{}, fmt.Errorf("read file %q: %w", path, err)
		}

		expanded, err := expandEnv(b)
		if err != nil {
			return AppConfig{}, fmt.Errorf("expand env in file %q: %w", path, err)
		}

		if err := k.Load(rawbytes.Provider(expanded), yaml.Parser()); err != nil {
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

func expandEnv(b []byte) ([]byte, error) {
	// Allow writing literal "$" without triggering substitution.
	const dollarSentinel = "\x00METATOOLS_DOLLAR\x00"

	s := string(b)
	s = strings.ReplaceAll(s, "$$", dollarSentinel)

	missing := make(map[string]struct{})
	for _, match := range envVarPattern.FindAllStringSubmatch(s, -1) {
		key := match[1]
		if _, ok := os.LookupEnv(key); !ok {
			missing[key] = struct{}{}
		}
	}
	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for key := range missing {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(keys, ", "))
	}

	s = os.ExpandEnv(s)
	s = strings.ReplaceAll(s, dollarSentinel, "$")
	return []byte(s), nil
}
