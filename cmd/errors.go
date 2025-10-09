package cmd

import "errors"

var (
	// ErrConfig indicates a configuration binding or initialization failure.
	ErrConfig = errors.New("failed to bind config")

	// ErrLogging indicates a logging setup or file operation failure.
	ErrLogging = errors.New("failed to setup logging")

	// ErrLLM indicates an LLM client initialization or request processing failure.
	ErrLLM = errors.New("failed to process LLM request")

	// ErrHomeDir indicates a failure to retrieve the user's home directory path.
	ErrHomeDir = errors.New("failed to get home directory")
)
