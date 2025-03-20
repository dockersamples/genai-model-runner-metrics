# Integration Tests for GenAI Chat Application

This directory contains integration tests for the GenAI Chat application using Testcontainers.

## Overview

These tests validate the functionality of the entire application stack including:

- Backend API endpoints
- Frontend UI behavior
- LLM response quality
- Performance characteristics

## Test Categories

1. **API Tests** - Validates backend API endpoints and streaming functionality
2. **Frontend Tests** - Tests the UI components using Playwright
3. **Performance Tests** - Measures response times and throughput
4. **Quality Tests** - Evaluates LLM response quality against predefined criteria

## Prerequisites

To run these tests, you need:

- Go 1.19 or higher
- Docker and Docker Compose
- Playwright browser automation (installed automatically during test execution)

## Dependencies

Add the following dependencies to your Go project:

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/stretchr/testify/assert
go get github.com/playwright-community/playwright-go
```

## Running the Tests

```bash
# Install Playwright dependencies (first time only)
go run github.com/playwright-community/playwright-go/cmd/playwright install

# Run all integration tests
go test -v ./tests/integration

# Run a specific test category
go test -v ./tests/integration -run TestGenAIAppIntegration

# Run tests in short mode (skips long-running tests)
go test -v ./tests/integration -short
```

## Test Environment

The tests create isolated Docker containers for:

1. Ollama service running Llama 3.2
2. Backend Go API server
3. Frontend React application (for UI tests)

All containers are connected via a dedicated Docker network to ensure proper communication.

## Customizing Tests

To add new test cases or modify existing ones:

- For API tests, add new test functions to `api_test.go`
- For UI tests, add new test steps to the Playwright tests in `frontend_test.go`
- For quality evaluation, add new test cases to the `testCases` array in `quality_test.go`

## Best Practices

1. Keep tests isolated - each test should run independently
2. Use descriptive test names that indicate what's being tested
3. Include both positive and negative test cases
4. For UI tests, add screenshots to help debug failures
5. For LLM quality tests, focus on pattern matching rather than exact text comparison
