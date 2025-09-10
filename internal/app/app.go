package app

import (
	"fmt"

	"github.com/theantichris/ghost/internal/llm"
)

// App represents the main application structure
type App struct {
	llmClient llm.LLMClient
}

// New initializes a new App instance with the provided LLM client.
func New(llmClient llm.LLMClient) (*App, error) {
	if llmClient == nil {
		return nil, fmt.Errorf("app init: %w", ErrLLMClientNil)
	}

	return &App{
		llmClient: llmClient,
	}, nil
}
