package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type loggerKey struct{}

var ErrLogger = errors.New("failed to create logger")

// initLogger creates and configures the application logger with JSON formatting
// and file output.
// The log is written to ~/.config/ghost/ghost.log and includes caller information
// and timestamps.
// Returns ErrLogger wrapped with the underlying error if initialization fails.
func initLogger() (*log.Logger, func() error, error) {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFilePath := filepath.Join(home, ".config", "ghost", "ghost.log")

	logDir := filepath.Dir(logFilePath)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLogger, err)
	}

	logger.SetOutput(logFile)

	cleanup := func() error {
		return logFile.Close()
	}

	return logger, cleanup, nil
}
