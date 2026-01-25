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
	"github.com/theantichris/ghost/v3/internal/llm"
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/internal/ui"
	"github.com/theantichris/ghost/v3/theme"
)

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

	ErrImageAnalysis = errors.New("visual recon failed")
	ErrPipedInput    = errors.New("data stream interrupted")
	ErrStreamDisplay = errors.New("output buffer overrun")
	ErrRender        = errors.New("rendering matrix collapsed")
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

	cmd.AddCommand(newChatCommand())

	return cmd, loggerCleanup, err
}

// run is called when the root command is executed.
// It collects the user prompt, any piped input, and flags.
// It initializes the message history, sends the request to the LLM, and prints
// the response.
func run(cmd *cobra.Command, args []string) error {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	url := viper.GetString("url")
	model := viper.GetString("model")
	format := strings.ToLower(viper.GetString("format"))
	userPrompt := args[0]

	messages := initMessages(systemPrompt, userPrompt, format)

	registry := registerTools(logger)

	pipedInput, err := getPipedInput(os.Stdin, logger)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPipedInput, err)
	}

	if pipedInput != "" {
		pipedMessage := llm.ChatMessage{
			Role:    llm.RoleUser,
			Content: pipedInput,
		}

		messages = append(messages, pipedMessage)
	}

	streamModel := ui.NewStreamModel(format)
	streamProgram := tea.NewProgram(streamModel)

	go func() {
		imageAnalysis, err := analyzeImages(cmd, url)
		if err != nil {
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
			return
		}

		messages = append(messages, imageAnalysis...)

		logger.Info("establishing neural link", "model", model, "url", url, "format", format)

		tools := registry.Definitions()

		if len(tools) > 0 {
			for {
				resp, err := llm.Chat(cmd.Context(), url, model, messages, tools)
				if err != nil {
					streamProgram.Send(ui.StreamErrorMsg{Err: err})
					return
				}

				if len(resp.ToolCalls) == 0 {
					break
				}

				messages = append(messages, resp)

				for _, toolCall := range resp.ToolCalls {
					logger.Debug("executing tool", "name", toolCall.Function.Name)

					result, err := registry.Execute(cmd.Context(), toolCall.Function.Name, toolCall.Function.Arguments)
					if err != nil {
						logger.Error("tool execution failed", "name", toolCall.Function.Name, "error", err)
						result = fmt.Sprintf("error: %s", err.Error())
					}

					messages = append(messages, llm.ChatMessage{Role: llm.RoleTool, Content: result})
				}
			}
		}

		if _, err = llm.StreamChat(cmd.Context(), url, model, messages, nil, func(chunk string) {
			streamProgram.Send(ui.StreamChunkMsg(chunk))
		}); err != nil {
			logger.Error("neural link severed", "error", err, "model", model, "url", url)
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
		} else {
			streamProgram.Send(ui.StreamDoneMsg{})
		}
	}()

	returnedModel, err := streamProgram.Run()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStreamDisplay, err)
	}

	finalModel := returnedModel.(ui.StreamModel)
	if finalModel.Err != nil {
		return finalModel.Err
	}

	render, err := theme.RenderContent(finalModel.Content(), format, isTTY)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRender, err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), render)

	return nil
}

// analyzeImages sends a request to the model to analyze images and returns a
// slice of chat messages with the reports.
func analyzeImages(cmd *cobra.Command, url string) ([]llm.ChatMessage, error) {
	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	visionModel := viper.GetString("vision.model")
	var imageAnalysis []llm.ChatMessage

	imagePaths, err := cmd.Flags().GetStringArray("image")
	if err != nil {
		return []llm.ChatMessage{}, err
	}

	for _, image := range imagePaths {
		filename := filepath.Base(image)
		logger.Debug("digitizing visual data", "filename", filename)

		encodedImage, err := encodeImage(image)
		if err != nil {
			return []llm.ChatMessage{}, err
		}

		prompt := fmt.Sprintf("Filename: %s\n\n%s", filename, visionPrompt)
		messages := initMessages(visionSystemPrompt, prompt, "markdown")
		messages[len(messages)-1].Images = []string{encodedImage} // Attach images to prompt message.

		logger.Info("initiating visual recon", "model", visionModel, "url", url, "filename", filename, "format", "markdown")

		response, err := llm.AnalyzeImages(cmd.Context(), url, visionModel, messages)
		if err != nil {
			return []llm.ChatMessage{}, err
		}

		logger.Debug("visual recon complete", "filename", filename, "analysis", response.Content)

		imageAnalysis = append(imageAnalysis, llm.ChatMessage{Role: llm.RoleUser, Content: response.Content})
	}

	return imageAnalysis, nil
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
		return "", fmt.Errorf("%w: %w", ErrPipedInput, err)
	}

	input := strings.TrimSpace(string(pipedInput))

	if len(input) > 0 {
		logger.Debug("intercepted data stream", "size_bytes", len(input))
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

// encodedImage takes an image path and returns a base64 encoded string.
func encodeImage(image string) (string, error) {
	imageBytes, err := os.ReadFile(image)
	if err != nil {
		return "", fmt.Errorf("%w: failed to read %s: %w", ErrImageAnalysis, image, err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)

	return encodedImage, nil
}

// registerTools creates and returns a new tool.Registry after registering tools.
func registerTools(logger *log.Logger) tool.Registry {
	registry := tool.NewRegistry()

	if tavilyAPIKey := viper.GetString("search.api-key"); tavilyAPIKey != "" {
		maxResults := viper.GetInt("search.max-results")
		if maxResults == 0 {
			maxResults = 5
		}

		registry.Register(tool.NewSearch(tavilyAPIKey, maxResults))
		logger.Debug("tool registered", "name", "web_search")
	}

	return registry
}
