package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/tui"
)

// chatCmd represents the chat command and its dependencies.
type chatCmd struct {
	logger    *log.Logger
	llmClient llm.LLMClient
}

// NewChatCmd creates a new chat command that launches an interactive TUI session
// for multi-turn conversations with the configured LLM.
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

// run initializes the LLM client and launches the interactive TUI chat interface.
func (chatCmd *chatCmd) run(cmd *cobra.Command, args []string) error {
	if chatCmd.llmClient == nil {
		llmClient, err := initializeLLMClient(chatCmd.logger)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrLLM, err)
		}

		chatCmd.llmClient = llmClient
	}

	model := tui.NewModel(chatCmd.llmClient, systemPrompt, chatCmd.logger)

	return tui.Run(model)
}
