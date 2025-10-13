package cmd

import "errors"

var (
	// ErrOutput indicates the application couldn't write output.
	ErrOutput = errors.New("failed to write output")

	// ErrNoPrompt indicates a prompt wasn't given as a CLI argument.
	ErrNoPrompt = errors.New("prompt not found")
)
