package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A simple quality evaluation structure
type EvaluationCase struct {
	Prompt          string
	ExpectedContent []string // Strings that should be in the response
	ForbiddenContent []string // Strings that should not be in the response
	MinLength       int
}

func TestGenAIQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping quality test in short mode")
	}

	// Setup the test containers (reusing from previous tests)
	ctx := context.Background()
	apiURL := setupTestEnvironment(t, ctx)
	
	// Define test cases for quality evaluation
	testCases := []EvaluationCase{
		{
			Prompt:          "What is Docker?",
			ExpectedContent: []string{"container", "application", "software", "package"},
			ForbiddenContent: []string{"I don't know", "sorry, I can't"},
			MinLength:       100,
		},
		{
			Prompt:          "Generate a brief poem about containers",
			ExpectedContent: []string{"container"},
			MinLength:       50,
		},
		{
			Prompt:          "Explain the difference between Docker and virtual machines",
			ExpectedContent: []string{"Docker", "virtual machine", "container", "hypervisor", "lightweight"},
			MinLength:       150,
		},
	}
	
	// Run quality tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("QualityTest_%d", i), func(t *testing.T) {
			testResponseQuality(t, apiURL, tc)
		})
	}
}

func testResponseQuality(t *testing.T, apiURL string, tc EvaluationCase) {
	// Create a chat request
	chatReq := ChatRequest{
		Message:  tc.Prompt,
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
	
	// Read the full response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %s", err)
	}
	
	response := string(body)
	
	// Length check
	assert.GreaterOrEqual(t, len(response), tc.MinLength, 
		"Response should meet minimum length requirement")
	
	// Check for expected content
	for _, expected := range tc.ExpectedContent {
		assert.Contains(t, strings.ToLower(response), strings.ToLower(expected), 
			"Response should contain '%s'", expected)
	}
	
	// Check for forbidden content
	for _, forbidden := range tc.ForbiddenContent {
		assert.NotContains(t, strings.ToLower(response), strings.ToLower(forbidden), 
			"Response should not contain '%s'", forbidden)
	}
	
	// Log the response for manual inspection
	t.Logf("Response for prompt '%s': %s", tc.Prompt, response)
}
