package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		t.Parallel()

		app, err := New(context.Background(), &llm.MockLLMClient{}, logger)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if app == nil {
			t.Error("expected app to be non-nil")
		}
	})

	t.Run("returns error for nil llmClient", func(t *testing.T) {
		t.Parallel()

		_, err := New(context.Background(), nil, logger)

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

		app, err := New(context.Background(), &llm.MockLLMClient{}, logger)
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run("test")
		if err != nil {
			t.Fatalf("expected no error running app, got %v", err)
		}
	})

	t.Run("returns error if llmClient.Chat fails", func(t *testing.T) {
		t.Parallel()

		llmClient := &llm.MockLLMClient{
			ReturnError: true,
		}

		app, err := New(context.Background(), llmClient, logger)
		if err != nil {
			t.Fatalf("expected no error creating app, got %v", err)
		}

		err = app.Run("test")
		if err == nil {
			t.Fatal("expected error running app, got nil")
		}

		if !errors.Is(err, llm.ErrMessageEmpty) {
			t.Errorf("expected ErrMessageEmpty, got %v", err)
		}
	})
}
