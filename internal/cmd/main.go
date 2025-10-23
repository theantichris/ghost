package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

// Run executes the root ghost command with the given context, arguments, version, output writer, and logger.
// It loads the configuration file, initializes the CLI command structure with flags and subcommands,
// and returns any errors that occur during execution.
func Run(ctx context.Context, args []string, version string, output io.Writer, logger *log.Logger) error {
	configFile, err := loadConfigFile(logger)
	if err != nil {
		return err
	}

	cmd := &cli.Command{
		Name:    "ghost",
		Usage:   "send a prompt to ghost",
		Version: version,
		Metadata: map[string]any{
			"output":     output,
			"logger":     logger,
			"configFile": configFile,
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "prompt",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "Ollama API URL",
				Value:    "http://localhost:11434",
				Sources:  cli.NewValueSourceChain(toml.TOML("host", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "model",
				Usage:    "default LLM",
				Value:    "llama3.1:8b",
				Sources:  cli.NewValueSourceChain(toml.TOML("model", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "system",
				Usage:    "the system prompt to override the model's",
				Value:    "",
				Sources:  cli.NewValueSourceChain(toml.TOML("system", configFile)),
				OnlyOnce: true,
			},
		},
		Before: before,
		Action: ghost,
		Commands: []*cli.Command{
			{
				Name:   "health",
				Usage:  "check ghost environment and dependencies",
				Action: health,
			},
		},
	}

	return cmd.Run(ctx, args)
}

// before initializes the LLM client and adds it to the root command's metadata.
var before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	logger := cmd.Metadata["logger"].(*log.Logger)

	llmClient, err := llm.NewOllama(cmd.String("host"), cmd.String("model"), logger)
	if err != nil {
		return ctx, err
	}

	cmd.Metadata["llmClient"] = llmClient

	return ctx, nil
}

// ghost is the action handler for the main ghost command that processes user prompts and generates LLM responses.
var ghost = func(ctx context.Context, cmd *cli.Command) error {
	if cmd.StringArg("prompt") == "" {
		return fmt.Errorf("%w", ErrNoPrompt)
	}

	llmClient := cmd.Metadata["llmClient"].(llm.LLMClient)

	response, err := llmClient.Generate(ctx, cmd.String("system"), cmd.StringArg("prompt"))
	if err != nil {
		return err
	}

	output := cmd.Metadata["output"].(io.Writer)
	_, _ = fmt.Fprintln(output, response)

	return nil
}
