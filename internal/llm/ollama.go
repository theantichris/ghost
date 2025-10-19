package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/charmbracelet/log"
)

// ollamaRequest holds the information for the API request.
type ollamaRequest struct {
	Model        string `json:"model"`  // The model name
	Stream       bool   `json:"stream"` // If false the response is returned as a single object
	SystemPrompt string `json:"system"` // System message to override what is in the model file
	UserPrompt   string `json:"prompt"` // The prompt to generate a response for
}

// ollamaResponse holds the information from the API request.
type ollamaResponse struct {
	Response string `json:"response"`
}

// Ollama is the client for the Ollama API.
type Ollama struct {
	host         string
	generateURL  string
	defaultModel string
	logger       *log.Logger
}

// NewOllama creates and returns a new Ollama client.
func NewOllama(host, defaultModel string, logger *log.Logger) (Ollama, error) {
	if strings.TrimSpace(host) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoHostURL)
	}

	if strings.TrimSpace(defaultModel) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoDefaultModel)
	}

	ollama := Ollama{
		host:         host,
		generateURL:  host + "/api/generate",
		defaultModel: defaultModel,
		logger:       logger,
	}

	logger.Info("initialized Ollama client", "url", ollama.host, "model", ollama.defaultModel)

	return ollama, nil
}

// Generate sends a request to /api/generate and returns the response
func (ollama Ollama) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	ollamaRequest := ollamaRequest{
		Model:        ollama.defaultModel,
		Stream:       false,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}

	ollama.logger.Debug("sending generate request to Ollama API", "url", ollama.generateURL, "model", ollama.defaultModel, "request", ollamaRequest)

	var ollamaResponse ollamaResponse
	err := requests.
		URL(ollama.generateURL).
		BodyJSON(&ollamaRequest).
		ToJSON(&ollamaResponse).
		Fetch(ctx)

	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOllama, err)
	}

	ollama.logger.Debug("response received from Ollama API", "response", ollamaResponse)

	return ollamaResponse.Response, nil
}
