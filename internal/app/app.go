package app

import (
	"fmt"
	"log/slog"

	"github.com/theantichris/ghost/internal/llm"
)

// App represents the main application structure
type App struct {
	llmClient llm.LLMClient
	logger    *slog.Logger
}

// New initializes a new App instance with the provided LLM client.
func New(llmClient llm.LLMClient, logger *slog.Logger) (*App, error) {
	if llmClient == nil {
		logger.Error("llmClient is nil", slog.String("component", "app"))

		return nil, fmt.Errorf("app init: %w", ErrLLMClientNil)
	}

	return &App{
		llmClient: llmClient,
		logger:    logger,
	}, nil
}

func (app *App) Run() error {
	// Send message to Ollama

	// Display response from Ollama

	return nil
}
