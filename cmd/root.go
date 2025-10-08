package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd creates and returns the root command for the Ghost CLI application.
// It sets up persistent flags for configuration, debug mode, model selection, and API settings.
func NewRootCmd(logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghost",
		Short: "A cyberpunk inspired AI assistant.",
		Long:  "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlag("ollama", cmd.PersistentFlags().Lookup("ollama")); err != nil {
				return fmt.Errorf("%w: %w", ErrConfig, err)
			}

			logger.Debug("bound persistent flag", "flag", "ollama", "value", viper.GetString("ollama"))

			if err := viper.BindPFlag("model", cmd.PersistentFlags().Lookup("model")); err != nil {
				return fmt.Errorf("%w: %w", ErrConfig, err)
			}

			logger.Debug("bound persistent flag", "flag", "model", "value", viper.GetString("model"))

			if err := viper.BindPFlag("timeout", cmd.PersistentFlags().Lookup("timeout")); err != nil {
				return fmt.Errorf("%w: %w", ErrConfig, err)
			}

			logger.Debug("bound persistent flag", "flag", "timeout", "value", viper.GetString("timeout"))

			return nil
		},
	}

	cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.config/ghost/config.toml)")
	cmd.PersistentFlags().String("model", "", "LLM model to use")
	cmd.PersistentFlags().String("ollama", "", "Ollama API base URL")
	cmd.PersistentFlags().Duration("timeout", 2*time.Minute, "HTTP client timeout for LLM requests")

	cmd.AddCommand(NewAskCmd(logger))
	cmd.AddCommand(NewChatCmd(logger))

	return cmd
}

// Execute creates and returns the root command for use with fang.Execute.
// It sets up the logger and registers the configuration initialization.
func Execute() *cobra.Command {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.WarnLevel,
	})

	cobra.OnInitialize(func() {
		initConfig(logger)
	})

	return NewRootCmd(logger)
}

// initConfig initializes the configuration for the Ghost CLI application by loading
// environment variables from .env file, setting up viper configuration paths, binding
// environment variables (OLLAMA_BASE_URL to ollama, DEFAULT_MODEL to model), and
// attempting to read the config file from multiple locations.
func initConfig(logger *log.Logger) {
	if err := godotenv.Load(); err != nil {
		logger.Debug(".env file not found, using environment variables")
	} else {
		logger.Debug(".env file loaded successfully")
	}

	config := viper.GetString("config")

	if config != "" {
		viper.SetConfigFile(config)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Error(ErrHomeDir.Error(), "error", err)
			logger.Debug("skipping home directory config path")
		} else {
			viper.AddConfigPath(filepath.Join(home, ".config", "ghost"))
			viper.SetConfigName("config")
			viper.SetConfigType("toml")
		}
	}

	viper.AutomaticEnv()
	if err := viper.BindEnv("ollama", "OLLAMA_BASE_URL"); err != nil {
		logger.Error("failed to bind ollama config", "error", err)
	}

	logger.Debug("bound environment variable", "var", "OLLAMA_BASE_URL")

	if err := viper.BindEnv("model", "DEFAULT_MODEL"); err != nil {
		logger.Error("failed to bind model config", "error", err)
	}

	logger.Debug("bound environment variable", "var", "DEFAULT_MODEL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debug("config file not found")
		} else {
			logger.Error("error loading config file", "error", err)
		}
	} else {
		logger.Debug("using config file", "file", viper.ConfigFileUsed())
	}

	logger.Debug("configuration loaded successfully", "ollama", viper.GetString("ollama"), "model", viper.GetString("model"), "debug", viper.GetBool("debug"))

	if err := initLogger(logger); err != nil {
		logger.Error("failed to setup file logging, continuing with stderr", "error", err)
	}
}

// initLogger configures file logging to ~/.config/ghost/ghost.log for debug output.
func initLogger(logger *log.Logger) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%w: failed to get home directory: %w", ErrLogging, err)
	}

	logFilePath := filepath.Join(home, ".config", "ghost", "ghost.log")

	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("%w: failed to create log directory: %w", ErrLogging, err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("%w: failed to open log file: %w", ErrLogging, err)
	}

	// logFile is intentionally not closed here as the logger needs to write to
	// it for the program's lifecycle. The file will be closed by the OS when the
	// program exits.
	logger.SetOutput(logFile)

	logger.SetLevel(log.DebugLevel)

	return nil
}
