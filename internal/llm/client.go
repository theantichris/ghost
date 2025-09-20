package llm

import (
	"context"
	"fmt"
)

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// Chat sends a chat request to the LLM API.
	Chat(ctx context.Context, message string) (string, error)
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct {
	ReturnError bool
}

// Chat is a mock implementation of the Chat method.
func (mock MockLLMClient) Chat(ctx context.Context, message string) (string, error) {
	if mock.ReturnError {
		return "", fmt.Errorf("ollama client chat: %w", ErrChatHistoryEmpty)
	}

	return "", nil
}
