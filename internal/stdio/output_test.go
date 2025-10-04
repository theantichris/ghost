package stdio

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestOutputWriterWrite(t *testing.T) {
	t.Run("writes token without think block", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("Hello World")

		expectedOutput := "Hello World"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		expectedTokens := "Hello World"

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})
}

func TestOutputWriterReset(t *testing.T) {
	t.Run("resets all state fields", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("<think>reasoning</think>Response")
		actualTokens = "accumulated tokens"
		writer.buffer.WriteString("buffer content")
		writer.insideThinkBlock = true
		writer.canPassThrough = true
		writer.newlinesTrimmed = true

		writer.Reset()

		if writer.buffer.Len() != 0 {
			t.Errorf("expected buffer length 0, got %d", writer.buffer.Len())
		}

		if writer.insideThinkBlock {
			t.Error("expected insideThinkBlock to be false")
		}

		if writer.canPassThrough {
			t.Error("expected canPassThrough to be false")
		}

		if writer.newlinesTrimmed {
			t.Error("expected newlinesTrimmed to be false")
		}

		expectedTokens := ""

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})

	t.Run("allows writer to be reused after reset", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("<think>first</think>First response")

		expectedFirstOutput := "First response"

		if actualOutput.String() != expectedFirstOutput {
			t.Errorf("expected first output %q, got %q", expectedFirstOutput, actualOutput.String())
		}

		writer.Reset()
		actualOutput.Reset()

		writer.Write("<think>second</think>Second response")

		expectedSecondOutput := "Second response"

		if actualOutput.String() != expectedSecondOutput {
			t.Errorf("expected second output %q, got %q", expectedSecondOutput, actualOutput.String())
		}
	})

	t.Run("clears tokens pointer", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		actualTokens := "initial tokens"
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Reset()

		expectedTokens := ""

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})
}

func TestOutputWriterFlush(t *testing.T) {
	t.Run("discards buffered content when think block never closed", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("<think>")
		writer.Write("incomplete reasoning")

		writer.Flush()

		expectedOutput := ""

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}

		expectedTokens := "<think>incomplete reasoning"

		if actualTokens != expectedTokens {
			t.Errorf("expected tokens %q, got %q", expectedTokens, actualTokens)
		}
	})

	t.Run("flushes partial tag when stream ends early", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("<th")

		writer.Flush()

		expectedOutput := "<th"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("does nothing when in pass-through mode", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Write("Normal text")

		writer.Flush()

		expectedOutput := "Normal text"

		if actualOutput.String() != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput.String())
		}
	})

	t.Run("does nothing when buffer is empty", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		var actualTokens string
		logger := log.New(io.Discard)

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
		}

		writer.Flush()

		expectedOutput := ""

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

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
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
			writer.Write(token)
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

		writer := &OutputWriter{
			Logger: logger,
			Output: &actualOutput,
			Tokens: &actualTokens,
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
			writer.Write(token)
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

		writer1 := &OutputWriter{
			Logger: logger,
			Output: &actualOutput1,
			Tokens: &actualTokens1,
		}

		writer1.Write("<think>first</think>Response")

		var actualOutput2 bytes.Buffer
		var actualTokens2 string
		writer2 := &OutputWriter{
			Logger: logger,
			Output: &actualOutput2,
			Tokens: &actualTokens2,
		}

		writer2.Write("<think>second</think>Another")

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
