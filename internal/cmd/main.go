package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

// maxPipedInputSize sets the maximum size for piped input to 10 megabytes.
const maxPipedInputSize = 10 << 20

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
				Name:     "vision-model",
				Usage:    "LLM to use for image requests",
				Value:    "qwen2.5vl:7b",
				Sources:  cli.NewValueSourceChain(toml.TOML("vision.model", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "system",
				Usage:    "the system prompt to override the model's",
				Value:    "",
				Sources:  cli.NewValueSourceChain(toml.TOML("system", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "vision-prompt",
				Usage:    "the system prompt to override the vision model's",
				Value:    "",
				Sources:  cli.NewValueSourceChain(toml.TOML("vision.system_prompt", configFile)),
				OnlyOnce: true,
			},
			&cli.StringSliceFlag{
				Name:  "image",
				Usage: "path to an image (can be used multiple times)",
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

	config := llm.Config{
		Host:         cmd.String("host"),
		DefaultModel: cmd.String("model"),
		VisionModel:  cmd.String("vision-model"),
	}

	llmClient, err := llm.NewOllama(config, logger)
	if err != nil {
		return ctx, err
	}

	cmd.Metadata["llmClient"] = llmClient

	return ctx, nil
}

// ghost is the action handler for the root command.
// It checks for piped input, processes user prompt, sends the prompt to the LLM
// client, and returns the response
var ghost = func(ctx context.Context, cmd *cli.Command) error {
	prompt, err := getPrompt(cmd)
	if err != nil {
		return err
	}

	pipedInput, err := getPipedInput()
	if err != nil {
		return err
	}

	if pipedInput != "" {
		prompt = fmt.Sprintf("%s\n\n%s", prompt, pipedInput)
	}

	// TODO: Check for images
	// TODO: Send generate request for images
	// TODO: Send images response with user prompt to Generate
	imagePaths := cmd.StringSlice("image")
	encodedImages, err := encodeImages(imagePaths)
	if err != nil {
		return err
	}

	llmClient := cmd.Metadata["llmClient"].(llm.LLMClient)

	response, err := llmClient.Generate(ctx, cmd.String("system"), prompt, encodedImages)
	if err != nil {
		return err
	}

	output := cmd.Metadata["output"].(io.Writer)
	fmt.Fprintln(output, response)

	return nil
}

// getPrompt checks the prompt argument and returns the prompt with trimmed spaces.
// Returns ErrNoPrompt if prompt is empty.
func getPrompt(cmd *cli.Command) (string, error) {
	prompt := strings.TrimSpace(cmd.StringArg("prompt"))

	if prompt == "" {
		return "", fmt.Errorf("%w", ErrNoPrompt)
	}

	return prompt, nil
}

// getPipedInput checks for and returns any input piped to the command.
// Returns an empty string if piped input doesn't exist or is empty.
// Returns ErrInput if piped input cannot be read.
func getPipedInput() (string, error) {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return "", nil
	}

	if fileInfo.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	pipedInput, err := io.ReadAll(io.LimitReader(os.Stdin, maxPipedInputSize))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInput, err)
	}

	input := strings.TrimSpace(string(pipedInput))

	return input, nil
}

// encodeImages takes a slice of paths and returns a slice of base64 encoded strings.
func encodeImages(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	encoded := make([]string, 0, len(paths))

	for _, path := range paths {
		imageBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read image %s: %w", ErrInput, path, err)
		}

		encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
		encoded = append(encoded, encodedImage)
	}

	return encoded, nil
}
