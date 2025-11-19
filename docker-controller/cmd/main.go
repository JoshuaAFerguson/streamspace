// Package main is the entry point for the StreamSpace Docker controller.
//
// This controller manages StreamSpace sessions using Docker containers instead
// of Kubernetes. It subscribes to NATS events and performs Docker operations.
//
// Key responsibilities:
//   - Session container lifecycle (create, start, stop, remove)
//   - Container networking and port mapping
//   - Volume management for persistent home directories
//   - Auto-hibernation (stop containers) and wake (start containers)
//
// Architecture:
//   - Subscribes to NATS events on streamspace.*.docker subjects
//   - Uses Docker API to manage containers
//   - Publishes status events back to NATS
//
// Deployment:
//   The controller can run as a standalone binary or Docker container with:
//   - Access to Docker socket (/var/run/docker.sock)
//   - NATS connection for event communication
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streamspace/docker-controller/pkg/docker"
	"github.com/streamspace/docker-controller/pkg/events"
)

func main() {
	var natsURL string
	var natsUser string
	var natsPassword string
	var controllerID string
	var dockerHost string
	var networkName string

	// Parse command-line flags
	flag.StringVar(&natsURL, "nats-url", getEnv("NATS_URL", "nats://localhost:4222"), "NATS server URL")
	flag.StringVar(&natsUser, "nats-user", getEnv("NATS_USER", ""), "NATS username")
	flag.StringVar(&natsPassword, "nats-password", getEnv("NATS_PASSWORD", ""), "NATS password")
	flag.StringVar(&controllerID, "controller-id", getEnv("CONTROLLER_ID", "streamspace-docker-controller-1"), "Unique controller ID")
	flag.StringVar(&dockerHost, "docker-host", getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"), "Docker host")
	flag.StringVar(&networkName, "network", getEnv("DOCKER_NETWORK", "streamspace"), "Docker network name")
	flag.Parse()

	log.Printf("StreamSpace Docker Controller starting...")
	log.Printf("NATS URL: %s", natsURL)
	log.Printf("Controller ID: %s", controllerID)
	log.Printf("Docker Host: %s", dockerHost)

	// Initialize Docker client
	dockerClient, err := docker.NewClient(dockerHost, networkName)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Initialize NATS event subscriber
	subscriber, err := events.NewSubscriber(events.Config{
		URL:      natsURL,
		User:     natsUser,
		Password: natsPassword,
	}, dockerClient, controllerID)

	if err != nil {
		log.Fatalf("Failed to create NATS subscriber: %v", err)
	}
	defer subscriber.Close()

	// Start subscriber in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := subscriber.Start(ctx); err != nil {
			log.Printf("NATS subscriber error: %v", err)
		}
	}()

	log.Printf("Docker controller started successfully")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Printf("Shutting down Docker controller...")
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
