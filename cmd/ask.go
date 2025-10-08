package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/stdio"
)

// askCmd represents the ask command and its dependencies.
type askCmd struct {
	logger    *log.Logger
	llmClient llm.LLMClient
}

// NewAskCmd creates a new ask command that sends queries to the configured LLM.
// It supports both direct command-line queries and piped input from stdin.
func NewAskCmd(logger *log.Logger) *cobra.Command {
	askCmd := &askCmd{
		logger: logger,
	}

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
	if askCmd.llmClient == nil {
		llmClient, err := initializeLLMClient(askCmd.logger)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrLLM, err)
		}

		askCmd.llmClient = llmClient
	}

	inputReader := stdio.NewInputReader(askCmd.logger)
	userInput, err := inputReader.Read(cmd, args)
	if err != nil {
		return err
	}

	askCmd.logger.Info("executing query", "queryLength", len(userInput))

	chatHistory := []llm.ChatMessage{
		{Role: llm.SystemRole, Content: systemPrompt},
		{Role: llm.UserRole, Content: userInput},
	}

	var tokens string

	// OutputWriter.Tokens is a pointer to accumulate all tokens (including think
	//  blocks) while filtering only non-think content to Output.
	outputWriter := &stdio.OutputWriter{
		Logger: askCmd.logger,
		Output: cmd.OutOrStdout(),
		Tokens: &tokens,
	}

	if err := askCmd.llmClient.Chat(cmd.Context(), chatHistory, outputWriter.Write); err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}
	outputWriter.Flush()

	if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("%w: %w", stdio.ErrIO, err)
	}

	askCmd.logger.Info("query completed successfully")

	return nil
}
