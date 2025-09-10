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
	godotenv.Load() // load .env file if it exists

	// Load CLI flags

	// Create llm client
	llmClient, err := llm.NewOllamaClient("http://localhost:11434", "llama2")
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

	// Create app
	_, err = app.New(llmClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Run app
}
