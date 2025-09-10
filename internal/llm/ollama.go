package llm

import (
	"fmt"
	"log/slog"
	"strings"
)

// Role defines the role of a message in the chat.
type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Tool      Role = "tool"
)

// ChatRequest represents a request to the Ollama chat API.
type ChatRequest struct {
	Model    string         `json:"model"`    // Required. The model name.
	Messages *[]ChatMessage `json:"messages"` // The messages of the chat, this can be used to keep a chat memory
}

// ChatMessage represents a single message in the chat.
type ChatMessage struct {
	Role    Role   `json:"role"`    // The role of the message, either system, user, assistant, or tool
	Content string `json:"content"` // The content of the message
}

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string       // Base Ollama server URL
	defaultModel string       // Default model to use
	logger       *slog.Logger // Logger for structured logging
}

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
		logger:       logger,
	}, nil
}

// Chat sends a message to the Ollama API.
func (ollama OllamaClient) Chat(message string) error {
	// Validate message to LLM
	if strings.TrimSpace(message) == "" {
		ollama.logger.Error("message is empty", slog.String("component", "ollama client"))
		return fmt.Errorf("ollama client chat: %w", ErrMessageEmpty)
	}

	// Create request payload

	// Send request to Ollama API

	// Handle response from Ollama API

	// Return response or error
	return nil
}
