package llm

import (
	"context"
	"testing"
)

func TestMockLLMClientChat(t *testing.T) {
	t.Run("returns no error when ReturnError is false", func(t *testing.T) {
		t.Parallel()

		client := &MockLLMClient{
			ReturnError: false,
		}

		messageHistory := []ChatMessage{
			{Role: User, Content: "Hello"},
		}

		_, err := client.Chat(context.Background(), messageHistory)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when ReturnError is true", func(t *testing.T) {
		t.Parallel()

		client := &MockLLMClient{
			ReturnError: true,
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
