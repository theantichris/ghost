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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

var (
	systemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Answer requests directly and briefly."
	noNewLine    bool
	timeout      time.Duration

	ErrLLMBaseURLEmpty = errors.New("LLM API base url is empty")
	ErrModelEmpty      = errors.New("model is empty")
	ErrReadPipeInput   = errors.New("failed to read piped input")
	ErrEmptyInput      = errors.New("input is empty")
	ErrLLMClientInit   = errors.New("failed to create LLM client")
)

// askCmd represents the command called with chat.
var askCmd = &cobra.Command{
	Use:   "ask [query]",
	Short: "Ask Ghost a question.",
	Long: `Ask Ghost a question.

		Examples:
			# Quick question from command line
			ghost ask "What is the capital of France?"

			# Pipe input to Ghost
			cat code.go | ghost ask "Explain this code`,
	RunE: runAsk,
	Args: cobra.ArbitraryArgs,
}

// init initializes the chat command.
func init() {
	rootCmd.AddCommand(askCmd)

	askCmd.Flags().BoolVarP(&noNewLine, "no-newline", "n", false, "Don't add newline after response (useful for scripts)")
	askCmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "HTTP timeout for LLM requests")
}

// runAsk sends the query to the LLM and returns the response.
func runAsk(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient()
	if err != nil {
		Logger.Error(ErrLLMClientInit.Error(), "error", err, "component", "askCmd")
		return fmt.Errorf("%w: %s", ErrLLMClientInit, err)
	}

	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var query string

	if isPiped {
		query, err = readPipedInput()
		if err != nil {
			Logger.Error(ErrReadPipeInput.Error(), "error", err, "component", "askCmd")
			return fmt.Errorf("%w: %s", ErrReadPipeInput, err)
		}

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, "")
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")
	} else {
		Logger.Error(ErrEmptyInput.Error(), "component", "askCmd")
		return fmt.Errorf("%w, provide a query or pipe input", ErrEmptyInput)
	}

	return runSingleQuery(llmClient, query, cmd.OutOrStdout())
}

// initializeLLMClient validates the client config, creates, and returns the client.
func initializeLLMClient() (*llm.OllamaClient, error) {
	ollamaBaseURL := viper.GetString("ollama_base_url")
	model := viper.GetString("model")

	if ollamaBaseURL == "" {
		return nil, fmt.Errorf("%w, set it via OLLAMA_BASE_URL environment variable or config file", ErrLLMBaseURLEmpty)
	}

	if model == "" {
		return nil, fmt.Errorf("%w, set it via DEFAULT_MODEL environment variable, config file, or --model flag", ErrModelEmpty)

	}

	Logger.Debug("initializing LLM client", "model", model, "base_url", ollamaBaseURL, "component", "chatCmd")

	httpClient := &http.Client{Timeout: timeout}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, model, httpClient, Logger)
	if err != nil {
		return nil, err
	}

	return llmClient, nil
}

// readPipedInput reads input piped from the CLI.
func readPipedInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

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

// runSingleQuery sends the query to the LLM client and writes the response.
func runSingleQuery(llmClient llm.LLMClient, query string, output io.Writer) error {
	ctx := context.Background()

	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.User, Content: query},
	}

	response, err := llmClient.Chat(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	message := stripThinkBlock(response.Content)

	if noNewLine {
		fmt.Fprint(output, message)
	} else {
		fmt.Fprintln(output, message)
	}

	return nil
}

// stripThinkBlock removes <think>...</think> from a string.
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
