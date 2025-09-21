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

		config := Config{Debug: false, Output: &bytes.Buffer{}}
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

		_, err := New(nil, logger, Config{Output: &bytes.Buffer{}})
		if err == nil {
			t.Fatalf("expected error for nil llmClient, got nil")
		}

		if !errors.Is(err, ErrLLMClientNil) {
			t.Errorf("expected ErrLLMClientNil, got %v", err)
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("outputs LLM messages and handles exit", func(t *testing.T) {
		t.Parallel()

		callCount := 0 // Used to simulate the two streams in Run()

		llmClient := &llm.MockLLMClient{
			StreamChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage,
				onToken func(string)) error {
				callCount++

				switch callCount {
				case 1:
					tokens := []string{"<think>thinking</think>Hello", ", ", "user", "!\n"}

					for _, token := range tokens {
						onToken(token)
					}
				case 2:
					tokens := []string{"<think>thinking</think>Goodbye", ", ", "user", "!\n"}

					for _, token := range tokens {
						onToken(token)
					}
				}

				return nil
			},
		}

		var outputBuff bytes.Buffer

		app, err := New(llmClient, logger, Config{Output: &outputBuff})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString(exitCommand+"\n"))
		if err != nil {
			t.Fatalf("expected no error running app, got %v", err)
		}

		actual := outputBuff.String()
		expected := "\nGhost: Hello, user!\n\n\nUser: \nGhost: Goodbye, user!\n\n"

		if actual != expected {
			t.Errorf("expected response %q, got %q", expected, actual)
		}
	})

	t.Run("returns error when chat fails at LLM greeting", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{
			StreamChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage,
				onToken func(string)) error {

				return ErrChatFailed
			},
		}

		var outputBuff bytes.Buffer

		app, err := New(llmClient, logger, Config{Output: &outputBuff})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString(exitCommand+"\n"))
		if err == nil {
			t.Fatalf("received no error when expecting one")
		}

		if !errors.Is(err, ErrChatFailed) {
			t.Errorf("expected error %v, got %v", ErrChatFailed, err)
		}
	})

	t.Run("returns error when chat fails in chat loop", func(t *testing.T) {
		t.Parallel()

		callCount := 0 // Used to simulate the two streams in Run()

		llmClient := &llm.MockLLMClient{
			StreamChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage,
				onToken func(string)) error {
				callCount++

				switch callCount {
				case 1:
					tokens := []string{"Hello", ", ", "user", "!\n"}

					for _, token := range tokens {
						onToken(token)
					}
				case 2:
					return ErrChatFailed
				}

				return nil
			},
		}

		var outputBuff bytes.Buffer

		app, err := New(llmClient, logger, Config{Output: &outputBuff})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), bytes.NewBufferString(exitCommand+"\n"))
		if err == nil {
			t.Fatalf("expected no error running app, got %v", err)
		}

		if !errors.Is(err, ErrChatFailed) {
			t.Errorf("expected error %v, got %v", ErrChatFailed, err)
		}

		actual := outputBuff.String()
		expected := "\nGhost: Hello, user!\n\n\nUser: \nGhost: \n"

		if actual != expected {
			t.Errorf("expected response %q, got %q", expected, actual)
		}
	})

	t.Run("returns error for faulty user input", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{
			StreamChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage,
				onToken func(string)) error {
				tokens := []string{"Hello", ", ", "user", "!\n"}

				for _, token := range tokens {
					onToken(token)
				}

				return nil
			},
		}

		var outputBuff bytes.Buffer

		app, err := New(llmClient, logger, Config{Output: &outputBuff})
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run(context.Background(), &faultyReader{})
		if err == nil {
			t.Fatal("expected error, received nil")
		}

		if !errors.Is(err, ErrReadingInput) {
			t.Errorf("expected error %v, got %v", ErrReadingInput, err)
		}
	})
}

func TestHandleLLMResponseError(t *testing.T) {
	t.Run("prints message for ErrClientResponse", func(t *testing.T) {
		t.Parallel()

		buffer := &bytes.Buffer{}

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = app.handleLLMResponseError(llm.ErrClientResponse)
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

		buffer := &bytes.Buffer{}

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = app.handleLLMResponseError(llm.ErrNon2xxResponse)
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

		buffer := &bytes.Buffer{}

		app, err := New(&llm.MockLLMClient{}, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = app.handleLLMResponseError(llm.ErrResponseBody)
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

		buffer := &bytes.Buffer{}

		app, err := New(llm.MockLLMClient{}, logger, Config{Output: buffer})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = app.handleLLMResponseError(llm.ErrUnmarshalResponse)
		if err != nil {
			t.Fatalf("expected handleLLMResponse to recover, got %v", err)
		}

		actual := strings.TrimSpace(buffer.String())

		if actual != msgUnmarshalResponse {
			t.Errorf("expected system message %q, got %q", msgUnmarshalResponse, actual)
		}
	})
}
