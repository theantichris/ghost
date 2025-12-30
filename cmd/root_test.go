package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/theantichris/ghost/internal/llm"
)

func TestInitMessages(t *testing.T) {
	tests := []struct {
		name     string
		system   string
		prompt   string
		format   string
		expected []llm.ChatMessage
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			prompt: "user prompt",
			expected: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			prompt: "user prompt",
			format: "json",
			expected: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: jsonPrompt},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with markdown format",
			system: "system prompt",
			prompt: "user prompt",
			format: "markdown",
			expected: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: markdownPrompt},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := initMessages(tt.system, tt.prompt, tt.format)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("expected messages (-want +got):\n%s", diff)
			}
		})
	}
}
