package llm

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
