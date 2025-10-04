package cmd

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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

func TestGetUserInput(t *testing.T) {
	t.Run("returns error when no input provided", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return false, nil
			},
		}

		cmd := &cobra.Command{}
		args := []string{}

		_, err := reader.getUserInput(cmd, args)
		if err == nil {
			t.Fatal("expected error when no input provided, got nil")
		}

		if !errors.Is(err, ErrInput) {
			t.Errorf("expected error to wrap ErrInput, got %v", err)
		}
	})

	t.Run("uses command-line arguments as query", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return false, nil
			},
		}

		cmd := &cobra.Command{}
		args := []string{"What", "is", "Go?"}

		actual, err := reader.getUserInput(cmd, args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := "What is Go?"
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})

	t.Run("reads piped input only", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return true, nil
			},
		}

		cmd := &cobra.Command{}
		cmd.SetIn(strings.NewReader("func main() {}\n"))
		args := []string{}

		actual, err := reader.getUserInput(cmd, args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := "func main() {}\n"
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})

	t.Run("combines piped input with arguments", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return true, nil
			},
		}

		cmd := &cobra.Command{}
		cmd.SetIn(strings.NewReader("func main() {}\n"))
		args := []string{"Explain", "this"}

		actual, err := reader.getUserInput(cmd, args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(actual, "func main() {}") {
			t.Errorf("expected actual to contain piped input, got %q", actual)
		}

		if !strings.Contains(actual, "Explain this") {
			t.Errorf("expected actual to contain args, got %q", actual)
		}

		if !strings.Contains(actual, "\n\n") {
			t.Errorf("expected actual to contain separator, got %q", actual)
		}
	})

	t.Run("returns error when stdin stat fails", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return false, errors.New("stat error")
			},
		}

		cmd := &cobra.Command{}
		args := []string{"test"}

		_, err := reader.getUserInput(cmd, args)
		if err == nil {
			t.Fatal("expected error when stdin stat fails, got nil")
		}

		if !errors.Is(err, ErrInput) {
			t.Errorf("expected error to wrap ErrInput, got %v", err)
		}
	})

	t.Run("returns error when reading piped input fails", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		reader := &inputReader{
			logger: logger,
			stdinDetector: func() (bool, error) {
				return true, nil
			},
		}

		cmd := &cobra.Command{}
		cmd.SetIn(&errorReader{failAt: 2})
		args := []string{}

		_, err := reader.getUserInput(cmd, args)
		if err == nil {
			t.Fatal("expected error when reading piped input fails, got nil")
		}

		if !errors.Is(err, ErrIO) {
			t.Errorf("expected error to wrap ErrIO, got %v", err)
		}
	})
}
