package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/storage"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/internal/ui"
)

var ErrHomeDir = errors.New("failed to retrieve user home directory")

func newChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chat",
		Short:   "starts ghost in chat mode",
		Long:    "starts ghost in chat mode, use :q to quit",
		Example: "ghost chat",
		Args:    cobra.NoArgs,
		RunE:    runChat,
	}

	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	tavilyAPIKey := viper.GetString("search.api-key")
	maxResults := viper.GetInt("search.max-results")

	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error(ErrHomeDir.Error(), "error", err)

			return fmt.Errorf("%w: %w", ErrHomeDir, err)
		}

		dataDir = filepath.Join(homeDir, ".local", "share")
	}

	storeDir := filepath.Join(dataDir, "ghost")
	store, err := storage.NewStore(storeDir)
	if err != nil {
		logger.Error("failed to create store", "path", storeDir, "error", err)

		return err
	}

	prompts := cmd.Context().Value(promptKey{}).(agent.Prompt)

	config := ui.ModelConfig{
		Context:   cmd.Context(),
		Logger:    logger,
		URL:       viper.GetString("url"),
		ChatLLM:   viper.GetString("model"),
		VisionLLM: viper.GetString("vision.model"),
		System:    prompts.System,
		Registry:  tool.NewRegistry(tavilyAPIKey, maxResults, logger),
		Store:     store,
	}

	chatModel := ui.NewChatModel(config)

	logger.Info("entering chat", "ollama_url", config.URL, "chat_model", config.ChatLLM, "vision_model", config.VisionLLM)
	program := tea.NewProgram(chatModel)
	_, err = program.Run()

	return err
}
