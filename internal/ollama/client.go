package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/theantichris/ghost/v4/internal/conversation"
)

const DefaultURL = "http://localhost:11434/api"
const maxErrBodySize = 64 << 10

// Client communicates with the Ollama HTTP API.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type chatRequest struct {
	Model    string                 `json:"model"`
	Messages []conversation.Message `json:"messages"`
	Stream   bool                   `json:"stream"`
}

type chatResponse struct {
	Message conversation.Message `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewClient constructs an Ollama API client.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse Ollama base URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("parse Ollama base URL: absolute URL required")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    parsedURL,
		httpClient: httpClient,
	}, nil
}

// Chat sends a non-streaming chat request to Ollama.
func (client *Client) Chat(ctx context.Context, model string, messages []conversation.Message) (conversation.Message, error) {
	payload, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return conversation.Message{}, fmt.Errorf("encode Ollama chat request: %w", err)
	}

	endpoint := client.baseURL.JoinPath("chat")

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.String(),
		bytes.NewReader(payload),
	)
	if err != nil {
		return conversation.Message{}, fmt.Errorf("create Ollama chat request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return conversation.Message{}, fmt.Errorf("send Ollama chat request: %w", err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var apiError errorResponse
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, maxErrBodySize)).Decode(&apiError)

		if decodeErr != nil || strings.TrimSpace(apiError.Error) == "" {
			return conversation.Message{}, fmt.Errorf("ollama chat request failed: %s", response.Status)
		}

		return conversation.Message{}, fmt.Errorf("ollama chat request failed: %s: %s", response.Status, apiError.Error)
	}

	var result chatResponse

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return conversation.Message{}, fmt.Errorf("decode Ollama chat response: %w", err)
	}

	return result.Message, nil
}
