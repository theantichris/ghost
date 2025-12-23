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
func health(ctx context.Context, cmd *cli.Command) error {
	output := cmd.Root().Metadata["output"].(io.Writer)

	host := cmd.String("host")
	chatModel := cmd.String("model")
	visionModel := cmd.String("vision-model")
	systemPrompt := cmd.String("system")
	visionSystemPrompt := cmd.String("vision-system")
	configFile := cmd.Root().Metadata["configFile"].(altsrc.StringSourcer)

	errorCount := 0

	fmt.Fprint(output, ">> initializing ghost diagnostics...\n\n")

	fmt.Fprintln(output, "SYSTEM CONFIG")

	if _, err := os.Stat(configFile.SourceURI()); err == nil {
		fmt.Fprintf(output, "  ◆ config loaded: %s\n", configFile.SourceURI())
	} else {
		fmt.Fprint(output, "  ◆ config file not loaded: using defaults\n")
	}

	fmt.Fprintf(output, "  ◆ host: %s\n", host)
	fmt.Fprintf(output, "  ◆ chat model: %s\n", chatModel)
	fmt.Fprintf(output, "  ◆ vision model: %s\n", visionModel)

	if systemPrompt == "" {
		fmt.Fprint(output, "  ◆ system prompt: empty\n")
	} else {
		fmt.Fprintf(output, "  ◆ system prompt: %s\n", systemPrompt)
	}

	if visionSystemPrompt == "" {
		fmt.Fprint(output, "  ◆ vision system prompt: empty\n\n")
	} else {
		fmt.Fprintf(output, "  ◆ vision system prompt: %s\n\n", visionSystemPrompt)
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

	if err = llmClient.Show(ctx, chatModel); err == nil {
		fmt.Fprintf(output, "  ◆ chat model %s ACTIVE\n", chatModel)
	} else {
		errorCount++

		if errors.Is(err, llm.ErrModelNotFound) {
			fmt.Fprintf(output, "  ✗ chat model %s NOT LOADED: pull model\n", chatModel)
		} else {
			fmt.Fprintf(output, "  ✗ chat model %s NOT LOADED: %s\n", chatModel, err)
		}
	}

	if err = llmClient.Show(ctx, visionModel); err == nil {
		fmt.Fprintf(output, "  ◆ vision model %s ACTIVE\n\n", visionModel)
	} else {
		errorCount++

		if errors.Is(err, llm.ErrModelNotFound) {
			fmt.Fprintf(output, "  ✗ vision model %s NOT LOADED: pull model\n\n", visionModel)
		} else {
			fmt.Fprintf(output, "  ✗ vision model %s NOT LOADED: %s\n\n", visionModel, err)
		}
	}

	if errorCount == 0 {
		fmt.Fprintln(output, ">> ghost online :: all systems nominal")
	} else {
		fmt.Fprintf(output, ">> ghost offline :: %d critical errors detected\n", errorCount)
	}

	return nil
}
