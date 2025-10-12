// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"log"
	"os"

	"github.com/theantichris/ghost/internal/cmd"
)

// main initializes and executes the Ghost CLI application using the fang framework.
func main() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
