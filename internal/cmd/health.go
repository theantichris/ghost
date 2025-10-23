package cmd

import (
	"context"
	"fmt"
	"io"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

// health prints information on application health and configuration.
var health = func(ctx context.Context, cmd *cli.Command) error {
	output := cmd.Root().Metadata["output"].(io.Writer)

	_, _ = fmt.Fprintln(output, "checking ghost health...")

	configFile := cmd.Root().Metadata["configFile"].(altsrc.StringPtrSourcer)
	if configFile.SourceURI() == "" {
		_, _ = fmt.Fprintln(output, "config file not found")
	} else {
		_, _ = fmt.Fprintln(output, "config file loaded")
	}

	// Check API
	// Check model

	return nil
}
