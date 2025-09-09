package main

import (
	"github.com/joho/godotenv"
	"github.com/theantichris/ghost/internal/app"
)

func main() {
	// Load environment variables from .env file if it exists
	godotenv.Load() // load .env file if it exists

	// Load CLI flags

	// Create llm client

	// Create app
	_ = app.New()

	// Run app
}
