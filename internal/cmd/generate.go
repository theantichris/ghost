package cmd

import (
	"context"

	"github.com/theantichris/ghost/internal/llm"
)

// generate sends the prompt to the LLM API and returns the response
func generate(ctx context.Context, systemPrompt, userPrompt string, llmClient llm.LLMClient) (string, error) {
	response, err := llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	return response, nil
}
