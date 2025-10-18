package llm

import (
	"context"
	"errors"
	"testing"
)

func TestMockClientGenerate(t *testing.T) {
	t.Run("mocks the Generate function", func(t *testing.T) {
		t.Parallel()

		mockGenerate := func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
			return "Hello, chummer!", nil
		}

		llmClient := MockLLMClient{
			GenerateFunc: mockGenerate,
		}

		response, _ := llmClient.Generate(context.Background(), "system prompt", "user prompt")

		if response != "Hello, chummer!" {
			t.Errorf("expected response %q, got %q", "Hello, chummer!", response)
		}
	})

	t.Run("mocks an error return", func(t *testing.T) {
		t.Parallel()

		llmError := errors.New("llm client error")

		llmClient := MockLLMClient{
			Error: llmError,
		}

		_, err := llmClient.Generate(context.Background(), "system prompt", "user prompt")

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, llmError) {
			t.Errorf("expected error %v, got %v", llmError, err)
		}
	})
}
