package llm

import "context"

// LLMClient is an interface representing a client for interacting with LLM API.
type LLMClient interface {
	// Chat sends a chat request to the LLM API.
	Chat(ctx context.Context, message string) (string, error)
}

// MockLLMClient is a mock implementation of the LLMClient interface for testing purposes.
type MockLLMClient struct{}

// Chat is a mock implementation of the Chat method.
func (mock MockLLMClient) Chat(ctx context.Context, message string) (string, error) {
	return "", nil
}
