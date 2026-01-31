package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	dockerbackend "github.com/jonwraymond/toolexec/runtime/backend/docker"
)

// Client implements toolruntime's ContainerRunner interface using Docker SDK.
type Client struct {
	docker *client.Client
}

// ClientConfig holds configuration for the Docker client.
type ClientConfig struct {
	// Host is the Docker daemon socket (default: unix:///var/run/docker.sock)
	Host string

	// APIVersion is the Docker API version (default: negotiated)
	APIVersion string
}

// NewClient creates a new Docker client.
func NewClient(cfg ClientConfig) (*Client, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}
	if cfg.Host != "" {
		opts = append(opts, client.WithHost(cfg.Host))
	}

	docker, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Client{docker: docker}, nil
}

// Docker returns the underlying Docker client for use with other components.
func (c *Client) Docker() *client.Client {
	return c.docker
}

// Run implements ContainerRunner.Run.
// It creates, starts, waits for, and removes a container atomically.
func (c *Client) Run(ctx context.Context, spec dockerbackend.ContainerSpec) (dockerbackend.ContainerResult, error) {
	start := time.Now()

	// Apply timeout if specified
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	// Create container config
	containerCfg := &container.Config{
		Image:      spec.Image,
		Cmd:        spec.Command,
		WorkingDir: spec.WorkingDir,
		Env:        spec.Env,
		Labels:     spec.Labels,
	}

	// Apply security settings
	if spec.Security.User != "" {
		containerCfg.User = spec.Security.User
	}

	// Create host config with resource limits and security
	hostCfg := &container.HostConfig{
		Resources:      buildResources(spec.Resources),
		ReadonlyRootfs: spec.Security.ReadOnlyRootfs,
		Mounts:         convertMounts(spec.Mounts),
		SecurityOpt:    buildSecurityOpts(spec.Security),
	}

	// Set network mode if specified
	if spec.Security.NetworkMode != "" {
		hostCfg.NetworkMode = container.NetworkMode(spec.Security.NetworkMode)
	}

	// Create container
	resp, err := c.docker.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return dockerbackend.ContainerResult{}, fmt.Errorf("create container: %w", err)
	}
	containerID := resp.ID

	// Always remove container on exit
	defer func() {
		removeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = c.docker.ContainerRemove(removeCtx, containerID, container.RemoveOptions{Force: true})
	}()

	// Start container
	if err := c.docker.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return dockerbackend.ContainerResult{}, fmt.Errorf("start container: %w", err)
	}

	// Wait for completion
	statusCh, errCh := c.docker.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	var exitCode int64
	select {
	case err := <-errCh:
		if err != nil {
			// Context cancellation - return the context error
			if ctx.Err() != nil {
				return dockerbackend.ContainerResult{}, ctx.Err()
			}
			return dockerbackend.ContainerResult{}, fmt.Errorf("wait container: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	case <-ctx.Done():
		return dockerbackend.ContainerResult{}, ctx.Err()
	}

	// Capture logs
	stdout, stderr, err := c.captureLogs(context.Background(), containerID)
	if err != nil {
		return dockerbackend.ContainerResult{}, fmt.Errorf("capture logs: %w", err)
	}

	return dockerbackend.ContainerResult{
		ExitCode: int(exitCode),
		Stdout:   stdout,
		Stderr:   stderr,
		Duration: time.Since(start),
	}, nil
}

// captureLogs retrieves stdout and stderr from a container.
func (c *Client) captureLogs(ctx context.Context, containerID string) (string, string, error) {
	logs, err := c.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", "", err
	}
	defer func() { _ = logs.Close() }()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, logs)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

// buildResources converts ResourceSpec to container.Resources.
func buildResources(spec dockerbackend.ResourceSpec) container.Resources {
	res := container.Resources{}

	if spec.MemoryBytes > 0 {
		res.Memory = spec.MemoryBytes
	}
	if spec.CPUQuota > 0 {
		res.CPUQuota = spec.CPUQuota
	}
	if spec.PidsLimit > 0 {
		res.PidsLimit = &spec.PidsLimit
	}

	return res
}

// convertMounts converts toolruntime mounts to Docker SDK mounts.
func convertMounts(mounts []dockerbackend.Mount) []mount.Mount {
	if len(mounts) == 0 {
		return nil
	}

	result := make([]mount.Mount, len(mounts))
	for i, m := range mounts {
		result[i] = mount.Mount{
			Type:     mount.Type(m.Type),
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}
	}
	return result
}

// buildSecurityOpts creates Docker security options from SecuritySpec.
func buildSecurityOpts(sec dockerbackend.SecuritySpec) []string {
	var opts []string
	if sec.SeccompProfile != "" {
		opts = append(opts, "seccomp="+sec.SeccompProfile)
	}
	return opts
}

// Close closes the Docker client connection.
func (c *Client) Close() error {
	return c.docker.Close()
}

// Ensure interface compliance at compile time.
var _ dockerbackend.ContainerRunner = (*Client)(nil)
