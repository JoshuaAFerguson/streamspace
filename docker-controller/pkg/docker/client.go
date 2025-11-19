// Package docker provides Docker container management for StreamSpace sessions.
package docker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Client wraps the Docker API client for StreamSpace operations.
type Client struct {
	docker      *client.Client
	networkName string
}

// NewClient creates a new Docker client.
func NewClient(host, networkName string) (*Client, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}

	if host != "" && host != "unix:///var/run/docker.sock" {
		opts = append(opts, client.WithHost(host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Test connection
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	return &Client{
		docker:      cli,
		networkName: networkName,
	}, nil
}

// Close closes the Docker client.
func (c *Client) Close() error {
	return c.docker.Close()
}

// SessionConfig holds configuration for creating a session container.
type SessionConfig struct {
	SessionID      string
	UserID         string
	TemplateID     string
	Image          string
	Memory         int64  // bytes
	CPUShares      int64
	VNCPort        int
	PersistentHome bool
	HomeVolume     string
	Env            map[string]string
}

// CreateSession creates a new session container.
func (c *Client) CreateSession(ctx context.Context, config SessionConfig) (string, error) {
	containerName := fmt.Sprintf("ss-%s", config.SessionID)

	// Build environment variables
	env := []string{
		fmt.Sprintf("SESSION_ID=%s", config.SessionID),
		fmt.Sprintf("USER_ID=%s", config.UserID),
		fmt.Sprintf("TEMPLATE_ID=%s", config.TemplateID),
	}
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Configure port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	if config.VNCPort > 0 {
		vncPort := nat.Port(fmt.Sprintf("%d/tcp", config.VNCPort))
		exposedPorts[vncPort] = struct{}{}
		portBindings[vncPort] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: ""}, // Auto-assign host port
		}
	}

	// Configure mounts
	var mounts []mount.Mount
	if config.PersistentHome && config.HomeVolume != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: config.HomeVolume,
			Target: "/config",
		})
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:        config.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"streamspace.io/managed":  "true",
			"streamspace.io/session":  config.SessionID,
			"streamspace.io/user":     config.UserID,
			"streamspace.io/template": config.TemplateID,
		},
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       mounts,
		Resources: container.Resources{
			Memory:    config.Memory,
			CPUShares: config.CPUShares,
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			c.networkName: {},
		},
	}

	// Create container
	resp, err := c.docker.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := c.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		// Clean up on failure
		c.docker.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	log.Printf("Created and started container %s for session %s", containerName, config.SessionID)
	return resp.ID, nil
}

// StopSession stops (hibernates) a session container.
func (c *Client) StopSession(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("ss-%s", sessionID)

	timeout := 30 // seconds
	if err := c.docker.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &timeout}); err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return nil // Already stopped/removed
		}
		return fmt.Errorf("failed to stop container: %w", err)
	}

	log.Printf("Stopped container %s for session %s", containerName, sessionID)
	return nil
}

// StartSession starts (wakes) a hibernated session container.
func (c *Client) StartSession(ctx context.Context, sessionID string) error {
	containerName := fmt.Sprintf("ss-%s", sessionID)

	if err := c.docker.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	log.Printf("Started container %s for session %s", containerName, sessionID)
	return nil
}

// RemoveSession removes a session container.
func (c *Client) RemoveSession(ctx context.Context, sessionID string, force bool) error {
	containerName := fmt.Sprintf("ss-%s", sessionID)

	if err := c.docker.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         force,
		RemoveVolumes: false, // Keep volumes for data persistence
	}); err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return nil // Already removed
		}
		return fmt.Errorf("failed to remove container: %w", err)
	}

	log.Printf("Removed container %s for session %s", containerName, sessionID)
	return nil
}

// GetSessionStatus returns the status of a session container.
func (c *Client) GetSessionStatus(ctx context.Context, sessionID string) (string, error) {
	containerName := fmt.Sprintf("ss-%s", sessionID)

	info, err := c.docker.ContainerInspect(ctx, containerName)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return "not_found", nil
		}
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	if info.State.Running {
		return "running", nil
	}
	if info.State.Paused {
		return "paused", nil
	}
	return "stopped", nil
}

// GetSessionURL returns the URL to access the session.
func (c *Client) GetSessionURL(ctx context.Context, sessionID string, vncPort int) (string, error) {
	containerName := fmt.Sprintf("ss-%s", sessionID)

	info, err := c.docker.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	portKey := fmt.Sprintf("%d/tcp", vncPort)
	if bindings, ok := info.NetworkSettings.Ports[nat.Port(portKey)]; ok && len(bindings) > 0 {
		return fmt.Sprintf("http://localhost:%s", bindings[0].HostPort), nil
	}

	return "", fmt.Errorf("VNC port not exposed")
}

// EnsureUserVolume creates a volume for user's persistent home if it doesn't exist.
func (c *Client) EnsureUserVolume(ctx context.Context, userID string) (string, error) {
	volumeName := fmt.Sprintf("streamspace-home-%s", userID)

	// Check if volume exists
	_, err := c.docker.VolumeInspect(ctx, volumeName)
	if err == nil {
		return volumeName, nil // Already exists
	}

	// Create volume
	_, err = c.docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
		Labels: map[string]string{
			"streamspace.io/managed": "true",
			"streamspace.io/user":    userID,
			"streamspace.io/type":    "home",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	log.Printf("Created volume %s for user %s", volumeName, userID)
	return volumeName, nil
}

// ListSessions returns all StreamSpace session containers.
func (c *Client) ListSessions(ctx context.Context) ([]string, error) {
	containers, err := c.docker.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "streamspace.io/managed=true"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var sessions []string
	for _, c := range containers {
		if sessionID, ok := c.Labels["streamspace.io/session"]; ok {
			sessions = append(sessions, sessionID)
		}
	}

	return sessions, nil
}
