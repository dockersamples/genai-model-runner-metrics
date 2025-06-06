#!/bin/bash

# Script to run the Docker Model Runner direct integration test
# Author: Test Team
# Created: March 2025

echo "======================================================================"
echo "   Testing Direct Integration with Docker Model Runner                "
echo "======================================================================"
echo ""
echo "This test demonstrates direct communication with Docker Model Runner"
echo "It will:"
echo "  1. Create a Socat container to connect to Model Runner"
echo "  2. Send prompts directly to the Llama3.2 model"
echo "  3. Verify responses for both basic text and structured outputs"
echo ""
echo "Prerequisites:"
echo "  - Docker Desktop with Model Runner enabled"
echo "  - Host-side TCP support enabled (port 12434)"
echo "  - Llama3.2 model pulled (ignaciolopezluna020/llama3.2:1B)"
echo ""

# Navigate to the tests directory
cd "$(dirname "$0")" || { echo "Failed to change to script directory"; exit 1; }

# Check Docker Model Runner availability
echo "Checking Docker Model Runner availability..."
if ! docker model ls >/dev/null 2>&1; then
  echo "⚠️  Warning: Docker Model Runner command not found."
  echo "    Please ensure Docker Desktop is running with Model Runner enabled."
  echo "    Test will continue but may fail if Model Runner isn't available."
fi

# Check if the model is already pulled
echo "Checking for required model..."
if docker model ls | grep -q "ignaciolopezluna020/llama3.2:1B"; then
  echo "✅ Model ignaciolopezluna020/llama3.2:1B is already available"
else
  echo "⚠️  Model ignaciolopezluna020/llama3.2:1B not found locally."
  echo "   Test will attempt to pull it if needed."
fi

echo ""
echo "Starting test now..."
echo "======================================================================"
# Run the test with verbose output and generous timeout
go test -v -timeout 2m -run TestModelRunnerIntegration ./integration

# Check the exit code
if [ $? -eq 0 ]; then
  echo ""
  echo "======================================================================"
  echo "✅ Test PASSED: Successfully tested direct integration with Docker Model Runner"
  echo "======================================================================"
else
  echo ""
  echo "======================================================================"
  echo "❌ Test FAILED: See error details above"
  echo "======================================================================"
fi

# Clean up any leftover containers
echo "Cleaning up any test containers..."
docker ps -a | grep 'testcontainers' | awk '{print $1}' | xargs -r docker rm -f >/dev/null 2>&1
