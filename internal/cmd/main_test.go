package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
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
			t.Fatalf("expect no error got, %s", err)
		}

		golden, err := os.ReadFile(filepath.Join("../../testdata", t.Name()+".golden"))
		if err != nil {
			t.Fatalf("error reading golden file, %s", err)
		}

		if !bytes.Equal(writer.Bytes(), golden) {
			t.Errorf("written output does not matching golden")
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
