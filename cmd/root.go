package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/ui"
	"github.com/theantichris/ghost/theme"
)

type loggerKey struct{}

const (
	Version = "dev"

	useText   = "ghost <prompt>"
	shortText = "ghost is a local cyberpunk AI assistant."
	longText  = `Ghost is a local cyberpunk AI assistant.
Send prompts directly or pipe data through for analysis.`
	exampleText = `  ghost "explain this code" < main.go
	cat error.log | ghost "what's wrong here"
	ghost "tell me a joke"`

	systemPrompt   = "You are ghost, a cyberpunk AI assistant."
	jsonPrompt     = "Format the response as json without enclosing backticks."
	markdownPrompt = "Format the response as markdown without enclosing backticks."
)

var (
	isTTY = term.IsTerminal(os.Stdout.Fd())

	ErrNoModel       = errors.New("model is required (set via --model flag, config file, or environment)")
	ErrInvalidFormat = errors.New("invalid format option, valid options are json or markdown")
	ErrLogger        = errors.New("failed to create logger")
)

// NewRootCmd creates and returns the root command.
func NewRootCmd() (*cobra.Command, func() error, error) {
	logger, loggerCleanup, err := initLogger()
	if err != nil {
		return nil, nil, err
	}

	var cfgFile string

	cmd := &cobra.Command{
		Use:     useText,
		Short:   shortText,
		Long:    longText,
		Example: exampleText,
		Args:    cobra.MinimumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(context.WithValue(cmd.Context(), loggerKey{}, logger))

			return initConfig(cmd, cfgFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args)
		},
	}

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	cmd.PersistentFlags().StringP("format", "f", "", "output format (JSON, markdown), unspecified for text")
	cmd.PersistentFlags().StringP("model", "m", "", "chat model to use")
	cmd.PersistentFlags().StringP("url", "u", "http://localhost:11434/api", "url to the Ollama API")

	return cmd, loggerCleanup, err
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command, cfgFile string) error {
	viper.SetEnvPrefix("GHOST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".config", "ghost"))
		viper.SetConfigName("config.toml")
		viper.SetConfigType("toml")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	model := viper.GetString("model")
	if model == "" {
		return ErrNoModel
	}

	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)
	logger.Debug("chat model", "value", model)

	return nil
}

// run is called when the root command is executed.
// It collects the user prompt, any piped input, and flags.
// It initializes the message history, sends the request to the LLM, and prints
// the response.
func run(cmd *cobra.Command, args []string) error {
	userPrompt := args[0]

	pipedInput, err := getPipedInput(os.Stdin)
	if err != nil {
		return err
	}

	if pipedInput != "" {
		userPrompt = fmt.Sprintf("%s\n\n%s", userPrompt, pipedInput)
	}

	format := strings.ToLower(viper.GetString("format"))

	err = validateFormat(format)
	if err != nil {
		return err
	}

	messages := initMessages(systemPrompt, userPrompt, format)

	url := viper.GetString("url")
	model := viper.GetString("model")

	streamModel := ui.NewStreamModel(format)
	streamProgram := tea.NewProgram(streamModel)

	go func() {
		_, err := llm.Chat(cmd.Context(), url, model, messages, func(chunk string) {
			streamProgram.Send(ui.StreamChunkMsg(chunk))
		})

		if err != nil {
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
		} else {
			streamProgram.Send(ui.StreamDoneMsg{})
		}
	}()

	returnedModel, err := streamProgram.Run()
	if err != nil {
		return err
	}

	streamModel = returnedModel.(ui.StreamModel)

	if streamModel.Err != nil {
		return streamModel.Err
	}

	content := streamModel.Content()

	render, err := theme.RenderContent(content, format, isTTY)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render)

	return nil
}

// getPipedInput detects, reads, and returns any input piped to the command.
func getPipedInput(file *os.File) (string, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return "", nil
	}

	if fileInfo.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	pipedInput, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return "", fmt.Errorf("failed to read piped input: %w", err)
	}

	input := strings.TrimSpace(string(pipedInput))

	return input, nil
}

// initMessages creates and returns the initial message history.
func initMessages(system, prompt, format string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: system},
	}

	if format != "" {
		switch format {
		case "json":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: jsonPrompt})
		case "markdown":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: markdownPrompt})
		}
	}

	messages = append(messages, llm.ChatMessage{Role: llm.RoleUser, Content: prompt})

	return messages
}

// validateFormat returns an error if the format flag isn't a valid value.
func validateFormat(format string) error {
	if format != "" && (format != "json" && format != "markdown") {
		return ErrInvalidFormat
	}

	return nil
}

// initLogger creates and configures the application logger with JSON formatting
// and file output.
// The log is written to ~/.config/ghost/ghost.log and includes caller information
// and timestamps.
// Returns ErrLogger wrapped with the underlying error if initialization fails.
func initLogger() (*log.Logger, func() error, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFilePath := filepath.Join(home, ".config", "ghost", "ghost.log")

	logDir := filepath.Dir(logFilePath)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logger.SetOutput(logFile)

	cleanup := func() error {
		return logFile.Close()
	}

	return logger, cleanup, nil
}
