package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

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
)

var ErrAskCmd = errors.New("failed to run ask command")

type askCmd struct {
	logger *slog.Logger
}

func NewAskCmd(logger *slog.Logger) *cobra.Command {
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

func (askCmd *askCmd) run(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient(askCmd.logger)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrLLMClientInit, err)
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAskCmd, err)
	}

	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var query string

	if isPiped {
		query, err = readPipedInput(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("%w: %s", ErrReadPipeInput, err)
		}

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, " ")
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")
	} else {
		return fmt.Errorf("%w, provide a query or pipe input", ErrEmptyInput)
	}

	return runSingleQuery(llmClient, query, cmd.OutOrStdout())
}

func initializeLLMClient(logger *slog.Logger) (llm.LLMClient, error) {
	ollamaBaseURL := viper.GetString("ollama")
	model := viper.GetString("model")

	if ollamaBaseURL == "" {
		return nil, fmt.Errorf("%w, set it via OLLAMA_BASE_URL environment variable, config file, or --ollama flag", ErrURLEmpty)
	}

	if model == "" {
		return nil, fmt.Errorf("%w, set it via DEFAULT_MODEL environment variable, config file, or --model flag", ErrModelEmpty)

	}

	timeout := viper.GetDuration("timeout")
	httpClient := &http.Client{Timeout: timeout}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, model, httpClient, logger)
	if err != nil {
		return nil, err
	}

	return llmClient, nil
}

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

	_, _ = fmt.Fprintln(output, message)

	return nil
}

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
