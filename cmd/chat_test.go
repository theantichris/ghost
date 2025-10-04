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

func TestChatCmdRun(t *testing.T) {
	t.Run("handles greeting and single message exchange", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				response := "Hello! How can I help?"
				for _, char := range response {
					onToken(string(char))
				}
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader("test message\n"))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		output := actualOutput.String()
		if !strings.Contains(output, "Hello! How can I help?") {
			t.Errorf("expected output to contain greeting, got %q", output)
		}
	})

	t.Run("handles /bye command", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		callCount := 0
		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				callCount++
				if callCount == 1 {
					onToken("Welcome!")
				} else {
					lastMsg := chatHistory[len(chatHistory)-1]
					if lastMsg.Content == "Goodbye!" {
						onToken("Farewell!")
					}
				}
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader("/bye\n"))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if callCount != 2 {
			t.Errorf("expected 2 LLM calls (greeting + goodbye), got %d", callCount)
		}

		output := actualOutput.String()
		if !strings.Contains(output, "Farewell!") {
			t.Errorf("expected output to contain goodbye message, got %q", output)
		}
	})

	t.Run("handles /exit command", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		callCount := 0
		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				callCount++
				onToken("Response")
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader("/exit\n"))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if callCount != 2 {
			t.Errorf("expected 2 LLM calls (greeting + goodbye), got %d", callCount)
		}
	})

	t.Run("skips empty input and continues", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		callCount := 0
		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				callCount++
				onToken("Response")
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader("\n\nhello\n/bye\n"))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if callCount != 3 {
			t.Errorf("expected 3 LLM calls (greeting + hello + goodbye), got %d", callCount)
		}
	})

	t.Run("handles EOF gracefully", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				onToken("Greeting")
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error on EOF, got %v", err)
		}
	})

	t.Run("returns error when LLM fails on greeting", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		expectedErr := errors.New("LLM connection failed")
		mockClient := &llm.MockLLMClient{
			Error: expectedErr,
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader(""))

		err := chatCmd.run(cmd, []string{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrLLM) {
			t.Errorf("expected error to wrap ErrLLM, got %v", err)
		}
	})

	t.Run("maintains chat history across turns", func(t *testing.T) {
		t.Parallel()

		var actualOutput bytes.Buffer
		logger := log.New(io.Discard)

		var capturedHistory []llm.ChatMessage
		callCount := 0

		mockClient := &llm.MockLLMClient{
			ChatFunc: func(ctx context.Context, chatHistory []llm.ChatMessage, onToken func(string)) error {
				callCount++
				capturedHistory = chatHistory
				onToken("Response")
				return nil
			},
		}

		chatCmd := &chatCmd{
			logger:    logger,
			llmClient: mockClient,
		}

		cmd := NewChatCmd(logger)
		cmd.SetOut(&actualOutput)
		cmd.SetIn(strings.NewReader("first message\n/bye\n"))

		err := chatCmd.run(cmd, []string{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(capturedHistory) < 4 {
			t.Errorf("expected at least 4 messages in history (system + greeting + user + assistant), got %d", len(capturedHistory))
		}

		if capturedHistory[0].Role != llm.SystemRole {
			t.Errorf("expected first message to be system, got %v", capturedHistory[0].Role)
		}
	})
}
