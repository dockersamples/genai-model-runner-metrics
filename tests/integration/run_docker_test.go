package integration

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestWithDockerCompose runs tests using docker-compose instead of testcontainers
// This is useful for CI environments or when you want to debug the tests
func TestWithDockerCompose(t *testing.T) {
	if os.Getenv("USE_DOCKER_COMPOSE") != "true" {
		t.Skip("Skipping docker-compose tests. Set USE_DOCKER_COMPOSE=true to run these tests.")
	}

	// Start the environment with docker-compose
	cmd := exec.Command("docker-compose", "-f", "../docker-compose.test.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to start docker-compose: %s", err)
	}
	defer func() {
		// Clean up after tests
		cmd := exec.Command("docker-compose", "-f", "../docker-compose.test.yml", "down")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}()

	// Wait for services to be ready
	t.Log("Waiting for services to be ready...")
	time.Sleep(30 * time.Second)

	// Define test URLs
	backendURL := "http://localhost:8080"
	frontendURL := "http://localhost:3000"

	// Run tests
	t.Run("API_Health", func(t *testing.T) {
		testHealthEndpoint(t, backendURL)
	})

	t.Run("API_Chat", func(t *testing.T) {
		testChatEndpoint(t, backendURL)
	})

	t.Run("API_Streaming", func(t *testing.T) {
		testStreamingResponse(t, backendURL)
	})

	// Uncomment to run browser tests if needed
	// t.Run("UI_Tests", func(t *testing.T) {
	// 	runPlaywrightTests(t, frontendURL)
	// })

	t.Log("Tests completed successfully!")
}
