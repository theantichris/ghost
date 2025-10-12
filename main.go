// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"log"
	"os"

	"github.com/theantichris/ghost/internal/cmd"
)

// main initializes and executes the root command (ghost).
func main() {
	if err := cmd.Run(context.Background(), os.Args, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
