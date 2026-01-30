package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageResolver_Resolve_LocalImage(t *testing.T) {
	cli := requireDocker(t)

	resolver := NewResolver(cli)

	// alpine:latest should already exist from our test suite
	resolved, err := resolver.Resolve(context.Background(), "alpine:latest")

	require.NoError(t, err)
	assert.Equal(t, "alpine:latest", resolved)
}

func TestImageResolver_Resolve_PullsImage(t *testing.T) {
	cli := requireDocker(t)

	// First, remove the image if it exists
	_, _ = cli.ImageRemove(context.Background(), "alpine:3.18", image.RemoveOptions{Force: true})

	resolver := NewResolver(cli)

	// Should pull the image
	resolved, err := resolver.Resolve(context.Background(), "alpine:3.18")

	require.NoError(t, err)
	assert.Equal(t, "alpine:3.18", resolved)

	// Verify the image now exists
	_, _, err = cli.ImageInspectWithRaw(context.Background(), "alpine:3.18")
	require.NoError(t, err)
}

func TestImageResolver_Resolve_InvalidImage(t *testing.T) {
	cli := requireDocker(t)

	resolver := NewResolver(cli)

	_, err := resolver.Resolve(context.Background(), "nonexistent-registry.invalid/no-such-image:v999")

	require.Error(t, err)
}

func TestImageResolver_Resolve_ContextCancel(t *testing.T) {
	cli := requireDocker(t)

	// First, remove the image to force a pull
	_, _ = cli.ImageRemove(context.Background(), "alpine:3.17", image.RemoveOptions{Force: true})

	resolver := NewResolver(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give context time to expire
	time.Sleep(5 * time.Millisecond)

	_, err := resolver.Resolve(ctx, "alpine:3.17")

	require.Error(t, err)
}
