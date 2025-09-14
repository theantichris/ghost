package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

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
		ctx:       ctx, // TODO: Should probably be passing context instead of adding it to app.
		llmClient: llmClient,
		logger:    logger,
	}, nil
}

// Run starts the application logic.
func (app *App) Run(input io.Reader) error {
	// TODO: Add prompt
	// TODO: Add chat labels
	// TODO: Fix tests
	app.logger.Info("starting chat loop", slog.String("component", "app"))

	scanner := bufio.NewScanner(input)

	var userInput string

	for {
		scanner.Scan()
		userInput = strings.TrimSpace(scanner.Text())

		if userInput == "/bye" {
			app.logger.Info("exiting chat loop", slog.String("component", "app"))
			break
		}

		if userInput == "" {
			continue
		}

		response, err := app.llmClient.Chat(app.ctx, userInput)
		if err != nil {
			app.logger.Error("failed to chat with Ollama", slog.String("component", "app"), slog.String("error", err.Error()))
			return fmt.Errorf("app run: %w", err)
		}

		fmt.Fprintf(os.Stdout, "\nOllama response: %s\n", response)
	}

	return nil
}
