package cmd

import (
	"testing"

	"github.com/theantichris/ghost/internal/llm"
)

func TestInitMessages(t *testing.T) {
	tests := []struct {
		name    string
		system  string
		prompt  string
		format  string
		wantErr bool
		err     bool
	}{
		{
			name:   "returns message history with no format",
			system: "system prompt",
			prompt: "user prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := initMessages(tt.system, tt.prompt, tt.format)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				expected := llm.ChatMessage{
					Role:    "system",
					Content: tt.system,
				}

				if actual[0] != expected {
					t.Errorf("expected message %v, got %v", expected, actual[0])
				}

				expected = llm.ChatMessage{
					Role:    "user",
					Content: tt.prompt,
				}

				if actual[1] != expected {
					t.Errorf("expected messages %v, got %v", expected, actual[1])
				}
			}
		})
	}
}
