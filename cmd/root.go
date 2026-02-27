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
	"github.com/theantichris/ghost/v3/internal/tool"
	"github.com/theantichris/ghost/v3/internal/tui"
	"github.com/theantichris/ghost/v3/style"
)

const Version = "dev"

type promptKey struct{}

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
		Use:   "ghost <prompt>",
		Short: "ghost is a local cyberpunk AI assistant.",
		Long:  "Ghost is a local cyberpunk AI Assistant.\nSend prompts directly or pipe data through for analysis.",
		Example: `  ghost "explain this code" < main.go
	cat error.log | ghost "what's wrong here"
	ghost "tell me a joke"`,
		Args: cobra.MinimumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(context.WithValue(cmd.Context(), loggerKey{}, logger))

			configDir, err := configDir()
			if err != nil {
				return err
			}

			prompts, err := agent.LoadPrompts(configDir, logger)
			if err != nil {
				return err
			}

			cmd.SetContext(context.WithValue(cmd.Context(), promptKey{}, prompts))

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
	prompts := cmd.Context().Value(promptKey{}).(agent.Prompt)

	format := strings.ToLower(viper.GetString("format"))
	images, err := cmd.Flags().GetStringArray("image")
	if err != nil {
		return err
	}

	modelConfig := tui.ModelConfig{
		Context:   cmd.Context(),
		Prompts:   prompts,
		Logger:    logger,
		URL:       viper.GetString("url"),
		ChatLLM:   viper.GetString("model"),
		VisionLLM: viper.GetString("vision.model"),
		Format:    format,
		Images:    images,
		Registry: tool.NewRegistry(
			viper.GetString("search.api-key"),
			viper.GetInt("search.max-results"),
			logger,
		),
	}

	streamModel, err := tui.NewStreamModel(modelConfig, args[0])
	if err != nil {
		return err
	}

	var programOpts []tea.ProgramOption
	if ttyIn, ttyOut, err := tea.OpenTTY(); err == nil {
		programOpts = append(programOpts, tea.WithInput(ttyIn), tea.WithOutput(ttyOut))
		defer func() { _ = ttyIn.Close() }()
		defer func() { _ = ttyOut.Close() }()
	} else {
		logger.Debug("TTY unavailable, using standard I/O", "error", err)
	}

	streamProgram := tea.NewProgram(streamModel, programOpts...)

	returnedModel, err := streamProgram.Run()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStreamDisplay, err)
	}

	// Bubble Tea clears the output once it exits so rerender the content to
	// Stdout.
	finalModel := returnedModel.(tui.StreamModel)
	if finalModel.Err != nil {
		return finalModel.Err
	}

	render, err := style.RenderContent(finalModel.Content(), format, isTTY)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRender, err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), render)

	return nil
}
