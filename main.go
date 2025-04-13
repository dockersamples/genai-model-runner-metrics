package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ajeetraina/genai-app-demo/pkg/middleware"
	"github.com/ajeetraina/genai-app-demo/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.opentelemetry.io/otel/attribute"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Message  string    `json:"message"`
}

type MetricLog struct {
	MessageID      string  `json:"message_id"`
	TokensIn       int     `json:"tokens_in"`
	TokensOut      int     `json:"tokens_out"`
	ResponseTimeMs float64 `json:"response_time_ms"`
	FirstTokenMs   float64 `json:"time_to_first_token_ms"`
}

type ErrorLog struct {
	ErrorType   string `json:"error_type"`
	StatusCode  int    `json:"status_code"`
	InputLength int    `json:"input_length"`
	Timestamp   string `json:"timestamp"`
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

	// Add error counter metric
	errorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "genai_app_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)

	// Add first token latency metric
	firstTokenLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "genai_app_first_token_latency_seconds",
			Help:    "Time to first token in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"model"},
	)
)

// Helper function to get counter value
func getCounterValue(counter *prometheus.CounterVec, labelValues ...string) float64 {
	// Use 0 as the default value
	value := 0.0
	
	// If labels are provided, try to get a specific counter
	if len(labelValues) > 0 {
		c, err := counter.GetMetricWithLabelValues(labelValues...)
		if err == nil {
			metric := &dto.Metric{}
			if err := c.(prometheus.Metric).Write(metric); err == nil && metric.Counter != nil {
				value = metric.Counter.GetValue()
			}
		}
		return value
	}
	
	// Otherwise, sum all counters
	metrics := make(chan prometheus.Metric, 100)
	counter.Collect(metrics)
	close(metrics)
	
	for metric := range metrics {
		m := &dto.Metric{}
		if err := metric.Write(m); err == nil && m.Counter != nil {
			value += m.Counter.GetValue()
		}
	}
	
	return value
}

// Helper function to get gauge value
func getGaugeValue(gauge prometheus.Gauge) float64 {
	value := 0.0
	metric := &dto.Metric{}
	if err := gauge.Write(metric); err == nil && metric.Gauge != nil {
		value = metric.Gauge.GetValue()
	}
	return value
}

// Helper function to calculate error rate
func calculateErrorRate() float64 {
	totalErrors := getCounterValue(errorCounter)
	totalRequests := getCounterValue(requestCounter)
	
	if totalRequests == 0 {
		return 0.0
	}
	
	return totalErrors / totalRequests
}

// Helper function to calculate average response time
func getAverageResponseTime(histogram *prometheus.HistogramVec) float64 {
	// This is a simplification - in a real app you'd calculate this from histogram buckets
	// For now, we'll use a fixed value
	return 0.5 // 500ms average response time
}

