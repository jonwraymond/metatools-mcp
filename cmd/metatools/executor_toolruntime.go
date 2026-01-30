//go:build toolruntime

package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/runtime/docker"
	"github.com/jonwraymond/toolcode"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolrun"
	"github.com/jonwraymond/toolruntime"
	dockerbackend "github.com/jonwraymond/toolruntime/backend/docker"
	"github.com/jonwraymond/toolruntime/backend/unsafe"
	"github.com/jonwraymond/toolruntime/toolcodeengine"
)

// maybeCreateExecutor wires toolruntime into toolcode when the build tag is set.
// It configures both unsafe (dev) and Docker (standard) backends based on availability.
func maybeCreateExecutor(idx toolindex.Index, docs tooldocs.Store, runner toolrun.Runner) (toolcode.Executor, error) {
	// Unsafe backend for dev profile - always available
	unsafeBackend := unsafe.New(unsafe.Config{
		Mode:         unsafe.ModeSubprocess,
		RequireOptIn: false, // Unsafe dev mode is explicit via build tag.
	})

	backends := map[toolruntime.SecurityProfile]toolruntime.Backend{
		toolruntime.ProfileDev: unsafeBackend,
	}

	defaultProfile := toolruntime.ProfileDev

	// Try to create Docker backend for standard profile
	dockerClient, err := docker.NewClient(docker.ClientConfig{})
	if err != nil {
		slog.Warn("Docker client unavailable, using unsafe-only mode", "error", err)
	} else {
		// Health check before proceeding
		healthChecker := docker.NewHealthCheck(dockerClient.Docker())
		if err := healthChecker.Ping(context.Background()); err != nil {
			slog.Warn("Docker daemon not responding, using unsafe-only mode", "error", err)
			_ = dockerClient.Close()
		} else {
			// Create full Docker backend with all interfaces
			imageResolver := docker.NewResolver(dockerClient.Docker())

			dockerBack := dockerbackend.New(dockerbackend.Config{
				ImageName:     getEnvOrDefault("METATOOLS_DOCKER_IMAGE", "toolruntime-sandbox:latest"),
				Client:        dockerClient,
				ImageResolver: imageResolver,
				HealthChecker: healthChecker,
				Logger:        &slogAdapter{},
			})

			backends[toolruntime.ProfileStandard] = dockerBack

			// Use standard profile by default if Docker is available and configured
			if os.Getenv("METATOOLS_RUNTIME_PROFILE") == "standard" {
				defaultProfile = toolruntime.ProfileStandard
				slog.Info("Docker backend enabled, using standard profile",
					"image", getEnvOrDefault("METATOOLS_DOCKER_IMAGE", "toolruntime-sandbox:latest"))
			} else {
				slog.Info("Docker backend available (use METATOOLS_RUNTIME_PROFILE=standard to enable)")
			}
		}
	}

	rt := toolruntime.NewDefaultRuntime(toolruntime.RuntimeConfig{
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
