package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
)

var (
	ErrModelNotFound    = errors.New("model not found")
	ErrUnexpectedStatus = errors.New("unexpected status")
	ErrDecodeChunk      = errors.New("error decoding chunk")
)

// ChatRequest holds the information for the chat endpoint.
type ChatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse holds the response from the chat endpoint.
type ChatResponse struct {
	Message ChatMessage `json:"message"`
}

// ChatMessage holds a single message in the chat history.
type ChatMessage struct {
	// Role holds the author of the message.
	// Values are system, user, assistant, tool.
	Role string `json:"role"`

	// Content holds the message content.
	Content string `json:"content"`
}

// Chat sends a request to the chat endpoint and returns the response message.
// onChunk is called for each streamed chunk of content.
func Chat(ctx context.Context, host, model string, messages []ChatMessage, onChunk func(string)) (ChatMessage, error) {
	request := ChatRequest{
		Model:    model,
		Stream:   true,
		Messages: messages,
	}

	var chatContent strings.Builder

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		AddValidator(nil).
		Handle(func(response *http.Response) error {
			defer func() {
				_ = response.Body.Close()
			}()

			if response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("%w: %s", ErrModelNotFound, request.Model)
			}

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("%w: %s", ErrUnexpectedStatus, response.Status)
			}

			decoder := json.NewDecoder(response.Body)

			for {
				var chunk ChatResponse

				if err := decoder.Decode(&chunk); err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("%w: %w", ErrDecodeChunk, err)
				}

				onChunk(chunk.Message.Content)

				chatContent.WriteString(chunk.Message.Content)
			}

			return nil
		}).
		Fetch(ctx)

	if err != nil {
		return ChatMessage{}, fmt.Errorf("%w", err)
	}

	chatMessage := ChatMessage{
		Role:    "assistant",
		Content: chatContent.String(),
	}

	return chatMessage, nil
}
