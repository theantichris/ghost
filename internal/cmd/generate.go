package cmd

import (
	"context"

	"github.com/theantichris/ghost/internal/llm"
)

// generate sends the system prompt and user prompt to the LLM API via the provided client and returns the response.
// Returns an error if the LLM client fails to generate a response
func generate(ctx context.Context, systemPrompt, userPrompt string, llmClient llm.LLMClient) (string, error) {
	response, err := llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	return response, nil
}
