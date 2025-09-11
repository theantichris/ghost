package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/theantichris/ghost/internal/llm"
)

// App represents the main application structure
type App struct {
	ctx       context.Context
	llmClient llm.LLMClient
	logger    *slog.Logger
}

// New initializes a new App instance with the provided LLM client.
func New(ctx context.Context, llmClient llm.LLMClient, logger *slog.Logger) (*App, error) {
	if llmClient == nil {
		logger.Error("llmClient is nil", slog.String("component", "app"))

		return nil, fmt.Errorf("app init: %w", ErrLLMClientNil)
	}

	logger.Info("ghost app initialized", slog.String("component", "app"))

	return &App{
		ctx:       ctx,
		llmClient: llmClient,
		logger:    logger,
	}, nil
}

// Run starts the application logic.
func (app *App) Run() error {
	// TODO: Get user input from CLI for message

	app.logger.Info("starting chat with Ollama", slog.String("component", "app"), slog.String("message", "Hello, Ollama!"))

	response, err := app.llmClient.Chat(app.ctx, "Hello, Ollama!")
	if err != nil {
		app.logger.Error("failed to chat with Ollama", slog.String("component", "app"), slog.String("error", err.Error()))
		return fmt.Errorf("app run: %w", err)
	}

	app.logger.Info("received response from Ollama", slog.String("component", "app"), slog.String("response", response))
	fmt.Fprintf(os.Stdout, "Ollama response: %s\n", response)

	return nil
}
