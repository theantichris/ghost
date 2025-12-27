package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

const (
	Version = "dev"
	host    = "http://localhost:11434/api"
	model   = "dolphin-mixtral:8x7b"
	system  = "You are ghost, a cyberpunk AI assistant."
)

var errPromptNotDetected = errors.New("prompt not detected")

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ghost <prompt>",
	Short: "A cyberpunk AI assistant powered by Ollama",
	Long: `Ghost is a local cyberpunk AI assistant.
Send prompts directly or pipe data through for analysis.`,
	Example: `  ghost "explain this code" < main.go
	cat error.log | ghost "what's wrong here"
	ghost "tell me a joke"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(cmd.ErrOrStderr(), errPromptNotDetected)
			os.Exit(1)
		}

		prompt := args[0]

		pipedInput, err := getPipedInput()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
		}

		if pipedInput != "" {
			prompt = fmt.Sprintf("%s\n\n%s", prompt, pipedInput)
		}

		messages := initMessages(system, prompt)

		_, err = llm.Chat(cmd.Context(), host, model, messages, onChunk)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
		}

		fmt.Fprintln(cmd.OutOrStdout())
	},
}

// init defines flags and configuration settings.
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/ghost/config.toml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath("ghost")
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
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
