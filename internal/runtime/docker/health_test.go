package docker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthChecker_Ping_Success(t *testing.T) {
	cli := requireDocker(t)

	checker := NewHealthCheck(cli)

	err := checker.Ping(context.Background())

	require.NoError(t, err)
}

func TestHealthChecker_Ping_ContextCancel(t *testing.T) {
	cli := requireDocker(t)

	checker := NewHealthCheck(cli)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := checker.Ping(ctx)

	require.Error(t, err)
}

func TestHealthChecker_Info_Success(t *testing.T) {
	cli := requireDocker(t)

	checker := NewHealthCheck(cli)

	info, err := checker.Info(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, info.Version, "Version should not be empty")
	assert.NotEmpty(t, info.APIVersion, "APIVersion should not be empty")
	assert.NotEmpty(t, info.OS, "OS should not be empty")
	assert.NotEmpty(t, info.Architecture, "Architecture should not be empty")
}

func TestHealthChecker_Info_ContextTimeout(t *testing.T) {
	cli := requireDocker(t)

	checker := NewHealthCheck(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give context time to expire
	time.Sleep(1 * time.Millisecond)

	_, err := checker.Info(ctx)

	require.Error(t, err)
}
