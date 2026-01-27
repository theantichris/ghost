package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestChatModel_HandleCommandMode(t *testing.T) {
	tests := []struct {
		name          string
		cmdBuffer     string
		msg           tea.Msg
		wantMode      Mode
		wantCmdBuffer string
		wantQuit      bool
	}{
		{
			name:          "q command quits",
			cmdBuffer:     "q",
			msg:           tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:      ModeCommand,
			wantCmdBuffer: "q",
			wantQuit:      true,
		},
		{
			name:          "invalid command returns to normal mode",
			cmdBuffer:     "invalid",
			msg:           tea.KeyPressMsg{Code: tea.KeyEnter},
			wantMode:      ModeNormal,
			wantCmdBuffer: "",
			wantQuit:      false,
		},
		{
			name:          "escape returns to normal mode",
			cmdBuffer:     "partial",
			msg:           tea.KeyPressMsg{Code: tea.KeyEscape},
			wantMode:      ModeNormal,
			wantCmdBuffer: "",
			wantQuit:      false,
		},
		{
			name:          "typing appends to buffer",
			cmdBuffer:     "q",
			msg:           tea.KeyPressMsg{Code: 'u', Text: "u"},
			wantMode:      ModeCommand,
			wantCmdBuffer: "qu",
			wantQuit:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModel()
			model.mode = ModeCommand
			model.cmdBuffer = tt.cmdBuffer
			model.ready = true

			newModel, cmd := model.Update(tt.msg)
			got := newModel.(ChatModel)

			if got.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", got.mode, tt.wantMode)
			}

			if got.cmdBuffer != tt.wantCmdBuffer {
				t.Errorf("cmdBuffer = %q, want %q", got.cmdBuffer, tt.wantCmdBuffer)
			}

			if tt.wantQuit && cmd == nil {
				t.Error("expected quit command, got nil")
			}

			if !tt.wantQuit && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}
		})
	}
}
