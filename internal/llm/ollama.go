package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/charmbracelet/log"
)

// generateRequest represents the JSON payload sent to the Ollama generate endpoint.
type generateRequest struct {
	Model        string   `json:"model"`            // The model name
	Stream       bool     `json:"stream"`           // If false the response is returned as a single object
	SystemPrompt string   `json:"system"`           // System message to override what is in the model file
	Prompt       string   `json:"prompt"`           // The prompt to generate a response for
	Images       []string `json:"images,omitempty"` // A list of base64 encoded images
}

// generateResponse represents the JSON response received from the Ollama generate endpoint.
type generateResponse struct {
	Response string `json:"response"`
}

// versionResponse represents the JSON response received from the Ollama version endpoint.
type versionResponse struct {
	Version string `json:"version"`
}

// showRequest represents the JSON payload sent to the Ollama show endpoint.
type showRequest struct {
	Model string `json:"model"`
}

// TODO: probably can remove this.
// Config holds the configuration values for the Ollama client.
type Config struct {
	Host string
}

// Ollama is the client for the Ollama API.
type Ollama struct {
	host        string
	generateURL string
	versionURL  string
	showURL     string
	logger      *log.Logger
}

// NewOllama creates and returns a new Ollama client.
func NewOllama(config Config, logger *log.Logger) (Ollama, error) {
	if strings.TrimSpace(config.Host) == "" {
		return Ollama{}, fmt.Errorf("%w", ErrNoHostURL)
	}

	ollama := Ollama{
		host:        config.Host,
		generateURL: config.Host + "/api/generate",
		versionURL:  config.Host + "/api/version",
		showURL:     config.Host + "/api/show",
		logger:      logger,
	}

	logger.Debug("initialized Ollama client", "url", ollama.host)

	return ollama, nil
}

// Generate sends a request to the Ollama API generate endpoint and streams the
// response through the callback.
// The callback is called for each chunk of text as it arrives.
// If images are included those are added to the request.
// Returns ErrOllama wrapped with the underlying error if the API request fails.
func (ollama Ollama) Generate(ctx context.Context, model, systemPrompt, prompt string, images []string, callback func(string) error) error {
	request := generateRequest{
		Model:        model,
		Stream:       true,
		SystemPrompt: systemPrompt,
		Prompt:       prompt,
		Images:       images,
	}

	ollama.logger.Debug("sending generate request to Ollama API", "url", ollama.generateURL, "model", model, "request", request)

	err := requests.
		URL(ollama.generateURL).
		BodyJSON(&request).
		Handle(func(response *http.Response) error {
			defer func() {
				_ = response.Body.Close()
			}()

			decoder := json.NewDecoder(response.Body)

			for {
				var apiResponse generateResponse

				if err := decoder.Decode(&apiResponse); err != nil {
					if err == io.EOF {
						break
					}

					return err
				}

				if apiResponse.Response != "" {
					if err := callback(apiResponse.Response); err != nil {
						return err
					}
				}
			}

			return nil
		}).
		Fetch(ctx)

	if err != nil {
		if requests.HasStatusErr(err, http.StatusNotFound) {
			return fmt.Errorf("%w: %w", ErrModelNotFound, err)
		}

		return fmt.Errorf("%w: %w", ErrOllama, err)
	}

	ollama.logger.Debug("streaming complete from Ollama API")

	return nil
}

// Version calls the version endpoint to and returns the version string.
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

// Show calls the show endpoint to verify the model exists and is accessible.
// Returns an error if the model is not found or the API request fails.
func (ollama Ollama) Show(ctx context.Context, model string) error {
	ollama.logger.Debug("sending show request to Ollama API", "url", ollama.showURL)

	request := showRequest{
		Model: model,
	}

	err := requests.
		URL(ollama.showURL).
		BodyJSON(&request).
		Fetch(ctx)

	if err != nil {
		if requests.HasStatusErr(err, http.StatusNotFound) {
			return fmt.Errorf("%w: %w", ErrModelNotFound, err)
		}

		return fmt.Errorf("%w: %w", ErrOllama, err)
	}

	return nil
}
