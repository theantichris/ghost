package conversation

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestTurnStream(t *testing.T) {
	chatErr := errors.New("stream interrupted")
	emitErr := errors.New("receiver unavailable")
	chunk := TurnEvent{TurnID: 42, Kind: TurnEventContent, Content: "Ghost"}
	completed := TurnEvent{TurnID: 42, Kind: TurnEventCompleted, Content: "Ghost"}

	tests := []struct {
		name       string
		chunks     []string
		chatErr    error
		rejectKind TurnEventKind
		wantEvents []TurnEvent
		wantErr    error
	}{
		{
			name: "completes an empty stream",
			wantEvents: []TurnEvent{
				{TurnID: 42, Kind: TurnEventCompleted},
			},
		},
		{
			name:   "preserves chunk order and whitespace",
			chunks: []string{"Ghost", " ", "online.\n"},
			wantEvents: []TurnEvent{
				chunk,
				{TurnID: 42, Kind: TurnEventContent, Content: " "},
				{TurnID: 42, Kind: TurnEventContent, Content: "online.\n"},
				{TurnID: 42, Kind: TurnEventCompleted, Content: "Ghost online.\n"},
			},
		},
		{
			name:    "preserves chat failure before content",
			chatErr: chatErr,
			wantErr: chatErr,
		},
		{
			name:       "does not complete partial content after chat failure",
			chunks:     []string{"Ghost"},
			chatErr:    chatErr,
			wantEvents: []TurnEvent{chunk},
			wantErr:    chatErr,
		},
		{
			name:       "content delivery error stops further chunks",
			chunks:     []string{"Ghost", " must not be delivered"},
			rejectKind: TurnEventContent,
			wantEvents: []TurnEvent{chunk},
			wantErr:    emitErr,
		},
		{
			name:       "preserves completion delivery error",
			chunks:     []string{"Ghost"},
			rejectKind: TurnEventCompleted,
			wantEvents: []TurnEvent{chunk, completed},
			wantErr:    emitErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := []Message{{Role: RoleSystem, Content: "You are Ghost."}}
			session := NewSession(initial...)
			turn := NewTurn(42, "qwen3:8b", session.Messages(), "Identify yourself.")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var events []TurnEvent
			calls := 0

			chat := func(gotCtx context.Context, model string, messages []Message, onContent func(string) error) error {
				calls++
				if gotCtx != ctx {
					t.Error("chat context differs from supplied context")
				}
				if model != "qwen3:8b" {
					t.Errorf("chat model = %q, want %q", model, "qwen3:8b")
				}
				wantMessages := []Message{
					{Role: RoleSystem, Content: "You are Ghost."},
					{Role: RoleUser, Content: "Identify yourself."},
				}
				if !slices.Equal(messages, wantMessages) {
					t.Errorf("chat messages = %#v, want %#v", messages, wantMessages)
				}
				for _, content := range test.chunks {
					if err := onContent(content); err != nil {
						return err
					}
				}
				return test.chatErr
			}

			err := turn.Stream(ctx, chat, func(event TurnEvent) error {
				events = append(events, event)
				if event.Kind == test.rejectKind {
					return emitErr
				}
				return nil
			})

			if !errors.Is(err, test.wantErr) {
				t.Errorf("Stream() error = %v, want %v", err, test.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "stream conversation turn:") {
				t.Errorf("Stream() error = %v, missing operation context", err)
			}
			if calls != 1 {
				t.Errorf("chat calls = %d, want 1", calls)
			}
			if !slices.Equal(events, test.wantEvents) {
				t.Errorf("Stream() events = %#v, want %#v", events, test.wantEvents)
			}
			if got := session.Messages(); !slices.Equal(got, initial) {
				t.Errorf("session history = %#v, want unchanged history %#v", got, initial)
			}
		})
	}
}

