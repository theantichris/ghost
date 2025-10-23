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

	host := cmd.String("host")
	model := cmd.String("model")
	configFile := cmd.Root().Metadata["configFile"].(altsrc.StringPtrSourcer)

	errors := 0

	fmt.Fprint(output, ">> initializing ghost diagnostics...\n\n")

	fmt.Fprintln(output, "SYSTEM CONFIG")
	fmt.Fprintf(output, "  ◆ host: %s\n", host)
	fmt.Fprintf(output, "  ◆ model: %s\n", model)
	fmt.Fprintf(output, "  ◆ config: %s\n\n", configFile.SourceURI())

	fmt.Fprintln(output, "NEURAL LINK STATUS")

	llmClient := cmd.Root().Metadata["llmClient"].(llm.LLMClient)
	version, err := llmClient.Version(ctx)
	if err == nil {
		fmt.Fprintf(output, "  ◆ ollama api CONNECTED [v%s]\n", version)
	} else {
		errors++
		fmt.Fprintf(output, "  ✗ ollama api CONNECTION FAILED: %s\n", err.Error())
	}

	if err = llmClient.Show(ctx); err == nil {
		_, _ = fmt.Fprintf(output, "  ◆ model %s active\n\n", model)
	} else {
		errors++
		_, _ = fmt.Fprintf(output, "  ✗ model %s not loaded: %s\n\n", model, err.Error())
	}

	if errors < 1 {
		fmt.Fprintln(output, ">> ghost online :: all systems nominal")
	} else {
		fmt.Fprintf(output, ">> ghost offline :: %d critical errors detected", errors)
	}

	return nil
}
