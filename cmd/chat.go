package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theantichris/ghost/internal/llm"
)

// chatCmd represents the chat command and its dependencies.
type chatCmd struct {
	logger *log.Logger
}

// NewChatCmd creates a new chat command that sends queries to the configured LLM.
// It supports both direct command-line queries and piped input from stdin.
func NewChatCmd(logger *log.Logger) *cobra.Command {
	chatCmd := &chatCmd{logger: logger}

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start a chat with Ghost.",
		Long: `Start a chat with Ghost.

		Examples:
			ghost chat`,
		RunE: chatCmd.run,
		Args: cobra.ArbitraryArgs,
	}

	return cmd
}

func (chatCmd *chatCmd) run(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient(chatCmd.logger)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.System, Content: "Greet the user"},
	}

	// TODO: Send message to LLM and stream response

	var tokens string
	outputWriter := &outputWriter{
		logger: chatCmd.logger,
		output: cmd.OutOrStdout(),
		tokens: &tokens,
	}

	if err := llmClient.StreamChat(cmd.Context(), chatHistory, outputWriter.write); err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("%w: %w", ErrIO, err)
	}

	// chatHistory = append(chatHistory, llm.ChatMessage{Role: llm.Assistant, Content: tokens})

	return nil
}
