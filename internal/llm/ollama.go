package llm

import (
	"bytes"
	"context"
	"encoding/json"
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
		return nil, ErrURLEmpty
	}

	if strings.TrimSpace(defaultModel) == "" {
		return nil, ErrModelEmpty
	}

	logger.Info("Ollama client initialized", slog.String("component", "ollama client"), slog.String("baseURL", baseURL), slog.String("defaultModel", defaultModel))

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// TODO: Can some error messages be returned as if from the API and recovered?

// Chat sends a message to the Ollama API.
func (ollama OllamaClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	if len(chatHistory) == 0 {
		return ChatMessage{}, ErrChatHistoryEmpty
	}

	chatRequest := ChatRequest{
		Model:    ollama.defaultModel,
		Messages: chatHistory,
		Stream:   false,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrMarshalRequest, err)
	}

	requestCTX, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	clientRequest, err := http.NewRequestWithContext(requestCTX, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrCreateRequest, err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	ollama.logger.Info("sending chat request to Ollama API", slog.String("component", "ollama client"), slog.String("url", ollama.baseURL+"/api/chat"), slog.String("method", http.MethodPost))
	ollama.logger.Debug("request payload", slog.String("component", "ollama client"), slog.String("payload", string(requestBody)))

	clientResponse, err := ollama.httpClient.Do(clientRequest)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrClientResponse, err)
	}

	defer clientResponse.Body.Close()

	if clientResponse.StatusCode/100 != 2 {
		responseBody, err := io.ReadAll(clientResponse.Body)
		if err != nil {
			return ChatMessage{}, fmt.Errorf("%w: status=%d %s: %s", ErrResponseBody, clientResponse.StatusCode, http.StatusText(clientResponse.StatusCode), err)
		}

		var apiError apiError
		if err := json.Unmarshal(responseBody, &apiError); err == nil && apiError.Error != "" {
			return ChatMessage{}, fmt.Errorf("%w: status=%d %s api_error=%s", ErrNon2xxResponse, clientResponse.StatusCode, http.StatusText(clientResponse.StatusCode), apiError.Error)
		}

		return ChatMessage{}, fmt.Errorf("%w: status=%d %s body=%s", ErrNon2xxResponse, clientResponse.StatusCode, http.StatusText(clientResponse.StatusCode), string(responseBody))
	}

	responseBody, err := io.ReadAll(clientResponse.Body)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrResponseBody, err)
	}

	ollama.logger.Info("received response from Ollama API", slog.String("component", "ollama client"), slog.String("status code", strconv.Itoa(clientResponse.StatusCode)))
	ollama.logger.Debug("response payload", slog.String("component", "ollama client"), slog.String("payload", string(responseBody)))

	var chatResponse ChatResponse
	err = json.Unmarshal(responseBody, &chatResponse)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
	}

	return chatResponse.Message, nil
}
