package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

type faultyReader struct {
	data []byte
	err  error
}

func (reader *faultyReader) Read(p []byte) (n int, err error) {
	if len(reader.data) > 0 {
		n = copy(p, reader.data)
		reader.data = reader.data[n:]
		return n, nil
	}

	return 0, reader.err
}

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		t.Parallel()

		config := Config{Debug: false}
		llmClient := &llm.MockLLMClient{}

		app, err := New(llmClient, logger, config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if app == nil {
			t.Fatal("expected app to be non-nil")
		}

		if app.llmClient != llmClient {
			t.Errorf("unexpected llmClient %v", app.llmClient)
		}

		if app.debug != config.Debug {
			t.Errorf("expected debug to equal %v", config.Debug)
		}

		if app.logger != logger {
			t.Errorf("unexpected logger %v, got %v", logger, app.logger)
		}
	})

	t.Run("returns error for nil llmClient", func(t *testing.T) {
		t.Parallel()

		_, err := New(nil, logger, Config{})
		if err == nil {
			t.Fatalf("expected error for nil llmClient, got nil")
		}

		if !errors.Is(err, ErrLLMClientNil) {
			t.Errorf("expected ErrLLMClientNil, got %v", err)
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("runs the app without error", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: &bytes.Buffer{}})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString(exitCommand+"\n"))
		if err != nil {
			t.Errorf("expected no error running app, got %v", err)
		}
	})

	t.Run("returns error if llmClient.Chat fails", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{Error: ErrChatFailed}
		app, err := New(llmClient, logger, Config{Output: &bytes.Buffer{}})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString("Hello\n"))
		if err == nil {
			t.Fatal("expected error running app, got nil")
		}

		if !errors.Is(err, ErrChatFailed) {
			t.Errorf("expected ErrChatFailed, got %v", err)
		}
	})

	t.Run("handles empty input gracefully", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: &bytes.Buffer{}})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString(" \n"))
		if err != nil {
			t.Errorf("expected no error running app with empty input, got %v", err)
		}
	})

	t.Run("returns error if reading input fails", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: &bytes.Buffer{}})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), &faultyReader{err: errors.New("boom")})
		if err == nil {
			t.Fatal("expected error running app with faulty reader, got nil")
		}

		if !errors.Is(err, ErrReadingInput) {
			t.Errorf("expected ErrReadingInput, got %v", err)
		}
	})

	t.Run("exits gracefully when EOF is reached", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: &bytes.Buffer{}})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString("Hello\n"))
		if err != nil {
			t.Errorf("expected no error running app until EOF, got %v", err)
		}
	})
}

func TestHandleLLMResponse(t *testing.T) {
	t.Run("prints message for ErrClientResponse", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{Error: llm.ErrClientResponse}
		buffer := &bytes.Buffer{}

		app, err := New(llmClient, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		chatHistory := []llm.ChatMessage{{Role: llm.System, Content: "test"}}
		_, err = app.handleLLMResponse(context.Background(), chatHistory)
		if err != nil {
			t.Fatalf("expected handleLLMResponse to recover, got %v", err)
		}

		actual := strings.TrimSpace(buffer.String())

		if actual != msgClientResponse {
			t.Errorf("expected system message %q, got %q", msgClientResponse, actual)
		}
	})

	t.Run("prints message for ErrNon2xxResponse", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{Error: llm.ErrNon2xxResponse}
		buffer := &bytes.Buffer{}

		app, err := New(llmClient, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		chatHistory := []llm.ChatMessage{{Role: llm.System, Content: "test"}}
		_, err = app.handleLLMResponse(context.Background(), chatHistory)
		if err != nil {
			t.Fatalf("expected handleLLMResponse to recover, got %v", err)
		}

		actual := strings.TrimSpace(buffer.String())

		if actual != msgNon2xxResponse {
			t.Errorf("expected system message %q, got %q", msgNon2xxResponse, actual)
		}
	})

	t.Run("prints message for ErrResponseBody", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{Error: llm.ErrResponseBody}
		buffer := &bytes.Buffer{}

		app, err := New(llmClient, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		chatHistory := []llm.ChatMessage{{Role: llm.System, Content: "test"}}
		_, err = app.handleLLMResponse(context.Background(), chatHistory)
		if err != nil {
			t.Fatalf("expected handleLLMResponse to recover, got %v", err)
		}

		actual := strings.TrimSpace(buffer.String())

		if actual != msgResponseBody {
			t.Errorf("expected system message %q, got %q", msgResponseBody, actual)
		}
	})

	t.Run("prints message for ErrUnmarshalResponse", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{Error: llm.ErrUnmarshalResponse}
		buffer := &bytes.Buffer{}

		app, err := New(llmClient, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		chatHistory := []llm.ChatMessage{{Role: llm.System, Content: "test"}}
		_, err = app.handleLLMResponse(context.Background(), chatHistory)
		if err != nil {
			t.Fatalf("expected handleLLMResponse to recover, got %v", err)
		}

		actual := strings.TrimSpace(buffer.String())

		if actual != msgUnmarshalResponse {
			t.Errorf("expected system message %q, got %q", msgUnmarshalResponse, actual)
		}
	})
}

func TestStripThinkBlock(t *testing.T) {
	t.Run("strips thinking block from response", func(t *testing.T) {
		t.Parallel()

		input := "<think>I need to think of a good joke.</think> Why did the chicken cross the road?"

		expected := "Why did the chicken cross the road?"
		actual := stripThinkBlock(input)

		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})

	t.Run("breaks if <think> is not found", func(t *testing.T) {
		t.Parallel()

		input := "Why did the chicken cross the road?</think>"
		actual := stripThinkBlock(input)

		if actual != input {
			t.Errorf("expected %q, got %q", input, actual)
		}
	})

	t.Run("breaks if </think> is not found", func(t *testing.T) {
		t.Parallel()

		input := "<think>Why did the chicken cross the road?"
		actual := stripThinkBlock(input)

		if actual != input {
			t.Errorf("expected %q, got %q", input, actual)
		}
	})
}
