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

	"github.com/joho/godotenv"
	"github.com/theantichris/ghost/internal/app"
	"github.com/theantichris/ghost/internal/llm"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load environment variables from .env file if it exists
	err := godotenv.Load()
	if err != nil {
		slog.Info(".env file not found, proceeding with existing environment variables")
	}
	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")

	// Parse command-line flags
	defaultModel := flag.String("model", os.Getenv("DEFAULT_MODEL"), "LLM model to use (overrides DEFAULT_MODEL env var)")
	flag.Parse()

	// Set up HTTP client with timeout
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Initialize Ollama LLM client
	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, *defaultModel, httpClient, logger)
	if err != nil {
		logger.Error("failed to create Ollama client", slog.Any("error", err.Error()))
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the main application
	app, err := app.New(ctx, llmClient, logger)
	if err != nil {
		logger.Error("failed to create app", slog.String("error", err.Error()))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// TODO: Run app
	app.Run()
}
