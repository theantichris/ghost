package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/theantichris/ghost/internal/llm"
)

const systemPrompt string = "You are Ghost, a concise terminal assistant. Greet the user warmly once at the start of the conversation, then answer their requests directly and briefly. Ask for clarification only when needed."

// App represents the main application structure
type App struct {
	llmClient llm.LLMClient
	logger    *slog.Logger
	debug     bool
}

// New initializes a new App instance with the provided LLM client.
func New(llmClient llm.LLMClient, debug bool, logger *slog.Logger) (*App, error) {
	if llmClient == nil {
		return nil, ErrLLMClientNil
	}

	logger.Info("ghost app initialized", slog.String("component", "app"))

	return &App{
		llmClient: llmClient,
		logger:    logger,
		debug:     debug,
	}, nil
}

// Run starts the application logic.
func (app *App) Run(ctx context.Context, input io.Reader) error {
	app.logger.Info("starting chat loop", slog.String("component", "app"))

	scanner := bufio.NewScanner(input)

	// Add system prompt and greeting
	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
	}

	llmResponse, err := app.llmClient.Chat(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrChatFailed, err)
	}

	chatHistory = append(chatHistory, llmResponse)

	response := stripThinkBlock(llmResponse.Content)
	fmt.Fprintf(os.Stdout, "\nGhost: %s\n", response)

	for {
		endChat := false
		fmt.Print("User: ")

		if ok := scanner.Scan(); !ok {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("%w: %s", ErrReadingInput, err)
			}

			break // EOF reached
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		if userInput == "/bye" {
			endChat = true
			userInput = "Goodbye!"
		}

		userMessage := llm.ChatMessage{Role: llm.User, Content: userInput}
		chatHistory = append(chatHistory, userMessage)

		llmResponse, err := app.llmClient.Chat(ctx, chatHistory)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrChatFailed, err)
		}

		chatHistory = append(chatHistory, llmResponse)

		response := stripThinkBlock(llmResponse.Content)
		fmt.Fprintf(os.Stdout, "\nGhost: %s\n", response)

		if endChat {
			break
		}
	}

	app.logger.Info("stopping chat loop", slog.String("component", "app"))

	if app.debug {
		spew.Dump(chatHistory)
	}

	return nil
}

// stripThinkBlock removes any <think>...</think> blocks from the message.
func stripThinkBlock(message string) string {
	openTag := "<think>"
	closeTag := "</think>"

	for {
		start := strings.Index(message, openTag)
		if start == -1 {
			break
		}

		end := strings.Index(message[start+len(openTag):], closeTag)
		if end == -1 {
			break
		}

		blockEnd := start + len(openTag) + end + len(closeTag)

		message = message[:start] + message[blockEnd:]
	}

	return strings.TrimSpace(message)
}
