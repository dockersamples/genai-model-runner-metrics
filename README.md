# GenAI Application Demo

A modern chat application demonstrating integration of frontend technologies with local Large Language Models (LLMs).

## Overview

This project is a full-stack GenAI chat application that showcases how to build a Generative AI interface with a React frontend and Go backend, connected to Llama 3.2 running on Ollama.

## Architecture

The application consists of three main components:

1. **Frontend**: React TypeScript application providing a responsive chat interface
2. **Backend**: Go server that handles API requests and connects to the LLM
3. **Model Runner**: Ollama service running the Llama 3.2 (1B parameter) model

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │ >>> │   Backend   │ >>> │    Ollama   │
│  (React/TS) │     │    (Go)     │     │  (Llama 3.2)│
└─────────────┘     └─────────────┘     └─────────────┘
      :3000              :8080              :11434
```

## Features

- Real-time chat interface with message history
- Streaming AI responses (tokens appear as they're generated)
- Dockerized deployment for easy setup
- Local LLM integration (no cloud API dependencies)
- Cross-origin resource sharing (CORS) enabled

## Prerequisites

- Docker and Docker Compose
- Git

## Quick Start

1. Clone this repository:
   ```bash
   git clone https://github.com/ajeetraina/genai-app-demo.git
   cd genai-app-demo
   ```

2. Start the application using Docker Compose:
   ```bash
   docker compose -f compose.yaml -f ollama-ci.yaml up
   ```

   This command combines both files to create a complete deployment with all three components:

    - The frontend React app
    - The backend Go server
    - The Ollama LLM service

There's also a third compose file called compose-ci.yaml which appears to be a simplified version possibly used for continuous integration scenarios.

3. Access the frontend at [http://localhost:3000](http://localhost:3000)

> Please Note: There are two compose files in the repository:
> 1. compose.yaml: This is the main Docker Compose file that sets up the core services:
>     - backend service: The Go API server
>     - frontend service: The React web application

> 2. ollama-ci.yaml: This is a separate compose file specifically for setting up the Ollama service which runs the LLM (Llama 3.2 1B model).

> These files are designed to be used together with Docker Compose's ability to merge multiple compose files. 

## Development Setup

### Frontend

The frontend is a React TypeScript application using Vite:

```bash
cd frontend
npm install
npm run dev
```

### Backend

The Go backend can be run directly:

```bash
go mod download
go run main.go
```

Make sure to set the environment variables in `backend.env` or provide them directly.

## Configuration

The backend connects to the LLM service using environment variables defined in `backend.env`:

- `BASE_URL`: URL for the model runner
- `MODEL`: Model identifier to use
- `API_KEY`: API key for authentication (defaults to "ollama")

## Deployment

The application is configured for easy deployment using Docker Compose. See the `compose.yaml` and `ollama-ci.yaml` files for details.

## License

[Add appropriate license information here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
