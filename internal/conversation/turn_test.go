package conversation

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestRunTurn(t *testing.T) {
	session := NewSession(Message{Role: RoleSystem, Content: "You are Ghost."})

	responses := []Message{
		{
			Role:    RoleAssistant,
			Content: "First response.",
		},
		{
			Role:    RoleAssistant,
			Content: "Second response.",
		},
	}

	var calls [][]Message

	chat := func(_ context.Context, model string, messages []Message) (Message, error) {
		if model != "qwen3:8b" {
			t.Errorf("model = %q, want %q", model, "qwen3:8b")
		}

		callIndex := len(calls)
		calls = append(calls, slices.Clone(messages))

		if callIndex >= len(responses) {
			t.Fatalf("Chat() call count exceeded available responses: %d", callIndex+1)
		}

		return responses[callIndex], nil
	}

	if _, err := RunTurn(context.Background(), session, "qwen3:8b", "First signal.", chat); err != nil {
		t.Fatalf("first RunTurn() error = %v, want nil", err)
	}

	if _, err := RunTurn(context.Background(), session, "qwen3:8b", "Second signal.", chat); err != nil {
		t.Fatalf("second RunTurn() error = %v, want nil", err)
	}

	wantCalls := [][]Message{
		{
			{
				Role:    RoleSystem,
				Content: "You are Ghost.",
			},
			{
				Role:    RoleUser,
				Content: "First signal.",
			},
		},
		{
			{
				Role:    RoleSystem,
				Content: "You are Ghost.",
			},
			{
				Role:    RoleUser,
				Content: "First signal.",
			},
			{
				Role:    RoleAssistant,
				Content: "First response.",
			},
			{
				Role:    RoleUser,
				Content: "Second signal.",
			},
		},
	}

	if len(calls) != len(wantCalls) {
		t.Fatalf("Chat() call count = %d, want %d", len(calls), len(wantCalls))
	}

	for index := range wantCalls {
		if !slices.Equal(calls[index], wantCalls[index]) {
			t.Errorf("Chat() call %d messages = %#v, want %#v", index+1, calls[index], wantCalls[index])
		}
	}

	wantHistory := []Message{
		{
			Role:    RoleSystem,
			Content: "You are Ghost.",
		},
		{
			Role:    RoleUser,
			Content: "First signal.",
		},
		{
			Role:    RoleAssistant,
			Content: "First response.",
		},
		{
			Role:    RoleUser,
			Content: "Second signal.",
		},
		{
			Role:    RoleAssistant,
			Content: "Second response.",
		},
	}

	if got := session.Messages(); !slices.Equal(got, wantHistory) {
		t.Errorf("Messages() = %#v, want %#v", got, wantHistory)
	}

	t.Run("leaves session unchanged when chat fails", func(t *testing.T) {
		chatErr := errors.New("model unavailable")

		initial := []Message{
			{
				Role:    RoleSystem,
				Content: "You are Ghost.",
			},
			{
				Role:    RoleUser,
				Content: "Previous signal",
			},
			{
				Role:    RoleAssistant,
				Content: "Previous response.",
			},
		}

		session := NewSession(initial...)

		chat := func(_ context.Context, _ string, _ []Message) (Message, error) {
			return Message{Role: RoleAssistant, Content: "Fabricated response."}, chatErr
		}

		response, err := RunTurn(context.Background(), session, "qwen3:8b", "Failed signal.", chat)

		if err == nil {
			t.Fatal("RunTurn() error = nil, want an error")
		}

		if !errors.Is(err, chatErr) {
			t.Errorf("RunTurn() error = %v, want errors.Is(error, %v)", err, chatErr)
		}

		if response != (Message{}) {
			t.Errorf("RunTurn() response = %#v, want zero-value Message", response)
		}

		if got := session.Messages(); !slices.Equal(got, initial) {
			t.Errorf("Messages() = %#v, want unchanged history %#v", got, initial)
		}
	})

	t.Run("propagates cancellation without changing session", func(t *testing.T) {
		initial := []Message{
			{
				Role:    RoleSystem,
				Content: "You are Ghost.",
			},
			{
				Role:    RoleUser,
				Content: "Previous signal.",
			},
			{
				Role:    RoleAssistant,
				Content: "Previous response.",
			},
		}

		session := NewSession(initial...)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		chat := func(ctx context.Context, _ string, _ []Message) (Message, error) {
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatalf("Chat() context error = %v, want context.Canceled", ctx.Err())
			}

			return Message{Role: RoleAssistant, Content: "Obsolete response."}, ctx.Err()
		}

		response, err := RunTurn(ctx, session, "qwen3:8b", "Cancelled signal.", chat)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("RunTurn() error = %v, want errors.Is(error, context.Canceled)", err)
		}

		if response != (Message{}) {
			t.Errorf("RunTurn() response = %#v, want zero-value Message", response)
		}

		if got := session.Messages(); !slices.Equal(got, initial) {
			t.Errorf("Messages() = %#v, want unchanged history %#v", got, initial)
		}
	})
}
