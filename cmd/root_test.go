package cmd

import (
	"io"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewRootCmd(t *testing.T) {
	t.Run("creates root command with correct configuration", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		cmd := NewRootCmd(logger)

		if cmd == nil {
			t.Fatal("expected command to be created, got nil")
		}

		if cmd.Use != "ghost" {
			t.Errorf("expected Use to be 'ghost', got %q", cmd.Use)
		}

		if cmd.Short != "A cyberpunk inspired AI assistant." {
			t.Errorf("expected Short to be 'A cyberpunk inspired AI assistant.', got %q", cmd.Short)
		}

		if cmd.Long != "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media." {
			t.Errorf("expected Long to be 'Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.', got %q", cmd.Long)
		}

		configFlag := cmd.PersistentFlags().Lookup("config")
		if configFlag == nil {
			t.Error("expected config flag to be set")
		}

		model := cmd.PersistentFlags().Lookup("model")
		if model == nil {
			t.Error("expected model flag to be set")
		}

		ollama := cmd.PersistentFlags().Lookup("ollama")
		if ollama == nil {
			t.Error("expected ollama flag to be set")
		}

		if cmd.PreRunE == nil {
			t.Error("expected PreRunE to be set")
		}

		// Check for subcommands.
		foundAsk := false
		foundChat := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == "ask" {
				foundAsk = true
			}

			if subCmd.Name() == "chat" {
				foundChat = true
			}
		}

		if !foundAsk {
			t.Error("expected ask subcommand to be added")
		}

		if !foundChat {
			t.Error("expected chat subcommand to be added")
		}
	})
}
