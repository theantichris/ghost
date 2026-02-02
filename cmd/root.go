package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/v3/internal/agent"
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
	visionModel := viper.GetString("vision.model")
	format := strings.ToLower(viper.GetString("format"))
	tavilyAPIKey := viper.GetString("search.api-key")
	maxResults := viper.GetInt("search.max-results")
	userPrompt := args[0]

	images, err := cmd.Flags().GetStringArray("image")
	if err != nil {
		return err
	}

	messages := agent.NewMessageHistory(agent.SystemPrompt, userPrompt, format)

	registry := tool.NewRegistry(tavilyAPIKey, maxResults, logger)

	pipedInput, err := agent.GetPipedInput(os.Stdin, logger)
	if err != nil {
		return fmt.Errorf("%w: %w", agent.ErrPipedInput, err)
	}

	if pipedInput != "" {
		pipedMessage := llm.ChatMessage{
			Role:    llm.RoleUser,
			Content: pipedInput,
		}

		messages = append(messages, pipedMessage)
	}

	streamModel := ui.NewStreamModel(format, logger)

	var programOpts []tea.ProgramOption
	if ttyIn, ttyOut, err := tea.OpenTTY(); err == nil {
		programOpts = append(programOpts, tea.WithInput(ttyIn), tea.WithOutput(ttyOut))
		defer func() { _ = ttyIn.Close() }()
		defer func() { _ = ttyOut.Close() }()
	} else {
		logger.Debug("TTY unavailable, using standard I/O", "error", err)
	}

	streamProgram := tea.NewProgram(streamModel, programOpts...)

	go func() {
		imageAnalysis, err := agent.AnalyseImages(cmd.Context(), url, visionModel, images, logger)
		if err != nil {
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
			return
		}

		messages = append(messages, imageAnalysis...)

		logger.Info("establishing neural link", "model", model, "url", url, "format", format)

		messages, err = agent.RunToolLoop(cmd.Context(), registry, url, model, messages, logger)
		if err != nil {
			streamProgram.Send(ui.StreamErrorMsg{Err: err})
			return
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
