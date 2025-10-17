package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

// ollamaRequest holds the information for the API request.
type ollamaRequest struct {
	Model        string `json:"model"`  // The model name
	Stream       bool   `json:"bool"`   // If false the response is returned as a single object
	SystemPrompt string `json:"system"` // System message to override what is in the model file
	UserPrompt   string `json:"prompt"` // The prompt to generate a response for
}

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
func (ollama Ollama) Generate(systemPrompt, userPrompt string) string {
	ollamaRequest := ollamaRequest{
		Model:        ollama.defaultModel,
		Stream:       false,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}

	requestBody, _ := json.Marshal(ollamaRequest)

	httpRequest, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, ollama.baseURL+"/api/generate", bytes.NewReader(requestBody))
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, _ := ollama.httpClient.Do(httpRequest)

	// Check status code

	body, _ := io.ReadAll(httpResponse.Body)
	var response ollamaResponse
	_ = json.Unmarshal(body, &response)

	return response.Response
}
