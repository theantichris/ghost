package conversation

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// ChatFunc requests one assistant response for an ordered message history.
type ChatFunc func(context.Context, string, []Message) (Message, error)

// ChatStreamFunc streams the assistant response for an ordered message history.
type ChatStreamFunc func(context.Context, string, []Message, func(string) error) error

// TurnID holds the ID of a single chat turn.
type TurnID uint64

// Turn holds the data for a single chat turn.
type Turn struct {
	id       TurnID
	model    string
	prompt   string
	messages []Message
}

// TurnEventKind defines the type of a turn event.
type TurnEventKind string

const (
	// TurnEventContent carries one incremental content chunk.
	TurnEventContent TurnEventKind = "content"

	// TurnEventCompleted carries the full response after successful completion.
	TurnEventCompleted TurnEventKind = "completed"
)

// TurnEvent holds the data for a single event in a chat turn.
type TurnEvent struct {
	TurnID  TurnID
	Kind    TurnEventKind
	Content string
}

// NewTurn creates a new chat turn with a snapshot of chat history with the new
// prompt.
func NewTurn(id TurnID, model string, history []Message, prompt string) Turn {
	messages := slices.Clone(history)
	messages = append(messages, Message{Role: RoleUser, Content: prompt})

	turn := Turn{
		id:       id,
		model:    model,
		prompt:   prompt,
		messages: messages,
	}

	return turn
}

// Stream emits content chunks synchronously followed by a completion event
// containing the assembled response.
// chat must call its content callback synchronously and stop when it returns an
// error.
// Errors or cancellation can follow partial delivery, callers must require a nil
// return before committing.
// Never changes session history.
func (turn Turn) Stream(ctx context.Context, chat ChatStreamFunc, emit func(TurnEvent) error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("stream conversation turn: %w", err)
	}

	var response strings.Builder

	err := chat(ctx, turn.model, slices.Clone(turn.messages), func(content string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := emit(TurnEvent{TurnID: turn.id, Kind: TurnEventContent, Content: content})
		if err != nil {
			return err
		}

		response.WriteString(content)

		return ctx.Err()
	})

	if err != nil {
		return fmt.Errorf("stream conversation turn: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("stream conversation turn: %w", err)
	}

	err = emit(TurnEvent{TurnID: turn.id, Kind: TurnEventCompleted, Content: response.String()})
	if err != nil {
		return fmt.Errorf("stream conversation turn: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("stream conversation turn: %w", err)
	}

	return nil
}

// RunTurnStream forwards turn events and commits the user/assistant exchange
// only after streaming succeeds.
// Any returned error leaves session history unchanged.
// The caller must serialize access to the session, callbacks must not modify it.
// The caller supplies a distinct ID for each new turn so consumers can identify
// obsolete events.
func RunTurnStream(ctx context.Context, session *Session, id TurnID, model string, prompt string, chat ChatStreamFunc, emit func(TurnEvent) error) (Message, error) {
	if err := ctx.Err(); err != nil {
		return Message{}, fmt.Errorf("run streaming conversation turn: %w", err)
	}

	turn := NewTurn(id, model, session.Messages(), prompt)
	var response Message
	err := turn.Stream(ctx, chat, func(event TurnEvent) error {
		if event.Kind == TurnEventCompleted {
			response = Message{Role: RoleAssistant, Content: event.Content}
		}

		return emit(event)
	})
	if err != nil {
		return Message{}, fmt.Errorf("run streaming conversation turn: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return Message{}, fmt.Errorf("run streaming conversation turn: %w", err)
	}

	session.AppendUser(prompt)
	session.AppendAssistant(response.Content)

	return response, nil
}

// RunTurn sends a user prompt with the session history and commits the
// user-assistant exchange only when chat succeeds. Any returned error leaves
// the session unchanged.
func RunTurn(ctx context.Context, session *Session, model string, prompt string, chat ChatFunc) (Message, error) {
	messages := session.Messages()
	messages = append(messages, Message{Role: RoleUser, Content: prompt})

	response, err := chat(ctx, model, messages)
	if err != nil {
		return Message{}, fmt.Errorf("run conversation turn: %w", err)
	}

	session.AppendUser(prompt)
	session.AppendAssistant(response.Content)

	return response, nil
}
