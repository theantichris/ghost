package main

import (
	"fmt"
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
	// Create request body
	// Send to chat endpoint
	// Print response
}
