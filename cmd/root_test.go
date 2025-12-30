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
		wantErr  bool
		err      bool
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			prompt: "user prompt",
			expected: []llm.ChatMessage{
				{Role: "system", Content: "system prompt"},
				{Role: "user", Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			prompt: "user prompt",
			format: "json",
			expected: []llm.ChatMessage{
				{Role: "system", Content: "system prompt"},
				{Role: "system", Content: "Format the response as json without enclosing backticks."},
				{Role: "user", Content: "user prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := initMessages(tt.system, tt.prompt, tt.format)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if diff := cmp.Diff(tt.expected, actual); diff != "" {
					t.Errorf("expected messages (-want +got):\n%s", diff)
				}
			}
		})
	}
}
