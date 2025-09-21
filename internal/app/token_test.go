package app

import (
	"bytes"
	"testing"
)

func TestTokenHandler(t *testing.T) {
	t.Run("passes through tokens without think blocks", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		var tokens string

		handler := &tokenHandler{
			output: &output,
			tokens: &tokens,
		}

		handler.handle("Hello ")
		handler.handle("world!")

		expected := "Hello world!"

		if output.String() != expected {
			t.Errorf("expected output %q, got %q", expected, output.String())
		}

		if tokens != expected {
			t.Errorf("expected tokens %q, got %q", expected, tokens)
		}
	})

	t.Run("filters complete think block in single token", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		var tokens string

		handler := &tokenHandler{
			output: &output,
			tokens: &tokens,
		}

		handler.handle("<think>internal thoughts</think>Hello world!")

		expectedOutput := "Hello world!"

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}

		expectedTokens := "<think>internal thoughts</think>Hello world!"
		if tokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, tokens)
		}
	})

	t.Run("filters think block split across tokens", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		var tokens string

		handler := &tokenHandler{
			output: &output,
			tokens: &tokens,
		}

		handler.handle("<th")
		handler.handle("ink>reasoning here</thi")
		handler.handle("nk>Hello world!")

		expectedOutput := "Hello world!"

		if output.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output.String())
		}

		expectedTokens := "<think>reasoning here</think>Hello world!"

		if tokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, tokens)
		}
	})
}
