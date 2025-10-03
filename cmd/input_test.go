package cmd

import (
	"errors"
	"io"
	"strings"
	"testing"
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

func TestReadPipedInput(t *testing.T) {
	t.Run("reads piped input", func(t *testing.T) {
		t.Parallel()

		input := strings.NewReader("cat main.go")

		output, err := readPipedInput(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedOutput := "cat main.go"

		if output != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, output)
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
