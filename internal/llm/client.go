package llm

import "context"

// LLMClient is an interface representing an OpenAPI client.
type LLMClient interface {
	//Generate sends a request to the generate endpoint and returns the response.
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