func TestTurnStreamCancellation(t *testing.T) {
	tests := []struct {
		name          string
		cancelBefore  bool
		cancelOnEvent TurnEventKind
		chat          func(context.CancelFunc, func(string) error) error
		wantEvents    []TurnEvent
	}{
		{
			name:         "does not start cancelled work",
			cancelBefore: true,
		},
		{
			name: "does not emit content when cancelled before callback",
			chat: func(cancel context.CancelFunc, onContent func(string) error) error {
				cancel()
				return onContent("obsolete content")
			},
		},
		{
			name:          "cancellation during content delivery stops further chunks",
			cancelOnEvent: TurnEventContent,
			chat: func(_ context.CancelFunc, onContent func(string) error) error {
				if err := onContent("partial"); err != nil {
					return err
				}
				return onContent("obsolete content")
			},
			wantEvents: []TurnEvent{
				{TurnID: 42, Kind: TurnEventContent, Content: "partial"},
			},
		},
		{
			name: "does not complete when chat cancels and returns nil",
			chat: func(cancel context.CancelFunc, onContent func(string) error) error {
				if err := onContent("partial"); err != nil {
					return err
				}
				cancel()
				return nil
			},
			wantEvents: []TurnEvent{
				{TurnID: 42, Kind: TurnEventContent, Content: "partial"},
			},
		},
		{
			name: "does not complete an empty stream cancelled by chat",
			chat: func(cancel context.CancelFunc, _ func(string) error) error {
				cancel()
				return nil
			},
		},
		{
			name:          "reports cancellation during completion delivery",
			cancelOnEvent: TurnEventCompleted,
			chat: func(_ context.CancelFunc, _ func(string) error) error {
				return nil
			},
			wantEvents: []TurnEvent{
				{TurnID: 42, Kind: TurnEventCompleted},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if test.cancelBefore {
				cancel()
			}
			var events []TurnEvent
			called := false
			chat := func(_ context.Context, _ string, _ []Message, onContent func(string) error) error {
				called = true
				if test.cancelBefore {
					t.Error("chat called for an already cancelled turn")
					return nil
				}
				return test.chat(cancel, onContent)
			}
			turn := NewTurn(42, "qwen3:8b", nil, "Identify yourself.")
			err := turn.Stream(ctx, chat, func(event TurnEvent) error {
				events = append(events, event)
				if event.Kind == test.cancelOnEvent {
					cancel()
				}
				return nil
			})

			if !errors.Is(err, context.Canceled) {
				t.Errorf("Stream() error = %v, want context.Canceled", err)
			}
			if called == test.cancelBefore {
				t.Errorf("chat called = %t, want %t", called, !test.cancelBefore)
			}
			if !slices.Equal(events, test.wantEvents) {
				t.Errorf("Stream() events = %#v, want %#v", events, test.wantEvents)
			}
		})
	}
}

func TestTurnStreamCancellationWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	started := make(chan struct{})
	finished := make(chan error, 1)
	chat := func(ctx context.Context, _ string, _ []Message, _ func(string) error) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}
	turn := NewTurn(42, "qwen3:8b", nil, "Identify yourself.")
	go func() {
		finished <- turn.Stream(ctx, chat, func(event TurnEvent) error {
			t.Errorf("unexpected event while waiting for content: %#v", event)
			return nil
		})
	}()

	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("chat did not start before timeout")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Stream() error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stream() did not stop after cancellation")
	}
}

func TestTurnStreamConcurrentSnapshots(t *testing.T) {
	initial := []Message{{Role: RoleSystem, Content: "You are Ghost."}}
	first := NewTurn(1, "qwen3:8b", initial, "First signal.")
	second := NewTurn(2, "qwen3:8b", initial, "Second signal.")
	tests := []struct {
		name string
		turn Turn
	}{
		{name: "first execution", turn: first},
		{name: "same turn executed concurrently", turn: first},
		{name: "different turn executed concurrently", turn: second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var events []TurnEvent
			chat := func(_ context.Context, _ string, messages []Message, onContent func(string) error) error {
				messages[0].Content = test.name
				if test.turn.messages[0].Content != "You are Ghost." {
					t.Error("chat mutation changed the stored turn snapshot")
				}
				return onContent(test.turn.prompt)
			}
			err := test.turn.Stream(t.Context(), chat, func(event TurnEvent) error {
				events = append(events, event)
				return nil
			})
			if err != nil {
				t.Fatalf("Stream() error = %v, want nil", err)
			}
			want := []TurnEvent{
				{TurnID: test.turn.id, Kind: TurnEventContent, Content: test.turn.prompt},
				{TurnID: test.turn.id, Kind: TurnEventCompleted, Content: test.turn.prompt},
			}
			if !slices.Equal(events, want) {
				t.Errorf("Stream() events = %#v, want %#v", events, want)
			}
		})
	}
}
