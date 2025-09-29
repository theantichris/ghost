package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestNewRootCmd(t *testing.T) {
	t.Run("creates root command with correct configuration", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.DiscardHandler)
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

		debugFlag := cmd.PersistentFlags().Lookup("debug")
		if debugFlag == nil {
			t.Error("expected debug flag to be set")
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
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == "ask" {
				found = true
				break
			}
		}

		if !found {
			t.Error("expected ask subcommand to be added")
		}
	})
}

func TestInitConfig(t *testing.T) {
	// TODO: Uncomment when log is implemented.
	// 	t.Run("updates logger level when debug is enabled", func(t *testing.T) {
	// 		logger := log.NewWithOptions(io.Discard, log.Options{
	// 			ReportCaller:    false,
	// 			ReportTimestamp: false,
	// 			Level:           log.WarnLevel,
	// 		})

	// 		viper.Reset()

	// 		tmpDir := t.TempDir()
	// 		configFile := filepath.Join(tmpDir, ".granola.toml")
	// 		configContent := `debug = true`

	// 		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
	// 			t.Fatalf("failed to write to test configFile: %v", err)
	// 		}

	// 		viper.Set("config", configFile)

	// 		initConfig(logger)

	// 		if logger.GetLevel() != log.DebugLevel {
	// 			t.Errorf("expected logger level to be DebugLevel, got %v", logger.GetLevel())
	// 		}

	// 		if !viper.GetBool("debug") {
	// 			t.Error("expected debug mode to be enabled in viper")
	// 		}
	// 	})

	t.Run("loads environment variables from .env file", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.DiscardHandler)
		viper.Reset()

		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		envContent := `DEBUG=true`

		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("failed to write to test .env file: %v", err)
		}

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get the current working directory: %v", err)
		}

		defer func() {
			if err := os.Chdir(oldWd); err != nil {
				t.Fatalf("failed to change to old working directory: %v", err)
			}
		}()

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change to temp directory: %v", err)
		}

		initConfig(logger)
		if !viper.GetBool("debug") {
			t.Error("expected DEBUG_MODE from .env to be loaded")
		}
	})
}
