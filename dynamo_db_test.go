package dynamotest_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"

	"github.com/rozen03/dynamotest"
)

func checkDockerInstanceRunning(containerID string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	defer cli.Close()

	container, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return false, err
	}

	return container.State.Running, nil
}

func checkDockerInstanceRemoved(containerID string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	defer cli.Close()

	_, err = cli.ContainerInspect(context.Background(), containerID)
	if err != nil && client.IsErrNotFound(err) {
		return true, nil
	}
	return false, err
}

func TestDockerInstanceCreation(t *testing.T) {
	t.Parallel()

	// Create DynamoDB client and start Docker instance
	client, clean := dynamotest.NewDynamoDB()

	// Get the container ID from the client (assuming it stores this information)
	containerID := client.ContainerID // Modify as per your implementation

	// Check if Docker instance is running
	running, err := checkDockerInstanceRunning(containerID)
	if err != nil {
		t.Fatalf("Error checking if Docker instance is running: %v", err)
	}
	if !running {
		t.Fatalf("Docker instance is not running")
	}

	// Clean up Docker instance
	clean()

	// Wait for a while to make sure the Docker instance is stopped
	time.Sleep(2 * time.Second)

	// Check if Docker instance is removed
	removed, err := checkDockerInstanceRemoved(containerID)
	if err != nil {
		t.Fatalf("Error checking if Docker instance is removed: %v", err)
	}
	if !removed {
		t.Fatalf("Docker instance is still running after clean up")
	}
}

func TestDockerInstanceDeletion(t *testing.T) {
	// you would ask why this test is needed,
	// it is to ensure that the Docker instance is removed after the test is complete.
	// just in case the other test is modified and the cleanup is not called.
	t.Parallel()

	// Create DynamoDB client and start Docker instance
	client, clean := dynamotest.NewDynamoDB()

	// Get the container ID from the client (assuming it stores this information)
	containerID := client.ContainerID // Modify as per your implementation

	// Clean up Docker instance
	clean()

	// Wait for a while to make sure the Docker instance is stopped
	time.Sleep(2 * time.Second)

	// Check if Docker instance is removed
	removed, err := checkDockerInstanceRemoved(containerID)
	if err != nil {
		t.Fatalf("Error checking if Docker instance is removed: %v", err)
	}
	if !removed {
		t.Fatalf("Docker instance is still running after clean up")
	}
}
