package docker

import (
	"context"
	"strings"
	"testing"
	"time"

	dockerbackend "github.com/jonwraymond/toolruntime/backend/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamRunner_RunStream_Success(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	events, err := c.RunStream(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo line1; echo line2; echo line3"},
	})
	require.NoError(t, err)

	var stdout []string
	var exitCode int
	var gotExit bool

	for event := range events {
		switch event.Type {
		case dockerbackend.StreamEventStdout:
			stdout = append(stdout, strings.TrimSpace(string(event.Data)))
		case dockerbackend.StreamEventExit:
			exitCode = event.ExitCode
			gotExit = true
		case dockerbackend.StreamEventError:
			t.Fatalf("unexpected error event: %v", event.Error)
		}
	}

	assert.True(t, gotExit, "should receive exit event")
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "line1")
	assert.Contains(t, stdout, "line2")
	assert.Contains(t, stdout, "line3")
}

func TestStreamRunner_RunStream_Stderr(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	events, err := c.RunStream(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo stdout && echo stderr >&2"},
	})
	require.NoError(t, err)

	var hasStdout, hasStderr bool

	for event := range events {
		switch event.Type {
		case dockerbackend.StreamEventStdout:
			if strings.Contains(string(event.Data), "stdout") {
				hasStdout = true
			}
		case dockerbackend.StreamEventStderr:
			if strings.Contains(string(event.Data), "stderr") {
				hasStderr = true
			}
		}
	}

	assert.True(t, hasStdout, "should receive stdout event")
	assert.True(t, hasStderr, "should receive stderr event")
}

func TestStreamRunner_RunStream_ExitEvent(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	events, err := c.RunStream(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "exit 42"},
	})
	require.NoError(t, err)

	var exitCode int
	var gotExit bool

	for event := range events {
		if event.Type == dockerbackend.StreamEventExit {
			exitCode = event.ExitCode
			gotExit = true
		}
	}

	assert.True(t, gotExit, "should receive exit event")
	assert.Equal(t, 42, exitCode)
}

func TestStreamRunner_RunStream_ContextCancel(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	events, err := c.RunStream(ctx, dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sleep", "60"},
	})
	require.NoError(t, err)

	// Cancel after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	var gotError bool
	start := time.Now()

	for event := range events {
		if event.Type == dockerbackend.StreamEventError {
			gotError = true
		}
	}

	elapsed := time.Since(start)

	assert.True(t, gotError, "should receive error event on cancel")
	assert.Less(t, elapsed, 5*time.Second, "should not wait for full sleep")
}

func TestStreamRunner_RunStream_Timeout(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	events, err := c.RunStream(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"sleep", "60"},
		Timeout: 500 * time.Millisecond,
	})
	require.NoError(t, err)

	var gotError bool
	start := time.Now()

	for event := range events {
		if event.Type == dockerbackend.StreamEventError {
			gotError = true
		}
	}

	elapsed := time.Since(start)

	assert.True(t, gotError, "should receive error event on timeout")
	assert.Less(t, elapsed, 5*time.Second, "should respect timeout")
}

// TestStreamRunner_Implements_ContainerRunner verifies StreamClient also implements ContainerRunner
func TestStreamRunner_Implements_ContainerRunner(t *testing.T) {
	requireDocker(t)

	c, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	// StreamClient embeds Client, so it should work as ContainerRunner
	result, err := c.Run(context.Background(), dockerbackend.ContainerSpec{
		Image:   "alpine:latest",
		Command: []string{"echo", "hello"},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello\n", result.Stdout)
}
