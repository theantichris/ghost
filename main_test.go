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
	"github.com/theantichris/ghost/v4/internal/ollama"
)

type errorWriter struct {
	err error
}

func (writer errorWriter) Write(_ []byte) (int, error) {
	return 0, writer.err
}

func TestRun(t *testing.T) {
	chatErr := errors.New("Ollama unavailable")
	writeErr := errors.New("output unavailable")

	promptMessages := []ollama.Message{
		{Role: conversation.RoleUser, Content: "Identify yourself."},
	}

	tests := []struct {
		name            string
		args            []string
		output          io.Writer
		response        ollama.Message
		chatErr         error
		wantChat        bool
		wantModel       string
		wantMessages    []ollama.Message
		wantOutput      string
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:         "prints assistant response",
			args:         []string{"--model", "qwen3:8b", "Identify", "yourself."},
			output:       &bytes.Buffer{},
			response:     ollama.Message{Role: conversation.RoleAssistant, Content: "Ghost online."},
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
			response:        ollama.Message{Role: conversation.RoleAssistant, Content: "Ghost online."},
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
			response:     ollama.Message{Role: conversation.RoleAssistant, Content: "Ghost online.\n"},
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
			var gotMessages []ollama.Message

			chat := func(_ context.Context, model string, messages []ollama.Message) (ollama.Message, error) {
				chatCalled = true
				gotModel = model
				gotMessages = append([]ollama.Message(nil), messages...)

				return test.response, test.chatErr
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
				t.Errorf("Chat() called = %t, want %t", chatCalled, test.wantChat)
			}

			if test.wantChat {
				if gotModel != test.wantModel {
					t.Errorf("Chat() model = %q, want %q", gotModel, test.wantModel)
				}

				if !slices.Equal(gotMessages, test.wantMessages) {
					t.Errorf("Chat() messages = %#v, want %#v", gotMessages, test.wantMessages)
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
