package cmd

import (
	"fmt"

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
// It initializes the LLM client, retrieves user input, and processes the query.
func (askCmd *askCmd) run(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient(askCmd.logger)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	inputReader := newInputReader(askCmd.logger)
	userInput, err := inputReader.getUserInput(cmd, args)
	if err != nil {
		return err
	}

	askCmd.logger.Info("executing query", "queryLength", len(userInput))

	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.User, Content: userInput},
	}

	response, err := llmClient.Chat(cmd.Context(), chatHistory)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	askCmd.logger.Info("received response", "contentLength", len(response.Content))

	message := stripThinkBlock(response.Content)

	askCmd.logger.Debug("stripped think blocks", "finalLength", len(message))

	if _, err = fmt.Fprintln(cmd.OutOrStdout(), message); err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	askCmd.logger.Info("query completed successfully")

	return nil
}
