package conversation

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestNewTurn(t *testing.T) {
	tests := []struct {
		name    string
		history []Message
		prompt  string
		want    []Message
	}{
		{
			name:   "starts with a user prompt",
			prompt: "Identify yourself.",
			want: []Message{
				{Role: RoleUser, Content: "Identify yourself."},
			},
		},
		{
			name: "preserves history order before the new prompt",
			history: []Message{
				{Role: RoleSystem, Content: "You are Ghost."},
				{Role: RoleUser, Content: "First signal."},
				{Role: RoleAssistant, Content: "First response."},
			},
			prompt: "Second signal.",
			want: []Message{
				{Role: RoleSystem, Content: "You are Ghost."},
				{Role: RoleUser, Content: "First signal."},
				{Role: RoleAssistant, Content: "First response."},
				{Role: RoleUser, Content: "Second signal."},
			},
		},
		{
			name:    "preserves prompt whitespace with empty history",
			history: []Message{},
			prompt:  "  signal\n",
			want: []Message{
				{Role: RoleUser, Content: "  signal\n"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			turn := NewTurn(42, "qwen3:8b", test.history, test.prompt)

			if turn.id != 42 {
				t.Errorf("NewTurn() ID = %d, want 42", turn.id)
			}
			if turn.model != "qwen3:8b" {
				t.Errorf("NewTurn() model = %q, want %q", turn.model, "qwen3:8b")
			}
			if turn.prompt != test.prompt {
				t.Errorf("NewTurn() prompt = %q, want %q", turn.prompt, test.prompt)
			}
			if !slices.Equal(turn.messages, test.want) {
				t.Errorf("NewTurn() messages = %#v, want %#v", turn.messages, test.want)
			}
		})
	}
}

func TestNewTurnHistoryIsolation(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{name: "history at capacity", capacity: 1},
		{name: "history with spare capacity", capacity: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backing := make([]Message, test.capacity)
			backing[0] = Message{Role: RoleSystem, Content: "You are Ghost."}
			history := backing[:1]
			before := slices.Clone(backing)

			first := NewTurn(1, "qwen3:8b", history, "First signal.")
			second := NewTurn(2, "qwen3:8b", history, "Second signal.")

			if !slices.Equal(backing, before) {
				t.Errorf("NewTurn() changed caller backing array = %#v, want %#v", backing, before)
			}

			history[0].Content = "Changed outside the turn."

			wantFirst := []Message{
				{Role: RoleSystem, Content: "You are Ghost."},
				{Role: RoleUser, Content: "First signal."},
			}
			wantSecond := []Message{
				{Role: RoleSystem, Content: "You are Ghost."},
				{Role: RoleUser, Content: "Second signal."},
			}

			if !slices.Equal(first.messages, wantFirst) {
				t.Errorf("first turn after caller mutation = %#v, want %#v", first.messages, wantFirst)
			}
			if !slices.Equal(second.messages, wantSecond) {
				t.Errorf("second turn after caller mutation = %#v, want %#v", second.messages, wantSecond)
			}
		})
	}
}

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
		defer cancel()

		chat := func(ctx context.Context, _ string, _ []Message) (Message, error) {
			cancel()
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

func TestRunTurnCancellationGuards(t *testing.T) {
	tests := []struct {
		name         string
		cancelBefore bool
		wantCalls    int
	}{
		{
			name:         "does not call chat for an already cancelled turn",
			cancelBefore: true,
			wantCalls:    0,
		},
		{
			name:      "discards a successful response when chat cancels",
			wantCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := []Message{
				{Role: RoleSystem, Content: "You are Ghost."},
				{Role: RoleUser, Content: "Previous signal."},
				{Role: RoleAssistant, Content: "Previous response."},
			}
			session := NewSession(initial...)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if test.cancelBefore {
				cancel()
			}
			calls := 0
			chat := func(gotCtx context.Context, _ string, _ []Message) (Message, error) {
				calls++
				if gotCtx != ctx {
					t.Error("chat context differs from supplied context")
				}
				cancel()
				return Message{Role: RoleAssistant, Content: "Obsolete response."}, nil
			}

			response, err := RunTurn(ctx, session, "qwen3:8b", "Cancelled signal.", chat)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("RunTurn() error = %v, want context.Canceled", err)
			}
			if response != (Message{}) {
				t.Errorf("RunTurn() response = %#v, want zero-value Message", response)
			}
			if calls != test.wantCalls {
				t.Errorf("chat calls = %d, want %d", calls, test.wantCalls)
			}
			if got := session.Messages(); !slices.Equal(got, initial) {
				t.Errorf("history after cancellation = %#v, want unchanged history %#v", got, initial)
			}
		})
	}
}
