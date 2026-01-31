package docker

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	dockerbackend "github.com/jonwraymond/toolexec/runtime/backend/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireDocker skips the test if Docker is not available.
func requireDocker(t *testing.T) *client.Client {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("Docker client not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skipf("Docker daemon not running: %v", err)
	}

	if err := ensureImage(ctx, cli, "alpine:latest"); err != nil {
		t.Skipf("Docker image unavailable: %v", err)
	}

	return cli
}

func ensureImage(ctx context.Context, cli *client.Client, imageRef string) error {
	args := filters.NewArgs()
	args.Add("reference", imageRef)

	images, err := cli.ImageList(ctx, image.ListOptions{Filters: args})
	if err != nil {
		return err
	}
	if len(images) > 0 {
		return nil
	}

	rc, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	_, _ = io.Copy(io.Discard, rc)

	return nil
}

func TestClient_Run_Success(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"echo", "hello"},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello\n", result.Stdout)
	assert.Empty(t, result.Stderr)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestClient_Run_NonZeroExit(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "exit 42"},
	})

	require.NoError(t, err)
	assert.Equal(t, 42, result.ExitCode)
}

func TestClient_Run_Stderr(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo stdout && echo stderr >&2"},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "stdout\n", result.Stdout)
	assert.Equal(t, "stderr\n", result.Stderr)
}

func TestClient_Run_Timeout(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	start := time.Now()
	_, err = c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sleep", "60"},
		Timeout: 500 * time.Millisecond,
	})

	elapsed := time.Since(start)

	// Should fail due to timeout
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context") || strings.Contains(err.Error(), "timeout"),
		"expected timeout error, got: %v", err)
	// Should not wait the full 60 seconds
	assert.Less(t, elapsed, 5*time.Second)
}

func TestClient_Run_ContextCancel(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err = c.Run(ctx, dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sleep", "60"},
	})

	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 5*time.Second)
}

func TestClient_Run_ImageNotFound(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	_, err = c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "nonexistent-image-that-does-not-exist:v999.999.999",
		Command: []string{"echo", "hello"},
	})

	require.Error(t, err)
	// The error should indicate the image was not found
	assert.True(t, strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "No such image") ||
		strings.Contains(err.Error(), "pull access denied"),
		"expected image not found error, got: %v", err)
}

func TestClient_Run_ContainerCleanup(t *testing.T) {
	requireDocker(t)

	cli := requireDocker(t)
	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	// Run a container that will fail
	_, _ = c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "exit 1"},
		Labels:  map[string]string{"test.cleanup": "true"},
	})

	// List containers with the label - should find none
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "test.cleanup=true"),
		),
	})
	require.NoError(t, err)
	assert.Empty(t, containers, "container should be cleaned up after execution")
}

func TestClient_Run_WorkingDir(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:      "alpine:latest",
		Command:    []string{"pwd"},
		WorkingDir: "/tmp",
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "/tmp\n", result.Stdout)
}

func TestClient_Run_EnvVars(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo $MY_VAR"},
		Env:     []string{"MY_VAR=hello-world"},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello-world\n", result.Stdout)
}

func TestClient_Run_MemoryLimit(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	// Set a very low memory limit (6MB) - alpine needs ~4MB minimum
	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"cat", "/sys/fs/cgroup/memory.max"},
		Resources: dockerbackend.ResourceSpec{
			MemoryBytes: 6 * 1024 * 1024, // 6MB
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	// The cgroup should show the memory limit
	assert.Contains(t, result.Stdout, "6291456") // 6MB in bytes
}

func TestClient_Run_ReadOnlyRootfs(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "touch /test-file 2>&1 || echo 'read-only'"},
		Security: dockerbackend.SecuritySpec{
			ReadOnlyRootfs: true,
		},
	})

	require.NoError(t, err)
	// Should fail to write because rootfs is read-only
	assert.Contains(t, result.Stdout, "read-only")
}

func TestClient_Run_NetworkNone(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"cat", "/sys/class/net/eth0/operstate"},
		Security: dockerbackend.SecuritySpec{
			NetworkMode: "none",
		},
	})

	// With network=none, eth0 won't exist
	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ExitCode, "eth0 should not exist with network=none")
}

func TestClient_Run_NonRootUser(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"id", "-u"},
		Security: dockerbackend.SecuritySpec{
			User: "nobody",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	// nobody user typically has UID 65534
	assert.NotEqual(t, "0\n", result.Stdout, "should not run as root")
}

func TestClient_Close(t *testing.T) {
	requireDocker(t)

	c, err := NewClient(ClientConfig{})
	require.NoError(t, err)

	// Close should not error
	err = c.Close()
	require.NoError(t, err)

	// Note: Docker SDK may reconnect on subsequent operations,
	// so we don't test for errors after close. The important thing
	// is that Close() itself doesn't error.
}
