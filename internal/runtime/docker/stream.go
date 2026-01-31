package docker

import (
	"bufio"
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	dockerbackend "github.com/jonwraymond/toolexec/runtime/backend/docker"
)

// StreamClient extends Client with streaming capabilities.
// It implements both ContainerRunner and StreamRunner interfaces.
type StreamClient struct {
	*Client
}

// NewStreamClient creates a client with streaming support.
func NewStreamClient(cfg ClientConfig) (*StreamClient, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &StreamClient{Client: client}, nil
}

// RunStream implements StreamRunner.RunStream with real-time output streaming.
// It returns a channel that receives streaming events for stdout, stderr, and exit.
func (c *StreamClient) RunStream(ctx context.Context, spec dockerbackend.ContainerSpec) (<-chan dockerbackend.StreamEvent, error) {
	events := make(chan dockerbackend.StreamEvent, 100)

	go c.runStreamInternal(ctx, spec, events)

	return events, nil
}

func (c *StreamClient) runStreamInternal(ctx context.Context, spec dockerbackend.ContainerSpec, events chan<- dockerbackend.StreamEvent) {
	defer close(events)

	// Apply timeout
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	// Create container
	containerCfg := &container.Config{
		Image:        spec.Image,
		Cmd:          spec.Command,
		WorkingDir:   spec.WorkingDir,
		Env:          spec.Env,
		Labels:       spec.Labels,
		AttachStdout: true,
		AttachStderr: true,
	}

	if spec.Security.User != "" {
		containerCfg.User = spec.Security.User
	}

	hostCfg := &container.HostConfig{
		Resources:      buildResources(spec.Resources),
		ReadonlyRootfs: spec.Security.ReadOnlyRootfs,
		Mounts:         convertMounts(spec.Mounts),
		SecurityOpt:    buildSecurityOpts(spec.Security),
	}

	if spec.Security.NetworkMode != "" {
		hostCfg.NetworkMode = container.NetworkMode(spec.Security.NetworkMode)
	}

	resp, err := c.docker.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		events <- dockerbackend.StreamEvent{Type: dockerbackend.StreamEventError, Error: err}
		return
	}
	containerID := resp.ID

	defer func() {
		removeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = c.docker.ContainerRemove(removeCtx, containerID, container.RemoveOptions{Force: true})
	}()

	// Attach to container for streaming
	attachResp, err := c.docker.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		events <- dockerbackend.StreamEvent{Type: dockerbackend.StreamEventError, Error: err}
		return
	}
	defer attachResp.Close()

	// Start container
	if err := c.docker.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		events <- dockerbackend.StreamEvent{Type: dockerbackend.StreamEventError, Error: err}
		return
	}

	// Stream output in background
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		c.streamOutput(attachResp.Reader, events)
	}()

	// Wait for completion
	statusCh, errCh := c.docker.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			events <- dockerbackend.StreamEvent{Type: dockerbackend.StreamEventError, Error: err}
			return
		}
	case status := <-statusCh:
		// Wait for output streaming to complete
		<-outputDone
		events <- dockerbackend.StreamEvent{
			Type:     dockerbackend.StreamEventExit,
			ExitCode: int(status.StatusCode),
		}
	case <-ctx.Done():
		events <- dockerbackend.StreamEvent{Type: dockerbackend.StreamEventError, Error: ctx.Err()}
	}
}

func (c *StreamClient) streamOutput(reader io.Reader, events chan<- dockerbackend.StreamEvent) {
	// Use stdcopy to demux Docker's multiplexed stdout/stderr stream
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	go func() {
		defer func() { _ = stdoutWriter.Close() }()
		defer func() { _ = stderrWriter.Close() }()
		_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, reader)
	}()

	// Stream stdout and stderr concurrently
	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			events <- dockerbackend.StreamEvent{
				Type: dockerbackend.StreamEventStdout,
				Data: append(scanner.Bytes(), '\n'),
			}
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			events <- dockerbackend.StreamEvent{
				Type: dockerbackend.StreamEventStderr,
				Data: append(scanner.Bytes(), '\n'),
			}
		}
	}()

	// Wait for both to complete
	<-done
	<-done
}

// Ensure interface compliance at compile time.
var _ dockerbackend.StreamRunner = (*StreamClient)(nil)
