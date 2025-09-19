package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/joho/godotenv"
	"github.com/theantichris/ghost/internal/app"
	"github.com/theantichris/ghost/internal/llm"
)

// main is the entry point for the ghost CLI application.
func main() {
	logger := createLogger()
	logger.Info("ghost CLI starting", slog.String("component", "main"))

	err := godotenv.Load()
	if err != nil {
		logger.Info(".env file not found, proceeding with existing environment variables", slog.String("component", "main"))
	} else {
		logger.Info(".env file loaded successfully", slog.String("component", "main"))
	}

	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")
	defaultModel := flag.String("model", os.Getenv("DEFAULT_MODEL"), "LLM model to use (overrides DEFAULT_MODEL env var)")
	flag.Parse()

	httpClient := &http.Client{Timeout: 0 * time.Second}
	llmClient := createLLMClient(ollamaBaseURL, *defaultModel, httpClient, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ghostApp, err := app.New(llmClient, logger)
	if err != nil {
		logger.Error("failed to create app", slog.String("component", "main"), slog.String("error", err.Error()))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = ghostApp.Run(ctx, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

// createLogger initializes and returns a structured logger.
func createLogger() *slog.Logger {
	logger := slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
	}))
	slog.SetDefault(logger)

	return logger
}

// createLLMClient initializes and returns an Ollama LLM client.
func createLLMClient(ollamaBaseURL, defaultModel string, httpClient *http.Client, logger *slog.Logger) *llm.OllamaClient {
	// TODO: Should return error and let main handle it
	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, defaultModel, httpClient, logger)
	if err != nil {
		logger.Error("failed to create Ollama client", slog.String("error", err.Error()))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		if errors.Is(err, llm.ErrURLEmpty) {
			logger.Error("OLLAMA_BASE_URL is required but not set")
			fmt.Fprintf(os.Stderr, "error: OLLAMA_BASE_URL is required but not set\n")
			os.Exit(2)
		}

		if errors.Is(err, llm.ErrModelEmpty) {
			logger.Error("DEFAULT_MODEL is required but not set")
			fmt.Fprintf(os.Stderr, "error: DEFAULT_MODEL is required but not set\n")
			os.Exit(3)
		}

		os.Exit(1)
	}

	return llmClient
}
