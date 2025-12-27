package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	pipedInput, err := getPipedInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if pipedInput != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, pipedInput)
	}

	messages := initMessages(system, prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err = llm.Chat(ctx, host, model, messages, onChunk)
	if err != nil {
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout)
}

func getPrompt(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("%w", errPromptNotDetected)
	}

	return args[1], nil
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
