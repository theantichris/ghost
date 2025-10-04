package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/internal/llm"
	"github.com/theantichris/ghost/internal/stdio"
)

func TestNewAskCmd(t *testing.T) {
	t.Run("creates ask command with correct configuration", func(t *testing.T) {
		t.Parallel()

		logger := log.New(io.Discard)
		cmd := NewAskCmd(logger)

		if cmd == nil {
			t.Fatal("expected command to be created, got nil")
		}

		if cmd.Use != "ask [query]" {
			t.Errorf("expected Use to be 'ask [query]', got %q", cmd.Use)
		}

		expectedShort := "Ask Ghost a question."
		if cmd.Short != expectedShort {
			t.Errorf("expected Short to be %q, got %q", expectedShort, cmd.Short)
		}

		if !strings.Contains(cmd.Long, "Ask Ghost a question") {
			t.Errorf("expected Long to contain 'Ask Ghost a question', got %q", cmd.Long)
		}

		if cmd.RunE == nil {
			t.Error("expected RunE to be set")
		}
	})
}

func TestAskCmdRun(t *testing.T) {
	t.Run("handles simple query", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				response := "The capital is Paris."
				for _, char := range response {
					onToken(string(char))
				}
				return nil
			},
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{"What", "is", "the", "capital", "of", "France?"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		output := actualOutput.String()
		if !strings.Contains(output, "The capital is Paris.") {
			t.Errorf("expected output to contain response, got %q", output)
		}
	})

	t.Run("handles arguments", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		var capturedHistory []llm.ChatMessage

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				capturedHistory = chatHistory
				onToken("Response")
				return nil
			},
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{"explain", "this", "code"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(capturedHistory) != 2 {
			t.Errorf("expected 2 messages in history (system + user), got %d", len(capturedHistory))
		}

		userMessage := capturedHistory[1]
		if !strings.Contains(userMessage.Content, "explain this code") {
			t.Errorf("expected user message to contain joined args, got %q", userMessage.Content)
		}
	})

	t.Run("returns error when no input provided", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				onToken("Response")
				return nil
			},
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{})

		if err == nil {
			t.Fatal("expected error when no input provided, got nil")
		}

		if !errors.Is(err, stdio.ErrIO) {
			t.Errorf("expected error to wrap ErrInput, got %v", err)
		}
	})

	t.Run("returns error when LLM fails", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		expectedErr := errors.New("LLM connection failed")
		mockClient := &llm.MockLLMClient{
			Error: expectedErr,
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{"test", "query"})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrLLM) {
			t.Errorf("expected error to wrap ErrLLM, got %v", err)
		}
	})

	t.Run("includes system prompt in chat history", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		var capturedHistory []llm.ChatMessage

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				capturedHistory = chatHistory
				onToken("Response")
				return nil
			},
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{"test"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(capturedHistory) < 2 {
			t.Fatalf("expected at least 2 messages in history, got %d", len(capturedHistory))
		}

		if capturedHistory[0].Role != llm.SystemRole {
			t.Errorf("expected first message to be system, got %v", capturedHistory[0].Role)
		}

		if !strings.Contains(capturedHistory[0].Content, "Ghost") {
			t.Errorf("expected system prompt to mention Ghost, got %q", capturedHistory[0].Content)
		}
	})

	t.Run("flushes output buffer", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				onToken("<th")
				return nil
			},
		}

		askCmd := &askCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewAskCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := askCmd.run(cmd, []string{"test"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		output := actualOutput.String()
		if !strings.Contains(output, "<th") {
			t.Errorf("expected partial tag to be flushed to output, got %q", output)
		}
	})
}
