package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/charmbracelet/log"
)

// generateRequest represents the JSON payload sent to the Ollama /api/generate endpoint.
type generateRequest struct {
	Model        string   `json:"model"`            // The model name
	Stream       bool     `json:"stream"`           // If false the response is returned as a single object
	SystemPrompt string   `json:"system"`           // System message to override what is in the model file
	UserPrompt   string   `json:"prompt"`           // The prompt to generate a response for
	Images       []string `json:"images,omitempty"` // A list of base64 encoded images
}

// generateResponse represents the JSON response received from the Ollama /api/generate endpoint.
type generateResponse struct {
	Response string `json:"response"`
}

// versionResponse represents the JSON response received from the Ollama /api/version endpoint.
type versionResponse struct {
	Version string `json:"version"`
}

// showRequest represents the JSON payload sent to the Ollama /api/show endpoint.
type showRequest struct {
	Model string `json:"model"`
}

// Config holds the configuration values for the Ollama client.
type Config struct {
	Host         string
	DefaultModel string
	VisionModel  string
}

// Ollama is the client for the Ollama API.
type Ollama struct {
	host         string
	generateURL  string
	versionURL   string
	showURL      string
	defaultModel string
	visionModel  string
	logger       *log.Logger
}

// NewOllama creates and returns a new Ollama client.
func NewOllama(config Config, logger *log.Logger) (Ollama, error) {
	if strings.TrimSpace(config.Host) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoHostURL)
	}

	if strings.TrimSpace(config.DefaultModel) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoDefaultModel)
	}

	if strings.TrimSpace(config.VisionModel) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoVisionModel)
	}

	ollama := Ollama{
		host:         config.Host,
		generateURL:  config.Host + "/api/generate",
		versionURL:   config.Host + "/api/version",
		showURL:      config.Host + "/api/show",
		defaultModel: config.DefaultModel,
		visionModel:  config.VisionModel,
		logger:       logger,
	}

	logger.Debug("initialized Ollama client", "url", ollama.host, "model", ollama.defaultModel)

	return ollama, nil
}

// Generate sends a request to /api/generate with the system and user prompt and returns the generated response.
// Returns ErrOllama wrapped with the underlying error if the API request fails.
func (ollama Ollama) Generate(ctx context.Context, systemPrompt, userPrompt string, images []string) (string, error) {
	request := generateRequest{
		Model:        ollama.defaultModel,
		Stream:       false,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Images:       images,
	}

	// TODO: Another config struct?
	// TODO: If images then send a request to the vision model.
	// TODO: append the response to the prompt for the default model.

	ollama.logger.Debug("sending generate request to Ollama API", "url", ollama.generateURL, "model", ollama.defaultModel, "request", request)

	var response generateResponse
	err := requests.
		URL(ollama.generateURL).
		BodyJSON(&request).
		ToJSON(&response).
		Fetch(ctx)

	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOllama, err)
	}

	ollama.logger.Debug("response received from Ollama API", "response", response)

	return response.Response, nil
}

// Version calls the /api/version endpoint to and returns the version string.
// Returns an error if the API request fails.
func (ollama Ollama) Version(ctx context.Context) (string, error) {
	ollama.logger.Debug("sending version request to Ollama API", "url", ollama.versionURL)

	var response versionResponse
	err := requests.
		URL(ollama.versionURL).
		ToJSON(&response).
		Fetch(ctx)

	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrOllama, err)
	}

	ollama.logger.Debug("response received from Ollama API", "response", response)

	return response.Version, nil
}

// Show calls the /api/show endpoint to verify the configured model exists and is accessible.
// Returns an error if the model is not found or the API request fails.
func (ollama Ollama) Show(ctx context.Context) error {
	ollama.logger.Debug("sending show request to Ollama API", "url", ollama.showURL)

	request := showRequest{
		Model: ollama.defaultModel,
	}

	err := requests.
		URL(ollama.showURL).
		BodyJSON(&request).
		Fetch(ctx)

	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllama, err)
	}

	return nil
}
