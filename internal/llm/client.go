package llm

import (
	"context"
)

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// StreamChat sends a message to the LLM API and streams the response.
	StreamChat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct {
	Content        string
	Error          error
	StreamChatFunc func(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
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
