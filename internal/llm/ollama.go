package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

	logger.Info("Ollama client initialized", slog.String("component", "ollama client"), slog.String("baseURL", baseURL), slog.String("defaultModel", defaultModel))

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// Chat sends a message to the Ollama API.
func (ollama OllamaClient) Chat(ctx context.Context, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		ollama.logger.Error("message is empty", slog.String("component", "ollama client"))
		return "", fmt.Errorf("ollama client chat: %w", ErrMessageEmpty)
	}

	chatRequest := ChatRequest{
		Model: ollama.defaultModel,
		Messages: &[]ChatMessage{
			{
				Role:    User,
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

	requestCTX, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	clientRequest, err := http.NewRequestWithContext(requestCTX, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		ollama.logger.Error("failed to create HTTP request", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	ollama.logger.Info("sending chat request to Ollama API", slog.String("component", "ollama client"), slog.String("url", ollama.baseURL+"/api/chat"), slog.String("method", http.MethodPost), slog.String("body", string(requestBody)))
	clientResponse, err := ollama.httpClient.Do(clientRequest)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ollama.logger.Error("request to Ollama API timed out", slog.String("component", "ollama client"), slog.String("error", err.Error()))
			return "", fmt.Errorf("ollama client chat: request to Ollama API timed out: %w", err)
		}

		ollama.logger.Error("failed to send request to Ollama API", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}
	defer clientResponse.Body.Close()

	if clientResponse.StatusCode/100 != 2 {
		responseBody, err := io.ReadAll(clientResponse.Body)
		if err != nil {
			ollama.logger.Error("failed to read error response body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
			return "", fmt.Errorf("ollama client chat: received non-2xx response from Ollama API: %d", clientResponse.StatusCode)
		}

		ollama.logger.Error("Ollama API error response", slog.String("component", "ollama client"), slog.Int("status_code", clientResponse.StatusCode), slog.String("response_body", string(responseBody)))
		return "", fmt.Errorf("ollama client chat: received %d response from Ollama API: %q", clientResponse.StatusCode, string(responseBody))
	}

	responseBody, err := io.ReadAll(clientResponse.Body)
	if err != nil {
		ollama.logger.Error("failed to read response body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	ollama.logger.Info("received response from Ollama API", slog.String("component", "ollama client"), slog.String("status code", strconv.Itoa(clientResponse.StatusCode)), slog.String("response", string(responseBody)))

	var chatResponse ChatResponse
	err = json.Unmarshal(responseBody, &chatResponse)
	if err != nil {
		ollama.logger.Error("failed to decode response body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		return "", fmt.Errorf("ollama client chat: %w", err)
	}

	return chatResponse.Message.Content, nil
}
