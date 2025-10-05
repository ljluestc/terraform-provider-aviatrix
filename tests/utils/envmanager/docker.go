package envmanager

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// DockerManager manages Docker-based test environments
type DockerManager struct {
	networkName string
	subnet      string
}

// NewDockerManager creates a new Docker manager
func NewDockerManager(networkName, subnet string) *DockerManager {
	return &DockerManager{
		networkName: networkName,
		subnet:      subnet,
	}
}

// CreateNetwork creates a Docker network for isolated testing
func (dm *DockerManager) CreateNetwork(ctx context.Context) error {
	// Check if network already exists
	checkCmd := exec.CommandContext(ctx, "docker", "network", "ls", "--filter", fmt.Sprintf("name=%s", dm.networkName), "-q")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check network: %w", err)
	}

	if len(strings.TrimSpace(string(output))) > 0 {
		// Network already exists
		return nil
	}

	// Create network
	args := []string{"network", "create", "--driver", "bridge"}
	if dm.subnet != "" {
		args = append(args, "--subnet", dm.subnet)
	}
	args = append(args, dm.networkName)

	cmd := exec.CommandContext(ctx, "docker", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create network: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RemoveNetwork removes the Docker network
func (dm *DockerManager) RemoveNetwork(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "network", "rm", dm.networkName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if network doesn't exist
		if !strings.Contains(string(output), "not found") {
			return fmt.Errorf("failed to remove network: %w\nOutput: %s", err, string(output))
		}
	}

	return nil
}

// CreateContainer creates a Docker container for the environment
func (dm *DockerManager) CreateContainer(ctx context.Context, env *Environment, image string, envVars map[string]string, volumes map[string]string) error {
	args := []string{"run", "-d", "--name", env.ID, "--network", dm.networkName}

	// Add environment variables
	for key, value := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volumes
	for host, container := range volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, container))
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create container: %w\nOutput: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	env.ContainerID = containerID
	env.NetworkID = dm.networkName

	// Get container IP
	ip, err := dm.GetContainerIP(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container IP: %w", err)
	}
	env.AddResource("container_ip", ip)

	return nil
}

// StopContainer stops a Docker container
func (dm *DockerManager) StopContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop container: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RemoveContainer removes a Docker container
func (dm *DockerManager) RemoveContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if container doesn't exist
		if !strings.Contains(string(output), "no such container") &&
			!strings.Contains(string(output), "No such container") {
			return fmt.Errorf("failed to remove container: %w\nOutput: %s", err, string(output))
		}
	}

	return nil
}

// GetContainerIP retrieves the IP address of a container
func (dm *DockerManager) GetContainerIP(ctx context.Context, containerID string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckContainerHealth checks if a container is healthy
func (dm *DockerManager) CheckContainerHealth(ctx context.Context, containerID string) (bool, error) {
	// Check if container is running
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check container status: %w\nOutput: %s", err, string(output))
	}

	running := strings.TrimSpace(string(output)) == "true"
	if !running {
		return false, nil
	}

	// Check health status if health check is defined
	cmd = exec.CommandContext(ctx, "docker", "inspect", "-f", "{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}", containerID)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check container health: %w\nOutput: %s", err, string(output))
	}

	healthStatus := strings.TrimSpace(string(output))
	if healthStatus == "none" {
		// No health check defined, consider healthy if running
		return true, nil
	}

	return healthStatus == "healthy", nil
}

// ExecCommand executes a command in a container
func (dm *DockerManager) ExecCommand(ctx context.Context, containerID string, command []string) (string, error) {
	args := append([]string{"exec", containerID}, command...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to execute command: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// GetContainerLogs retrieves logs from a container
func (dm *DockerManager) GetContainerLogs(ctx context.Context, containerID string, tail int) (string, error) {
	args := []string{"logs"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// PruneContainers removes stopped containers
func (dm *DockerManager) PruneContainers(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "container", "prune", "-f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prune containers: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ListContainers lists all containers in the test network
func (dm *DockerManager) ListContainers(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("network=%s", dm.networkName), "--format", "{{.ID}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w\nOutput: %s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []string
	for _, line := range lines {
		if line != "" {
			containers = append(containers, strings.TrimSpace(line))
		}
	}

	return containers, nil
}
