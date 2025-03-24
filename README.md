# GenAI Application Demo using Model Runner

A modern chat application demonstrating integration of frontend technologies with local Large Language Models (LLMs).

## Overview

This project is a full-stack GenAI chat application that showcases how to build a Generative AI interface with a React frontend and Go backend, connected to the Docker Model Runner.

## Two Methods

There are two ways you can use Model Runner:

- Using Internal DNS
- Using TCP


### Using Internal DNS

This methods points to the same Model Runner (llama.cpp engine) but through different connection method. 
It uses the internal Docker DNS resolution (model-runner.docker.internal)



#### Architecture

The application consists of three main components:

1. **Frontend**: React TypeScript application providing a responsive chat interface
2. **Backend**: Go server that handles API requests and connects to the LLM
3. **Model Runner**: Llama 3.2 (1B parameter) model

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │ >>> │   Backend   │ >>> │ Model Runner│
│  (React/TS) │     │    (Go)     │     │ 
└─────────────┘     └─────────────┘     └─────────────┘
      :3000              :8080              
```

##### Features

- Real-time chat interface with message history
- Streaming AI responses (tokens appear as they're generated)
- Dockerized deployment for easy setup
- Local LLM integration (no cloud API dependencies)
- Cross-origin resource sharing (CORS) enabled
- Comprehensive integration tests using Testcontainers

##### Prerequisites

- Docker and Docker Compose
- Git
- Go 1.19 or higher
- Download the model before proceeding further

```
docker model pull ignaciolopezluna020/llama3.2:1B
```

##### Quick Start

1. Clone this repository:
   ```bash
   git clone https://github.com/ajeetraina/genai-app-demo.git
   cd genai-app-demo

   ```

2. Start the application using Docker Compose:
   ```bash
   docker compose up -d -build
   ```

3. Access the frontend at [http://localhost:3000](http://localhost:3000)

##### Development Setup

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


## Using TCP 

This menthods points to the same Model Runner (`llama.cpp engine`) but through different connection method. 
It uses the host-side TCP support via `host.docker.internal:12434`


## Configuration

The backend connects to the LLM service using environment variables defined in `backend.env`:

- `BASE_URL`: URL for the model runner
- `MODEL`: Model identifier to use
- `API_KEY`: API key for authentication 

## Deployment

The application is configured for easy deployment using Docker Compose. See the `compose.yaml` and `ollama-ci.yaml` files for details.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
