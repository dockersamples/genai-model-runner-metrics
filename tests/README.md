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

Alternatively, you can use the provided Makefile:

```bash
# Setup dependencies
make -C tests setup

# Run all tests
make -C tests test

# Run specific test categories
make -C tests test-api
make -C tests test-frontend
make -C tests test-performance
make -C tests test-quality

# Run tests in short mode
make -C tests test-short

# Clean up test artifacts
make -C tests clean
```

## Test Environment

The tests create isolated Docker containers for:

1. Ollama service running Llama 3.2
2. Backend Go API server
3. Frontend React application (for UI tests)

All containers are connected via a dedicated Docker network to ensure proper communication.

## Test Reports

Test results are displayed in the console. Screenshots for UI tests are saved to the system's temporary directory for visual inspection when failures occur.

## Extending the Tests

To add new test cases or modify existing ones:

- For API tests, add new test functions to `api_test.go`
- For UI tests, add new test steps to the Playwright tests in `frontend_test.go`
- For quality evaluation, add new test cases to the `testCases` array in `quality_test.go`
