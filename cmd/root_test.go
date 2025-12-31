package cmd

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/theantichris/ghost/internal/llm"
)

func TestInitMessages(t *testing.T) {
	tests := []struct {
		name   string
		system string
		prompt string
		format string
		want   []llm.ChatMessage
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			prompt: "user prompt",
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
		{
			name:   "returns message history with JSON format",
			system: "system prompt",
			prompt: "user prompt",
			format: "json",
			want: []llm.ChatMessage{
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
			want: []llm.ChatMessage{
				{Role: llm.RoleSystem, Content: "system prompt"},
				{Role: llm.RoleSystem, Content: markdownPrompt},
				{Role: llm.RoleUser, Content: "user prompt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := initMessages(tt.system, tt.prompt, tt.format)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("initMessages() mismatch (-want +got):\n%s", diff)
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
			name:   "does not return error for empty format",
			format: "",
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

			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateFormat() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("validateFormat() err = %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("validateFormat() error = %v, want no error", err)
			}
		})
	}
}
