package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

// errorReader simulates read errors.
type errorReader struct {
	failAt int
	calls  int
}

// Read handles read operations for errorReader.
func (err *errorReader) Read(p []byte) (int, error) {
	err.calls++

	if err.calls == err.failAt {
		return 0, errors.New("simulated I/O error")
	}

	if err.calls == 1 {
		copy(p, []byte("partial data\n"))

		return 13, nil
	}

	return 0, io.EOF
}

func TestRunSingleQuest(t *testing.T) {
	t.Run("queries with newline", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke."
		expectedOutput := "Why did the chicken cross the road?\n"

		mockClient := &llm.MockLLMClient{
			Content: "Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := runSingleQuery(mockClient, false, query, &output)

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

		err := runSingleQuery(mockClient, false, query, &output)
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

		err := runSingleQuery(mockClient, false, query, &output)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}
	})

	t.Run("queries without newline", func(t *testing.T) {
		t.Parallel()

		query := "Tell me a joke?"
		expectedOutput := "Why did the chicken cross the road?"

		defer func() {
			noNewLine = false
		}()

		mockClient := &llm.MockLLMClient{
			Content: "Why did the chicken cross the road?",
		}

		var output bytes.Buffer

		err := runSingleQuery(mockClient, true, query, &output)
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

		err := runSingleQuery(mockClient, false, "Hello", &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, llm.ErrResponseBody) {
			t.Errorf("expected error %v, got %v", llm.ErrResponseBody, err)
		}
	})
}

func TestReadPipedInput(t *testing.T) {
	t.Run("reades piped input", func(t *testing.T) {
		t.Parallel()

		input := strings.NewReader("cat main.go")

		output, err := readPipedInput(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedOutput := "cat main.go"

		if output != expectedOutput {
			t.Errorf("expected  output %q, got %q", expectedOutput, output)
		}
	})

	t.Run("handler reader error", func(t *testing.T) {
		t.Parallel()

		errReader := &errorReader{failAt: 2}

		_, err := readPipedInput(errReader)
		if err == nil {
			t.Error("expected error for I/O failure, got nil")
		}

		if !strings.Contains(err.Error(), "simulated I/O error") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
