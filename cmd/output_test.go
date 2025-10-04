package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

type errorWriter struct {
	failAt int
	calls  int
}

func (err *errorWriter) Write(p []byte) (int, error) {
	err.calls++

	if err.calls == err.failAt {
		return 0, errors.New("simulated I/O error")
	}

	return len(p), nil
}

func TestOutputWriterWrite(t *testing.T) {
	t.Run("writes token without think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("Hello World")

		expectedOutput := "Hello World"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		expectedTokens := "Hello World"

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})

	t.Run("strips think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<think>")
		writer.write("reasoning here")
		writer.write("</think>")
		writer.write("Actual response")

		expectedOutput := "Actual response"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		expectedTokens := "<think>reasoning here</think>Actual response"

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})

	t.Run("strips think block with close tag in one token", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<think>reasoning</think>Response")

		expectedOutput := "Response"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("trims leading whitespace before first output", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("  \n\t  Hello World")

		expectedOutput := "Hello World"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("trims leading whitespace after think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<think>reasoning</think>  \n\t  Response")

		expectedOutput := "Response"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("enables pass-through after detecting no think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("Normal ")
		writer.write("response ")
		writer.write("here")

		expectedOutput := "Normal response here"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		if !writer.canPassThrough {
			t.Error("expected canPassThrough to be true")
		}
	})

	t.Run("enables pass-through after closing think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<think>reasoning</think>First ")
		writer.write("Second")

		expectedOutput := "First Second"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		if !writer.canPassThrough {
			t.Error("expected canPassThrough to be true after closing think block")
		}
	})

	t.Run("handles partial open tag", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<th")
		writer.write("ink>reasoning</think>Response")

		expectedOutput := "Response"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("logs error on write failure during normal output", func(t *testing.T) {
		t.Parallel()

		var actualTokens string
		logger := log.New(io.Discard)
		errWriter := &errorWriter{failAt: 1}

		writer := &outputWriter{
			logger: logger,
			output: errWriter,
			tokens: &actualTokens,
		}

		writer.write("Hello")

		expectedCalls := 1

		if errWriter.calls != expectedCalls {
			t.Errorf("expected %d write call, got %d", expectedCalls, errWriter.calls)
		}
	})

	t.Run("logs error on write failure during pass-through", func(t *testing.T) {
		t.Parallel()

		var actualTokens string
		logger := log.New(io.Discard)
		errWriter := &errorWriter{failAt: 1}

		writer := &outputWriter{
			logger:         logger,
			output:         errWriter,
			tokens:         &actualTokens,
			canPassThrough: true,
		}

		writer.write("  \nHello")

		expectedCalls := 1

		if errWriter.calls != expectedCalls {
			t.Errorf("expected %d write call, got %d", expectedCalls, errWriter.calls)
		}
	})

	t.Run("handles empty tokens", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("")

		expectedOutput := ""

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("accumulates tokens even when output is stripped", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		writer.write("<think>")
		writer.write("internal reasoning")
		writer.write("</think>")

		expectedTokens := "<think>internal reasoning</think>"

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}

		expectedOutput := ""

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("handles whitespace-only token during pass-through", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger:          logger,
			output:          &actualOutput,
			tokens:          &actualTokens,
			canPassThrough:  true,
			newlinesTrimmed: true,
		}

		writer.write("   ")

		expectedOutput := "   "

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})
}

