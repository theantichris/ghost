package llm

import "context"

// LLMClient is an interface representing an Ollama API client for LLM operations.
type LLMClient interface {
	// Generate sends a system prompt and user prompt to the LLM and returns the generated response text.
	// Returns an error if the API request fails or the response cannot be parsed.
	Generate(ctx context.Context, systemPrompt, userPrompt string, images []string) (string, error)

	// Version retrieves the version string of the Ollama API server.
	// Returns an error if the API request fails.
	Version(ctx context.Context) (string, error)

	// Show validates that the model exists and is accessible on the Ollama API server.
	// Returns an error if the model is not found or the API request fails.
	Show(ctx context.Context, model string) error
}
