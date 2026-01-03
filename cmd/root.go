package cmd

import (
	"context"
	"encoding/base64"
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
)

var (
	isTTY = term.IsTerminal(os.Stdout.Fd())

	ErrNoModel       = errors.New("model is required (set via --model flag, config file, or environment)")
	ErrInvalidFormat = errors.New("invalid format option, valid options are json or markdown")
	ErrLogger        = errors.New("failed to create logger")
	ErrImageAnalysis = errors.New("failed to analyze images")
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
	cmd.PersistentFlags().StringArrayP("image", "i", []string{}, "path to image file(s) (can be specified multiple times)")
	cmd.PersistentFlags().StringP("model", "m", "", "chat model to use")
	cmd.PersistentFlags().StringP("url", "u", "http://localhost:11434/api", "url to the Ollama API")
	cmd.PersistentFlags().StringP("vision-model", "V", "", "vision model to use")

	return cmd, loggerCleanup, err
}

// run is called when the root command is executed.
// It collects the user prompt, any piped input, and flags.
// It initializes the message history, sends the request to the LLM, and prints
// the response.
func run(cmd *cobra.Command, args []string) error {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	// Get config
	url := viper.GetString("url")
	model := viper.GetString("model")
	format := strings.ToLower(viper.GetString("format"))
	userPrompt := args[0]

	// Create message history
	messages := initMessages(systemPrompt, userPrompt, format)

	// Add piped input
	pipedInput, err := getPipedInput(os.Stdin, logger)
	if err != nil {
		return err
	}

	if pipedInput != "" {
		pipedMessage := llm.ChatMessage{
			Role:    llm.RoleUser,
			Content: pipedInput,
		}

		messages = append(messages, pipedMessage)
	}

	// Add image analysis
	imagePaths, err := cmd.Flags().GetStringArray("image")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrImageAnalysis, err)
	}

	if len(imagePaths) > 0 {
		imageAnalysis, err := analyzeImages(cmd, url, imagePaths)
		if err != nil {
			return err
		}

		messages = append(messages, imageAnalysis)
	}

	// Send request
	streamModel := ui.NewStreamModel(format)
	streamProgram := tea.NewProgram(streamModel)

	logger.Info("sending chat request", "model", model, "url", url, "format", format, "has_piped_input", pipedInput != "")

	go func() {
		_, err := llm.StreamChat(cmd.Context(), url, model, messages, func(chunk string) {
			streamProgram.Send(ui.StreamChunkMsg(chunk))
		})

		if err != nil {
			logger.Error("llm request failed", "error", err, "model", model, "url", url)
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
		} else {
			streamProgram.Send(ui.StreamDoneMsg{})
		}
	}()

	// Handle response
	returnedModel, err := streamProgram.Run()
	if err != nil {
		return err
	}

	finalModel := returnedModel.(ui.StreamModel)
	if finalModel.Err != nil {
		return finalModel.Err
	}

	content := finalModel.Content()
	render, err := theme.RenderContent(content, format, isTTY)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render)

	return nil
}

// analyzeImages sends a request to the model to analyze images and returns a
// chat message with the report.
func analyzeImages(cmd *cobra.Command, url string, imagePaths []string) (llm.ChatMessage, error) {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	logger.Debug("encoding images", "count", len(imagePaths))

	visionModel := viper.GetString("vision.model")

	if visionModel == "" {
		return llm.ChatMessage{}, fmt.Errorf("%w: no vision model", ErrImageAnalysis)
	}

	encodedImages, err := encodeImages(imagePaths)
	if err != nil {
		return llm.ChatMessage{}, err
	}

	messages := initMessages(visionSystemPrompt, visionPrompt, "markdown")
	messages[len(messages)-1].Images = encodedImages // Attach images to prompt message.

	logger.Info("starting image analysis request", "model", visionModel, "url", url, "image_count", len(encodedImages), "format", "markdown")

	response, err := llm.AnalyzeImages(cmd.Context(), url, visionModel, messages)
	if err != nil {
		return llm.ChatMessage{}, err
	}

	imageAnalysis := llm.ChatMessage{
		Role:    llm.RoleTool,
		Content: response.Content,
	}

	return imageAnalysis, nil
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

	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}

		logger.Debug("no config file found, using flags/env only")
	} else {
		logger.Debug("loaded config", "file", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	model := viper.GetString("model")
	if model == "" {
		return ErrNoModel
	}

	err = validateFormat(viper.GetString("format"))
	if err != nil {
		return err
	}

	_ = viper.BindPFlag("vision.model", cmd.Flags().Lookup("vision-model"))

	return nil
}

// getPipedInput detects, reads, and returns any input piped to the command.
func getPipedInput(file *os.File, logger *log.Logger) (string, error) {
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

	if len(input) > 0 {
		logger.Debug("received piped input", "size_bytes", len(input))
	}

	return input, nil
}

// initMessages creates and returns an initial message history.
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

// encodedImages takes a slice of paths and returns a slice of base64 encoded strings.
func encodeImages(paths []string) ([]string, error) {
	if len(paths) < 1 {
		return []string{}, nil
	}

	encodedImages := make([]string, 0, len(paths))

	for _, path := range paths {
		imageBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read %s: %w", ErrImageAnalysis, path, err)
		}

		encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
		encodedImages = append(encodedImages, encodedImage)
	}

	return encodedImages, nil
}