func TestIsOpenTag(t *testing.T) {
	t.Run("returns true for open tag", func(t *testing.T) {
		t.Parallel()

		actualResult := isOpenTag("<think>")
		expectedResult := true

		if actualResult != expectedResult {
			t.Errorf("expected %v for '<think>', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns true for open tag with content", func(t *testing.T) {
		t.Parallel()

		actualResult := isOpenTag("<think>reasoning")
		expectedResult := true

		if actualResult != expectedResult {
			t.Errorf("expected %v for '<think>reasoning', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns false for partial open tag", func(t *testing.T) {
		t.Parallel()

		actualResult := isOpenTag("<thin")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for '<thin', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns false for non-tag content", func(t *testing.T) {
		t.Parallel()

		actualResult := isOpenTag("Hello")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for 'Hello', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		t.Parallel()

		actualResult := isOpenTag("")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for empty string, got %v", expectedResult, actualResult)
		}
	})
}

func TestIsCloseTag(t *testing.T) {
	t.Run("returns true and index for close tag", func(t *testing.T) {
		t.Parallel()

		actualFound, actualIndex := isCloseTag("</think>")
		expectedFound := true
		expectedIndex := 0

		if actualFound != expectedFound {
			t.Errorf("expected found %v for '</think>', got %v", expectedFound, actualFound)
		}

		if actualIndex != expectedIndex {
			t.Errorf("expected index %d, got %d", expectedIndex, actualIndex)
		}
	})

	t.Run("returns true and index for close tag with content", func(t *testing.T) {
		t.Parallel()

		actualFound, actualIndex := isCloseTag("reasoning</think>")
		expectedFound := true
		expectedIndex := 9

		if actualFound != expectedFound {
			t.Errorf("expected found %v for 'reasoning</think>', got %v", expectedFound, actualFound)
		}

		if actualIndex != expectedIndex {
			t.Errorf("expected index %d, got %d", expectedIndex, actualIndex)
		}
	})

	t.Run("returns false for partial close tag", func(t *testing.T) {
		t.Parallel()

		actualFound, _ := isCloseTag("</thin")
		expectedFound := false

		if actualFound != expectedFound {
			t.Errorf("expected found %v for '</thin', got %v", expectedFound, actualFound)
		}
	})

	t.Run("returns false for non-tag content", func(t *testing.T) {
		t.Parallel()

		actualFound, _ := isCloseTag("Hello")
		expectedFound := false

		if actualFound != expectedFound {
			t.Errorf("expected found %v for 'Hello', got %v", expectedFound, actualFound)
		}
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		t.Parallel()

		actualFound, _ := isCloseTag("")
		expectedFound := false

		if actualFound != expectedFound {
			t.Errorf("expected found %v for empty string, got %v", expectedFound, actualFound)
		}
	})
}

func TestThinkBlockNotPossible(t *testing.T) {
	t.Run("returns false for empty string", func(t *testing.T) {
		t.Parallel()

		actualResult := thinkBlockNotPossible("")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for empty string, got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns false for valid prefix of open tag", func(t *testing.T) {
		t.Parallel()

		tests := []string{"<", "<t", "<th", "<thin", "<think"}
		expectedResult := false

		for _, test := range tests {
			actualResult := thinkBlockNotPossible(test)

			if actualResult != expectedResult {
				t.Errorf("expected %v for %q, got %v", expectedResult, test, actualResult)
			}
		}
	})

	t.Run("returns false for complete open tag", func(t *testing.T) {
		t.Parallel()

		actualResult := thinkBlockNotPossible("<think>")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for '<think>', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns true when content does not start with open tag", func(t *testing.T) {
		t.Parallel()

		tests := []string{"Hello", "response", "<other>", " <think>"}
		expectedResult := true

		for _, test := range tests {
			actualResult := thinkBlockNotPossible(test)

			if actualResult != expectedResult {
				t.Errorf("expected %v for %q, got %v", expectedResult, test, actualResult)
			}
		}
	})

	t.Run("returns true for 7+ characters not starting with open tag", func(t *testing.T) {
		t.Parallel()

		actualResult := thinkBlockNotPossible("Normal response")
		expectedResult := true

		if actualResult != expectedResult {
			t.Errorf("expected %v for 'Normal response', got %v", expectedResult, actualResult)
		}
	})

	t.Run("returns false for open tag with content", func(t *testing.T) {
		t.Parallel()

		actualResult := thinkBlockNotPossible("<think>reasoning")
		expectedResult := false

		if actualResult != expectedResult {
			t.Errorf("expected %v for '<think>reasoning', got %v", expectedResult, actualResult)
		}
	})
}

func TestOutputWriterIntegration(t *testing.T) {
	t.Run("handles think block detection with character-by-character streaming", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		streamTokens := []string{
			"<",
			"think",
			">",
			"reasoning",
			"</",
			"think",
			">",
			"Response",
		}

		for _, token := range streamTokens {
			writer.write(token)
		}

		expectedOutput := "Response"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		expectedTokens := strings.Join(streamTokens, "")

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})

	t.Run("handles response with no think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &outputWriter{
			logger: logger,
			output: &actualOutput,
			tokens: &actualTokens,
		}

		streamTokens := []string{
			"  \n\t",
			"Hello, ",
			"how ",
			"can ",
			"I ",
			"help?",
		}

		for _, token := range streamTokens {
			writer.write(token)
		}

		expectedOutput := "Hello, how can I help?"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("handles multiple think blocks", func(t *testing.T) {
		t.Parallel()

		var actualOutput1 bytes.Buffer
		var actualTokens1 string
		logger := log.New(io.Discard)

		writer1 := &outputWriter{
			logger: logger,
			output: &actualOutput1,
			tokens: &actualTokens1,
		}

		writer1.write("<think>first</think>Response")

		var actualOutput2 bytes.Buffer
		var actualTokens2 string
		writer2 := &outputWriter{
			logger: logger,
			output: &actualOutput2,
			tokens: &actualTokens2,
		}

		writer2.write("<think>second</think>Another")

		expectedOutput1 := "Response"

		if actualOutput1.String() != expectedOutput1 {
			t.Errorf("expected first output %q, got %q", expectedOutput1, actualOutput1.String())
		}

		expectedOutput2 := "Another"

		if actualOutput2.String() != expectedOutput2 {
			t.Errorf("expected second output %q, got %q", expectedOutput2, actualOutput2.String())
		}
	})
}
