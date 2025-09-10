package llm

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Tool      Role = "tool"
)

type OllamaChatRequest struct {
	Model    string               `json:"model"`           // Required. The model name.
	Messages *[]OllamaChatMessage `json:"message"`         // The messages of the chat, this can be used to keep a chat memory
	Tools    json.RawMessage      `json:"tools,omitempty"` //  List of tools in JSON for the model to use if supported
}

type OllamaChatMessage struct {
	Role      Role            `json:"role"`                 // The role of the message, either system, user, assistant, or tool
	Content   string          `json:"content"`              // The content of the message
	ToolCalls json.RawMessage `json:"tool_calls,omitempty"` // A list of tools in JSON that the model wants to use
	ToolName  string          `json:"tool_name,omitempty"`  // Add the name of the tool that was executed to inform the model of the result
}

// OllamaClient is a client for interacting with the Ollama API.
type OllamaClient struct {
	baseURL      string       // Base Ollama server URL
	defaultModel string       // Default model to use
	stream       bool         // Whether to use streaming responses
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
		stream:       true,
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
