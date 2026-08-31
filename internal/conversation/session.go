package conversation

import "slices"

// Role identifies a participant in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents one entry in a conversation.
type Message struct {
	Role    Role
	Content string
}

// Session owns the ordered history for one conversation.
type Session struct {
	messages []Message
}

// NewSession constructs a session with a copy of the supplied initial context.
func NewSession(initial ...Message) *Session {
	return &Session{messages: slices.Clone(initial)}
}

// AppendUser appends a user message to the conversation.
func (session *Session) AppendUser(content string) {
	session.messages = append(session.messages, Message{
		Role:    RoleUser,
		Content: content,
	})
}

// AppendAssistant appends an assistant message to the conversation.
func (session *Session) AppendAssistant(content string) {
	session.messages = append(session.messages, Message{
		Role:    RoleAssistant,
		Content: content,
	})
}

// Messages returns a copy of the conversation history.
func (session *Session) Messages() []Message {
	return slices.Clone(session.messages)
}
