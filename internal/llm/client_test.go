package llm

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestMockLLMClientChat(t *testing.T) {
	t.Run("returns no error when Error is not set", func(t *testing.T) {
		t.Parallel()

		client := &MockLLMClient{}

		messageHistory := []ChatMessage{
			{Role: User, Content: "Hello"},
		}

		_, err := client.Chat(context.Background(), messageHistory)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when Error is set", func(t *testing.T) {
		t.Parallel()

		client := &MockLLMClient{
			Error: ErrValidation,
		}

		messageHistory := []ChatMessage{
			{Role: User, Content: "Hello"},
		}

		_, err := client.Chat(context.Background(), messageHistory)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestMockLLMClientStreamChat(t *testing.T) {
	t.Run("returns error when Error is set", func(t *testing.T) {
		t.Parallel()

		client := &MockLLMClient{
			Error: ErrValidation,
		}

		messageHistory := []ChatMessage{{Role: User, Content: "Hello"}}

		err := client.StreamChat(context.Background(), messageHistory, func(token string) {})
		if err == nil {
			t.Fatal("expected SteamChat to return error but got nil")
		}

		if !errors.Is(err, ErrValidation) {
			t.Errorf("expected error %v, got %v", ErrValidation, err)
		}
	})

	t.Run("runs StreamChat", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer

		mockStreamChat := func(ctx context.Context, messageHistory []ChatMessage, onToken func(string)) error {
			buffer.WriteString("Hello")

			return nil
		}

		client := &MockLLMClient{
			StreamChatFunc: mockStreamChat,
		}

		messageHistory := []ChatMessage{{Role: User, Content: "Hello"}}

		err := client.StreamChat(context.Background(), messageHistory, func(string) {})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if buffer.String() != "Hello" {
			t.Errorf("expected output %q, got %q", "Hello", buffer.String())
		}
	})
}
