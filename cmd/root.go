package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

const (
	Version = "dev"
	system  = "You are ghost, a cyberpunk AI assistant."
)

var (
	cfgFile    string
	ErrNoModel = errors.New("model is required (set via --model flag, config file, or environment)")
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "ghost <prompt>",

	Short: "A cyberpunk AI assistant powered by Ollama",

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
		prompt := args[0]

		pipedInput, err := getPipedInput()
		if err != nil {
			return err
		}

		if pipedInput != "" {
			prompt = fmt.Sprintf("%s\n\n%s", prompt, pipedInput)
		}

		messages := initMessages(system, prompt)

		url := viper.GetString("url")
		model := viper.GetString("model")

		_, err = llm.Chat(cmd.Context(), url, model, messages, onChunk)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout())

		return nil
	},
}

// init defines flags and configuration settings.
func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/ghost/config.toml)")
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
func initMessages(system, prompt string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: prompt},
	}

	return messages
}

// onChunk is the callback to print streaming content.
func onChunk(chunk string) {
	fmt.Fprint(os.Stdout, chunk)
}
