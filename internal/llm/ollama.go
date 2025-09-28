package llm

import (
	"bufio"
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
		logger.Error("Ollama client initialization failed", "error", ErrURLEmpty, "component", "llm.NewOllamaClient")

		return nil, ErrURLEmpty
	}

	if strings.TrimSpace(defaultModel) == "" {
		logger.Error("Ollama client initialization failed", "error", ErrModelEmpty, "component", "llm.NewOllamaClient")

		return nil, ErrModelEmpty
	}

	logger.Info("Ollama client initialized", "baseURL", baseURL, "defaultModel", defaultModel, "component", "llm.NewOllamaClient")

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

	ollama.logger.Info("sending chat request to Ollama API", "url", ollama.baseURL+"/api/chat", "method", http.MethodPost, "component", "llm.OllamaClient.StreamChat")
	ollama.logger.Debug("request payload", "payload", string(requestBody), "component", "llm.OllamaClient.StreamChat")

	response, err := ollama.httpClient.Do(request)
	if err != nil {
		ollama.logger.Error(ErrClientResponse.Error(), "error", err, "component", "llm.OllamaClient.StreamChat")

		return fmt.Errorf("%w: %s", ErrClientResponse, err)
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
			return fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
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
			return fmt.Errorf("%w: %s", ErrTimeout, err)
		}

		return fmt.Errorf("%w: %s", ErrResponseBody, err)
	}

	return nil
}

// Chat sends a message to the Ollama API.
func (ollama *OllamaClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	requestBody, err := ollama.preparePayload(chatHistory, false)
	if err != nil {
		return ChatMessage{}, err
	}

	clientRequest, cancel, err := ollama.createHTTPRequest(ctx, requestBody)
	if err != nil {
		return ChatMessage{}, err
	}

	defer cancel()

	ollama.logger.Info("sending chat request to Ollama API", "url", ollama.baseURL+"/api/chat", "method", http.MethodPost, "component", "llm.OllamaClient.Chat")
	ollama.logger.Debug("request payload", "payload", string(requestBody), "component", "llm.OllamaClient.Chat")

	clientResponse, err := ollama.httpClient.Do(clientRequest)
	if err != nil {
		ollama.logger.Error(ErrClientResponse.Error(), "error", err, "component", "llm.OllamaClient.Chat")

		return ChatMessage{}, fmt.Errorf("%w: %s", ErrClientResponse, err)
	}

	defer func() {
		if err := clientResponse.Body.Close(); err != nil {
			ollama.logger.Error("failed to close response body", "error", err, "component", "llm.OllamaClient.Chat")
		}
	}()

	err = ollama.checkForHTTPError(clientResponse.StatusCode, clientResponse.Body)
	if err != nil {
		return ChatMessage{}, err
	}

	responseBody, err := io.ReadAll(clientResponse.Body)
	if err != nil {
		ollama.logger.Error(ErrResponseBody.Error(), "error", err, "component", "llm.OllamaClient.Chat")

		return ChatMessage{}, fmt.Errorf("%w: %s", ErrResponseBody, err)
	}

	ollama.logger.Info("received response from Ollama API", "status code", strconv.Itoa(clientResponse.StatusCode), "component", "llm.OllamaClient.Chat")
	ollama.logger.Debug("response payload", "payload", string(responseBody), "component", "llm.OllamaClient.Chat")

	var chatResponse ChatResponse
	err = json.Unmarshal(responseBody, &chatResponse)
	if err != nil {
		ollama.logger.Error(ErrUnmarshalResponse.Error(), "error", err, "component", "llm.OllamaClient.Chat")

		return ChatMessage{}, fmt.Errorf("%w: %s", ErrUnmarshalResponse, err)
	}

	return chatResponse.Message, nil
}

// preparePayload takes the chat history and returns the marshaled request body.
func (ollama *OllamaClient) preparePayload(chatHistory []ChatMessage, stream bool) ([]byte, error) {
	if len(chatHistory) == 0 {
		return nil, ErrChatHistoryEmpty
	}

	chatRequest := ChatRequest{
		Model:    ollama.defaultModel,
		Messages: chatHistory,
		Stream:   stream,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMarshalRequest, err)
	}

	return requestBody, nil
}

// createHTTPRequest creates the HTTP request with timeout and headers.
func (ollama *OllamaClient) createHTTPRequest(ctx context.Context, requestBody []byte) (*http.Request, context.CancelFunc, error) {
	requestCTX, cancel := context.WithTimeout(ctx, 2*time.Minute)

	clientRequest, err := http.NewRequestWithContext(requestCTX, http.MethodPost, ollama.baseURL+"/api/chat", bytes.NewReader(requestBody))
	if err != nil {
		cancel()

		return nil, nil, fmt.Errorf("%w: %s", ErrCreateRequest, err)
	}

	clientRequest.Header.Set("Content-Type", "application/json")

	return clientRequest, cancel, nil
}

// checkForHTTPError returns the correct error for the HTTP status code.
func (ollama *OllamaClient) checkForHTTPError(statusCode int, body io.ReadCloser) error {
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
