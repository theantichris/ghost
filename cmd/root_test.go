package cmd

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/go-cmp/cmp"
	"github.com/theantichris/ghost/v3/internal/agent"
	"github.com/theantichris/ghost/v3/internal/llm"
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
				{Role: llm.RoleSystem, Content: agent.JSONPrompt},
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
				{Role: llm.RoleSystem, Content: agent.MarkdownPrompt}, {Role: llm.RoleUser, Content: "user prompt"},
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

func TestGetPipedInput(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "trims whitespace",
			content: "  hello world  \n",
			want:    "hello world",
		},
		{
			name:    "empty input",
			content: "",
			want:    "",
		},
		{
			name:    "multiline input",
			content: "line1\nline2\nline3",
			want:    "line1\nline2\nline3",
		},
		{
			name:    "only whitespace",
			content: "   \n\t\n  ",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard)

			tmpFile, err := os.CreateTemp("", "ghost-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			defer func(tempFile *os.File) {
				_ = os.Remove(tempFile.Name())
			}(tmpFile)

			defer func(tempFile *os.File) {
				_ = tmpFile.Close()
			}(tmpFile)

			_, err = tmpFile.WriteString(tt.content)
			if err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			_, err = tmpFile.Seek(0, 0)
			if err != nil {
				t.Fatalf("Failed to seek temp file: %v", err)
			}

			got, err := getPipedInput(tmpFile, logger)

			if (err != nil) != tt.wantErr {
				t.Fatalf("getPipedInput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("getPipedInput() = %q, want %q", got, tt.want)
			}
		})
	}
}
