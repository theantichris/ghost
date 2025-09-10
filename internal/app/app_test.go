package app

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{}, logger)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if app == nil {
			t.Error("expected app to be non-nil")
		}
	})

	t.Run("returns error for nil llmClient", func(t *testing.T) {
		t.Parallel()

		_, err := New(nil, logger)

		if err == nil {
			t.Fatalf("expected error for nil llmClient, got nil")
		}

		if !errors.Is(err, ErrLLMClientNil) {
			t.Errorf("expected ErrLLMClientNil, got %v", err)
		}
	})
}
