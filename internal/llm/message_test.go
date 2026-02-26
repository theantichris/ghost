package llm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewMessageHistory(t *testing.T) {
	tests := []struct {
		name   string
		system string
		format string
		want   []ChatMessage
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			want: []ChatMessage{
				{Role: RoleSystem, Content: "system prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			format: "json",
			want: []ChatMessage{
				{Role: RoleSystem, Content: "system prompt"},
				{Role: RoleSystem, Content: "json prompt"},
			},
		},
		{
			name:   "returns message history with markdown format",
			system: "system prompt",
			format: "markdown",
			want: []ChatMessage{
				{Role: RoleSystem, Content: "system prompt"},
				{Role: RoleSystem, Content: "markdown prompt"},
			},
		},
		{
			name:   "ignores unknown format",
			system: "system prompt",
			format: "unknown",
			want: []ChatMessage{
				{Role: RoleSystem, Content: "system prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessageHistory(tt.system, "json prompt", "markdown prompt", tt.format)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewMessageHistory() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
