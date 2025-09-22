package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

const (
	systemPrompt            = "You are Ghost, a cyberpunk insprired terminal based assistant. Answer requests directly and briefly."
	interactiveSystemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Greet the user once at the start of the conversation then answer requests directly and briefly. Ask for clarification when needed."
	userLabel               = "\nUser: "
	ghostLabel              = "\nGhost: "
	exitCommand             = "/bye"
)

var (
	interactive bool
	noNewLine   bool
)

// chatCmd represents the command called with chat.
var chatCmd = &cobra.Command{
	Use:   "chat [query]",
	Short: "Chat with Ghost.",
	Long: `Chat with Ghost.

		Examples:
			# Quick question from command line
			ghost chat "What is the capital of France?"

			# Pipe input to Ghost
			cat code.go | ghost chat "Explain this code`,
	RunE: runChat,
	Args: cobra.ArbitraryArgs,
}

// init initializes the chat command.
func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().BoolVarP(&noNewLine, "no-newline", "n", false, "Don't add newline after response (useful for scripts)")
	chatCmd.Flags().Duration("timeout", 2*time.Minute, "HTTP timeout for LLM requests")

	viper.BindPFlag("timeout", chatCmd.Flags().Lookup("timeout"))
}

// runChat initializes the LLM Client, sends the query to the LLM and returns the response.
func runChat(cmd *cobra.Command, args []string) error {
	llmClient, err := initializeLLMClient()
	if err != nil {
		return err
	}

	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var query string

	if isPiped {
		query, err = readPipedInput()
		if err != nil {
			return fmt.Errorf("failed to read piped input: %w", err)
		}

		if len(args) > 0 {
			query = query + "\n\n" + strings.Join(args, "")
		}
	} else if len(args) > 0 {
		query = strings.Join(args, " ")
	} else {
		return fmt.Errorf("no input, provide a query or pipe input")
	}

	return runSingleQuery(llmClient, query)
}

// initializeLLMClient validates the client config, creates, and returns the client.
func initializeLLMClient() (*llm.OllamaClient, error) {
	ollamaBaseURL := viper.GetString("ollama_base_url")
	model := viper.GetString("model")
	timeout := viper.GetDuration("timeout")

	if ollamaBaseURL == "" {
		return nil, fmt.Errorf("Ollama base URL not set, set it via OLLAMA_BASE_URL environment variable or config file")
	}

	if model == "" {
		return nil, fmt.Errorf("model not set, set it via DEFAULT_MODEL environment variable, config file, or --model flag")

	}

	Logger.Debug("Initializing LLM client", slog.String("component", "chatCmd"), slog.String("model", model), slog.String("base_url", ollamaBaseURL))

	httpClient := &http.Client{
		Timeout: timeout,
	}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, model, httpClient, Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM Client: %w", err)
	}

	return llmClient, nil
}

// readPipedInput reads input piped from the CLI.
func readPipedInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					lines = append(lines, line)
				}

				break
			}

			return "", err
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, ""), nil
}

// runSingleQuery sends the query to the LLM client and prints the response.
func runSingleQuery(llmClient *llm.OllamaClient, query string) error {
	ctx := context.Background()

	chatHistory := []llm.ChatMessage{
		{Role: llm.System, Content: systemPrompt},
		{Role: llm.User, Content: query},
	}

	response, err := llmClient.Chat(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	if noNewLine {
		fmt.Print(response.Content)
	} else {
		fmt.Println(response.Content)
	}

	return nil
}
