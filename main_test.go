package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type errorWriter struct {
	err error
}

func (writer errorWriter) Write(_ []byte) (int, error) {
	return 0, writer.err
}

func TestRun(t *testing.T) {
	writeErr := errors.New("output unavailable")

	tests := []struct {
		name            string
		output          io.Writer
		wantOutput      string
		wantErr         error
		wantErrContains string
	}{
		{
			name:       "writes bootstrap message",
			output:     &bytes.Buffer{},
			wantOutput: bootstrapMsg + "\n",
		},
		{
			name:            "wraps output error",
			output:          errorWriter{err: writeErr},
			wantErr:         writeErr,
			wantErrContains: "write bootstrap message",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.output)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("run() error = %v, want errors.Is(error, %v)", err, test.wantErr)
			}

			if test.wantErrContains != "" && !strings.Contains(err.Error(), test.wantErrContains) {
				t.Errorf("run() error = %q, want it to contain %q", err, test.wantErrContains)
			}

			buffer, ok := test.output.(*bytes.Buffer)
			if !ok {
				return
			}

			if got := buffer.String(); got != test.wantOutput {
				t.Errorf("run() output = %q, want %q", got, test.wantOutput)
			}
		})
	}
}
