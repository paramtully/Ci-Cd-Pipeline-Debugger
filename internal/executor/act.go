package executor

import "github.com/moby/moby/client"

type DockerExecutor struct {
	docker *client.Client
}

// NewDockerExecutor wraps an Engine API client for job execution.
// REQUIRES: docker is non-nil
// MODIFIES: nothing
// EFFECTS: returns a DockerExecutor that uses docker for container operations
func NewDockerExecutor(docker *client.Client) *DockerExecutor {
	return &DockerExecutor{docker: docker}
}
