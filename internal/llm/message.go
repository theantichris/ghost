package llm

// Role represents the author of a message in the chat history.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage holds a single message in the chat history.
type ChatMessage struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	Images    []string   `json:"images,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessageHistory takes system and format prompts and returns an initial message
// history with system messages.
func NewMessageHistory(system, jsonPrompt, markdownPrompt, format string) []ChatMessage {
	messages := []ChatMessage{
		{Role: RoleSystem, Content: system},
	}

	if format != "" {
		switch format {
		case "json":
			messages = append(messages, ChatMessage{Role: RoleSystem, Content: jsonPrompt})
		case "markdown":
			messages = append(messages, ChatMessage{Role: RoleSystem, Content: markdownPrompt})
		}
	}

	return messages
}
