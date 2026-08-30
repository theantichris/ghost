package conversation

// Role identifies a participant in an chat conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents one message in an Ollama chat conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}
