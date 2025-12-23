// Package main is the entry point for the Ghost CLI application.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/cmd"
	"github.com/theantichris/ghost/internal/exitcode"
)

// version holds the application version, defaults to "dev".
var version = "dev"

// ErrLogger indicates a failure to create or initialize the application logger.
var ErrLogger = exitcode.New(errors.New("failed to create logger"), exitcode.ExDefault)

// main initializes and executes the root command (ghost).
func main() {
	logger, err := initLogger()
	if err != nil {
		fmt.Printf("failed to initialize logger: %s\n", err.Error())
		os.Exit(int(exitcode.ExSoftware))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := cmd.Run(ctx, os.Args, version, os.Stdout, logger); err != nil {
		exitCode := exitcode.GetExitCode(err)

		fmt.Printf("failed to run ghost: %s\n", err)
		logger.Error("failed to run ghost", "error", err)
		os.Exit(int(exitCode))
	}
}

// initLogger creates and configures the application logger with JSON formatting and file output.
// The log is written to ~/.config/ghost/ghost.log and includes caller information and timestamps.
// Returns ErrLogger wrapped with the underlying error if initialization fails.
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

	// logger will be closed automatically by the OS on application exit.
	logger.SetOutput(logFile)

	return logger, nil
}
