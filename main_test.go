package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/theantichris/ghost/v4/internal/conversation"
)

type errorWriter struct {
	err error
}

func TestRunIncrementalOutput(t *testing.T) {
	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{name: "empty response", want: "\n"},
		{name: "preserves whitespace", chunks: []string{"  Ghost", "\t", " online.  "}, want: "  Ghost\t online.  \n"},
		{name: "preserves multiple trailing newlines", chunks: []string{"Ghost", "\n", "\n"}, want: "Ghost\n\n"},
		{name: "ignores empty chunks when checking suffix", chunks: []string{"Ghost\n", ""}, want: "Ghost\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			chat := func(_ context.Context, _ string, _ []conversation.Message, onContent func(string) error) error {
				var delivered string
				for _, chunk := range test.chunks {
					if err := onContent(chunk); err != nil {
						return err
					}
					delivered += chunk
					if got := output.String(); got != delivered {
						t.Fatalf("output before chat returns = %q, want %q", got, delivered)
					}
				}
				return nil
			}
			if err := run(t.Context(), []string{"--model", "qwen3:8b", "Hello"}, &output, chat); err != nil {
				t.Fatalf("run() error = %v, want nil", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("run() output = %q, want %q", got, test.want)
			}
		})
	}
}

type writeFunc func([]byte) (int, error)

func (write writeFunc) Write(data []byte) (int, error) {
	return write(data)
}

func TestRunStreamingFailures(t *testing.T) {
	chatErr := errors.New("stream interrupted")
	writeErr := errors.New("stdout unavailable")
	tests := []struct {
		name         string
		cancelBefore bool
		cancelAfter  bool
		chatErr      error
		failWrite    int
		wantOutput   string
		wantErr      error
	}{
		{name: "partial request failure", chatErr: chatErr, wantOutput: "firstsecond", wantErr: chatErr},
		{name: "first chunk write failure", failWrite: 1, wantErr: writeErr},
		{name: "later chunk write failure", failWrite: 2, wantOutput: "first", wantErr: writeErr},
		{name: "final newline write failure", failWrite: 3, wantOutput: "firstsecond", wantErr: writeErr},
		{name: "already canceled", cancelBefore: true, wantErr: context.Canceled},
		{name: "canceled after partial output", cancelAfter: true, wantOutput: "first", wantErr: context.Canceled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if test.cancelBefore {
				cancel()
			}
			var output bytes.Buffer
			writes := 0
			writer := writeFunc(func(data []byte) (int, error) {
				writes++
				if writes == test.failWrite {
					return 0, writeErr
				}
				return output.Write(data)
			})
			chatCalled := false
			chat := func(chatCtx context.Context, _ string, _ []conversation.Message, onContent func(string) error) error {
				chatCalled = true
				if chatCtx != ctx {
					t.Error("chat context differs from supplied context")
				}
				for _, chunk := range []string{"first", "second"} {
					if err := onContent(chunk); err != nil {
						return err
					}
					if test.cancelAfter {
						cancel()
						return chatCtx.Err()
					}
				}
				return test.chatErr
			}
			err := run(ctx, []string{"--model", "qwen3:8b", "Hello"}, writer, chat)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("run() error = %v, want errors.Is(error, %v)", err, test.wantErr)
			}
			if !strings.Contains(err.Error(), "request Ollama chat response") {
				t.Errorf("run() error = %v, missing request context", err)
			}
			if test.failWrite != 0 && !strings.Contains(err.Error(), "write assistant response") {
				t.Errorf("run() error = %v, missing write context", err)
			}
			if chatCalled == test.cancelBefore {
				t.Errorf("chat called = %t, want %t", chatCalled, !test.cancelBefore)
			}
			if got := output.String(); got != test.wantOutput {
				t.Errorf("run() output = %q, want %q", got, test.wantOutput)
			}
			if test.failWrite != 0 && writes != test.failWrite {
				t.Errorf("write calls = %d, want %d", writes, test.failWrite)
			}
		})
	}
}

func (writer errorWriter) Write(_ []byte) (int, error) {
	return 0, writer.err
}

