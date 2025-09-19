package main

import (
	"context"
	"errors"
	"flag"
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
	defaultModel := flag.String("model", "", "LLM model to use (overrides DEFAULT_MODEL env var)")
	isDebugMode := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	logger := createLogger(*isDebugMode)
	logger.Info("ghost CLI starting", slog.String("component", "main"))

	loadEnv(logger)

	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")

	if defaultModel == nil || *defaultModel == "" {
		val := os.Getenv("DEFAULT_MODEL")
		defaultModel = &val
	}

	httpClient := &http.Client{Timeout: 0 * time.Second}

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, *defaultModel, httpClient, logger)
	if err != nil {
		if errors.Is(err, llm.ErrURLEmpty) {
			logger.Error("OLLAMA_BASE_URL environment variable is not set", slog.String("component", "main"))
			os.Exit(2)
		}

		if errors.Is(err, llm.ErrModelEmpty) {
			logger.Error("DEFAULT_MODEL environment variable is not set and -model not passed", slog.String("component", "main"))
			os.Exit(3)
		}

		os.Exit(1)
	}

	ghostApp, err := app.New(llmClient, logger)
	if err != nil {
		logger.Error(err.Error(), slog.String("component", "main"))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ghostApp.Run(ctx, os.Stdin)
	if err != nil {
		logger.Error(err.Error(), slog.String("component", "main"))
		os.Exit(1)
	}
}

// loadEnv loads environment variables from a .env file if it exists.
func loadEnv(logger *slog.Logger) {
	err := godotenv.Load()
	if err != nil {
		logger.Info(".env file not found, proceeding with existing environment variables", slog.String("component", "main"))
	} else {
		logger.Info(".env file loaded successfully", slog.String("component", "main"))
	}
}

// createLogger initializes and returns a structured logger.
func createLogger(debugMode bool) *slog.Logger {
	var logLevel = slog.LevelInfo

	if debugMode {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:      logLevel,
		TimeFormat: time.RFC3339,
	}))

	slog.SetDefault(logger)

	return logger
}
