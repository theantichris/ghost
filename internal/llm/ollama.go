package llm

import (
	"fmt"
	"strings"
)

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string // Base Ollama server URL
	defaultModel string // Default model to use
	stream       bool   // Whether to use streaming responses
}

// Chat sends a chat request to the Ollama API.
func (ollama OllamaClient) Chat() {}

// NewOllamaClient initializes a new OllamaClient with the given baseURL and defaultModel.
func NewOllamaClient(baseURL, defaultModel string) (*OllamaClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("ollama client init: %w", ErrURLEmpty)
	}

	if strings.TrimSpace(defaultModel) == "" {
		return nil, fmt.Errorf("ollama client init: %w", ErrModelEmpty)
	}

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		stream:       true,
	}, nil
}
