package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theantichris/ghost/internal/llm"
)

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
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInput, err)
	}

	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	askCmd.logger.Debug("detected input mode", "piped", isPiped)

	var query string

	if isPiped {
		query, err = readPipedInput(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIO, err)
		}

		askCmd.logger.Debug("read piped input", "bytes", len(query))

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, " ")

			askCmd.logger.Debug("combined piped input with arguments", "argCount", len(args))
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")

		askCmd.logger.Debug("using direct arguments as query", "argCount", len(args))
	} else {
		askCmd.logger.Warn("no input provided")

		return fmt.Errorf("%w: provide a query or pipe input", ErrInput)
	}

	ctx := cmd.Context()

	askCmd.logger.Info("executing query", "queryLength", len(query))

	return processQuery(ctx, llmClient, query, cmd.OutOrStdout(), askCmd.logger)
}

// processQuery sends a single query to the LLM and writes the response to the output.
// It constructs a chat history with the system prompt and user query,
// then strips any think blocks from the response before outputting.
func processQuery(ctx context.Context, llmClient llm.LLMClient, query string, output io.Writer, logger *log.Logger) error {
	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.User, Content: query},
	}

	response, err := llmClient.Chat(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	logger.Info("received response", "contentLength", len(response.Content))

	message := stripThinkBlock(response.Content)

	logger.Debug("stripped think blocks", "finalLength", len(message))

	if _, err = fmt.Fprintln(output, message); err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	logger.Info("query completed successfully")

	return nil
}
