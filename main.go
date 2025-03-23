package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Message  string    `json:"message"`
}

// ModelRequest represents a request to the model
type ModelRequest struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

// ModelResponse represents a response from the model
type ModelResponse struct {
	Choices []struct {
		Text         string `json:"text,omitempty"`
		Delta        struct {
			Content string `json:"content"`
		} `json:"delta,omitempty"`
	} `json:"choices"`
}

func main() {
	baseURL := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// If we have a message, add it to the messages
		if req.Message != "" {
			req.Messages = append(req.Messages, Message{
				Role:    "user",
				Content: req.Message,
			})
		}

		// Create the model request
		modelReq := ModelRequest{
			Model:       model,
			Messages:    req.Messages,
			Temperature: 0.2,
			MaxTokens:   500,
			Stream:      true,
		}

		// Convert request to JSON
		jsonData, err := json.Marshal(modelReq)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Send request to model
		modelURL := fmt.Sprintf("%s/completions", baseURL)
		client := &http.Client{}
		modelResp, err := client.Post(
			modelURL,
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			http.Error(w, "Error communicating with model", http.StatusInternalServerError)
			return
		}
		defer modelResp.Body.Close()

		// Process the streaming response
		decoder := json.NewDecoder(modelResp.Body)
		for {
			var response ModelResponse
			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("Error decoding response: %v\n", err)
				break
			}

			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				content := response.Choices[0].Delta.Content
				_, err := fmt.Fprintf(w, "%s", content)
				if err != nil {
					fmt.Printf("Error writing to response: %v\n", err)
					break
				}
				w.(http.Flusher).Flush()
			} else if len(response.Choices) > 0 && response.Choices[0].Text != "" {
				// For non-streaming responses
				_, err := fmt.Fprintf(w, "%s", response.Choices[0].Text)
				if err != nil {
					fmt.Printf("Error writing to response: %v\n", err)
					break
				}
				w.(http.Flusher).Flush()
			}
		}
	})

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
