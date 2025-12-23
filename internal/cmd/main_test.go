package cmd

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"

	"github.com/urfave/cli/v3"
)

func TestBefore(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "initializes LLM Client and adds to metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			cmd := cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Value: "http://test.dev",
					},
					&cli.StringFlag{
						Name:  "model",
						Value: "default:model",
					},
					&cli.StringFlag{
						Name:  "vision-model",
						Value: "vision:model",
					},
				},
				Metadata: map[string]any{
					"logger": logger,
				},
			}

			_, err := beforeHook(context.Background(), &cmd)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			actual := cmd.Metadata["llmClient"]

			if actual == nil {
				t.Fatalf("expected LLMClient, got nil")
			}

			if _, ok := actual.(llm.LLMClient); !ok {
				t.Errorf("expected LLM Client to be of type LLMClient, got %v", actual)
			}
		})
	}

}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		images    []string
		config    config
		llmClient llm.LLMClient
		expected  string
		isError   bool
		Error     error
	}{
		{
			name:   "sends prompt to LLM generate without images",
			prompt: "test prompt",
			images: []string{},
			config: config{
				systemPrompt: "system prompt",
			},
			llmClient: llm.MockLLMClient{
				GenerateFunc: func(ctx context.Context, systemPrompt string, userPrompt string, images []string) (string, error) {
					return "this prompt is good", nil
				},
			},
			expected: "this prompt is good",
		},
		{
			name:   "returns error for LLM generate without images",
			prompt: "test prompt",
			images: []string{},
			config: config{
				systemPrompt: "system prompt",
			},
			llmClient: llm.MockLLMClient{
				Error: llm.ErrOllama,
			},
			isError: true,
			Error:   llm.ErrOllama,
		},
		{
			name:   "sends prompt to LLM generate with images",
			prompt: "test prompt",
			images: []string{"test/image.png"},
			config: config{
				systemPrompt:       "system prompt",
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockLLMClient{
				GenerateFunc: func(ctx context.Context, systemPrompt string, userPrompt string, images []string) (string, error) {
					return "this prompt is good", nil
				},
			},
			expected: "this prompt is good",
		},
		{
			name:   "returns error for LLM generate with images",
			prompt: "test prompt",
			images: []string{"test/image.png"},
			config: config{
				systemPrompt:       "system prompt",
				visionSystemPrompt: "vision system prompt",
				visionPrompt:       "vision prompt",
			},
			llmClient: llm.MockLLMClient{
				Error: llm.ErrOllama,
			},
			isError: true,
			Error:   llm.ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := generate(context.Background(), tt.prompt, tt.images, tt.config, tt.llmClient)

			if !tt.isError {
				if err != nil {
					t.Fatalf("expect no error, got %v", err)
				}

				if actual != tt.expected {
					t.Errorf("expected response %q, got %q", tt.expected, actual)
				}
			}

			if tt.isError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if !errors.Is(err, tt.Error) {
					t.Errorf("expected error %v, got %v", tt.Error, err)
				}
			}
		})
	}
}
