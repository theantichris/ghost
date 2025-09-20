package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/theantichris/ghost/internal/llm"
)

const systemPrompt string = "You are Ghost, a concise terminal assistant. Greet the user warmly once at the start of the conversation, then answer their requests directly and briefly. Ask for clarification only when needed."

const (
	msgClientResponse    string = "(system) I couldn't reach the LLM. Check your network or make sure the host is running then try again"
	msgNon2xxResponse    string = "(system) The LLM response with an error. Verify the model is pulled and the server is healthy before retrying."
	msgResponseBody      string = "(system) I couldn't read the LLM's reply. This might be a transient issue, please try again in a moment."
	msgUnmarshalResponse string = "(system) The LLM sent back something I couldn't parse. It may be busy, try your request again shortly."
)

// Config holds the optional configuration options for App.
type Config struct {
	Output io.Writer
	Debug  bool
}

// App represents the main application structure
type App struct {
	llmClient llm.LLMClient
	logger    *slog.Logger
	output    io.Writer
	debug     bool
}

// New initializes a new App instance with the provided LLM client.
func New(llmClient llm.LLMClient, logger *slog.Logger, config Config) (*App, error) {
	if llmClient == nil {
		return nil, ErrLLMClientNil
	}

	logger.Info("ghost app initialized", slog.String("component", "app"))

	if config.Output == nil {
		config.Output = os.Stdout
	}

	return &App{
		llmClient: llmClient,
		logger:    logger,
		output:    config.Output,
		debug:     config.Debug,
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

	chatHistory, err := app.handleLLMResponse(ctx, chatHistory)
	if err != nil {
		return err
	}

	for {
		endChat := false
		fmt.Fprint(app.output, "User: ")

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

		chatHistory, err = app.handleLLMResponse(ctx, chatHistory)
		if err != nil {
			return err
		}

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

// handleLLMResponse gets the response from the LLMClient, adds it to the chat history and prints the content.
func (app *App) handleLLMResponse(ctx context.Context, chatHistory []llm.ChatMessage) ([]llm.ChatMessage, error) {
	llmResponse, err := app.llmClient.Chat(ctx, chatHistory)
	if err != nil {
		if errors.Is(err, llm.ErrClientResponse) {
			fmt.Fprintf(app.output, "\n%s\n", msgClientResponse)

			return chatHistory, nil
		}

		if errors.Is(err, llm.ErrNon2xxResponse) {
			fmt.Fprintf(app.output, "\n%s\n", msgNon2xxResponse)

			return chatHistory, nil
		}

		if errors.Is(err, llm.ErrResponseBody) {
			fmt.Fprintf(app.output, "\n%s\n", msgResponseBody)

			return chatHistory, nil
		}

		if errors.Is(err, llm.ErrUnmarshalResponse) {
			fmt.Fprintf(app.output, "\n%s\n", msgUnmarshalResponse)

			return chatHistory, nil
		}

		return chatHistory, fmt.Errorf("%w: %s", ErrChatFailed, err)
	}

	chatHistory = append(chatHistory, llmResponse)

	response := stripThinkBlock(llmResponse.Content)
	fmt.Fprintf(app.output, "\nGhost: %s\n", response)

	return chatHistory, nil
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
