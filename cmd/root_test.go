package cmd

import (
	"errors"
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

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
		err     error
	}{
		{
			name:   "does not return error for json",
			format: "json",
		},
		{
			name:   "does not return error for markdown",
			format: "markdown",
		},
		{
			name:    "returns error for invalid format",
			format:  "butts",
			wantErr: true,
			err:     ErrInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormat(tt.format)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if !errors.Is(err, ErrInvalidFormat) {
					t.Errorf("expected error %v, got %v", ErrInvalidFormat, err)
				}
			}
		})
	}
}
