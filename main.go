package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Message  string    `json:"message"`
}

// Define metrics
var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "genai_app_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "genai_app_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	chatTokensCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "genai_app_chat_tokens_total",
			Help: "Total number of tokens processed in chat",
		},
		[]string{"direction", "model"},
	)
	
	modelLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "genai_app_model_latency_seconds",
			Help:    "Model response time in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 30, 60},
		},
		[]string{"model", "operation"},
	)
	
	activeRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "genai_app_active_requests",
			Help: "Number of currently active requests",
		},
	)
)

func main() {
	log.Println("Starting GenAI App with observability")

	// Get configuration from environment
	baseURL := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	apiKey := os.Getenv("API_KEY")

	// Create OpenAI client
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	// Create router
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

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Add metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Add chat endpoint
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeRequests.Inc()
		defer activeRequests.Dec()
		
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
			log.Printf("Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			requestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", http.StatusBadRequest)).Inc()
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := r.Context()

		// Count input tokens (rough estimate)
		inputTokens := 0
		for _, msg := range req.Messages {
			inputTokens += len(msg.Content) / 4 // Rough estimate
		}
		inputTokens += len(req.Message) / 4
		
		// Track metrics for input tokens
		chatTokensCounter.WithLabelValues("input", model).Add(float64(inputTokens))

		var messages []openai.ChatCompletionMessageParamUnion
		for _, msg := range req.Messages {
			var message openai.ChatCompletionMessageParamUnion
			switch msg.Role {
			case "user":
				message = openai.UserMessage(msg.Content)
			case "assistant":
				message = openai.AssistantMessage(msg.Content)
			}

			messages = append(messages, message)
		}

		// Start model timing
		modelStartTime := time.Now()
		var firstTokenTime time.Time
		outputTokens := 0

		// Adds the user message to the conversation
		messages = append(messages, openai.UserMessage(req.Message))
		
		param := openai.ChatCompletionNewParams{
			Messages: openai.F(messages),
			Model:    openai.F(model),
		}

		stream := client.Chat.Completions.NewStreaming(ctx, param)

		for stream.Next() {
			chunk := stream.Current()

			// Record first token time
			if firstTokenTime.IsZero() && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				firstTokenTime = time.Now()
			}

			// Stream each chunk as it arrives
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				outputTokens++
				_, err := fmt.Fprintf(w, "%s", chunk.Choices[0].Delta.Content)
				if err != nil {
					log.Printf("Error writing to stream: %v", err)
					return
				}
				w.(http.Flusher).Flush()
			}
		}

		// Record metrics
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
		requestCounter.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
		chatTokensCounter.WithLabelValues("output", model).Add(float64(outputTokens))
		modelLatency.WithLabelValues(model, "inference").Observe(time.Since(modelStartTime).Seconds())
		
		if !firstTokenTime.IsZero() {
			ttft := firstTokenTime.Sub(modelStartTime).Seconds()
			log.Printf("Time to first token: %.3f seconds", ttft)
		}

		if err := stream.Err(); err != nil {
			log.Printf("Error in stream: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
	}

	// Start metrics server on a separate port
	metricsServer := &http.Server{
		Addr:    ":9090",
		Handler: promhttp.Handler(),
	}
	
	go func() {
		log.Println("Starting metrics server on :9090")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Start the main server
	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown servers
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Fatalf("Metrics server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
