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
		name       string
		llmClient  llm.LLMClient
		configFile string
		isError    bool
	}{
		{
			name:       "prints output for no config file",
			llmClient:  llm.MockLLMClient{},
			configFile: "",
		},
		{
			name:       "prints output for loading config file",
			llmClient:  llm.MockLLMClient{},
			configFile: "/home/.config/ghost/config.toml",
		},
		{
			name: "prints output for Ollama API version",
			llmClient: llm.MockLLMClient{
				VersionFunc: func(ctx context.Context) (string, error) {
					return "0.12.6", nil
				},
			},
		},
		{
			name: "prints output for Ollama API error",
			llmClient: llm.MockLLMClient{
				Error: llm.ErrOllama,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			cmd := &cli.Command{
				Metadata: map[string]any{
					"output":     output,
					"configFile": altsrc.NewStringPtrSourcer(&tt.configFile),
					"llmClient":  tt.llmClient,
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
