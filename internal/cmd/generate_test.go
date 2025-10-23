package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		llmClient llm.LLMClient
		prompt    string
		expected  string
		isError   bool
		err       error
	}{
		{
			name: "generates a LLM response",
			llmClient: llm.MockLLMClient{
				GenerateFunc: func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
					return "The terms are often interchangeable.", nil
				},
			},
			prompt:   "What is the difference between a netrunner and a decker?",
			expected: "The terms are often interchangeable.",
		},
		{
			name: "returns LLM error",
			llmClient: llm.MockLLMClient{
				Error: llm.ErrOllama,
			},
			prompt:  "What is the difference between a netrunning and a decker?",
			isError: true,
			err:     llm.ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := generate(context.Background(), "system prompt", tt.prompt, tt.llmClient)

			if !tt.isError && err != nil {
				t.Fatalf("expected no error got, %s", err)
			}

			if tt.isError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
			}

			if !tt.isError && response != tt.expected {
				t.Errorf("expected response to be %q, got %q", tt.expected, response)
			}
		})
	}
}
