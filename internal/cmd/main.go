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

// Run executes the root command (ghost) printing out a test string.
func Run(ctx context.Context, args []string, version string, output io.Writer, logger *log.Logger) error {
	var userPrompt string

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
				Name:        "prompt",
				Destination: &userPrompt,
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if userPrompt == "" {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			llmClient, err := llm.NewOllama(cmd.String("host"), cmd.String("model"), logger)
			if err != nil {
				return err
			}

			// TODO: Should return the output and print it here.
			// TODO: Move generate to its own file.
			return generate(ctx, cmd.String("system"), userPrompt, llmClient, output)
		},
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

// generate sends the prompt to the LLM API, processes the response, and displays the results.
func generate(ctx context.Context, systemPrompt, userPrompt string, llmClient llm.LLMClient, output io.Writer) error {
	response, err := llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(output, response); err != nil {
		return fmt.Errorf("%w: %w", ErrOutput, err)
	}

	return nil
}
