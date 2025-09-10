package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/theantichris/ghost/internal/app"
	"github.com/theantichris/ghost/internal/llm"
)

func main() {
	godotenv.Load()
	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")
	defaultModel := os.Getenv("DEFAULT_MODEL")

	// TODO: load CLI flags

	llmClient, err := llm.NewOllamaClient(ollamaBaseURL, defaultModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		if errors.Is(err, llm.ErrURLEmpty) {
			os.Exit(2)
		}

		if errors.Is(err, llm.ErrModelEmpty) {
			os.Exit(3)
		}

		os.Exit(1)
	}

	_, err = app.New(llmClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// TODO: Run app
}
