package agent

import "github.com/theantichris/ghost/v3/internal/llm"

// NewMessageHistory creates and returns an initial message history with system
// messages.
func NewMessageHistory(system, format string) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: system},
	}

	if format != "" {
		switch format {
		case "json":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: JSONPrompt})
		case "markdown":
			messages = append(messages, llm.ChatMessage{Role: llm.RoleSystem, Content: MarkdownPrompt})
		}
	}

	return messages
}
