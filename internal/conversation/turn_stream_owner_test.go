package conversation

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestRunTurnStream(t *testing.T) {
	session := NewSession(Message{Role: RoleSystem, Content: "You are Ghost."})
	tests := []struct {
		name   string
		id     TurnID
		prompt string
		chunks []string
	}{
		{
			name:   "commits an assembled response",
			id:     1,
			prompt: "First signal.",
			chunks: []string{"Ghost", " online.\n"},
		},
		{
			name:   "preserves the previous exchange and commits an empty response",
			id:     2,
			prompt: "Second signal.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := session.Messages()
			wantRequest := append(slices.Clone(before), Message{Role: RoleUser, Content: test.prompt})
			wantResponse := Message{Role: RoleAssistant, Content: strings.Join(test.chunks, "")}
			var events []TurnEvent
			calls := 0
			chat := func(ctx context.Context, model string, messages []Message, onContent func(string) error) error {
				calls++
				if ctx != t.Context() {
					t.Error("chat context differs from supplied context")
				}
				if model != "qwen3:8b" {
					t.Errorf("chat model = %q, want %q", model, "qwen3:8b")
				}
				if !slices.Equal(messages, wantRequest) {
					t.Errorf("chat messages = %#v, want %#v", messages, wantRequest)
				}
				for _, content := range test.chunks {
					if err := onContent(content); err != nil {
						return err
					}
				}
				return nil
			}
			response, err := RunTurnStream(t.Context(), session, test.id, "qwen3:8b", test.prompt, chat, func(event TurnEvent) error {
				if got := session.Messages(); !slices.Equal(got, before) {
					t.Errorf("history during event delivery = %#v, want unchanged history %#v", got, before)
				}
				events = append(events, event)
				return nil
			})
			if err != nil {
				t.Fatalf("RunTurnStream() error = %v, want nil", err)
			}
			if response != wantResponse {
				t.Errorf("RunTurnStream() response = %#v, want %#v", response, wantResponse)
			}
			if calls != 1 {
				t.Errorf("chat calls = %d, want 1", calls)
			}
			wantHistory := append(slices.Clone(wantRequest), wantResponse)
			if got := session.Messages(); !slices.Equal(got, wantHistory) {
				t.Errorf("history after success = %#v, want %#v", got, wantHistory)
			}
			var wantEvents []TurnEvent
			for _, content := range test.chunks {
				wantEvents = append(wantEvents, TurnEvent{TurnID: test.id, Kind: TurnEventContent, Content: content})
			}
			wantEvents = append(wantEvents, TurnEvent{TurnID: test.id, Kind: TurnEventCompleted, Content: wantResponse.Content})
			if !slices.Equal(events, wantEvents) {
				t.Errorf("forwarded events = %#v, want %#v", events, wantEvents)
			}
		})
	}
}

func TestRunTurnStreamFailureLeavesHistoryUnchanged(t *testing.T) {
	chatErr := errors.New("stream interrupted")
	emitErr := errors.New("receiver unavailable")
	tests := []struct {
		name            string
		cancelBefore    bool
		cancelAfterChat bool
		cancelOnEvent   TurnEventKind
		rejectKind      TurnEventKind
		chatErr         error
		wantErr         error
		wantKinds       []TurnEventKind
	}{
		{
			name:         "already cancelled",
			cancelBefore: true,
			wantErr:      context.Canceled,
		},
		{
			name:      "chat failure after partial delivery",
			chatErr:   chatErr,
			wantErr:   chatErr,
			wantKinds: []TurnEventKind{TurnEventContent},
		},
		{
			name:            "chat cancels after partial delivery and returns nil",
			cancelAfterChat: true,
			wantErr:         context.Canceled,
			wantKinds:       []TurnEventKind{TurnEventContent},
		},
		{
			name:       "content receiver rejects partial delivery",
			rejectKind: TurnEventContent,
			wantErr:    emitErr,
			wantKinds:  []TurnEventKind{TurnEventContent},
		},
		{
			name:          "content receiver cancels",
			cancelOnEvent: TurnEventContent,
			wantErr:       context.Canceled,
			wantKinds:     []TurnEventKind{TurnEventContent},
		},
		{
			name:       "completion receiver rejects the assembled response",
			rejectKind: TurnEventCompleted,
			wantErr:    emitErr,
			wantKinds:  []TurnEventKind{TurnEventContent, TurnEventCompleted},
		},
		{
			name:          "completion receiver cancels after response capture",
			cancelOnEvent: TurnEventCompleted,
			wantErr:       context.Canceled,
			wantKinds:     []TurnEventKind{TurnEventContent, TurnEventCompleted},
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
			called := false
			chat := func(_ context.Context, _ string, _ []Message, onContent func(string) error) error {
				called = true
				if err := onContent("partial response"); err != nil {
					return err
				}
				if test.cancelAfterChat {
					cancel()
				}
				return test.chatErr
			}
			var kinds []TurnEventKind
			response, err := RunTurnStream(ctx, session, 42, "qwen3:8b", "Failed signal.", chat, func(event TurnEvent) error {
				kinds = append(kinds, event.Kind)
				if got := session.Messages(); !slices.Equal(got, initial) {
					t.Errorf("history during event delivery = %#v, want %#v", got, initial)
				}
				if event.Kind == test.cancelOnEvent {
					cancel()
				}
				if event.Kind == test.rejectKind {
					return emitErr
				}
				return nil
			})
			if !errors.Is(err, test.wantErr) {
				t.Errorf("RunTurnStream() error = %v, want %v", err, test.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "run streaming conversation turn:") {
				t.Errorf("RunTurnStream() error = %v, missing operation context", err)
			}
			if response != (Message{}) {
				t.Errorf("RunTurnStream() response = %#v, want zero-value Message", response)
			}
			if called == test.cancelBefore {
				t.Errorf("chat called = %t, want %t", called, !test.cancelBefore)
			}
			if !slices.Equal(kinds, test.wantKinds) {
				t.Errorf("forwarded event kinds = %#v, want %#v", kinds, test.wantKinds)
			}
			if got := session.Messages(); !slices.Equal(got, initial) {
				t.Errorf("history after failure = %#v, want unchanged history %#v", got, initial)
			}
		})
	}
}
