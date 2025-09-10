package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// Role defines the role of a message in the chat.
type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Tool      Role = "tool"
)

// ChatRequest represents a request to the Ollama chat API.
type ChatRequest struct {
	Model    string         `json:"model"`    // Required. The model name.
	Messages *[]ChatMessage `json:"messages"` // The messages of the chat, this can be used to keep a chat memory
	Stream   bool           `json:"stream"`   // Whether to stream the response
}

// ChatMessage represents a single message in the chat.
type ChatMessage struct {
	Role    Role   `json:"role"`    // The role of the message, either system, user, assistant, or tool
	Content string `json:"content"` // The content of the message
}

// ChatResponse represents a response from the Ollama chat API.
type ChatResponse struct {
	Message ChatMessage `json:"message"` // The response message from the assistant
}

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string       // Base Ollama server URL
	defaultModel string       // Default model to use
	httpClient   *http.Client // HTTP client for making requests
	logger       *slog.Logger // Logger for structured logging
}

// NewOllamaClient initializes a new OllamaClient with the given baseURL and defaultModel.
func NewOllamaClient(baseURL, defaultModel string, httpClient *http.Client, logger *slog.Logger) (*OllamaClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		logger.Error("baseURL is empty", slog.String("component", "ollama client"))
		return nil, fmt.Errorf("ollama client init: %w", ErrURLEmpty)
	}

	if strings.TrimSpace(defaultModel) == "" {
		logger.Error("defaultModel is empty", slog.String("component", "ollama client"))
		return nil, fmt.Errorf("ollama client init: %w", ErrModelEmpty)
	}

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// Chat sends a message to the Ollama API.
func (ollama OllamaClient) Chat(ctx context.Context, message string) (string, error) {
	// TODO: Info logging
	// Validate message to LLM
	if strings.TrimSpace(message) == "" {
		ollama.logger.Error("message is empty", slog.String("component", "ollama client"))
		return "", fmt.Errorf("ollama client chat: %w", ErrMessageEmpty)
	}

	// Create request payload
	chatRequest := ChatRequest{
		Model: ollama.defaultModel,
		Messages: &[]ChatMessage{
			{
				Role:    System,
				Content: message,
			},
		},
		Stream: false,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		ollama.logger.Error("failed to marshal request body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	// Send request to Ollama API
	clientRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		ollama.logger.Error("failed to create HTTP request", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	// Handle response from Ollama API
	clientResponse, err := ollama.httpClient.Do(clientRequest)
	if err != nil {
		if err == context.DeadlineExceeded {
			ollama.logger.Error("request to Ollama API timed out", slog.String("component", "ollama client"), slog.String("error", err.Error()))
			return "", fmt.Errorf("ollama client chat: request to Ollama API timed out: %w", err)
		}

		ollama.logger.Error("failed to send request to Ollama API", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	if clientResponse.StatusCode/100 != 2 {
		ollama.logger.Error("received non-2xx response from Ollama API", slog.String("component", "ollama client"), slog.Int("status_code", clientResponse.StatusCode))
		return "", fmt.Errorf("ollama client chat: received non-2xx response from Ollama API: %d", clientResponse.StatusCode)
	}

	// Get message.content from response
	var chatResponse ChatResponse
	err = json.NewDecoder(clientResponse.Body).Decode(&chatResponse)
	if err != nil {
		ollama.logger.Error("failed to decode response body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	// Return response or error
	return chatResponse.Message.Content, nil
}
