package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/charmbracelet/log"
)

// ollamaRequest holds the information for the API request.
type ollamaRequest struct {
	Model        string `json:"model"`  // The model name
	Stream       bool   `json:"bool"`   // If false the response is returned as a single object
	SystemPrompt string `json:"system"` // System message to override what is in the model file
	UserPrompt   string `json:"prompt"` // The prompt to generate a response for
}

// ollamaResponse holds the information from the API request.
type ollamaResponse struct {
	Response string `json:"response"`
}

// Ollama is the client for the Ollama API.
type Ollama struct {
	baseURL      string
	defaultModel string
	httpClient   *http.Client
	logger       *log.Logger
}

var ErrOllama = errors.New("failed to get API response")

// NewOllama creates and returns a new Ollama client.
func NewOllama(baseURL, defaultModel string, httpClient *http.Client, logger *log.Logger) (Ollama, error) {
	if strings.TrimSpace(baseURL) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoBaseURL)
	}

	if strings.TrimSpace(defaultModel) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoDefaultModel)
	}

	ollama := Ollama{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}

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

	var ollamaResponse ollamaResponse

	err := requests.
		URL(ollama.baseURL + "/api/generate").
		BodyJSON(&ollamaRequest).
		ToJSON(&ollamaResponse).
		Fetch(ctx)

	if err != nil {
		return ollamaResponse.Response, fmt.Errorf("%w: %w", ErrOllama, err)
	}

	return ollamaResponse.Response, nil
}
