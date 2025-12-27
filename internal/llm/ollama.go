package llm

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/requests"
)

const host = "http://localhost:11434/api"

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

// Chat sends a request to the chat endpoint and returns the response message
// content.
func Chat(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	request := ChatRequest{
		Model:    model,
		Stream:   false,
		Messages: messages,
	}

	var chatResponse ChatResponse

	err := requests.
		URL(host + "/chat").
		BodyJSON(&request).
		ToJSON(&chatResponse).
		Fetch(ctx)

	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return chatResponse.Message.Content, nil
}
