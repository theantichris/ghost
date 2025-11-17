package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

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

	errorCount := 0

	fmt.Fprint(output, ">> initializing ghost diagnostics...\n\n")

	fmt.Fprintln(output, "SYSTEM CONFIG")
	fmt.Fprintf(output, "  ◆ host: %s\n", host)
	fmt.Fprintf(output, "  ◆ model: %s\n", model)

	if _, err := os.Stat(configFile.SourceURI()); err == nil {
		fmt.Fprintf(output, "  ◆ config loaded: %s\n", configFile.SourceURI())
	} else {
		fmt.Fprintf(output, "  ◆ config file not loaded: %s\n", err)
	}

	if systemPrompt == "" {
		fmt.Fprint(output, "  ◆ system prompt: empty\n\n")
	} else {
		fmt.Fprintf(output, "  ◆ system prompt: %s\n\n", systemPrompt)
	}

	fmt.Fprintln(output, "NEURAL LINK STATUS")

	llmClient := cmd.Root().Metadata["llmClient"].(llm.LLMClient)
	version, err := llmClient.Version(ctx)
	if err == nil {
		fmt.Fprintf(output, "  ◆ ollama api CONNECTED [v%s]\n", version)
	} else {
		errorCount++
		fmt.Fprintf(output, "  ✗ ollama api CONNECTION FAILED: %s\n", err.Error())
	}

	if err = llmClient.Show(ctx, model); err == nil {
		fmt.Fprintf(output, "  ◆ model %s active\n\n", model)
	} else {
		errorCount++

		if errors.Is(err, llm.ErrModelNotFound) {
			fmt.Fprintf(output, "  ✗ model %s not loaded: pull model\n\n", model)
		} else {
			fmt.Fprintf(output, "  ✗ model %s not loaded: %s\n\n", model, err)
		}
	}

	if errorCount == 0 {
		fmt.Fprintln(output, ">> ghost online :: all systems nominal")
	} else {
		fmt.Fprintf(output, ">> ghost offline :: %d critical errors detected\n", errorCount)
	}

	return nil
}
