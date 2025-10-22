package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

// health prints information on application health and configuration.
var health = func(ctx context.Context, cmd *cli.Command) error {
	output := cmd.Root().Metadata["output"].(io.Writer)

	if _, err := fmt.Fprintln(output, "ghost is healthy"); err != nil {
		return fmt.Errorf("%w: %w", ErrOutput, err)
	}

	return nil
}
