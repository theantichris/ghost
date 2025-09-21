package llm

import (
	"context"
)

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// Chat sends a message to the LLM API and returns the response.
	Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error)
	// StreamChat sends a message to the LLM API and streams the response.
	StreamChat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct {
	Error          error
	StreamChatFunc func(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
}

// Chat is a mock implementation of the Chat method.
func (mock *MockLLMClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	if mock.Error != nil {
		return ChatMessage{}, mock.Error
	}

	return ChatMessage{}, nil
}

// StreamChat is a mock implementation of the StreamChat method.
func (mock *MockLLMClient) StreamChat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error {
	if mock.Error != nil {
		return mock.Error
	}

	if mock.StreamChatFunc != nil {
		return mock.StreamChatFunc(ctx, chatHistory, onToken)
	}

	return nil
}
