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
				Usage:    "LLM to use for basic chat",
				Value:    "llama3.1:8b",
				Sources:  cli.NewValueSourceChain(toml.TOML("model", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "vision-model",
				Usage:    "LLM to use for analyzing images",
				Value:    "qwen2.5vl:7b",
				Sources:  cli.NewValueSourceChain(toml.TOML("vision.model", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "system",
				Usage:    "the system prompt to override the basic chat model",
				Value:    "",
				Sources:  cli.NewValueSourceChain(toml.TOML("system", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "vision-system",
				Usage:    "the system prompt to override the vision model",
				Value:    "",
				Sources:  cli.NewValueSourceChain(toml.TOML("vision.system_prompt", configFile)),
				OnlyOnce: true,
			},
			&cli.StringFlag{
				Name:     "vision-prompt",
				Usage:    "the prompt to send for image analysis",
				Value:    "Analyze the attached image(s)",
				Sources:  cli.NewValueSourceChain(toml.TOML("vision.prompt", configFile)),
				OnlyOnce: true,
			},
			&cli.StringSliceFlag{
				Name:  "image",
				Usage: "path to an image (can be used multiple times)",
			},
		},
		Before: beforeHook,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			generateConfig := config{
				host:               cmd.String("host"),
				model:              cmd.String("model"),
				visionModel:        cmd.String("vision-model"),
				systemPrompt:       cmd.String("system"),
				visionSystemPrompt: cmd.String("vision-system"),
				visionPrompt:       cmd.String("vision-prompt"),
			}

			prompt := strings.TrimSpace(cmd.StringArg("prompt"))
			if prompt == "" {
				return fmt.Errorf("%w", ErrNoPrompt)
			}

			pipedInput, err := getPipedInput()
			if err != nil {
				return err
			}

			// Add piped input to the prompt.
			if pipedInput != "" {
				prompt = fmt.Sprintf("%s\n\n%s", prompt, pipedInput)
			}

			var encodedImages []string
			images := cmd.StringSlice("image")
			if len(images) > 0 {
				encodedImages, err = encodeImages(images)

				if err != nil {
					return err
				}
			}

			llmClient := cmd.Metadata["llmClient"].(llm.LLMClient)

			// Stream callback that writes chunks directly to output
			streamCallback := func(chunk string) error {
				_, err := fmt.Fprint(output, chunk)

				return err
			}

			err = generate(ctx, prompt, encodedImages, generateConfig, llmClient, streamCallback)
			if err != nil {
				return err
			}

			fmt.Fprintln(output)

			return nil
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

// before initializes the LLM client and adds it to the root command's metadata.
func beforeHook(ctx context.Context, cmd *cli.Command) (context.Context, error) {
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

// generate sends the prompt to the LLM client's generate function.
// If there is piped input it appends it to the prompt.
// If there are images it sends those to the LLM to be analyzed and appends the
// results to the prompt.
// The callback is called for each chunk of streamed text.
func generate(ctx context.Context, prompt string, images []string, config config, llmClient llm.LLMClient, callback func(string) error) error {
	// If images send a request to analyze them and add the response to the prompt
	// Don't stream to user
	if len(images) > 0 {
		var visionResponse strings.Builder

		err := llmClient.Generate(ctx, config.visionSystemPrompt, config.visionPrompt, images, func(chunk string) error {
			visionResponse.WriteString(chunk)

			return nil
		})

		if err != nil {
			return err
		}

		prompt = fmt.Sprintf("%s\n\n%s", prompt, visionResponse.String())
	}

	// Send the main request and stream response to user
	var fullResponse strings.Builder

	err := llmClient.Generate(ctx, config.systemPrompt, prompt, nil, func(chunk string) error {
		fullResponse.WriteString(chunk)

		return callback(chunk)
	})

	if err != nil {
		return err
	}

	return nil
}

// analyzeImages sends images to the vision model for analysis.
// The analysis is accumulated silently (not streamed to user).
// Returns the complete vision analysis text.
func analyzeImages(ctx context.Context, llmClient llm.LLMClient, config config, images []string) (string, error) {
	var visionResponse strings.Builder

	err := llmClient.Generate(ctx, config.visionSystemPrompt, config.visionPrompt, images, func(chunk string) error {
		visionResponse.WriteString(chunk)

		return nil
	})

	if err != nil {
		return "", err
	}

	return visionResponse.String(), nil
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
		return []string{}, nil
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
