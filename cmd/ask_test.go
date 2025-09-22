package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

func TestRunSingleQuest(t *testing.T) {
	t.Run("queries with newline", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke?"
		expectedOutput := "Why did the chicken cross the road?\n"

		mockClient := &llm.MockLLMClient{
			Content: "Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := runSingleQuery(mockClient, query, &output)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}
	})

	t.Run("queries without newline", func(t *testing.T) {
		t.Parallel()

		noNewLine = true
		query := "Tell me a joke?"
		expectedOutput := "Why did the chicken cross the road?"

		mockClient := &llm.MockLLMClient{
			Content: "Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := runSingleQuery(mockClient, query, &output)
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
			Error: llm.ErrResponseBody,
		}

		err := runSingleQuery(mockClient, "Hello", &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, llm.ErrResponseBody) {
			t.Errorf("expected error %v, got %v", llm.ErrResponseBody, err)
		}
	})
}
