package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// errorWrite is used to test output errors.
type errorWriter struct {
	err error
}

// Write will return an error if one is set, otherwise the length of str.
func (writer *errorWriter) Write(str []byte) (int, error) {
	if writer.err != nil {
		return 0, writer.err
	}

	return len(str), nil
}

func TestRun(t *testing.T) {
	t.Run("writes default text", func(t *testing.T) {
		t.Parallel()

		var writer bytes.Buffer

		err := Run(context.Background(), []string{}, &writer)
		if err != nil {
			t.Fatalf("expect no error got, %v", err)
		}

		actualOutput := writer.String()
		expectedOutput := "ghost system online\n"

		if actualOutput != expectedOutput {
			t.Errorf("expected output %q, got %q", expectedOutput, actualOutput)
		}
	})

	t.Run("returns error for bad output", func(t *testing.T) {
		t.Parallel()

		writer := errorWriter{err: errors.New("error printing output")}

		err := Run(context.Background(), []string{}, &writer)
		if err == nil {
			t.Fatal("expect error, got nil")
		}

		if !errors.Is(err, ErrOutput) {
			t.Errorf("expected error %v, got %v", ErrOutput, err)
		}
	})
}
