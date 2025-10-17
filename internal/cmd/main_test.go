package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/sebdah/goldie/v2"
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

func TestHandleLLMRequest(t *testing.T) {
	tests := []struct {
		name     string
		writer   io.Writer
		prompt   string
		isGolden bool
		isError  bool
		Error    error
	}{
		{
			name:     "writes user prompt",
			writer:   &bytes.Buffer{},
			prompt:   "what is the capital of tennessee",
			isGolden: true,
		},
		{
			name:    "returns error for bad output",
			writer:  &errorWriter{err: errors.New("error printing output")},
			prompt:  "what is the capital of tennessee",
			isError: true,
			Error:   ErrOutput,
		},
		{
			name:    "returns error for no prompt",
			writer:  &bytes.Buffer{},
			isError: true,
			Error:   ErrNoPrompt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleLLMRequest(tt.prompt, tt.writer)

			if !tt.isError && err != nil {
				t.Fatalf("expected no error got, %s", err)
			}

			if tt.isError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, tt.Error) {
					t.Errorf("expected error %v, got %v", tt.Error, err)
				}
			}

			if tt.isGolden {
				buffer, ok := tt.writer.(*bytes.Buffer)
				if !ok {
					t.Fatalf("expected writer to be of type %T, got %T", &bytes.Buffer{}, buffer)
				}

				g := goldie.New(t)
				g.Assert(t, t.Name(), buffer.Bytes())
			}
		})
	}
}
