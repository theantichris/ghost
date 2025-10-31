package cmd

import (
	"context"
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

			_, err := before(context.Background(), &cmd)
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
