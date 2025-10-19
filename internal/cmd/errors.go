package cmd

import (
	"errors"

	"github.com/theantichris/ghost/internal/exitcode"
)

var (
	// ErrOutput indicates the application couldn't write output.
	ErrOutput = exitcode.New(errors.New("failed to write output"), exitcode.ExIOErr)

	// ErrNoPrompt indicates a prompt wasn't given as a CLI argument.
	ErrNoPrompt = exitcode.New(errors.New("no prompt provided"), exitcode.ExNoInput)
)
