// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/cmd"
)

var ErrLogger = errors.New("failed to create logger")

// main initializes and executes the root command (ghost).
func main() {
	logger, err := initLogger()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Run(context.Background(), os.Args, os.Stdout, logger); err != nil {
		logger.Fatal(err)
	}
}

// initLogger creates and returns the logger.
func initLogger() (*log.Logger, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFilePath := filepath.Join(home, ".config", "ghost", "ghost.log")

	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logger.SetOutput(logFile)

	return logger, nil
}
