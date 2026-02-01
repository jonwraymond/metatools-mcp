package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	dockerbackend "github.com/jonwraymond/toolexec/runtime/backend/docker"
)

// Resolver implements toolruntime's ImageResolver interface using Docker SDK.
type Resolver struct {
	docker *client.Client
}

// NewResolver creates a new image resolver.
func NewResolver(docker *client.Client) *Resolver {
	return &Resolver{docker: docker}
}

// Resolve ensures the image is available locally, pulling if necessary.
// Returns the resolved image reference (same as input for now).
func (r *Resolver) Resolve(ctx context.Context, imageRef string) (string, error) {
	// Check if image exists locally
	_, err := r.docker.ImageInspect(ctx, imageRef)
	if err == nil {
		return imageRef, nil // Already exists
	}

	// Check context before attempting pull
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Pull the image
	reader, err := r.docker.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("pull image %s: %w", imageRef, err)
	}
	defer func() { _ = reader.Close() }()

	// Drain the reader to complete the pull (captures progress)
	decoder := json.NewDecoder(reader)
	for {
		var event map[string]any
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			// Check for context cancellation
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "", fmt.Errorf("decode pull progress: %w", err)
		}

		// Check for error in the pull response
		if errMsg, ok := event["error"].(string); ok {
			return "", fmt.Errorf("pull error: %s", errMsg)
		}
	}

	return imageRef, nil
}

// Ensure interface compliance at compile time.
var _ dockerbackend.ImageResolver = (*Resolver)(nil)
