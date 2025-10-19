package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	// model is the Ollama model name.
	model = "dolphin-mixtral:8x7b"

	// systemPrompt defines the default system level instruction for Ghost's LLM interactions
	systemPrompt = "You are Ghost, a cyberpunk inspired terminal based assistant. Answer requests directly and briefly."
)

// Run executes the root command (ghost) printing out a test string.
func Run(ctx context.Context, args []string, output io.Writer, logger *log.Logger) error {
	var userPrompt string

	configFile, err := loadConfigFile(logger)
	if err != nil {
		return err
	}

	// TODO: add model flag
	// TODO: rename baseURL
	// TODO: evaluate log levels

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
				Value:    "http://localhost:11434",
				Sources:  cli.NewValueSourceChain(toml.TOML("host", configFile)),
				OnlyOnce: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if userPrompt == "" {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			llmClient, err := llm.NewOllama(cmd.String("host"), model, logger)
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

// loadConfigFile attempts to load config.toml from ~/.config/ghost.
func loadConfigFile(logger *log.Logger) (altsrc.StringPtrSourcer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return altsrc.StringPtrSourcer{}, fmt.Errorf("%w", ErrConfigFile)
	}

	configFile := filepath.Join(homeDir, ".config/ghost", "config.toml")

	var sourcer altsrc.StringPtrSourcer
	if _, err := os.Stat(configFile); err != nil {
		logger.Debug("config file not found", "file", configFile)
	} else {
		sourcer = altsrc.NewStringPtrSourcer(&configFile)
		logger.Debug("loading config file", "file", configFile)
	}

	return sourcer, nil
}