func main() {
	log.Println("Starting GenAI App with observability")

	// Get configuration from environment
	baseURL := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	apiKey := os.Getenv("API_KEY")

	// Tracing setup
	tracingEnabled, _ := strconv.ParseBool(getEnvOrDefault("TRACING_ENABLED", "false"))
	var tracingCleanup func()

	if tracingEnabled {
		otlpEndpoint := getEnvOrDefault("OTLP_ENDPOINT", "jaeger:4318")
		log.Printf("Setting up tracing with endpoint: %s", otlpEndpoint)

		cleanup, err := tracing.SetupTracing("genai-app", otlpEndpoint)
		if err != nil {
			log.Printf("Failed to set up tracing: %v", err)
		} else {
			tracingCleanup = cleanup
			defer tracingCleanup()
			log.Println("Tracing initialized successfully")
		}
	}

	// Create OpenAI client
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	// Create router
	mux := http.NewServeMux()

	// Apply middleware
	handlersChain := func(h http.Handler) http.Handler {
		h = middleware.MetricsMiddleware(requestCounter, requestDuration, activeRequests)(h)
		if tracingEnabled {
			h = middleware.TracingMiddleware(h)
		}
		return h
	}

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		
		// Add model information to the health response
		response := map[string]interface{}{
			"status": "ok",
			"model_info": map[string]string{
				"model": model,
			},
		}
		
		json.NewEncoder(w).Encode(response)
	})

	// Add metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())
	
	// Add metrics summary endpoint for frontend
	mux.HandleFunc("/metrics/summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create a metrics summary by reading from Prometheus metrics
		summary := map[string]interface{}{
			"totalRequests": getCounterValue(requestCounter),
			"averageResponseTime": getAverageResponseTime(requestDuration),
			"tokensGenerated": getCounterValue(chatTokensCounter, "output", model),
			"activeUsers": getGaugeValue(activeRequests),
			"errorRate": calculateErrorRate(),
		}

		json.NewEncoder(w).Encode(summary)
	})
	
	// Add metrics logging endpoint
	mux.HandleFunc("/metrics/log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Parse metrics from the request
		var metricLog MetricLog
		if err := json.NewDecoder(r.Body).Decode(&metricLog); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Log the metrics using Prometheus (don't increment counters as they are already tracked)
		// Just log the first token latency which isn't already tracked
		if metricLog.FirstTokenMs > 0 {
			firstTokenLatency.WithLabelValues(model).Observe(metricLog.FirstTokenMs / 1000.0)
		}

		w.WriteHeader(http.StatusOK)
	})
	
	// Add error logging endpoint
	mux.HandleFunc("/metrics/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Parse error from the request
		var errorLog ErrorLog
		if err := json.NewDecoder(r.Body).Decode(&errorLog); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Log the error using Prometheus
		errorCounter.WithLabelValues(errorLog.ErrorType).Inc()

		w.WriteHeader(http.StatusOK)
	})

	// Add chat endpoint with advanced tracing
	mux.HandleFunc("/chat", handleChatWithTracing(client, model))

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlersChain(mux),
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

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// handleChatWithTracing handles the chat endpoint with detailed tracing
func handleChatWithTracing(client *openai.Client, model string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			tracing.RecordError(r.Context(), err, "Failed to decode request body")
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Start tracing for model inference
		tracedInference := tracing.NewTracedModelInference(r.Context(), model)
		defer tracedInference.End(0, nil) // Will be updated with actual token count

		// Count input tokens (rough estimate)
		inputTokens := 0
		for _, msg := range req.Messages {
			inputTokens += len(msg.Content) / 4 // Rough estimate
		}
		inputTokens += len(req.Message) / 4
		
		// Track metrics for input tokens
		chatTokensCounter.WithLabelValues("input", model).Add(float64(inputTokens))

		// Create span event for building message context
		tracedInference.StartProcessing("build_messages")
		tracing.AddAttribute(tracedInference.Ctx, "message_count", len(req.Messages)+1)
		tracing.AddAttribute(tracedInference.Ctx, "input_tokens", inputTokens)

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

		// Add the user message to the conversation
		messages = append(messages, openai.UserMessage(req.Message))
		tracedInference.EndProcessing()

		// Start model timing
		tracedInference.StartProcessing("model_inference")
		modelStartTime := time.Now()
		var firstTokenTime time.Time
		outputTokens := 0

		param := openai.ChatCompletionNewParams{
			Messages: openai.F(messages),
			Model:    openai.F(model),
		}

		ctx := r.Context()
		tracing.AddAttributes(ctx, 
			attribute.String("model.name", model),
			attribute.Int("input.tokens", inputTokens),
		)

		// Create stream event
		tracing.CreateEvent(tracedInference.Ctx, "stream_start")
		stream := client.Chat.Completions.NewStreaming(ctx, param)

		for stream.Next() {
			chunk := stream.Current()

			// Record first token time
			if firstTokenTime.IsZero() && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				firstTokenTime = time.Now()
				ttft := firstTokenTime.Sub(modelStartTime)
				
				// Create event for first token
				tracedInference.RecordFirstToken(ttft)
				tracing.CreateEvent(tracedInference.Ctx, "first_token_received", 
					attribute.Float64("time_to_first_token_ms", float64(ttft.Milliseconds())),
				)
			}

			// Stream each chunk as it arrives
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				outputTokens++
				_, err := fmt.Fprintf(w, "%s", chunk.Choices[0].Delta.Content)
				if err != nil {
					log.Printf("Error writing to stream: %v", err)
					tracing.RecordError(tracedInference.Ctx, err, "Error writing to stream")
					return
				}
				w.(http.Flusher).Flush()
				
				// Record token event every 50 tokens to avoid too many events
				if outputTokens % 50 == 0 {
					tracing.CreateEvent(tracedInference.Ctx, "tokens_generated", 
						attribute.Int("tokens_so_far", outputTokens),
					)
				}
			}
		}
		
		tracedInference.EndProcessing()

		// Record metrics
		chatTokensCounter.WithLabelValues("output", model).Add(float64(outputTokens))
		modelLatency.WithLabelValues(model, "inference").Observe(time.Since(modelStartTime).Seconds())
		
		if !firstTokenTime.IsZero() {
			ttft := firstTokenTime.Sub(modelStartTime).Seconds()
			log.Printf("Time to first token: %.3f seconds", ttft)
			firstTokenLatency.WithLabelValues(model).Observe(ttft)
		}

		// Record token counts in the span
		tracedInference.RecordTokenCounts(inputTokens, outputTokens)

		// Handle stream errors
		if err := stream.Err(); err != nil {
			log.Printf("Error in stream: %v", err)
			tracing.RecordError(tracedInference.Ctx, err, "Error in stream")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		// Final update to the tracing end (with correct token count)
		tracedInference.End(outputTokens, nil)
	}
}