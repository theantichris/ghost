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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
	Error   string      `json:"error"`
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
		Messages: toChatMessages(messages),
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

	return fromChatMessage(result.Message), nil
}

// ChatStream sends a streaming chat request to Ollama.
func (client *Client) ChatStream(ctx context.Context, model string, messages []conversation.Message, onContent func(string) error) error {
	payload, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: toChatMessages(messages),
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("encode Ollama chat request: %w", err)
	}

	endpoint := client.baseURL.JoinPath("chat")

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint.String(),
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create Ollama chat request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send Ollama chat request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var apiError errorResponse
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, maxErrBodySize)).Decode(&apiError)
		if decodeErr != nil || strings.TrimSpace(apiError.Error) == "" {
			return fmt.Errorf("ollama chat request failed: %s", response.Status)
		}

		return fmt.Errorf("ollama chat request failed: %s: %s", response.Status, apiError.Error)
	}

	decoder := json.NewDecoder(response.Body)
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("consume Ollama chat stream: %w", ctx.Err())
		}

		var frame chatResponse
		if err := decoder.Decode(&frame); err != nil {
			return fmt.Errorf("decode Ollama chat stream: %w", err)
		}

		if frame.Error != "" {
			return fmt.Errorf("Ollama chat stream failed: %s", frame.Error)
		}

		if frame.Message.Content != "" {
			err := onContent(frame.Message.Content)
			if err != nil {
				return fmt.Errorf("deliver Ollama chat content: %w", err)
			}
		}

		if ctx.Err() != nil {
			return fmt.Errorf("consume Ollama chat stream: %w", ctx.Err())
		}

		if frame.Done {
			return nil
		}
	}
}

func toChatMessages(messages []conversation.Message) []chatMessage {
	result := make([]chatMessage, len(messages))

	for index, message := range messages {
		result[index] = chatMessage{Role: string(message.Role), Content: message.Content}
	}

	return result
}

func fromChatMessage(message chatMessage) conversation.Message {
	return conversation.Message{Role: conversation.Role(message.Role), Content: message.Content}
}
