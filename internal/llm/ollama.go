package llm

import (
	"fmt"
	"log/slog"
	"strings"
)

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string       // Base Ollama server URL
	defaultModel string       // Default model to use
	stream       bool         // Whether to use streaming responses
	logger       *slog.Logger // Logger for structured logging
}

// Chat sends a chat request to the Ollama API.
func (ollama OllamaClient) Chat() {}

// NewOllamaClient initializes a new OllamaClient with the given baseURL and defaultModel.
func NewOllamaClient(baseURL, defaultModel string, logger *slog.Logger) (*OllamaClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		logger.Error("baseURL is empty", slog.String("component", "ollama client"))

		return nil, fmt.Errorf("ollama client init: %w", ErrURLEmpty)
	}

	if strings.TrimSpace(defaultModel) == "" {
		logger.Error("defaultModel is empty", slog.String("component", "ollama client"))

		return nil, fmt.Errorf("ollama client init: %w", ErrModelEmpty)
	}

	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		stream:       true,
		logger:       logger,
	}, nil
}
