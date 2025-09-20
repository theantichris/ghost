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
	llmClient llm.LLMClient
	logger    *slog.Logger
}

// New initializes a new App instance with the provided LLM client.
func New(llmClient llm.LLMClient, logger *slog.Logger) (*App, error) {
	if llmClient == nil {
		return nil, ErrLLMClientNil
	}

	logger.Info("ghost app initialized", slog.String("component", "app"))

	return &App{
		llmClient: llmClient,
		logger:    logger,
	}, nil
}

// Run starts the application logic.
func (app *App) Run(ctx context.Context, input io.Reader) error {
	app.logger.Info("starting chat loop", slog.String("component", "app"))

	scanner := bufio.NewScanner(input)

	chatHistory := []llm.ChatMessage{}

	for {
		fmt.Print("User: ")

		if ok := scanner.Scan(); !ok {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("%w: %s", ErrReadingInput, err)
			}

			break // EOF reached
		}

		userInput := strings.TrimSpace(scanner.Text())
		userMessage := llm.ChatMessage{Role: llm.User, Content: userInput}
		chatHistory = append(chatHistory, userMessage)

		if userInput == "/bye" {
			// TODO: Add goodbye message from LLM
			app.logger.Info("stopping chat loop", slog.String("component", "app"))
			break
		}

		if userInput == "" {
			continue
		}

		llmResponse, err := app.llmClient.Chat(ctx, chatHistory)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrChatFailed, err)
		}

		chatHistory = append(chatHistory, llmResponse)

		fmt.Fprintf(os.Stdout, "\nGhost: %s\n", llmResponse.Content)
	}

	return nil
}
