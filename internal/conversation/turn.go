package conversation

import (
	"context"
	"fmt"
	"slices"
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
