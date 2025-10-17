package llm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

// Ollama is the client for the Ollama API.
type Ollama struct {
	baseURL      string
	defaultModel string
	httpClient   *http.Client
	logger       *log.Logger
}

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
func (ollama Ollama) Generate() {
	// Create request body
	// Create HTTP request
	// Send HTTP request
	// Progress response
	// Return response
}
