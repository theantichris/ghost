package agent

import "github.com/theantichris/ghost/v3/internal/llm"

// NewMessageHistory creates and returns an initial message history.
func NewMessageHistory(system, prompt, format string) []llm.ChatMessage {
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

	messages = append(messages, llm.ChatMessage{Role: llm.RoleUser, Content: prompt})

	return messages
}
