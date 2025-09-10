package app

import (
	"errors"
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		t.Parallel()

		app, err := New(&llm.MockLLMClient{})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if app == nil {
			t.Error("expected app to be non-nil")
		}
	})

	t.Run("returns error for nil llmClient", func(t *testing.T) {
		t.Parallel()

		_, err := New(nil)

		if err == nil {
			t.Fatalf("expected error for nil llmClient, got nil")
		}

		if !errors.Is(err, ErrLLMClientNil) {
			t.Errorf("expected ErrLLMClientNil, got %v", err)
		}
	})
}
