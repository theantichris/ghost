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
	fmt.Printf("host: %s\n", host)
	fmt.Printf("model: %s\n", model)
	fmt.Printf("system: %s\n", system)

	// Get Ollama host URL
	// Get model
	// Get system prompt
	// Get user prompt
	// Create request body
	// Send to chat endpoint
	// Print response
}
