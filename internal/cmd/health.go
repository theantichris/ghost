package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/theantichris/ghost/internal/llm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

// health prints information on application health and configuration.
var health = func(ctx context.Context, cmd *cli.Command) error {
	output := cmd.Root().Metadata["output"].(io.Writer)

	_, _ = fmt.Fprint(output, "checking ghost health...\n")

	configFile := cmd.Root().Metadata["configFile"].(altsrc.StringPtrSourcer)
	if configFile.SourceURI() == "" {
		_, _ = fmt.Fprint(output, "config file not found\n")
	} else {
		_, _ = fmt.Fprint(output, "config file loaded\n")
	}

	llmClient := cmd.Root().Metadata["llmClient"].(llm.LLMClient)
	version, err := llmClient.Version(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(output, "failed to reach Ollama API: %v\n", err)
	} else {
		_, _ = fmt.Fprintf(output, "Ollama version %s\n", version)
	}

	// Check model

	return nil
}
