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

const (
	userLabel  string = "\nUser: "
	ghostLabel string = "\nGhost: "
)

const (
	exitCommand string = "/bye"
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
	chatHistory := []llm.ChatMessage{{Role: llm.System, Content: systemPrompt}}

	llmMessage, err := app.getLLMMessage(ctx, chatHistory)
	if err != nil {
		return err
	}
	chatHistory = append(chatHistory, llmMessage)

	app.logger.Info("starting chat loop", slog.String("component", "app"))
	userInputScanner := bufio.NewScanner(input)

	for {
		userMessage, endChat, err := app.getUserMessage(userInputScanner)
		if err != nil {
			if errors.Is(err, ErrUserInputEmpty) {
				continue // Don't send empty user input.
			}

			return err
		}
		chatHistory = append(chatHistory, userMessage)

		llmMessage, err := app.getLLMMessage(ctx, chatHistory)
		if err != nil {
			return err
		}
		chatHistory = append(chatHistory, llmMessage)

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

// getLLMMessage sends the chat history to the LLM, streams the response and returns the LLM's message.
func (app *App) getLLMMessage(ctx context.Context, chatHistory []llm.ChatMessage) (llm.ChatMessage, error) {
	fmt.Fprint(app.output, ghostLabel)

	var tokens string
	err := app.llmClient.StreamChat(ctx, chatHistory, func(token string) {
		fmt.Fprint(app.output, token)
		tokens += token

		// TODO: Remove <think> blocks.
	})

	fmt.Fprint(app.output, "\n") // Add newline after LLM message.

	if err != nil {
		if err := app.handleLLMResponseError(err); err != nil {
			return llm.ChatMessage{}, err
		}
	}

	message := llm.ChatMessage{Role: llm.Assistant, Content: tokens}

	return message, nil
}

// getUserMessage gets the user's input and returns a llm.ChatMessage and endChat flag.
func (app *App) getUserMessage(scanner *bufio.Scanner) (llm.ChatMessage, bool, error) {
	var endChat = false

	fmt.Fprint(app.output, userLabel)

	if ok := scanner.Scan(); !ok {
		if err := scanner.Err(); err != nil {
			return llm.ChatMessage{}, endChat, fmt.Errorf("%w: %s", ErrReadingInput, err)
		}
	}

	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		return llm.ChatMessage{}, endChat, ErrUserInputEmpty
	}

	if input == exitCommand {
		endChat = true
		input = "Goodbye!"
	}

	message := llm.ChatMessage{Role: llm.User, Content: input}

	return message, endChat, nil
}

// handleLLMResponseError shows a system message for recoverable errors and returns unrecoverable errors.
func (app *App) handleLLMResponseError(err error) error {
	errorMap := map[error]string{
		llm.ErrClientResponse:    msgClientResponse,
		llm.ErrNon2xxResponse:    msgNon2xxResponse,
		llm.ErrResponseBody:      msgResponseBody,
		llm.ErrUnmarshalResponse: msgUnmarshalResponse,
	}

	for error, msg := range errorMap {
		if errors.Is(err, error) {
			fmt.Fprintf(app.output, "\n%s\n", msg)

			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrChatFailed, err)
}
