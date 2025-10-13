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
	cmd := &cli.Command{
		Name:  commands["ghost"].Name,
		Usage: commands["ghost"].Usage,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			if _, err := fmt.Fprintln(output, cmd.Args().Get(0)); err != nil {
				return fmt.Errorf("%w: %w", ErrOutput, err)
			}

			return nil
		},
	}

	return cmd.Run(ctx, args)
}
