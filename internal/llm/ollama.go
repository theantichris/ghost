package llm

import (
	"net/http"

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
	ollama := Ollama{
		baseURL:      baseURL,
		defaultModel: defaultModel,
		httpClient:   httpClient,
		logger:       logger,
	}

	return ollama, nil
}
