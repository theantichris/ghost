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
	"github.com/theantichris/ghost/internal/exitcode"
)

var version = "dev"

var ErrLogger = errors.New("failed to create logger")

// main initializes and executes the root command (ghost).
func main() {
	logger, err := initLogger()
	if err != nil {
		fmt.Printf("failed to initialize logger: %s\n", err.Error())
		os.Exit(int(exitcode.ExSoftware))
	}

	if err := cmd.Run(context.Background(), os.Args, version, os.Stdout, logger); err != nil {
		exitCode := exitcode.GetExitCode(err)

		fmt.Printf("failed to run ghost: %s\n", err)
		logger.Error("failed to run ghost", "error", err)
		os.Exit(int(exitCode))
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

	// logger will be closed automatically by the OS on application exit.
	logger.SetOutput(logFile)

	return logger, nil
}
