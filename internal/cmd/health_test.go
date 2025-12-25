package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/theantichris/ghost/internal/llm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name         string
		llmClient    llm.Client
		configFile   string
		host         string
		model        string
		systemPrompt string
		isError      bool
	}{
		{
			name:       "prints output for no config file",
			llmClient:  llm.MockClient{},
			configFile: "",
		},
		{
			name:       "prints output for loading config file",
			llmClient:  llm.MockClient{},
			configFile: "/home/.config/ghost/config.toml",
		},
		{
			name: "prints output for Ollama API version",
			llmClient: llm.MockClient{
				VersionFunc: func(ctx context.Context) (string, error) {
					return "0.12.6", nil
				},
			},
		},
		{
			name: "prints output for Ollama API error",
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
		},
		{
			name: "prints output for model check",
			llmClient: llm.MockClient{
				ShowFunc: func(ctx context.Context, model string) error {
					return nil
				},
			},
		},
		{
			name: "prints output model check error",
			llmClient: llm.MockClient{
				Error: llm.ErrOllama,
			},
		},
		{
			name: "prints output model not found check error",
			llmClient: llm.MockClient{
				Error: llm.ErrModelNotFound,
			},
		},
		{
			name:         "prints output with system prompt configured",
			llmClient:    llm.MockClient{},
			configFile:   "",
			systemPrompt: "You are Ghost, a cyberpunk AI assistant",
		},
		{
			name:         "prints output with empty system prompt",
			llmClient:    llm.MockClient{},
			configFile:   "",
			systemPrompt: "",
		},
		{
			name: "prints output for version error with model success",
			llmClient: llm.MockClient{
				VersionFunc: func(ctx context.Context) (string, error) {
					return "", llm.ErrOllama
				},
				ShowFunc: func(ctx context.Context, model string) error {
					return nil
				},
			},
		},
		{
			name: "prints output for version success with model error",
			llmClient: llm.MockClient{
				VersionFunc: func(ctx context.Context) (string, error) {
					return "0.12.6", nil
				},
				ShowFunc: func(ctx context.Context, model string) error {
					return llm.ErrOllama
				},
			},
		},
		{
			name:       "prints output with remote host",
			llmClient:  llm.MockClient{},
			host:       "http://192.168.1.100:11434",
			configFile: "",
		},
		{
			name:       "prints output with different model",
			llmClient:  llm.MockClient{},
			model:      "codellama:13b",
			configFile: "",
		},
		{
			name:         "prints output with all custom values",
			llmClient:    llm.MockClient{},
			configFile:   "/custom/path/config.toml",
			host:         "http://remote:11434",
			model:        "llama3.1:70b",
			systemPrompt: "Custom system prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			host := tt.host
			if host == "" {
				host = "http://localhost:11434"
			}

			model := tt.model
			if model == "" {
				model = "test:model"
			}

			cmd := &cli.Command{
				Metadata: map[string]any{
					"output":     output,
					"configFile": altsrc.StringSourcer(tt.configFile),
					"llmClient":  tt.llmClient,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Value: host,
					},
					&cli.StringFlag{
						Name:  "model",
						Value: model,
					},
					&cli.StringFlag{
						Name:  "system",
						Value: tt.systemPrompt,
					},
				},
			}

			err := health(context.Background(), cmd)
			if !tt.isError && err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			g := goldie.New(t, goldie.WithDiffEngine(goldie.ColoredDiff))
			g.Assert(t, t.Name(), output.Bytes())
		})
	}
}
