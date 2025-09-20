package llm

import (
	"context"
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
			Error: ErrChatHistoryEmpty,
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