func TestRun(t *testing.T) {
	chatErr := errors.New("Ollama unavailable")
	writeErr := errors.New("output unavailable")

	promptMessages := []conversation.Message{
		{Role: conversation.RoleUser, Content: "Identify yourself."},
	}

	tests := []struct {
		name            string
		args            []string
		output          io.Writer
		chunks          []string
		chatErr         error
		wantChat        bool
		wantModel       string
		wantMessages    []conversation.Message
		wantOutput      string
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:         "prints assistant response",
			args:         []string{"--model", "qwen3:8b", "Identify", "yourself."},
			output:       &bytes.Buffer{},
			chunks:       []string{"Ghost", " online."},
			wantChat:     true,
			wantModel:    "qwen3:8b",
			wantMessages: promptMessages,
			wantOutput:   "Ghost online.\n",
		},
		{
			name:            "requires model",
			args:            []string{"Identify yourself."},
			output:          &bytes.Buffer{},
			wantErr:         true,
			wantErrIs:       errModelRequired,
			wantErrContains: "model is required",
		},
		{
			name:            "requires prompt",
			args:            []string{"--model", "qwen3:8b"},
			output:          &bytes.Buffer{},
			wantErr:         true,
			wantErrIs:       errPromptRequired,
			wantErrContains: "prompt is required",
		},
		{
			name:            "wraps argument parsing error",
			args:            []string{"--unknown"},
			output:          &bytes.Buffer{},
			wantErr:         true,
			wantErrContains: "parse command arguments",
		},
		{
			name:            "wraps chat error",
			args:            []string{"--model", "qwen3:8b", "Identify yourself."},
			output:          &bytes.Buffer{},
			chatErr:         chatErr,
			wantChat:        true,
			wantModel:       "qwen3:8b",
			wantMessages:    promptMessages,
			wantErr:         true,
			wantErrIs:       chatErr,
			wantErrContains: "request Ollama chat response",
		},
		{
			name:            "wraps output error",
			args:            []string{"--model", "qwen3:8b", "Identify yourself."},
			output:          errorWriter{err: writeErr},
			chunks:          []string{"Ghost", " online."},
			wantChat:        true,
			wantModel:       "qwen3:8b",
			wantMessages:    promptMessages,
			wantErr:         true,
			wantErrIs:       writeErr,
			wantErrContains: "write assistant response",
		},
		{
			name:         "does not duplicate final newline",
			args:         []string{"--model", "qwen3:8b", "Identify yourself."},
			output:       &bytes.Buffer{},
			chunks:       []string{"Ghost online.", "\n"},
			wantChat:     true,
			wantModel:    "qwen3:8b",
			wantMessages: promptMessages,
			wantOutput:   "Ghost online.\n",
		},
		{
			name:            "requires nonblank model",
			args:            []string{"--model", "  ", "Identify yourself."},
			output:          &bytes.Buffer{},
			wantErr:         true,
			wantErrIs:       errModelRequired,
			wantErrContains: "model is required",
		},
		{
			name:            "requires nonblank prompt",
			args:            []string{"--model", "qwen3:8b", "  "},
			output:          &bytes.Buffer{},
			wantErr:         true,
			wantErrIs:       errPromptRequired,
			wantErrContains: "prompt is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chatCalled := false

			var gotModel string
			var gotMessages []conversation.Message

			chat := func(_ context.Context, model string, messages []conversation.Message, onContent func(string) error) error {
				chatCalled = true
				gotModel = model
				gotMessages = append([]conversation.Message(nil), messages...)

				for _, content := range test.chunks {
					if err := onContent(content); err != nil {
						return err
					}
				}
				return test.chatErr
			}

			err := run(context.Background(), test.args, test.output, chat)

			if test.wantErr && err == nil {
				t.Fatal("run() error = nil, want an error")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("run() error = %v, want nil", err)
			}

			if test.wantErrIs != nil && !errors.Is(err, test.wantErrIs) {
				t.Errorf("run() error = %v, want errors.Is(error, %v)", err, test.wantErrIs)
			}

			if test.wantErrContains != "" && !strings.Contains(err.Error(), test.wantErrContains) {
				t.Errorf("run() error = %q, want it to contain %q", err, test.wantErrContains)
			}

			if chatCalled != test.wantChat {
				t.Errorf("ChatStream() called = %t, want %t", chatCalled, test.wantChat)
			}

			if test.wantChat {
				if gotModel != test.wantModel {
					t.Errorf("ChatStream() model = %q, want %q", gotModel, test.wantModel)
				}

				if !slices.Equal(gotMessages, test.wantMessages) {
					t.Errorf("ChatStream() messages = %#v, want %#v", gotMessages, test.wantMessages)
				}
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
