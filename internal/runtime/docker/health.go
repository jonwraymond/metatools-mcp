package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	dockerbackend "github.com/jonwraymond/toolruntime/backend/docker"
)

// HealthCheck implements toolruntime's HealthChecker interface using Docker SDK.
type HealthCheck struct {
	docker *client.Client
}

// NewHealthCheck creates a new health checker.
func NewHealthCheck(docker *client.Client) *HealthCheck {
	return &HealthCheck{docker: docker}
}

// Ping checks if the Docker daemon is responsive.
func (h *HealthCheck) Ping(ctx context.Context) error {
	_, err := h.docker.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon ping: %w", err)
	}
	return nil
}

// Info returns daemon information including version, OS, and architecture.
func (h *HealthCheck) Info(ctx context.Context) (dockerbackend.DaemonInfo, error) {
	info, err := h.docker.Info(ctx)
	if err != nil {
		return dockerbackend.DaemonInfo{}, fmt.Errorf("docker daemon info: %w", err)
	}

	return dockerbackend.DaemonInfo{
		Version:      info.ServerVersion,
		APIVersion:   h.docker.ClientVersion(),
		OS:           info.OSType,
		Architecture: info.Architecture,
		RootDir:      info.DockerRootDir,
	}, nil
}

// Ensure interface compliance at compile time.
var _ dockerbackend.HealthChecker = (*HealthCheck)(nil)
