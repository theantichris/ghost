package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/MatusOllah/slogcolor"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrRootCmd = errors.New("failed to run ghost command")

// TODO: should this be named ghost and not root?

// NewRootCmd creates and returns the root command for the Ghost CLI application.
// It sets up persistent flags for configuration, debug mode, model selection, and API settings.
func NewRootCmd(logger *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghost",
		Short: "A cyberpunk inspired AI assistant.",
		Long:  "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			if err := viper.BindPFlag("ollama", cmd.PersistentFlags().Lookup("ollama")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			if err := viper.BindPFlag("model", cmd.PersistentFlags().Lookup("model")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			return nil
		},
	}

	// TODO: Should I be binding the config file?
	cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.ghost.toml)")
	cmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	cmd.PersistentFlags().String("model", "", "LLM model to use")
	cmd.PersistentFlags().String("ollama", "", "Ollama API base URL")

	cmd.AddCommand(NewAskCmd(logger))

	return cmd
}

// TODO: Research returning errors in Cobra functions.

// Execute initializes and runs the Ghost CLI application.
// It sets up the logger, configuration, and executes the root command.
// Returns the command for use with fang.Execute or nil on error.
func Execute() *cobra.Command {
	logger := slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level: slog.LevelWarn,
	}))

	cobra.OnInitialize(func() {
		initConfig(logger)
	})

	cmd := NewRootCmd(logger)

	if err := cmd.Execute(); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)

		return nil
	}

	return cmd
}

func initConfig(logger *slog.Logger) {
	if err := godotenv.Load(); err != nil {
		logger.Debug(".env file not found, using environment variables", "component", "cmd.RootCmd")
	} else {
		logger.Debug(".env file loaded successfully", "component", "cmd.RootCmd")
	}

	config := viper.GetString("config")

	if config != "" {
		viper.SetConfigFile(config)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ghost")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	if err := viper.BindEnv("ollama", "OLLAMA_BASE_URL"); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)
	}

	if err := viper.BindEnv("model", "DEFAULT_MODEL"); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debug("config file not found", "component", "cmd.RootCmd")
		} else {
			logger.Error("error loading config file", "error", err)
		}
	} else {
		logger.Debug("using config file", "file", viper.ConfigFileUsed(), "component", "cmd.RootCmd")
	}

	// if viper.GetBool("debug") {
	// TODO: Set debug level once I switch logger.
	// }
}
