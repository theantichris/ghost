package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/theantichris/ghost/internal/llm"
)

func TestProcessQuery(t *testing.T) {
	t.Run("processes queries with newline", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke."
		expectedOutput := "Why did the chicken cross the road?\n"

		mockClient := &llm.MockLLMClient{
			Content: "Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := processQuery(context.Background(), mockClient, query, &output, log.New(io.Discard))

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}
	})

	t.Run("strips <think> block from response", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke."
		expectedOutput := "Why did the chicken cross the road?\n"

		mockClient := &llm.MockLLMClient{
			Content: "<think>...</think>Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := processQuery(context.Background(), mockClient, query, &output, log.New(io.Discard))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}
	})

	t.Run("does not strip <think> block from response without </think>", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke."
		expectedOutput := "<think>...Why did the chicken cross the road?\n"

		mockClient := &llm.MockLLMClient{
			Content: "<think>...Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := processQuery(context.Background(), mockClient, query, &output, log.New(io.Discard))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}
	})

	t.Run("returns error when LLM response fails", func(t *testing.T) {
		t.Parallel()

		mockClient := &llm.MockLLMClient{
			Error: llm.ErrResponse,
		}

		err := processQuery(context.Background(), mockClient, "Hello", &bytes.Buffer{}, log.New(io.Discard))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, ErrLLM) {
			t.Errorf("expected error %v, got %v", ErrLLM, err)
		}
	})
}
