package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenAIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup the test containers (reusing from previous tests)
	ctx := context.Background()
	apiURL := setupTestEnvironment(t, ctx)
	
	// Test parameters
	shortPrompt := "Hello, how are you?"
	mediumPrompt := "Explain what Docker is in a paragraph."
	longPrompt := "Write a short story about a container that becomes self-aware."
	
	// Run the performance tests
	t.Run("ShortPromptLatency", func(t *testing.T) {
		testPromptLatency(t, apiURL, shortPrompt, 5*time.Second)
	})
	
	t.Run("MediumPromptLatency", func(t *testing.T) {
		testPromptLatency(t, apiURL, mediumPrompt, 10*time.Second)
	})
	
	t.Run("LongPromptLatency", func(t *testing.T) {
		testPromptLatency(t, apiURL, longPrompt, 15*time.Second)
	})
	
	t.Run("ResponseThroughput", func(t *testing.T) {
		testResponseThroughput(t, apiURL, mediumPrompt)
	})
}

func testPromptLatency(t *testing.T, apiURL string, prompt string, maxLatency time.Duration) {
	// Create a chat request
	chatReq := ChatRequest{
		Message:  prompt,
		Messages: []Message{},
	}
	
	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		t.Fatalf("Failed to marshal chat request: %s", err)
	}
	
	// Measure time to first token
	start := time.Now()
	
	resp, err := http.Post(fmt.Sprintf("%s/chat", apiURL), "application/json", 
		strings.NewReader(string(reqBody)))
	if err != nil {
		t.Fatalf("Failed to call chat endpoint: %s", err)
	}
	defer resp.Body.Close()
	
	// Read the first byte (first token)
	buffer := make([]byte, 1)
	_, err = resp.Body.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response: %s", err)
	}
	
	timeToFirstToken := time.Since(start)
	
	t.Logf("Time to first token for prompt '%s': %v", prompt, timeToFirstToken)
	assert.Less(t, timeToFirstToken, maxLatency, 
		"Time to first token should be less than %v", maxLatency)
}

func testResponseThroughput(t *testing.T, apiURL string, prompt string) {
	// Create a chat request
	chatReq := ChatRequest{
		Message:  prompt,
		Messages: []Message{},
	}
	
	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		t.Fatalf("Failed to marshal chat request: %s", err)
	}
	
	// Send the request
	resp, err := http.Post(fmt.Sprintf("%s/chat", apiURL), "application/json", 
		strings.NewReader(string(reqBody)))
	if err != nil {
		t.Fatalf("Failed to call chat endpoint: %s", err)
	}
	defer resp.Body.Close()
	
	// Measure throughput (tokens per second)
	start := time.Now()
	buffer := make([]byte, 4096)
	totalBytes := 0
	chunkCount := 0
	
	for {
		n, err := resp.Body.Read(buffer)
		if err != nil || n == 0 {
			break
		}
		
		totalBytes += n
		chunkCount++
		
		// If we've read a significant amount or several chunks, we can stop
		// This is to make the test run faster
		if totalBytes > 1000 || chunkCount > 10 {
			break
		}
	}
	
	elapsed := time.Since(start)
	bytesPerSecond := float64(totalBytes) / elapsed.Seconds()
	
	// Very rough approximation: 1 token ≈ 4 bytes for English text
	tokensPerSecond := bytesPerSecond / 4
	
	t.Logf("Response throughput: %.2f bytes/sec (approx. %.2f tokens/sec)", 
		bytesPerSecond, tokensPerSecond)
	
	// A very basic assertion - actual performance will vary based on hardware
	assert.Greater(t, tokensPerSecond, 1.0, 
		"Token generation speed should be at least 1 token per second")
}
