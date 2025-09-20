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
	ReturnError                bool
	ReturnErrClientResponse    bool
	ReturnErrNon2xxResponse    bool
	ReturnErrResponseBody      bool
	ReturnErrUnmarshalResponse bool
}

// Chat is a mock implementation of the Chat method.
func (mock MockLLMClient) Chat(ctx context.Context, chatHistory []ChatMessage) (ChatMessage, error) {
	if mock.ReturnError {
		return ChatMessage{}, ErrChatHistoryEmpty
	}

	if mock.ReturnErrClientResponse {
		return ChatMessage{}, ErrClientResponse
	}

	if mock.ReturnErrNon2xxResponse {
		return ChatMessage{}, ErrNon2xxResponse
	}

	if mock.ReturnErrResponseBody {
		return ChatMessage{}, ErrResponseBody
	}

	if mock.ReturnErrUnmarshalResponse {
		return ChatMessage{}, ErrUnmarshalResponse
	}

	return ChatMessage{}, nil
}
