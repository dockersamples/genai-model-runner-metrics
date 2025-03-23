# Testing GenAI Applications with Docker Model Runner and Testcontainers

This directory contains integration tests for the GenAI application, demonstrating how to test applications that use Docker Model Runner with Testcontainers.

## Overview

These tests showcase two approaches to testing GenAI applications:

1. **Direct Model Runner Testing**: Tests that directly interact with Docker Model Runner through a Socat container.
2. **Full Application Testing**: Tests that run the backend application and connect it to Docker Model Runner.

## Prerequisites

- Docker Desktop 4.40+ with Model Runner enabled
- Go 1.19 or higher
- Docker Model Runner CLI (model-runner)
- The Llama3.2 model pulled from Docker Hub (`docker model pull ignaciolopezluna020/llama3.2:1b`)

## Setting Up Model Runner

Before running the tests, ensure Docker Model Runner is enabled in Docker Desktop:

1. Open Docker Desktop settings
2. Navigate to "Features in development"
3. Enable "Docker Model Runner"
4. Enable "Host-side TCP support" (set port to 12434)
5. Apply & Restart

## Test Files

- `model_runner_test.go`: Tests direct interaction with Docker Model Runner
- `app_integration_test.go`: Tests the full application stack with Model Runner

## Running the Tests

From the `tests` directory, run:

```bash
# Run all tests
go test -v ./...

# Run only Model Runner tests
go test -v -run "TestModelRunner" ./...

# Run full application tests
go test -v -run "TestFullApp" ./...
```

## How It Works

### Direct Model Runner Testing

The `TestModelRunnerIntegration` test:

1. Creates a Socat container that tunnels to Model Runner's internal DNS (`model-runner.docker.internal`)
2. Ensures the required model is available (pulls it if needed)
3. Sends direct API requests to Model Runner
4. Validates the responses

### Full Application Testing

The `TestFullAppWithModelRunner` test:

1. Creates a Socat container to tunnel to Model Runner
2. Starts the backend application with environment variables pointing to the Socat container
3. Sends chat requests to the backend, which in turn calls Model Runner
4. Validates the complete flow works correctly

## How Testcontainers is Used

In these tests, Testcontainers is used to:

1. Create and manage the Socat container that provides network access to Model Runner
2. Handle container lifecycle (creation, configuration, and cleanup)
3. Provide isolation for testing

## Benefits of This Approach

1. **Clean Environment**: Tests run in isolation with fresh containers
2. **Realistic Testing**: Tests the actual interaction with Model Runner
3. **GPU Acceleration**: Takes advantage of Model Runner's GPU acceleration
4. **Complete Testing**: Tests the entire application stack
5. **No Mock Dependencies**: Uses real Model Runner service instead of mocks

## Debugging Tips

1. If the tests fail, check that Docker Model Runner is enabled and running
2. Ensure the model is available (`docker model ls`)
3. Check Docker Desktop logs for any issues with Model Runner
4. Increase timeout values if model pulling takes too long

## Extending the Tests

You can extend these tests to:

1. Test different models
2. Add performance benchmarks
3. Test error handling and edge cases
4. Test with different configuration parameters
5. Add evaluator logic to validate response quality
