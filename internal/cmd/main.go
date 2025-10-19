package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/urfave/cli/v3"
)

const (
	// model is the Ollama model name.
	model = "dolphin-mixtral:8x7b"

	// ollamaURL is the base address to the Ollama API
	ollamaURL = "http://100.92.199.66:11434"

	// systemPrompt defines the default system level instruction for Ghost's LLM interactions
	systemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Answer requests directly and briefly."
)

// Run executes the root command (ghost) printing out a test string.
func Run(ctx context.Context, args []string, output io.Writer, logger *log.Logger) error {
	var userPrompt string

	// TODO: add model flag

	cmd := &cli.Command{
		Name:      commands["ghost"].Name,
		Usage:     commands["ghost"].Usage,
		ArgsUsage: "[prompt]",
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
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if userPrompt == "" {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			// TODO: rename ollamaURL/baseURL to host

			llmClient, err := llm.NewOllama(ollamaURL, model, logger)
			if err != nil {
				return err
			}

			return generate(ctx, userPrompt, llmClient, output)
		},
	}

	return cmd.Run(ctx, args)
}

// generate sends the prompt to the LLM API, processes the response, and displays the results.
func generate(ctx context.Context, userPrompt string, llmClient llm.LLMClient, output io.Writer) error {
	response, err := llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(output, response); err != nil {
		return fmt.Errorf("%w: %w", ErrOutput, err)
	}

	return nil
}
