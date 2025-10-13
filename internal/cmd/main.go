package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

var ErrOutput = errors.New("failed to write output")
var ErrNoPrompt = errors.New("prompt not found")

// command holds information about the application commands.
type command struct {
	Name  string
	Usage string
}

// commandList is a map of commands and their information.
type commandList map[string]command

var commands = commandList{
	"ghost": {Name: "ghost", Usage: "send ghost a prompt"},
}

// Run executes the root command (ghost) printing out a test string.
func Run(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("%w", ErrNoPrompt)
	}
	cmd := &cli.Command{
		Name:  commands["ghost"].Name,
		Usage: commands["ghost"].Usage,
		Action: func(context.Context, *cli.Command) error {
			if _, err := fmt.Fprintln(output, args[1]); err != nil {
				return fmt.Errorf("%w: %w", ErrOutput, err)
			}

			return nil
		},
	}

	return cmd.Run(ctx, args)
}
