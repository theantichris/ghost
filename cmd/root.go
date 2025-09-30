package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrRootCmd = errors.New("failed to run ghost command")

// NewRootCmd creates and returns the root command for the Ghost CLI application.
// It sets up persistent flags for configuration, debug mode, model selection, and API settings.
func NewRootCmd(logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghost",
		Short: "A cyberpunk inspired AI assistant.",
		Long:  "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlag("ollama", cmd.PersistentFlags().Lookup("ollama")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			logger.Debug("bound persistent flag", "flag", "ollama", "value", viper.GetString("ollama"))

			if err := viper.BindPFlag("model", cmd.PersistentFlags().Lookup("model")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			logger.Debug("bound persistent flag", "flag", "model", "value", viper.GetString("model"))

			if err := viper.BindPFlag("log_file", cmd.PersistentFlags().Lookup("log-file")); err != nil {
				return fmt.Errorf("%w: %s", ErrRootCmd, err)
			}

			logger.Debug("bound persistent flag", "flag", "log_file", "value", viper.GetString("log_file"))

			return nil
		},
	}

	cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.ghost/config.toml)")
	cmd.PersistentFlags().String("model", "", "LLM model to use")
	cmd.PersistentFlags().String("ollama", "", "Ollama API base URL")
	cmd.PersistentFlags().String("log-file", "", "log file path (default: ~/.ghost/ghost.log)")

	cmd.AddCommand(NewAskCmd(logger))

	return cmd
}

// Execute creates and returns the root command for use with fang.Execute.
// It sets up the logger and registers the configuration initialization.
func Execute() *cobra.Command {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	cobra.OnInitialize(func() {
		initConfig(logger)
	})

	return NewRootCmd(logger)
}

// initConfig initializes the configuration for the Ghost CLI application.
// It loads environment variables from .env file, sets up viper configuration paths,
// binds environment variables (OLLAMA_BASE_URL to ollama, DEFAULT_MODEL to model),
// and attempts to read the config file from multiple locations.
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
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".ghost"))
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	if err := viper.BindEnv("ollama", "OLLAMA_BASE_URL"); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)
	}

	logger.Debug("bound environment variable", "var", "OLLAMA_BASE_URL")

	if err := viper.BindEnv("model", "DEFAULT_MODEL"); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)
	}

	logger.Debug("bound environment variable", "var", "DEFAULT_MODEL")

	if err := viper.BindEnv("log_file", "LOG_FILE"); err != nil {
		logger.Error(ErrRootCmd.Error(), "error", err)
	}

	logger.Debug("bound environment variable", "var", "LOG_FILE")

	viper.SetDefault("log_file", "~/.ghost/ghost.log")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debug("config file not found")
		} else {
			logger.Error("error loading config file")
		}
	} else {
		logger.Debug("using config file", "file", viper.ConfigFileUsed())
	}

	logger.Debug("configuration loaded successfully", "ollama", viper.GetString("ollama"), "model", viper.GetString("model"), "debug", viper.GetBool("debug"))

	if err := setupFileLogging(logger); err != nil {
		logger.Error("failed to setup file logging", "error", err)
	}
}

// setupFileLogging configures file logging if a log file path is specified.
func setupFileLogging(logger *log.Logger) error {
	logFilePath := viper.GetString("log_file")
	if logFilePath == "" {
		logger.Debug("file logging disabled (empty path)")
		return nil
	}

	if strings.HasPrefix(logFilePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		logFilePath = filepath.Join(home, logFilePath[2:])
	}

	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(os.Stderr, logFile)
	logger.SetOutput(multiWriter)

	logger.Info("file logging enabled", "path", logFilePath)

	return nil
}
