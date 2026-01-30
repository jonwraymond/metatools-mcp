// Package docker provides Docker SDK implementations of toolruntime's container interfaces.
//
// This package implements:
//   - ContainerRunner: executes code in Docker containers with full lifecycle management
//   - ImageResolver: ensures images are available locally, pulling if necessary
//   - HealthChecker: verifies Docker daemon availability
//   - StreamRunner: provides real-time stdout/stderr streaming
//
// # Architecture
//
// The implementations in this package satisfy interfaces defined in toolruntime/backend/docker.
// This separation allows toolruntime to remain agnostic to the specific container runtime
// while metatools-mcp provides the concrete Docker SDK integration.
//
// # Usage
//
// Create a client and wire it into toolruntime's Docker backend:
//
//	client, err := docker.NewClient(docker.ClientConfig{})
//	if err != nil {
//	    return err
//	}
//
//	backend := dockerbackend.New(dockerbackend.Config{
//	    ImageName:     "toolruntime-sandbox:latest",
//	    Client:        client,
//	    ImageResolver: docker.NewResolver(client.Docker()),
//	    HealthChecker: docker.NewHealthCheck(client.Docker()),
//	})
//
// # Testing
//
// Tests require a running Docker daemon. Use -short to skip integration tests:
//
//	go test ./internal/runtime/docker/... -short
//
// # Security
//
// The client enforces security settings from ContainerSpec:
//   - Resource limits (memory, CPU, PIDs)
//   - Network isolation (NetworkMode: "none")
//   - Read-only root filesystem
//   - Non-root user execution
//   - Seccomp profiles
//
// Containers are always removed after execution, even on errors.
package docker
