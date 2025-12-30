package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/ui"
	"github.com/theantichris/ghost/theme"
)

const (
	Version      = "dev"
	systemPrompt = "You are ghost, a cyberpunk AI assistant."
	jsonPrompt   = "Format the response as json without enclosing backticks."
)

var (
	cfgFile          string
	ErrNoModel       = errors.New("model is required (set via --model flag, config file, or environment)")
	ErrInvalidFormat = errors.New("invalid format option, valid options are json")
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "ghost <prompt>",

	Short: "Ghost is a local cyberpunk AI assistant.",

	Long: `Ghost is a local cyberpunk AI assistant.
Send prompts directly or pipe data through for analysis.`,

	Example: `  ghost "explain this code" < main.go
	cat error.log | ghost "what's wrong here"
	ghost "tell me a joke"`,

	Args: cobra.MinimumNArgs(1),

	// PersistentPreRunE is called after flags are parsed but before the command's
	// RunE function is called.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		userPrompt := args[0]

		pipedInput, err := getPipedInput()
		if err != nil {
			return err
		}

		if pipedInput != "" {
			userPrompt = fmt.Sprintf("%s\n\n%s", userPrompt, pipedInput)
		}

		format := strings.ToLower(viper.GetString("format"))

		if format != "" && format != "json" {
			return ErrInvalidFormat
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

		if format == "json" && term.IsTerminal(os.Stdout.Fd()) {
			content = theme.JSON(content)
		}

		fmt.Fprintln(cmd.OutOrStdout(), content)

		return nil
	},
}

// init defines flags and configuration settings.
func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	RootCmd.PersistentFlags().StringP("format", "f", "", "output format (JSON), unspecified for text")
	RootCmd.PersistentFlags().StringP("model", "m", "", "chat model to use")
	RootCmd.PersistentFlags().StringP("url", "u", "http://localhost:11434/api", "url to the Ollama API")
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
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

	return nil
}

// getPipedInput detects, reads, and returns any input piped to the command.
func getPipedInput() (string, error) {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return "", nil
	}

	if fileInfo.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	pipedInput, err := io.ReadAll(io.LimitReader(os.Stdin, 10<<20))
	if err != nil {
		return "", fmt.Errorf("failed to read piped input: %w", err)
	}

	input := strings.TrimSpace(string(pipedInput))

	return input, nil
}

// initMessages creates and returns the initial message history.
// Returns an error for an invalid output format.
func initMessages(system, prompt, format string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: system},
	}

	if format != "" {
		switch format {
		case "json":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: jsonPrompt})
		}
	}

	messages = append(messages, llm.ChatMessage{Role: llm.RoleUser, Content: prompt})

	return messages
}
