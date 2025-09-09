package app

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates a new app instance", func(t *testing.T) {
		app, err := New("qwen2.5:7b-instruct", true)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if app.model != "qwen2.5:7b-instruct" {
			t.Errorf("expected model to be 'qwen2.5:7b-instruct', got '%s'", app.model)
		}

		if app.stream != true {
			t.Errorf("expected stream to be true, got %v", app.stream)
		}
	})

	t.Run("returns error for empty model", func(t *testing.T) {
		_, err := New("", true)

		if err == nil {
			t.Fatal("expected error for empty model, got nil")
		}

		if !errors.Is(err, ErrModelEmpty) {
			t.Errorf("expected ErrModelEmpty, got %v", err)
		}
	})
}
