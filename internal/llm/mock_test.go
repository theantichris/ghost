package llm

import (
	"context"
	"errors"
	"testing"
)

func TestMockClientGenerate(t *testing.T) {
	t.Run("mocks the Generate function", func(t *testing.T) {
		t.Parallel()

		mockGenerate := func(ctx context.Context, systemPrompt, userPrompt string, images []string) (string, error) {
			return "Hello, chummer!", nil
		}

		llmClient := MockLLMClient{
			GenerateFunc: mockGenerate,
		}

		response, _ := llmClient.Generate(context.Background(), "system prompt", "user prompt", []string{})

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

		_, err := llmClient.Generate(context.Background(), "system prompt", "user prompt", []string{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, llmError) {
			t.Errorf("expected error %v, got %v", llmError, err)
		}
	})
}

func TestMockClientVersion(t *testing.T) {
	t.Run("mocks the Version function", func(t *testing.T) {
		t.Parallel()

		mockVersion := func(ctx context.Context) (string, error) {
			return "0.12.6", nil
		}

		llmClient := MockLLMClient{
			VersionFunc: mockVersion,
		}

		response, _ := llmClient.Version(context.Background())

		if response != "0.12.6" {
			t.Errorf("expected response %q, got %q", "0.12.6", response)
		}
	})

	t.Run("mocks an error return", func(t *testing.T) {
		t.Parallel()

		llmError := errors.New("llm client error")

		llmClient := MockLLMClient{
			Error: llmError,
		}

		_, err := llmClient.Version(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, llmError) {
			t.Errorf("expected error %v, got %v", llmError, err)
		}
	})
}

func TestMockClientShow(t *testing.T) {
	t.Run("mocks the Show function", func(t *testing.T) {
		t.Parallel()

		mockShow := func(ctx context.Context) error {
			return nil
		}

		llmClient := MockLLMClient{
			ShowFunc: mockShow,
		}

		if err := llmClient.Show(context.Background()); err != nil {
			t.Errorf("expect no error, got %v", err)
		}

	})

	t.Run("mocks an error return", func(t *testing.T) {
		t.Parallel()

		llmError := errors.New("llm client error")

		llmClient := MockLLMClient{
			Error: llmError,
		}

		if err := llmClient.Show(context.Background()); err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
