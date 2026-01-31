//go:build toolruntime

package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/runtime/docker"
	"github.com/jonwraymond/metatools-mcp/internal/runtime/wasm"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
	"github.com/jonwraymond/toolexec/runtime"
	dockerbackend "github.com/jonwraymond/toolexec/runtime/backend/docker"
	"github.com/jonwraymond/toolexec/runtime/backend/unsafe"
	wasmbackend "github.com/jonwraymond/toolexec/runtime/backend/wasm"
	"github.com/jonwraymond/toolexec/runtime/toolcodeengine"
)

// maybeCreateExecutor wires toolruntime into toolcode when the build tag is set.
// It configures both unsafe (dev) and Docker (standard) backends based on availability.
func maybeCreateExecutor(idx index.Index, docs tooldoc.Store, runner run.Runner) (code.Executor, error) {
	// Unsafe backend for dev profile - always available
	unsafeBackend := unsafe.New(unsafe.Config{
		Mode:         unsafe.ModeSubprocess,
		RequireOptIn: false, // Unsafe dev mode is explicit via build tag.
	})

	backends := map[runtime.SecurityProfile]runtime.Backend{
		runtime.ProfileDev: unsafeBackend,
	}

	defaultProfile := runtime.ProfileDev

	// Try to create Docker backend for standard profile
	var dockerBack runtime.Backend
	dockerClient, err := docker.NewClient(docker.ClientConfig{})
	if err != nil {
		slog.Warn("Docker client unavailable", "error", err)
	} else {
		// Health check before proceeding
		healthChecker := docker.NewHealthCheck(dockerClient.Docker())
		if err := healthChecker.Ping(context.Background()); err != nil {
			slog.Warn("Docker daemon not responding", "error", err)
			_ = dockerClient.Close()
		} else {
			// Create full Docker backend with all interfaces
			imageResolver := docker.NewResolver(dockerClient.Docker())

			dockerBack = dockerbackend.New(dockerbackend.Config{
				ImageName:     getEnvOrDefault("METATOOLS_DOCKER_IMAGE", "toolruntime-sandbox:latest"),
				Client:        dockerClient,
				ImageResolver: imageResolver,
				HealthChecker: healthChecker,
				Logger:        &slogAdapter{},
			})

			slog.Info("Docker backend available",
				"image", getEnvOrDefault("METATOOLS_DOCKER_IMAGE", "toolruntime-sandbox:latest"))
		}
	}

	// Try to create WASM backend for edge/lightweight profile
	// WASM provides strong isolation without Docker dependencies
	var wasmBack runtime.Backend
	if os.Getenv("METATOOLS_WASM_ENABLED") == "true" {
		wasmClient, err := wasm.NewClient(wasm.ClientConfig{
			MaxMemoryPages:         256, // 16MB default
			EnableCompilationCache: true,
		})
		if err != nil {
			slog.Warn("WASM client unavailable", "error", err)
		} else {
			healthChecker := wasm.NewHealthCheck(wasmClient)
			if err := healthChecker.Ping(context.Background()); err != nil {
				slog.Warn("WASM runtime not responding", "error", err)
				_ = wasmClient.Close(context.Background())
			} else {
				wasmBack = wasmbackend.New(wasmbackend.Config{
					Runtime:        "wazero",
					MaxMemoryPages: 256,
					EnableWASI:     true,
					Client:         wasmClient,
					HealthChecker:  healthChecker,
					Logger:         &slogAdapter{},
				})

				slog.Info("WASM backend available",
					"runtime", "wazero",
					"memoryPages", 256)
			}
		}
	}

	// Select standard backend preference (docker default, wasm optional)
	standardBackend := ""
	preferred := os.Getenv("METATOOLS_RUNTIME_BACKEND")
	switch preferred {
	case "wasm":
		if wasmBack != nil {
			backends[runtime.ProfileStandard] = wasmBack
			standardBackend = "wasm"
		} else {
			slog.Warn("WASM backend requested but unavailable")
		}
	case "docker":
		if dockerBack != nil {
			backends[runtime.ProfileStandard] = dockerBack
			standardBackend = "docker"
		} else {
			slog.Warn("Docker backend requested but unavailable")
		}
	default:
		if dockerBack != nil {
			backends[runtime.ProfileStandard] = dockerBack
			standardBackend = "docker"
		} else if wasmBack != nil {
			backends[runtime.ProfileStandard] = wasmBack
			standardBackend = "wasm"
			slog.Info("Docker unavailable, using WASM for standard profile")
		}
	}

	if standardBackend != "" {
		slog.Info("Standard backend selected", "backend", standardBackend)
	}

	// Honor requested runtime profile
	if os.Getenv("METATOOLS_RUNTIME_PROFILE") == "standard" {
		if standardBackend != "" {
			defaultProfile = runtime.ProfileStandard
		} else {
			slog.Warn("Standard profile requested but no standard backend available")
		}
	}

	rt := runtime.NewDefaultRuntime(runtime.RuntimeConfig{
		Backends:       backends,
		DefaultProfile: defaultProfile,
	})

	engine, err := toolcodeengine.New(toolcodeengine.Config{
		Runtime: rt,
		Profile: defaultProfile,
	})
	if err != nil {
		return nil, err
	}

	return code.NewDefaultExecutor(code.Config{
		Index:          idx,
		Docs:           docs,
		Run:            runner,
		Engine:         engine,
		DefaultTimeout: 10 * time.Second,
		MaxToolCalls:   64,
		MaxChainSteps:  8,
	})
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// slogAdapter adapts slog to toolruntime's Logger interface.
type slogAdapter struct{}

func (s *slogAdapter) Info(msg string, args ...any)  { slog.Info(msg, args...) }
func (s *slogAdapter) Warn(msg string, args ...any)  { slog.Warn(msg, args...) }
func (s *slogAdapter) Error(msg string, args ...any) { slog.Error(msg, args...) }
