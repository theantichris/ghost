package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

const systemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Answer requests directly and briefly."

var (
	ErrURLEmpty      = errors.New("OLLAMA_BASE_URL not configured")
	ErrModelEmpty    = errors.New("DEFAULT_MODEL name not configured")
	ErrReadPipeInput = errors.New("failed to read piped input")
	ErrEmptyInput    = errors.New("input is empty")
	ErrLLMClientInit = errors.New("failed to create LLM client")
	ErrLLMResponse   = errors.New("failed to print LLM response")
)

var ErrAskCmd = errors.New("failed to run ask command")

// askCmd represents the ask command and its dependencies.
type askCmd struct {
	logger *log.Logger
}

// NewAskCmd creates a new ask command that sends queries to the configured LLM.
// It supports both direct command-line queries and piped input from stdin.
func NewAskCmd(logger *log.Logger) *cobra.Command {
	askCmd := &askCmd{logger: logger}

	cmd := &cobra.Command{
		Use:   "ask [query]",
		Short: "Ask Ghost a question.",
		Long: `Ask Ghost a question.

		Examples:
			# Quick question from command line
			ghost ask "What is the capital of France?"

			# Pipe input to Ghost
			cat code.go | ghost ask "Explain this code"`,
		RunE: askCmd.run,
		Args: cobra.ArbitraryArgs,
	}

	return cmd
}

// run executes the ask command logic.
// It handles both piped input and command-line arguments, sends the query to the LLM,
// and outputs the response.
func (askCmd *askCmd) run(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient(askCmd.logger)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrLLMClientInit, err)
	}

	askCmd.logger.Info("LLM client initialized successfully")

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAskCmd, err)
	}

	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	askCmd.logger.Debug("detected input mode", "piped", isPiped)
	var query string

	if isPiped {
		query, err = readPipedInput(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("%w: %s", ErrReadPipeInput, err)
		}

		askCmd.logger.Debug("read piped input", "bytes", len(query))

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, " ")

			askCmd.logger.Debug("combined piped input with agurments", "args", strings.Join(args, " "))
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")

		askCmd.logger.Debug("using direct arguments as query", args, strings.Join(args, " "))
	} else {
		askCmd.logger.Warn("no input provided")

		return fmt.Errorf("%w, provide a query or pipe input", ErrEmptyInput)
	}

	ctx := cmd.Context()

	askCmd.logger.Info("executing query", "queryLength", len(query))

	return runSingleQuery(ctx, llmClient, query, cmd.OutOrStdout(), askCmd.logger)
}

// initializeLLMClient creates and configures an LLM client using configuration from viper.
// It requires OLLAMA_BASE_URL and DEFAULT_MODEL to be set via environment variables,
// config file, or command-line flags.
func initializeLLMClient(logger *log.Logger) (llm.LLMClient, error) {
	ollamaBaseURL := viper.GetString("ollama")
	model := viper.GetString("model")

	if ollamaBaseURL == "" {
		return nil, fmt.Errorf("%w, set it via OLLAMA_BASE_URL environment variable, config file, or --ollama flag", ErrURLEmpty)
	}

	if model == "" {
		return nil, fmt.Errorf("%w, set it via DEFAULT_MODEL environment variable, config file, or --model flag", ErrModelEmpty)

	}

	timeout := viper.GetDuration("timeout")

	logger.Info("creating Ollama client", "baseURL", ollamaBaseURL, "model", model, "timeout", timeout)

	httpClient := &http.Client{Timeout: timeout}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, model, httpClient, logger)
	if err != nil {
		return nil, err
	}

	return llmClient, nil
}

// readPipedInput reads all input from the provided reader until EOF.
// It's used to capture piped input from stdin.
func readPipedInput(input io.Reader) (string, error) {
	reader := bufio.NewReader(input)

	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					lines = append(lines, line)
				}

				break
			}

			return "", err
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, ""), nil
}

// runSingleQuery sends a single query to the LLM and writes the response to the output.
// It constructs a chat history with the system prompt and user query,
// then strips any think blocks from the response before outputting.
func runSingleQuery(ctx context.Context, llmClient llm.LLMClient, query string, output io.Writer, logger *log.Logger) error {
	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.User, Content: query},
	}

	response, err := llmClient.Chat(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	logger.Info("received response", "contentLength", len(response.Content))

	message := stripThinkBlock(response.Content)

	logger.Debug("stripped think blocks", "finalLength", len(message))

	if _, err = fmt.Fprintln(output, message); err != nil {
		return fmt.Errorf("%w: %s", ErrLLMResponse, err)
	}

	logger.Info("query completed successfully")

	return nil
}

// stripThinkBlock removes <think>...</think> blocks from the message.
// These blocks may contain internal reasoning that shouldn't be shown to the user.
func stripThinkBlock(message string) string {
	openTag := "<think>"
	closeTag := "</think>"

	for {
		start := strings.Index(message, openTag)

		if start == -1 {
			break // No <think> block.
		}

		end := strings.Index(message[start+len(openTag):], closeTag)
		if end == -1 {
			break // No </think> block.
		}

		blockEnd := start + len(openTag) + end + len(closeTag)

		message = message[:start] + message[blockEnd:]
	}

	return strings.TrimSpace(message)
}
