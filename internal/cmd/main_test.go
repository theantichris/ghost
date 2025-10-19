package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/theantichris/ghost/internal/llm"
)

// errorWrite is used to test output errors.
type errorWriter struct {
	err error
}

// Write will return an error if one is set, otherwise the length of str.
func (writer *errorWriter) Write(str []byte) (int, error) {
	if writer.err != nil {
		return 0, writer.err
	}

	return len(str), nil
}

func TestHandleLLMRequest(t *testing.T) {
	tests := []struct {
		name      string
		llmClient llm.LLMClient
		writer    io.Writer
		prompt    string
		isGolden  bool
		isErr     bool
		err       error
	}{
		{
			name:   "generates a LLM response",
			writer: &bytes.Buffer{},
			llmClient: llm.MockLLMClient{
				GenerateFunc: func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
					return "The terms are often interchangeable.", nil
				},
			},
			prompt:   "What is the difference between a netrunner and a decker?",
			isGolden: true,
		},
		{
			name:      "returns error for bad output",
			llmClient: llm.MockLLMClient{},
			writer:    &errorWriter{err: errors.New("error printing output")},
			prompt:    "What is the difference between a netrunner and a decker?",
			isErr:     true,
			err:       ErrOutput,
		},
		{
			name: "returns LLM error",
			llmClient: llm.MockLLMClient{
				Error: llm.ErrOllama,
			},
			writer: &bytes.Buffer{},
			prompt: "What is the difference between a netrunning and a decker?",
			isErr:  true,
			err:    llm.ErrOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generate(context.Background(), "system prompt", tt.prompt, tt.llmClient, tt.writer)

			if !tt.isErr && err != nil {
				t.Fatalf("expected no error got, %s", err)
			}

			if tt.isErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
			}

			if tt.isGolden {
				buffer, ok := tt.writer.(*bytes.Buffer)
				if !ok {
					t.Fatalf("expected writer to be of type %T, got %T", &bytes.Buffer{}, buffer)
				}

				g := goldie.New(t)
				g.Assert(t, t.Name(), buffer.Bytes())
			}
		})
	}
}
