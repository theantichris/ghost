package llm

import (
	"context"
)

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// Chat sends a message to the LLM API and streams the response.
	Chat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct {
	Content  string
	Error    error
	ChatFunc func(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error
}

// Chat is a mock implementation of the Chat method.
func (mock *MockLLMClient) Chat(ctx context.Context, chatHistory []ChatMessage, onToken func(string)) error {
	if mock.Error != nil {
		return mock.Error
	}

	if mock.ChatFunc != nil {
		return mock.ChatFunc(ctx, chatHistory, onToken)
	}

	return nil
}
