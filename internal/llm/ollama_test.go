package llm

import (
	"errors"
	"io"
	"log/slog"
	"testing"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestNewOllamaClient(t *testing.T) {
	t.Run("creates new Ollama client with default", func(t *testing.T) {
		t.Parallel()

		client, err := NewOllamaClient("http://test.dev", "llama2", logger)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if client.baseURL != "http://test.dev" {
			t.Errorf("expected baseURL to be 'http://test.dev', got '%s'", client.baseURL)
		}

		if client.defaultModel != "llama2" {
			t.Errorf("expected defaultModel to be 'llama2', got '%s'", client.defaultModel)
		}
	})

	t.Run("returns error for empty baseURL", func(t *testing.T) {
		t.Parallel()

		_, err := NewOllamaClient("", "llama2", logger)

		if err == nil {
			t.Fatal("expected error for empty baseURL, got nil")
		}

		if !errors.Is(err, ErrURLEmpty) {
			t.Errorf("expected ErrURLEmpty, got %v", err)
		}
	})

	t.Run("returns error for empty defaultModel", func(t *testing.T) {
		t.Parallel()

		_, err := NewOllamaClient("http://test.dev", "", logger)

		if err == nil {
			t.Fatal("expected error for empty defaultModel, got nil")
		}

		if !errors.Is(err, ErrModelEmpty) {
			t.Errorf("expected ErrModelEmpty, got %v", err)
		}
	})
}

func TestChat(t *testing.T) {
	t.Run("calls Chat method without error", func(t *testing.T) {
		t.Parallel()

		client, err := NewOllamaClient("http://test.dev", "llama2", logger)
		if err != nil {
			t.Fatalf("expected no error creating client, got %v", err)
		}

		err = client.Chat("Hello, there!")
		if err != nil {
			t.Fatalf("expected no error calling Chat, got %v", err)
		}
	})

	t.Run("returns error for missing message", func(t *testing.T) {
		t.Parallel()

		app, err := NewOllamaClient("http://test.dev", "llama2", logger)
		if err != nil {
			t.Fatal("expected no error creating client, got", err)
		}

		err = app.Chat("")
		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}

		if !errors.Is(err, ErrMessageEmpty) {
			t.Errorf("expected ErrMessageEmpty, got %v", err)
		}
	})
}
