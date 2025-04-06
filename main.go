package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajeetraina/genai-app-demo/pkg/health"
	"github.com/ajeetraina/genai-app-demo/pkg/logger"
	"github.com/ajeetraina/genai-app-demo/pkg/metrics"
	"github.com/ajeetraina/genai-app-demo/pkg/middleware"
	"github.com/ajeetraina/genai-app-demo/pkg/tracing"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Message  string    `json:"message"`
}

type TokenCounter struct {
	Input  int
	Output int
}

func main() {
	// Initialize structured logging
	logger.Init(getEnvOrDefault("LOG_LEVEL", "info"), getEnvOrDefault("LOG_PRETTY", "true") == "true")
	log.Info().Msg("Starting GenAI App with observability")

	// Initialize tracing if enabled
	if getEnvOrDefault("TRACING_ENABLED", "false") == "true" {
		shutdownTracing, err := tracing.SetupTracing(
			"genai-app",
			getEnvOrDefault("OTLP_ENDPOINT", ""),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize tracing")
		} else {
			defer shutdownTracing()
			log.Info().Msg("Tracing initialized")
		}
	}

	// Get configuration from environment
	baseURL := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	apiKey := os.Getenv("API_KEY")

	// Create OpenAI client
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	// Create router with middleware
	mux := http.NewServeMux()

	// Add CORS handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	// Add health check endpoints
	mux.HandleFunc("/health", health.HandleHealth())
	mux.HandleFunc("/readiness", health.HandleReadiness())

	// Add metrics endpoint
	mux.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	}))

	// Add chat endpoint with rate limiting
	chatHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error().Err(err).Msg("Invalid request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			metrics.ErrorCounter.WithLabelValues("invalid_request", "chat").Inc()
			return
		}

		// Start tracing span for the request
		ctx, endSpan := tracing.StartSpan(r.Context(), "chat_completion")
		defer endSpan()

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Initialize token counter
		tokenCounter := TokenCounter{}

		// Create messages array from request
		var messages []openai.ChatCompletionMessageParamUnion
		for _, msg := range req.Messages {
			var message openai.ChatCompletionMessageParamUnion
			switch msg.Role {
			case "user":
				message = openai.UserMessage(msg.Content)
				tokenCounter.Input += len(msg.Content) / 4 // Rough estimate of tokens
			case "assistant":
				message = openai.AssistantMessage(msg.Content)
				tokenCounter.Output += len(msg.Content) / 4 // Rough estimate of tokens
			}

			messages = append(messages, message)
		}

		// Add the current user message
		messages = append(messages, openai.UserMessage(req.Message))
		tokenCounter.Input += len(req.Message) / 4 // Rough estimate of tokens

		// Start timing for model latency metrics
		startTime := time.Now()
		var firstTokenTime time.Time

		// Create chat completion parameters
		param := openai.ChatCompletionNewParams{
			Messages: openai.F(messages),
			Model:    openai.F(model),
		}

		// Create streaming request
		stream := client.Chat.Completions.NewStreaming(ctx, param)

		// Process the stream
		for stream.Next() {
			chunk := stream.Current()

			// Record first token time if not already set
			if firstTokenTime.IsZero() && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				firstTokenTime = time.Now()
			}

			// Stream each chunk as it arrives
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				token := chunk.Choices[0].Delta.Content
				tokenCounter.Output += 1 // Increment token count
				if _, err := w.Write([]byte(token)); err != nil {
					log.Error().Err(err).Msg("Error writing to stream")
					metrics.ErrorCounter.WithLabelValues("stream_write", "chat").Inc()
					return
				}
				w.(http.Flusher).Flush()
			}
		}

		// Check for errors
		if err := stream.Err(); err != nil {
			log.Error().Err(err).Msg("Error in stream")
			metrics.ErrorCounter.WithLabelValues("stream_error", "chat").Inc()
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Record metrics for the request
		metrics.RecordModelInference(
			model,
			startTime,
			tokenCounter.Input,
			tokenCounter.Output,
			firstTokenTime,
		)

		log.Info().
			Str("model", model).
			Int("input_tokens", tokenCounter.Input).
			Int("output_tokens", tokenCounter.Output).
			Dur("total_duration", time.Since(startTime)).
			Dur("time_to_first_token", firstTokenTime.Sub(startTime)).
			Msg("Chat completion finished")
	})

	// Apply middleware to the chat handler
	rateLimit := middleware.RateLimiter(60) // 60 requests per minute
	wrappedChatHandler := middleware.RequestLogger(rateLimit(chatHandler))
	mux.Handle("/chat", wrappedChatHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
	}

	// Start metrics server on a separate port
	metricsServer := metrics.SetupMetricsServer(":9090")
	go func() {
		log.Info().Str("addr", ":9090").Msg("Starting metrics server")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	// Start the main server
	go func() {
		log.Info().Str("addr", ":8080").Msg("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown servers
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Metrics server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
