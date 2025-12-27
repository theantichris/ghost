package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/theantichris/ghost/internal/llm"
)

const (
	host   = "http://localhost:11434/api"
	model  = "dolphin-mixtral:8x7b"
	system = "You are ghost, a cyberpunk AI assistant."
)

var (
	errPromptNotDetected = errors.New("prompt not detected")
)

func main() {
	prompt, err := getPrompt(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	messages := createMessages(system, prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	chatResponse, err := llm.Chat(ctx, host, model, messages)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(chatResponse.Content)
}

func getPrompt(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("%w", errPromptNotDetected)
	}

	return args[1], nil
}

func createMessages(system, prompt string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: prompt},
	}

	return messages
}
