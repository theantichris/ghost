package llm

import (
	"context"
)

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// Chat sends a chat request to the LLM API.
	Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error)
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct {
	Error error
}

// Chat is a mock implementation of the Chat method.
func (mock MockLLMClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	if mock.Error != nil {
		return ChatMessage{}, mock.Error
	}

	return ChatMessage{}, nil
}
