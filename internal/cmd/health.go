package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/theantichris/ghost/internal/llm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

// health is the action handler for the health subcommand that displays system diagnostics.
// It prints the current configuration, checks Ollama API connectivity, verifies API version, and validates that the configured model is available.
var health = func(ctx context.Context, cmd *cli.Command) error {
	output := cmd.Root().Metadata["output"].(io.Writer)

	host := cmd.String("host")
	model := cmd.String("model")
	systemPrompt := cmd.String("system")
	configFile := cmd.Root().Metadata["configFile"].(altsrc.StringSourcer)

	errors := 0

	fmt.Fprint(output, ">> initializing ghost diagnostics...\n\n")

	fmt.Fprintln(output, "SYSTEM CONFIG")
	fmt.Fprintf(output, "  ◆ host: %s\n", host)
	fmt.Fprintf(output, "  ◆ model: %s\n", model)
	fmt.Fprintf(output, "  ◆ config: %s\n", configFile.SourceURI())
	fmt.Fprintf(output, "  ◆ system prompt: %s\n\n", systemPrompt)

	fmt.Fprintln(output, "NEURAL LINK STATUS")

	llmClient := cmd.Root().Metadata["llmClient"].(llm.LLMClient)
	version, err := llmClient.Version(ctx)
	if err == nil {
		fmt.Fprintf(output, "  ◆ ollama api CONNECTED [v%s]\n", version)
	} else {
		errors++
		fmt.Fprintf(output, "  ✗ ollama api CONNECTION FAILED: %s\n", err.Error())
	}

	if err = llmClient.Show(ctx, model); err == nil {
		fmt.Fprintf(output, "  ◆ model %s active\n\n", model)
	} else {
		errors++
		fmt.Fprintf(output, "  ✗ model %s not loaded: %s\n\n", model, err.Error())
	}

	if errors == 0 {
		fmt.Fprintln(output, ">> ghost online :: all systems nominal")
	} else {
		fmt.Fprintf(output, ">> ghost offline :: %d critical errors detected\n", errors)
	}

	return nil
}
