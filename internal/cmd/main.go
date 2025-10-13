package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

// TODO: setup config files
// TODO: Ollama URL, default model, system prompt
// TODO: handle LLM request and response

// Run executes the root command (ghost) printing out a test string.
func Run(ctx context.Context, args []string, output io.Writer, logger *log.Logger) error {
	var prompt string

	cmd := &cli.Command{
		Name:      commands["ghost"].Name,
		Usage:     commands["ghost"].Usage,
		ArgsUsage: "[prompt]",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "prompt",
				Destination: &prompt,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if prompt == "" {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			if _, err := fmt.Fprintln(output, prompt); err != nil {
				return fmt.Errorf("%w: %w", ErrOutput, err)
			}

			return nil
		},
	}

	return cmd.Run(ctx, args)
}
