// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/cmd"
)

// main initializes and executes the root command (ghost).
func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	if err := cmd.Run(context.Background(), os.Args, os.Stdout, logger); err != nil {
		log.Fatal(err)
	}
}
