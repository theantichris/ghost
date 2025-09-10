package llm

import (
	"fmt"
	"strings"
)

// Define ollama client struct
type OllamaClient struct {
	baseURL      string // Base Ollama server URL
	defaultModel string // Default model to use
}

// Instantiate ollama llm client
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
	}, nil
}
