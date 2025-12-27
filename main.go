package main

import (
	"fmt"
	"os"
)

const (
	host   = "https://localhost:11434"
	model  = "dolphin-mixtral:8x7b"
	system = "You are ghost, a cyberpunk AI assistant."
)

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	// Role holds the author of the message.
	// Values are system, user, assistant, tool.
	Role string `json:"role"`

	// Content holds the message history.
	Content string `json:"content"`
}

func main() {
	// Get Ollama host URL
	fmt.Printf("host: %s\n", host)

	// Get user prompt
	args := os.Args

	if len(args) < 2 {
		fmt.Println("you must send a prompt to ghost")
		os.Exit(1)
	}

	prompt := args[1]

	// Create message history
	messages := []chatMessage{
		{
			Role:    "system",
			Content: system,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Create request body
	chatRequest := chatRequest{
		Model:    model,
		Messages: messages,
	}

	fmt.Printf("request: %v\n", chatRequest)

	// Send to chat endpoint

	// Print response
}
