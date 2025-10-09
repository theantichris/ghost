package cmd

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theantichris/ghost/internal/llm"
)

// systemPrompt defines the default system-level instruction for Ghost's LLM interactions.
const systemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Answer requests directly and briefly."

// initializeLLMClient creates and configures an LLM client using configuration from viper,
// requiring OLLAMA_BASE_URL and DEFAULT_MODEL to be set via environment variables, config
// file, or command-line flags.
func initializeLLMClient(logger *log.Logger) (llm.LLMClient, error) {
	ollamaBaseURL := viper.GetString("ollama")
	model := viper.GetString("model")

	if ollamaBaseURL == "" {
		return nil, fmt.Errorf("%w: set OLLAMA_BASE_URL via environment variable, config file, or --ollama flag", ErrConfig)
	}

	if model == "" {
		return nil, fmt.Errorf("%w: set DEFAULT_MODEL via environment variable, config file, or --model flag", ErrConfig)
	}

	httpClient := &http.Client{}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, model, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLM, err)
	}

	logger.Info("LLM client initialized successfully")

	return llmClient, nil
}
