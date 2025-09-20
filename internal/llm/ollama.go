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

// Chat sends a message to the Ollama API.
func (ollama OllamaClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	requestBody, err := ollama.preparePayload(chatHistory)
	if err != nil {
		return ChatMessage{}, err
	}

	requestCTX, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	clientRequest, err := ollama.createHTTPRequest(requestCTX, requestBody)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrCreateRequest, err)
	}

	ollama.logger.Info("sending chat request to Ollama API", slog.String("component", "ollama client"), slog.String("url", ollama.baseURL+"/api/chat"), slog.String("method", http.MethodPost))
	ollama.logger.Debug("request payload", slog.String("component", "ollama client"), slog.String("payload", string(requestBody)))

	clientResponse, err := ollama.httpClient.Do(clientRequest)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w: %s", ErrClientResponse, err)
	}

	defer func() {
		if err := clientResponse.Body.Close(); err != nil {
			ollama.logger.Error("failed to close response body", slog.String("component", "ollama client"), slog.String("error", err.Error()))
		}
	}()

	err = ollama.checkForHTTPError(clientResponse.StatusCode, clientRequest.Body)
	if err != nil {
		return ChatMessage{}, err
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

// preparePayload takes the chat history and returns the marshaled request body.
func (ollama OllamaClient) preparePayload(chatHistory []ChatMessage) ([]byte, error) {
	if len(chatHistory) == 0 {
		return nil, ErrChatHistoryEmpty
	}

	chatRequest := ChatRequest{
		Model:    ollama.defaultModel,
		Messages: chatHistory,
		Stream:   false,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMarshalRequest, err)
	}

	return requestBody, nil
}

// createHTTPRequest creates the HTTP request with timeout and headers.
func (ollama OllamaClient) createHTTPRequest(ctx context.Context, requestBody []byte) (*http.Request, error) {
	clientRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateRequest, err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	return clientRequest, nil
}

// checkForHTTPError returns the correct error for the HTTP status code.
func (ollama OllamaClient) checkForHTTPError(statusCode int, body io.ReadCloser) error {
	if statusCode/100 != 2 {
		responseBody, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("%w: status=%d %s: %s", ErrResponseBody, statusCode, http.StatusText(statusCode), err)
		}

		var apiError apiError
		if err := json.Unmarshal(responseBody, &apiError); err == nil && apiError.Error != "" {
			return fmt.Errorf("%w: status=%d %s api_error=%s", ErrNon2xxResponse, statusCode, http.StatusText(statusCode), apiError.Error)
		}

		return fmt.Errorf("%w: status=%d %s body=%s", ErrNon2xxResponse, statusCode, http.StatusText(statusCode), string(responseBody))
	}

	return nil
}
