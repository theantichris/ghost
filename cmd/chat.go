package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/stdio"
)

// chatCmd represents the chat command and its dependencies.
type chatCmd struct {
	logger    *log.Logger
	llmClient llm.LLMClient
}

// NewChatCmd creates a new chat command that sends queries to the configured LLM.
// It supports both direct command-line queries and piped input from stdin.
func NewChatCmd(logger *log.Logger) *cobra.Command {
	chatCmd := &chatCmd{
		logger: logger,
	}

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
	if chatCmd.llmClient == nil {
		llmClient, err := initializeLLMClient(chatCmd.logger)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrLLM, err)
		}

		chatCmd.llmClient = llmClient
	}

	chatHistory := []llm.ChatMessage{
		{Role: llm.SystemRole, Content: systemPrompt},
		{Role: llm.SystemRole, Content: "Greet the user"},
	}

	var tokens string
	writer := &stdio.OutputWriter{
		Logger: chatCmd.logger,
		Output: cmd.OutOrStdout(),
		Tokens: &tokens,
	}

	// Send system and greeting prompt.
	if err := chatCmd.llmClient.Chat(cmd.Context(), chatHistory, writer.Write); err != nil {
		return fmt.Errorf("%w: %w", ErrLLM, err)
	}
	writer.Flush()

	chatHistory = append(chatHistory, llm.ChatMessage{Role: llm.AssistantRole, Content: tokens})

	if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("%w: %w", stdio.ErrIO, err)
	}

	inputScanner := bufio.NewScanner(cmd.InOrStdin())
	endChat := false

	for !endChat {
		if ok := inputScanner.Scan(); !ok {
			if err := inputScanner.Err(); err != nil {
				return fmt.Errorf("%w: %w", stdio.ErrIO, err)
			}

			break // Reached EOF.
		}

		input := strings.TrimSpace(inputScanner.Text())

		if input == "" {
			continue
		}

		// Setup exit routine.
		if input == "/bye" || input == "/exit" {
			endChat = true
			input = "Goodbye!"
		}

		chatHistory = append(chatHistory, llm.ChatMessage{Role: llm.UserRole, Content: input})

		writer.Reset()
		if err := chatCmd.llmClient.Chat(cmd.Context(), chatHistory, writer.Write); err != nil {
			return fmt.Errorf("%w: %w", ErrLLM, err)
		}
		writer.Flush()

		chatHistory = append(chatHistory, llm.ChatMessage{Role: llm.AssistantRole, Content: tokens})

		if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("%w: %w", stdio.ErrIO, err)
		}
	}

	return nil
}
