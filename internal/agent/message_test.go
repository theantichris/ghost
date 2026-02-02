package agent

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/theantichris/ghost/v3/internal/llm"
)

func TestNewMessageHistory(t *testing.T) {
	tests := []struct {
		name   string
		system string
		format string
		want   []llm.ChatMessage
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			format: "json",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: JSONPrompt},
			},
		},
		{
			name:   "returns message history with markdown format",
			system: "system prompt",
			format: "markdown",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: MarkdownPrompt},
			},
		},
		{
			name:   "ignores unknown format",
			system: "system prompt",
			format: "unknown",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessageHistory(tt.system, tt.format)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewMessageHistory() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
