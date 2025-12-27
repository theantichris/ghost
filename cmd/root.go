package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

const (
	host   = "http://localhost:11434/api"
	model  = "dolphin-mixtral:8x7b"
	system = "You are ghost, a cyberpunk AI assistant."
)

var errPromptNotDetected = errors.New("prompt not detected")

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghost",
	Short: "A cyberpunk AI assistant",
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		_, err = llm.Chat(ctx, host, model, messages, onChunk)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
		}

		fmt.Fprintln(cmd.OutOrStdout())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/ghost/config.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

func initMessages(system, prompt string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: prompt},
	}

	return messages
}

func onChunk(chunk string) {
	fmt.Fprint(os.Stdout, chunk)
}
