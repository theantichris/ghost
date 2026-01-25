package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	url := viper.GetString("url")
	model := viper.GetString("model")

	logger.Info("entering ghost chat", "model", model, "url", url)

	chatModel := ui.NewChatModel(cmd.Context(), url, model, systemPrompt, logger)
	program := tea.NewProgram(chatModel)

	_, err := program.Run()

	return err
}
