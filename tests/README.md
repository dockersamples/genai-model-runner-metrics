# Integration Tests for GenAI Application

This directory contains integration tests for the GenAI application with Testcontainers support.

## Quick Start

```bash
# Install dependencies
go get github.com/stretchr/testify/assert github.com/stretchr/testify/require
go mod tidy

# Run a single test to verify setup
go test -v ./integration -run TestSimple

# Run a specific integration test
go test -v ./integration -run TestGenAIAppIntegration

# Run extended performance tests
go test -v ./integration -run TestExtendedPerformance

# Run all tests
go test -v ./integration
```

## Available Tests

- **TestSimple**: Basic test to verify compilation and environment setup
- **TestBasicTestcontainer**: Validates the Testcontainers environment setup
- **TestGenAIAppIntegration**: Tests various API endpoints with different prompt types
- **TestLLMResponseQuality**: Validates the quality of LLM responses
- **TestLLMPerformance**: Measures performance metrics of the LLM service
- **TestMultiTurnConversation**: Tests context maintenance in conversations
- **TestChatPerformance**: Checks chat endpoint response times
- **TestChatQuality**: Verifies chat response quality for specific prompts
- **TestDockerIntegration**: Tests Docker-based deployments
- **TestExtendedPerformance**: Runs extended load tests over a longer period

## Test Structure

- `setup.go`: Contains the test environment setup code
- `test_helpers.go`: Helper functions for testing API endpoints
- `chat_request.go`: Functions for sending chat requests
- `quality_test.go`: Tests for chat response quality
- `performance_test.go`: Tests for API performance
- `llm_quality_test.go`: Tests for LLM response quality
- `genai_integration_test.go`: Tests for API endpoints integration
- `extended_test.go`: Extended performance tests
- `basic_testcontainer_test.go`: Basic test for Testcontainers functionality
- `simple_test.go`: Minimal test to verify package compilation

## Running Tests

### Basic Tests

To verify your testing environment is set up properly:

```bash
go test -v ./integration -run TestSimple
go test -v ./integration -run TestBasicTestcontainer
```

### Functional API Tests

Tests the API endpoints and response quality:

```bash
go test -v ./integration -run TestGenAIAppIntegration
go test -v ./integration -run TestChatQuality
```

### Performance Tests

Measures response times and performance characteristics:

```bash
go test -v ./integration -run TestChatPerformance
go test -v ./integration -run TestLLMPerformance
```

### Extended Performance Tests

Runs load tests for an extended period (30 seconds by default):

```bash
go test -v ./integration -run TestExtendedPerformance
```

### Short Mode Tests

Skip long-running tests:

```bash
go test -v ./integration -short
```

## Prerequisites

- Go 1.19 or higher
- Docker (for Testcontainers functionality)
- Running GenAI application at http://localhost:8080 (for API tests)

## Notes

- The tests expect your application to be running at http://localhost:8080
- The Testcontainers functionality is currently simplified, but the structure is in place for full container-based testing
- You can customize test duration and parameters in the respective test files