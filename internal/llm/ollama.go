package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string       // Base Ollama server URL
	defaultModel string       // Default model to use
	httpClient   *http.Client // HTTP client for making requests
	logger       *log.Logger  // Logger for structured logging
}

// NewOllamaClient initializes a new OllamaClient with the given baseURL and defaultModel.
func NewOllamaClient(baseURL, defaultModel string, httpClient *http.Client, logger *log.Logger) (*OllamaClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		logger.Error("Ollama client initialization failed", "error", ErrValidation)

		return nil, fmt.Errorf("%w: base URL is empty", ErrValidation)
	}

	if strings.TrimSpace(defaultModel) == "" {
		logger.Error("Ollama client initialization failed", "error", ErrValidation)

		return nil, fmt.Errorf("%w: model name is empty", ErrValidation)
	}

	logger.Info("Ollama client initialized", "baseURL", baseURL, "defaultModel", defaultModel)

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// StreamChat sends a message to the Ollama API and streams the response.
func (ollama *OllamaClient) StreamChat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error {
	requestBody, err := ollama.preparePayload(chatHistory, true)
	if err != nil {
		return err
	}

	request, cancel, err := ollama.createHTTPRequest(ctx, requestBody)
	if err != nil {
		return err
	}

	defer cancel()

	ollama.logger.Info("sending chat request to Ollama API", "url", ollama.baseURL+"/api/chat", "method", http.MethodPost, "queryLength", len(string(requestBody)))

	response, err := ollama.httpClient.Do(request)
	if err != nil {
		ollama.logger.Error(ErrResponse.Error(), "error", err)

		return fmt.Errorf("%w: %w", ErrResponse, err)
	}

	err = ollama.checkForHTTPError(response.StatusCode, response.Body)
	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	scanner := bufio.NewScanner(response.Body)
	const maxTokenBytes = 1024 * 1024
	buffer := make([]byte, 0, 64+1024)
	scanner.Buffer(buffer, maxTokenBytes)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		var chunk ChatResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("%w: %w", ErrResponse, err)
		}

		if chunk.Message.Content != "" && onToken != nil {
			onToken(chunk.Message.Content)
		}

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: timeout: %w", ErrResponse, err)
		}

		return fmt.Errorf("%w: %w", ErrResponse, err)
	}

	return nil
}

// preparePayload takes the chat history and returns the marshaled request body.
func (ollama *OllamaClient) preparePayload(chatHistory []ChatMessage, stream bool) ([]byte, error) {
	if len(chatHistory) == 0 {
		return nil, fmt.Errorf("%w: chat history is empty", ErrValidation)
	}

	chatRequest := ChatRequest{
		Model:    ollama.defaultModel,
		Messages: chatHistory,
		Stream:   stream,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	return requestBody, nil
}

// createHTTPRequest creates the HTTP request with timeout and headers.
func (ollama *OllamaClient) createHTTPRequest(ctx context.Context, requestBody []byte) (*http.Request, context.CancelFunc, error) {
	requestCTX, cancel := context.WithTimeout(ctx, 2*time.Minute)

	clientRequest, err := http.NewRequestWithContext(requestCTX, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		cancel()

		return nil, nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	return clientRequest, cancel, nil
}

// checkForHTTPError returns the correct error for the HTTP status code.
func (ollama *OllamaClient) checkForHTTPError(statusCode int, body io.ReadCloser) error {
	if statusCode/100 != 2 {
		responseBody, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("%w: status=%d %s: %w", ErrResponse, statusCode, http.StatusText(statusCode), err)
		}

		var apiError apiError
		if err := json.Unmarshal(responseBody, &apiError); err == nil && apiError.Error != "" {
			return fmt.Errorf("%w: status=%d %s api_error=%s", ErrResponse, statusCode, http.StatusText(statusCode), apiError.Error)
		}

		return fmt.Errorf("%w: status=%d %s body=%s", ErrResponse, statusCode, http.StatusText(statusCode), string(responseBody))
	}

	return nil
}
