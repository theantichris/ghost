package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/internal/ui"
)

const (
	chatUseText     = "chat"
	chatShortText   = "starts ghost in chat mode"
	chatLongText    = "starts ghost in chat mode, use :q to quit"
	chatExampleText = "ghost chat"
)

func newChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     chatUseText,
		Short:   chatShortText,
		Long:    chatLongText,
		Example: chatExampleText,
		Args:    cobra.NoArgs,
		RunE:    runChat,
	}

	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	tavilyAPIKey := viper.GetString("search.api-key")
	maxResults := viper.GetInt("search.max-results")

	config := ui.ModelConfig{
		Context:     cmd.Context(),
		Logger:      logger,
		URL:         viper.GetString("url"),
		Model:       viper.GetString("model"),
		VisionModel: viper.GetString("vision.model"),
		System:      agent.SystemPrompt,
		Registry:    tool.NewRegistry(tavilyAPIKey, maxResults, logger),
	}

	chatModel := ui.NewChatModel(config)

	logger.Info("entering chat", "ollama_url", config.URL, "chat_model", config.Model, "vision_model", config.VisionModel)
	program := tea.NewProgram(chatModel)

	_, err := program.Run()

	return err
}
