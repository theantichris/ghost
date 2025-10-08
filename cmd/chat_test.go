package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewChatCmd(t *testing.T) {
	t.Run("creates chat command with correct configuration", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		cmd := NewChatCmd(logger)

		if cmd == nil {
			t.Fatal("expected command to be created, got nil")
		}

		if cmd.Use != "chat" {
			t.Errorf("expected Use to be 'chat', got %q", cmd.Use)
		}

		expectedShort := "Start a chat with Ghost."
		if cmd.Short != expectedShort {
			t.Errorf("expected Short to be %q, got %q", expectedShort, cmd.Short)
		}

		if !strings.Contains(cmd.Long, "Start a chat with Ghost") {
			t.Errorf("expected Long to contain 'Start a chat with Ghost', got %q", cmd.Long)
		}

		if cmd.RunE == nil {
			t.Error("expected RunE to be set")
		}
	})
}
