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

func main() {
	// Get Ollama host URL
	fmt.Printf("host: %s\n", host)

	// Get model
	fmt.Printf("model: %s\n", model)

	// Get system prompt
	fmt.Printf("system: %s\n", system)

	// Get user prompt
	args := os.Args

	if len(args) < 2 {
		fmt.Println("you must send a prompt to ghost")
		os.Exit(1)
	}

	prompt := args[1]

	fmt.Printf("Prompt: %s\n", prompt)

	// Create request body
	// Send to chat endpoint
	// Print response
}
