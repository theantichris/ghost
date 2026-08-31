package conversation

import (
	"context"
	"fmt"
)

// ChatFunc requests one assistant response for an ordered message history.
type ChatFunc func(context.Context, string, []Message) (Message, error)

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
